package domain

import (
	"context"

	"github.com/google/uuid"
)

type Good struct {
	ID              uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name            string    `gorm:"not null"`
	Category        string    `gorm:"not null"`
	ImageLink       string
	Description     string
	Price           int `gorm:"not null"`
	Volume          int
	QuantityInStock int
}

type OrderItem struct {
	GoodID   uuid.UUID `json:"product_id"`
	Quantity int       `json:"quantity"`
}
type ReserveProductsEvent struct {
	OrderID  uuid.UUID   `json:"order_id"`
	SagaID   uuid.UUID   `json:"saga_id"`
	Products []OrderItem `json:"products"`
}

type ReserveProductsEventReply struct {
	OrderID  uuid.UUID   `json:"order_id"`
	SagaID   uuid.UUID   `json:"saga_id"`
	Products []OrderItem `json:"products"`
	TotalSum int         `json:"total_sum"`
}

type GoodRepository interface {
	SaveGood(ctx context.Context, good *Good) error
	ListGoods(ctx context.Context) ([]*Good, error)
	DeleteGood(ctx context.Context, goodID uuid.UUID) error
	UpdateGood(ctx context.Context, good *Good) error
	ReserveProducts(ctx context.Context, goods []OrderItem) (int, error)
}

type InventoryInteractor interface {
	AddGood(ctx context.Context, name, category, description, imageLink string, price, volume, quantityInStock int) error
	ListProducts(ctx context.Context) ([]*Good, error)
	DeleteGood(ctx context.Context, goodID uuid.UUID) error
	UpdateGood(ctx context.Context, goodID uuid.UUID, name, category, description, imageLink string, price, volume, quantityInStock int) error
	ReserveProducts(ctx context.Context, event ReserveProductsEvent)
}
