package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending           = OrderStatus("pending")
	OrderStatusPaymentPending    = OrderStatus("payment_pending")
	OrderStatusPaid              = OrderStatus("paid")
	OrderStatusProductOutOfStock = OrderStatus("product_out_of_stock")
	OrderStatusCancelRequested   = OrderStatus("cancel_requested")
	OrderStatusCanceled          = OrderStatus("canceled")
	OrderStatusFailed            = OrderStatus("failed")
)

type Order struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	AccntID   uint
	ProductID uuid.UUID
	Status    string
	Qty       int
}
