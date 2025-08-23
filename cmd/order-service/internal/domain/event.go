package domain

import "github.com/google/uuid"

type ReserveProductsEventReply struct {
	OrderID  uuid.UUID   `json:"order_id"`
	SagaID   uuid.UUID   `json:"saga_id"`
	Products []OrderItem `json:"products"`
	TotalSum int         `json:"total_sum"`
}
