package impl

import (
	"context"
	"github.com/minh6824pro/nxrGO/dto"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
)

type merchantService struct {
	repo repositories.MerchantRepository
}

func NewMerchantService(repo repositories.MerchantRepository) services.MerchantService {
	return &merchantService{repo}
}

func (merchantService *merchantService) Create(ctx context.Context, m *dto.CreateMerchantInput) (*models.Merchant, error) {
	return merchantService.repo.Create(ctx, MerchantInputDtoMapper(m))
}

func (merchantService *merchantService) GetByID(ctx context.Context, id uint) (*models.Merchant, error) {

	return merchantService.repo.GetByID(ctx, id)
}

//func (merchantService *merchantService) Update(ctx context.Context, m *models.Merchant) error {
//	existing, err := merchantService.repo.GetByID(ctx, m.ID)
//
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			return gorm.ErrRecordNotFound
//		}
//		return err
//	}
//
//	existing.Name = m.Name
//
//	return merchantService.repo.Update(ctx, existing)
//}

func (merchantService *merchantService) Delete(ctx context.Context, id uint) error {
	_, err := merchantService.repo.GetByID(ctx, id)

	if err != nil {
		return err
	}

	return merchantService.repo.Delete(ctx, id)
}

func (merchantService *merchantService) List(ctx context.Context) ([]models.Merchant, error) {
	return merchantService.repo.List(ctx)
}

func (merchantService *merchantService) Patch(ctx context.Context, id uint, input *dto.UpdateMerchantInput) (*models.Merchant, error) {
	existing, err := merchantService.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		existing.Name = input.Name
	}
	if err := merchantService.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func MerchantInputDtoMapper(m *dto.CreateMerchantInput) *models.Merchant {
	return &models.Merchant{
		Name: m.Name,
	}
}
