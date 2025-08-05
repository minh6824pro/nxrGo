package impl

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"github.com/minh6824pro/nxrGO/services"
)

type categoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(r repositories.CategoryRepository) services.CategoryService {
	return &categoryService{r}
}

func (categoryService *categoryService) Create(ctx context.Context, c *models.Category) (*models.Category, error) {
	return categoryService.repo.Create(ctx, c)
}

func (categoryService *categoryService) GetByID(ctx context.Context, id uint) (*models.Category, error) {
	return categoryService.repo.GetByID(ctx, id)
}

func (categoryService *categoryService) Update(ctx context.Context, c *models.Category) error {
	return categoryService.repo.Update(ctx, c)
}

func (categoryService *categoryService) Delete(ctx context.Context, id uint) error {
	return categoryService.repo.Delete(ctx, id)
}

func (categoryService *categoryService) List(ctx context.Context) ([]models.Category, error) {
	return categoryService.repo.List(ctx)
}

func (categoryService *categoryService) Patch(ctx context.Context, id uint, updates map[string]interface{}) (*models.Category, error) {
	existing, err := categoryService.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Optional updates
	if name, ok := updates["name"].(string); ok {
		existing.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		existing.Description = desc
	}

	if err := categoryService.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}
