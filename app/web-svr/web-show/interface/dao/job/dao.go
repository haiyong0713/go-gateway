package job

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/stat/prom"
	"go-gateway/app/web-svr/web-show/interface/conf"
)

// Dao struct
type Dao struct {
	db *xsql.DB
}

// PromError err
func PromError(name string, format string, args ...interface{}) {
	prom.BusinessErrCount.Incr(name)
	log.Error(format, args...)
}

// New conf
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{db: xsql.NewMySQL(c.MySQL.Operation)}
	return
}

// Ping Dao
func (dao *Dao) Ping(c context.Context) error {
	return dao.db.Ping(c)
}

// Close Dao
func (dao *Dao) Close() {
	dao.db.Close()
}
