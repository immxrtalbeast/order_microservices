package main

import (
	"fmt"
	"immxrtalbeast/order_microservices/auth-service/internal/app"
	"immxrtalbeast/order_microservices/auth-service/internal/config"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger()
	log.Info("starting application")
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
	dsn := fmt.Sprintf("postgresql://postgres.sqgurzgprfcomirlwgqw:%s@aws-0-eu-north-1.pooler.supabase.com:6543/postgres", os.Getenv("DB_PASS"))
	application := app.New(log, cfg.GRPC.Port, dsn, cfg.TokenTTL, os.Getenv("APP_SECRET"))

	application.GRPCServer.MustRun()
	log.Info("db connected")
}

func setupLogger() *slog.Logger {
	var log *slog.Logger

	log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)
	return log
}
