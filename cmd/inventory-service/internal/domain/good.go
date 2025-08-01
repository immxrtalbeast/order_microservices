package domain

import (
	"context"

	"github.com/google/uuid"
)

type Good struct {
	ID              uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name            string    `gorm:"not null"`
	ImageLink       string
	Description     string
	Price           int `gorm:"not null"`
	QuantityInStock int
}

type GoodRepository interface {
	SaveGood(ctx context.Context, good *Good) error
	ListGoods(ctx context.Context) ([]*Good, error)
	DeleteGood(ctx context.Context, goodID uuid.UUID) error
	UpdateGood(ctx context.Context, good *Good) error
}

type InventoryInteractor interface {
	AddGood(ctx context.Context, name string, description string, imageLink string, price int, quantityInStock int) error
	ListProducts(ctx context.Context) ([]*Good, error)
	DeleteGood(ctx context.Context, goodID uuid.UUID) error
	UpdateGood(ctx context.Context, goodID uuid.UUID, name string, description string, imageLink string, price int, quantityInStock int) error
}
