package impl

import (
	"context"
	"errors"
	"fmt"
	"github.com/minh6824pro/nxrGO/cache"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/models/CacheModel"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func NewProductService(db *gorm.DB, productRepo repositories.ProductRepository, brandRepo repositories.BrandRepository, merchanRepo repositories.MerchantRepository,
	categoryRepo repositories.CategoryRepository, productVariantRepo repositories.ProductVariantRepository, variantOptionValueRepo repositories.VariantOptionValueRepository,
	variantOptionRepo repositories.VariantOptionRepository, productCache cache.ProductCacheService,
	productVariantService services.ProductVariantService) services.ProductService {
	return &productService{
		db:                     db,
		productRepo:            productRepo,
		brandRepo:              brandRepo,
		merchantRepo:           merchanRepo,
		categoryRepo:           categoryRepo,
		productVariantRepo:     productVariantRepo,
		variantOptionValueRepo: variantOptionValueRepo,
		variantOptionRepo:      variantOptionRepo,
		productCacheService:    productCache,
		productVariantService:  productVariantService,
	}
}

type productService struct {
	db                     *gorm.DB
	productRepo            repositories.ProductRepository
	brandRepo              repositories.BrandRepository
	merchantRepo           repositories.MerchantRepository
	categoryRepo           repositories.CategoryRepository
	productVariantRepo     repositories.ProductVariantRepository
	variantOptionValueRepo repositories.VariantOptionValueRepository
	variantOptionRepo      repositories.VariantOptionRepository
	productVariantService  services.ProductVariantService
	productCacheService    cache.ProductCacheService
}

//	func (productService *productService) Create(ctx context.Context, input dto.CreateProductInput) (*models.Product, error) {
//		// Xử lý Brand
//		brandID, err := productService.getOrCreateBrand(ctx, input.BrandID, input.BrandName)
//		if err != nil {
//			return nil, err
//		}
//
//		// Xử lý Category
//		categoryID, err := productService.getOrCreateCategory(ctx, input.CategoryID, input.CategoryName)
//		if err != nil {
//			return nil, err
//		}
//
//		// Xử lý Merchant
//		merchantID, err := productService.getOrCreateMerchant(ctx, input.MerchantID, input.MerchantName)
//		if err != nil {
//			return nil, err
//		}
//
//		// Tạo Product
//		product := &models.Product{
//			Name:          input.Name,
//			Description:   input.Description,
//			Image:         input.Image,
//			BrandID:       brandID,
//			CategoryID:    categoryID,
//			MerchantID:    merchantID,
//			AverageRating: 0,
//			NumberRating:  0,
//			Active:        true,
//		}
//
//		// Tạo trong DB
//		createdProduct, err := productService.productRepo.Create(ctx, product)
//		if err != nil {
//			return nil, err
//		}
//
//		// Xử lý Variant
//		for _, variant := range input.Variants {
//			productVariantCreated := &models.ProductVariant{
//				Quantity:  variant.Quantity,
//				Price:     variant.Price,
//				Image:     variant.Image,
//				ProductID: createdProduct.ID,
//			}
//			productVariantCreated, err := productService.productVariantRepo.Create(ctx, productVariantCreated)
//			if err != nil {
//				return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Error creating product variant ", http.StatusInternalServerError, err)
//
//			}
//			for _, option := range variant.OptionValues {
//				variantOptionValue := models.VariantOptionValue{
//					VariantID: productVariantCreated.ID,
//					OptionID:  option.OptionID,
//					Value:     option.Value,
//				}
//				if _, err := productService.variantOptionValueRepo.Create(ctx, &variantOptionValue); err != nil {
//					return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Error creating variant option value", http.StatusInternalServerError, err)
//				}
//			}
//		}
//		return createdProduct, nil
//	}
func (productService *productService) Create(ctx context.Context, input dto.CreateProductInput) (*uint, error) {
	// ✅ Bắt đầu transaction
	tx := productService.db.Begin()
	if tx.Error != nil {
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Failed to start transaction", http.StatusInternalServerError, tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	//// ✅ Truyền tx vào các hàm getOrCreate*
	//brandID, err := productService.getOrCreateBrand(ctx, tx, input.BrandID, input.BrandName)
	//if err != nil {
	//	tx.Rollback()
	//	return nil, err
	//}
	if _, err := productService.brandRepo.GetByIDTx(ctx, tx, *input.BrandID); err != nil {
		tx.Rollback()
		return nil, err
	}

	//categoryID, err := productService.getOrCreateCategory(ctx, tx, input.CategoryID, input.CategoryName)
	//if err != nil {
	//	tx.Rollback()
	//	return nil, err
	//}
	if _, err := productService.categoryRepo.GetByIDTx(ctx, tx, *input.CategoryID); err != nil {
		tx.Rollback()
		return nil, err
	}
	//merchantID, err := productService.getOrCreateMerchant(ctx, tx, input.MerchantID, input.MerchantName)
	//if err != nil {
	//	tx.Rollback()
	//	return nil, err
	//}

	if _, err := productService.merchantRepo.GetByIDTx(ctx, tx, *input.MerchantID); err != nil {
		tx.Rollback()
		return nil, err
	}

	product := &models.Product{
		Name:          input.Name,
		Description:   input.Description,
		Image:         input.Image,
		BrandID:       *input.BrandID,
		CategoryID:    *input.CategoryID,
		MerchantID:    *input.MerchantID,
		AverageRating: 0,
		NumberRating:  0,
		Active:        true,
	}

	createdProduct, err := productService.productRepo.CreateWithTx(ctx, tx, product)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, variant := range input.Variants {
		productVariantCreated := &models.ProductVariant{
			Quantity:  variant.Quantity,
			Price:     variant.Price,
			Image:     variant.Image,
			ProductID: createdProduct.ID,
		}

		productVariantCreated, err = productService.productVariantRepo.CreateWithTx(ctx, tx, productVariantCreated)
		if err != nil {
			tx.Rollback()
			return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Error creating product variant", http.StatusInternalServerError, err)
		}

		for _, option := range variant.OptionValues {
			_, err := productService.variantOptionRepo.GetByIDTx(ctx, tx, option.OptionID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				return nil, customErr.NewError(customErr.INVALID_INPUT, fmt.Sprintf("Option ID %d not found", option.OptionID), http.StatusBadRequest, nil)
			}
			if err != nil {
				tx.Rollback()
				return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Error checking option existence", http.StatusInternalServerError, err)
			}
			variantOptionValue := &models.VariantOptionValue{
				VariantID: productVariantCreated.ID,
				OptionID:  option.OptionID,
				Value:     option.Value,
			}

			if _, err := productService.variantOptionValueRepo.CreateWithTx(ctx, tx, variantOptionValue); err != nil {
				tx.Rollback()
				return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Error creating variant option value", http.StatusInternalServerError, err)
			}
		}
	}

	// ✅ Commit transaction nếu mọi thứ đều OK
	if err := tx.Commit().Error; err != nil {
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Failed to commit transaction", http.StatusInternalServerError, err)
	}
	//if err := productService.db.WithContext(ctx).
	//	Preload("Merchant").
	//	Preload("Brand").
	//	Preload("Category").
	//	Preload("Variants").
	//	Preload("Variants.OptionValues").
	//	Preload("Variants.OptionValues.Option").
	//	First(&createdProduct, createdProduct.ID).Error; err != nil {
	//	return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Failed to load full product", http.StatusInternalServerError, err)
	//}

	return &createdProduct.ID, nil
}

func (productService *productService) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	return productService.productRepo.GetByID(ctx, id)
}

func (productService *productService) List(ctx context.Context) ([]models.Product, error) {
	return productService.productRepo.List(ctx)

}

func (productService *productService) Delete(ctx context.Context, id uint) error {
	_, err := productService.productRepo.GetByID(ctx, id)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gorm.ErrRecordNotFound
		}
		return err
	}

	return productService.productRepo.Delete(ctx, id)

}

func (productService *productService) Patch(ctx context.Context, product *models.Product) error {
	//TODO implement me
	panic("implement me")
}

func (productService *productService) getOrCreateBrand(ctx context.Context, tx *gorm.DB, id *uint, name *string) (uint, error) {
	if id != nil {
		return *id, nil
	}
	if name == nil {
		return 0, errors.New("brand_id or brand_name is required")
	}
	brand, err := productService.brandRepo.GetByNameTx(ctx, tx, *name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		brand = &models.Brand{Name: *name}
		brand, err = productService.brandRepo.CreateTx(ctx, tx, brand)
	}
	if err != nil {
		return 0, err
	}
	return brand.ID, nil
}
func (productService *productService) getOrCreateCategory(ctx context.Context, tx *gorm.DB, id *uint, name *string) (uint, error) {
	if id != nil {
		return *id, nil
	}
	if name == nil {
		return 0, errors.New("category_id or category_name is required")
	}
	category, err := productService.categoryRepo.GetByNameTx(ctx, tx, *name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		category = &models.Category{Name: *name}
		category, err = productService.categoryRepo.CreateTx(ctx, tx, category)
	}
	if err != nil {
		return 0, err
	}
	return category.ID, nil
}

func (productService *productService) getOrCreateMerchant(ctx context.Context, tx *gorm.DB, id *uint, name *string) (uint, error) {
	if id != nil {
		return *id, nil
	}
	if name == nil {
		return 0, errors.New("merchant_id or merchant_name is required")
	}

	// Tìm theo tên trong transaction
	merchant, err := productService.merchantRepo.GetByNameTx(ctx, tx, *name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		merchant = &models.Merchant{Name: *name}
		merchant, err = productService.merchantRepo.CreateTx(ctx, tx, merchant)
	}
	if err != nil {
		return 0, err
	}
	return merchant.ID, nil
}

func (productService *productService) GetProductList(ctx context.Context, priceMin, priceMax *float64,
	priceAsc *bool, totalBuyDesc *bool, page, pageSize int) ([]*CacheModel.ProductMiniCache, int, error) {
	var ListProductCache []*CacheModel.ProductMiniCache

	// TODO IMPL +
	// GetDB
	listProductFilter, total, err := productService.productRepo.GetProductListFilter(ctx, priceMin, priceMax, priceAsc, totalBuyDesc, page, pageSize)
	log.Printf("Product list: %v", listProductFilter)
	if err != nil {
		return nil, 0, customErr.NewError(customErr.UNEXPECTED_ERROR, "Failed to get product list", http.StatusInternalServerError, err)
	}
	err = productService.productCacheService.PingRedis(ctx)
	if err != nil {
		ListProductCache, err = productService.GetProducInfo(ctx, listProductFilter)
	} else {
		ListProductCache, err = productService.GetProductCacheInfo(ctx, listProductFilter)
	}
	return ListProductCache, total, nil
}

func (productService *productService) GetProductCacheInfo(ctx context.Context, list []CacheModel.ListProductQueryCache) ([]*CacheModel.ProductMiniCache, error) {
	// Get cache
	cacheList, missingList, err := productService.productCacheService.GetProductMiniCacheBulk(ctx, list)
	if err != nil {
		return nil, err
	}

	// Nếu còn missing → load từ DB
	if len(missingList) > 0 {
		missingProducts, err := productService.GetProducInfo(ctx, missingList)
		if err != nil {
			return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Failed to get product list", http.StatusInternalServerError, err)
		}
		err = productService.productCacheService.CacheMiniProducts(ctx, missingProducts)
		if err != nil {
			log.Printf("Failed to cache mini products: %v", err)
		}
		// Map lại missingProducts theo ProductID để gán nhanh vào cacheList
		missingMap := make(map[uint]*CacheModel.ProductMiniCache, len(missingProducts))
		for _, p := range missingProducts {
			missingMap[p.ID] = p
		}

		// Gán vào cacheList đúng vị trí ban đầu
		for i, item := range list {
			if cacheList[i] == nil {
				if mp, ok := missingMap[item.ProductID]; ok {
					cacheList[i] = mp
				}
			}
		}
	}

	return cacheList, nil

}

func (productService *productService) GetProductForMiniCache(ctx context.Context, productId uint, variantId uint) (*CacheModel.ProductMiniCache, error) {

	product, err := productService.productRepo.GetByIdPreloadVariant(ctx, productId)

	if err != nil {
		return nil, err
	}
	var varIds []uint
	for _, v := range product.Variants {
		varIds = append(varIds, v.ID)

	}

	variants, err := productService.productVariantService.CheckAndCacheProductVariants(ctx, varIds)
	if err != nil {
		return nil, err
	}
	var totalQuantity uint
	var price float64
	for _, variant := range variants {
		totalQuantity += variant.Quantity
		if variant.ID == variantId {
			price = variant.Price
		}
	}
	model := CacheModel.ProductMiniCache{
		ID:            product.ID,
		Name:          product.Name,
		AverageRating: product.AverageRating,
		NumberRating:  product.NumberRating,
		Image:         product.Image,
		TotalBuy:      product.TotalBuy,
		Variants:      varIds,
		TotalQuantity: totalQuantity,
		Price:         price,
	}

	log.Println(model, " Getproductforminicache")
	return &model, nil
}

//func (productService *productService) GetProducInfo(ctx context.Context, list []CacheModel.ListProductQueryCache) ([]CacheModel.ProductMiniCache, error) {
//	ids := make([]uint, 0, len(list))
//	for _, product := range list {
//		ids = append(ids, product.ProductID)
//	}
//	// Batch query preload variants
//	var products []models.Product
//	if err := productService.db.WithContext(ctx).
//		Preload("Variants").
//		Where("id IN ?", ids).
//		Find(&products).Error; err != nil {
//		return nil, err
//	}
//	variantIds := make([]uint, 0, len(products))
//	for _, product := range products {
//		for _, variant := range product.Variants {
//			variantIds = append(variantIds, variant.ID)
//		}
//	}
//
//	productVariants, err := productService.productVariantRepo.GetByIDSForRedisCache(ctx, ids)
//	if err != nil {
//		return nil, err
//	}
//	panic("implement me")
//}

func (productService *productService) GetProducInfo(
	ctx context.Context,
	list []CacheModel.ListProductQueryCache,
) ([]*CacheModel.ProductMiniCache, error) {

	// 1. Get IDs from list
	ids := make([]uint, 0, len(list))
	for _, product := range list {
		ids = append(ids, product.ProductID)
	}

	// 2. Batch query preload variants
	var products []models.Product
	if err := productService.db.WithContext(ctx).
		Preload("Variants").
		Where("id IN ?", ids).
		Find(&products).Error; err != nil {
		return nil, err
	}

	// 3. Get variants
	productVariants, err := productService.productVariantRepo.GetByIDSForRedisCache(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Map variantID → variant cache
	variantMap := make(map[uint]models.ProductVariant, len(productVariants))
	for _, v := range productVariants {
		variantMap[v.ID] = v
	}

	// Map productID → variantID từ list (để lấy price đúng variant)
	productVariantIDMap := make(map[uint]uint, len(list))
	for _, item := range list {
		productVariantIDMap[item.ProductID] = item.VariantID
	}

	// Map productID → product (để dễ lookup)
	productMap := make(map[uint]models.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	// 4. Mapping sang ProductMiniCache
	result := make([]*CacheModel.ProductMiniCache, 0, len(list))
	for _, item := range list {
		prod, ok := productMap[item.ProductID]
		if !ok {
			continue
		}

		// Tính totalQuantity + variants list
		var totalQuantity uint
		variantIDs := make([]uint, 0, len(prod.Variants))
		for _, v := range prod.Variants {
			variantIDs = append(variantIDs, v.ID)
			if pv, ok := variantMap[v.ID]; ok {
				totalQuantity += pv.Quantity
			}
		}

		// Lấy price từ variantID được chỉ định trong list
		var price float64
		if vID, ok := productVariantIDMap[item.ProductID]; ok {
			if pv, ok := variantMap[vID]; ok {
				price = pv.Price
			}
		}

		// Append vào kết quả
		result = append(result, &CacheModel.ProductMiniCache{
			ID:            prod.ID,
			Name:          prod.Name,
			AverageRating: prod.AverageRating,
			NumberRating:  prod.NumberRating,
			Image:         prod.Image,
			TotalBuy:      prod.TotalBuy,
			TotalQuantity: totalQuantity,
			Variants:      variantIDs,
			Price:         price,
		})
	}

	return result, nil
}
