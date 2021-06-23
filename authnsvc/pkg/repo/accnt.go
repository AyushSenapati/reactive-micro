package repo

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/dto"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/model"
)

// UserRepository defines all the DB operations that the service supports
type UserRepository interface {
	CreateUser(ctx context.Context, name, email, hashedPswd string, role model.Role) (uint, error)
	ListUser(ctx context.Context, qp *dto.BasicQueryParam) ([]dto.GetAccountResponse, *dto.Page, error)
	ListAccountsByIDs(ctx context.Context, aids []uint, qp *dto.BasicQueryParam) ([]dto.GetAccountResponse, error)
	GetUserByEmail(ctx context.Context, email string) (model.User, error)
	GetUserByID(ctx context.Context, uid uint) (model.User, error)
	GetRoleByName(ctx context.Context, name string) (model.Role, error)
	UpdateUser(ctx context.Context, uid uint, user map[string]interface{}) error
	DeleteUser(ctx context.Context, uid uint) error
	CreateRole(ctx context.Context, name string) (int8, error)
	ListRole(ctx context.Context, qp *dto.BasicQueryParam) ([]model.Role, error)
	DeleteRole(ctx context.Context, rid int8) error
}

type basicUserRepo struct {
	db *gorm.DB
}

func NewBasicUserRepo(db *gorm.DB) UserRepository {
	if db == nil {
		return nil
	}

	// auto-migrate tables
	db.AutoMigrate(&model.User{}, &model.Role{})

	return &basicUserRepo{
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

func (b *basicUserRepo) CreateUser(ctx context.Context, name, email, hashedPswd string, roleObj model.Role) (uint, error) {
	u := model.User{Name: name, Email: email, Password: hashedPswd, Role: roleObj}
	err := b.db.Create(&u).Error
	return u.ID, err
}

func queryMerger(q ...string) string {
	return strings.Join(q, " ")
}

func orderByQry(q ...string) string {
	return fmt.Sprintf("order by %s", strings.Join(q, ", "))
}

func paginate(page, pageSize int) (string, *dto.Page) {
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

	return fmt.Sprintf("offset %d limit %d", offset, pageSize), &dto.Page{Page: page}
}

func (b *basicUserRepo) ListUser(ctx context.Context, qp *dto.BasicQueryParam) ([]dto.GetAccountResponse, *dto.Page, error) {
	var (
		q, pageQry string
		accnts     []dto.GetAccountResponse
		err        error
		pageInfo   *dto.Page
	)

	selectQry := "select u.id, u.email, u.name, r.name as role, count(u.id) over() as total_records from users u"
	joinQry := "join roles r on r.id = u.role_id"

	// if qp != nil {
	// 	err = b.db.Scopes(
	// 		orderBy(qp.Filter.OrederBy),
	// 		Paginate(qp.Paginator.Page, qp.Paginator.PageSize),
	// 	).Joins("Role").Find(&users).Error
	// } else {
	// 	err = b.db.Joins("Role").Find(&users).Error
	// }
	if qp != nil {
		pageQry, pageInfo = paginate(qp.Paginator.Page, qp.Paginator.PageSize)
		for i, o := range qp.Filter.OrederBy {
			qp.Filter.OrederBy[i] = "u." + o
		}

		q = queryMerger(
			selectQry,
			joinQry,
			orderByQry(qp.Filter.OrederBy...),
			pageQry,
		)
	} else {
		q = queryMerger(selectQry, joinQry)
	}

	err = b.db.Debug().Raw(q).Scan(&accnts).Error

	// if records found is zero because of pagination, try filtering records with out
	// pagination and set total records, so that client could set correct page number
	if len(accnts) <= 0 {
		q = queryMerger(selectQry, joinQry)
		err = b.db.Debug().Raw(q).Scan(&accnts).Error
		if len(accnts) > 0 {
			pageInfo.TotalRecords = accnts[0].TotalRecords
			accnts = []dto.GetAccountResponse{}
		}
	} else {
		records := accnts[0].TotalRecords
		pageInfo.TotalRecords = records
		pageInfo.PageSize = len(accnts)
	}

	return accnts, pageInfo, err
}

func (b *basicUserRepo) ListAccountsByIDs(ctx context.Context, aids []uint, qp *dto.BasicQueryParam) ([]dto.GetAccountResponse, error) {
	var accnts []dto.GetAccountResponse
	values := []string{}
	for _, aid := range aids {
		values = append(values, fmt.Sprintf("(%s)", fmt.Sprint(aid)))
	}
	selectQry := "select u.id, u.email, u.name, r.name as role from users u"
	joinQry := "join roles r on r.id = u.role_id"
	filterQry := fmt.Sprintf("where u.id = any ( values %s )", strings.Join(values, ","))
	q := queryMerger(selectQry, joinQry, filterQry)
	err := b.db.Debug().Raw(q, values).Scan(&accnts).Error
	return accnts, err
}

func (b *basicUserRepo) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	usrObj := model.User{}
	err := b.db.Joins("Role").Where("email = ?", email).First(&usrObj).Error
	return usrObj, err
}

func (b *basicUserRepo) GetUserByID(ctx context.Context, uid uint) (model.User, error) {
	usrObj := model.User{}
	err := b.db.Joins("Role").First(&usrObj, uid).Error
	return usrObj, err
}

func (b *basicUserRepo) GetRoleByName(ctx context.Context, name string) (role model.Role, err error) {
	err = b.db.Where("name = ?", name).First(&role).Error
	return
}

func (b *basicUserRepo) UpdateUser(ctx context.Context, uid uint, user map[string]interface{}) error {
	usrObj := model.User{ID: uid}
	if roleName, found := user["role"]; found {
		roleObj, err := b.GetRoleByName(ctx, roleName.(string))
		if err != nil {
			return err
		}
		delete(user, "role")
		user["role_id"] = roleObj.ID
	}
	err := b.db.Model(&usrObj).Updates(user).Error
	return err
}

func (b *basicUserRepo) DeleteUser(ctx context.Context, uid uint) error {
	return b.db.Delete(&model.User{}, uid).Error
}

func (b *basicUserRepo) CreateRole(ctx context.Context, name string) (rid int8, err error) {
	r := model.Role{Name: name}
	err = b.db.Create(&r).Error
	return r.ID, err
}

func (b *basicUserRepo) ListRole(ctx context.Context, qp *dto.BasicQueryParam) (roles []model.Role, err error) {
	fields := []string{"id", "name"}
	if qp != nil {
		err = b.db.Scopes(
			orderBy(qp.Filter.OrederBy),
			Paginate(qp.Paginator.Page, qp.Paginator.PageSize),
		).Select([]string{"id", "name"}).Find(&roles).Error
	} else {
		err = b.db.Select(fields).Find(&roles).Error
	}
	return
}

func (b *basicUserRepo) DeleteRole(ctx context.Context, rid int8) error {
	return b.db.Delete(&model.Role{}, rid).Error
}
