package service

import (
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin"
)

var (
	s *Service
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.steins-gate-job")
		flag.Set("conf_token", "3e8e118cb06ae86b60c701d17c517c8a")
		flag.Set("tree_id", "116790")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	if os.Getenv("UT_LOCAL_TEST") != "" {
		flag.Set("conf", "../../configs")
	}
	flag.Parse()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	s = New()
	m.Run()
	os.Exit(0)
}

func TestService_Remove(t *testing.T) {
	s.removeHvarRec()
}
