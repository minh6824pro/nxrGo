package impl

import (
	"github.com/payOSHQ/payos-lib-golang"
	"time"
)

func CreatePayOSPayment(orderCode int, amount float64, items []payos.Item, description, returnUrl, cancelUrl string) (*payos.CheckoutResponseDataType, error) {
	//payload := dto.PayOSCreatePaymentRequest{
	//	OrderCode:   orderCode,
	//	Amount:      amount,
	//	Description: description,
	//	ReturnUrl:   returnUrl,
	//	CancelUrl:   cancelUrl,
	//}
	expiredAt := int(time.Now().Add(5 * time.Minute).Unix())

	body := payos.CheckoutRequestType{
		OrderCode:   int64(orderCode) + 123456,
		Amount:      int(amount),
		Items:       items,
		Description: description,
		ReturnUrl:   returnUrl,
		CancelUrl:   cancelUrl,
		ExpiredAt:   &expiredAt,
	}

	resp, err := payos.CreatePaymentLink(body)
	if err != nil {
		return nil, err
	}

	return resp, nil

}
