package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/services"
	"github.com/minh6824pro/nxrGO/utils"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type ProductController struct {
	service services.ProductService
}

func NewProductController(s services.ProductService) *ProductController {
	return &ProductController{s}
}

func (pc *ProductController) Create(c *gin.Context) {
	var input dto.CreateProductInput
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

	id, err := pc.service.Create(c.Request.Context(), input)
	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "success", "id": id})
}

func (pc *ProductController) List(ctx *gin.Context) {
	list, err := pc.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, list)
}

func (pc *ProductController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	product, err := pc.service.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Product with id %d not found", id)})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, product)
}

func (pc *ProductController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := pc.service.Delete(ctx.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Product with id %d not found", id)})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return

	}
	ctx.JSON(http.StatusOK, gin.H{"messaage": "deleted"})
}
