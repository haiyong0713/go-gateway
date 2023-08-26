package component

import (
	"context"

	"go-common/library/database/sql"
	"go-gateway/app/app-svr/app-show/interface/conf"
)

var (
	GlobalShowDB *sql.DB
)

func InitByCfg(cfg *conf.Config) error {
	GlobalShowDB = sql.NewMySQL(cfg.MySQL.Show)
	return Ping()
}

func Ping() error {
	if GlobalShowDB != nil {
		return GlobalShowDB.Ping(context.Background())
	}
	return nil
}

func Close() error {
	if GlobalShowDB != nil {
		return GlobalShowDB.Close()
	}
	return nil
}
