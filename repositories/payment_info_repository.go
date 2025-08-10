package repositories

import (
	"context"
	"github.com/minh6824pro/nxrGO/models"
)

type PaymentInfoRepository interface {
	Create(ctx context.Context, payment *models.PaymentInfo) error
	Save(ctx context.Context, payment *models.PaymentInfo) error
}
