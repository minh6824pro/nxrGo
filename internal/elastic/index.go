package elastic

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/minh6824pro/nxrGO/internal/config"
	"log"
	"strings"
)

type ElasticClient struct {
	ES *elasticsearch.Client
}

func NewElasticClient() *ElasticClient {
	ES, err := config.GetElasticClient()
	if err != nil {
		log.Println(err)
	}
	return &ElasticClient{ES: ES}
}
func ProductIndexMapping() string {
	return `{
  "settings": {
    "analysis": {
      "analyzer": {
        "name_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "asciifolding"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id":            { "type": "keyword" },
      "name": {
        "type": "text",
        "analyzer": "name_analyzer",
        "search_analyzer": "name_analyzer"
      },
      "average_rating":{ "type": "double" },
      "number_rating": { "type": "float" },
      "image":         { "type": "keyword" },
      "total_buy":     { "type": "long" },
      "location":      { "type": "text" },
      "merchant":      { "type": "keyword" },
		"geo_point":      { "type": "geo_point" },
      "brand":         { "type": "keyword" },
      "category":      { "type": "keyword" },
      "prices":        { "type": "double" }
    }
  }
}`
}

// EnsureProductIndex kiểm tra index, nếu chưa có thì tạo
func (c *ElasticClient) EnsureProductIndex(ctx context.Context) error {
	// Check xem index đã tồn tại chưa
	exists, err := c.ES.Indices.Exists([]string{"products"}, c.ES.Indices.Exists.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("cannot check index existence: %w", err)
	}
	defer exists.Body.Close()

	if exists.StatusCode == 200 {
		log.Println("Index 'products' already exists, skip creating.")
		return nil
	}

	// Nếu chưa tồn tại thì tạo mới
	mapping := ProductIndexMapping()
	res, err := c.ES.Indices.Create(
		"products",
		c.ES.Indices.Create.WithBody(strings.NewReader(mapping)),
		c.ES.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("cannot create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating index: %s", res.String())
	}

	log.Println("Index 'products' created successfully")
	return nil
}
