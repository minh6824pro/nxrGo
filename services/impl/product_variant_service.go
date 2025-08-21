package impl

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/cache"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/event"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/models/CacheModel"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
	"github.com/minh6824pro/nxrGO/utils"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"time"
)

type productVariantService struct {
	productRepo         repositories.ProductRepository
	productVariantRepo  repositories.ProductVariantRepository
	productVariantCache cache.ProductVariantRedis
	updateStockAgg      *event.UpdateStockAggregator
}

func NewProductVariantService(productRepo repositories.ProductRepository, productVariantRepo repositories.ProductVariantRepository,
	productVariantCache cache.ProductVariantRedis, updateStockAgg *event.UpdateStockAggregator) services.ProductVariantService {
	return &productVariantService{
		productRepo:         productRepo,
		productVariantRepo:  productVariantRepo,
		productVariantCache: productVariantCache,
		updateStockAgg:      updateStockAgg,
	}
}

func (p productVariantService) Create(ctx context.Context, input dto.CreateProductVariantInput) (*models.ProductVariant, error) {
	log.Print("----------------------")
	product, err := p.productRepo.GetByID(ctx, input.ProductID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, customErr.NewError(customErr.ITEM_NOT_FOUND, fmt.Sprintf("Product %d not found", input.ProductID), http.StatusBadRequest, err)
	}
	image := ""
	if input.Image != "" {
		image = input.Image
	}

	var optionValues []models.VariantOptionValue
	for _, v := range input.OptionValues {
		optionValues = append(optionValues, models.VariantOptionValue{
			OptionID: v.OptionID,
			Value:    v.Value,
		})
	}

	productVariant := &models.ProductVariant{
		ProductID:    product.ID,
		Quantity:     input.Quantity,
		Price:        input.Price,
		Image:        image,
		OptionValues: optionValues,
	}

	createdProductVariant, err := p.productVariantRepo.Create(ctx, productVariant)
	if err != nil {
		return nil, err
	}
	return createdProductVariant, nil
}

func (p productVariantService) GetByID(ctx context.Context, id uint) (*models.ProductVariant, error) {
	//TODO implement me
	panic("implement me")
}

func (p productVariantService) List(ctx context.Context) ([]models.ProductVariant, error) {
	//TODO implement me
	panic("implement me")
}

func (p productVariantService) Delete(ctx context.Context, id uint) error {
	//TODO implement me
	panic("implement me")
}

func (p productVariantService) Patch(ctx context.Context, productVariant *models.ProductVariant) error {
	//TODO implement me
	panic("implement me")
}

func (p productVariantService) IncreaseStock(c *gin.Context, id uint, input dto.UpdateStockRequest) (*models.ProductVariant, error) {

	if input.Quantity == 0 {
		return nil, customErr.NewError(customErr.INVALID_INPUT, "Quantity can't be 0", http.StatusBadRequest, nil)
	}
	pv, err := p.productVariantRepo.GetByID(c, id)
	if err != nil {
		return nil, err
	}
	if input.Quantity > 0 {
		err := p.productVariantRepo.IncreaseQuantity(c, map[uint]uint{
			id: input.Quantity,
		})
		if err != nil {
			return nil, err
		}
		pv.Quantity += uint(input.Quantity)
		// Remove cache
		err = p.productVariantCache.DeleteProductVariantHash(id)
		if err != nil {
			log.Println("Cache 1", err)
		}
	} else {
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected error 3", http.StatusInternalServerError, nil)
	}

	return pv, nil
}

func (p productVariantService) DecreaseStock(c *gin.Context, id uint, input dto.UpdateStockRequest) (*models.ProductVariant, error) {

	if input.Quantity == 0 {
		return nil, customErr.NewError(customErr.INVALID_INPUT, "Quantity can't be 0", http.StatusBadRequest, nil)
	}
	pv, err := p.productVariantRepo.GetByID(c, id)
	if err != nil {
		return nil, err
	}
	err = p.productVariantCache.PingRedis(c)
	if err != nil {
		// Redis die -> db
		pv, err = p.productVariantRepo.CheckAndDecreaseStock(c, id, input.Quantity)
		if err != nil {
			return nil, err
		}
		return pv, nil
	} else {
		// Check exists in redis cache
		pvCache, err := p.productVariantCache.GetProductVariantHash(id)
		if err != nil {
			log.Print(err)
		}
		// If not get cache
		if len(pvCache) == 0 {
			pv, err := p.productVariantRepo.GetByIDForRedisCache(c, id)
			if err != nil {
				return nil, err
			}

			err = p.productVariantCache.SaveProductVariantHash(*pv)
			if err != nil {
				return nil, err
			}
			// Retrieve again from cache after save
			pvCache, err = p.productVariantCache.GetProductVariantHash(id)
			if err != nil {
				return nil, err
			}
		}
		quantityStr, ok := pvCache["quantity"]
		if !ok {
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected Error 1", http.StatusInternalServerError, err)
		}
		quantityInt, err := strconv.Atoi(quantityStr)
		if err != nil {
			return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Unexpected Error 2", http.StatusInternalServerError, err)
		}
		if input.Quantity <= uint(quantityInt) {
			err := p.productVariantRepo.DecreaseQuantity(c, map[uint]uint{
				id: input.Quantity,
			})
			if err != nil {
				return nil, err
			}
			// Remove cache
			err = p.productVariantCache.DeleteProductVariantHash(id)
			if err != nil {
				return nil, err
			}
			pv.Quantity += uint(-input.Quantity)
			return pv, nil

		} else {
			return nil, customErr.NewError(customErr.BAD_REQUEST, fmt.Sprintf("Cannot reduce stock: quantity available is less than requested,available %d", quantityInt), http.StatusBadRequest, nil)
		}
	}

}

//	func (p productVariantService) loadAndCacheProductVariants(ctx context.Context, ids []uint) ([]models.ProductVariant, error) {
//		// Get List from Db
//		variants, err := p.productVariantRepo.GetByIDSForRedisCache(ctx, ids)
//		if err != nil {
//			return nil, err
//		}
//		// If DB returns fewer records, return ITEM_NOT_FOUND error for missing ids
//		if len(variants) != len(ids) {
//			found := map[uint]bool{}
//			for _, v := range variants {
//				found[v.ID] = true
//			}
//			for _, id := range ids {
//				if !found[id] {
//					return nil, customErr.NewError(
//						customErr.ITEM_NOT_FOUND,
//						fmt.Sprintf("Product variant not found: %d", id),
//						http.StatusBadRequest, nil,
//					)
//				}
//			}
//		}
//
//		// Save Redis cache
//		for _, pv := range variants {
//			if err := p.productVariantCache.SaveProductVariantHash(pv); err != nil {
//				log.Printf("warning: save productVariant %d to redis failed: %v", pv.ID, err)
//			}
//		}
//
//		return variants, nil
//	}
func (p productVariantService) ListByIds(ctx context.Context, list dto.ListProductVariantIds) ([]dto.VariantCartInfoResponse, error) {
	productsVariant, err := p.productVariantRepo.ListByIds(ctx, list)
	if err != nil {
		return nil, err
	}
	return MapToVariantCartResponse(ctx, productsVariant)
}
func (p productVariantService) CheckAndCacheProductVariants(ctx context.Context, ids []uint) ([]CacheModel.VariantLite, error) {

	var result []CacheModel.VariantLite
	var missingIDs []uint

	// 1. Check Redis cache trước
	for _, id := range ids {
		cache, err := p.productVariantCache.GetProductVariantHash(id)
		if err != nil {
			log.Printf("warning: get productVariant %d from redis failed: %v", id, err)
			missingIDs = append(missingIDs, id)
			continue
		}

		if len(cache) == 0 {
			// key không tồn tại
			missingIDs = append(missingIDs, id)
			continue
		}

		// Lấy id, quantity, price từ hash
		idVal, _ := strconv.ParseUint(cache["id"], 10, 64)
		quantityVal, _ := strconv.ParseUint(cache["quantity"], 10, 64)
		priceVal, _ := strconv.ParseInt(cache["price"], 10, 32)

		result = append(result, CacheModel.VariantLite{
			ID:       uint(idVal),
			Quantity: uint(quantityVal),
			Price:    float64(priceVal),
		})
	}

	// 2. Nếu tất cả có cache → trả về luôn
	if len(missingIDs) == 0 {
		return result, nil
	}

	// 3. Lấy các missing IDs từ DB
	variants, err := p.productVariantRepo.GetByIDSForRedisCache(ctx, missingIDs)
	if err != nil {
		return nil, err
	}

	// 4. Kiểm tra DB trả về đầy đủ không
	if len(variants) != len(missingIDs) {
		found := map[uint]bool{}
		for _, v := range variants {
			found[v.ID] = true
		}
		for _, id := range missingIDs {
			if !found[id] {
				return nil, customErr.NewError(
					customErr.ITEM_NOT_FOUND,
					fmt.Sprintf("Product variant not found: %d", id),
					http.StatusBadRequest, nil,
				)
			}
		}
	}

	// 5. Lưu Redis cache cho các variant vừa lấy và append vào result
	for _, pv := range variants {
		if err := p.productVariantCache.SaveProductVariantHash(pv); err != nil {
			log.Printf("warning: save productVariant %d to redis failed: %v", pv.ID, err)
		}

		result = append(result, CacheModel.VariantLite{
			ID:       pv.ID,
			Quantity: pv.Quantity,
			Price:    pv.Price,
		})
	}

	return result, nil
}

func MapToVariantCartResponse(ctx context.Context, list []models.ProductVariant) ([]dto.VariantCartInfoResponse, error) {
	var result []dto.VariantCartInfoResponse
	for _, productVariant := range list {
		optStr := ""
		for i, opt := range productVariant.OptionValues {
			if i != len(productVariant.OptionValues)-1 {
				optStr += opt.Option.Name + ": " + opt.Value + ", "

			} else {
				optStr += opt.Option.Name + ": " + opt.Value
			}
		}
		cartInfo := dto.VariantCartInfoResponse{
			ID:           productVariant.ID,
			Price:        productVariant.Price,
			Quantity:     productVariant.Quantity,
			ProductName:  productVariant.Product.Name,
			ProductID:    productVariant.Product.ID,
			Option:       optStr,
			MerchantName: productVariant.Product.Merchant.Name,
			MerchantID:   productVariant.Product.Merchant.ID,
			Timestamp:    time.Now().Unix(),
			Image:        productVariant.Image,
		}
		signature := utils.GenerateProductVariantSignature(cartInfo.ID, cartInfo.Price, cartInfo.MerchantID, cartInfo.Timestamp)
		cartInfo.Signature = signature
		result = append(result, cartInfo)
	}
	return result, nil
}
