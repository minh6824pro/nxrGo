package impl

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"gorm.io/gorm"
)

type brandGormRepository struct {
	db *gorm.DB
}

func NewBrandGormRepository(db *gorm.DB) repositories.BrandRepository {
	return &brandGormRepository{db}
}

func (r *brandGormRepository) Create(ctx context.Context, b *models.Brand) (*models.Brand, error) {
	if err := r.db.WithContext(ctx).Create(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

func (r *brandGormRepository) GetByID(ctx context.Context, id uint) (*models.Brand, error) {
	var b models.Brand
	if err := r.db.WithContext(ctx).First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *brandGormRepository) Update(ctx context.Context, b *models.Brand) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *brandGormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Brand{}, id).Error
}

func (r *brandGormRepository) List(ctx context.Context) ([]models.Brand, error) {
	var list []models.Brand
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *brandGormRepository) GetByName(ctx context.Context, name string) (*models.Brand, error) {
	var b models.Brand
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}
