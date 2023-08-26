package feature

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/conf"
)

var (
	dao        *Dao
	treeID     = 1
	buildLtID  = 1
	path       = "/x/web-interface/user"
	keyName    = "ogvCard"
	creator    = "lisi"
	creatorUID = 1
	c          = context.Background()
)

func init() {
	dir, _ := filepath.Abs("../../cmd/feed-admin-test.toml")
	_ = flag.Set("conf", dir)
	if err := conf.Init(); err != nil {
		panic(err)
	}
	log.Init(conf.Conf.Log)
	if dao == nil {
		dao = New(conf.Conf)
	}
	time.Sleep(time.Second)
}

func WithDao(f func(d *Dao)) func() {
	return func() {
		f(dao)
	}
}

func printDetail(detail interface{}) {
	data, _ := json.MarshalIndent(detail, "", "\t")
	fmt.Printf("%+v\n", string(data))
}
