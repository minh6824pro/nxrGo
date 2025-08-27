package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/minh6824pro/nxrGO/internal/config"
	"github.com/minh6824pro/nxrGO/internal/database"
	"github.com/minh6824pro/nxrGO/internal/elastic/document"
	"github.com/minh6824pro/nxrGO/internal/models"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

type ProductElasticRepo struct {
	es *elasticsearch.Client
	db *gorm.DB
}

var index = "products"

func NewProductElasticRepo() ProductElasticRepository {
	ES, err := config.GetElasticClient()
	if err != nil {
		log.Fatal(err)
	}
	DB := database.DB
	return &ProductElasticRepo{es: ES, db: DB}
}

// Insert document
func (r *ProductElasticRepo) Insert(ctx context.Context, p document.ProductDocument) {
	body, _ := json.Marshal(p)
	res, err := r.es.Index(
		"products",
		strings.NewReader(string(body)),
		r.es.Index.WithDocumentID(p.ID),
		r.es.Index.WithContext(ctx),
		r.es.Index.WithRefresh("true"), // document sẵn sàng tìm kiếm ngay
	)
	if err != nil {
		log.Fatalf("Index error: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Failed to insert document %s: %s", p.ID, res.String())
	} else {
		log.Printf("Document %s inserted successfully", p.ID)
	}
}

func (r *ProductElasticRepo) BulkInsert(ctx context.Context, products []document.ProductDocument) {
	var b strings.Builder

	for _, p := range products {
		// Action line cho Bulk API
		meta := map[string]map[string]string{
			"index": {"_index": "products", "_id": p.ID},
		}
		metaLine, _ := json.Marshal(meta)
		b.Write(metaLine)
		b.WriteByte('\n')

		// Document line
		docLine, _ := json.Marshal(p)
		b.Write(docLine)
		b.WriteByte('\n')
	}

	res, err := r.es.Bulk(
		strings.NewReader(b.String()),
		r.es.Bulk.WithContext(ctx),
		r.es.Bulk.WithRefresh("true"),
	)
	if err != nil {
		log.Fatalf("Bulk insert error: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Bulk insert failed: %s", res.String())
	} else {
		log.Printf("Bulk insert successful")
	}
}

func (r *ProductElasticRepo) DBToElastic(ctx context.Context) {
	var product []models.Product
	err := r.db.Table("products").
		Preload("Merchant").
		Preload("Brand").
		Preload("Category").
		Preload("Variants").
		Preload("Variants.OptionValues").
		Find(&product).Error
	if err != nil {
		log.Fatal(err)
	}
	r.BulkInsert(ctx, MapProductToProductDocument(product))
}

func MapProductToProductDocument(products []models.Product) []document.ProductDocument {
	var docs []document.ProductDocument
	for _, product := range products {
		var priceArray []float64
		for _, variant := range product.Variants {
			price := variant.Price
			priceArray = append(priceArray, price)
		}
		var doc = document.ProductDocument{
			ID:            strconv.Itoa(int(product.ID)),
			Name:          product.Name,
			AverageRating: product.AverageRating,
			NumberRating:  product.NumberRating,
			Image:         product.Image,
			TotalBuy:      product.TotalBuy,
			Location:      product.Merchant.Location,
			GeoPoint:      fmt.Sprintf("%s,%s", product.Merchant.Latitude, product.Merchant.Longitude),
			Merchant:      product.Merchant.Name,
			Brand:         product.Brand.Name,
			Category:      product.Category.Name,
			Price:         priceArray,
		}

		docs = append(docs, doc)
	}
	return docs
}

func (r *ProductElasticRepo) GetProductList(
	ctx context.Context,
	name string,
	priceMin, priceMax *float64,
	priceAsc, totalBuyDesc *bool,
	page, pageSize int, lat, lon *float64,
) (products []document.ProductDocument, totalPages, currentPage int, err error) {

	// --- Build Query DSL ---
	boolQuery := map[string]interface{}{"must": []interface{}{}}

	// --- Name Search ---
	if name != "" {
		// Option 1: Multi-match query (tìm trong nhiều fields)
		nameQuery := map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query": name,
				"fields": []string{
					"name^3",     // boost name field (quan trọng hơn)
					"brand^1",    // search cả trong description
					"category^2", // search trong category
				},
				"type":                 "best_fields",
				"fuzziness":            "AUTO", // cho phép typo
				"minimum_should_match": "75%",
			},
		}

		boolQuery["must"] = append(boolQuery["must"].([]interface{}), nameQuery)
	}

	// --- Price Range ---
	if priceMin != nil || priceMax != nil {
		priceRange := map[string]interface{}{"range": map[string]interface{}{"prices": map[string]interface{}{}}}
		if priceMin != nil {
			priceRange["range"].(map[string]interface{})["prices"].(map[string]interface{})["gte"] = *priceMin
		}
		if priceMax != nil {
			priceRange["range"].(map[string]interface{})["prices"].(map[string]interface{})["lte"] = *priceMax
		}
		boolQuery["must"] = append(boolQuery["must"].([]interface{}), priceRange)
	}

	query := map[string]interface{}{
		"from":  page * pageSize,
		"size":  pageSize,
		"query": map[string]interface{}{"bool": boolQuery},
	}

	// --- Highlight để show matched text ---
	if name != "" {
		query["highlight"] = map[string]interface{}{
			"fields": map[string]interface{}{
				"name": map[string]interface{}{
					"pre_tags":  []string{"<mark>"},
					"post_tags": []string{"</mark>"},
				},
				"description": map[string]interface{}{
					"pre_tags":  []string{"<mark>"},
					"post_tags": []string{"</mark>"},
				},
			},
		}
	}

	// --- Sort ---
	sorts := []interface{}{}

	// Nếu có search text, sort theo relevance score trước
	if name != "" {
		sorts = append(sorts, map[string]interface{}{"_score": map[string]interface{}{"order": "desc"}})
	}

	if priceAsc != nil {
		order := "desc"
		if *priceAsc {
			order = "asc"
		}
		sorts = append(sorts, map[string]interface{}{"prices": map[string]interface{}{"order": order}})
	} else if totalBuyDesc != nil && *totalBuyDesc {
		sorts = append(sorts, map[string]interface{}{"total_buy": map[string]interface{}{"order": "desc"}})
	}

	// Nếu có lat,long thì sort theo khoảng cách
	if lat != nil && lon != nil {
		sorts = append(sorts, map[string]interface{}{
			"_geo_distance": map[string]interface{}{
				"geo_point": map[string]interface{}{
					"lat": *lat,
					"lon": *lon,
				},
				"order":         "asc", // gần trước, xa sau
				"unit":          "km",
				"mode":          "min",
				"distance_type": "arc",
			},
		})
	}

	if len(sorts) > 0 {
		query["sort"] = sorts
	}

	// --- Encode query ---
	body, err := json.Marshal(query)
	if err != nil {
		return nil, 0, 0, err
	}

	// --- Execute Search ---
	res, err := r.es.Search(
		r.es.Search.WithContext(ctx),
		r.es.Search.WithIndex("products"), // dùng index thật
		r.es.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, 0, 0, err
	}
	defer res.Body.Close()

	// --- Parse response ---
	var resp map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, 0, 0, err
	}

	// --- Total hits và page info ---
	hitsRaw, ok := resp["hits"].(map[string]interface{})
	if !ok {
		return nil, 0, 0, fmt.Errorf("resp[\"hits\"] không phải map: %+v", resp)
	}

	totalRaw, ok := hitsRaw["total"].(map[string]interface{})
	if !ok {
		return nil, 0, 0, fmt.Errorf("hits[\"total\"] không phải map: %+v", hitsRaw)
	}

	value, ok := totalRaw["value"].(float64)
	if !ok {
		return nil, 0, 0, fmt.Errorf("total[\"value\"] không phải float64: %+v", totalRaw)
	}

	totalHits := int(value)
	totalPages = (totalHits + pageSize - 1) / pageSize
	currentPage = page

	// --- Parse hits ---
	rawHits, ok := hitsRaw["hits"].([]interface{})
	if !ok {
		return nil, totalPages, currentPage, fmt.Errorf("hits[\"hits\"] không phải array: %+v", hitsRaw)
	}

	products = make([]document.ProductDocument, 0, len(rawHits))
	for _, h := range rawHits {
		hMap, ok := h.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := hMap["_source"]
		if !ok {
			continue
		}

		b, _ := json.Marshal(source)
		var p document.ProductDocument
		if err := json.Unmarshal(b, &p); err != nil {
			continue // skip nếu unmarshal lỗi
		}

		// Thêm highlight info nếu có
		if highlight, exists := hMap["highlight"]; exists {
			// Bạn có thể thêm highlight vào struct ProductDocument
			// hoặc xử lý highlight ở đây
			fmt.Printf("Highlight for product %s: %+v\n", p.Name, highlight)
		}

		products = append(products, p)
	}

	return products, totalPages, currentPage, nil
}
