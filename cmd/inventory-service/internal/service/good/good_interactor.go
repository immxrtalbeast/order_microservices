package good

import (
	"context"
	"fmt"
	"immxrtalbeast/order_microservices/internal/pkg/kafka"
	"immxrtalbeast/order_microservices/inventory-service/internal/domain"
	"immxrtalbeast/order_microservices/inventory-service/internal/lib/logger/sl"
	"log/slog"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type GoodInteractor struct {
	log      *slog.Logger
	goodRepo domain.GoodRepository
	producer *kafka.Producer
}

func NewGoodInteractor(goodRepo domain.GoodRepository, log *slog.Logger, producer *kafka.Producer) *GoodInteractor {
	return &GoodInteractor{goodRepo: goodRepo, log: log, producer: producer}
}

func (gi *GoodInteractor) AddGood(ctx context.Context, name string, description string, imageLink string, price int, volume int, quantityInStock int) error {
	const op = "service.good.save"
	log := gi.log.With(
		slog.String("op", op),
		slog.String("good", name),
		slog.Int("volume", volume),
	)
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "InvetoryService.AddGood")
	span.SetAttributes(
		attribute.String("good.name", name),
		attribute.Int("good.volume", volume),
	)
	defer span.End()
	log.Info("adding good")
	good := &domain.Good{
		Name:            name,
		Description:     description,
		ImageLink:       imageLink,
		Price:           price,
		Volume:          volume,
		QuantityInStock: quantityInStock,
	}

	if err := gi.goodRepo.SaveGood(ctx, good); err != nil {
		log.Error("failed to save good", sl.Err(err))
		span.RecordError(err)
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("Good saved")
	return nil
}

func (gi *GoodInteractor) ListProducts(ctx context.Context) ([]*domain.Good, error) {
	const op = "service.good.list"
	log := gi.log.With(
		slog.String("op", op),
	)
	log.Info("getting list of goods")
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "InvetoryService.ListProducts")
	defer span.End()
	goods, err := gi.goodRepo.ListGoods(ctx)
	if err != nil {
		log.Error("failed to get list of goods", sl.Err(err))
		span.RecordError(err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("list provided")
	return goods, nil
}

func (gi *GoodInteractor) DeleteGood(ctx context.Context, goodID uuid.UUID) error {
	const op = "service.good.delete"
	log := gi.log.With(
		slog.String("op", op),
		slog.String("goodID", goodID.String()),
	)
	log.Info("deleting good")
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "InvetoryService.DeleteGood")
	span.SetAttributes(
		attribute.String("good.id", goodID.String()),
	)
	defer span.End()
	if err := gi.goodRepo.DeleteGood(ctx, goodID); err != nil {
		log.Error("failed to delete good", sl.Err(err))
		span.RecordError(err)
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("good deleted")
	return nil
}

func (gi *GoodInteractor) UpdateGood(ctx context.Context, goodID uuid.UUID, name string, description string, imageLink string, price int, volume int, quantityInStock int) error {
	const op = "service.good.update"
	log := gi.log.With(
		slog.String("op", op),
		slog.String("goodID", goodID.String()),
		slog.String("name", name),
		slog.String("description", description),
		slog.String("imageLink", imageLink),
		slog.Int("price", price),
		slog.Int("volume", volume),
		slog.Int("quantity", quantityInStock),
	)
	log.Info("updating good")
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "InvetoryService.UpdateGood")
	span.SetAttributes(
		attribute.String("good.id", goodID.String()),
	)
	defer span.End()
	good := &domain.Good{
		ID:              goodID,
		Name:            name,
		Description:     description,
		ImageLink:       imageLink,
		Price:           price,
		Volume:          volume,
		QuantityInStock: quantityInStock,
	}
	if err := gi.goodRepo.UpdateGood(ctx, good); err != nil {
		log.Error("failed to update good", sl.Err(err))
		span.RecordError(err)
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("good updated")
	return nil
}

func (gi *GoodInteractor) ReserveProducts(ctx context.Context, event domain.ReserveProductsEvent) {
	const op = "service.good.update"
	log := gi.log.With(
		slog.String("op", op),
		slog.String("order_id", event.OrderID.String()),
		slog.String("saga_id", event.SagaID.String()),
		slog.Any("products", event.Products),
	)
	log.Info("reserving goods")
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "InvetoryService.ReserveProducts")
	span.SetAttributes(
		attribute.String("saga.id", event.SagaID.String()),
	)
	defer span.End()
	if err := gi.goodRepo.ReserveProducts(ctx, event.Products); err != nil {
		span.RecordError(err)
		log.Error("failed to reserve products", sl.Err(err))
		if err := gi.producer.PublishEventWithEventType(ctx, "InventoryReservedEventFailed", event, "InventoryReservedEventFailed"); err != nil {
			span.RecordError(err)
			log.Error("Failed to publish event", sl.Err(err))
		}
		return
	}
	log.Info("goods reserved")
	if err := gi.producer.PublishEventWithEventType(ctx, "InventoryReservedEvent", event, "InventoryReservedEvent"); err != nil {
		span.RecordError(err)
		log.Error("Failed to publish event", sl.Err(err))
	}

}
