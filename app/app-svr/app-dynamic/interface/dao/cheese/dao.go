package cheese

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	dynCheeseGrpc "git.bilibili.co/bapis/bapis-go/cheese/service/dynamic"
)

type Dao struct {
	c          *conf.Config
	cheeseGrpc dynCheeseGrpc.DynamicClient
	// http client
	client *bm.Client
	// hosts
	attachCard string
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:          c,
		client:     bm.NewClient(c.HTTPClient),
		attachCard: c.Hosts.ApiCo + "/pugv/internal/dynamic/attach/card",
	}
	var err error
	if d.cheeseGrpc, err = dynCheeseGrpc.NewClient(c.DynCheeseGRPC); err != nil {
		panic(err)
	}
	return d
}
