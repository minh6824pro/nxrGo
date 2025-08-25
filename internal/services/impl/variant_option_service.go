package impl

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/repositories"
	"github.com/minh6824pro/nxrGO/internal/services"
)

type variantOptionService struct {
	repo repositories.VariantOptionRepository
}

func NewVariantOptionService(repo repositories.VariantOptionRepository) services.VariantOptionService {
	return &variantOptionService{repo}
}

func (r *variantOptionService) Create(ctx context.Context, variantOption *dto.CreateVariantOptionInput) (*models.VariantOption, error) {
	return r.repo.Create(ctx, CreateVariantOptionDtoMapper(variantOption))
}

func (r *variantOptionService) GetByID(ctx context.Context, id uint) (*models.VariantOption, error) {
	return r.repo.GetByID(ctx, id)
}

func (r *variantOptionService) List(ctx context.Context) ([]models.VariantOption, error) {
	return r.repo.List(ctx)
}

func (r *variantOptionService) Delete(ctx context.Context, id uint) error {
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return r.repo.Delete(ctx, id)
}

func (r *variantOptionService) Patch(ctx context.Context, id uint, variantOption *dto.UpdateVariantOptionInput) (*models.VariantOption, error) {
	existing, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if variantOption.Name != "" {
		existing.Name = variantOption.Name
	}
	if err := r.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func CreateVariantOptionDtoMapper(m *dto.CreateVariantOptionInput) *models.VariantOption {
	return &models.VariantOption{
		Name: m.Name,
	}
}
