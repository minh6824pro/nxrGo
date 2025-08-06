package impl

import (
	"context"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"net/http"

	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"gorm.io/gorm"
)

type productGormRepository struct {
	db *gorm.DB
}

func NewProductGormRepository(db *gorm.DB) repositories.ProductRepository {
	return &productGormRepository{db}
}

func (r *productGormRepository) Create(ctx context.Context, p *models.Product) (*models.Product, error) {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return nil, err
	}
	// Preload ngay sau khi tạo
	if err := r.db.WithContext(ctx).
		Preload("Brand").
		Preload("Category").
		Preload("Merchant").
		First(p, p.ID).Error; err != nil {
		return nil, err
	}

	return p, nil
}

func (r *productGormRepository) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	var p models.Product
	err := r.db.WithContext(ctx).
		Preload("Brand").
		Preload("Merchant").
		Preload("Category").
		First(&p, id).Error

	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productGormRepository) Update(ctx context.Context, p *models.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *productGormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Product{}, id).Error
}

func (r *productGormRepository) List(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("Brand").
		Preload("Category").
		Find(&products).Error; err != nil {
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}

	return products, nil
}
