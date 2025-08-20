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

func GenerateProductVariantSignature(id uint, price float64, merchantId uint, timestamp time.Time) string {
	data := fmt.Sprintf(variantFormat, id, price, merchantId, timestamp.Unix())
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func ValidateProductVariantSignature(id uint, price float64, merchantId uint, timestamp time.Time, signature string) bool {
	expectedSig := GenerateProductVariantSignature(id, price, merchantId, timestamp)
	return hmac.Equal([]byte(expectedSig), []byte(signature))
}
