package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/ordersvc/pkg/model"
)

// OrderRepository defines all the DB operations that the service supports
type OrderRepository interface {
	CreateOrder(ctx context.Context, aid uint, qty int, product_id uuid.UUID, status model.OrderStatus) (uuid.UUID, error)
	ListOrder(ctx context.Context, qp *dto.BasicQueryParam) ([]model.Order, error)
	ListOrderByIDs(ctx context.Context, oids []uuid.UUID, qp *dto.BasicQueryParam) ([]model.Order, error)
	GetOrderByID(ctx context.Context, oid uuid.UUID) (model.Order, error)
	UpdateOrderStatus(ctx context.Context, oid uuid.UUID, status model.OrderStatus) error
}

type basicOrderRepo struct {
	db *gorm.DB
}

func NewBasicOrderRepo(db *gorm.DB) OrderRepository {
	if db == nil {
		return nil
	}

	// auto-migrate tables
	db.AutoMigrate(&model.Order{})

	return &basicOrderRepo{
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

func (b *basicOrderRepo) CreateOrder(ctx context.Context, aid uint, qty int, product_id uuid.UUID, status model.OrderStatus) (uuid.UUID, error) {
	orderID := uuid.New()
	orderObj := model.Order{ID: orderID, AccntID: aid, ProductID: product_id, Qty: qty, Status: string(status)}
	err := b.db.Create(&orderObj).Error
	return orderObj.ID, err
}

func (b *basicOrderRepo) ListOrder(ctx context.Context, qp *dto.BasicQueryParam) (orders []model.Order, err error) {
	if qp != nil {
		err = b.db.Scopes(
			orderBy(qp.Filter.OrederBy),
			Paginate(qp.Paginator.Page, qp.Paginator.PageSize),
		).Find(&orders).Error
	} else {
		err = b.db.Find(&orders).Error
	}
	return
}

func (b *basicOrderRepo) ListOrderByIDs(ctx context.Context, oids []uuid.UUID, qp *dto.BasicQueryParam) ([]model.Order, error) {
	var orders []model.Order
	values := []string{}
	for _, oid := range oids {
		values = append(values, fmt.Sprintf("('%s')", oid.String()))
	}
	q := fmt.Sprintf("select * from orders o where o.id = any ( values %s )", strings.Join(values, ","))
	err := b.db.Debug().Raw(q, values).Scan(&orders).Error
	return orders, err
}

func (b *basicOrderRepo) GetOrderByID(ctx context.Context, oid uuid.UUID) (model.Order, error) {
	orderObj := model.Order{ID: oid}
	err := b.db.Find(orderObj).Error
	return orderObj, err
}

func (b *basicOrderRepo) UpdateOrderStatus(ctx context.Context, oid uuid.UUID, status model.OrderStatus) error {
	orderObj := model.Order{ID: oid}
	return b.db.Model(orderObj).UpdateColumn("status", string(status)).Error
}
