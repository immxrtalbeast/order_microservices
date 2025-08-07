package main

import (
	"context"
	"fmt"
	"immxrtalbeast/order_microservices/internal/pkg/kafka"
	"immxrtalbeast/order_microservices/inventory-service/grpcapp"
	"immxrtalbeast/order_microservices/inventory-service/internal/config"
	"immxrtalbeast/order_microservices/inventory-service/internal/domain"
	"immxrtalbeast/order_microservices/inventory-service/internal/lib/logger/slogpretty"
	"immxrtalbeast/order_microservices/inventory-service/internal/service/good"
	"immxrtalbeast/order_microservices/inventory-service/internal/storage/psql"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("starting application")
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
	dsn := fmt.Sprintf("postgresql://postgres.sqgurzgprfcomirlwgqw:%s@aws-0-eu-north-1.pooler.supabase.com:6543/postgres", os.Getenv("DB_PASS"))
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("failed to connect database")
	}
	log.Info("db connected")
	db.AutoMigrate(&domain.Good{})
	producer := kafka.NewProducer(
		[]string{os.Getenv("KAFKA_ADDRESS")},
		"saga",
	)
	defer producer.Close()

	goodRepo := psql.NewGoodRepository(db)
	goodInteractor := good.NewGoodInteractor(goodRepo, log, producer)

	consumer := kafka.NewConsumer(
		[]string{os.Getenv("KAFKA_ADDRESS")},
		"inventory",
		"order-service-group",
	)
	defer consumer.Close()
	go processInventoryEvents(consumer, goodInteractor, log)
	grpcApp := grpcapp.New(log, goodInteractor, cfg.GRPC.Port)
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

func processInventoryEvents(consumer *kafka.Consumer, goodInteractor *good.GoodInteractor, log *slog.Logger) {
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
		log.Info("Order created event received", event)
		goodInteractor.ReserveProducts(processCtx, event)
		processCancel()
	}
}

// {
//   "order_id": "a34ce156-0353-4312-8a20-a80f5e684096",
//   "saga_id": "a32ce156-0353-4312-8a20-a80f5e684096",
//   "products": [
//     {"product_id": "a30ce156-0353-4312-8a20-a80f5e684096", "quantity": 2}
//     ]
// }
