package main

import (
	"fmt"
	"immxrtalbeast/order_microservices/inventory-service/grpcapp"
	"immxrtalbeast/order_microservices/inventory-service/internal/config"
	"immxrtalbeast/order_microservices/inventory-service/internal/domain"
	"immxrtalbeast/order_microservices/inventory-service/internal/service/good"
	"immxrtalbeast/order_microservices/inventory-service/internal/storage/psql"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger()
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
	goodRepo := psql.NewGoodRepository(db)
	goodInteractor := good.NewGoodInteractor(goodRepo)
	grpcApp := grpcapp.New(log, goodInteractor, cfg.GRPC.Port)
	grpcApp.MustRun()

}

func setupLogger() *slog.Logger {
	var log *slog.Logger

	log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)
	return log
}
