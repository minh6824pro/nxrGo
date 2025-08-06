package controllers

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/minh6824pro/nxrGO/dto"
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
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			validationDetails := make([]map[string]string, len(ve))
			for i, fe := range ve {
				validationDetails[i] = map[string]string{
					"field":   fe.Field(),
					"message": fmt.Sprintf("%s is %s", fe.Field(), fe.Tag()),
				}
			}

			customErr.WriteError(ctx, customErr.NewErrorWithMeta(
				customErr.VALIDATION_ERROR,
				"Invalid request body",
				http.StatusBadRequest,
				err,
				map[string]any{"details": validationDetails},
			))
			return
		}
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

//func (c *MerchantController) Update(ctx *gin.Context) {
//	id, _ := strconv.Atoi(ctx.Param("id"))
//	var m models.Merchant
//	if err := ctx.ShouldBindJSON(&m); err != nil {
//		if errors.Is(err, io.EOF) {
//			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Request body is empty"})
//
//			return
//		}
//		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//	m.ID = uint(id)
//
//	if err := c.service.Update(ctx.Request.Context(), &m); err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Merchant with id %d not found", id)})
//			return
//		}
//		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	ctx.JSON(http.StatusOK, m)
//}

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

//func (c *MerchantController) Patch(ctx *gin.Context) {
//	id, _ := strconv.Atoi(ctx.Param("id"))
//
//	var updates map[string]interface{}
//	if err := ctx.ShouldBindJSON(&updates); err != nil {
//		customErr.WriteError(ctx, err)
//		return
//	}
//
//	category, err := c.service.Patch(ctx.Request.Context(), uint(id), updates)
//	if err != nil {
//		customErr.WriteError(ctx, err)
//		return
//	}
//
//	ctx.JSON(http.StatusOK, category)
//}

func (c *MerchantController) Patch(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))

	var input dto.UpdateMerchantInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			validationDetails := make([]map[string]string, len(ve))
			for i, fe := range ve {
				validationDetails[i] = map[string]string{
					"field":   fe.Field(),
					"message": fmt.Sprintf("%s is %s", fe.Field(), fe.Tag()),
				}
			}

			customErr.WriteError(ctx, customErr.NewErrorWithMeta(
				customErr.VALIDATION_ERROR,
				"Invalid request body",
				http.StatusBadRequest,
				err,
				map[string]any{"details": validationDetails},
			))
			return

		}
		customErr.WriteError(ctx, err)
		return
	}

	updated, err := c.service.Patch(ctx.Request.Context(), uint(id), &input)
	if err != nil {
		customErr.WriteError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, updated)
}
