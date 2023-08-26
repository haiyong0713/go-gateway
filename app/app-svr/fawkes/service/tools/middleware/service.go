package middleware

import (
	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
)

var (
	fkDao *fkdao.Dao
)

func Init() {
	fkDao = fkdao.New(conf.Conf)
}
