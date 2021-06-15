package model

import (
	"time"

	"github.com/google/uuid"
)

type Merchant struct {
	ID      uuid.UUID `gorm:"primaryKey"`
	Name    string    `gorm:"unique"`
	AdminID uint      `json:"admin_id,omitempty"`
}

type Product struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// name and merchant_id are composite keys
	Name       string    `gorm:"primaryKey"`
	MerchantID uuid.UUID `gorm:"primaryKey"`
	Qty        int
	Price      float32
	Desc       string
}

type ReservedProduct struct {
	OID uuid.UUID `gorm:"primaryKey"`
	PID uuid.UUID
	Qty int
}
