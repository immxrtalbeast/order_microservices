package jwt

import (
	"immxrtalbeast/order_microservices/auth-service/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewToken(user *domain.User, duration time.Duration, secret string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", nil
	}
	return tokenString, nil
}
