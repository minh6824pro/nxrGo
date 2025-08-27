package elastic

import (
	"context"
	"github.com/minh6824pro/nxrGO/internal/elastic/document"
)

type ProductElasticRepository interface {
	Insert(ctx context.Context, p document.ProductDocument)
	BulkInsert(ctx context.Context, products []document.ProductDocument)
	DBToElastic(ctx context.Context)
	GetProductList(
		ctx context.Context,
		name string,
		priceMin, priceMax *float64,
		priceAsc, totalBuyDesc *bool,
		page, pageSize int, lat, lon *float64,
	) (products []document.ProductDocument, totalPages, currentPage int, err error)
}
