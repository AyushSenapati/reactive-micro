package dto

import "github.com/google/uuid"

type CreateMerchantRequest struct {
	Name string `json:"name"`
}

type CreateMerchantResponse struct {
	ID  uuid.UUID `json:"id,omitempty"`
	Err error     `json:"error,omitempty"`
}

func (resp CreateMerchantResponse) Failed() error {
	return resp.Err
}

type MerchantDetailResponse struct {
	ID   uuid.UUID `json:"id,omitempty"`
	Name string    `json:"name,omitempty"`
	Err  error     `json:"err,omitempty"`
}

type ListMerchantResponse struct {
	Merchants []MerchantDetailResponse `json:"merchants,omitempty"`
	Err       error                    `json:"err,omitempty"`
}

func (resp ListMerchantResponse) Failed() error {
	return resp.Err
}
