package client

import (
	"context"
	"encoding/json"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/domain"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/lib/logger/sl"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/service/order"
	"log/slog"
	"time"

	mykafka "github.com/ozzus/order_kafka"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/propagation"
)

func ProcessOrderEvents(consumer *mykafka.Consumer, orderInteractor *order.OrderInteractor, log *slog.Logger) {
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
		case "InventoryReservedEvent":
			var event domain.ReserveProductsEventReply
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Error("failed to unmarshal event", "type", eventType, "error", err)
				continue
			}
			log.Info("products reserve command received", "event", event)

			go func() {
				defer processCancel()
				orderInteractor.SetTotalSum(processCtx, event)
			}()

		case "OrderStatusUpdateCommand":
			var command domain.OrderStatusUpdateCommand
			if err := json.Unmarshal(msg.Value, &command); err != nil {
				log.Error("failed to unmarshal command", "type", eventType, "error", err)
				processCancel()
				continue
			}
			go func() {
				defer processCancel()
				if err := orderInteractor.UpdateOrderStatus(processCtx, command.OrderID, command.Status); err != nil {
					log.Error("failed to update order status", sl.Err(err), "order_id", command.OrderID, "status", command.Status)
				}
			}()

		default:
			processCancel()
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
