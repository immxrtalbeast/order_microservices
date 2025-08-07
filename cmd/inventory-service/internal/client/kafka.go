package client

import (
	"context"
	"immxrtalbeast/order_microservices/internal/pkg/kafka"
	"immxrtalbeast/order_microservices/inventory-service/internal/domain"
	"immxrtalbeast/order_microservices/inventory-service/internal/service/good"
	"log/slog"
	"time"
)

func ProcessInventoryEvents(consumer *kafka.Consumer, goodInteractor *good.GoodInteractor, log *slog.Logger) {
	for {
		readCtx, readCancel := context.WithTimeout(context.Background(), 5*time.Second)
		var event domain.ReserveProductsEvent

		_, err := consumer.ReadEvent(readCtx, &event)
		readCancel()
		if err != nil {
			if err == context.DeadlineExceeded {
				continue
			}
			time.Sleep(1 * time.Second)
			continue
		}
		processCtx, processCancel := context.WithTimeout(context.Background(), 30*time.Second)
		log.Info("Products reserve command received", event)
		go func() {
			defer processCancel()
			goodInteractor.ReserveProducts(processCtx, event)
		}()
	}
}
