package config

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
	)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migration successfully")
}
