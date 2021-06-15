package dto

import (
	"time"

	"github.com/google/uuid"
)

type RechargeWalletRequest struct {
	Amount float32 `json:"amount"`
}

type RechargeWalletResponse struct {
	TXID uuid.UUID `json:"transaction_id,omitempty"`
	Err  error     `json:"err,omitempty"`
}

func (resp RechargeWalletResponse) Failed() error {
	return resp.Err
}

type TransactionResponse struct {
	ID         uuid.UUID `json:"txn_id"`
	ExecutedAt time.Time `json:"executed_at"`
	Amount     float32   `json:"amount"`
	MadeBy     uint      `json:"made_by,omitempty"`
	IsCredit   bool      `json:"is_credit"`
}

type ListTransactionsResponse struct {
	Transactions []TransactionResponse `json:"transactions,omitempty"`
	Err          error                 `json:"err,omitempty"`
}

func (resp ListTransactionsResponse) Failed() error {
	return resp.Err
}
