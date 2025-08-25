package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)

var secret = os.Getenv("PRODUCT_SECRET")
var variantFormat = "id:%d.price:%.2f.merchant_id:%d.timestamp:%d."
var shippingFeeFormat = "merchant:%d.delivery:%d.price:%.2f.lat:%s.lon:%s.timestamp:%d."

func GenerateProductVariantSignature(id uint, price float64, merchantId uint, timestamp int64) string {
	data := fmt.Sprintf(variantFormat, id, price, merchantId, timestamp)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func ValidateProductVariantSignature(id uint, price float64, merchantId uint, timestamp int64, signature string) bool {
	// Check timestamp
	if timestamp < LastResetTime(time.Now()) {
		return false
	}
	expectedSig := GenerateProductVariantSignature(id, price, merchantId, timestamp)
	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

func GenerateShippingFeeSignature(merchantId uint, deliveryId uint, shippingFee float64, lat, lon string, timestamp int64) string {
	data := fmt.Sprintf(shippingFeeFormat, merchantId, deliveryId, shippingFee, lat, lon, timestamp)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func ValidateShippingFeeSignature(merchantId uint, deliveryId uint, shippingFee float64, lat, lon string, timestamp int64, signature string) bool {
	// Check timestamp
	if timestamp < LastResetTime(time.Now()) {
		return false
	}
	expectedSig := GenerateShippingFeeSignature(merchantId, deliveryId, shippingFee, lat, lon, timestamp)
	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// At 4 AM
func LastResetTime(now time.Time) int64 {
	resetToday := time.Date(
		now.Year(), now.Month(), now.Day(),
		4, 0, 0, 0, now.Location(),
	)

	if now.Before(resetToday) {
		resetToday.Add(-24 * time.Hour)
	}
	return resetToday.Unix()
}
