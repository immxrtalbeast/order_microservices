package client

import (
	"context"
	"encoding/json"
	"immxrtalbeast/order_microservices/inventory-service/internal/domain"
	"immxrtalbeast/order_microservices/inventory-service/internal/lib/logger/sl"
	"immxrtalbeast/order_microservices/inventory-service/internal/service/good"
	"log/slog"
	"time"

	mykafka "github.com/immxrtalbeast/order_kafka"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/propagation"
)

func ProcessInventoryEvents(consumer *mykafka.Consumer, goodInteractor *good.GoodInteractor, log *slog.Logger) {
	log.Info("listening kafka")
	propagator := propagation.TraceContext{}
	for {
		readCtx, readCancel := context.WithTimeout(context.Background(), 1*time.Second)

		msg, err := consumer.ReadRawMessage(readCtx)
		readCancel()
		if err != nil {
			if err == context.DeadlineExceeded {
				continue
			}
			time.Sleep(1 * time.Second)
			continue
		}
		log.Info("EventReceived")
		eventType, err := getHeaderValue(msg.Headers, "Event-Type")
		if err != nil {
			log.Error("error while parsing eventType", sl.Err(err), "eventType is", eventType)
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
		case "InventoryReserveItemsCommand":
			var event domain.ReserveProductsEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Error("failed to unmarshal event", "type", eventType, "error", err)
				continue
			}
			log.Info("Products reserve command received", event)

			go func() {
				defer processCancel()
				goodInteractor.ReserveProducts(processCtx, event)
			}()

		default:
			continue
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
