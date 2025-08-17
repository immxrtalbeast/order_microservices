package controller

import (
	ordergrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/order"
	order "immxrtalbeast/order_microservices/protos/gen/go/order"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderController struct {
	orderService *ordergrpc.Client
}

func NewOrderController(orderService *ordergrpc.Client) *OrderController {
	return &OrderController{orderService: orderService}
}

func (c *OrderController) CreateOrder(ctx *gin.Context) {
	type OrderItem struct {
		ProductID string  `json:"product_id" binding:"required"`
		Quantity  int32   `json:"quantity" binding:"required,min=1"`
		Price     float64 `json:"price" binding:"required,min=0.01"`
	}

	type CreateOrderRequest struct {
		Items []OrderItem `json:"items" binding:"required"`
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	items := make([]*order.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &order.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	resp, err := c.orderService.CreateOrder(ctx, userID.(string), items)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"order_id": resp.OrderId,
		"status":   resp.Status.String(),
	})
}

func (c *OrderController) GetOrder(ctx *gin.Context) {
	orderID := ctx.Param("id")

	if _, err := uuid.Parse(orderID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "order ID is required"})
		return
	}

	resp, err := c.orderService.GetOrder(ctx, orderID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp.Order)
}

func (c *OrderController) ListOrders(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 32)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 32)

	resp, err := c.orderService.ListOrdersByUser(ctx, userID.(string), int32(limit), int32(offset))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"orders": resp.Orders})
}
