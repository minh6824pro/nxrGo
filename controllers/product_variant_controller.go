package controllers

import "github.com/minh6824pro/nxrGO/services"

type ProductVariantController struct {
	service services.ProductVariantService
}

func NewProductVariantController(service services.ProductVariantService) *ProductVariantController {
	return &ProductVariantController{}
}
