package impl

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"gorm.io/gorm"
	"net/http"
)

type brandGormRepository struct {
	db *gorm.DB
}

func NewBrandGormRepository(db *gorm.DB) repositories.BrandRepository {
	return &brandGormRepository{db}
}

func (r *brandGormRepository) Create(ctx context.Context, b *models.Brand) (*models.Brand, error) {
	if err := r.db.WithContext(ctx).Create(b).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return nil, customErr.NewError(customErr.DUPLICATED_ERROR, "Brand already exists", http.StatusBadRequest, nil)
			}
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return b, nil
}

func (r *brandGormRepository) GetByID(ctx context.Context, id uint) (*models.Brand, error) {
	var b models.Brand
	if err := r.db.WithContext(ctx).First(&b, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.NewError(customErr.ITEM_NOT_FOUND, "Brand not found", http.StatusBadRequest, nil)
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return &b, nil
}

func (r *brandGormRepository) Update(ctx context.Context, b *models.Brand) error {
	if err := r.db.WithContext(ctx).Save(b).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return customErr.NewError(customErr.DUPLICATED_ERROR, "Brand already exists", http.StatusBadRequest, nil)
			}
		}
		return customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return nil
}

func (r *brandGormRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.Merchant{}, id).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1451 {
				return customErr.NewError(customErr.FOREIGN_KEY_CONSTRAINT_ERROR, "Cannot delete brand because it is associated with existing products", http.StatusBadRequest, nil)
			}
		}
		return customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return nil
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
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return &b, nil
}
