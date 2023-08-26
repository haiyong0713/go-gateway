package bender

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"

	"go-gateway/app/app-svr/fawkes/service/conf"
)

var task *Task

func init() {
	err := conf.Init()
	if err != nil {
		panic(err)
	}
	dao := fawkes.New(conf.Conf)
	task = NewTask(conf.Conf, dao, "test")
	time.Sleep(time.Second)
}

func TestTask_HandlerFunc(t *testing.T) {
	Convey("TestTask_HandlerFunc", t, func() {
		task.HandlerFunc(context.Background())
	})
}
