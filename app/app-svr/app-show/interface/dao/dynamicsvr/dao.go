package dynamicsvr

import (
	httpx "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-show/interface/conf"
	//暂时关闭解决go-main引用问题
	//dyngrpc "go-main/app/dynamic/service/activity/api"
)

type Dao struct {
	c              *conf.Config
	client         *httpx.Client
	dynamicInfoURL string
	activeUserURL  string
	feedDynamicURL string
	briefDynURL    string
	hasFeedURL     string
	//暂时关闭解决go-main引用问题
	//dynClient      dyngrpc.ActPromoRPCClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:              c,
		client:         httpx.NewClient(c.HTTPDynamic),
		dynamicInfoURL: c.Host.Dynamic + _dynamicInfoURI,
		activeUserURL:  c.Host.Dynamic + _activeUserURI,
		feedDynamicURL: c.Host.Dynamic + _feedDynamicURI,
		briefDynURL:    c.Host.Dynamic + _briefDynURI,
		hasFeedURL:     c.Host.Dynamic + _hasFeedURI,
	}
	//暂时关闭解决go-main引用问题
	//var err error
	//if d.dynClient, err = dyngrpc.NewClient(c.DynGRPC); err != nil {
	//	panic(err)
	//}
	return
}
