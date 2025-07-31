package auth

import (
	"context"
	"errors"
	"fmt"
	"immxrtalbeast/order_microservices/auth-service/internal/domain"
	"immxrtalbeast/order_microservices/auth-service/internal/lib/jwt"
	"immxrtalbeast/order_microservices/auth-service/internal/lib/logger/sl"
	"immxrtalbeast/order_microservices/auth-service/internal/storage/psql"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log       *slog.Logger
	usrRepo   domain.UserRepository
	tokenTTL  time.Duration
	appSecret string
}

func New(
	log *slog.Logger, usrRepo domain.UserRepository, tokenTTL time.Duration, appSecret string) *Auth {
	return &Auth{
		usrRepo:   usrRepo,
		log:       log,
		tokenTTL:  tokenTTL,
		appSecret: appSecret,
	}
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func (a *Auth) Login(ctx context.Context, email string, password string) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrRepo.User(ctx, email)
	if err != nil {
		if errors.Is(err, psql.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("user logged in successfully")

	token, err := jwt.NewToken(&user, a.tokenTTL, a.appSecret)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (uuid.UUID, error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}
	user := domain.User{
		Email:    email,
		PassHash: passHash,
	}

	id, err := a.usrRepo.SaveUser(ctx, &user)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
// 	const op = "Auth.IsAdmin"

// 	log := a.log.With(
// 		slog.String("op", op),
// 		slog.Int64("user_id", userID),
// 	)

// 	log.Info("checking if user is admin")

// 	isAdmin, err := a.usrRepo.IsAdmin(ctx, userID)
// 	if err != nil {
// 		return false, fmt.Errorf("%s: %w", op, err)
// 	}

// 	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

// 	return isAdmin, nil
// }
