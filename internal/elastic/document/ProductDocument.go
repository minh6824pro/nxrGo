package document

type ProductDocument struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	AverageRating float64   `json:"average_rating"`
	NumberRating  float32   `json:"number_rating"`
	Image         string    `json:"image"`
	TotalBuy      uint      `json:"total_buy"`
	Location      string    `json:"location"`
	Merchant      string    `json:"merchant"`
	GeoPoint      string    `json:"geo_point"`
	Brand         string    `json:"brand"`
	Category      string    `json:"category"`
	Price         []float64 `json:"prices"`
}
