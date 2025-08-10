package impl

import (
	"context"
	"errors"
	"fmt"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"gorm.io/gorm"
	"net/http"
)

type productVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantGormRepository(db *gorm.DB) repositories.ProductVariantRepository {
	return &productVariantRepository{db: db}
}

func (r *productVariantRepository) Create(ctx context.Context, variant *models.ProductVariant) (*models.ProductVariant, error) {
	if err := r.db.WithContext(ctx).Create(variant).Error; err != nil {
		return nil, err
	}
	return variant, nil
}

func (r *productVariantRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, variant *models.ProductVariant) (*models.ProductVariant, error) {
	if err := tx.WithContext(ctx).Create(variant).Error; err != nil {
		return nil, err
	}
	return variant, nil
}
func (r *productVariantRepository) GetByID(ctx context.Context, id uint) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	err := r.db.WithContext(ctx).
		Preload("Options").
		First(&variant, id).Error
	if err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *productVariantRepository) GetByIDNoPreload(ctx context.Context, id uint) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	err := r.db.WithContext(ctx).
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

func (r *productVariantRepository) List(ctx context.Context) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	err := r.db.WithContext(ctx).
		Preload("Options").
		Find(&variants).Error
	return variants, err
}

func (r *productVariantRepository) CheckExistsAndQuantity(ctx context.Context, id uint, quantity uint) error {
	var variant models.ProductVariant
	err := r.db.WithContext(ctx).First(&variant, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customErr.NewError(customErr.ITEM_NOT_FOUND, fmt.Sprintf("Product variant not found: %d", id), http.StatusBadRequest, nil)

		}
		return err
	}
	if quantity > variant.Quantity {
		return customErr.NewError(customErr.INSUFFICIENT_STOCK, fmt.Sprintf("Product : %d Insufficient stock", id), http.StatusBadRequest, nil)

	}
	return nil
}

func (r *productVariantRepository) GetByIDs(ctx context.Context, ids []uint) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&variants).Error; err != nil {
		return nil, err
	}
	return variants, nil
}
