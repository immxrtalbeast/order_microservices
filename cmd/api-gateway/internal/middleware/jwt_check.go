package middleware

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(appSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(appSecret) == "" {
			c.AbortWithStatusJSON(500, gin.H{"error": "auth is not configured"})
			return
		}

		var tokenString string
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			if !strings.HasPrefix(authHeader, "Bearer ") {
				c.AbortWithStatusJSON(401, gin.H{"error": "Bearer token required"})
				return
			}
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			cookie, err := c.Cookie("jwt")
			if err != nil {
				c.AbortWithStatusJSON(401, gin.H{"error": "JWT required"})
				return
			}
			tokenString = cookie
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(appSecret), nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(401, gin.H{"error": "Token expired"})
				return
			}
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		expiresAt, err := claims.GetExpirationTime()
		if err != nil || expiresAt == nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token claims"})
			return
		}
		if time.Now().After(expiresAt.Time) {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token expired"})
			return
		}

		userID, ok := claims["uid"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid user ID in token"})
			return
		}

		c.Set("userID", userID)
		if isAdmin, ok := claims["is_admin"].(bool); ok {
			c.Set("isAdmin", isAdmin)
		} else {
			c.Set("isAdmin", false)
		}

		c.Next()
	}
}

func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, ok := c.Get("isAdmin")
		if !ok || isAdmin != true {
			c.AbortWithStatusJSON(403, gin.H{"error": "admin access required"})
			return
		}

		c.Next()
	}
}
