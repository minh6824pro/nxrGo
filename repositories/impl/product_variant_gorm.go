package impl

import (
	"context"
	"errors"
	"fmt"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	"github.com/minh6824pro/nxrGO/repositories"
	"gorm.io/gorm"
	"net/http"
)

type productVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantGormRepository(db *gorm.DB) repositories.ProductVariantRepository {
	return &productVariantRepository{db: db}
}

func (r *productVariantRepository) Create(ctx context.Context, variant *models.ProductVariant) (*models.ProductVariant, error) {
	if err := r.db.WithContext(ctx).Create(variant).Error; err != nil {
		return nil, err
	}
	return variant, nil
}

func (r *productVariantRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, variant *models.ProductVariant) (*models.ProductVariant, error) {
	if err := tx.WithContext(ctx).Create(variant).Error; err != nil {
		return nil, err
	}
	return variant, nil
}
func (r *productVariantRepository) GetByID(ctx context.Context, id uint) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	err := r.db.WithContext(ctx).
		Preload("Options").
		First(&variant, id).Error
	if err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *productVariantRepository) GetByIDNoPreload(ctx context.Context, id uint) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	err := r.db.WithContext(ctx).
		First(&variant, id).Error
	if err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *productVariantRepository) Update(ctx context.Context, variant *models.ProductVariant) error {
	return r.db.WithContext(ctx).Save(variant).Error
}

func (r *productVariantRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ProductVariant{}, id).Error
}

func (r *productVariantRepository) List(ctx context.Context) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	err := r.db.WithContext(ctx).
		Preload("Options").
		Find(&variants).Error
	return variants, err
}

func (r *productVariantRepository) CheckExistsAndQuantity(ctx context.Context, id uint, quantity uint) error {
	var variant models.ProductVariant
	err := r.db.WithContext(ctx).First(&variant, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customErr.NewError(customErr.ITEM_NOT_FOUND, fmt.Sprintf("Product variant not found: %d", id), http.StatusBadRequest, nil)

		}
		return err
	}
	if quantity > variant.Quantity {
		return customErr.NewError(customErr.INSUFFICIENT_STOCK, fmt.Sprintf("Product : %d Insufficient stock", id), http.StatusBadRequest, nil)

	}
	return nil
}

func (r *productVariantRepository) GetByIDSForRedisCache(ctx context.Context, ids []uint) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&variants).Error; err != nil {
		return nil, err
	}

	// Struct for quantity reserved in draft_orders
	type ReservedQty struct {
		ProductVariantID uint
		TotalQuantity    uint
	}

	var reserved []ReservedQty
	if err := r.db.WithContext(ctx).
		Table("order_items oi").
		Select("oi.product_variant_id, COALESCE(SUM(oi.quantity), 0) as total_quantity").
		Joins("INNER JOIN draft_orders do ON do.id = oi.order_id").
		Where("oi.product_variant_id IN ? AND oi.order_type = ? AND do.to_order is null or do.to_order !=0 ", ids, "draft_order").
		Group("oi.product_variant_id").
		Scan(&reserved).Error; err != nil {
		return nil, err
	}

	reservedMap := make(map[uint]uint)
	for _, r := range reserved {
		reservedMap[r.ProductVariantID] = r.TotalQuantity
	}

	// Update real quantity
	for i, v := range variants {
		if qty, ok := reservedMap[v.ID]; ok {
			variants[i].Quantity = v.Quantity - qty
		}
	}

	return variants, nil
}

func (r *productVariantRepository) UpdateQuantity(ctx context.Context, quantityMap map[uint]uint) error {
	// Begin Transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for variantID, qty := range quantityMap {
		// Update quantity for each product variant
		if err := tx.Model(&models.ProductVariant{}).
			Where("id = ?", variantID).
			UpdateColumn("quantity", gorm.Expr("quantity - ?", qty)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
