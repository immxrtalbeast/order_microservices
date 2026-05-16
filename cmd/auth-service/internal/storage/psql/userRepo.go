package psql

import (
	"context"
	"errors"
	"immxrtalbeast/order_microservices/auth-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrAppNotFound = errors.New("app not found")
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
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, domain.ErrUserExists
		}
		return uuid.Nil, result.Error
	}
	return user.ID, nil
}

func (r *UserRepository) User(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, err
}

func (r *UserRepository) IsAdmin(ctx context.Context, uid uuid.UUID) (bool, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Select("is_admin").Where("id = ?", uid).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, domain.ErrUserNotFound
	}
	if err != nil {
		return false, err
	}
	return user.IsAdmin, nil
}
