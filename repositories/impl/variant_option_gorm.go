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

type variantOptionGormRepository struct {
	db *gorm.DB
}

func NewVariantOptionGormRepository(db *gorm.DB) repositories.VariantOptionRepository {
	return &variantOptionGormRepository{db}
}

func (r *variantOptionGormRepository) Create(ctx context.Context, option *models.VariantOption) (*models.VariantOption, error) {
	if err := r.db.WithContext(ctx).Create(option).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, customErr.NewError(customErr.DUPLICATED_ERROR, "Variant option already exists", http.StatusBadRequest, nil)
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return option, nil
}

func (r *variantOptionGormRepository) GetByID(ctx context.Context, id uint) (*models.VariantOption, error) {
	var option models.VariantOption
	if err := r.db.WithContext(ctx).First(&option, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.NewError(customErr.ITEM_NOT_FOUND, "Variant option not found", http.StatusBadRequest, nil)
		}

		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return &option, nil
}

func (r *variantOptionGormRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.VariantOption{}, id).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1451 {
			return customErr.NewError(customErr.FOREIGN_KEY_CONSTRAINT_ERROR, "Cannot delete merchant because it is associated with existing products", http.StatusBadRequest, nil)
		}
		return customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return nil
}

func (r *variantOptionGormRepository) List(ctx context.Context) ([]models.VariantOption, error) {
	var list []models.VariantOption
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return list, nil
}

func (r *variantOptionGormRepository) Update(ctx context.Context, variantOption *models.VariantOption) error {
	return r.db.WithContext(ctx).Save(variantOption).Error
}
