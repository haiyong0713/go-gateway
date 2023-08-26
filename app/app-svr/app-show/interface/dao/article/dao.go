package article

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-show/interface/conf"

	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
)

// Dao is article dao.
type Dao struct {
	c         *conf.Config
	artClient artapi.ArticleGRPCClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	// grpc
	var err error
	if d.artClient, err = artapi.NewClient(c.ArticleGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return
}

// ArticleMetas .
func (d *Dao) ArticleMetas(c context.Context, cvids []int64, from int32) (map[int64]*artmdl.Meta, error) {
	rly, e := d.artClient.ArticleMetas(c, &artapi.ArticleMetasReq{Ids: cvids, From: from})
	if e != nil {
		return nil, e
	}
	if rly == nil {
		return make(map[int64]*artmdl.Meta), nil
	}
	return rly.Res, nil
}
