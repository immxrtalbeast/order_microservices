package main

import (
	"context"
	authgrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/auth"
	inventorygrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/inventory"
	ordergrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/order"
	"immxrtalbeast/order_microservices/api-gateway/internal/config"
	"immxrtalbeast/order_microservices/api-gateway/internal/controller"
	"immxrtalbeast/order_microservices/api-gateway/internal/middleware"
	"immxrtalbeast/order_microservices/api-gateway/internal/tracing"
	"log/slog"
	"os"
	"strings"

	kafka "github.com/immxrtalbeast/order_kafka"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger()

	tracer, err := tracing.InitTracer("api-gateway", cfg.Clients.Jaeger.Address)
	if err != nil {
		panic(err)
	}
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	log.Info(
		"starting api-gateway",
	)
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
	appSecret := os.Getenv("APP_SECRET")
	if appSecret == "" {
		panic("APP_SECRET is required")
	}

	authClient, err := authgrpc.New(
		context.Background(),
		cfg.Clients.Auth.Address,
		cfg.Clients.Auth.Timeout,
		cfg.Clients.Auth.RetriesCount,
	)
	authMiddleware := middleware.AuthMiddleware(appSecret)

	if err != nil {
		log.Error("failed to connect auth service", slog.Any("error", err))
		panic("failed to connect authClient")
	}
	inventoryClient, err := inventorygrpc.New(
		context.Background(),
		cfg.Clients.Inventory.Address,
		cfg.Clients.Inventory.Timeout,
		cfg.Clients.Inventory.RetriesCount,
	)
	if err != nil {
		log.Error("failed to connect inventory service", slog.Any("error", err))
		panic("failed to connect inventory service")
	}

	orderClient, err := ordergrpc.New(
		context.Background(),
		cfg.Clients.Order.Address,
		cfg.Clients.Order.Timeout,
		cfg.Clients.Order.RetriesCount,
	)
	if err != nil {
		log.Error("failed to connect order service", slog.Any("error", err))
		panic("failed to connect order service")
	}
	orderStatusProducer := kafka.NewProducer(
		[]string{os.Getenv("KAFKA_ADDRESS")},
		"saga-replies",
	)
	defer orderStatusProducer.Close()

	userController := controller.NewUserController(authClient, cfg.TokenTTL)
	inventoryController := controller.NewInventoryController(inventoryClient)
	orderController := controller.NewOrderController(orderClient, orderStatusProducer)

	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOriginFunc = func(origin string) bool {
		return strings.HasPrefix(origin, "http://localhost:")
	}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{
		"Authorization",
		"Content-Type",
		"Origin",
		"Accept",
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	router.Use(cors.New(corsConfig))
	router.Use(otelgin.Middleware("api-gateway"))
	api := router.Group("/api/v1")
	{
		api.POST("/register", userController.Register)
		api.POST("/login", userController.Login)
	}
	inventory := api.Group("/inventory")
	inventory.Use(authMiddleware)
	{
		inventory.GET("/goods", inventoryController.ListGoods)
		inventory.POST("/add-good", middleware.AdminOnlyMiddleware(), inventoryController.AddGood)
		inventory.PATCH("/update-good", middleware.AdminOnlyMiddleware(), inventoryController.UpdateGood)
		inventory.DELETE("/:id", middleware.AdminOnlyMiddleware(), inventoryController.DeleteGood)
	}
	order := api.Group("/order")
	order.Use(authMiddleware)
	{
		order.POST("/create-order", orderController.CreateOrder)
		order.GET("/order/:id", orderController.GetOrder)
		order.GET("/list-orders/:id", orderController.ListOrders)
		order.PATCH("/:id/cancel", orderController.CancelOrder)
		order.DELETE("/:id", orderController.DeleteOrder)
	}
	admin := api.Group("/admin")
	admin.Use(authMiddleware, middleware.AdminOnlyMiddleware())
	{
		admin.GET("/orders", orderController.ListAllOrders)
		admin.PATCH("/orders/:id/status", orderController.UpdateOrderStatus)
	}
	if err := router.Run(":8080"); err != nil {
		panic(err)
	}

}

func setupLogger() *slog.Logger {
	var log *slog.Logger

	log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)
	return log
}
