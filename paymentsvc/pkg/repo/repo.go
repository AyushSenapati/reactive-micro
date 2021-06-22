package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/paymentsvc/pkg/model"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

// PaymentRepository defines all the DB operations that the service supports
type PaymentRepository interface {
	EnableWallet(ctx context.Context, aid uint, balance float32) error
	ExecuteTX(ctx context.Context, aid uint, amount float32, isCredit bool) (uuid.UUID, error)
	ListTxnsByIDs(ctx context.Context, txids []uuid.UUID, qp *dto.BasicQueryParam) ([]model.Transaction, error)
	ListTxns(ctx context.Context, qp *dto.BasicQueryParam) ([]model.Transaction, error)
}

type basicPaymentRepo struct {
	db *gorm.DB
}

func NewBasicOrderRepo(db *gorm.DB) PaymentRepository {
	if db == nil {
		return nil
	}

	// auto-migrate tables
	db.AutoMigrate(&model.Wallet{}, &model.Transaction{})

	return &basicPaymentRepo{
		db: db,
	}
}

func orderBy(orderBy []string) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		for _, o := range orderBy {
			tx = tx.Order(o)
		}
		return tx
	}
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func (b *basicPaymentRepo) EnableWallet(ctx context.Context, aid uint, balance float32) error {
	wo := model.Wallet{AccntID: aid, Balance: balance}
	return b.db.Create(wo).Error
}

func (b *basicPaymentRepo) ExecuteTX(ctx context.Context, aid uint, amount float32, isCredit bool) (uuid.UUID, error) {
	txid := uuid.New()
	var err error

	err = b.db.Transaction(func(tx *gorm.DB) error {
		wo := model.Wallet{AccntID: aid}
		var result *gorm.DB

		if isCredit {
			result = tx.Model(&wo).Where("accnt_id = ?", aid).Update("balance", gorm.Expr("balance + ?", amount))
		} else {
			result = tx.Model(&wo).Where(
				"accnt_id = ? and balance > ?", aid, amount).Update("balance", gorm.Expr("balance - ?", amount))
		}
		if result.RowsAffected == 0 {
			if result.Error == nil {
				return ErrInsufficientBalance
			}
			return result.Error
		}

		txo := model.Transaction{ID: txid, Amount: amount, MadeBy: aid, IsCredit: isCredit}
		err = tx.Create(&txo).Error
		return err
	})

	return txid, err
}

func (b *basicPaymentRepo) ListTxns(ctx context.Context, qp *dto.BasicQueryParam) (txs []model.Transaction, err error) {
	if qp != nil {
		err = b.db.Scopes(
			orderBy(qp.Filter.OrederBy),
			Paginate(qp.Paginator.Page, qp.Paginator.PageSize),
		).Find(&txs).Error
	} else {
		err = b.db.Find(&txs).Error
	}
	return
}

func (b *basicPaymentRepo) ListTxnsByIDs(ctx context.Context, txnids []uuid.UUID, qp *dto.BasicQueryParam) (txns []model.Transaction, err error) {
	values := []string{}
	for _, txnid := range txnids {
		values = append(values, fmt.Sprintf("('%s')", txnid.String()))
	}
	q := fmt.Sprintf(
		"select * from transactions t where t.id = any ( values %s )",
		strings.Join(values, ","))
	err = b.db.Debug().Raw(q, values).Scan(&txns).Error
	return
}
