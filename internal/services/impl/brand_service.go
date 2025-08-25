package impl

import (
	"context"
	"errors"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/repositories"
	"github.com/minh6824pro/nxrGO/internal/services"
	"gorm.io/gorm"
)

type brandService struct {
	repo repositories.BrandRepository
}

func NewBrandService(repo repositories.BrandRepository) services.BrandService {
	return &brandService{repo}
}

func (brandService *brandService) Create(ctx context.Context, b *dto.CreateBrandInput) (*models.Brand, error) {
	return brandService.repo.Create(ctx, CreateBrandInputDtoMapper(b))
}

func (brandService *brandService) GetByID(ctx context.Context, id uint) (*models.Brand, error) {
	return brandService.repo.GetByID(ctx, id)
}

func (brandService *brandService) Update(ctx context.Context, b *models.Brand) error {
	existing, err := brandService.repo.GetByID(ctx, b.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		return err
	}

	existing.Name = b.Name
	return brandService.repo.Update(ctx, existing)
}

func (brandService *brandService) Delete(ctx context.Context, id uint) error {
	_, err := brandService.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return brandService.repo.Delete(ctx, id)
}

func (brandService *brandService) List(ctx context.Context) ([]models.Brand, error) {
	return brandService.repo.List(ctx)
}

func (brandService *brandService) Patch(ctx context.Context, id uint, input *dto.UpdateMerchantInput) (*models.Brand, error) {
	existing, err := brandService.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		existing.Name = input.Name
	}
	if err := brandService.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func CreateBrandInputDtoMapper(m *dto.CreateBrandInput) *models.Brand {
	return &models.Brand{
		Name: m.Name,
	}
}

func UpdateBrandInputDtoMapper(m *dto.UpdateBrandInput) *models.Brand {
	return &models.Brand{
		Name: m.Name,
	}
}
