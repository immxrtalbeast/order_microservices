package controller

import (
	inventorygrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/inventory"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InventoryController struct {
	inventoryService *inventorygrpc.Client
}

func NewInventoryController(inventoryService *inventorygrpc.Client) *InventoryController {
	return &InventoryController{inventoryService: inventoryService}
}

func (c *InventoryController) AddGood(ctx *gin.Context) {
	type AddGoodRequest struct {
		Name            string `json:"name" binding:"required"`
		Description     string `json:"description"`
		Price           int    `json:"price" binding:"required"`
		ImageLink       string `json:"image_link"`
		QuantityInStock int    `json:"quantity_in_stock"`
	}

	var req AddGoodRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}
	if err := c.inventoryService.AddGood(ctx, req.Name, req.Description, req.ImageLink, req.Price, req.QuantityInStock); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to add good",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "good added successfuly",
	})
}

func (c *InventoryController) DeleteGood(ctx *gin.Context) {
	goodID := ctx.Param("id")
	if err := c.inventoryService.DeleteGood(ctx, uuid.MustParse(goodID)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to delete good",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "good deleted successfully",
	})
}

func (c *InventoryController) ListGoods(ctx *gin.Context) {
	goods, err := c.inventoryService.ListProducts(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to get list of goods",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"goods": goods,
	})
}
func (c *InventoryController) UpdateGood(ctx *gin.Context) {
	type UpdateGoodRequest struct {
		ID              string `json:"id" binding:"required"`
		Name            string `json:"name" binding:"required"`
		Description     string `json:"description"`
		Price           int    `json:"price" binding:"required"`
		ImageLink       string `json:"image_link"`
		QuantityInStock int    `json:"quantity_in_stock"`
	}
	var req UpdateGoodRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}
	if err := c.inventoryService.UpdateGood(ctx, uuid.MustParse(req.ID), req.Name, req.Description, req.ImageLink, req.Price, req.QuantityInStock); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to update good",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "good updated successfully",
	})
}
