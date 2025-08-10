package dto

type WebhookType struct {
	Code      string           `json:"code"`      // Mã lỗi
	Desc      string           `json:"desc"`      // Mô tả lỗi
	Success   bool             `json:"success"`   // Trạng thái của webhook
	Data      *WebhookDataType `json:"data"`      // Dữ liệu webhook
	Signature string           `json:"signature"` // Chữ ký số của dữ liệu webhook, dùng để kiểm tra tính toàn vẹn của dữ liệu
}

type WebhookDataType struct {
	OrderCode              int64   `json:"orderCode"`              // Mã đơn hàng
	Amount                 int     `json:"amount"`                 // Số tiền chuyển khoản
	Description            string  `json:"description"`            // Mô tả đơn hàng, được dùng làm nội dung chuyển khoản
	AccountNumber          string  `json:"accountNumber"`          // Số tài khoản của kênh thanh toán
	Reference              string  `json:"reference"`              // Mã tham chiếu của giao dịch
	TransactionDateTime    string  `json:"transactionDateTime"`    // Thời gian giao dịch
	Currency               string  `json:"currency"`               // Đơn vị tiền tệ
	PaymentLinkId          string  `json:"paymentLinkId"`          // Mã link thanh toán
	Code                   string  `json:"code"`                   // Mã lỗi
	Desc                   string  `json:"desc"`                   // Mô tả lỗi
	CounterAccountBankId   *string `json:"counterAccountBankId"`   // Mã ngân hàng đối ứng
	CounterAccountBankName *string `json:"counterAccountBankName"` // Tên ngân hàng đối ứng
	CounterAccountName     *string `json:"counterAccountName"`     // Tên chủ tài khoản đối ứng
	CounterAccountNumber   *string `json:"counterAccountNumber"`   // Số tài khoản đối ứng
	VirtualAccountName     *string `json:"virtualAccountName"`     // Tên chủ tài khoản ảo
	VirtualAccountNumber   *string `json:"virtualAccountNumber"`   // Số tài khoản ảo
}
