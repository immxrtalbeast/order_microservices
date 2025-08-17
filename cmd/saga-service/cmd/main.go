package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"immxrtalbeast/order_microservices/internal/pkg/kafka"
	"immxrtalbeast/order_microservices/internal/pkg/tracing"
	"immxrtalbeast/order_microservices/saga-service/internal/client"
	"immxrtalbeast/order_microservices/saga-service/internal/config"
	"immxrtalbeast/order_microservices/saga-service/internal/domain"
	"immxrtalbeast/order_microservices/saga-service/internal/lib/logger/slogpretty"
	"immxrtalbeast/order_microservices/saga-service/internal/service/saga"
	"immxrtalbeast/order_microservices/saga-service/storage/psql"

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
		panic("failed to connect database")
	}
	log.Info("db connected")
	db.AutoMigrate(&domain.Saga{})

	producer := kafka.NewProducer(
		[]string{os.Getenv("KAFKA_ADDRESS")},
		"inventory",
	)
	defer producer.Close()
	sagaRepo := psql.NewSagaRepository(db)
	sagaInteractor := saga.NewSagaInteractor(log, producer, sagaRepo)
	consumer := kafka.NewConsumer(
		[]string{os.Getenv("KAFKA_ADDRESS")},
		"saga",
		"order-service-group",
	)
	defer consumer.Close()
	client.ProcessSagaEvents(consumer, sagaInteractor, log)
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
