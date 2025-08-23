package order

import (
	"context"
	"fmt"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/domain"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/lib"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/lib/logger/sl"
	"immxrtalbeast/order_microservices/internal/pkg/kafka"
	"log/slog"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
		slog.Any("items", items),
	)

	log.Info("creating order")
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "OrderService.CreateOrder")
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.Int("goods.listLength", len(items)),
	)
	defer span.End()

	order := &domain.Order{
		ID:     uuid.New(),
		UserID: userID,
		Items:  items,
		Total:  0,
		Status: "PENDING",
	}

	log = log.With(slog.String("order_id", order.ID.String()))
	log.Debug("order details")

	if _, err := oi.orderRepo.SaveOrder(ctx, order); err != nil {
		log.Error("failed to create order", sl.Err(err))
		span.RecordError(err)
		return uuid.Nil, "", fmt.Errorf("%s: %w", op, err)
	}

	products := lib.ConvertItemstoEventItems(order.Items)

	event := domain.OrderCreatedEvent{
		OrderID:  order.ID,
		Products: products,
		UserID:   order.UserID,
	}

	if err := oi.producer.PublishEventWithEventType(ctx, "OrderCreatedEvent", event, "OrderCreatedEvent"); err != nil {
		log.Error("failed to publish event", sl.Err(err))
		span.RecordError(err)
	}

	log.Info("Order created")
	return order.ID, order.Status, nil
}

func (oi *OrderInteractor) DeleteOrder(ctx context.Context, orderID uuid.UUID) error {
	const op = "service.order.delete"
	log := oi.log.With(
		slog.String("op", op),
		slog.String("order_id", orderID.String()),
	)
	log.Info("deleting order")
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "OrderService.DeleteOrder")
	span.SetAttributes(
		attribute.String("order.id", orderID.String()),
	)
	defer span.End()
	if err := oi.orderRepo.DeleteOrder(ctx, orderID); err != nil {
		log.Error("failed to delete order", sl.Err(err))
		span.RecordError(err)
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("order deleted")
	return nil
}

func (oi *OrderInteractor) Order(ctx context.Context, orderID uuid.UUID) (domain.Order, error) {
	const op = "service.order.get"
	log := oi.log.With(
		slog.String("op", op),
		slog.String("order_id", orderID.String()),
	)

	log.Info("getting order")
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "OrderService.GetOrder")
	span.SetAttributes(
		attribute.String("order.id", orderID.String()),
	)
	defer span.End()
	order, err := oi.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		log.Error("failed to get order", sl.Err(err))
		span.RecordError(err)
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

	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "OrderService.ListOrdersByUser")
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
	)
	defer span.End()

	orders, err := oi.orderRepo.ListOrdersByUser(ctx, userID, limit, offset)
	if err != nil {
		log.Error("failed to get a list of orders by user", sl.Err(err))
		span.RecordError(err)
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
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "OrderService.UpdateOrderStatus")
	span.SetAttributes(
		attribute.String("order.id", orderID.String()),
		attribute.String("order.status", status),
	)
	defer span.End()

	if err := oi.orderRepo.UpdateOrderStatus(ctx, orderID, status); err != nil {
		log.Error("failed to update order status", sl.Err(err))
		span.RecordError(err)
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("order status updated")
	return nil
}

func (oi *OrderInteractor) SetTotalSum(ctx context.Context, event domain.ReserveProductsEventReply) error {
	const op = "service.order.set-total-sum"
	log := oi.log.With(
		slog.String("op", op),
		slog.String("order_id", event.OrderID.String()),
		slog.String("saga_id", event.SagaID.String()),
		slog.Any("products", event.Products),
	)
	log.Info("setting sum")
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "OrderService.SetTotalSum")
	span.SetAttributes(
		attribute.String("saga.id", event.SagaID.String()),
	)
	defer span.End()
	if err := oi.orderRepo.SetTotalSum(ctx, event.OrderID, event.TotalSum); err != nil {
		log.Error("failed to update order sum", sl.Err(err))
		span.RecordError(err)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
