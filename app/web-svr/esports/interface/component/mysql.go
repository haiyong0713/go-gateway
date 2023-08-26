package component

import (
	"go-common/library/database/sql"

	"go-gateway/app/web-svr/esports/interface/conf"
)

var (
	GlobalDB         *sql.DB
	GlobalDBOfMaster *sql.DB
)

func initDB(cfg *conf.Config) {
	GlobalDB = sql.NewMySQL(cfg.Mysql)
	GlobalDBOfMaster = sql.NewMySQL(cfg.MysqlMaster)
}
