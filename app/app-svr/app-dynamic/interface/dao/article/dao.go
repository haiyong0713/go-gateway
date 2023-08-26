package article

import (
	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
)

type Dao struct {
	c           *conf.Config
	articleGRPC articlegrpc.ArticleGRPCClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.articleGRPC, err = articlegrpc.NewClient(c.ArticleGRPC); err != nil {
		panic(err)
	}
	return
}
