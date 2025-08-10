package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/minh6824pro/nxrGO/cache"
	"github.com/minh6824pro/nxrGO/configs"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
	"github.com/payOSHQ/payos-lib-golang"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"time"
)

type orderService struct {
	db                 *gorm.DB
	productVariantRepo repositories.ProductVariantRepository
	orderItemRepo      repositories.OrderItemRepository
	orderRepo          repositories.OrderRepository
	draftOrderRepo     repositories.DraftOrderRepository
	paymentInfoRepo    repositories.PaymentInfoRepository
	redisClient        *redis.Client
	pendingReplies     map[string]chan dto.OrderProcessingResult // correlationID -> reply channel
}

func NewOrderService(db *gorm.DB, productVariantRepo repositories.ProductVariantRepository, orderItemRepo repositories.OrderItemRepository,
	orderRepo repositories.OrderRepository, draftOrderRepo repositories.DraftOrderRepository,
	paymentInfoRepo repositories.PaymentInfoRepository) services.OrderService {
	service := &orderService{
		db:                 db,
		productVariantRepo: productVariantRepo,
		orderRepo:          orderRepo,
		orderItemRepo:      orderItemRepo,
		draftOrderRepo:     draftOrderRepo,
		paymentInfoRepo:    paymentInfoRepo,
		redisClient:        nil,
		pendingReplies:     make(map[string]chan dto.OrderProcessingResult),
	}

	return service
}

//func (o *orderService) Create(ctx context.Context, input dto.CreateOrderInput) (*models.Order, error) {
//	var total float64 = 0
//
//	// Check items existence & quantity
//	for _, orderItem := range input.OrderItems {
//		// Check redis
//		val, err := cache.GetProductVariant(fmt.Sprintf(cache.KeyPattern, orderItem.ProductVariantID))
//		if err != nil {
//			log.Println(err)
//		}
//		// No redis -> query db
//		if val == nil {
//			productVariant, err := o.productVariantRepo.GetByIDNoPreload(ctx, orderItem.ProductVariantID)
//			if err != nil {
//				if errors.Is(err, gorm.ErrRecordNotFound) {
//					return nil, customErr.NewError(customErr.ITEM_NOT_FOUND, fmt.Sprintf("Product variant not found: %d", orderItem.ProductVariantID), http.StatusBadRequest, nil)
//				}
//				return nil, err
//			}
//			err = cache.SaveProductVariant(fmt.Sprintf(cache.KeyPattern, orderItem.ProductVariantID), *productVariant, 12*time.Hour)
//			if err != nil {
//				log.Println(err)
//				return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected error", http.StatusInternalServerError, nil)
//			}
//			val = productVariant
//		}
//		if orderItem.Quantity > val.Quantity {
//			return nil, customErr.NewError(customErr.INSUFFICIENT_STOCK, fmt.Sprintf("Product variant : %d Insufficient stock", orderItem.ProductVariantID), http.StatusBadRequest, nil)
//		}
//		if orderItem.Price != val.Price {
//			return nil, customErr.NewError(customErr.INVALID_PRICE, fmt.Sprintf("Product variant: %d Price Invalid", orderItem.ProductVariantID), http.StatusBadRequest, nil)
//		}
//
//		total += val.Price * float64(orderItem.Quantity)
//	}
//
//	// Check Total Price
//	if total+input.ShippingFee != input.Total {
//		fmt.Println(input.Total)
//		fmt.Println(total)
//		return nil, customErr.NewError(customErr.INVALID_PRICE, "Total Price Invalid", http.StatusBadRequest, nil)
//	}
//
//	// Create Draft Order
//	draft_order := models.DraftOrder{
//		UserID:          input.UserID,
//		Status:          models.OrderStatePending,
//		Total:           total,
//		ShippingAddress: input.ShippingAddress,
//		ShippingFee:     input.ShippingFee,
//		PaymentMethod:   input.PaymentMethod,
//		PhoneNumber:     input.PhoneNumber,
//	}
//
//	if err := o.draftOrderRepo.Create(ctx, &draft_order); err != nil {
//		return nil, customErr.NewError(customErr.INTERNAL_ERROR, fmt.Sprintf("Draft order creation error: %v", err.Error()), http.StatusBadRequest, nil)
//	}
//
//	var orderItems []models.OrderItem
//	// Create orderItems
//	for _, item := range input.OrderItems {
//		orderItem := models.OrderItem{
//			OrderID:          draft_order.ID,
//			ProductVariantID: item.ProductVariantID,
//			OrderType:        models.OrderTypeDraftOrder,
//			Quantity:         item.Quantity,
//			Price:            item.Price,
//			TotalPrice:       total,
//		}
//		if err := o.orderItemRepo.Create(ctx, &orderItem); err != nil {
//			return nil, err
//		}
//		orderItems = append(orderItems, orderItem)
//	}
//	// process Payment Info
//	//o.CreatePayment(ctx, &draft_order, orderItems)
//	//CreatePayment(&draft_order, orderItems)
//	return nil, nil
//}

// Create - phiên bản dùng Redis Hash + Lua trả tất cả MISS, batch DB query cho những MISS
func (o *orderService) Create(ctx context.Context, input dto.CreateOrderInput) (*models.Order, error) {
	var total float64 = 0

	luaScript := `
local itemCount = tonumber(ARGV[1])
local idx = 2

-- 1) Kiểm tra xem có key nào MISS không (EXISTS)
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
    for i = 1, #missed do table.insert(ret, missed[i]) end
    return ret
end

-- 2) Nếu không MISS, kiểm tra stock/price và chuẩn bị updates
idx = 2
local updates = {}
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
    table.insert(updates, { key = key, qty = qtyRequested })
end

-- 3) Nếu tất cả ok -> trừ stock
for i = 1, #updates do
    redis.call("HINCRBY", updates[i].key, "quantity", -updates[i].qty)
end

return {"OK"}
`

	// build keys & args
	keys := make([]string, 0, len(input.OrderItems))
	args := []interface{}{len(input.OrderItems)}
	for _, oi := range input.OrderItems {
		keys = append(keys, fmt.Sprintf(cache.KeyPattern, oi.ProductVariantID))
		// ARGV: variantId, quantity, price (lua dùng tonumber khi cần)
		args = append(args, oi.ProductVariantID, oi.Quantity, oi.Price)
	}

	// helper để convert interface{} -> string an toàn
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

	const maxRetries = 3
	var res interface{}
	var err error

	// lần đầu chạy script
	res, err = configs.RedisClient.Eval(ctx, luaScript, keys, args...).Result()
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
			// arr[1:] là list variantId bị miss
			var missingIDs []uint
			log.Print("Redis key miss: ", missingIDs)
			for i := 1; i < len(arr); i++ {
				idStr := toStr(arr[i])
				id64, parseErr := strconv.ParseUint(idStr, 10, 64)
				if parseErr != nil {
					return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Invalid variant id from redis", http.StatusInternalServerError, nil)
				}
				missingIDs = append(missingIDs, uint(id64))
			}

			// Batch query DB một lần cho all missingIDs
			variants, err := o.productVariantRepo.GetByIDs(ctx, missingIDs)
			if err != nil {
				return nil, err
			}

			// Nếu DB không trả đủ, trả về lỗi ITEM_NOT_FOUND cho id mất
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

			// Lưu tất cả variants vừa query vào Redis (Hash)
			for _, pv := range variants {
				// lưu dạng hash: id, quantity, price
				if err := cache.SaveProductVariantHash(pv, 12*time.Hour); err != nil {
					// log error nhưng không fail toàn bộ request; tuy nhiên có thể tùy chỉnh
					log.Printf("warning: save productVariant %d to redis failed: %v", pv.ID, err)
				}
			}

			// retry: chạy lại lua script
			res, err = configs.RedisClient.Eval(ctx, luaScript, keys, args...).Result()
			if err != nil {
				return nil, err
			}
			continue // vòng for sẽ parse res mới

		case "INSUFFICIENT":
			variantId := ""
			if len(arr) > 1 {
				variantId = toStr(arr[1])
			}
			return nil, customErr.NewError(customErr.INSUFFICIENT_STOCK, fmt.Sprintf("Product variant : %s Insufficient stock", variantId), http.StatusBadRequest, nil)

		case "INVALID_PRICE":
			variantId := ""
			if len(arr) > 1 {
				variantId = toStr(arr[1])
			}
			return nil, customErr.NewError(customErr.INVALID_PRICE, fmt.Sprintf("Product variant: %s Price Invalid", variantId), http.StatusBadRequest, nil)

		case "OK":
			// tất cả ok -> tính total từ input (giá đã được xác thực)
			for _, oi := range input.OrderItems {
				total += oi.Price * float64(oi.Quantity)
			}
			// nhảy ra khỏi loop để tiếp phần tạo order
			attempt = maxRetries // break outer loop
			break

		default:
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected Lua script status", http.StatusInternalServerError, nil)
		}
	}

	// Nếu sau maxRetries vẫn chưa OK => lỗi
	// (ví dụ: lưu cache thất bại liên tục)
	// Bạn có thể tùy chỉnh message/loại lỗi
	resArr, _ := res.([]interface{})
	if len(resArr) == 0 || toStr(resArr[0]) != "OK" {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Failed to reserve stock after retries", http.StatusInternalServerError, nil)
	}

	// Check tổng giá
	if total+input.ShippingFee != input.Total {
		return nil, customErr.NewError(customErr.INVALID_PRICE, "Total Price Invalid", http.StatusBadRequest, nil)
	}

	// Create Draft Order (giữ nguyên flow của bạn)
	draftOrder := models.DraftOrder{
		UserID:          input.UserID,
		Status:          models.OrderStatePending,
		Total:           total,
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

	// TODO: trả về order thực tế nếu cần (hiện giữ giống trước: nil,nil hoặc build model Order từ draftOrder + items)
	return nil, nil
}
func (o *orderService) CreatePayment(ctx context.Context, draftOrder *models.DraftOrder, orderItems []models.OrderItem) error {

	paymentLink := ""
	if draftOrder.PaymentMethod == models.PaymentMethodBank {
		// Create PayOS payment link
		paymentData, err := CreatePayOSPayment(int(draftOrder.ID), draftOrder.Total, MapOrderItemsToPayOSItems(orderItems, int(draftOrder.ShippingFee)), "Thanh toán đơn hàng", "returnUrl", "cancelUrl")
		if err != nil {
			log.Println("CreatePayment error", err.Error())
			return customErr.NewError(customErr.INTERNAL_ERROR, "CreatePayment error", http.StatusInternalServerError, err)
		}
		paymentLink = paymentData.CheckoutUrl
	}
	var paymentInfo = &models.PaymentInfo{
		Amount:      draftOrder.Total,
		Status:      models.PaymentPending,
		PaymentLink: paymentLink,
	}
	if err := o.paymentInfoRepo.Create(ctx, paymentInfo); err != nil {
		log.Printf(err.Error(), "while creating payment info")
		return err
	}

	//draftOrder.PaymentInfo = paymentInfo
	//if err := db.Save(&order).Error; err != nil {
	//	log.Printf(err.Error(), "while saving payment info to order")
	//}

	return nil
}

func (o *orderService) startReplyConsumer() {
	msgs, err := configs.RMQChannel.Consume(configs.OrderReplyQueue, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume from order.reply: %v", err)
	}

	for d := range msgs {
		var result dto.OrderProcessingResult
		if err := json.Unmarshal(d.Body, &result); err != nil {
			log.Printf("Invalid reply message format: %v", err)
			continue
		}

		correlationID := d.CorrelationId
		if replyChan, exists := o.pendingReplies[correlationID]; exists {
			replyChan <- result
		}
	}
}

func (o orderService) GetById(ctx context.Context, orderID uint, userID uint) (*models.Order, error) {
	return o.orderRepo.GetById(ctx, orderID, userID)
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
