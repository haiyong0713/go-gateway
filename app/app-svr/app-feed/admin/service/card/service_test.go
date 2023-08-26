package card

import (
	"context"
	"flag"
	"path/filepath"

	"go-gateway/app/app-svr/app-feed/admin/conf"
)

var (
	s *Service
	c = context.Background()
)

func init() {
	dir, _ := filepath.Abs("../../cmd/feed-admin-test.toml")
	//dir, _ := filepath.Abs("/Users/litongyu/configs/feed/feed-admin.toml")
	flag.Set("conf", dir)
	flag.Set("deploy.env", "uat")
	conf.Init()
	New(conf.Conf)
	s = New(conf.Conf)
}
