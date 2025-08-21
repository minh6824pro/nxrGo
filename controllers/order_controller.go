package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/services"
	"github.com/minh6824pro/nxrGO/utils"
	"io"
	"log"
	"net/http"
	"strconv"
)

type OrderController struct {
	service services.OrderService
}

func NewOrderController(service services.OrderService) *OrderController {
	return &OrderController{service}
}

// Create godoc
// @Summary      Create a new order
// @Description  Create a new order with the provided items. Requires authentication.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        order  body      dto.CreateOrderInput  true  "Create order request"
// @Success      201    {object}  dto.CreateOrderResponse  "Success response with order data"
// @Router       /orders [post]
func (o *OrderController) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		customErr.WriteError(c, customErr.NewError(
			customErr.UNAUTHORIZED,
			"Unauthorized",
			http.StatusUnauthorized,
			nil))
		return
	}
	var input dto.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		if errors.Is(err, io.EOF) {
			customErr.WriteError(c, customErr.NewError(
				customErr.BAD_REQUEST,
				"Request body is empty",
				http.StatusBadRequest,
				err,
			))
			return
		}
		if utils.HandleValidationError(c, err) {
			return
		}
		customErr.WriteError(c, customErr.NewError(customErr.BAD_REQUEST, "Invalid request body", http.StatusBadRequest, err))
		return
	}
	input.UserID = userID.(uint)

	createdOrder, err := o.service.Create(c, input)

	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, createdOrder)
}

// GetById godoc
// @Summary      Get Order that associate with user by order id
// @Description  Get order by order id. Requires authentication.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param id path string true "Order ID"
// @Success      200    {object}  models.Order  "Success response with order data"
// @Router       /orders/{id} [get]
func (o *OrderController) GetById(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		customErr.WriteError(c, customErr.NewError(
			customErr.UNAUTHORIZED,
			"Unauthorized",
			http.StatusUnauthorized,
			nil))
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	order, err := o.service.GetById(c, uint(id), userID.(uint))
	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": order})
}

// UpdateDb godoc
// @Summary     Mock Call for update product variant quantity in DB and clean draft orders that are not converted to real orders
// @Description	Update product variant quantity in DB and clean draft orders that are not converted to real orders
// @Tags		database
// @Accept		json
// Produce		json
// @Security	BearerAuth
// @Router		/orders/updatedb [post]
func (o *OrderController) UpdateDb(c *gin.Context) {
	err := o.service.UpdateQuantity(c)
	if err != nil {
		customErr.WriteError(c, err)
		log.Println(err)
		return
	}
	return
}

// UpdateOrderStatus godoc
// @Summary      Update order status
// @Description  Update order status. Requires Admin Role.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param id path string true "Order ID"
// @Param        order  body      dto.OrderEventRequest  true  "Create order request"
// @Success      200    {array}  models.Order  "Success response with order data"
// @Router       /orders/status/{id} [patch]
func (o *OrderController) UpdateOrderStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var event dto.OrderEventRequest
	if err := c.ShouldBindJSON(&event); err != nil {
		if errors.Is(err, io.EOF) {
			customErr.WriteError(c, customErr.NewError(
				customErr.BAD_REQUEST,
				"Request body is empty",
				http.StatusBadRequest,
				err,
			))
			return
		}
		if utils.HandleValidationError(c, err) {
			return
		}
		customErr.WriteError(c, err)
		return
	}
	order, err := o.service.UpdateOrderStatus(c, uint(id), event.Event)
	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": order})
}

// GetByStatus godoc
// @Summary      Get Order that associate with user by order id and status
// @Description  Get order by order id and status. Requires authentication.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        order  body      dto.OrderStatusRequest  true  "Create order request"
// @Success      200    {array}  models.Order  "Success response with order data"
// @Router       /orders/status [get]
func (o *OrderController) GetByStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		customErr.WriteError(c, customErr.NewError(
			customErr.UNAUTHORIZED,
			"Unauthorized",
			http.StatusUnauthorized,
			nil))
		return
	}
	var event dto.OrderStatusRequest
	if err := c.ShouldBindJSON(&event); err != nil {
		if errors.Is(err, io.EOF) {
			customErr.WriteError(c, customErr.NewError(
				customErr.BAD_REQUEST,
				"Request body is empty",
				http.StatusBadRequest,
				err,
			))
			return
		}
		if utils.HandleValidationError(c, err) {
			return
		}
		customErr.WriteError(c, err)
		return
	}
	orders, err := o.service.GetsByStatus(c, event.OrderStatus, userID.(uint))
	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": orders})
}

// ReBuy godoc
// @Summary      Create a new order from an existing order of same user
// @Description  Create a new order from an existing order of same user. Requires authentication.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param id path string true "Order ID"
// @Success      201    {object}  dto.CreateOrderResponse  "Success response with order data"
// @Router       /orders/rebuy/{id} [post]
func (o *OrderController) ReBuy(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		customErr.WriteError(c, customErr.NewError(
			customErr.UNAUTHORIZED,
			"Unauthorized",
			http.StatusUnauthorized,
			nil))
		return
	}
	orderId, _ := strconv.Atoi(c.Param("id"))

	reBuy, err := o.service.ReBuy(c, uint(orderId), userID.(uint))
	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, reBuy)

}

// List godoc
// @Summary      Get list of orders for current user
// @Description  Retrieve all orders for the authenticated user
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  dto.CreateOrderResponse  "List of orders"
// @Router       /orders [get]
func (o *OrderController) List(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		customErr.WriteError(c, customErr.NewError(
			customErr.UNAUTHORIZED,
			"Unauthorized",
			http.StatusUnauthorized,
			nil))
		return
	}
	list, err := o.service.ListByUserId(c, userID.(uint))
	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (o *OrderController) ChangePaymentMethod(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		customErr.WriteError(c, customErr.NewError(
			customErr.UNAUTHORIZED,
			"Unauthorized",
			http.StatusUnauthorized,
			nil))
		return
	}
	var payment dto.ChangePaymentMethodRequest
	if err := c.ShouldBindJSON(&payment); err != nil {

		if errors.Is(err, io.EOF) {
			customErr.WriteError(c, customErr.NewError(
				customErr.BAD_REQUEST,
				"Request body is empty",
				http.StatusBadRequest,
				err,
			))
			return
		}
		if utils.HandleValidationError(c, err) {
			return
		}
		customErr.WriteError(c, err)
		return
	}
	changed, err := o.service.ChangePaymentMethod(c, payment, userID.(uint))
	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": changed})
}

func (o *OrderController) ListByAdmin(c *gin.Context) {
	list, err := o.service.ListByAdmin(c)

	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// GET /shippingFee?merchantId=1&merchantId=2&lat=16.45&lat=10.77&lon=107.59&lon=106.70
func (o *OrderController) GetShippingFee(c *gin.Context) {
	merchantIDs := c.QueryArray("merchantId")
	userLat := c.Query("lat")
	userLon := c.Query("lon")

	if userLat == "" || userLon == "" || len(merchantIDs) == 0 {
		customErr.WriteError(c, customErr.NewError(customErr.BAD_REQUEST, "merchantId, lat and lon are required", http.StatusBadRequest, nil))
		return
	}

	var shippingFee []*dto.ShippingFeeResponse
	for _, mID := range merchantIDs {
		id, err := strconv.Atoi(mID)
		if err != nil {
			log.Println("Error converting merchantId to int", err, mID)
			continue
		}

		fee, err := o.service.CalculateShippingFee(c, uint(id), userLon, userLat)
		if err != nil {
			log.Println("Error calculating shipping fee", err, mID)
			continue
		}

		shippingFee = append(shippingFee, fee...)
	}

	c.JSON(http.StatusOK, gin.H{"data": shippingFee})
}
