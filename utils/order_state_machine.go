package utils

import "github.com/minh6824pro/nxrGO/models"

type OrderEvent string

const (
	EventConfirm       OrderEvent = "confirm"
	EventProcess       OrderEvent = "process"
	EventShip          OrderEvent = "ship"
	EventDeliver       OrderEvent = "deliver"
	EventCancel        OrderEvent = "cancel"
	EventRequestReturn OrderEvent = "request_return"
	EventReturn        OrderEvent = "return"
	EventRefund        OrderEvent = "refund"
)

// StateMachine defines valid transitions
var stateMachine = map[models.OrderStatus]map[OrderEvent]models.OrderStatus{
	models.OrderStatePending: {
		EventConfirm: models.OrderStateConfirmed,
		EventCancel:  models.OrderStateCancelled,
	},
	models.OrderStateConfirmed: {
		EventProcess: models.OrderStateProcessing,
		EventCancel:  models.OrderStateCancelled,
	},
	models.OrderStateProcessing: {
		EventShip:   models.OrderStateShipped,
		EventCancel: models.OrderStateCancelled,
	},
	models.OrderStateShipped: {
		EventDeliver:       models.OrderStateDelivered,
		EventRequestReturn: models.OrderStateReturnRequested,
	},
	models.OrderStateDelivered: {
		EventRequestReturn: models.OrderStateReturnRequested,
	},
	models.OrderStateReturnRequested: {
		EventReturn: models.OrderStateReturned,
	},
	models.OrderStateReturned: {
		EventRefund: models.OrderStateRefunded,
	},
}

func CanTransition(current models.OrderStatus, event OrderEvent) (models.OrderStatus, bool) {
	if nextStates, ok := stateMachine[current]; ok {
		if next, ok := nextStates[event]; ok {
			return next, true
		}
	}
	return "", false
}
