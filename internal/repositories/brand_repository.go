package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/models"
	"gorm.io/gorm"
)

type BrandRepository interface {
	Create(ctx context.Context, brand *models.Brand) (*models.Brand, error)
	GetByID(ctx context.Context, id uint) (*models.Brand, error)
	GetByIDTx(ctx context.Context, tx *gorm.DB, id uint) (*models.Brand, error)
	Update(ctx context.Context, brand *models.Brand) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Brand, error)
	GetByName(ctx context.Context, name string) (*models.Brand, error)
	CreateTx(ctx context.Context, tx *gorm.DB, brand *models.Brand) (*models.Brand, error)
	GetByNameTx(ctx context.Context, tx *gorm.DB, name string) (*models.Brand, error)
}
