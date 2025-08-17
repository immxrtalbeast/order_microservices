package client

import (
	"context"
	"immxrtalbeast/order_microservices/internal/pkg/kafka"
	"immxrtalbeast/order_microservices/inventory-service/internal/domain"
	"immxrtalbeast/order_microservices/inventory-service/internal/service/good"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/propagation"
)

func ProcessInventoryEvents(consumer *kafka.Consumer, goodInteractor *good.GoodInteractor, log *slog.Logger) {
	log.Info("listening kafka")
	propagator := propagation.TraceContext{}
	for {
		readCtx, readCancel := context.WithTimeout(context.Background(), 1*time.Second)
		var event domain.ReserveProductsEvent

		msg, err := consumer.ReadEvent(readCtx, &event)
		readCancel()
		if err != nil {
			if err == context.DeadlineExceeded {
				continue
			}
			time.Sleep(1 * time.Second)
			continue
		}
		baseCtx := context.Background()
		carrier := propagation.MapCarrier{}
		for _, header := range msg.Headers {
			carrier[header.Key] = string(header.Value)
		}

		ctx := propagator.Extract(baseCtx, carrier)
		processCtx, processCancel := context.WithTimeout(ctx, 30*time.Second)
		log.Info("Products reserve command received", event)
		go func() {
			defer processCancel()
			goodInteractor.ReserveProducts(processCtx, event)
		}()
	}
}
