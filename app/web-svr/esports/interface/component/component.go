package component

import (
	"go-gateway/app/web-svr/esports/interface/conf"
)

func InitComponents() (err error) {
	initMemcahced(conf.Conf)
	InitRedis(conf.Conf)
	initDB(conf.Conf)
	return
}
