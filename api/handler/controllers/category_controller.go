package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/services"
	"github.com/minh6824pro/nxrGO/internal/utils"
	customErr "github.com/minh6824pro/nxrGO/pkg/errors"
	"gorm.io/gorm"
	"io"
	"net/http"
	"strconv"
)

type CategoryController struct {
	service services.CategoryService
}

func NewCategoryController(s services.CategoryService) *CategoryController {
	return &CategoryController{s}
}

// Create godoc
// @Summary      Create a new category
// @Description  Create a new category with the input payload
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body      dto.CreateCategoryInput  true  "Create category"
// @Success      201       {object}  models.Category
// @Router       /categories [post]
func (c *CategoryController) Create(ctx *gin.Context) {
	var cat dto.CreateCategoryInput
	if err := ctx.ShouldBindJSON(&cat); err != nil {
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
	createCate, err := c.service.Create(ctx.Request.Context(), &cat)
	if err != nil {
		customErr.WriteError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, createCate)
}

// GetByID godoc
// @Summary      Get category by id
// @Description  Get category by id
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id path string true "Category ID"
// @Success      200 {object} dto.CategoryResponse
// @Router       /categories/{id} [get]
func (c *CategoryController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	cat, err := c.service.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		customErr.WriteError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, cat)
}

func (c *CategoryController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var cat models.Category
	if err := ctx.ShouldBindJSON(&cat); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat.ID = uint(id)

	if err := c.service.Update(ctx.Request.Context(), &cat); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Category with id %d not found", id)})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, cat)
}

// Delete godoc
// @Summary      Delete category by id
// @Description  Delete category by id
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id path string true "Category ID"
// @Router       /categories/{id} [delete]
func (c *CategoryController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.service.Delete(ctx.Request.Context(), uint(id)); err != nil {
		customErr.WriteError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// List godoc
// @Summary      Get list of categories
// @Description  Get list of categories
// @Tags         categories
// @Accept       json
// @Produce      json
// @Success      200 {array} dto.CategoryResponse "Success response with category data"
// @Router       /categories [get]
func (c *CategoryController) List(ctx *gin.Context) {
	list, err := c.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, list)
}

// Patch godoc
// @Summary      Patch category by id
// @Description  Patch category by id
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id path string true "Category ID"
// @Param        category body dto.UpdateCategoryInput true "Patch category request"
// @Success      200 {object} dto.CategoryResponse "Success response with category data"
// @Router       /categories/{id} [patch]
func (c *CategoryController) Patch(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))

	var input dto.UpdateCategoryInput
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
