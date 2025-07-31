package psql

import (
	"context"
	"errors"
	"immxrtalbeast/order_microservices/auth-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return uuid.Nil, result.Error
	}
	return user.ID, nil
}

func (r *UserRepository) User(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return user, err
}

// func (r *UserRepository) IsAdmin(ctx context.Context, uid uuid.UUID) (bool, error){

// }
