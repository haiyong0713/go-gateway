package relation

import (
	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

type Dao struct {
	c *conf.Config
	// grpc
	relGRPC relationgrpc.RelationClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.relGRPC, err = relationgrpc.NewClient(c.RelationGRPC); err != nil {
		panic(err)
	}
	return
}
