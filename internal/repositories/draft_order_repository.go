package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"

	"gorm.io/gorm"
)

type DraftOrderRepository interface {
	CreateTx(ctx context.Context, tx *gorm.DB, order *models.DraftOrder) (*models.DraftOrder, error)
	Create(ctx context.Context, order *models.DraftOrder) error
	Delete(ctx context.Context, id uint) error
	GetById(ctx context.Context, orderID uint) (*models.DraftOrder, error)
	Save(ctx context.Context, order *models.DraftOrder) error
	GetsForDbUpdate(ctx context.Context) ([]models.DraftOrder, error)
	CleanDraft(ctx context.Context) error
	ListByUserIdToOrderNull(ctx context.Context, draftOrderID uint) ([]*models.DraftOrder, error)
	ListByAdmin(ctx context.Context) ([]*models.DraftOrder, error)
	GetByParentId(ctx context.Context, parentId uint) ([]models.DraftOrder, error)
	GetForSplit(ctx context.Context, id uint) ([]dto.OrderItemForSplit, error)
}
