package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/services"
	"github.com/minh6824pro/nxrGO/internal/utils"
	customErr "github.com/minh6824pro/nxrGO/pkg/errors"
	"io"
	"net/http"
	"strconv"
)

type ProductVariantController struct {
	service services.ProductVariantService
}

func NewProductVariantController(service services.ProductVariantService) *ProductVariantController {
	return &ProductVariantController{
		service: service,
	}
}

// Create godoc
// @Summary      Create product variant
// @Description  Create product variant
// @Tags         Product Variants
// @Accept       json
// @Produce      json
// @Param        productVariant  body      dto.CreateProductVariantInput  true  "Product Variant Input"
// @Success      201  {object}   models.ProductVariant
// @Router       /product_variants [post]
func (pc *ProductVariantController) Create(c *gin.Context) {
	var input dto.CreateProductVariantInput
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
		customErr.WriteError(c, err)
		return
	}
	create, err := pc.service.Create(c, input)
	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": create})
}

// IncreaseStock godoc
// @Summary		Increase product variant quantity
// @Description	Increase product variant quantity
// @Tags		Product Variants
// @Accept		json
// @Produce		json
// @Param id path string true "Product Variant ID"
// @Param		updateStockRequest body dto.UpdateStockRequest true "Update Stock Request"
// @Success		200 {object} models.ProductVariant
// @Router		/product_variants/{id}/increase_stock [patch]
func (pc *ProductVariantController) IncreaseStock(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var input dto.UpdateStockRequest
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
		customErr.WriteError(c, err)
		return
	}

	updated, err := pc.service.IncreaseStock(c, uint(id), input)
	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": updated})
}

// DecreaseStock godoc
// @Summary		Decrease product variant quantity
// @Description	Decrease product variant quantity
// @Tags		Product Variants
// @Accept		json
// @Produce		json
// @Param		updateStockRequest body dto.UpdateStockRequest true "Update Stock Request"
// @Param id path string true "Product Variant ID"
// @Success		200 {object} models.ProductVariant
// @Router		/product_variants/{id}/decrease_stock [patch]
func (pc *ProductVariantController) DecreaseStock(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var input dto.UpdateStockRequest
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
		customErr.WriteError(c, err)
		return
	}

	updated, err := pc.service.DecreaseStock(c, uint(id), input)
	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": updated})
}

// ListByIds godoc
// @Summary		Get products for cart info
// @Description	Get product for cart info
// @Tags		Product Variants
// @Accept		json
// @Produce		json
// @Param		ListProductVariantIds body dto.ListProductVariantIds true "List contain ids need to fetch"
// @Success		200 {object} models.ProductVariant
// @Router		/product_variants/listbyids [get]
func (pc *ProductVariantController) ListByIds(c *gin.Context) {
	var input dto.ListProductVariantIds

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
	ids, err := pc.service.ListByIds(c, input)
	if err != nil {
		customErr.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ids})
}
