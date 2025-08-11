package controllers

import (
	"errors"
	"github.com/minh6824pro/nxrGO/dto"
	"github.com/minh6824pro/nxrGO/utils"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/services"
)

type MerchantController struct {
	service services.MerchantService
}

func NewMerchantController(s services.MerchantService) *MerchantController {
	return &MerchantController{s}
}

// Create godoc
// @Summary      Create a new merchant
// @Description  Create a new merchant with the input payload
// @Tags         merchants
// @Accept       json
// @Produce      json
// @Param        merchant  body      dto.CreateMerchantInput  true  "Create merchant"
// @Success      201       {object}  dto.MerchantResponse
// @Router       /merchants [post]
func (c *MerchantController) Create(ctx *gin.Context) {
	var m dto.CreateMerchantInput
	if err := ctx.ShouldBindJSON(&m); err != nil {

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
	createMerchant, err := c.service.Create(ctx.Request.Context(), &m)
	if err != nil {
		customErr.WriteError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, createMerchant)
}

// GetByID godoc
// @Summary 	Get merchant by id
// @Description Get merchant by id
// @Tags		merchants
// @Accept		json
// @Produce		json
// @Param id path string true "Merchant ID"
// @Success 	200		{object}	dto.MerchantResponse
// @Router		/merchants/{id} [get]
func (c *MerchantController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	m, err := c.service.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		customErr.WriteError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, m)
}

// Delete godoc
// @Summary 	Delete merchant by id
// @Description Delete merchant by id
// @Tags		merchants
// @Accept		json
// @Produce		json
// @Param id path string true "Merchant ID"
// @Success 	200	"Deleted successfully"
// @Router		/merchants/{id} [delete]
func (c *MerchantController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.service.Delete(ctx.Request.Context(), uint(id)); err != nil {
		customErr.WriteError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// List godoc
// @Summary 	Get List merchant
// @Description	Get List merchant
// @Tags 		merchants
// @Accept		json
// @Produce		json
// @Success		200		{array}	dto.MerchantResponse "Success response with merchant data"
// @Router		/merchants [get]
func (c *MerchantController) List(ctx *gin.Context) {
	list, err := c.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, list)
}

// Patch godoc
// @Summary		Patch merchant by id
// @Description Patch merchant by id
// @Tags		merchants
// @Accept		json
// @Produce		json
// @Param id path string true "Merchant ID"
// @Param 		merchant body dto.UpdateMerchantInput true "Patch merchant request"
// @Success		200		{object}	dto.MerchantResponse "Success response with merchant data"
// @Router		/merchants/{id} [patch]
func (c *MerchantController) Patch(ctx *gin.Context) {
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
