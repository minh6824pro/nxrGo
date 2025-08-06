package impl

import (
	"context"
	"errors"
	"fmt"
	"github.com/minh6824pro/nxrGO/dto"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
	"gorm.io/gorm"
	"net/http"
)

func NewProductService(db *gorm.DB, productRepo repositories.ProductRepository, brandRepo repositories.BrandRepository, merchanRepo repositories.MerchantRepository, categoryRepo repositories.CategoryRepository, productVariantRepo repositories.ProductVariantRepository, variantOptionValueRepo repositories.VariantOptionValueRepository, variantOptionRepo repositories.VariantOptionRepository) services.ProductService {
	return &productService{
		db:                     db,
		productRepo:            productRepo,
		brandRepo:              brandRepo,
		merchantRepo:           merchanRepo,
		categoryRepo:           categoryRepo,
		productVariantRepo:     productVariantRepo,
		variantOptionValueRepo: variantOptionValueRepo,
		variantOptionRepo:      variantOptionRepo,
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
func (productService *productService) Create(ctx context.Context, input dto.CreateProductInput) (*models.Product, error) {
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

	// ✅ Truyền tx vào các hàm getOrCreate*
	brandID, err := productService.getOrCreateBrand(ctx, tx, input.BrandID, input.BrandName)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	categoryID, err := productService.getOrCreateCategory(ctx, tx, input.CategoryID, input.CategoryName)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	merchantID, err := productService.getOrCreateMerchant(ctx, tx, input.MerchantID, input.MerchantName)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	product := &models.Product{
		Name:          input.Name,
		Description:   input.Description,
		Image:         input.Image,
		BrandID:       brandID,
		CategoryID:    categoryID,
		MerchantID:    merchantID,
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

	return createdProduct, nil
}

func (productService productService) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	return productService.productRepo.GetByID(ctx, id)
}

func (productService productService) List(ctx context.Context) ([]models.Product, error) {
	return productService.productRepo.List(ctx)

}

func (productService productService) Delete(ctx context.Context, id uint) error {
	_, err := productService.productRepo.GetByID(ctx, id)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gorm.ErrRecordNotFound
		}
		return err
	}

	return productService.productRepo.Delete(ctx, id)

}

func (productService productService) Patch(ctx context.Context, product *models.Product) error {
	//TODO implement me
	panic("implement me")
}

func (s *productService) getOrCreateBrand(ctx context.Context, tx *gorm.DB, id *uint, name *string) (uint, error) {
	if id != nil {
		return *id, nil
	}
	if name == nil {
		return 0, errors.New("brand_id or brand_name is required")
	}
	brand, err := s.brandRepo.GetByNameTx(ctx, tx, *name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		brand = &models.Brand{Name: *name}
		brand, err = s.brandRepo.CreateTx(ctx, tx, brand)
	}
	if err != nil {
		return 0, err
	}
	return brand.ID, nil
}
func (s *productService) getOrCreateCategory(ctx context.Context, tx *gorm.DB, id *uint, name *string) (uint, error) {
	if id != nil {
		return *id, nil
	}
	if name == nil {
		return 0, errors.New("category_id or category_name is required")
	}
	category, err := s.categoryRepo.GetByNameTx(ctx, tx, *name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		category = &models.Category{Name: *name}
		category, err = s.categoryRepo.CreateTx(ctx, tx, category)
	}
	if err != nil {
		return 0, err
	}
	return category.ID, nil
}

func (s *productService) getOrCreateMerchant(ctx context.Context, tx *gorm.DB, id *uint, name *string) (uint, error) {
	if id != nil {
		return *id, nil
	}
	if name == nil {
		return 0, errors.New("merchant_id or merchant_name is required")
	}

	// Tìm theo tên trong transaction
	merchant, err := s.merchantRepo.GetByNameTx(ctx, tx, *name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		merchant = &models.Merchant{Name: *name}
		merchant, err = s.merchantRepo.CreateTx(ctx, tx, merchant)
	}
	if err != nil {
		return 0, err
	}
	return merchant.ID, nil
}
