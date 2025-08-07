package saga

import (
	"context"
	"immxrtalbeast/order_microservices/internal/pkg/kafka"
	"immxrtalbeast/order_microservices/saga-service/internal/domain"
	"immxrtalbeast/order_microservices/saga-service/internal/lib/logger/sl"
	"log/slog"
	"time"

	"github.com/google/uuid"
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
	time.Sleep(15 * time.Second)
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

	go si.ExecuteSaga(saga, event.OrderID, event.Products)
	return nil
}

func (si *SagaInteractor) ExecuteSaga(saga *domain.Saga, orderID uuid.UUID, products []domain.OrderItem) {
	const op = "service.saga.execute"
	log := si.log.With(
		slog.String("op", op),
		slog.String("sagaID", saga.ID),
	)
	log.Info("Executing saga...")
	event := domain.ReserveItemsCommand{
		SagaID:   uuid.MustParse(saga.ID),
		OrderID:  orderID,
		Products: products,
	}
	if err := si.producer.PublishEvent(context.Background(), "ReserveItemsCommand", event); err != nil {
		log.Error("Failed to publish event", sl.Err(err))
		return
	}
	log.Info("Command to reserve items sended!")
	return
}

func (si *SagaInteractor) HandleProductsReserved(ctx context.Context, event domain.ProductsReservedEvent) {
	const op = "service.saga.ProductsReserved"
	log := si.log.With(
		slog.String("op", op),
		slog.String("sagaID", event.SagaID.String()),
	)
	saga, err := si.sagaRepo.Saga(ctx, event.SagaID)
	if err != nil {
		log.Error("Failed to get saga", sl.Err(err))
		return
	}
	saga.CurrentStep = string(domain.StateInventoryReserved)
	saga.UpdatedAt = time.Now()
	err = si.sagaRepo.UpdateSaga(ctx, saga)
	if err != nil {
		log.Error("Failed to save saga", sl.Err(err))
		return
	}
	log.Info("Products reservation handled")
	return
}
