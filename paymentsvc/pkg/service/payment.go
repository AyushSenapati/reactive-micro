package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/dto"
	svcevent "github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/model"
	"github.com/google/uuid"
)

func (svc *basicPaymentService) RechargeWallet(ctx context.Context, aid uint, amount float32) (uuid.UUID, error) {
	if amount <= 0.0 {
		return uuid.Nil, errors.New("some positive amount is required")
	}

	txid, err := svc.repo.ExecuteTX(ctx, aid, amount, true)
	if err != nil {
		return uuid.Nil, err
	}

	eventPublisher := svcevent.NewEventPublisher()
	eventErr := eventPublisher.AddEvent(svcevent.NewEvent(
		ctx, svcevent.EventUpsertPolicy,
		svcevent.EventUpsertPolicyPayload{
			Sub:          fmt.Sprint(aid),
			ResourceType: "transactions",
			ResourceID:   txid.String(),
			Action:       "get",
		},
	))
	svc.cl.LogIfError(ctx, eventErr)

	eventErr = eventPublisher.Publish(svc.nc)
	svc.cl.LogIfError(ctx, eventErr)
	if eventErr == nil {
		svc.cl.Debug(ctx, fmt.Sprintf("published events: %s", svcevent.EventUpsertPolicy))
	}

	return txid, err
}

func (svc *basicPaymentService) ListTransactions(ctx context.Context, txids []uuid.UUID, qp *dto.BasicQueryParam) dto.ListTransactionsResponse {
	var txObjs []model.Transaction
	var err error

	if len(txids) > 0 {
		txObjs, err = svc.repo.ListTxsByIDs(ctx, txids, qp)
	} else {
		txObjs, err = svc.repo.ListTxs(ctx, qp)
	}

	if err != nil {
		svc.cl.Error(ctx, fmt.Sprintf("err getting transactions [%v]", err))
		return dto.ListTransactionsResponse{Err: err}
	}

	var txs []dto.TransactionResponse
	for _, o := range txObjs {
		txs = append(txs, dto.TransactionResponse{
			ID:         o.ID,
			ExecutedAt: o.ExecutedAt,
			Amount:     o.Amount,
			MadeBy:     o.MadeBy,
			IsCredit:   o.IsCredit,
		})
	}

	return dto.ListTransactionsResponse{Transactions: txs}
}
