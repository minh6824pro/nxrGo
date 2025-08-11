package controllers

import (
	"errors"
	"fmt"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/utils"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/services"
	"gorm.io/gorm"
)

type BrandController struct {
	service services.BrandService
}

func NewBrandController(s services.BrandService) *BrandController {
	return &BrandController{s}
}

// Create godoc
// @Summary      Create a new brand
// @Description  Create a new brand with the input payload
// @Tags         brands
// @Accept       json
// @Produce      json
// @Param        brand  body      dto.CreateBrandInput  true  "Create brand payload"
// @Success      201    {object}  dto.BrandResponse
// @Router       /brands [post]
func (c *BrandController) Create(ctx *gin.Context) {
	var b dto.CreateBrandInput
	if err := ctx.ShouldBindJSON(&b); err != nil {

		if errors.Is(err, io.EOF) {
			customErr.WriteError(ctx, customErr.NewError(
				customErr.BAD_REQUEST,
				"Request body is empty",
				http.StatusBadRequest,
				err,
			))
			return
		}
		if utils.HandleValidationError(ctx, err) {
			return
		}
		customErr.WriteError(ctx, err)
		return
	}
	createdBrand, err := c.service.Create(ctx.Request.Context(), &b)
	if err != nil {
		customErr.WriteError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, createdBrand)
}

// GetByID godoc
// @Summary      Get brand by id
// @Description  Get brand by id
// @Tags         brands
// @Accept       json
// @Produce      json
// @Param        id path string true "Brand ID"
// @Success      200 {object} dto.BrandResponse
// @Router       /brands/{id} [get]
func (c *BrandController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	b, err := c.service.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		customErr.WriteError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, b)
}

func (c *BrandController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var b models.Brand
	if err := ctx.ShouldBindJSON(&b); err != nil {
		if errors.Is(err, io.EOF) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Request body is empty"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b.ID = uint(id)

	if err := c.service.Update(ctx.Request.Context(), &b); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Brand with id %d not found", id)})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, b)
}

// Delete godoc
// @Summary      Delete brand by id
// @Description  Delete brand by id
// @Tags         brands
// @Accept       json
// @Produce      json
// @Param        id path string true "Brand ID"
// @Success      200 "Deleted successfully"
// @Router       /brands/{id} [delete]
func (c *BrandController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.service.Delete(ctx.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Brand with id %d not found", id)})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// List godoc
// @Summary      Get list of brands
// @Description  Get list of brands
// @Tags         brands
// @Accept       json
// @Produce      json
// @Success      200 {array} dto.BrandResponse "Success response with brand data"
// @Router       /brands [get]
func (c *BrandController) List(ctx *gin.Context) {
	list, err := c.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, list)
}

// Patch godoc
// @Summary      Patch brand by id
// @Description  Patch brand by id
// @Tags         brands
// @Accept       json
// @Produce      json
// @Param        id path string true "Brand ID"
// @Param        brand body dto.UpdateBrandInput true "Patch brand request"
// @Success      200 {object} dto.BrandResponse "Success response with brand data"
// @Router       /brands/{id} [patch]
func (c *BrandController) Patch(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))

	var input dto.UpdateMerchantInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		if errors.Is(err, io.EOF) {
			customErr.WriteError(ctx, customErr.NewError(
				customErr.BAD_REQUEST,
				"Request body is empty",
				http.StatusBadRequest,
				err,
			))
			return
		}
		utils.HandleValidationError(ctx, err)
		return
	}

	updated, err := c.service.Patch(ctx.Request.Context(), uint(id), &input)
	if err != nil {
		customErr.WriteError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, updated)
}
