package domain

import "github.com/google/uuid"

type OrderCreatedEvent struct {
	OrderID  uuid.UUID   `json:"order_id"`
	Products []OrderItem `json:"products"`
	UserID   uuid.UUID   `json:"user_id"`
}

type ProductsReservedEvent struct {
	OrderID  uuid.UUID   `json:"order_id"`
	SagaID   uuid.UUID   `json:"saga_id"`
	Products []OrderItem `json:"products"`
}

type Event interface {
	EventType() string
}

func (e OrderCreatedEvent) EventType() string {
	return "OrderCreatedEvent"
}

func (e ProductsReservedEvent) EventType() string {
	return "ProductsReservedEvent"
}
