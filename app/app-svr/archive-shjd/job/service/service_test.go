package service

import (
	"context"
	"flag"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go-gateway/app/app-svr/archive-shjd/job/conf"
)

var (
	s *Service
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.archive-job-shjd")
		flag.Set("conf_token", "ae3c651ac38199f41bb2822cd1adce7e")
		flag.Set("tree_id", "57290")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "uat-config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	//dir, _ := filepath.Abs("../cmd/archive-job-kisjd-test.toml")
	//flag.Set("conf", dir)
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	s = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func Test_Ping(t *testing.T) {
	Convey("Ping", t, func() {
		s.Ping()
	})
}

func Test_DelteVideoCache(t *testing.T) {
	Convey("DelteVideoCache", t, func() {
		s.DelVideoCache(context.Background(), 1, 1)
	})
}

func Test_UpdateVideoCache(t *testing.T) {
	Convey("UpdateVideoCache", t, func() {
		s.UpdateVideoCache(context.Background(), 1, 1)
	})
}

func Test_Close(t *testing.T) {
	Convey("Close", t, func() {
		s.Close()
	})
}
