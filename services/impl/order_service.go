package impl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/cache"
	"github.com/minh6824pro/nxrGO/configs"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/event"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
	"github.com/minh6824pro/nxrGO/utils"
	"github.com/payOSHQ/payos-lib-golang"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type orderService struct {
	db                  *gorm.DB
	productVariantRepo  repositories.ProductVariantRepository
	orderItemRepo       repositories.OrderItemRepository
	orderRepo           repositories.OrderRepository
	draftOrderRepo      repositories.DraftOrderRepository
	merchantRepo        repositories.MerchantRepository
	paymentInfoRepo     repositories.PaymentInfoRepository
	productVariantCache cache.ProductVariantRedis
	eventBus            event.EventPublisher
	updateStockAgg      *event.UpdateStockAggregator
}

func NewOrderService(db *gorm.DB, productVariantRepo repositories.ProductVariantRepository, orderItemRepo repositories.OrderItemRepository,
	orderRepo repositories.OrderRepository, merchantRepo repositories.MerchantRepository, draftOrderRepo repositories.DraftOrderRepository,
	paymentInfoRepo repositories.PaymentInfoRepository,
	productVariantCache cache.ProductVariantRedis,
	eventBus event.EventPublisher, updateStockAgg *event.UpdateStockAggregator) services.OrderService {
	service := &orderService{
		db:                  db,
		productVariantRepo:  productVariantRepo,
		orderRepo:           orderRepo,
		orderItemRepo:       orderItemRepo,
		draftOrderRepo:      draftOrderRepo,
		merchantRepo:        merchantRepo,
		paymentInfoRepo:     paymentInfoRepo,
		productVariantCache: productVariantCache,
		eventBus:            eventBus,
		updateStockAgg:      updateStockAgg,
	}
	service.registerEventHandlers()

	return service
}

func (o *orderService) Create(ctx context.Context, input dto.CreateOrderInput) (*dto.CreateOrderResponse, error) {
	// Validate info with signature
	var totalPrice float64
	for _, oi := range input.OrderItems {
		if !utils.ValidateProductVariantSignature(oi.ProductVariantID, oi.Price, oi.MerchantID, oi.Timestamp, oi.Signature) {
			return nil, customErr.NewError(customErr.BAD_REQUEST, "Product information invalid", http.StatusBadRequest, nil)
		}
		totalPrice += oi.Price * float64(oi.Quantity)

	}
	if totalPrice != input.Total-input.ShippingFee {
		return nil, customErr.NewError(customErr.INVALID_PRICE, "Invalid total price", http.StatusBadRequest, nil)
	}

	// Validate shipping fee
	var totalShippingFee float64
	for _, shipping := range input.ShippingFeeInput {
		if !utils.ValidateShippingFeeSignature(shipping.MerchantID, shipping.DeliveryID, shipping.Fee, input.Latitude, input.Longitude, shipping.Timestamp, shipping.Signature) {
			log.Println("sig ", shipping.Signature)
			return nil, customErr.NewError(customErr.INVALID_PRICE, "Shipping Fee invalid", http.StatusBadRequest, nil)
		}
		totalShippingFee += shipping.Fee
	}
	if totalShippingFee != input.ShippingFee {
		return nil, customErr.NewError(customErr.INVALID_PRICE, "Invalid total shippingFee", http.StatusBadRequest, nil)
	}
	// Begin check quantity
	var orderItems []models.OrderItem
	var draftOrder models.DraftOrder

	// Check redis available
	err := o.productVariantCache.PingRedis(ctx)
	if err != nil {
		// Process with DB
		log.Printf("Create with Db")
		draftOrder, orderItems, err = o.CreateOrderWithDb(ctx, input)
		if err != nil {
			return nil, err
		}

	} else {
		// Process with Redis
		var total float64 = 0

		luaScript := `
local itemCount = tonumber(ARGV[1])
local idx = 2

-- 1) Check for MISS (missing keys)
local missed = {}
for i = 1, itemCount do
    local variantId = ARGV[idx]
    idx = idx + 2
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

-- 2) Check stock only
idx = 2
local updates = {}
for i = 1, itemCount do
    local variantId = ARGV[idx]
    local qtyRequested = tonumber(ARGV[idx + 1])
    idx = idx + 2
    local key = KEYS[i]

    local stock = tonumber(redis.call("HGET", key, "quantity"))

    if stock == nil then
        return {"MISS", variantId}
    end
    if qtyRequested > stock then
        return {"INSUFFICIENT", variantId}
    end

    table.insert(updates, { key = key, qty = qtyRequested })
end

-- 3) Deduct stock if all checks pass
for i = 1, #updates do
    redis.call("HINCRBY", updates[i].key, "quantity", -updates[i].qty)
end

return {"OK"}
`

		// build keys & args
		keys := make([]string, 0, len(input.OrderItems))
		args := []interface{}{len(input.OrderItems)}
		for _, oi := range input.OrderItems {
			keys = append(keys, fmt.Sprintf(cache.ProductVariantKeyPattern, oi.ProductVariantID))
			// ARGV: variantId, quantity
			args = append(args, oi.ProductVariantID, oi.Quantity)
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
				// Get list of missing variantIds
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

				_, err := o.loadAndCacheProductVariants(ctx, missingIDs)
				if err != nil {
					return nil, err
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

		// Remove landing page cache
		for _, oi := range input.OrderItems {
			oi := oi
			go func() {
				o.productVariantCache.DeleteMiniProduct(oi.ProductVariantID)
			}()
		}

		// Create Draft Order
		draftOrder = models.DraftOrder{
			UserID:          input.UserID,
			Status:          models.OrderStatePending,
			ShippingAddress: input.ShippingAddress,
			PaymentMethod:   input.PaymentMethod,
			PhoneNumber:     input.PhoneNumber,
			DeliveryMode:    input.ShippingFeeInput[0].Mode,
			Latitude:        input.Latitude,
			Longitude:       input.Longitude,
		}
		if err := o.draftOrderRepo.Create(ctx, &draftOrder); err != nil {
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, fmt.Sprintf("Draft order creation error: %v", err.Error()), http.StatusBadRequest, nil)
		}

		// Create order items
		for _, item := range input.OrderItems {
			orderItem := models.OrderItem{
				OrderID:          draftOrder.ID,
				ProductVariantID: item.ProductVariantID,
				OrderType:        models.OrderTypeDraftOrder,
				Quantity:         item.Quantity,
				Price:            item.Price,
				TotalPrice:       item.Price * float64(item.Quantity),
				MerchantID:       item.MerchantID,
			}
			if err := o.orderItemRepo.Create(ctx, &orderItem); err != nil {
				return nil, err
			}
			orderItems = append(orderItems, orderItem)
		}
		// Create Delivery detail
		newDeliveryDetail := models.DeliveryDetail{
			OrderID:    draftOrder.ID,
			OrderType:  models.OrderTypeDraftOrder,
			DeliveryID: input.ShippingFeeInput[0].DeliveryID,
		}
		if err := o.db.Create(&newDeliveryDetail).Error; err != nil {
			return nil, err
		}
		draftOrder.Delivery = newDeliveryDetail
	}

	// Create PaymentInfo
	if err := o.CreatePayment(ctx, &draftOrder, orderItems, input.Total, input.ShippingFee); err != nil {
		return nil, err
	}

	// Check if split order needed

	// Create map for unique merchant id
	uniqueMerchants := make(map[uint]struct{})

	for _, oi := range input.OrderItems {
		uniqueMerchants[oi.MerchantID] = struct{}{}
	}

	// To slice
	merchantIDs := make([]uint, 0, len(uniqueMerchants))
	for id := range uniqueMerchants {
		merchantIDs = append(merchantIDs, id)
	}
	if len(merchantIDs) > 1 {
		// Parse response
		response, err := o.MapDraftOrderToCreateOrderResponse(ctx, &draftOrder)
		if err != nil {
			return nil, err
		}
		if draftOrder.PaymentMethod == models.PaymentMethodCOD {
			// Split order if COD
			subDraftOrders, err := o.SplitOrder(ctx, &draftOrder, orderItems, merchantIDs, input.ShippingFeeInput)
			if err != nil {
				return nil, err
			}
			//Convert to order
			go func() {

				_, err := o.DraftsOrderToOrder(context.Background(), subDraftOrders)
				if err != nil {
					log.Println("Error in draft order to order:", err)
					return
				}

			}()
		} else {
			// Mark need to split after payment success
			temp := uint(0)
			draftOrder.ParentID = &temp
			if err = o.draftOrderRepo.Save(ctx, &draftOrder); err != nil {
				log.Println("Error in draft order to order:", err)
			}

		}
		return response, nil

	} else {
		// If not split

		// Convert into order if payment method = COD
		if draftOrder.PaymentMethod == models.PaymentMethodCOD {
			order, err := o.DraftOrderToOrder(ctx, &draftOrder, orderItems)
			if err != nil {
				return nil, err
			}
			return o.MapOrderToCreateOrderResponse(ctx, &order)

		} else {
			// Parse response
			draftOrder.OrderItems = orderItems
			return o.MapDraftOrderToCreateOrderResponse(ctx, &draftOrder)
		}
	}

}
func (o *orderService) CreatePayment(ctx context.Context, draftOrder *models.DraftOrder, orderItems []models.OrderItem, total float64, shippingFee float64) error {

	var paymentInfo = &models.PaymentInfo{
		ID:          GeneratePaymentInfoID(),
		Total:       total,
		ShippingFee: shippingFee,
		OrderID:     draftOrder.ID,
		OrderType:   models.OrderTypeDraftOrder,
		Status:      models.PaymentPending,
	}
	if err := o.paymentInfoRepo.Create(ctx, paymentInfo); err != nil {
		log.Printf(err.Error(), "while creating payment info")
		return customErr.NewError(customErr.INTERNAL_ERROR, "CreatePayment error", http.StatusInternalServerError, err)
	}
	paymentLink := ""
	if draftOrder.PaymentMethod == models.PaymentMethodBank {
		// Create PayOS payment link
		paymentData, err := CreatePayOSPayment(paymentInfo.ID, 10000, MapOrderItemsToPayOSItems(orderItems, int(shippingFee)), fmt.Sprintf("Thanh toán đơn hàng %d", draftOrder.ID), "http://localhost:5173/success", "http://localhost:5173/cancel")
		if err != nil {
			log.Println("CreatePayment error", err.Error())
			log.Print(paymentInfo.ID)
			return customErr.NewError(customErr.INTERNAL_ERROR, "CreatePayment error", http.StatusInternalServerError, err)
		}
		paymentLink = paymentData.CheckoutUrl

		// Publish event after create payOS link
		paymentEvent := event.PayOSPaymentCreatedEvent{
			Id:            paymentInfo.ID,
			OrderID:       draftOrder.ID,
			PaymentLink:   paymentData.CheckoutUrl,
			Total:         10000,
			PaymentMethod: string(draftOrder.PaymentMethod),
			CreatedAt:     time.Now(),
		}

		if err := o.eventBus.PublishPaymentCreated(paymentEvent); err != nil {
			log.Printf("Failed to publish payment created event: %v", err)
		}
		log.Printf("PayOs payment created for order %d: %s", draftOrder.ID, paymentData.CheckoutUrl)

	}

	paymentInfo.PaymentLink = paymentLink
	if err := o.paymentInfoRepo.Save(ctx, paymentInfo); err != nil {
		log.Printf(err.Error(), "while saving payment info")
		return customErr.NewError(customErr.INTERNAL_ERROR, "Save Payment error", http.StatusInternalServerError, err)
	}
	draftOrder.PaymentInfos = append(draftOrder.PaymentInfos, *paymentInfo)
	if err := o.draftOrderRepo.Save(ctx, draftOrder); err != nil {
		log.Printf(err.Error(), "while saving draftOrder")
		return err
	}
	return nil
}

// No split order
func (o *orderService) DraftOrderToOrder(ctx context.Context, draftOrder *models.DraftOrder, orderItems []models.OrderItem) (models.Order, error) {
	order := models.Order{
		UserID:          draftOrder.UserID,
		Status:          draftOrder.Status,
		PaymentMethod:   draftOrder.PaymentMethod,
		ShippingAddress: draftOrder.ShippingAddress,
		PhoneNumber:     draftOrder.PhoneNumber,
		OrderItems:      orderItems,
		PaymentInfos:    draftOrder.PaymentInfos,
		ParentID:        draftOrder.ParentID,
		DeliveryMode:    draftOrder.DeliveryMode,
		Delivery:        draftOrder.Delivery,
		Latitude:        draftOrder.Latitude,
		Longitude:       draftOrder.Longitude,
	}
	if err := o.orderRepo.Create(ctx, &order); err != nil {
		return order, err
	}
	log.Print(order.OrderItems)

	draftOrder.ToOrderID = &order.ID
	draftOrder.PaymentInfos = nil
	draftOrder.OrderItems = nil
	draftOrder.Delivery = models.DeliveryDetail{}
	if err := o.draftOrderRepo.Save(ctx, draftOrder); err != nil {
		return order, err
	}
	return order, nil
}

func (o *orderService) DraftsOrderToOrder(ctx context.Context, draftOrder []*models.DraftOrder) ([]models.Order, error) {
	var orders []models.Order
	// Create parent order
	i := len(draftOrder) - 1
	order := models.Order{
		UserID:          draftOrder[i].UserID,
		Status:          draftOrder[i].Status,
		PaymentMethod:   draftOrder[i].PaymentMethod,
		ShippingAddress: draftOrder[i].ShippingAddress,
		PhoneNumber:     draftOrder[i].PhoneNumber,
		OrderItems:      nil,
		PaymentInfos:    draftOrder[i].PaymentInfos,
		ParentID:        draftOrder[i].ParentID,
		DeliveryMode:    draftOrder[i].DeliveryMode,
		Delivery:        draftOrder[i].Delivery,
		Latitude:        draftOrder[i].Latitude,
		Longitude:       draftOrder[i].Longitude,
	}
	if err := o.orderRepo.Create(ctx, &order); err != nil {
		return nil, err
	}
	orders = append(orders, order)
	draftOrder[i].ToOrderID = &order.ID
	draftOrder[i].PaymentInfos = nil
	draftOrder[i].OrderItems = nil
	draftOrder[i].Delivery = models.DeliveryDetail{}
	if err := o.draftOrderRepo.Save(ctx, draftOrder[i]); err != nil {
	}
	// create sub order
	for i = 0; i < len(draftOrder)-1; i++ {
		subOrder := models.Order{
			UserID:          draftOrder[i].UserID,
			Status:          draftOrder[i].Status,
			PaymentMethod:   draftOrder[i].PaymentMethod,
			ShippingAddress: draftOrder[i].ShippingAddress,
			PhoneNumber:     draftOrder[i].PhoneNumber,
			OrderItems:      draftOrder[i].OrderItems,
			PaymentInfos:    draftOrder[i].PaymentInfos,
			ParentID:        &order.ID,
			DeliveryMode:    draftOrder[i].DeliveryMode,
			Delivery:        draftOrder[i].Delivery,
			Latitude:        draftOrder[i].Latitude,
			Longitude:       draftOrder[i].Longitude,
		}
		if err := o.orderRepo.Create(ctx, &subOrder); err != nil {
			log.Println("Create sub order error")
		}
		orders = append(orders, subOrder)
		draftOrder[i].ToOrderID = &subOrder.ID
		draftOrder[i].PaymentInfos = nil
		draftOrder[i].OrderItems = nil
		draftOrder[i].Delivery = models.DeliveryDetail{}
		err := o.draftOrderRepo.Save(ctx, draftOrder[i])
		if err != nil {
			return nil, err
		}
	}
	return orders, nil
}

func (o *orderService) DraftOrderToOrderResponse(ctx context.Context, draftOrder *models.DraftOrder, orderItems []models.OrderItem) (models.Order, error) {
	order := models.Order{
		ID:              draftOrder.ID,
		UserID:          draftOrder.UserID,
		Status:          draftOrder.Status,
		PaymentMethod:   draftOrder.PaymentMethod,
		ShippingAddress: draftOrder.ShippingAddress,
		PhoneNumber:     draftOrder.PhoneNumber,
		PaymentInfos:    draftOrder.PaymentInfos,
		OrderItems:      orderItems,
	}
	return order, nil
}

// TODO IMPLEMENT SPLIT ORDER
func (o *orderService) PayOSPaymentSuccess(ctx context.Context, paymentInfoID int64) {
	paymentInfo, err := o.paymentInfoRepo.GetByID(ctx, paymentInfoID)
	if err != nil {
		log.Printf(err.Error(), "while getting draftOrder to update PayOSPayment")
		return
	}
	// If draft -> convert to order
	if paymentInfo.OrderType == models.OrderTypeDraftOrder {

		if nextStatus, ok := utils.CanTransitionPayment(paymentInfo.Status, utils.EventPaySuccess); ok {
			paymentInfo.Status = nextStatus
			err = o.paymentInfoRepo.Save(ctx, paymentInfo)
			if err != nil {
				log.Printf(err.Error(), "while saving payment info")
			}
		} else {
			log.Printf("error 1 while transitioning payment info from PayOSPayment payment id: %d", paymentInfo.ID)
		}
		draftOrder, err := o.draftOrderRepo.GetById(ctx, paymentInfo.OrderID)
		if err != nil {
			log.Printf(err.Error(), "while getting draftOrder")
			return
		}
		// Check if order need to split
		if draftOrder.ParentID != nil && *draftOrder.ParentID == 0 {
			log.Println("NEED SPLIT")
			//split
			infos, err2 := o.draftOrderRepo.GetForSplit(ctx, draftOrder.ID)
			if err2 != nil {
				log.Println(err2, " While split order (bank payment success)")
				return
			}
			//Get merchant id distinct
			var merchantIDs []uint
			merchantIDMap := make(map[uint]bool)
			itemAndMerchantMap := make(map[uint]uint) // key: OrderItem ID, value: MerchantID

			for _, info := range infos {
				if !merchantIDMap[info.MerchantID] {
					merchantIDMap[info.MerchantID] = true
					merchantIDs = append(merchantIDs, info.MerchantID)
				}
				itemAndMerchantMap[info.ID] = info.MerchantID
			}
			//Inject merchant ID for orderItems
			for i := range draftOrder.OrderItems {
				draftOrder.OrderItems[i].MerchantID = itemAndMerchantMap[draftOrder.OrderItems[i].ID]
			}
			log.Println(merchantIDs)
			var shippingFeeResponse []dto.ShippingFeeResponse
			for _, merchantID := range merchantIDs {
				fee, err := o.CalculateShippingFee(ctx, merchantID, draftOrder.Longitude, draftOrder.Latitude, infos[0].DeliveryID)
				if err != nil {
					return
				}
				shippingFeeResponse = append(shippingFeeResponse, fee...)

			}
			for _, oi := range draftOrder.OrderItems {
				log.Println(oi.MerchantID, "merchant")
			}
			draftsSplit, err := o.SplitOrder(ctx, draftOrder, draftOrder.OrderItems, merchantIDs, shippingFeeResponse)
			if err != nil {
				log.Printf(err.Error(), "while split order (bank payment success)")
				return
			}

			_, err = o.DraftsOrderToOrder(ctx, draftsSplit)
			if err != nil {
				log.Printf(err.Error(), "while convert drafts order to order(bank payment success)")
				return
			}

		} else {
			// not split
			_, err = o.DraftOrderToOrder(ctx, draftOrder, draftOrder.OrderItems)
			if err != nil {
				log.Printf(err.Error(), "while create Order to update PayOSPayment")
				return
			}
		}
	} else {
		// if order update
		if nextStatus, ok := utils.CanTransitionPayment(paymentInfo.Status, utils.EventPaySuccess); ok {
			paymentInfo.Status = nextStatus
			err = o.paymentInfoRepo.Save(ctx, paymentInfo)
			if err != nil {
				log.Printf(err.Error(), "while saving payment info")
			}
		} else {
			log.Printf("error 1 while transitioning payment info from PayOSPayment payment id: %d", paymentInfo.ID)
		}
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
		log.Printf("Tracking payment created for order %d: %s", e.Id, e.PaymentLink)
		var data *payos.PaymentLinkDataType
		for {
			var err error
			data, err = payos.GetPaymentLinkInformation(strconv.FormatInt(e.Id, 10))
			if err != nil {
				log.Printf("Error getting payment info: %v", err)
				return
			}

			if data.Status != "PENDING" {
				break
			}

			time.Sleep(10 * time.Second)
		}

		if data.Status == "PAID" {
			o.PayOSPaymentSuccess(context.Background(), e.Id)
		} else {
			reasonStr := "Cancelled/Expired via payos"
			if data.CancellationReason != nil {
				reasonStr = *data.CancellationReason
			}
			o.PayOSPaymentCancelled(context.Background(), e.Id, data.Status, reasonStr)
		}
		log.Printf("Payment status updated and no longer pending for order %d", e.Id)
	})
}

func (o *orderService) PayOSPaymentCancelled(ctx context.Context, paymentInfoId int64, status string, reason string) {
	paymentInfo, err := o.paymentInfoRepo.GetByID(ctx, paymentInfoId)
	if err != nil {
		log.Printf(err.Error(), "while getting draftOrder to update PayOSPayment")
		return
	}
	val := uint(0)

	// TODO IMPLEMENT IF ordertype is draft order
	if paymentInfo.OrderType == models.OrderTypeDraftOrder {
		draftOrder, err := o.draftOrderRepo.GetById(ctx, paymentInfo.OrderID)
		if err != nil {
			log.Printf(err.Error(), "while getting draftOrder")
			return
		}
		if draftOrder.PaymentInfos[0].ID == paymentInfoId && draftOrder.ToOrderID == nil {
			// Latest payment => cancel order
			log.Println("la payment moi nhat nen xu ly cancelled")
			draftOrder.ToOrderID = &val
			if err := o.draftOrderRepo.Save(ctx, draftOrder); err != nil {
				log.Printf(err.Error(), "while saving draft order to update PayOSPayment Fail")
			}
			orderItems := draftOrder.OrderItems

			err = o.productVariantCache.IncrementStock(orderItems)
			if err != nil {
				log.Printf(err.Error(), "while incrementing stock variant after cancelled payment")
			}

		}

		// Not latest => update paymentinfo
		if nextStatus, ok := utils.CanTransitionPayment(paymentInfo.Status, utils.EventPayCancel); ok {
			paymentInfo.Status = nextStatus
			paymentInfo.CancellationReason = reason
			now := time.Now()
			paymentInfo.CancellationAt = &now
			err = o.paymentInfoRepo.Save(ctx, paymentInfo)
			if err != nil {
				log.Printf(err.Error(), "while saving payment info")
			}
		} else {
			log.Printf("error 2 while transitioning payment info from PayOSPayment payment id: %d", paymentInfo.ID)
		}
	} else {
		// if is order
		order, err2 := o.orderRepo.GetById(ctx, paymentInfo.OrderID)
		if err2 != nil {
			log.Printf(err.Error(), "while getting Order to update PayOSPayment")
			return
		}
		if order.PaymentInfos[0].ID == paymentInfoId {
			// Latest payment -> cancel payment & order
			if nextStatus, ok := utils.CanTransitionPayment(paymentInfo.Status, utils.EventPayCancel); ok {
				paymentInfo.Status = nextStatus
				paymentInfo.CancellationReason = reason
				now := time.Now()
				paymentInfo.CancellationAt = &now
				err = o.paymentInfoRepo.Save(ctx, paymentInfo)

				if nextOrderStatus, ok := utils.CanTransitionOrder(order.Status, utils.EventCancel); ok == nil {
					order.Status = nextOrderStatus
					err = o.orderRepo.Save(ctx, order)
				}
				// publish cancel event -> add stock
				o.updateStockAgg.AddOrder(*order)
				if err != nil {
					log.Printf(err.Error(), "while saving payment info")
				}
			} else {
				log.Printf("error 2 while transitioning payment info from PayOSPayment payment id: %d", paymentInfo.ID)
			}
		} else {
			// Not latest payment info -> cancel payment
			if nextStatus, ok := utils.CanTransitionPayment(paymentInfo.Status, utils.EventPayCancel); ok {
				paymentInfo.Status = nextStatus
				paymentInfo.CancellationReason = reason
				now := time.Now()
				paymentInfo.CancellationAt = &now
				err = o.paymentInfoRepo.Save(ctx, paymentInfo)

			} else {
				log.Printf("error 2 while transitioning payment info from PayOSPayment payment id: %d", paymentInfo.ID)
			}

		}
	}
}

func (o *orderService) UpdateQuantity(ctx context.Context) error {
	o.updateStocks()
	// Get draft order that are converted to  order
	draftOrders, err := o.draftOrderRepo.GetsForDbUpdate(ctx)
	if err != nil {
		return err
	}
	totalQuantityByVariant := make(map[uint]uint)

	// Sum quantity for each key: Product Variant id
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

	err = o.productVariantRepo.DecreaseQuantity(ctx, totalQuantityByVariant)
	if err != nil {
		return err
	}

	// Delete draft order that are converted to order
	for _, d := range draftOrders {
		err := o.draftOrderRepo.Delete(ctx, d.ID)
		if err != nil {
			return err
		}
		log.Printf("Deleted draft order %d", d.ID)
	}

	log.Printf("UpdateOrderStatus Db successfully")

	// Delete redis cache
	for key, _ := range totalQuantityByVariant {
		err := o.productVariantCache.DeleteProductVariantHash(key)
		if err != nil {
			log.Print(err)
		}
	}
	log.Println("map delete redis", totalQuantityByVariant)
	log.Printf("Delete redis Cache successfully")

	//Clean draft order that can't be converted to order
	err = o.CleanDraft(ctx)
	if err != nil {
		return err
	}
	log.Printf("Cleaned draft orders")
	return nil
}

func (o *orderService) CleanDraft(ctx context.Context) error {
	err := o.draftOrderRepo.CleanDraft(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (o *orderService) updateStocks() {
	data := o.updateStockAgg.Flush()
	for key, value := range data {
		err := o.db.Model(&models.ProductVariant{}).
			Where("id = ?", key).
			UpdateColumn("quantity", gorm.Expr("quantity + ?", value)).
			Error
		if err != nil {
			log.Printf("Error updating product variant: %d: %v", key, err)
		}
		err = o.productVariantCache.DeleteProductVariantHash(key)
		if err != nil {
			log.Printf("Error deleting product variant cache: %v", err)
		}
	}
	log.Print("Update stocks successfully")
}

func (o *orderService) GetsByStatus(ctx context.Context, status models.OrderStatus, userId uint) ([]*models.Order, error) {

	return o.orderRepo.GetsByStatusAndUserId(ctx, status, userId)
}

func (o *orderService) UpdateOrderStatus(ctx context.Context, orderId uint, event utils.OrderEvent) (*models.Order, error) {
	order, err := o.orderRepo.GetById(ctx, orderId)
	if err != nil {
		return nil, err
	}
	if nextStatus, err := utils.CanTransitionOrder(order.Status, event); err != nil {
		log.Printf(err.Error())
		return nil, err
	} else {

		// Increase stock if cancel before ship or  after completing return_shipping
		if (nextStatus == models.OrderStateCancelled && models.IsBeforeOrderStatus(order.Status, models.OrderStateProcessing)) ||
			nextStatus == models.OrderStateReturned {
			log.Printf("Order %d is already processing cancel/return", order.ID)
			o.updateStockAgg.AddOrder(*order)
		}
		// Update Status
		order.Status = nextStatus
		err = o.orderRepo.Update(ctx, order)
		if err != nil {
			log.Printf(err.Error(), "while saving order")
			return nil, err
		}

		return order, nil
	}
}

func (o *orderService) ReBuy(ctx context.Context, orderId uint, userId uint) (*dto.CreateOrderResponse, error) {

	order, err := o.orderRepo.GetByIdAndUserId(ctx, orderId, userId)
	if err != nil {
		return nil, err
	}

	create, err := o.Create(ctx, OrderToCreateOrderInput(order))
	if err != nil {
		return nil, err
	}

	return create, nil
}

// TODO REVIEW
func OrderToCreateOrderInput(order *models.Order) dto.CreateOrderInput {
	var items []dto.CreateOrderItem

	for _, detail := range order.OrderItems {
		item := dto.CreateOrderItem{
			ProductVariantID: detail.ProductVariantID,
			Quantity:         detail.Quantity,
			Price:            detail.Price,
		}
		items = append(items, item)
	}

	return dto.CreateOrderInput{
		UserID:          order.UserID,
		Total:           order.PaymentInfos[0].Total,
		PaymentMethod:   order.PaymentMethod,
		ShippingAddress: order.ShippingAddress,
		ShippingFee:     order.PaymentInfos[0].ShippingFee,
		PhoneNumber:     order.PhoneNumber,
		OrderItems:      items,
	}
}

func (o *orderService) MapOrderToCreateOrderResponse(ctx context.Context, order *models.Order) (*dto.CreateOrderResponse, error) {
	var orderItems []dto.OrderItemResponse

	for _, item := range order.OrderItems {
		orderItems = append(orderItems, dto.OrderItemResponse{
			ID:               item.ID,
			OrderID:          item.OrderID,
			OrderType:        item.OrderType,
			ProductVariantID: item.ProductVariantID,
			Quantity:         item.Quantity,
			Price:            item.Price,
			TotalPrice:       item.TotalPrice,
		})
	}

	var paymentInfo dto.PaymentInfoResponse
	if len(order.PaymentInfos) > 0 {
		paymentInfo = dto.PaymentInfoResponse{
			ID:                 order.PaymentInfos[0].ID,
			Total:              order.PaymentInfos[0].Total,
			ShippingFee:        order.PaymentInfos[0].ShippingFee,
			Status:             order.PaymentInfos[0].Status,
			PaymentLink:        order.PaymentInfos[0].PaymentLink,
			CancellationReason: order.PaymentInfos[0].CancellationReason,
		}
	}

	data := dto.OrderData{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		PaymentMethod:   string(order.PaymentMethod),
		ShippingAddress: order.ShippingAddress,
		PhoneNumber:     order.PhoneNumber,
		PaymentInfo:     paymentInfo,
		OrderItems:      orderItems,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		DeliveryMode:    order.DeliveryMode,
	}

	return &dto.CreateOrderResponse{
		Data:    data,
		Message: "success",
	}, nil
}

func (o *orderService) MapDraftOrderToCreateOrderResponse(ctx context.Context, order *models.DraftOrder) (*dto.CreateOrderResponse, error) {
	var orderItems []dto.OrderItemResponse

	for _, item := range order.OrderItems {
		orderItems = append(orderItems, dto.OrderItemResponse{
			ID:               item.ID,
			OrderID:          item.OrderID,
			OrderType:        item.OrderType,
			ProductVariantID: item.ProductVariantID,
			Quantity:         item.Quantity,
			Price:            item.Price,
			TotalPrice:       item.TotalPrice,
		})
	}

	var paymentInfo dto.PaymentInfoResponse
	if len(order.PaymentInfos) > 0 {
		paymentInfo = dto.PaymentInfoResponse{
			ID:                 order.PaymentInfos[0].ID,
			Total:              order.PaymentInfos[0].Total,
			ShippingFee:        order.PaymentInfos[0].ShippingFee,
			Status:             order.PaymentInfos[0].Status,
			PaymentLink:        order.PaymentInfos[0].PaymentLink,
			CancellationReason: order.PaymentInfos[0].CancellationReason,
		}
	}

	data := dto.OrderData{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		PaymentMethod:   string(order.PaymentMethod),
		ShippingAddress: order.ShippingAddress,
		DeliveryMode:    order.DeliveryMode,
		PhoneNumber:     order.PhoneNumber,
		PaymentInfo:     paymentInfo,
		OrderItems:      orderItems,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}

	return &dto.CreateOrderResponse{
		Data:    data,
		Message: "success",
	}, nil
}

func (o *orderService) MapOrdersToCreateOrderResponses(ctx context.Context, orders []*models.Order) ([]*dto.CreateOrderResponse, error) {
	var responses []*dto.CreateOrderResponse
	for _, order := range orders {
		resp, err := o.MapOrderToCreateOrderResponse(ctx, order)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

func (o *orderService) MapDraftOrdersToListOrderDataResponses(ctx context.Context, orders []*models.DraftOrder) ([]*dto.OrderData, error) {

	var responses []*dto.OrderData
	for _, order := range orders {
		resp, err := o.MapDraftOrderToListOrderDataResponse(ctx, order)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}
func (o *orderService) MapOrdersToListOrderDataResponses(ctx context.Context, orders []*models.Order) ([]*dto.OrderData, error) {

	var responses []*dto.OrderData
	for _, order := range orders {
		resp, err := o.MapOrderToListOrderDataResponse(ctx, order)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

func (o *orderService) MapOrderToListOrderDataResponse(ctx context.Context, order *models.Order) (*dto.OrderData, error) {
	var orderItems []dto.OrderItemResponse

	for _, item := range order.OrderItems {
		productString := item.Variant.Product.Name
		for _, opt := range item.Variant.OptionValues {
			productString = productString + " " + opt.Value
		}
		orderItems = append(orderItems, dto.OrderItemResponse{
			ID:               item.ID,
			OrderID:          item.OrderID,
			OrderType:        item.OrderType,
			ProductVariantID: item.ProductVariantID,
			Quantity:         item.Quantity,
			Price:            item.Price,
			TotalPrice:       item.TotalPrice,
			ProductName:      productString,
		})
	}

	var paymentInfo dto.PaymentInfoResponse
	if len(order.PaymentInfos) > 0 {
		paymentInfo = dto.PaymentInfoResponse{
			ID:                 order.PaymentInfos[0].ID,
			Total:              order.PaymentInfos[0].Total,
			ShippingFee:        order.PaymentInfos[0].ShippingFee,
			Status:             order.PaymentInfos[0].Status,
			PaymentLink:        order.PaymentInfos[0].PaymentLink,
			CancellationReason: order.PaymentInfos[0].CancellationReason,
		}
	}

	data := &dto.OrderData{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		PaymentMethod:   string(order.PaymentMethod),
		ShippingAddress: order.ShippingAddress,
		PhoneNumber:     order.PhoneNumber,
		PaymentInfo:     paymentInfo,
		OrderItems:      orderItems,
		DeliveryMode:    order.DeliveryMode,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		OrderType:       models.OrderTypeOrder,
		ParentID:        order.ParentID,
	}

	return data, nil
}

func (o *orderService) MapDraftOrderToListOrderDataResponse(ctx context.Context, order *models.DraftOrder) (*dto.OrderData, error) {
	var orderItems []dto.OrderItemResponse

	for _, item := range order.OrderItems {
		productString := item.Variant.Product.Name
		for _, opt := range item.Variant.OptionValues {
			productString = productString + " " + opt.Value
		}
		orderItems = append(orderItems, dto.OrderItemResponse{
			ID:               item.ID,
			OrderID:          item.OrderID,
			OrderType:        item.OrderType,
			ProductVariantID: item.ProductVariantID,
			Quantity:         item.Quantity,
			Price:            item.Price,
			TotalPrice:       item.TotalPrice,
			ProductName:      productString,
		})
	}

	var paymentInfo dto.PaymentInfoResponse
	if len(order.PaymentInfos) > 0 {
		paymentInfo = dto.PaymentInfoResponse{
			ID:                 order.PaymentInfos[0].ID,
			Total:              order.PaymentInfos[0].Total,
			ShippingFee:        order.PaymentInfos[0].ShippingFee,
			Status:             order.PaymentInfos[0].Status,
			PaymentLink:        order.PaymentInfos[0].PaymentLink,
			CancellationReason: order.PaymentInfos[0].CancellationReason,
		}
	}

	data := &dto.OrderData{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		PaymentMethod:   string(order.PaymentMethod),
		ShippingAddress: order.ShippingAddress,
		PhoneNumber:     order.PhoneNumber,
		PaymentInfo:     paymentInfo,
		OrderItems:      orderItems,
		DeliveryMode:    order.DeliveryMode,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		OrderType:       models.OrderTypeDraftOrder,
		ParentID:        order.ParentID,
	}

	return data, nil
}

func (o *orderService) ListByUserId(ctx context.Context, userID uint) ([]*dto.OrderData, error) {
	var results []*dto.OrderData

	orders, err := o.orderRepo.ListByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	drafts, err := o.draftOrderRepo.ListByUserIdToOrderNull(ctx, userID)

	if err != nil {
		return nil, err
	}

	// Draft order Bank method Split
	childItemsMap := make(map[uint][]models.OrderItem)
	var parents []*models.DraftOrder

	// Split parent and child
	for _, draft := range drafts {
		if draft.ParentID != nil && *draft.ParentID != 0 {
			childItemsMap[*draft.ParentID] = append(childItemsMap[*draft.ParentID], draft.OrderItems...)
		} else {
			parents = append(parents, draft)
		}
	}

	// merge child item into parent
	for i := range parents {
		if items, ok := childItemsMap[parents[i].ID]; ok {
			parents[i].OrderItems = append(parents[i].OrderItems, items...)
		}
	}

	// Map to DTO
	orderRs, err := o.MapOrdersToListOrderDataResponses(ctx, orders)
	if err != nil {
		return nil, err
	}

	draftRs, err := o.MapDraftOrdersToListOrderDataResponses(ctx, parents)
	if err != nil {
		return nil, err
	}
	results = append(results, orderRs...)
	results = append(results, draftRs...)

	// Sort by create at
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	return results, nil
}

func (o *orderService) ListByAdmin(c *gin.Context) ([]*dto.OrderData, error) {

	var results []*dto.OrderData
	orders, err := o.orderRepo.ListByAdmin(c)
	if err != nil {
		return nil, err
	}
	drafts, err := o.draftOrderRepo.ListByAdmin(c)
	if err != nil {
		return nil, err
	}
	orderRs, err := o.MapOrdersToListOrderDataResponses(c, orders)
	if err != nil {
		return nil, err
	}
	draftRs, err := o.MapDraftOrdersToListOrderDataResponses(c, drafts)
	if err != nil {
		return nil, err
	}
	results = append(results, orderRs...)
	results = append(results, draftRs...)
	return results, nil
}

func (o *orderService) loadAndCacheProductVariants(ctx context.Context, ids []uint) ([]models.ProductVariant, error) {
	// Get List from Db
	variants, err := o.productVariantRepo.GetByIDSForRedisCache(ctx, ids)
	if err != nil {
		return nil, err
	}
	// If DB returns fewer records, return ITEM_NOT_FOUND error for missing ids
	if len(variants) != len(ids) {
		found := map[uint]bool{}
		for _, v := range variants {
			found[v.ID] = true
		}
		for _, id := range ids {
			if !found[id] {
				return nil, customErr.NewError(
					customErr.ITEM_NOT_FOUND,
					fmt.Sprintf("Product variant not found: %d", id),
					http.StatusBadRequest, nil,
				)
			}
		}
	}

	// Save Redis cache
	for _, pv := range variants {
		if err := o.productVariantCache.SaveProductVariantHash(pv); err != nil {
			log.Printf("warning: save productVariant %d to redis failed: %v", pv.ID, err)
		}
	}

	return variants, nil
}

func (o *orderService) CreateOrderWithDb(ctx context.Context, input dto.CreateOrderInput) (models.DraftOrder, []models.OrderItem, error) {
	var createdDraftOrder models.DraftOrder
	var createdItems []models.OrderItem

	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Lock tất cả variants liên quan
		var variantIDs []uint
		for _, item := range input.OrderItems {
			variantIDs = append(variantIDs, item.ProductVariantID)
		}

		var variants []models.ProductVariant
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", variantIDs).
			Find(&variants).Error; err != nil {
			return err
		}
		// Map quantity and productVariant ID
		orderQtyMap := make(map[uint]uint)
		for _, item := range input.OrderItems {
			orderQtyMap[item.ProductVariantID] = item.Quantity
		}

		// 2. Get ReservedQty
		type ReservedQty struct {
			ProductVariantID uint
			TotalQuantity    uint
		}
		var reservedList1 []ReservedQty
		var reservedList2 []ReservedQty

		if err := tx.Table("order_items oi").
			Select("oi.product_variant_id, COALESCE(SUM(oi.quantity), 0) as total_quantity").
			Joins("INNER JOIN draft_orders do ON do.id = oi.order_id").
			Where("oi.product_variant_id IN ? AND oi.order_type = ? AND do.to_order IS NULL", variantIDs, "draft_order").
			Group("oi.product_variant_id").
			Scan(&reservedList1).Error; err != nil {
			return err
		}

		if err := tx.Table("order_items oi").
			Select("oi.product_variant_id, COALESCE(SUM(oi.quantity), 0) as total_quantity").
			Joins("INNER JOIN draft_orders do ON do.to_order = oi.order_id").
			Where("oi.product_variant_id IN ? AND oi.order_type = ? AND do.to_order !=0 AND do.to_order IS NOT NULL", variantIDs, "order").
			Group("oi.product_variant_id").
			Scan(&reservedList2).Error; err != nil {
			return err
		}

		reservedMap := make(map[uint]uint)
		for _, r := range reservedList1 {
			reservedMap[r.ProductVariantID] += r.TotalQuantity
		}
		for _, r := range reservedList2 {
			reservedMap[r.ProductVariantID] += r.TotalQuantity
		}

		log.Println(reservedMap)
		// 3. Map Variant for checking stock
		variantMap := make(map[uint]models.ProductVariant)
		for _, v := range variants {
			variantMap[v.ID] = v
		}

		// 4. Validate stock
		for _, item := range input.OrderItems {
			pv, ok := variantMap[item.ProductVariantID]
			if !ok {
				return fmt.Errorf("variant %d not found", item.ProductVariantID)
			}
			available := pv.Quantity - reservedMap[item.ProductVariantID]
			if item.Quantity > available {
				return fmt.Errorf("variant %d out of stock: requested %d, available %d",
					item.ProductVariantID, item.Quantity, available)
			}
		}

		// 5. Create  draft_order
		createdDraftOrder = models.DraftOrder{
			UserID:          input.UserID,
			PaymentMethod:   input.PaymentMethod,
			Status:          models.OrderStatePending,
			ShippingAddress: input.ShippingAddress,
			PhoneNumber:     input.PhoneNumber,
			DeliveryMode:    input.ShippingFeeInput[0].Mode,
			Latitude:        input.Latitude,
			Longitude:       input.Longitude,
		}
		if err := tx.Create(&createdDraftOrder).Error; err != nil {
			return err
		}

		// 6. Create order_items
		for _, item := range input.OrderItems {
			newItem := models.OrderItem{
				OrderID:          createdDraftOrder.ID,
				OrderType:        models.OrderTypeDraftOrder,
				ProductVariantID: item.ProductVariantID,
				Quantity:         item.Quantity,
				Price:            item.Price,
				TotalPrice:       float64(item.Quantity) * item.Price,
				MerchantID:       item.MerchantID,
			}
			createdItems = append(createdItems, newItem)
		}

		if err := tx.Create(&createdItems).Error; err != nil {
			return err
		}

		// 7. Create Delivery detail
		newDeliveryDetail := models.DeliveryDetail{
			OrderID:    createdDraftOrder.ID,
			OrderType:  models.OrderTypeDraftOrder,
			DeliveryID: input.ShippingFeeInput[0].DeliveryID,
		}
		if err := tx.Create(&newDeliveryDetail).Error; err != nil {
			return err
		}
		createdDraftOrder.Delivery = newDeliveryDetail
		return nil

	})

	if err != nil {
		return models.DraftOrder{}, nil, err
	}
	createdDraftOrder.OrderItems = createdItems
	return createdDraftOrder, createdItems, nil
}

func GeneratePaymentInfoID() int64 {
	node := configs.GetSnowflakeNode()
	return node.Generate().Int64() / 1000
}

func (o *orderService) ChangePaymentMethod(c *gin.Context, paymentChange dto.ChangePaymentMethodRequest, userID uint) (*models.Order, error) {
	payment, order, draft, err := o.paymentInfoRepo.GetByIdAndUserIdAndOrderId(c, paymentChange.PaymentID, userID, paymentChange.OrderId)

	if payment.Status != models.PaymentPending {
		return nil, customErr.NewError(customErr.BAD_REQUEST, "Cant change payment method", http.StatusBadRequest, nil)
	}
	if err != nil {
		return nil, err
	}

	if payment.OrderType == models.OrderTypeDraftOrder {

		if draft.ID == 0 {
			return nil, customErr.NewError(customErr.ITEM_NOT_FOUND, "Cant find order", http.StatusBadRequest, err)
		}
		if paymentChange.PaymentMethod == models.PaymentMethodBank {
			return nil, customErr.NewError(customErr.BAD_REQUEST, "Cant change payment method from BANK to BANK", http.StatusBadRequest, errors.New("Cant change BANK to BANK method"))
		}
		if draft.ToOrderID != nil {
			return nil, customErr.NewError(customErr.BAD_REQUEST, "Can't change this order payment method", http.StatusBadRequest, err)
		}
		// Change to COD
		// Check if order need to split
		//TODO
		if draft.ParentID != nil && *draft.ParentID == 0 {
			//split
			infos, err2 := o.draftOrderRepo.GetForSplit(c, draft.ID)
			if err2 != nil {
				log.Println(err2, " While split order (bank payment success)")
				return nil, customErr.NewError(customErr.BAD_REQUEST, "Cant change payment method", http.StatusBadRequest, err)
			}

			//Get merchant id distinct
			var merchantIDs []uint
			merchantIDMap := make(map[uint]bool)
			itemAndMerchantMap := make(map[uint]uint) // key: OrderItem ID, value: MerchantID

			for _, info := range infos {
				if !merchantIDMap[info.MerchantID] {
					merchantIDMap[info.MerchantID] = true
					merchantIDs = append(merchantIDs, info.MerchantID)
				}
				itemAndMerchantMap[info.ID] = info.MerchantID
			}

			var shippingFeeResponse []dto.ShippingFeeResponse
			for _, merchantID := range merchantIDs {
				fee, err := o.CalculateShippingFee(c, merchantID, draft.Longitude, draft.Latitude, infos[0].DeliveryID)
				if err != nil {
					return nil, customErr.NewError(customErr.BAD_REQUEST, "Cant change payment method 2", http.StatusBadRequest, err)
				}
				shippingFeeResponse = append(shippingFeeResponse, fee...)
			}

			// Change payment method and convert to order
			orderUpdated, err := o.ChangeToCODPaymentFromDraft(c, draft, payment.Total, payment.ShippingFee)
			if err != nil {
				return orderUpdated, err
			}

			// Inject merchant ID for orderItems
			for i := range orderUpdated.OrderItems {
				orderUpdated.OrderItems[i].MerchantID = itemAndMerchantMap[orderUpdated.OrderItems[i].ID]
			}
			// Split
			o.SplitOrderAfterChangePaymentMethod(c, orderUpdated, orderUpdated.OrderItems, merchantIDs, shippingFeeResponse)

			cancelReason := "Change payment method"
			_, err = payos.CancelPaymentLink(strconv.FormatInt(payment.ID, 10), &cancelReason)
			if err != nil {
				log.Println("Payment id: ", payment.ID, " cant cancel payment payos", err)
			}
			return orderUpdated, nil
		} else {
			//nosplit
			orderUpdated, err := o.ChangeToCODPaymentFromDraft(c, draft, payment.Total, payment.ShippingFee)
			if err != nil {
				return orderUpdated, err
			}
			// Cancel payment link
			cancelReason := "Change payment method"
			_, err = payos.CancelPaymentLink(strconv.FormatInt(payment.ID, 10), &cancelReason)
			if err != nil {
				log.Println("Payment id: ", payment.ID, " cant cancel payment payos", err)
			}
			return orderUpdated, nil
		}
	} else {
		//TODO IMPLEMENT ORDER change
		if order.ID == 0 {
			return nil, customErr.NewError(customErr.ITEM_NOT_FOUND, "Cant find order", http.StatusBadRequest, err)
		}
		if order.PaymentMethod == paymentChange.PaymentMethod {
			return nil, customErr.NewError(customErr.BAD_REQUEST, fmt.Sprintf("Cant change payment method from %s to %s", order.PaymentMethod, paymentChange.PaymentMethod), http.StatusBadRequest, nil)
		}
		if paymentChange.PaymentMethod == models.PaymentMethodCOD {
			updatedOrder, err2 := o.ChangeToCODPaymentFromOrder(c, order, payment.Total, payment.ShippingFee)
			if err2 != nil {
				return nil, err2
			}
			cancelReason := "Change payment method"
			_, err = payos.CancelPaymentLink(strconv.FormatInt(payment.ID, 10), &cancelReason)
			if err != nil {
				log.Println("Payment id: ", payment.ID, " cant cancel payment payos", err)
			}
			return updatedOrder, nil
		} else if paymentChange.PaymentMethod == models.PaymentMethodBank {
			updatedOrder, err := o.ChangeToBankPaymentFromOrder(c, order, payment.Total, payment.ShippingFee)
			if err != nil {
				return nil, err
			}
			payment.Status = models.PaymentCanceled
			payment.CancellationReason = "Change payment method"
			if err := o.paymentInfoRepo.Save(c, &payment); err != nil {
				log.Println("Fail  to cancel previous COD payment")
			}
			return updatedOrder, nil
		}
	}
	return order, nil
}

func (o *orderService) ChangeToCODPaymentFromDraft(c *gin.Context, draft *models.DraftOrder, total float64, shippingFee float64) (*models.Order, error) {

	order := &models.Order{
		UserID:          draft.UserID,
		Status:          models.OrderStatePending,
		PaymentMethod:   models.PaymentMethodCOD,
		ShippingAddress: draft.ShippingAddress,
		PhoneNumber:     draft.PhoneNumber,
		DeliveryMode:    draft.DeliveryMode,
		Longitude:       draft.Longitude,
		Latitude:        draft.Latitude,
		Delivery:        draft.Delivery,
	}
	if err := o.orderRepo.Create(c, order); err != nil {
		return nil, err
	}

	var paymentInfo = &models.PaymentInfo{
		ID:          GeneratePaymentInfoID(),
		Total:       total,
		ShippingFee: shippingFee,
		OrderID:     order.ID,
		OrderType:   models.OrderTypeOrder,
		Status:      models.PaymentPending,
	}
	if err := o.paymentInfoRepo.Create(c, paymentInfo); err != nil {
		log.Printf(err.Error(), "while creating payment info")
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "CreatePayment error", http.StatusInternalServerError, err)
	}
	for _, oi := range draft.OrderItems {
		oi.OrderID = order.ID
		oi.OrderType = models.OrderTypeOrder
		if err := o.orderItemRepo.Save(c, &oi); err != nil {
			log.Printf(err.Error(), "while saving order item")
		}
		order.OrderItems = append(order.OrderItems, oi)
	}
	order.OrderItems = draft.OrderItems

	draft.OrderItems = nil
	draft.ToOrderID = &order.ID
	draft.Delivery = models.DeliveryDetail{}
	if err := o.draftOrderRepo.Save(c, draft); err != nil {
		log.Printf(err.Error(), "while saving draft")
	}
	order.PaymentInfos = nil
	order.PaymentInfos = append(order.PaymentInfos, *paymentInfo)
	return order, nil
}

func (o *orderService) ChangeToCODPaymentFromOrder(c *gin.Context, order *models.Order, total float64, shippingFee float64) (*models.Order, error) {

	order.PaymentMethod = models.PaymentMethodCOD

	if err := o.orderRepo.Save(c, order); err != nil {
		return nil, err
	}

	var paymentInfo = &models.PaymentInfo{
		ID:          GeneratePaymentInfoID(),
		Total:       total,
		ShippingFee: shippingFee,
		OrderID:     order.ID,
		OrderType:   models.OrderTypeOrder,
		Status:      models.PaymentPending,
	}
	if err := o.paymentInfoRepo.Create(c, paymentInfo); err != nil {
		log.Printf(err.Error(), "while creating payment info")
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "CreatePayment error", http.StatusInternalServerError, err)
	}
	order.PaymentInfos = nil
	order.PaymentInfos = append(order.PaymentInfos, *paymentInfo)
	return order, nil
}

func (o *orderService) ChangeToBankPaymentFromOrder(c *gin.Context, order *models.Order, total float64, shippingFee float64) (*models.Order, error) {

	order.PaymentMethod = models.PaymentMethodBank
	var paymentInfo = &models.PaymentInfo{
		ID:          GeneratePaymentInfoID(),
		Total:       total,
		ShippingFee: shippingFee,
		OrderID:     order.ID,
		OrderType:   models.OrderTypeOrder,
		Status:      models.PaymentPending,
	}

	err := o.orderRepo.Save(c, order)
	if err != nil {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Change order payment error", http.StatusInternalServerError, err)
	}
	bankPayment, err := CreatePayOSPayment(paymentInfo.ID, paymentInfo.Total, MapOrderItemsToPayOSItems(order.OrderItems, int(paymentInfo.ShippingFee)), "Thanh toan don hang", "localhost:5173", "localhost:5173")
	if err != nil {
		log.Printf(err.Error(), "while creating payment info")
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "CreatePayment error", http.StatusInternalServerError, err)
	}

	// Publish event after create payOS link
	paymentEvent := event.PayOSPaymentCreatedEvent{
		Id:            paymentInfo.ID,
		OrderID:       order.ID,
		PaymentLink:   bankPayment.CheckoutUrl,
		Total:         10000,
		PaymentMethod: string(order.PaymentMethod),
		CreatedAt:     time.Now(),
	}

	if err := o.eventBus.PublishPaymentCreated(paymentEvent); err != nil {
		log.Printf("Failed to publish payment created event: %v", err)
	}
	log.Printf("PayOs payment created for order %d: %s", order.ID, bankPayment.CheckoutUrl)

	// Save payment link & return
	paymentInfo.PaymentLink = bankPayment.CheckoutUrl

	if err := o.paymentInfoRepo.Create(c, paymentInfo); err != nil {
		log.Printf(err.Error(), "while creating payment info")
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "CreatePayment error", http.StatusInternalServerError, err)
	}
	order.PaymentInfos = nil
	order.PaymentInfos = append(order.PaymentInfos, *paymentInfo)
	return order, nil
}
func (o *orderService) SplitOrder(ctx context.Context, draftOrder *models.DraftOrder, orderItems []models.OrderItem, merchantIDs []uint, shippingFeeResponses []dto.ShippingFeeResponse) ([]*models.DraftOrder, error) {

	groups := make(map[uint][]models.OrderItem)

	// Group order items by merchantID
	for _, oi := range orderItems {
		groups[oi.MerchantID] = append(groups[oi.MerchantID], oi)
	}
	// Slice to save sub draft
	var subDraftOrders []*models.DraftOrder

	for _, merchantID := range merchantIDs {
		orderItemsSplit := groups[merchantID]
		draftOrderSplit := &models.DraftOrder{
			UserID:          draftOrder.UserID,
			Status:          models.OrderStatePending,
			ShippingAddress: draftOrder.ShippingAddress,
			PaymentMethod:   draftOrder.PaymentMethod,
			PhoneNumber:     draftOrder.PhoneNumber,
			ParentID:        &draftOrder.ID,
			DeliveryMode:    draftOrder.DeliveryMode,
			Latitude:        draftOrder.Latitude,
			Longitude:       draftOrder.Longitude,
		}
		if err := o.draftOrderRepo.Create(ctx, draftOrderSplit); err != nil {
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Split order error", http.StatusInternalServerError, err)
		}

		// Create sub delivery detail
		subDeliveryDetail := &models.DeliveryDetail{
			OrderID:    draftOrderSplit.ID,
			DeliveryID: draftOrder.Delivery.DeliveryID,
			OrderType:  models.OrderTypeDraftOrder,
		}

		if err := o.db.WithContext(ctx).Create(&subDeliveryDetail).Error; err != nil {
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Split order error 2", http.StatusInternalServerError, err)
		}

		// Change order items reference
		var total float64
		for _, orderItemSplit := range orderItemsSplit {
			orderItemSplit.OrderID = draftOrderSplit.ID
			total += orderItemSplit.TotalPrice
			if err := o.orderItemRepo.Save(ctx, &orderItemSplit); err != nil {
				return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Change order item reference error", http.StatusInternalServerError, err)
			}
		}

		// Find shipping fee of merchant
		var tempShippingFee float64
		for _, ship := range shippingFeeResponses {
			if ship.MerchantID == merchantID {
				tempShippingFee = ship.Fee
				break
			}
		}
		log.Println("1, tempshippingfee: ", tempShippingFee, " total:", total)
		// Create payment for sub draft
		var paymentSplit = &models.PaymentInfo{
			ID:          GeneratePaymentInfoID(),
			Total:       total + tempShippingFee,
			ShippingFee: tempShippingFee,
			OrderID:     draftOrderSplit.ID,
			OrderType:   models.OrderTypeDraftOrder,
			Status:      draftOrder.PaymentInfos[0].Status,
			ParentID:    &draftOrder.PaymentInfos[0].ID,
		}
		log.Println("2, paymentsplit: ", paymentSplit.Total, " ,", paymentSplit.ShippingFee)
		if err := o.paymentInfoRepo.Create(ctx, paymentSplit); err != nil {
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, "CreatePayment error", http.StatusInternalServerError, err)
		}

		// Append sub draft to slice
		draftOrderSplit.OrderItems = orderItemsSplit
		draftOrderSplit.PaymentInfos = []models.PaymentInfo{*paymentSplit}
		draftOrderSplit.Delivery = *subDeliveryDetail
		subDraftOrders = append(subDraftOrders, draftOrderSplit)
	}

	// Reset  parent draft
	temp1 := uint(0)
	temp2 := int64(0)
	draftOrder.OrderItems = nil
	draftOrder.ParentID = &temp1
	draftOrder.PaymentInfos[0].ParentID = &temp2
	err := o.draftOrderRepo.Save(ctx, draftOrder)
	if err != nil {
		log.Println("Error while split draft order ", err)
	}

	err = o.paymentInfoRepo.Save(ctx, &draftOrder.PaymentInfos[0])
	if err != nil {
		log.Println("Error while split draft order ", err)
	}
	subDraftOrders = append(subDraftOrders, draftOrder)

	return subDraftOrders, nil
}

func (o *orderService) SplitOrderAfterChangePaymentMethod(ctx context.Context, order *models.Order, orderItems []models.OrderItem, merchantIDs []uint, shippingFeeResponses []dto.ShippingFeeResponse) ([]*models.Order, error) {

	// TODO IMPLEMENT
	groups := make(map[uint][]models.OrderItem)

	// Group order items by merchantID
	for _, oi := range orderItems {
		groups[oi.MerchantID] = append(groups[oi.MerchantID], oi)
	}
	// Slice to save sub order
	var subOrders []*models.Order

	// Create order item for sub order
	for _, merchantID := range merchantIDs {
		orderItemsSplit := groups[merchantID]
		orderSplit := &models.Order{
			UserID:          order.UserID,
			Status:          models.OrderStatePending,
			ShippingAddress: order.ShippingAddress,
			PaymentMethod:   order.PaymentMethod,
			PhoneNumber:     order.PhoneNumber,
			ParentID:        &order.ID,
			DeliveryMode:    order.DeliveryMode,
			Latitude:        order.Latitude,
			Longitude:       order.Longitude,
		}
		if err := o.orderRepo.Create(ctx, orderSplit); err != nil {
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Split order error", http.StatusInternalServerError, err)
		}

		// Create sub delivery detail
		subDeliveryDetail := &models.DeliveryDetail{
			OrderID:    orderSplit.ID,
			DeliveryID: shippingFeeResponses[0].DeliveryID,
			OrderType:  models.OrderTypeOrder,
		}

		if err := o.db.WithContext(ctx).Create(&subDeliveryDetail).Error; err != nil {
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Split order error 2", http.StatusInternalServerError, err)
		}

		// Change order items reference
		var total float64
		for _, orderItemSplit := range orderItemsSplit {
			orderItemSplit.OrderID = orderSplit.ID
			total += orderItemSplit.TotalPrice
			orderItemSplit.OrderType = models.OrderTypeOrder
			if err := o.orderItemRepo.Save(ctx, &orderItemSplit); err != nil {
				return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Change order item reference error", http.StatusInternalServerError, err)
			}
		}

		// Find shipping fee of merchant
		var tempShippingFee float64
		for _, ship := range shippingFeeResponses {
			if ship.MerchantID == merchantID {
				tempShippingFee = ship.Fee
				break
			}
		}
		log.Println("1, tempshippingfee: ", tempShippingFee, " total:", total)
		// Create payment for sub draft
		var paymentSplit = &models.PaymentInfo{
			ID:          GeneratePaymentInfoID(),
			Total:       total + tempShippingFee,
			ShippingFee: tempShippingFee,
			OrderID:     orderSplit.ID,
			OrderType:   models.OrderTypeOrder,
			Status:      order.PaymentInfos[0].Status,
			ParentID:    &order.PaymentInfos[0].ID,
		}
		log.Println("2, paymentsplit: ", paymentSplit.Total, " ,", paymentSplit.ShippingFee)
		if err := o.paymentInfoRepo.Create(ctx, paymentSplit); err != nil {
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, "CreatePayment error", http.StatusInternalServerError, err)
		}

		// Append sub draft to slice
		orderSplit.OrderItems = orderItemsSplit
		orderSplit.PaymentInfos = []models.PaymentInfo{*paymentSplit}
		orderSplit.Delivery = *subDeliveryDetail
		subOrders = append(subOrders, orderSplit)
	}

	// Reset  parent draft
	temp1 := uint(0)
	temp2 := int64(0)
	order.OrderItems = nil
	order.ParentID = &temp1
	order.PaymentInfos[0].ParentID = &temp2
	err := o.orderRepo.Save(ctx, order)
	if err != nil {
		log.Println("Error while split draft order ", err)
	}

	err = o.paymentInfoRepo.Save(ctx, &order.PaymentInfos[0])
	if err != nil {
		log.Println("Error while split draft order ", err)
	}
	subOrders = append(subOrders, order)

	return subOrders, nil
}

func (o *orderService) CalculateShippingFees(c context.Context, merchantID uint, destLon, destLat string) ([]*dto.ShippingFeeResponse, error) {
	merchant, err := o.merchantRepo.GetByID(c, merchantID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://localhost:5000/route/v1/driving/%s,%s;%s,%s?overview=false",
		merchant.Longitude, merchant.Latitude, destLon, destLat)

	log.Println("Routing url: ", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var osrmResp dto.OSRMResponse
	if err := json.Unmarshal(body, &osrmResp); err != nil {
		return nil, err
	}

	if len(osrmResp.Routes) == 0 {
		return nil, fmt.Errorf("no route found")
	}
	distanceKm := float64(int(osrmResp.Routes[0].Distance / 1000.0))

	var shippingFee []*dto.ShippingFeeResponse
	deliveries, err := o.merchantRepo.GetDeliveriesInfo(c)
	if err != nil {
		return nil, err
	}
	for _, delivery := range deliveries {
		fee := delivery.BasePrice + delivery.PricePerKm*distanceKm
		feeDto := &dto.ShippingFeeResponse{
			Name:       delivery.Name,
			Mode:       delivery.DeliveryMode,
			Fee:        fee,
			MerchantID: merchantID,
			Timestamp:  time.Now().Unix(),
			DeliveryID: delivery.ID,
		}
		feeDto.Signature = utils.GenerateShippingFeeSignature(merchantID, delivery.ID, fee, destLat, destLon, feeDto.Timestamp)
		shippingFee = append(shippingFee, feeDto)
	}
	return shippingFee, nil
}

func (o *orderService) CalculateShippingFee(c context.Context, merchantID uint, destLon, destLat string, deliveryID uint) ([]dto.ShippingFeeResponse, error) {
	merchant, err := o.merchantRepo.GetByID(c, merchantID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://localhost:5000/route/v1/driving/%s,%s;%s,%s?overview=false",
		merchant.Longitude, merchant.Latitude, destLon, destLat)

	log.Println("Routing url: ", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var osrmResp dto.OSRMResponse
	if err := json.Unmarshal(body, &osrmResp); err != nil {
		return nil, err
	}

	if len(osrmResp.Routes) == 0 {
		return nil, fmt.Errorf("no route found")
	}
	distanceKm := float64(int(osrmResp.Routes[0].Distance / 1000.0))

	var shippingFee []dto.ShippingFeeResponse
	delivery, err := o.merchantRepo.GetDeliveryInfo(c, deliveryID)
	if err != nil {
		return nil, err
	}
	fee := delivery.BasePrice + delivery.PricePerKm*distanceKm
	feeDto := dto.ShippingFeeResponse{
		Name:       delivery.Name,
		Mode:       delivery.DeliveryMode,
		Fee:        fee,
		MerchantID: merchantID,
		Timestamp:  time.Now().Unix(),
		DeliveryID: delivery.ID,
	}
	feeDto.Signature = utils.GenerateShippingFeeSignature(merchantID, delivery.ID, fee, destLat, destLon, feeDto.Timestamp)
	shippingFee = append(shippingFee, feeDto)

	return shippingFee, nil
}
