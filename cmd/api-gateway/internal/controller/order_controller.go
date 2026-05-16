package controller

import (
	"context"
	"errors"
	ordergrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/order"
	"net/http"
	"strconv"
	"strings"

	order "github.com/immxrtalbeast/order_protos/gen/go/order"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	mykafka "github.com/immxrtalbeast/order_kafka"
)

type OrderController struct {
	orderService *ordergrpc.Client
	producer     *mykafka.Producer
}

type orderStatusUpdateCommand struct {
	OrderID uuid.UUID `json:"order_id"`
	Status  string    `json:"status"`
}

func NewOrderController(orderService *ordergrpc.Client, producer *mykafka.Producer) *OrderController {
	return &OrderController{orderService: orderService, producer: producer}
}

func (c *OrderController) CreateOrder(ctx *gin.Context) {
	type OrderItem struct {
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int32  `json:"quantity" binding:"required,min=1"`
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

func (c *OrderController) DeleteOrder(ctx *gin.Context) {
	orderID := ctx.Param("id")
	if err := c.orderService.DeleteOrder(ctx, uuid.MustParse(orderID)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to delete order",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "good deleted successfully",
	})
}

func (c *OrderController) CancelOrder(ctx *gin.Context) {
	orderID := ctx.Param("id")
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	orderResp, err := c.orderService.GetOrder(ctx, orderID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "order not found", "details": err.Error()})
		return
	}
	if orderResp.Order.GetUserId() != userID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "order belongs to another user"})
		return
	}
	if err := c.publishStatus(ctx, orderID, "CANCELLED"); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel order", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"message": "order cancellation accepted", "status": "CANCELLED"})
}

func (c *OrderController) UpdateOrderStatus(ctx *gin.Context) {
	orderID := ctx.Param("id")
	type request struct {
		Status string `json:"status" binding:"required"`
	}
	var req request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	status, err := normalizeOrderStatus(req.Status)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.publishStatus(ctx, orderID, status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order status", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"message": "order status update accepted", "status": status})
}

func (c *OrderController) publishStatus(ctx context.Context, orderID string, status string) error {
	uid, err := uuid.Parse(orderID)
	if err != nil {
		return err
	}
	if c.producer == nil {
		return errors.New("order status producer is not configured")
	}
	return c.producer.PublishEventWithEventType(ctx, "OrderStatusUpdateCommand", orderStatusUpdateCommand{
		OrderID: uid,
		Status:  status,
	}, "OrderStatusUpdateCommand")
}

func normalizeOrderStatus(status string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "PROCESSING", "PREPARING":
		return "PROCESSING", nil
	case "COMPLETED", "READY":
		return "COMPLETED", nil
	case "CANCELLED", "CANCELED":
		return "CANCELLED", nil
	default:
		return "", errors.New("unsupported order status")
	}
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

func (c *OrderController) ListAllOrders(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "50"), 10, 32)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 32)

	resp, err := c.orderService.ListAllOrders(ctx, int32(limit), int32(offset))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"orders": resp.Orders})
}
