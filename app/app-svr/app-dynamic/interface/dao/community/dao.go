package community

import (
	cmtGrpc "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
)

type Dao struct {
	c       *conf.Config
	cmtGrpc cmtGrpc.ReplyInterfaceClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.cmtGrpc, err = cmtGrpc.NewClient(c.CommunityGRPC); err != nil {
		panic(err)
	}
	return d
}
