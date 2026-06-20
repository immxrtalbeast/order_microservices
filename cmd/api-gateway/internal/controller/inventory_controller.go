package controller

import (
	"fmt"
	inventorygrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/inventory"
	"immxrtalbeast/order_microservices/api-gateway/internal/lib"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InventoryController struct {
	inventoryService *inventorygrpc.Client
}

const maxProductImageSize = 5 << 20

func NewInventoryController(inventoryService *inventorygrpc.Client) *InventoryController {
	return &InventoryController{inventoryService: inventoryService}
}

func (c *InventoryController) AddGood(ctx *gin.Context) {

	type AddGoodRequest struct {
		Name            string `form:"name" binding:"required"`
		Category        string `form:"category" binding:"required"`
		Description     string `form:"description"`
		Volume          int    `form:"volume" binding:"required,min=1"`
		Price           int    `form:"price" binding:"required,min=1"`
		QuantityInStock int    `form:"quantity_in_stock" binding:"min=0"`
	}

	var req AddGoodRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}
	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "No file uploaded",
			"details": err.Error()})
		return
	}
	if file.Size <= 0 || file.Size > maxProductImageSize {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image size must be between 1 byte and 5 MB"})
		return
	}
	contentType := file.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !allowedTypes[contentType] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
		return
	}
	uploadedFile, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read uploaded file"})
		return
	}
	defer uploadedFile.Close()

	header := make([]byte, 512)
	n, err := io.ReadFull(uploadedFile, header)
	if err != nil && err != io.ErrUnexpectedEOF {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to inspect uploaded file"})
		return
	}
	detectedType := http.DetectContentType(header[:n])
	if !allowedTypes[detectedType] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file content"})
		return
	}
	contentType = detectedType

	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s", uuid.NewString(), filepath.Base(file.Filename)))
	if err := ctx.SaveUploadedFile(file, tempPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file", "details": err.Error()})
		return
	}
	defer os.Remove(tempPath)
	publicURL, err := lib.UploadToSupabase(tempPath, file.Filename, contentType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to storage", "details": err.Error()})
		return
	}

	if err := c.inventoryService.AddGood(ctx, req.Name, req.Category, req.Description, publicURL, req.Price, req.QuantityInStock, int32(req.Volume)); err != nil {
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
	parsedGoodID, err := uuid.Parse(goodID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid good ID format"})
		return
	}
	if err := c.inventoryService.DeleteGood(ctx, parsedGoodID); err != nil {
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
		Category        string `json:"category" binding:"required"`
		Description     string `json:"description"`
		Price           int    `json:"price" binding:"required,min=1"`
		ImageLink       string `json:"image_link"`
		QuantityInStock int    `json:"quantity_in_stock" binding:"min=0"`
	}
	var req UpdateGoodRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}
	parsedGoodID, err := uuid.Parse(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid good ID format"})
		return
	}
	if err := c.inventoryService.UpdateGood(ctx, parsedGoodID, req.Name, req.Category, req.Description, req.ImageLink, req.Price, req.QuantityInStock); err != nil {
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
