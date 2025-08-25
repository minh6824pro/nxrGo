package impl

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/repositories"
	"github.com/minh6824pro/nxrGO/internal/services"
)

type categoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(r repositories.CategoryRepository) services.CategoryService {
	return &categoryService{r}
}

func (categoryService *categoryService) Create(ctx context.Context, c *dto.CreateCategoryInput) (*models.Category, error) {
	return categoryService.repo.Create(ctx, CreateCategoryInputDtoMapper(c))
}

func (categoryService *categoryService) GetByID(ctx context.Context, id uint) (*models.Category, error) {
	return categoryService.repo.GetByID(ctx, id)
}

func (categoryService *categoryService) Update(ctx context.Context, c *models.Category) error {
	return categoryService.repo.Update(ctx, c)
}

func (categoryService *categoryService) Delete(ctx context.Context, id uint) error {
	_, err := categoryService.repo.GetByID(ctx, id)

	if err != nil {
		return err
	}

	return categoryService.repo.Delete(ctx, id)
}

func (categoryService *categoryService) List(ctx context.Context) ([]models.Category, error) {
	return categoryService.repo.List(ctx)
}

func (categoryService *categoryService) Patch(ctx context.Context, id uint, input *dto.UpdateCategoryInput) (*models.Category, error) {
	existing, err := categoryService.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.Description != nil {
		existing.Description = *input.Description
	}

	if err := categoryService.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func CreateCategoryInputDtoMapper(m *dto.CreateCategoryInput) *models.Category {
	var name, description string

	if m.Name != nil {
		name = *m.Name
	}
	if m.Description != nil {
		description = *m.Description
	}

	return &models.Category{
		Name:        name,
		Description: description,
	}
}
