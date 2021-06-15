package dto

import "github.com/google/uuid"

type CreateOrderRequest struct {
	PID uuid.UUID `json:"product_id"`
	Qty int       `json:"qty"`
}

type CreateOrderResponse struct {
	OID uuid.UUID `json:"order_id,omitempty"`
	Err error     `json:"error,omitempty"`
}

func (resp CreateOrderResponse) Failed() error {
	return resp.Err
}

type GetOrderResponse struct {
	OID      string `json:"order_id"`
	Status   string `json:"status"`
	Qty      int    `json:"qty"`
	ProdName string `json:"product_name"`
}

type ListOrderResponse struct {
	Orders []GetOrderResponse `json:"orders,omitempty"`
	Err    error              `json:"error,omitempty"`
}

func (resp ListOrderResponse) Failed() error {
	return resp.Err
}
