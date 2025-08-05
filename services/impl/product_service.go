package impl

import (
	"context"
	"errors"
	"github.com/minh6824pro/nxrGO/dto"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
	"gorm.io/gorm"
)

func NewProductService(productRepo repositories.ProductRepository, brandRepo repositories.BrandRepository, merchanRepo repositories.MerchantRepository, categoryRepo repositories.CategoryRepository) services.ProductService {
	return &productService{
		productRepo:  productRepo,
		brandRepo:    brandRepo,
		merchantRepo: merchanRepo,
		categoryRepo: categoryRepo,
	}
}

type productService struct {
	productRepo  repositories.ProductRepository
	brandRepo    repositories.BrandRepository
	merchantRepo repositories.MerchantRepository
	categoryRepo repositories.CategoryRepository
}

func (productService *productService) Create(ctx context.Context, input dto.CreateProductInput) (*models.Product, error) {
	// Xử lý Brand
	brandID, err := productService.getOrCreateBrand(ctx, input.BrandID, input.BrandName)
	if err != nil {
		return nil, err
	}

	// Xử lý Category
	categoryID, err := productService.getOrCreateCategory(ctx, input.CategoryID, input.CategoryName)
	if err != nil {
		return nil, err
	}

	// Xử lý Merchant
	merchantID, err := productService.getOrCreateMerchant(ctx, input.MerchantID, input.MerchantName)
	if err != nil {
		return nil, err
	}

	// Tạo Product
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

	// Tạo trong DB
	createdProduct, err := productService.productRepo.Create(ctx, product)
	if err != nil {
		return nil, err
	}

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

func (s *productService) getOrCreateBrand(ctx context.Context, id *uint, name *string) (uint, error) {
	if id != nil {
		return *id, nil
	}
	if name == nil {
		return 0, errors.New("brand_id or brand_name is required")
	}
	brand, err := s.brandRepo.GetByName(ctx, *name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		brand = &models.Brand{Name: *name}
		brand, err = s.brandRepo.Create(ctx, brand)
	}
	if err != nil {
		return 0, err
	}
	return brand.ID, nil
}

func (s *productService) getOrCreateCategory(ctx context.Context, id *uint, name *string) (uint, error) {
	if id != nil {
		return *id, nil
	}
	if name == nil {
		return 0, errors.New("category_id or category_name is required")
	}
	category, err := s.categoryRepo.GetByName(ctx, *name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		category = &models.Category{Name: *name}
		category, err = s.categoryRepo.Create(ctx, category)
	}
	if err != nil {
		return 0, err
	}
	return category.ID, nil
}

func (s *productService) getOrCreateMerchant(ctx context.Context, id *uint, name *string) (uint, error) {
	if id != nil {
		return *id, nil
	}
	if name == nil {
		return 0, errors.New("merchant_id or merchant_name is required")
	}
	merchant, err := s.merchantRepo.GetByName(ctx, *name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		merchant = &models.Merchant{Name: *name}
		merchant, err = s.merchantRepo.Create(ctx, merchant)
	}
	if err != nil {
		return 0, err
	}
	return merchant.ID, nil
}
