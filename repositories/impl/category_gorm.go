package impl

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/repositories"
	"net/http"

	"github.com/minh6824pro/nxrGO/models"
	"gorm.io/gorm"
)

type categoryGormRepository struct {
	db *gorm.DB
}

func NewCategoryGormRepository(db *gorm.DB) repositories.CategoryRepository {
	return &categoryGormRepository{db}
}

func (r *categoryGormRepository) Create(ctx context.Context, c *models.Category) (*models.Category, error) {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, customErr.NewError(customErr.DUPLICATED_ERROR, "Category already exists", http.StatusBadRequest, nil)
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)

	}
	return c, nil
}

func (r *categoryGormRepository) CreateTx(ctx context.Context, tx *gorm.DB, c *models.Category) (*models.Category, error) {
	if err := tx.WithContext(ctx).Create(c).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, customErr.NewError(customErr.DUPLICATED_ERROR, "Category already exists", http.StatusBadRequest, nil)
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)

	}
	return c, nil
}

func (r *categoryGormRepository) GetByID(ctx context.Context, id uint) (*models.Category, error) {
	var c models.Category
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.NewError(customErr.ITEM_NOT_FOUND, "Category not found", http.StatusBadRequest, nil)
		}
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)

	}
	return &c, nil
}

func (r *categoryGormRepository) Update(ctx context.Context, c *models.Category) error {
	if err := r.db.WithContext(ctx).Save(c).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return customErr.NewError(customErr.DUPLICATED_ERROR, "Category already exists", http.StatusBadRequest, nil)
		}
		return customErr.NewError(customErr.INTERNAL_ERROR, "Unknown error", http.StatusInternalServerError, nil)
	}
	return nil
}

func (r *categoryGormRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.Category{}, id).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1451 {
				return customErr.NewError(customErr.FOREIGN_KEY_CONSTRAINT_ERROR, "Cannot delete category because it is associated with existing products", http.StatusBadRequest, nil)
			}
		}
		return customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return nil
}

func (r *categoryGormRepository) List(ctx context.Context) ([]models.Category, error) {
	var list []models.Category
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, customErr.NewError(customErr.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}
	return list, nil
}

func (r *categoryGormRepository) GetByName(ctx context.Context, name string) (*models.Category, error) {
	var c models.Category
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryGormRepository) GetByNameTx(ctx context.Context, tx *gorm.DB, name string) (*models.Category, error) {
	var c models.Category
	if err := tx.WithContext(ctx).Where("name = ?", name).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}
