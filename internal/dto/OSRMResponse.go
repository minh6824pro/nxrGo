package dto

type OSRMResponse struct {
	Routes []struct {
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
			Type        string      `json:"type"`
		} `json:"geometry"`
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
	} `json:"routes"`
	Code string `json:"code"`
}
