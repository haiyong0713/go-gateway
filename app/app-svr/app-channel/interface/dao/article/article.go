package article

import (
	"context"

	"go-gateway/app/app-svr/app-channel/interface/conf"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	artclient "git.bilibili.co/bapis/bapis-go/article/service"

	"github.com/pkg/errors"
)

// Dao is archive dao.
type Dao struct {
	// grpc
	artClient artclient.ArticleGRPCClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.artClient, err = artclient.NewClient(c.ArticleGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Articles(c context.Context, aids []int64) (artm map[int64]*article.Meta, err error) {
	arg := &artclient.ArticleMetasReq{Ids: aids}
	res, err := d.artClient.ArticleMetas(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	artm = res.Res
	return
}
