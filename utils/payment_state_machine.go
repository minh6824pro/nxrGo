package utils

import "github.com/minh6824pro/nxrGO/models"

type PaymentEvent string

const (
	EventPaySuccess PaymentEvent = "pay_success"
	EventPayFailed  PaymentEvent = "pay_failed"
	EventPayCancel  PaymentEvent = "pay_cancel"
)

// Payment State Machine
var paymentStateMachine = map[models.PaymentStatus]map[PaymentEvent]models.PaymentStatus{
	models.PaymentPending: {
		EventPaySuccess: models.PaymentSuccess,
		EventPayFailed:  models.PaymentFailed,
		EventPayCancel:  models.PaymentCanceled,
	},
}

// CanTransitionPayment checks if a payment status can transition via a given event
func CanTransitionPayment(current models.PaymentStatus, event PaymentEvent) (models.PaymentStatus, bool) {
	if nextStates, ok := paymentStateMachine[current]; ok {
		if next, ok := nextStates[event]; ok {
			return next, true
		}
	}
	return "", false
}
