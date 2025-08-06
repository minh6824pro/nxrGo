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

type merchantGormRepository struct {
	db *gorm.DB
}

func NewMerchantGormRepository(db *gorm.DB) repositories.MerchantRepository {
	return &merchantGormRepository{db}
}

func (r *merchantGormRepository) Create(ctx context.Context, m *models.Merchant) (*models.Merchant, error) {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return nil, customErr.NewError(customErr.DUPLICATED_ERROR, "Merchant already exists", http.StatusBadRequest, nil)
			}
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return m, nil
}

func (r *merchantGormRepository) GetByID(ctx context.Context, id uint) (*models.Merchant, error) {
	var m models.Merchant
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.NewError(customErr.ITEM_NOT_FOUND, "Merchant not found", http.StatusBadRequest, nil)
		}

		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return &m, nil
}

func (r *merchantGormRepository) Update(ctx context.Context, m *models.Merchant) error {

	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return customErr.NewError(customErr.DUPLICATED_ERROR, "Merchant already exists", http.StatusBadRequest, nil)
			}
		}
		return customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return nil
}

func (r *merchantGormRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.Merchant{}, id).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1451 {
				return customErr.NewError(customErr.FOREIGN_KEY_CONSTRAINT_ERROR, "Cannot delete merchant because it is associated with existing products", http.StatusBadRequest, nil)
			}
		}
		return customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return nil
}

func (r *merchantGormRepository) List(ctx context.Context) ([]models.Merchant, error) {
	var list []models.Merchant
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
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
