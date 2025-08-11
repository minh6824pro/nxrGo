package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/services"
	"github.com/minh6824pro/nxrGO/utils"
	"net/http"
	"strconv"
)

type OrderController struct {
	service services.OrderService
}

func NewOrderController(service services.OrderService) *OrderController {
	return &OrderController{service}
}

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
		customErr.WriteError(c, customErr.NewError(
			customErr.BAD_REQUEST,
			"invalid request body", http.StatusBadRequest, err,
		))
		if utils.HandleValidationError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	input.UserID = userID.(uint)

	createdOrder, err := o.service.Create(c, input)

	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "success", "data": createdOrder})
}

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

func (o *OrderController) UpdateDb(c *gin.Context) {
	err := o.service.UpdateQuantity(c)
	if err != nil {
		return
	}
	return
}
