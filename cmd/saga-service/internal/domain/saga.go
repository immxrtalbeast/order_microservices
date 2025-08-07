package domain

import (
	"time"
)

type Saga struct {
	ID          string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CurrentStep string `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ErrorReason string
}
