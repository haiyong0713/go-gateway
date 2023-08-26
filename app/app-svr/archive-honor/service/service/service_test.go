package service

import (
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/archive-honor/service/conf"
)

var (
	s *Service
)

func TestMain(m *testing.M) {
	//if os.Getenv("DEPLOY_ENV") != "" {
	//	flag.Set("app_id", "main.app-svr.archive-honor-service")
	//	flag.Set("conf_token", "6a91870821701a2c4e6b49d7fc270af2")
	//	flag.Set("tree_id", "136937")
	//	flag.Set("conf_version", "docker-1")
	//	flag.Set("deploy_env", "uat")
	//	flag.Set("conf_host", "config.bilibili.co")
	//	flag.Set("conf_path", "/tmp")
	//	flag.Set("region", "sh")
	//	flag.Set("zone", "sh001")
	//} else {
	flag.Set("conf", "../cmd/archive-honor-service.toml")
	//}
	flag.Parse()
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	cfg := &conf.Config{}
	if err := paladin.Get("archive-honor-service.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	s = New(cfg)
	m.Run()
	os.Exit(0)
}
