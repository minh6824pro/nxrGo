package CacheModel

type ProductMiniCache struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	AverageRating float64 `json:"average_rating"`
	NumberRating  float32 `json:"number_rating"`
	Image         string  `json:"image"`
	TotalBuy      uint    `json:"total_buy"`
	TotalQuantity uint    `json:"total_quantity"`
	Price         float64 `json:"price"`
}
