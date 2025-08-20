package client

import (
	"context"
	"encoding/json"
	"errors"
	mykafka "immxrtalbeast/order_microservices/internal/pkg/kafka"
	"immxrtalbeast/order_microservices/saga-service/internal/domain"
	"immxrtalbeast/order_microservices/saga-service/internal/lib/logger/sl"
	"immxrtalbeast/order_microservices/saga-service/internal/service/saga"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/propagation"
)

func ProcessSagaEvents(consumer *mykafka.Consumer, sagaInteractor *saga.SagaInteractor, log *slog.Logger) {
	log.Info("listening kafka")
	propagator := propagation.TraceContext{}
	for {
		readCtx, readCancel := context.WithTimeout(context.Background(), 1*time.Second)

		msg, err := consumer.ReadRawMessage(readCtx)
		readCancel()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			log.Error("Unknown error", sl.Err(err))
			time.Sleep(1 * time.Second)
			continue
		}
		log.Info("EventReceived")
		eventType, err := getHeaderValue(msg.Headers, "Event-Type")
		if err != nil {
			log.Error("error while parsing eventType", sl.Err(err), "eventType is", eventType)
			continue
		}
		log.Info("EventType is ", eventType)
		if eventType == "" {
			log.Error("missing event type header")
			continue
		}

		baseCtx := context.Background()
		carrier := propagation.MapCarrier{}
		for _, header := range msg.Headers {
			carrier[header.Key] = string(header.Value)
		}

		ctx := propagator.Extract(baseCtx, carrier)
		processCtx, processCancel := context.WithTimeout(ctx, 30*time.Second)
		switch eventType {
		case "OrderCreatedEvent":
			var event domain.OrderCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Error("failed to unmarshal event", "type", eventType, "error", err)
				continue
			}
			log.Info("Order created event received", "event", event)
			go func() {
				defer processCancel()
				sagaInteractor.StartSaga(processCtx, event)
			}()

		case "InventoryReservedEvent":
			var event domain.ProductsReservedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Error("failed to unmarshal event", "type", eventType, "error", err)
				continue
			}
			go func() {
				defer processCancel()
				sagaInteractor.HandleProductsReserved(processCtx, event)
			}()

		case "InventoryReservedEventFailed":
			var event domain.ProductsReservedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Error("failed to unmarshal event", "type", eventType, "error", err)
				continue
			}
			go func() {
				defer processCancel()
				sagaInteractor.HandleProductsReservedError(processCtx, event)
			}()

		case "PaymentProcessedEvent":
			// ... аналогично

		case "ProductsReservationFailedEvent":
		// ... обработка события об ошибке

		default:
			log.Error("unknown event type", "type", eventType)
		}
	}
}

func getHeaderValue(headers []kafka.Header, key string) (string, error) {
	for _, h := range headers {
		if h.Key == key {
			return string(h.Value), nil
		}
	}
	return "", nil
}
