package model

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	AccntID uint    `gorm:"index"`
	Balance float32 `gorm:"balance"`
}

type Transaction struct {
	ID         uuid.UUID `gorm:"primaryKey"`
	ExecutedAt time.Time `gorm:"autoCreateTime"`
	Amount     float32
	MadeBy     uint
	IsCredit   bool
}
