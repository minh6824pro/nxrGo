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

type variantOptionValueGormRepository struct {
	db *gorm.DB
}

func NewVariantOptionValueGormRepository(db *gorm.DB) repositories.VariantOptionValueRepository {
	return &variantOptionValueGormRepository{db: db}
}

func (v *variantOptionValueGormRepository) Create(ctx context.Context, variant *models.VariantOptionValue) (*models.VariantOptionValue, error) {
	if err := v.db.WithContext(ctx).Create(variant).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, customErr.NewError(customErr.DUPLICATED_ERROR, "Variant option value already exists", http.StatusBadRequest, nil)
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return variant, nil
}

func (v *variantOptionValueGormRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, variant *models.VariantOptionValue) (*models.VariantOptionValue, error) {
	if err := tx.WithContext(ctx).Create(variant).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, customErr.NewError(customErr.DUPLICATED_ERROR, "Variant option value already exists", http.StatusBadRequest, nil)
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return variant, nil
}
