package impl

import (
	"context"
	"errors"
	"fmt"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/repositories"
	customErr "github.com/minh6824pro/nxrGO/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"net/http"
	"strings"
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
		Preload("OptionValues").
		First(&variant, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.NewError(customErr.ITEM_NOT_FOUND, fmt.Sprintf("Product Variant id %d not found ", id), http.StatusBadRequest, nil)
		}
		return nil, customErr.NewError(customErr.INTERNAL_ERROR, "Unable to fetch product variant", http.StatusInternalServerError, err)
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

func (r *productVariantRepository) GetByIDSForProductMiniCache(ctx context.Context, productIds []uint) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	// Fetch product variants from DB including associated Product data (ID, Name)
	//    Preload is used to load the Product relationship eagerly.
	if err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Product").
		Where("product_id IN ?", productIds).
		Find(&variants).Error; err != nil {
		return nil, err
	}

	// Get product varaint ids
	var productVariantIDs []uint
	for _, variant := range variants {
		productVariantIDs = append(productVariantIDs, variant.ID)
	}
	// Struct to hold reserved quantity info from draft_orders
	type ReservedQtyWithProduct struct {
		ProductVariantID uint
		TotalQuantity    uint
	}

	var reserved1 []ReservedQtyWithProduct
	var reserved2 []ReservedQtyWithProduct

	// Query the total reserved quantity per product variant from draft_orders
	if err := r.db.WithContext(ctx).
		Table("order_items oi").
		Select("oi.product_variant_id, COALESCE(SUM(oi.quantity), 0) as total_quantity").
		Joins("INNER JOIN draft_orders do ON do.id = oi.order_id").
		Where("oi.product_variant_id IN ? AND oi.order_type = ? AND do.to_order IS NULL", productVariantIDs, "draft_order").
		Group("oi.product_variant_id").
		Scan(&reserved1).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Table("order_items oi").
		Select("oi.product_variant_id, COALESCE(SUM(oi.quantity), 0) as total_quantity").
		Joins("INNER JOIN draft_orders do ON do.to_order = oi.order_id").
		Where("oi.product_variant_id IN ? AND oi.order_type = ? AND do.to_order !=0 AND do.to_order IS NOT NULL", productVariantIDs, "order").
		Group("oi.product_variant_id").
		Scan(&reserved2).Error; err != nil {
		return nil, err
	}

	// Build a map for quick lookup of reserved quantities by ProductVariantID
	reservedMap := make(map[uint]uint)
	for _, r := range reserved1 {
		reservedMap[r.ProductVariantID] += r.TotalQuantity
	}
	for _, r := range reserved2 {
		reservedMap[r.ProductVariantID] += r.TotalQuantity
	}
	log.Println(reservedMap)
	// Adjust the quantity of each product variant by subtracting reserved quantity if any
	for i, v := range variants {
		if reservedQty, ok := reservedMap[v.ID]; ok {
			// Subtract reserved quantity from available stock
			variants[i].Quantity = v.Quantity - reservedQty
		}
	}

	return variants, nil
}

func (r *productVariantRepository) GetByIDSForRedisCache(ctx context.Context, productVariantIds []uint) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	// Fetch product variants from DB including associated Product data (ID, Name)
	//    Preload is used to load the Product relationship eagerly.
	if err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Product").
		Where("id IN ?", productVariantIds).
		Find(&variants).Error; err != nil {
		return nil, err
	}
	// Struct to hold reserved quantity info from draft_orders
	type ReservedQtyWithProduct struct {
		ProductVariantID uint
		TotalQuantity    uint
	}

	var reserved1 []ReservedQtyWithProduct
	var reserved2 []ReservedQtyWithProduct

	// Query the total reserved quantity per product variant from draft_orders
	if err := r.db.WithContext(ctx).
		Table("order_items oi").
		Select("oi.product_variant_id, COALESCE(SUM(oi.quantity), 0) as total_quantity").
		Joins("INNER JOIN draft_orders do ON do.id = oi.order_id").
		Where("oi.product_variant_id IN ? AND oi.order_type = ? AND do.to_order IS NULL", productVariantIds, "draft_order").
		Group("oi.product_variant_id").
		Scan(&reserved1).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Table("order_items oi").
		Select("oi.product_variant_id, COALESCE(SUM(oi.quantity), 0) as total_quantity").
		Joins("INNER JOIN draft_orders do ON do.to_order = oi.order_id").
		Where("oi.product_variant_id IN ? AND oi.order_type = ? AND do.to_order !=0 AND do.to_order IS NOT NULL", productVariantIds, "order").
		Group("oi.product_variant_id").
		Scan(&reserved2).Error; err != nil {
		return nil, err
	}

	// Build a map for quick lookup of reserved quantities by ProductVariantID
	reservedMap := make(map[uint]uint)
	for _, r := range reserved1 {
		reservedMap[r.ProductVariantID] += r.TotalQuantity
	}
	for _, r := range reserved2 {
		reservedMap[r.ProductVariantID] += r.TotalQuantity
	}
	// Adjust the quantity of each product variant by subtracting reserved quantity if any
	for i, v := range variants {
		if reservedQty, ok := reservedMap[v.ID]; ok {
			// Subtract reserved quantity from available stock
			variants[i].Quantity = v.Quantity - reservedQty
		}
	}

	return variants, nil
}

func (r *productVariantRepository) GetByIDForRedisCache(ctx context.Context, id uint) (*models.ProductVariant, error) {
	variants, err := r.GetByIDSForRedisCache(ctx, []uint{id})
	if err != nil {
		return nil, err
	}

	if len(variants) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &variants[0], nil
}

func (r *productVariantRepository) IncreaseQuantity(ctx context.Context, quantityMap map[uint]uint) error {
	// Begin Transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for variantID, qty := range quantityMap {
		// Update  quantity for each product variant
		if err := tx.Model(&models.ProductVariant{}).
			Where("id = ?", variantID).
			UpdateColumn("quantity", gorm.Expr("quantity + ?", qty)).Error; err != nil {
			tx.Rollback()
			log.Print(err.Error())
			return customErr.NewError(customErr.INTERNAL_ERROR, "Product Variant Update Failed", http.StatusBadRequest, err)
		}
	}

	return tx.Commit().Error
}

func (r *productVariantRepository) DecreaseQuantity(ctx context.Context, quantityMap map[uint]uint) error {
	// Begin Transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for variantID, qty := range quantityMap {
		// Update  quantity for each product variant
		if err := tx.Model(&models.ProductVariant{}).
			Where("id = ?", variantID).
			UpdateColumn("quantity", gorm.Expr("quantity - ?", qty)).Error; err != nil {
			tx.Rollback()
			log.Print(err.Error())
			return customErr.NewError(customErr.INTERNAL_ERROR, "Product Variant Update Failed", http.StatusBadRequest, err)
		}
	}

	return tx.Commit().Error
}

func (r *productVariantRepository) CheckAndDecreaseStock(ctx context.Context, pvID uint, quantity uint) (*models.ProductVariant, error) {
	var updatedVariant models.ProductVariant

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Lock variant
		var variant models.ProductVariant
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", pvID).
			First(&variant).Error; err != nil {
			return err
		}

		// 2. Lấy reserved quantity
		var reserved1, reserved2 uint
		if err := tx.Table("order_items oi").
			Select("COALESCE(SUM(oi.quantity), 0)").
			Joins("INNER JOIN draft_orders do ON do.id = oi.order_id").
			Where("oi.product_variant_id = ? AND oi.order_type = ? AND to_order IS NULL", pvID, "draft_order").
			Scan(&reserved1).Error; err != nil {
			return err
		}

		if err := tx.Table("order_items oi").
			Select("COALESCE(SUM(oi.quantity), 0)").
			Joins("INNER JOIN draft_orders do ON do.to_order = oi.order_id").
			Where("oi.product_variant_id = ? AND oi.order_type = ? AND do.to_order !=0 AND do.to_order IS NOT NULL", pvID, "order").
			Scan(&reserved2).Error; err != nil {
			return err
		}

		reservedQty := reserved1 + reserved2

		// 3. Check stock
		available := variant.Quantity - reservedQty
		if quantity > available {
			return fmt.Errorf("variant %d out of stock: requested %d, available %d", pvID, quantity, available)
		}

		// 4. Update quantity (trừ luôn)
		if err := tx.Model(&variant).
			Where("id = ?", pvID).
			Update("quantity", gorm.Expr("quantity - ?", quantity)).Error; err != nil {
			return err
		}

		// 5. Load lại variant sau khi trừ để return
		if err := tx.Where("id = ?", pvID).First(&updatedVariant).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &updatedVariant, nil
}

func (r *productVariantRepository) ListByIds(ctx context.Context, list dto.ListProductVariantIds) ([]models.ProductVariant, error) {
	productVariantIds := list.Ids

	var productVariants []models.ProductVariant
	if err := r.db.WithContext(ctx).
		Table("product_variants").
		Preload("Product").
		Preload("Product.Merchant").
		Preload("OptionValues").
		Preload("OptionValues.Option").
		Where("id in ?", productVariantIds).
		Order(fmt.Sprintf("FIELD(id, %s)", uintSliceToCSV(productVariantIds))).
		Find(&productVariants).Error; err != nil {
		return nil, err
	}

	// Struct to hold reserved quantity info from draft_orders
	type ReservedQtyWithProduct struct {
		ProductVariantID uint
		TotalQuantity    uint
	}

	var reserved1 []ReservedQtyWithProduct
	var reserved2 []ReservedQtyWithProduct

	// Query the total reserved quantity per product variant from draft_orders
	if err := r.db.WithContext(ctx).
		Table("order_items oi").
		Select("oi.product_variant_id, COALESCE(SUM(oi.quantity), 0) as total_quantity").
		Joins("INNER JOIN draft_orders do ON do.id = oi.order_id").
		Where("oi.product_variant_id IN ? AND oi.order_type = ? AND do.to_order IS NULL", productVariantIds, "draft_order").
		Group("oi.product_variant_id").
		Scan(&reserved1).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Table("order_items oi").
		Select("oi.product_variant_id, COALESCE(SUM(oi.quantity), 0) as total_quantity").
		Joins("INNER JOIN draft_orders do ON do.to_order = oi.order_id").
		Where("oi.product_variant_id IN ? AND oi.order_type = ? AND do.to_order !=0 AND do.to_order IS NOT NULL", productVariantIds, "order").
		Group("oi.product_variant_id").
		Scan(&reserved2).Error; err != nil {
		return nil, err
	}

	// Build a map for quick lookup of reserved quantities by ProductVariantID
	reservedMap := make(map[uint]uint)
	for _, r := range reserved1 {
		reservedMap[r.ProductVariantID] += r.TotalQuantity
	}
	for _, r := range reserved2 {
		reservedMap[r.ProductVariantID] += r.TotalQuantity
	}

	for i, v := range productVariants {
		if reservedQty, ok := reservedMap[v.ID]; ok {
			// Subtract reserved quantity from available stock
			productVariants[i].Quantity = v.Quantity - reservedQty
		}
	}

	return productVariants, nil
}

func uintSliceToCSV(ids []uint) string {
	strIds := make([]string, len(ids))
	for i, id := range ids {
		strIds[i] = fmt.Sprintf("%d", id)
	}
	return strings.Join(strIds, ",")
}
