package saga

import (
	"context"
	"immxrtalbeast/order_microservices/internal/pkg/kafka"
	"immxrtalbeast/order_microservices/saga-service/internal/domain"
	"immxrtalbeast/order_microservices/saga-service/internal/lib/logger/sl"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

type SagaInteractor struct {
	log      *slog.Logger
	producer *kafka.Producer
	sagaRepo domain.SagaRepository
}

func NewSagaInteractor(log *slog.Logger, producer *kafka.Producer, sagaRepo domain.SagaRepository) *SagaInteractor {
	return &SagaInteractor{log: log, producer: producer, sagaRepo: sagaRepo}
}

func (si *SagaInteractor) StartSaga(ctx context.Context, event domain.OrderCreatedEvent) error {
	const op = "service.saga.start"
	log := si.log.With(
		slog.String("op", op),
	)
	log.Info("starting saga")
	tracer := otel.Tracer("saga-service")
	ctx, span := tracer.Start(ctx, "SagaService.StartSaga")
	defer span.End()
	saga := &domain.Saga{
		CurrentStep: string(domain.StateOrderCreated),
		UserID:      event.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	sagaID, err := si.sagaRepo.SaveSaga(ctx, saga)
	if err != nil {
		log.Error("failed to save saga", sl.Err(err))
		if err := si.producer.PublishEvent(context.Background(), "StartSagaError", ""); err != nil {
			log.Error("Failed to publish event", sl.Err(err))
		}
		return nil
	}
	log.Info("Saga saved! SagaID: ", sagaID)

	si.ExecuteSaga(ctx, saga, event.OrderID, event.Products)
	return nil
}

func (si *SagaInteractor) ExecuteSaga(ctx context.Context, saga *domain.Saga, orderID uuid.UUID, products []domain.OrderItem) {
	const op = "service.saga.execute"
	log := si.log.With(
		slog.String("op", op),
		slog.String("sagaID", saga.ID),
	)
	log.Info("Executing saga...")
	tracer := otel.Tracer("saga-service")
	ctx, span := tracer.Start(ctx, "SagaService.ExecuteSaga")
	defer span.End()
	command := domain.ReserveItemsCommand{
		SagaID:   uuid.MustParse(saga.ID),
		OrderID:  orderID,
		Products: products,
	}
	if err := si.producer.PublishEventWithEventType(ctx, "InventoryReserveItemsCommand", command, "InventoryReserveItemsCommand"); err != nil {
		log.Error("Failed to publish event", sl.Err(err))
		span.RecordError(err)
		return
	}
	log.Info("Command to reserve items sended!")

}

func (si *SagaInteractor) HandleProductsReserved(ctx context.Context, event domain.ProductsReservedEvent) {
	const op = "service.saga.ProductsReserved"
	log := si.log.With(
		slog.String("op", op),
		slog.String("sagaID", event.SagaID.String()),
	)
	tracer := otel.Tracer("saga-service")
	ctx, span := tracer.Start(ctx, "SagaService.HandleProductsReserved")
	defer span.End()
	saga, err := si.sagaRepo.Saga(ctx, event.SagaID)
	if err != nil {
		log.Error("Failed to get saga", sl.Err(err))
		span.RecordError(err)
		return
	}
	saga.CurrentStep = string(domain.StateInventoryReserved)
	saga.UpdatedAt = time.Now()
	err = si.sagaRepo.UpdateSaga(ctx, saga)
	if err != nil {
		log.Error("Failed to save saga", sl.Err(err))
		span.RecordError(err)
		return
	}
	log.Info("Products reservation handled")

}

func (si *SagaInteractor) HandleProductsReservedError(ctx context.Context, event domain.ProductsReservedEvent) {
	const op = "service.saga.HandleProductsReservedError"
	log := si.log.With(
		slog.String("op", op),
		slog.String("sagaID", event.SagaID.String()),
	)
	tracer := otel.Tracer("saga-service")
	ctx, span := tracer.Start(ctx, "SagaService.HandleProductsReservedError")
	defer span.End()
	saga, err := si.sagaRepo.Saga(ctx, event.SagaID)
	if err != nil {
		log.Error("Failed to get saga", sl.Err(err))
		span.RecordError(err)
		return
	}

	if err = si.sagaRepo.UpdateSaga(ctx, saga); err != nil {
		log.Error("Failed to save saga", sl.Err(err))
		span.RecordError(err)
		return
	}
	command := domain.CancelOrderCommand{
		OrderID: event.OrderID,
		SagaID:  event.SagaID,
	}
	if err := si.producer.PublishEventWithEventType(ctx, "OrderCancel", command, "OrderCancel"); err != nil {
		log.Error("Failed to publish event", sl.Err(err))
		span.RecordError(err)
		return
	}
	log.Info("Command to cancel order sended!")
}

func (si *SagaInteractor) HandleCancelOrderCommand(ctx context.Context, command domain.CancelOrderCommand) {
	const op = "service.saga.HandleCancelOrderCommand"
	log := si.log.With(
		slog.String("op", op),
		slog.String("sagaID", command.SagaID.String()),
		slog.String("order_id", command.OrderID.String()),
	)
	log.Info("handling cancel order command")
	tracer := otel.Tracer("saga-service")
	ctx, span := tracer.Start(ctx, "SagaService.HandleCancelOrderCommand")
	defer span.End()

	saga, err := si.sagaRepo.Saga(ctx, command.SagaID)
	if err != nil {
		log.Error("failed to get saga", sl.Err(err))
		span.RecordError(err)
		return
	}

	saga.CurrentStep = string(domain.StateCompensated)
	saga.UpdatedAt = time.Now()

	if err := si.sagaRepo.UpdateSaga(ctx, saga); err != nil {
		log.Error("failed to update saga", sl.Err(err))
		span.RecordError(err)
		return
	}

	releaseCommand := domain.ReleaseInventoryCommand{
		OrderID: command.OrderID,
		SagaID:  command.SagaID,
	}

	if err := si.producer.PublishEventWithEventType(ctx, "ReleaseInventoryCommand", releaseCommand, "ReleaseInventoryCommand"); err != nil {
		log.Error("failed to publish release command", sl.Err(err))
		span.RecordError(err)
		return
	}

	log.Info("cancel order command handled successfully")
}

func (si *SagaInteractor) HandleCompensateOrderCommand(ctx context.Context, command domain.CompensateOrderCommand) {
	const op = "saga.HandleCompensateOrderCommand"
	log := si.log.With(
		slog.String("op", op),
		slog.String("orderID", command.OrderID.String()),
		slog.String("sagaID", command.SagaID.String()),
	)

	log.Info("handling compensate order command")
	tracer := otel.Tracer("saga-service")
	ctx, span := tracer.Start(ctx, "SagaService.HandleCompensateOrderCommand")
	defer span.End()

	saga, err := si.sagaRepo.Saga(ctx, command.SagaID)
	if err != nil {
		log.Error("failed to get saga", sl.Err(err))
		span.RecordError(err)
		return
	}

	saga.CurrentStep = string(domain.StateCompensated)
	saga.UpdatedAt = time.Now()

	if err := si.sagaRepo.UpdateSaga(ctx, saga); err != nil {
		log.Error("failed to update saga", sl.Err(err))
		span.RecordError(err)
		return
	}

	releaseCommand := domain.ReleaseInventoryCommand{
		OrderID: command.OrderID,
		SagaID:  command.SagaID,
	}

	if err := si.producer.PublishEventWithEventType(ctx, "ReleaseInventoryCommand", releaseCommand, "ReleaseInventoryCommand"); err != nil {
		log.Error("failed to publish release command", sl.Err(err))
		span.RecordError(err)
		return
	}

	log.Info("compensate order command handled successfully")
}
