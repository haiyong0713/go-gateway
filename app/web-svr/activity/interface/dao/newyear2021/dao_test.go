package newyear2021

import (
	"flag"

	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

var testDao *Dao

func init() {
	flag.Set("conf", "../../cmd/activity-test.toml")
	if err := conf.Init(); err != nil {
		panic(err)
	}
	if err := component.InitByCfg(conf.Conf.MySQL.Like, conf.Conf.Redis.Config); err != nil {
		panic(err)
	}
	client.New(conf.Conf)
	testDao = New(conf.Conf)

}
