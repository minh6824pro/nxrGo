package configs

import (
	"github.com/minh6824pro/nxrGO/database"
	"github.com/minh6824pro/nxrGO/models"
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
	)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migration successfully")
}
