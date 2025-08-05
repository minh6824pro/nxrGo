package impl

import (
	"context"
	"github.com/minh6824pro/nxrGO/repositories"

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
		return nil, err
	}
	return c, nil
}

func (r *categoryGormRepository) GetByID(ctx context.Context, id uint) (*models.Category, error) {
	var c models.Category
	err := r.db.WithContext(ctx).First(&c, id).Error
	return &c, err
}

func (r *categoryGormRepository) Update(ctx context.Context, c *models.Category) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *categoryGormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Category{}, id).Error
}

func (r *categoryGormRepository) List(ctx context.Context) ([]models.Category, error) {
	var list []models.Category
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *categoryGormRepository) GetByName(ctx context.Context, name string) (*models.Category, error) {
	var c models.Category
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}
