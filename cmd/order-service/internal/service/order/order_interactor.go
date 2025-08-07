package order

import (
	"context"
	"fmt"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/domain"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/lib/logger/sl"
	"immxrtalbeast/order_microservices/internal/pkg/kafka"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type OrderInteractor struct {
	orderRepo domain.OrderRepository
	log       *slog.Logger
	producer  *kafka.Producer
}

func NewOrderInteractor(orderRepo domain.OrderRepository, log *slog.Logger, producer *kafka.Producer) *OrderInteractor {
	return &OrderInteractor{orderRepo: orderRepo, log: log, producer: producer}
}

func (oi *OrderInteractor) CreateOrder(ctx context.Context, userID uuid.UUID, items []domain.OrderItem) (uuid.UUID, string, error) {
	const op = "service.order.create"
	log := oi.log.With(
		slog.String("op", op),
		slog.String("user_id", userID.String()),
		slog.Int("items", len(items)),
	)

	log.Info("creating order")
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	order := &domain.Order{
		ID:        uuid.New(),
		UserID:    userID,
		Items:     items,
		Total:     total,
		Status:    "PENDING",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	log = log.With(slog.String("order_id", order.ID.String()))
	log.Debug("order details",
		slog.Float64("total", total),
	)

	if _, err := oi.orderRepo.SaveOrder(ctx, order); err != nil {
		log.Error("failed to create order", sl.Err(err))
		return uuid.Nil, "", fmt.Errorf("%s: %w", op, err)
	}

	event := domain.OrderCreatedEvent{
		OrderID:  order.ID,
		Products: order.Items,
		UserID:   order.UserID,
	}

	if err := oi.producer.PublishEventWithEventType(ctx, "OrderCreatedEvent", event, "OrderCreatedEvent"); err != nil {
		log.Error("failed to publish event", sl.Err(err))
	}

	log.Info("Order created")
	return order.ID, order.Status, nil
}

func (oi *OrderInteractor) Order(ctx context.Context, orderID uuid.UUID) (domain.Order, error) {
	const op = "service.order.get"
	log := oi.log.With(
		slog.String("op", op),
		slog.String("order_id", orderID.String()),
	)

	log.Info("getting order")

	order, err := oi.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		log.Error("failed to get order", sl.Err(err))
		return domain.Order{}, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("Order getting")
	return order, nil
}

func (oi *OrderInteractor) ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Order, error) {
	const op = "service.order.list_by_user"
	log := oi.log.With(
		slog.String("op", op),
		slog.String("user_id", userID.String()),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	log.Info("getting a list of orders by user")

	orders, err := oi.orderRepo.ListOrdersByUser(ctx, userID, limit, offset)
	if err != nil {
		log.Error("failed to get a list of orders by user", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("orders listed", slog.Int("count", len(orders)))
	return orders, nil
}

func (oi *OrderInteractor) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	const op = "service.order.update_status"
	log := oi.log.With(
		slog.String("op", op),
		slog.String("order_id", orderID.String()),
		slog.String("status", status),
	)
	log.Info("updating order status")
	if err := oi.orderRepo.UpdateOrderStatus(ctx, orderID, status); err != nil {
		log.Error("failed to update order status", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("order status updated")
	return nil
}
