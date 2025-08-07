package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OrderCreatedEvent struct {
	OrderID  uuid.UUID   `json:"order_id"`
	Products []OrderItem `json:"products"`
	UserID   uuid.UUID   `json:"user_id"`
}

type Order struct {
	ID        uuid.UUID   `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    uuid.UUID   `gorm:"type:uuid;not null;index"` // Связь с пользователем
	Items     []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	Total     float64     `gorm:"type:decimal(10,2);not null"`
	Status    string      `gorm:"type:varchar(20);not null;default:'CREATED'"`
	CreatedAt time.Time   `gorm:"autoCreateTime"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime"`
}

type OrderItem struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null;index"` // Внешний ключ
	ProductID uuid.UUID `gorm:"type:uuid;not null"`
	Quantity  int       `gorm:"not null"`
	Price     float64   `gorm:"type:decimal(10,2);not null"`
}

type OrderItemEvent struct {
	GoodID   uuid.UUID `json:"product_id"`
	Quantity int       `json:"quantity"`
}

type OrderRepository interface {
	SaveOrder(ctx context.Context, order *Order) (uuid.UUID, error)
	GetOrder(ctx context.Context, orderID uuid.UUID) (Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error
	ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Order, error)
}

type OrderInteractor interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, orderItem []OrderItem) (uuid.UUID, string, error)
	Order(ctx context.Context, orderID uuid.UUID) (Order, error)
	ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Order, error)
}
