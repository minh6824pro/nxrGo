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

func NewProductVariantService(productRepo repositories.ProductRepository, productVariantRepo repositories.ProductVariantRepository) services.ProductVariantService {
	return &productVariantService{
		productRepo:        productRepo,
		productVariantRepo: productVariantRepo,
	}
}

type productVariantService struct {
	productRepo        repositories.ProductRepository
	productVariantRepo repositories.ProductVariantRepository
}

func (p *productVariantService) Create(ctx context.Context, input dto.CreateProductVariantInput) (*models.ProductVariant, error) {

	product, err := p.productRepo.GetByID(ctx, input.ProductID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	image := ""
	if input.Image != "" {
		image = input.Image
	}
	productVariant := &models.ProductVariant{
		ProductID: product.ID,
		Quantity:  input.Quantity,
		Price:     input.Price,
		Image:     image,
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
