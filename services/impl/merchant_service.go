package impl

import (
	"context"
	"errors"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
	"gorm.io/gorm"
)

type merchantService struct {
	repo repositories.MerchantRepository
}

func NewMerchantService(repo repositories.MerchantRepository) services.MerchantService {
	return &merchantService{repo}
}

func (merchantService *merchantService) Create(ctx context.Context, m *models.Merchant) (*models.Merchant, error) {
	return merchantService.repo.Create(ctx, m)
}

func (merchantService *merchantService) GetByID(ctx context.Context, id uint) (*models.Merchant, error) {

	return merchantService.repo.GetByID(ctx, id)
}

func (merchantService *merchantService) Update(ctx context.Context, m *models.Merchant) error {
	existing, err := merchantService.repo.GetByID(ctx, m.ID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		return err
	}

	existing.Name = m.Name

	return merchantService.repo.Update(ctx, existing)
}

func (merchantService *merchantService) Delete(ctx context.Context, id uint) error {
	_, err := merchantService.repo.GetByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		return err
	}

	return merchantService.repo.Delete(ctx, id)
}

func (merchantService *merchantService) List(ctx context.Context) ([]models.Merchant, error) {
	return merchantService.repo.List(ctx)
}

func (merchantService *merchantService) Patch(ctx context.Context, id uint, updates map[string]interface{}) (*models.Merchant, error) {
	existing, err := merchantService.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Optional updates
	if name, ok := updates["name"].(string); ok {
		existing.Name = name
	}
	if err := merchantService.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}
