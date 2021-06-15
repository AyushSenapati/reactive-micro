package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/model"
)

// InventoryRepository defines all the DB operations that the service supports
type InventoryRepository interface {
	// CreateOrder(ctx context.Context, aid uint, qty int, product_id uuid.UUID, status model.OrderStatus) (uuid.UUID, error)
	// ListOrder(ctx context.Context, qp *dto.BasicQueryParam) ([]model.Order, error)
	// ListOrderByIDs(ctx context.Context, oids []uuid.UUID, qp *dto.BasicQueryParam) ([]model.Order, error)
	CreateMerchant(ctx context.Context, merchantName string, adminID uint) (uuid.UUID, error)
	ListMerchant(ctx context.Context) (merchants []model.Merchant, err error)
	ListMerchantByIDs(ctx context.Context, mids []uuid.UUID) ([]model.Merchant, error)

	CreateProduct(ctx context.Context, name, desc string, mid uuid.UUID, qty int, price float32) (uuid.UUID, error)
	ListProduct(ctx context.Context, qp *dto.BasicQueryParam) ([]model.Product, error)
	ListProductByIDs(ctx context.Context, pids []uuid.UUID, qp *dto.BasicQueryParam) ([]model.Product, error)
	ReserveProduct(ctx context.Context, oid, pid uuid.UUID, qty int) (payble float32, err error)
	RemoveReservedProduct(ctx context.Context, oid uuid.UUID) error
	UndoReserveProduct(ctx context.Context, oid uuid.UUID) error
}

type basicInventoryRepo struct {
	db *gorm.DB
}

func NewBasicOrderRepo(db *gorm.DB) InventoryRepository {
	if db == nil {
		return nil
	}

	// auto-migrate tables
	db.AutoMigrate(&model.Merchant{}, &model.Product{}, &model.ReservedProduct{})

	return &basicInventoryRepo{
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

func (b *basicInventoryRepo) CreateMerchant(ctx context.Context, name string, adminID uint) (uuid.UUID, error) {
	mid := uuid.New()
	mo := model.Merchant{ID: mid, AdminID: adminID, Name: name}
	err := b.db.Create(&mo).Error
	return mo.ID, err
}

func (b *basicInventoryRepo) ListMerchant(ctx context.Context) (merchants []model.Merchant, err error) {
	err = b.db.Debug().Find(&merchants).Error
	return
}

func (b *basicInventoryRepo) ListMerchantByIDs(ctx context.Context, mids []uuid.UUID) (merchants []model.Merchant, err error) {
	err = b.db.Debug().Find(&merchants, mids).Error
	return
}

func (b *basicInventoryRepo) CreateProduct(ctx context.Context, name, desc string, mid uuid.UUID, qty int, price float32) (uuid.UUID, error) {
	pid := uuid.New()
	po := model.Product{ID: pid, Name: name, MerchantID: mid, Qty: qty, Price: price, Desc: desc}
	err := b.db.Create(&po).Error
	return po.ID, err
}

func (b *basicInventoryRepo) ListProduct(ctx context.Context, qp *dto.BasicQueryParam) (products []model.Product, err error) {
	if qp != nil {
		err = b.db.Debug().Scopes(
			orderBy(qp.Filter.OrederBy),
			Paginate(qp.Paginator.Page, qp.Paginator.PageSize),
		).Find(&products).Error
	} else {
		err = b.db.Find(&products).Error
	}
	return
}

func (b *basicInventoryRepo) ListProductByIDs(ctx context.Context, pids []uuid.UUID, qp *dto.BasicQueryParam) (products []model.Product, err error) {
	// err = b.db.Debug().Find(&products).Error
	values := []string{}
	for _, pid := range pids {
		values = append(values, fmt.Sprintf("('%s')", pid.String()))
	}
	q := fmt.Sprintf("select * from products p where p.id = any ( values %s )", strings.Join(values, ","))
	// fmt.Println(b.db.Explain(q))
	err = b.db.Debug().Raw(q, values).Scan(&products).Error
	return
}

func (b *basicInventoryRepo) ReserveProduct(ctx context.Context, oid, pid uuid.UUID, qty int) (float32, error) {
	po := model.Product{ID: pid}

	err := b.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Debug().Find(&po).Error; err != nil {
			return err
		}

		result := tx.Debug().Model(&po).Where("qty > 0").Update("qty", gorm.Expr("qty - ?", qty))
		if result.RowsAffected == 0 {
			return fmt.Errorf("failed updating prodID: %v [err: %v]", pid, result.Error)
		}

		rpo := model.ReservedProduct{OID: oid, PID: pid, Qty: qty}
		if err := tx.Debug().Create(&rpo).Error; err != nil {
			return err
		}

		// returning nil will commit the whole transaction
		return nil
	})

	if err != nil {
		return 0, err
	}
	return float32(qty) * po.Price, nil
}

func (b *basicInventoryRepo) RemoveReservedProduct(ctx context.Context, oid uuid.UUID) error {
	rpo := model.ReservedProduct{OID: oid}
	return b.db.Debug().Delete(&rpo, "o_id = ?", oid).Error
}

func (b *basicInventoryRepo) UndoReserveProduct(ctx context.Context, oid uuid.UUID) error {
	err := b.db.Transaction(func(tx *gorm.DB) error {
		rpo := model.ReservedProduct{OID: oid}
		err := tx.Debug().Find(&rpo).Error
		if err != nil {
			return err
		}
		err = tx.Debug().Delete(&rpo).Error
		if err != nil {
			return err
		}

		po := model.Product{ID: rpo.PID}
		err = tx.Debug().Model(&po).UpdateColumn("qty", gorm.Expr("qty + ?", rpo.Qty)).Error
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
