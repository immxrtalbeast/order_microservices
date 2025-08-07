package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Saga struct {
	ID          string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CurrentStep string    `gorm:"not null"`
	UserID      uuid.UUID `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ErrorReason string
}

type SagaState string

const (
	StateOrderCreated      SagaState = "ORDER_CREATED"
	StateInventoryReserved SagaState = "INVENTORY_RESERVED"
	StatePaymentProcessing SagaState = "PAYMENT_PROCESSING"
	StateCompleted         SagaState = "COMPLETED"
	StateCompensating      SagaState = "COMPENSATING"
	StateCompensated       SagaState = "COMPENSATED"
)

type SagaInteractor interface {
	StartSaga(ctx context.Context)
}

type SagaRepository interface {
	SaveSaga(ctx context.Context, saga *Saga) (uuid.UUID, error)
	Saga(ctx context.Context, sagaID uuid.UUID) (*Saga, error)
	UpdateSaga(ctx context.Context, saga *Saga) error
}
