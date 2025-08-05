package impl

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"gorm.io/gorm"
)

type merchantGormRepository struct {
	db *gorm.DB
}

func NewMerchantGormRepository(db *gorm.DB) repositories.MerchantRepository {
	return &merchantGormRepository{db}
}

func (r *merchantGormRepository) Create(ctx context.Context, m *models.Merchant) (*models.Merchant, error) {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (r *merchantGormRepository) GetByID(ctx context.Context, id uint) (*models.Merchant, error) {
	var m models.Merchant
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *merchantGormRepository) Update(ctx context.Context, m *models.Merchant) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *merchantGormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Merchant{}, id).Error
}

func (r *merchantGormRepository) List(ctx context.Context) ([]models.Merchant, error) {
	var list []models.Merchant
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *merchantGormRepository) GetByName(ctx context.Context, name string) (*models.Merchant, error) {
	var m models.Merchant
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
