package service

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/archive/service/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	s *Service
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.archive-service")
		flag.Set("conf_token", "Y2LJhIsHx87nJaOBSxuG5TeZoLdBFlrE")
		flag.Set("tree_id", "2302")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	s = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestAllTypes(t *testing.T) {
	var (
		c = context.TODO()
	)
	convey.Convey("AllTypes", t, func(ctx convey.C) {
		t := s.AllTypes(c)
		ctx.Convey("Then t should not be nil.", func(ctx convey.C) {
			ctx.So(t, convey.ShouldNotBeNil)
		})
	})
}
