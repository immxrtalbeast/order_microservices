package psql

import (
	"context"

	"immxrtalbeast/order_microservices/inventory-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GoodRepository struct {
	db *gorm.DB
}

func NewGoodRepository(db *gorm.DB) *GoodRepository {
	return &GoodRepository{db: db}
}

func (r *GoodRepository) SaveGood(ctx context.Context, good *domain.Good) error {
	err := r.db.WithContext(ctx).Create(&good).Error
	return err
}

func (r *GoodRepository) DeleteGood(ctx context.Context, goodID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", goodID).Delete(&domain.Good{}).Error
}

func (r *GoodRepository) ListGoods(ctx context.Context) ([]*domain.Good, error) {
	var goods []*domain.Good
	err := r.db.WithContext(ctx).
		Model(&domain.Good{}).
		Scan(&goods).
		Error
	return goods, err
}

func (r *GoodRepository) UpdateGood(ctx context.Context, good *domain.Good) error {
	result := r.db.WithContext(ctx).Model(&domain.Good{}).
		Where("id = ?", good.ID).
		Omit("id").
		Updates(&good)

	return result.Error
}
