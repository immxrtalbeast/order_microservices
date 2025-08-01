package app

import (
	grpcapp "immxrtalbeast/order_microservices/auth-service/internal/app/grpc"
	"immxrtalbeast/order_microservices/auth-service/internal/domain"
	"immxrtalbeast/order_microservices/auth-service/internal/services/auth"
	"immxrtalbeast/order_microservices/auth-service/internal/storage/psql"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	dsn string,
	tokenTTL time.Duration,
	appSecret string,
) *App {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&domain.User{})

	usrRepo := psql.NewUserRepository(db)
	authService := auth.New(log, usrRepo, tokenTTL, appSecret)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
