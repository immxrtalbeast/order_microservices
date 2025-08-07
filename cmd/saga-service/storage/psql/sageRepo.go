package psql

import (
	"context"
	"immxrtalbeast/order_microservices/saga-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SagaRepository struct {
	db *gorm.DB
}

func NewSagaRepository(db *gorm.DB) *SagaRepository {
	return &SagaRepository{db: db}
}

func (r *SagaRepository) SaveSaga(ctx context.Context, saga *domain.Saga) (uuid.UUID, error) {
	err := r.db.WithContext(ctx).Create(&saga).Error
	return uuid.MustParse(saga.ID), err
}

func (r *SagaRepository) Saga(ctx context.Context, sagaID uuid.UUID) (*domain.Saga, error) {
	var saga *domain.Saga
	err := r.db.WithContext(ctx).Model(&domain.Saga{}).Where("id = ?", sagaID).Scan(&saga).Error
	return saga, err
}

func (r *SagaRepository) UpdateSaga(ctx context.Context, saga *domain.Saga) error {
	result := r.db.WithContext(ctx).Model(&domain.Saga{}).
		Where("id = ?", saga.ID).
		Omit("id").
		Updates(&saga)

	return result.Error
}
