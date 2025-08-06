package impl

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"gorm.io/gorm"
)

type productVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantRepository(db *gorm.DB) repositories.ProductVariantRepository {
	return &productVariantRepository{db: db}
}

func (r *productVariantRepository) Create(ctx context.Context, variant *models.ProductVariant) (*models.ProductVariant, error) {
	if err := r.db.WithContext(ctx).Create(variant).Error; err != nil {
		return nil, err
	}
	return variant, nil
}

func (r *productVariantRepository) FindByID(ctx context.Context, id uint) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	err := r.db.WithContext(ctx).
		Preload("Options").
		First(&variant, id).Error
	if err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *productVariantRepository) Update(ctx context.Context, variant *models.ProductVariant) error {
	return r.db.WithContext(ctx).Save(variant).Error
}

func (r *productVariantRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ProductVariant{}, id).Error
}

func (r *productVariantRepository) FindAll(ctx context.Context) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	err := r.db.WithContext(ctx).
		Preload("Options").
		Find(&variants).Error
	return variants, err
}
