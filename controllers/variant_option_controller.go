package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/services"
	"github.com/minh6824pro/nxrGO/utils"
	"io"
	"net/http"
	"strconv"
)

type VariantOptionController struct {
	service services.VariantOptionService
}

func NewVariantOptionController(service services.VariantOptionService) *VariantOptionController {
	return &VariantOptionController{service}
}

// GetByID godoc
// @Summary      Get variant option by id
// @Description  Get variant option by id
// @Tags         variant-options
// @Accept       json
// @Produce      json
// @Param        id path string true "Variant Option ID"
// @Success      200 {object} models.VariantOption
// @Router       /variant_options/{id} [get]
func (c *VariantOptionController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	m, err := c.service.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		customErr.WriteError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, m)
}

// Create godoc
// @Summary      Create a new variant option
// @Description  Create a new variant option with input payload
// @Tags         variant-options
// @Accept       json
// @Produce      json
// @Param        variantOption body dto.CreateVariantOptionInput true "Create variant option payload"
// @Success      201 {object} models.VariantOption
// @Router       /variant_options [post]
func (c *VariantOptionController) Create(ctx *gin.Context) {
	var v dto.CreateVariantOptionInput
	if err := ctx.ShouldBind(&v); err != nil {
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
		return
	}
	createdVariantOption, err := c.service.Create(ctx.Request.Context(), &v)
	if err != nil {
		customErr.WriteError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, createdVariantOption)
}

// List godoc
// @Summary      Get list of variant options
// @Description  Get list of variant options
// @Tags         variant-options
// @Accept       json
// @Produce      json
// @Success      200 {array} models.VariantOption
// @Router       /variant_options [get]
func (c *VariantOptionController) List(ctx *gin.Context) {
	list, err := c.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, list)
}

// Delete godoc
// @Summary      Delete variant option by id
// @Description  Delete variant option by id
// @Tags         variant-options
// @Accept       json
// @Produce      json
// @Param        id path string true "Variant Option ID"
// @Success      200 "Deleted successfully"
// @Router       /variant_options/{id} [delete]
func (c *VariantOptionController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.service.Delete(ctx.Request.Context(), uint(id)); err != nil {
		customErr.WriteError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Patch godoc
// @Summary      Patch variant option by id
// @Description  Patch variant option by id
// @Tags         variant-options
// @Accept       json
// @Produce      json
// @Param        id path string true "Variant Option ID"
// @Param        variantOption body dto.UpdateVariantOptionInput true "Patch variant option request"
// @Success      200 {object} models.VariantOption
// @Router       /variant_options/{id} [patch]
func (c *VariantOptionController) Patch(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))

	var input dto.UpdateVariantOptionInput
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
