package impl

import (
	"context"
	"fmt"
	"github.com/minh6824pro/nxrGO/cache"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/event"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
	"github.com/minh6824pro/nxrGO/utils"
	"github.com/payOSHQ/payos-lib-golang"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"time"
)

type orderService struct {
	db                  *gorm.DB
	productVariantRepo  repositories.ProductVariantRepository
	orderItemRepo       repositories.OrderItemRepository
	orderRepo           repositories.OrderRepository
	draftOrderRepo      repositories.DraftOrderRepository
	paymentInfoRepo     repositories.PaymentInfoRepository
	productVariantCache cache.ProductVariantRedis
	eventBus            event.EventPublisher
	pendingReplies      map[string]chan dto.OrderProcessingResult // correlationID -> reply channel
}

func NewOrderService(db *gorm.DB, productVariantRepo repositories.ProductVariantRepository, orderItemRepo repositories.OrderItemRepository,
	orderRepo repositories.OrderRepository, draftOrderRepo repositories.DraftOrderRepository,
	paymentInfoRepo repositories.PaymentInfoRepository,
	productVariantCache cache.ProductVariantRedis,
	eventBus event.EventPublisher) services.OrderService {
	service := &orderService{
		db:                  db,
		productVariantRepo:  productVariantRepo,
		orderRepo:           orderRepo,
		orderItemRepo:       orderItemRepo,
		draftOrderRepo:      draftOrderRepo,
		paymentInfoRepo:     paymentInfoRepo,
		productVariantCache: productVariantCache,
		eventBus:            eventBus,
		pendingReplies:      make(map[string]chan dto.OrderProcessingResult),
	}
	service.registerEventHandlers()

	return service
}

func (o *orderService) Create(ctx context.Context, input dto.CreateOrderInput) (*models.Order, error) {
	var total float64 = 0

	luaScript := `
local itemCount = tonumber(ARGV[1])
local expectedTotal = tonumber(ARGV[2]) -- total client gửi vào
local idx = 3

-- 1) Check for MISS (missing keys)
local missed = {}
for i = 1, itemCount do
    local variantId = ARGV[idx]
    idx = idx + 3
    local key = KEYS[i]
    if redis.call("EXISTS", key) == 0 then
        table.insert(missed, variantId)
    end
end
if #missed > 0 then
    local ret = {"MISS"}
    for i = 1, #missed do
        table.insert(ret, missed[i])
    end
    return ret
end

-- 2) Check stock, price and calculate total
idx = 3
local updates = {}
local totalPrice = 0
for i = 1, itemCount do
    local variantId = ARGV[idx]
    local qtyRequested = tonumber(ARGV[idx + 1])
    local priceExpected = tonumber(ARGV[idx + 2])
    idx = idx + 3
    local key = KEYS[i]

    local stock = tonumber(redis.call("HGET", key, "quantity"))
    local price = tonumber(redis.call("HGET", key, "price"))

    if stock == nil or price == nil then
        return {"MISS", variantId}
    end
    if qtyRequested > stock then
        return {"INSUFFICIENT", variantId}
    end
    if priceExpected ~= price then
        return {"INVALID_PRICE", variantId}
    end

    totalPrice = totalPrice + (price * qtyRequested)

    table.insert(updates, { key = key, qty = qtyRequested })
end

-- 3) Check total price
if totalPrice ~= expectedTotal then
    return {"INVALID_TOTAL",totalPrice}
end

-- 4) Deduct stock if all checks pass
for i = 1, #updates do
    redis.call("HINCRBY", updates[i].key, "quantity", -updates[i].qty)
end

return {"OK"}

`

	// build keys & args
	keys := make([]string, 0, len(input.OrderItems))
	args := []interface{}{len(input.OrderItems), input.Total - input.ShippingFee}
	for _, oi := range input.OrderItems {
		keys = append(keys, fmt.Sprintf(cache.ProductVariantKeyPattern, oi.ProductVariantID))
		// ARGV: variantId, quantity, price (Lua uses toNumber when needed)
		args = append(args, oi.ProductVariantID, oi.Quantity, oi.Price)
	}

	// Helper to safely convert interface{} to string
	toStr := func(v interface{}) string {
		switch t := v.(type) {
		case string:
			return t
		case []byte:
			return string(t)
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	const maxRetries = 2
	var res interface{}
	var err error

	// First run of the Lua script
	res, err = o.productVariantCache.EvalLua(ctx, luaScript, keys, args...)
	if err != nil {
		return nil, err
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		arr, ok := res.([]interface{})
		if !ok || len(arr) == 0 {
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected redis lua response", http.StatusInternalServerError, nil)
		}

		status := toStr(arr[0])

		switch status {
		case "MISS":
			// arr[1:] contains list of missing variantIds
			var missingIDs []uint
			for i := 1; i < len(arr); i++ {
				idStr := toStr(arr[i])
				id64, parseErr := strconv.ParseUint(idStr, 10, 64)
				if parseErr != nil {
					return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Invalid variant id from redis", http.StatusInternalServerError, nil)
				}
				missingIDs = append(missingIDs, uint(id64))
			}
			log.Print("Redis key miss: ", missingIDs)

			// Batch query DB once for all missingIDs
			variants, err := o.productVariantRepo.GetByIDSForRedisCache(ctx, missingIDs)
			if err != nil {
				return nil, err
			}

			// If DB returns fewer records, return ITEM_NOT_FOUND error for missing ids
			if len(variants) != len(missingIDs) {
				found := map[uint]bool{}
				for _, v := range variants {
					found[v.ID] = true
				}
				for _, id := range missingIDs {
					if !found[id] {
						return nil, customErr.NewError(
							customErr.ITEM_NOT_FOUND,
							fmt.Sprintf("Product variant not found: %d", id),
							http.StatusBadRequest, nil,
						)
					}
				}
			}

			// Save all queried variants to Redis as Hash
			for _, pv := range variants {
				// save as hash: id, quantity, price
				if err := o.productVariantCache.SaveProductVariantHash(pv); err != nil {
					log.Printf("warning: save productVariant %d to redis failed: %v", pv.ID, err)
				}
			}

			// retry: run Lua script again
			res, err = o.productVariantCache.EvalLua(ctx, luaScript, keys, args...)
			if err != nil {
				return nil, err
			}
			continue

		case "INSUFFICIENT":
			variantId := ""
			if len(arr) > 1 {
				variantId = toStr(arr[1])
			}
			log.Print("Redis key miss: ", variantId)
			return nil, customErr.NewError(customErr.INSUFFICIENT_STOCK, fmt.Sprintf("Product variant : %s Insufficient stock", variantId), http.StatusBadRequest, nil)

		case "INVALID_PRICE":
			variantId := ""
			if len(arr) > 1 {
				variantId = toStr(arr[1])
			}
			log.Print(fmt.Sprintf("Product %s invalid price: ", variantId))

			return nil, customErr.NewError(customErr.INVALID_PRICE, fmt.Sprintf("Product variant: %s Price Invalid", variantId), http.StatusBadRequest, nil)
		case "INVALID_TOTAL":
			expectedTotal := float64(0)
			if len(arr) > 1 {
				expectedTotalStr := toStr(arr[1])
				expectedTotal, err = strconv.ParseFloat(expectedTotalStr, 64)
				if err != nil {
					return nil, customErr.NewError(customErr.INVALID_PRICE, "Total price not match", http.StatusBadRequest, nil)
				}
				expectedTotal += input.ShippingFee
			}
			log.Print("Invalid total, expected: ", expectedTotal)

			return nil, customErr.NewError(customErr.INVALID_PRICE, fmt.Sprintf("Total price not match, expected: %.2f", expectedTotal), http.StatusBadRequest, nil)
		case "OK":
			// all OK -> calculate total from input (price already verified)
			for _, oi := range input.OrderItems {
				total += oi.Price * float64(oi.Quantity)
			}
			// exit loop to continue creating order
			attempt = maxRetries // break outer loop
			break

		default:
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected Lua script status", http.StatusInternalServerError, nil)
		}
	}

	// If after maxRetries still not OK => error
	resArr, _ := res.([]interface{})
	if len(resArr) == 0 || toStr(resArr[0]) != "OK" {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Failed to reserve stock after retries", http.StatusInternalServerError, nil)
	}

	// Create Draft Order
	draftOrder := models.DraftOrder{
		UserID:          input.UserID,
		Status:          models.OrderStatePending,
		Total:           input.Total,
		ShippingAddress: input.ShippingAddress,
		ShippingFee:     input.ShippingFee,
		PaymentMethod:   input.PaymentMethod,
		PhoneNumber:     input.PhoneNumber,
	}
	if err := o.draftOrderRepo.Create(ctx, &draftOrder); err != nil {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, fmt.Sprintf("Draft order creation error: %v", err.Error()), http.StatusBadRequest, nil)
	}

	// Create order items
	var orderItems []models.OrderItem
	for _, item := range input.OrderItems {
		orderItem := models.OrderItem{
			OrderID:          draftOrder.ID,
			ProductVariantID: item.ProductVariantID,
			OrderType:        models.OrderTypeDraftOrder,
			Quantity:         item.Quantity,
			Price:            item.Price,
			TotalPrice:       item.Price * float64(item.Quantity),
		}
		if err := o.orderItemRepo.Create(ctx, &orderItem); err != nil {
			return nil, err
		}
		orderItems = append(orderItems, orderItem)
	}

	if err := o.CreatePayment(ctx, &draftOrder, orderItems); err != nil {
		return nil, err
	}

	if draftOrder.PaymentMethod == models.PaymentMethodCOD {
		order, err := o.DraftOrderToOrder(ctx, &draftOrder, orderItems)
		if err != nil {
			return nil, err
		}
		return &order, nil

	} else {
		order, err := o.DraftOrderToOrderResponse(ctx, &draftOrder, orderItems)
		if err != nil {
			return nil, err
		}
		return &order, nil
	}
}
func (o *orderService) CreatePayment(ctx context.Context, draftOrder *models.DraftOrder, orderItems []models.OrderItem) error {

	paymentLink := ""
	if draftOrder.PaymentMethod == models.PaymentMethodBank {
		// Create PayOS payment link
		paymentData, err := CreatePayOSPayment(int(draftOrder.ID), 10000, MapOrderItemsToPayOSItems(orderItems, int(draftOrder.ShippingFee)), "Thanh toán đơn hàng", "http://localhost:5173/success", "http://localhost:5173/cancel")
		if err != nil {
			log.Println("CreatePayment error", err.Error())
			return customErr.NewError(customErr.INTERNAL_ERROR, "CreatePayment error", http.StatusInternalServerError, err)
		}
		paymentLink = paymentData.CheckoutUrl

		// Publish event after create payOS link
		paymentEvent := event.PayOSPaymentCreatedEvent{
			DraftOrderID:  draftOrder.ID,
			PaymentLink:   paymentData.CheckoutUrl,
			Amount:        float64(10000),
			PaymentMethod: string(draftOrder.PaymentMethod),
			CreatedAt:     time.Now(),
		}

		if err := o.eventBus.PublishPaymentCreated(paymentEvent); err != nil {
			log.Printf("Failed to publish payment created event: %v", err)
		}

	}
	var paymentInfo = &models.PaymentInfo{
		Amount:      draftOrder.Total,
		Status:      models.PaymentPending,
		PaymentLink: paymentLink,
	}
	if err := o.paymentInfoRepo.Create(ctx, paymentInfo); err != nil {
		log.Printf(err.Error(), "while creating payment info")
		return customErr.NewError(customErr.INTERNAL_ERROR, "CreatePayment error", http.StatusInternalServerError, err)
	}

	draftOrder.PaymentInfo = paymentInfo
	if err := o.draftOrderRepo.Save(ctx, draftOrder); err != nil {
		log.Printf(err.Error(), "while saving draftOrder")
		return err
	}

	return nil
}

func (o *orderService) DraftOrderToOrder(ctx context.Context, draftOrder *models.DraftOrder, orderItems []models.OrderItem) (models.Order, error) {
	order := models.Order{
		UserID:          draftOrder.UserID,
		Status:          draftOrder.Status,
		Total:           draftOrder.Total,
		PaymentMethod:   draftOrder.PaymentMethod,
		ShippingAddress: draftOrder.ShippingAddress,
		ShippingFee:     draftOrder.ShippingFee,
		PhoneNumber:     draftOrder.PhoneNumber,
		PaymentInfoID:   draftOrder.PaymentInfoID,
		PaymentInfo:     draftOrder.PaymentInfo,
	}
	if err := o.orderRepo.Create(ctx, &order); err != nil {
		return order, err
	}

	orderItemsNew := []models.OrderItem{}
	for _, oi := range orderItems {
		oi.OrderID = order.ID
		oi.OrderType = models.OrderTypeOrder
		if err := o.orderItemRepo.Save(ctx, &oi); err != nil {
			return order, err
		}
		orderItemsNew = append(orderItemsNew, oi)
	}
	order.OrderItems = orderItemsNew

	draftOrder.ToOrderID = &order.ID
	if err := o.draftOrderRepo.Save(ctx, draftOrder); err != nil {
		return order, err
	}
	return order, nil
}

func (o *orderService) DraftOrderToOrderResponse(ctx context.Context, draftOrder *models.DraftOrder, orderItems []models.OrderItem) (models.Order, error) {
	order := models.Order{
		ID:              draftOrder.ID,
		UserID:          draftOrder.UserID,
		Status:          draftOrder.Status,
		Total:           draftOrder.Total,
		PaymentMethod:   draftOrder.PaymentMethod,
		ShippingAddress: draftOrder.ShippingAddress,
		ShippingFee:     draftOrder.ShippingFee,
		PhoneNumber:     draftOrder.PhoneNumber,
		PaymentInfoID:   draftOrder.PaymentInfoID,
		PaymentInfo:     draftOrder.PaymentInfo,
		OrderItems:      orderItems,
	}
	return order, nil
}

func (o *orderService) PayOSPaymentSuccess(ctx context.Context, draftOrderID uint) {
	draftOrder, err := o.draftOrderRepo.GetById(ctx, draftOrderID)
	if err != nil {
		log.Printf(err.Error(), "while getting draftOrder to update PayOSPayment")
		return
	}
	if nextStatus, ok := utils.CanTransitionPayment(draftOrder.PaymentInfo.Status, utils.EventPaySuccess); ok {
		draftOrder.PaymentInfo.Status = nextStatus
		err = o.paymentInfoRepo.Save(ctx, draftOrder.PaymentInfo)
		if err != nil {
			log.Printf(err.Error(), "while saving payment info")
		}
	} else {
		log.Printf(err.Error(), "while transitioning payment info from PayOSPayment")
	}
	_, err = o.DraftOrderToOrder(ctx, draftOrder, draftOrder.OrderItems)
	if err != nil {
		log.Printf(err.Error(), "while create Order to update PayOSPayment")
		return
	}
}

func (o *orderService) GetById(ctx context.Context, orderID uint, userID uint) (*models.Order, error) {
	return o.orderRepo.GetByIdAndUserId(ctx, orderID, userID)
}

func MapOrderItemsToPayOSItems(orderItem []models.OrderItem, shippingFee int) []payos.Item {
	var items []payos.Item

	for _, oi := range orderItem {
		item := payos.Item{
			Name:     fmt.Sprintf("%s (Variant #%d)", oi.Variant.Product.Name, oi.ProductVariantID),
			Price:    int(oi.Price),
			Quantity: int(oi.Quantity),
		}
		items = append(items, item)
	}

	items = append(items, payos.Item{
		Name:     "Shipping Fee",
		Price:    shippingFee,
		Quantity: 1,
	})
	return items
}

func (o *orderService) registerEventHandlers() {
	o.eventBus.Subscribe(func(e event.PayOSPaymentCreatedEvent) {
		log.Printf("Payment created for order %d: %s", e.DraftOrderID, e.PaymentLink)
		var data *payos.PaymentLinkDataType
		for {
			var err error
			data, err = payos.GetPaymentLinkInformation(strconv.FormatUint(uint64(e.DraftOrderID), 10))
			if err != nil {
				log.Printf("Error getting payment info: %v", err)
				return
			}

			log.Println("Current status:", data.Status)

			if data.Status != "PENDING" {
				break
			}

			time.Sleep(10 * time.Second)
		}

		if data.Status == "SUCCESS" {
			o.PayOSPaymentSuccess(context.Background(), e.DraftOrderID)
		}
		log.Printf("Payment status updated and no longer pending for order %d", e.DraftOrderID)
		o.PayOSPaymentCancelled(context.Background(), e.DraftOrderID, data.Status)
	})
}

func (o *orderService) PayOSPaymentCancelled(ctx context.Context, draftId uint, status string) {
	draftOrder, err := o.draftOrderRepo.GetById(ctx, draftId)
	if err != nil {
		log.Printf(err.Error(), "while getting draftOrder to update PayOSPayment")
		return
	}
	val := uint(0)

	draftOrder.ToOrderID = &val
	if err := o.draftOrderRepo.Save(ctx, draftOrder); err != nil {
		log.Printf(err.Error(), "while saving draft order to update PayOSPayment Fail")
	}
	orderItems := draftOrder.OrderItems

	err = o.productVariantCache.IncrementStock(orderItems)
	if err != nil {
		log.Printf(err.Error(), "while incrementing stock variant after cancelled payment")
	}

	if nextStatus, ok := utils.CanTransitionPayment(draftOrder.PaymentInfo.Status, utils.EventPayCancel); ok {
		draftOrder.PaymentInfo.Status = nextStatus
		err = o.paymentInfoRepo.Save(ctx, draftOrder.PaymentInfo)
		if err != nil {
			log.Printf(err.Error(), "while saving payment info")
		}
	} else {
		log.Printf(err.Error(), "while transitioning payment info from PayOSPayment")
	}
}

func (o *orderService) UpdateQuantity(ctx context.Context) error {
	draftOrders, err := o.draftOrderRepo.GetsForDbUpdate(ctx)
	if err != nil {
		return err
	}
	totalQuantityByVariant := make(map[uint]uint)

	for _, d := range draftOrders {
		log.Printf("Draft Order ID: %d, Order ID: %d", d.ID, d.ToOrderID)
		order, err := o.orderRepo.GetById(ctx, *d.ToOrderID)
		if err != nil {
			return err
		}
		orderItem := order.OrderItems
		for _, oi := range orderItem {
			totalQuantityByVariant[oi.ProductVariantID] += oi.Quantity
		}
	}
	log.Printf("Total quantity of variants: %d", totalQuantityByVariant)

	err = o.productVariantRepo.UpdateQuantity(ctx, totalQuantityByVariant)
	if err != nil {
		return err
	}

	for _, d := range draftOrders {
		err := o.draftOrderRepo.Delete(ctx, d.ID)
		if err != nil {
			return err
		}
		log.Printf("Deleted draft order %d", d.ID)
	}

	log.Printf("Update Db successfully")

	var keys []uint
	for k := range totalQuantityByVariant {
		keys = append(keys, k)
	}

	// Reset redis cache
	variants, err := o.productVariantRepo.GetByIDSForRedisCache(ctx, keys)
	if err != nil {
		return err
	}

	for _, pv := range variants {
		// save as hash: id, quantity, price
		if err := o.productVariantCache.SaveProductVariantHash(pv); err != nil {
			log.Printf("warning: save productVariant %d to redis failed: %v", pv.ID, err)
		}
	}
	log.Printf("Reset Redis Cache successfully")

	//Clean draft
	err = o.CleanDraft(ctx)
	if err != nil {
		return err
	}
	log.Printf("Clean draft orders")
	return nil
}

func (o *orderService) CleanDraft(ctx context.Context) error {
	err := o.draftOrderRepo.CleanDraft(ctx)
	if err != nil {
		return err
	}
	return nil
}
