package domain

import (
	"github.com/google/uuid"
)

type Good struct {
	ID              uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name            string    `gorm:"not null"`
	ImageLink       string
	Description     string
	Price           int `gorm:"not null"`
	QuantityInStock int
}
