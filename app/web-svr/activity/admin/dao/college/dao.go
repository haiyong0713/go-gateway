package college

import (
	"context"
	"go-common/library/database/orm"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/model/college"

	"github.com/jinzhu/gorm"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package college

// Dao dao interface
type Dao interface {
	BatchAddCollege(c context.Context, tx *xsql.Tx, rank []*college.College) (err error)
	BacthInsertOrUpdateCollege(c context.Context, collegeInfo *college.College) (err error)
	BacthInsertOrUpdateAidList(c context.Context, collegeInfo *college.AIDList) (err error)
	BeginTran(c context.Context) (tx *xsql.Tx, err error)
	GetDB() *gorm.DB
	GetAidDB() *gorm.DB
	Close()
}

// BeginTran begin transcation.
func (d *dao) BeginTran(c context.Context) (tx *xsql.Tx, err error) {
	return d.db.Begin(c)
}

// Dao dao.
type dao struct {
	c     *conf.Config
	db    *xsql.DB
	DB    *gorm.DB
	AidDB *gorm.DB
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:     c,
		DB:    orm.NewMySQL(c.ORM),
		AidDB: orm.NewMySQL(c.ORM),
		db:    xsql.NewMySQL(c.MySQL.Lottery),
	}
	return
}

func (d *dao) GetDB() *gorm.DB {
	return d.DB
}

func (d *dao) GetAidDB() *gorm.DB {
	return d.AidDB
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// Close Dao
func (d *dao) Close() {
}
