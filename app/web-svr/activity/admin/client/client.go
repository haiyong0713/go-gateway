package client

import (
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/interface/api"
)

var (
	ActivityClient api.ActivityClient
)

func InitClients(cfg *conf.Config) (err error) {
	ActivityClient, err = api.NewClient(cfg.ActClient)
	if err != nil {
		log.Error("InitClients: init activity client err:", err)
		panic(err)
	}

	return
}
