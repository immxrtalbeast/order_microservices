package main

import (
	"context"
	"fmt"
	"immxrtalbeast/order_microservices/cmd/order-service/grpcapp"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/config"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/domain"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/lib/logger/sl"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/lib/logger/slogpretty"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/service/order"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/storage/psql"
	"immxrtalbeast/order_microservices/internal/pkg/kafka"
	"immxrtalbeast/order_microservices/internal/pkg/tracing"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("starting application")
	if err := godotenv.Load(".env"); err != nil {
		log.Error("failed to load .env file", sl.Err(err))
		os.Exit(1)
	}
	tracer, err := tracing.InitTracer("order-service")
	if err != nil {
		panic(err)
	}
	defer func() { _ = tracer.Shutdown(context.Background()) }()
	dsn := fmt.Sprintf("postgresql://postgres.sqgurzgprfcomirlwgqw:%s@aws-0-eu-north-1.pooler.supabase.com:6543/postgres", os.Getenv("DB_PASS"))
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Error("failed to connect database", sl.Err(err))
		os.Exit(1)
	}
	log.Info("db connected")

	db.AutoMigrate(&domain.Order{}, &domain.OrderItem{})
	producer := kafka.NewProducer(
		[]string{os.Getenv("KAFKA_ADDRESS")},
		"saga-replies",
	)
	defer producer.Close()

	orderRepo := psql.NewOrderRepository(db)
	orderInteractor := order.NewOrderInteractor(orderRepo, log, producer)

	consumer := kafka.NewConsumer(
		[]string{os.Getenv("KAFKA_ADDRESS")},
		"saga-replies",
		"order-service-group",
	)
	defer consumer.Close()

	// go processOrderEvents(consumer, orderInteractor, log)
	grpcApp := grpcapp.New(log, orderInteractor, cfg.GRPC.Port)
	grpcApp.MustRun()

}

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

// func processOrderEvents(consumer *kafka.Consumer, orderInteractor *order.OrderInteractor, log *slog.Logger) {
// 	for  {
// 		readCtx, readCancel := context.WithTimeout(context.Background(), 5*time.Second)
// 		var event domain.ReserveProductsEvent

// 		_, err := consumer.ReadEvent(readCtx, &event)
// 		readCancel()
// 		if err != nil {
// 			if err == context.DeadlineExceeded {
// 				continue
// 			}
// 			time.Sleep(1 * time.Second)
// 			continue
// 		}
// 		processCtx, processCancel := context.WithTimeout(context.Background(), 30*time.Second)
// 		log.Info("Order created event received", event)
// 		orderInteractor.ReserveProducts(processCtx, event)
// 		processCancel()
// 	}
// }
