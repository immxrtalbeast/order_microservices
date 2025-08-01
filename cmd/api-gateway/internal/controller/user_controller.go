package controller

import (
	authgrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/auth"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	authService *authgrpc.Client
	tokenTTL    time.Duration
}

func NewUserController(authService *authgrpc.Client, tokenTTL time.Duration) *UserController {
	return &UserController{authService: authService, tokenTTL: tokenTTL}
}

func (c *UserController) Register(ctx *gin.Context) {
	type RegisterRequest struct {
		Email string `json:"email" binding:"required,min=3,max=50"`
		Pass  string `json:"password" binding:"required,min=5,max=50"`
	}

	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	userID, err := c.authService.Register(ctx, req.Email, req.Pass)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to register",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "user created",
		"userID":  userID,
	})

}

func (c *UserController) Login(ctx *gin.Context) {
	type LoginRequest struct {
		Login string `json:"email" binding:"required,min=3,max=50"`
		Pass  string `json:"password" binding:"required,min=5,max=50"`
	}
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}
	token, err := c.authService.Login(ctx, req.Login, req.Pass)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to login",
			"details": err.Error(),
		})
		return
	}
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		"jwt",                     // Имя куки
		token,                     // Значение токена
		int(c.tokenTTL.Seconds()), // Макс возраст в секундах
		"/",                       // Путь
		"",                        // Домен (пусто для текущего домена)
		false,                     // Secure (использовать true в production для HTTPS)
		false,                     // HttpOnly
	)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "loggin success",
	})
}
