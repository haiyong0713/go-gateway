package component

import (
	"context"

	"go-common/library/database/sql"
)

var (
	GlobalDB *sql.DB
)

func InitByCfg(cfg *sql.Config) error {
	GlobalDB = sql.NewMySQL(cfg)

	return Ping()
}

func Ping() error {
	return GlobalDB.Ping(context.Background())
}
