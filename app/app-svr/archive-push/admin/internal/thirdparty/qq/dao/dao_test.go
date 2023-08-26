package dao

import (
	"flag"
	"go-common/library/conf/paladin"
	"go-common/library/log"
)

var (
	testD *Dao
)

func init() {
	flag.Set("conf", "/Users/zhouhaotian/Projects/go-gateway/app/app-svr/archive-push/admin/configs")
	flag.Set("deploy.env", "uat")
	log.Init(nil)
	paladin.Init()
	testD, _, _ = Init()
}
