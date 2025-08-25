package dto

type PayOSCreatePaymentRequest struct {
	OrderCode   int     `json:"orderCode"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	BuyerEmail  string  `json:"buyerEmail,omitempty"`
	BuyerPhone  string  `json:"buyerPhone,omitempty"`
	ReturnUrl   string  `json:"returnUrl"`
	CancelUrl   string  `json:"cancelUrl"`
}
