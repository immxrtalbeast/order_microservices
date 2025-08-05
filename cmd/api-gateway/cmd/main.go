package main

import (
	"context"
	authgrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/auth"
	inventorygrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/inventory"
	"immxrtalbeast/order_microservices/api-gateway/internal/config"
	"immxrtalbeast/order_microservices/api-gateway/internal/controller"
	"immxrtalbeast/order_microservices/api-gateway/internal/middleware"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger()

	log.Info(
		"starting api-gateway",
	)
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
	authClient, err := authgrpc.New(
		context.Background(),
		cfg.Clients.Auth.Address,
		cfg.Clients.Auth.Timeout,
		cfg.Clients.Auth.RetriesCount,
	)
	authMiddleware := middleware.AuthMiddleware(os.Getenv("APP_SECRET"))

	if err != nil {
		panic("failed to connect authClient")
	}
	inventoryClient, err := inventorygrpc.New(
		context.Background(),
		cfg.Clients.Inventory.Address,
		cfg.Clients.Inventory.Timeout,
		cfg.Clients.Inventory.RetriesCount,
	)
	if err != nil {
		panic("failed to connect authClient")
	}

	userController := controller.NewUserController(authClient, cfg.TokenTTL)
	inventoryController := controller.NewInventoryController(inventoryClient)

	router := gin.Default()
	api := router.Group("/api/v1")
	{
		api.POST("/register", userController.Register)
		api.POST("/login", userController.Login)
	}
	inventory := api.Group("/inventory")
	inventory.Use(authMiddleware)
	{
		inventory.POST("/add-good", inventoryController.AddGood)
		inventory.GET("/goods", inventoryController.ListGoods)
		inventory.PATCH("/update-good", inventoryController.UpdateGood)
		inventory.DELETE("/:id", inventoryController.DeleteGood)
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
