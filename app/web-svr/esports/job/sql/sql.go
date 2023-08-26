package sql

import (
	"context"

	"go-common/library/database/sql"
	"go-gateway/app/web-svr/esports/job/conf"
)

var (
	GlobalDB *sql.DB
)

func InitByCfg() error {
	GlobalDB = sql.NewMySQL(conf.Conf.Mysql)

	return Ping()
}

func Ping() error {
	return GlobalDB.Ping(context.Background())
}
