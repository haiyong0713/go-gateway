package service

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/archive/job/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	s *Service
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.archive-job")
		flag.Set("conf_token", "MmQwIqWAyIaIu8CKb7MKcNSYlGGhoudN")
		flag.Set("tree_id", "2301")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	if os.Getenv("UT_LOCAL_TEST") != "" {
		flag.Set("conf", "../cmd/archive-job-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	s = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func Test_PopFail(t *testing.T) {
	Convey("PopFail", t, func() {
		s.PopFail(context.TODO(), "")
	})
}

func Test_TranResult(t *testing.T) {
	Convey("tranResult", t, func() {
		_, _, _, err := s.tranResult(context.TODO(), 10098500)
		So(err, ShouldBeNil)
	})
}
