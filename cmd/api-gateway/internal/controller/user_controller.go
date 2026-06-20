package controller

import (
	authgrpc "immxrtalbeast/order_microservices/api-gateway/internal/clients/auth"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

	isAdmin := false
	if parsedToken, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{}); err == nil {
		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
			if claimValue, ok := claims["is_admin"].(bool); ok {
				isAdmin = claimValue
			}
		}
	}

	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		"jwt",
		token,
		int(c.tokenTTL.Seconds()),
		"/",
		"",
		os.Getenv("COOKIE_SECURE") == "true",
		true,
	)

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "login success",
		"token":    token,
		"is_admin": isAdmin,
	})
}
