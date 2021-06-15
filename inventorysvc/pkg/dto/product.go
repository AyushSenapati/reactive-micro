package dto

import "github.com/google/uuid"

type CreateProductRequest struct {
	MID   uuid.UUID `json:"merchant_id"`
	Name  string    `json:"name"`
	Desc  string    `json:"description"`
	Qty   int       `json:"qty"`
	Price float32   `json:"price"`
}

type CreateProductResponse struct {
	ID  uuid.UUID `json:"id,omitempty"`
	Err error     `json:"err,omitempty"`
}

func (resp CreateProductResponse) Failed() error {
	return resp.Err
}

type ProductDetailsResponse struct {
	ID       uuid.UUID `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	LowStock *bool     `json:"low_stock,omitempty"`
	Price    float32   `json:"price,omitempty"`
	Desc     string    `json:"description,omitempty"`
	Err      error     `json:"error,omitempty"`
}

type ListProductResponse struct {
	Products []ProductDetailsResponse `json:"products,omitempty"`
	Err      error                    `json:"error,omitempty"`
}

func (resp ListProductResponse) Failed() error {
	return resp.Err
}
