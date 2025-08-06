package services

import (
	"context"
	"github.com/minh6824pro/nxrGO/dto"
	"github.com/minh6824pro/nxrGO/models"
)

type CategoryService interface {
	Create(ctx context.Context, c *dto.CreateCategoryInput) (*models.Category, error)
	GetByID(ctx context.Context, id uint) (*models.Category, error)
	Update(ctx context.Context, c *models.Category) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Category, error)
	Patch(ctx context.Context, id uint, input *dto.UpdateCategoryInput) (*models.Category, error)
}
