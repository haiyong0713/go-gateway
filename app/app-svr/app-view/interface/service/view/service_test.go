package view

import (
	"flag"
	"path/filepath"
	"time"

	"go-gateway/app/app-svr/app-view/interface/conf"
)

var (
	s *Service
)

func init() {
	dir, _ := filepath.Abs("../../cmd/app-view-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	s = New(conf.Conf)
	time.Sleep(time.Second)
}
