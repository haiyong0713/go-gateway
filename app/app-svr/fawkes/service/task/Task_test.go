package task

import (
	"time"

	"go-gateway/app/app-svr/fawkes/service/conf"
)

var S *Service

func init() {
	err := conf.Init()
	if err != nil {
		panic(err)
	}
	S = New(conf.Conf)
	time.Sleep(time.Second)
}
