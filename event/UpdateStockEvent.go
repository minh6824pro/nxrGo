package event

import (
	"github.com/minh6824pro/nxrGO/models"
	"log"
	"sync"
)

type UpdateStockAggregator struct {
	mu   sync.Mutex
	data map[uint]int // productVariantID -> quantity
}

func NewUpdateStockAggregator() *UpdateStockAggregator {
	return &UpdateStockAggregator{
		data: make(map[uint]int),
	}
}

// AddOrder receive order & update to map
func (u *UpdateStockAggregator) AddOrder(order models.Order) {
	u.mu.Lock()
	defer u.mu.Unlock()

	for _, oi := range order.OrderItems {
		u.data[oi.ProductVariantID] += int(oi.Quantity)
	}

	log.Println(u.data)
}

func (u *UpdateStockAggregator) AddStock(id uint, quantity int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.data[id] += quantity

}

// Flush to db and reset map
func (u *UpdateStockAggregator) Flush() map[uint]int {
	u.mu.Lock()
	defer u.mu.Unlock()

	flushed := u.data
	u.data = make(map[uint]int)
	return flushed
}

func (u *UpdateStockAggregator) RemoveStock(id uint, quantity int) {

}
