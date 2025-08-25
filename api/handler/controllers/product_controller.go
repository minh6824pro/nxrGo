package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/services"
	"github.com/minh6824pro/nxrGO/internal/utils"
	customErr "github.com/minh6824pro/nxrGO/pkg/errors"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

type ProductController struct {
	service services.ProductService
}

func NewProductController(s services.ProductService) *ProductController {
	return &ProductController{service: s}
}

// Create godoc
// @Summary      Create a new product
// @Description  Create a new product
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        product  body      dto.CreateProductInput  true  "Create product request"
// @Success      201    id  "Success response with product data"
// @Router       /products [post]
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

// List  godoc
// @Summary      Get all product ( No preload relationship)
// @Description  Get all product ( No preload relationship)
// @Tags         products
// @Accept       json
// @Produce      json
// @Success      200    {array}  models.Product  "Success response with product data"
// @Router       /products [get]
func (pc *ProductController) List(ctx *gin.Context) {
	list, err := pc.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, list)
}

// GetByID  godoc
// @Summary      Get a product by id
// @Description  Get a product by id
// @Tags         products
// @Accept       json
// @Produce      json
// @Param id path string true "Product ID"
// @Success      200    {object}  models.Product  "Success response with product data"
// @Router       /products/{id} [get]
func (pc *ProductController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	product, err := pc.service.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			customErr.WriteError(ctx, customErr.NewError(customErr.ITEM_NOT_FOUND, "product not found", http.StatusNotFound, err))
			return
		}
		customErr.WriteError(ctx, err)
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

// ListProductQuery godoc
// @Summary      Get landing page products with query options
// @Description  Retrieve a paginated list of products for landing page with optional filters and sorting.
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        priceMin       query     number  false  "Minimum price filter (float64)"  example(10000)
// @Param        priceMax       query     number  false  "Maximum price filter (float64)"  example(500000)
// @Param        priceAsc       query     bool    false  "Sort by price ascending"         example(true)
// @Param        totalBuyDesc   query     bool    false  "Sort by total purchases descending" example(true)
// @Param        page           query     int     false  "Page number (starts from 0)"     example(0)
// @Param        pageSize       query     int     false  "Number of items per page"        example(16)
// @Success      200            {array}  CacheModel.ProductMiniCache "Success response with products and total count"
// @Router       /products/query [post]
func (pc *ProductController) ListProductQuery(ctx *gin.Context) {
	// Lấy query param
	priceMinStr := ctx.Query("priceMin") // string
	priceMaxStr := ctx.Query("priceMax")
	priceAscStr := ctx.Query("priceAsc")
	filterTotalBuyStr := ctx.Query("totalBuyDesc")
	pageStr := ctx.DefaultQuery("page", "0")
	pageSizeStr := ctx.DefaultQuery("pageSize", "16")

	// Parse float64 pointer
	var priceMin, priceMax *float64
	if priceMinStr != "" {
		v, err := strconv.ParseFloat(priceMinStr, 64)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "priceMin invalid"})
			return
		}
		priceMin = &v
	}
	if priceMaxStr != "" {
		v, err := strconv.ParseFloat(priceMaxStr, 64)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "priceMax invalid"})
			return
		}
		priceMax = &v
	}

	// Parse bool pointer
	var priceAsc, filterTotalBuy *bool
	if priceAscStr != "" {
		v, err := strconv.ParseBool(priceAscStr)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "priceAsc invalid"})
			return
		}
		priceAsc = &v
	}
	if filterTotalBuyStr != "" {
		v, err := strconv.ParseBool(filterTotalBuyStr)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "totalBuyDesc invalid"})
			return
		}
		filterTotalBuy = &v
	}

	// Parse int
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "page invalid"})
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "pageSize invalid"})
		return
	}
	if priceAsc != nil && filterTotalBuy != nil {
		customErr.WriteError(ctx, customErr.NewError(customErr.BAD_REQUEST, "Sorting is allowed only by totalBuyDesc or priceAsc", http.StatusBadRequest, err))
		return
	}
	// Gọi service
	products, total, err := pc.service.GetProductList(ctx, priceMin, priceMax, priceAsc, filterTotalBuy, page, pageSize)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		log.Println(err)
		return
	}

	ctx.JSON(200, gin.H{"data": products, "total": total})
}

func (pc *ProductController) ListProductManagement(ctx *gin.Context) {

	pageStr := ctx.DefaultQuery("page", "0")
	pageSizeStr := ctx.DefaultQuery("pageSize", "16")

	// Parse int
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "page invalid"})
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "pageSize invalid"})
		return
	}

	products, total, err := pc.service.GetProductListManagement(ctx, nil, nil, nil, nil, page, pageSize)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		log.Println(err)
		return
	}

	ctx.JSON(200, gin.H{"data": products, "total": total})
}
