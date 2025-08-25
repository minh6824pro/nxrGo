package config

import (
	"github.com/minh6824pro/nxrGO/internal/database"
	"github.com/minh6824pro/nxrGO/internal/models"
	"log"
)

func AutoMigrate() {
	err := database.DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Merchant{},
		&models.Brand{},
		&models.Category{},
		&models.ProductVariant{},
		&models.VariantOption{},
		&models.VariantOptionValue{},
		&models.Order{},
		&models.OrderItem{},
		&models.PaymentInfo{},
		&models.DraftOrder{},
		&models.Delivery{},
		&models.DeliveryDetail{},
	)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migration successfully")
}
