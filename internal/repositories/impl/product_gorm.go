package impl

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/models/CacheModel"
	"github.com/minh6824pro/nxrGO/internal/repositories"

	"github.com/minh6824pro/nxrGO/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

type productGormRepository struct {
	db *gorm.DB
}

func NewProductGormRepository(db *gorm.DB) repositories.ProductRepository {
	return &productGormRepository{db}
}

func (r *productGormRepository) Create(ctx context.Context, p *models.Product) (*models.Product, error) {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return nil, err
	}
	// Preload ngay sau khi tạo
	if err := r.db.WithContext(ctx).
		Preload("Brand").
		Preload("Category").
		Preload("Merchant").
		First(p, p.ID).Error; err != nil {
		return nil, err
	}

	return p, nil
}
func (r *productGormRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, p *models.Product) (*models.Product, error) {
	if err := tx.WithContext(ctx).Create(p).Error; err != nil {
		return nil, errors.NewError(errors.INTERNAL_ERROR, "Unexpected Error", http.StatusInternalServerError, err)
	}

	return p, nil
}

func (r *productGormRepository) GetByID(ctx context.Context, id uint) (*models.Product, error) {

	var p models.Product
	err := r.db.WithContext(ctx).
		Preload("Brand").
		Preload("Merchant").
		Preload("Category").
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return db.Select("product_variants.id, product_variants.product_id, product_variants.price, product_variants.image, product_variants.version, get_available_quantity(product_variants.id) as quantity")
		}).
		Preload("Variants.OptionValues").
		Preload("Variants.OptionValues.Option").
		First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil

}

func (r *productGormRepository) Update(ctx context.Context, p *models.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *productGormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Product{}, id).Error
}

func (r *productGormRepository) List(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("Brand").
		Preload("Category").
		Preload("Variants").
		Preload("Variants.OptionValues").
		Preload("Variants.OptionValues.Option").
		Find(&products).Error; err != nil {
		return nil, errors.NewError(errors.UNEXPECTED_ERROR, "Unexpected error", http.StatusInternalServerError, err)
	}

	return products, nil
}

func (r *productGormRepository) ListWithPagination(ctx context.Context, page int, size int) ([]models.Product, int64, int64, error) {
	var products []models.Product
	var total int64

	// Count record
	if err := r.db.WithContext(ctx).Model(&models.Product{}).Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	offset := (page - 1) * size

	if err := r.db.WithContext(ctx).
		Limit(size).
		Offset(offset).
		Preload("Variants", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "price", "product_id")
		}).
		Find(&products).Error; err != nil {
		return nil, 0, 0, err
	}

	// Calc total page
	totalPages := (total + int64(size) - 1) / int64(size)

	return products, total, totalPages, nil
}

func (r *productGormRepository) GetAllProductId(ctx context.Context) ([]uint, error) {
	var ids []uint
	if err := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
func (r *productGormRepository) GetByIdPreloadVariant(ctx context.Context, id uint) (*models.Product, error) {
	var p models.Product
	err := r.db.WithContext(ctx).
		Preload("Variants").
		First(&p, id).Error

	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productGormRepository) GetProductListFilter(
	ctx context.Context,
	priceMin, priceMax *float64,
	priceAsc *bool,
	totalBuyDescStr *bool,
	page, pageSize int) ([]CacheModel.ListProductQueryCache, int, error) { // thêm totalPage trả về

	// Subquery: filter variant + xếp hạng theo giá
	ranked := r.db.Table("product_variants v").
		Select(`
            v.id,
            v.product_id,
            v.price,
            ROW_NUMBER() OVER (PARTITION BY v.product_id ORDER BY v.price ASC, v.id ASC) AS rn
        `)

	if priceMin != nil {
		ranked = ranked.Where("v.price >= ?", *priceMin)
	}
	if priceMax != nil {
		ranked = ranked.Where("v.price <= ?", *priceMax)
	}

	// Main query base
	query := r.db.WithContext(ctx).
		Table("products p").
		Joins("JOIN (?) rv ON rv.product_id = p.id AND rv.rn = 1", ranked).
		Where("p.deleted_at IS NULL AND p.active = 1")

	// Count total item
	var totalItem int64
	if err := query.Count(&totalItem).Error; err != nil {
		return nil, 0, err
	}

	// Order
	if priceAsc != nil {
		if *priceAsc {
			query = query.Order("rv.price ASC")
		} else {
			query = query.Order("rv.price DESC")
		}
	} else if totalBuyDescStr != nil && *totalBuyDescStr {
		query = query.Order("p.total_buy DESC")
	} else {
		query = query.Order("p.created_at DESC")
	}

	// Phân trang
	offset := page * pageSize
	query = query.Select(`
            p.id AS product_id,
            rv.id AS variant_id,
            p.total_buy
        `).Limit(pageSize).Offset(offset)

	var results []CacheModel.ListProductQueryCache
	if err := query.Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	// Tính total page
	totalPage := int((totalItem + int64(pageSize) - 1) / int64(pageSize))

	return results, totalPage, nil
}

func (r *productGormRepository) GetProductListFilterOptimized(
	ctx context.Context,
	priceMin, priceMax *float64,
	priceAsc *bool,
	totalBuyDescStr *bool,
	page, pageSize int,
) ([]CacheModel.ListProductQueryCache, int, error) {

	// subquery với ROW_NUMBER() để chọn 1 variant duy nhất cho mỗi product (order by price, id)
	ranked := r.db.Raw(`
  SELECT id, product_id, price, available_qty FROM (
    SELECT
      v.id,
      v.product_id,
      v.price,
      get_available_quantity(v.id) AS available_qty,
      ROW_NUMBER() OVER (PARTITION BY v.product_id ORDER BY v.price ASC, v.id ASC) AS rn
    FROM product_variants v
    WHERE get_available_quantity(v.id) > 0
  ) t
  WHERE t.rn = 1
`)
	// Filter giá min/max
	if priceMin != nil {
		ranked = ranked.Where("v.price >= ?", *priceMin)
	}
	if priceMax != nil {
		ranked = ranked.Where("v.price <= ?", *priceMax)
	}

	// Main query
	query := r.db.WithContext(ctx).
		Table("products p").
		Joins("JOIN (?) rv ON rv.product_id = p.id", ranked).
		Where("p.deleted_at IS NULL AND p.active = 1")
	// Count total
	var totalItem int64
	if err := query.Select("COUNT(DISTINCT p.id)").Count(&totalItem).Error; err != nil {
		return nil, 0, err
	}

	// Order
	if priceAsc != nil {
		if *priceAsc {
			query = query.Order("rv.price ASC")
		} else {
			query = query.Order("rv.price DESC")
		}
	} else if totalBuyDescStr != nil && *totalBuyDescStr {
		query = query.Order("p.total_buy DESC")
	} else {
		query = query.Order("p.created_at DESC")
	}

	// Pagination
	offset := page * pageSize
	query = query.Select(`
			p.id AS product_id,
			rv.id AS variant_id,
			rv.available_qty as quantity,
			p.total_buy
		`).
		Limit(pageSize).
		Offset(offset)

	// Execute
	var results []CacheModel.ListProductQueryCache
	if err := query.Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	totalPage := int((totalItem + int64(pageSize) - 1) / int64(pageSize))
	return results, totalPage, nil
}
