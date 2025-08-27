package config

import (
	"github.com/elastic/go-elasticsearch/v8"
	"log"
	"os"
)

var EsClient *elasticsearch.Client

func InitElastic() {
	URL := os.Getenv("ELASTICSEARCH_URL")
	//Index := os.Getenv("ELASTICSEARCH_INDEX")

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{URL},
	})
	if err != nil {
		log.Println("Error creating Elasticsearch client: ", err)
	}

	// Ping
	_, err = client.Info()
	if err != nil {
		log.Println("Error pinging Elasticsearch: ", err)
	} else {
		log.Println("✅ Connected to Elasticsearch at ", URL)
	}
	EsClient = client

}

func GetElasticClient() (*elasticsearch.Client, error) {
	return EsClient, nil
}
