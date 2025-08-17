package main

import (
	"context"
	"fmt"
	"immxrtalbeast/order_microservices/auth-service/internal/app"
	"immxrtalbeast/order_microservices/auth-service/internal/config"
	"immxrtalbeast/order_microservices/auth-service/internal/lib/logger/slogpretty"
	"immxrtalbeast/order_microservices/internal/pkg/tracing"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("starting application")
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
	tracer, err := tracing.InitTracer("auth-service")
	if err != nil {
		panic(err)
	}
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	dsn := fmt.Sprintf("postgresql://postgres.sqgurzgprfcomirlwgqw:%s@aws-0-eu-north-1.pooler.supabase.com:6543/postgres", os.Getenv("DB_PASS"))
	application := app.New(log, cfg.GRPC.Port, dsn, cfg.TokenTTL, os.Getenv("APP_SECRET"))

	application.GRPCServer.MustRun()
	log.Info("db connected")

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
