package main

import (
	"context"
	authgrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/auth"
	"immxrtalbeast/order_microservices/api-gateway/internal/config"
	"immxrtalbeast/order_microservices/api-gateway/internal/controller"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger()

	log.Info(
		"starting api-gateway",
	)

	authClient, err := authgrpc.New(
		context.Background(),
		cfg.Clients.Auth.Address,
		cfg.Clients.Auth.Timeout,
		cfg.Clients.Auth.RetriesCount,
	)
	if err != nil {
		panic("failed to connect authClient")
	}

	userController := controller.NewUserController(authClient, cfg.TokenTTL)

	router := gin.Default()
	api := router.Group("/api/v1")
	{
		api.POST("/register", userController.Register)
		api.POST("/login", userController.Login)
	}
	router.Run(":8080")

}

func setupLogger() *slog.Logger {
	var log *slog.Logger

	log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)
	return log
}
