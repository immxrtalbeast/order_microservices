package main

import (
	"context"
	"log/slog"
	"os"

	"immxrtalbeast/order_microservices/internal/pkg/tracing"
	"immxrtalbeast/order_microservices/saga-service/internal/client"
	"immxrtalbeast/order_microservices/saga-service/internal/config"
	"immxrtalbeast/order_microservices/saga-service/internal/domain"
	"immxrtalbeast/order_microservices/saga-service/internal/lib/logger/slogpretty"
	"immxrtalbeast/order_microservices/saga-service/internal/service/saga"
	"immxrtalbeast/order_microservices/saga-service/storage/psql"

	kafka "github.com/immxrtalbeast/order_kafka"
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
	tracer, err := tracing.InitTracer("saga-service")
	if err != nil {
		panic(err)
	}
	defer func() { _ = tracer.Shutdown(context.Background()) }()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://postgres:postgres@postgres:5432/order_microservices"
	}
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
		"saga-commands",
	)
	defer producer.Close()
	sagaRepo := psql.NewSagaRepository(db)
	sagaInteractor := saga.NewSagaInteractor(log, producer, sagaRepo)

	repliesConsumer := kafka.NewConsumer(
		[]string{os.Getenv("KAFKA_ADDRESS")},
		"saga-replies",
		"saga-service-replies-group",
	)
	defer repliesConsumer.Close()
	client.ProcessSagaEvents(repliesConsumer, sagaInteractor, log)
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
