package order

import (
	"context"
	"fmt"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

type OrderInteractor struct {
	orderRepo domain.OrderRepository
}

func NewOrderInteractor(orderRepo domain.OrderRepository) *OrderInteractor {
	return &OrderInteractor{orderRepo: orderRepo}
}

func (oi *OrderInteractor) CreateOrder(ctx context.Context, userID uuid.UUID, items []domain.OrderItem) (uuid.UUID, string, error) {
	const op = "service.order.create"

	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	order := &domain.Order{
		ID:        uuid.New(),
		UserID:    userID,
		Items:     items,
		Total:     total,
		Status:    "CREATED",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if _, err := oi.orderRepo.SaveOrder(ctx, order); err != nil {
		return uuid.Nil, "", fmt.Errorf("%s: %w", op, err)
	}
	return order.ID, order.Status, nil
}

func (oi *OrderInteractor) GetOrder(ctx context.Context, orderID uuid.UUID) (domain.Order, error) {
	const op = "service.order.get"
	order, err := oi.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("%s: %w", op, err)
	}
	return order, nil
}

func (oi *OrderInteractor) ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Order, error) {
	const op = "service.order.list"
	orders, err := oi.orderRepo.ListOrdersByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return orders, nil
}

func (oi *OrderInteractor) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	const op = "service.order.update_status"
	if err := oi.orderRepo.UpdateOrderStatus(ctx, orderID, status); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
