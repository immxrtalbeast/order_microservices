package domain

import "github.com/google/uuid"

type OrderItem struct {
	GoodID   uuid.UUID `json:"product_id"`
	Quantity int       `json:"quantity"`
}

type ReserveItemsCommand struct {
	OrderID  uuid.UUID   `json:"order_id"`
	SagaID   uuid.UUID   `json:"saga_id"`
	Products []OrderItem `json:"products"`
}

type CancelOrderCommand struct {
	OrderID uuid.UUID `json:"order_id"`
	SagaID  uuid.UUID `json:"saga_id"`
}

type ReleaseInventoryCommand struct {
	OrderID uuid.UUID `json:"order_id"`
	SagaID  uuid.UUID `json:"saga_id"`
}

type CompensateOrderCommand struct {
	OrderID uuid.UUID `json:"order_id"`
	SagaID  uuid.UUID `json:"saga_id"`
}
