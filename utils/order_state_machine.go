package utils

import (
	"fmt"
	customErr "github.com/minh6824pro/nxrGO/errors"
	"github.com/minh6824pro/nxrGO/models"
	"net/http"
)

type OrderEvent string

const (
	EventConfirm       OrderEvent = "confirm"
	EventProcess       OrderEvent = "process"
	EventShip          OrderEvent = "ship"
	EventDeliver       OrderEvent = "deliver"
	EventDone          OrderEvent = "done"
	EventCancel        OrderEvent = "cancel"
	EventRequestReturn OrderEvent = "request_return"
	EventReturnShip    OrderEvent = "return_ship"
	EventReturn        OrderEvent = "return"
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
		EventDone:          models.OrderStateDone,
		EventRequestReturn: models.OrderStateReturnRequested,
	},
	models.OrderStateReturnRequested: {
		EventReturnShip: models.OrderStateReturnShipping,
	},
	models.OrderStateReturnShipping: {
		EventReturn: models.OrderStateReturned,
	},
	models.OrderStateDone: {},
}

func CanTransitionOrder(current models.OrderStatus, event OrderEvent) (models.OrderStatus, error) {
	if nextStates, ok := stateMachine[current]; ok {
		if next, ok := nextStates[event]; ok {
			return next, nil
		}
	}
	return "", customErr.NewError(customErr.BAD_REQUEST, fmt.Sprintf("Cant transition from %s to %s", current, event), http.StatusBadRequest, nil)
}
