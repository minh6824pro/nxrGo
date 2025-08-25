package dto

type PayOSCreatePaymentResponse struct {
	Bin           string `json:"bin"`           // Mã BIN ngân hàng
	AccountNumber string `json:"accountNumber"` // Số tài khoản của kênh thanh toán
	AccountName   string `json:"accountName"`   // Tên chủ tài khoản của kênh thanh toán
	Amount        int    `json:"amount"`        // Tổng tiền đơn hàng
	Description   string `json:"description"`   // Mô tả đơn hàng, được dùng làm nội dung chuyển khoản
	OrderCode     int64  `json:"orderCode"`     // Mã đơn hàng
	Currency      string `json:"currency"`      // Đơn vị tiền tệ
	PaymentLinkId string `json:"paymentLinkId"` // Mã link thanh toán
	Status        string `json:"status"`        // Trạng thái của link thanh toán
	CheckoutUrl   string `json:"checkoutUrl"`   // Đường dẫn trang thanh toán
	QRCode        string `json:"qrCode"`        // Mã QR thanh toán
}
