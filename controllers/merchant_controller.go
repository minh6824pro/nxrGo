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

func (c *MerchantController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	m, err := c.service.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		customErr.WriteError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, m)
}

func (c *MerchantController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.service.Delete(ctx.Request.Context(), uint(id)); err != nil {
		customErr.WriteError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (c *MerchantController) List(ctx *gin.Context) {
	list, err := c.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, list)
}

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
