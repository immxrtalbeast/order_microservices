package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Email    string    `gorm:"unique;not null"`
	PassHash []byte    `gorm:"not null"`
	IsAdmin  bool      `gorm:"not null;default:false"`
}

type UserRepository interface {
	SaveUser(ctx context.Context, user *User) (uid uuid.UUID, err error)
	User(ctx context.Context, email string) (User, error)
	IsAdmin(ctx context.Context, uid uuid.UUID) (bool, error)
}
