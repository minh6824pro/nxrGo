package event

import (
	"time"
)

type PayOSPaymentCreatedEvent struct {
	Id            int64
	OrderID       uint
	PaymentLink   string
	Total         float64
	PaymentMethod string
	CreatedAt     time.Time
}

type EventPublisher interface {
	PublishPaymentCreated(event PayOSPaymentCreatedEvent) error
	Subscribe(handler func(PayOSPaymentCreatedEvent))
}

type ChannelEventPublisher struct {
	ch chan PayOSPaymentCreatedEvent
}

func NewChannelEventPublisher() *ChannelEventPublisher {
	return &ChannelEventPublisher{ch: make(chan PayOSPaymentCreatedEvent, 10)}
}

func (p *ChannelEventPublisher) PublishPaymentCreated(event PayOSPaymentCreatedEvent) error {
	p.ch <- event
	return nil
}

func (p *ChannelEventPublisher) Subscribe(handler func(PayOSPaymentCreatedEvent)) {
	go func() {
		for e := range p.ch {
			handler(e)
		}
	}()
}
