package article

import (
	"context"

	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	artmdl "go-gateway/app/app-svr/app-interface/interface-legacy/dao/article/model"
	artrpc "go-gateway/app/app-svr/app-interface/interface-legacy/dao/article/rpc/client"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	artclient "git.bilibili.co/bapis/bapis-go/article/service"

	"github.com/pkg/errors"
)

// Dao is atticle dao
type Dao struct {
	artClient artclient.ArticleGRPCClient
	artRPC    *artrpc.Service
}

// New initial tag dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		artRPC: artrpc.New(c.ArticleRPC),
	}
	var err error
	if d.artClient, err = artclient.NewClient(c.ArticleGRPC); err != nil {
		panic(err)
	}
	return
}

// UpArticles get article data from api.
func (d *Dao) UpArticles(c context.Context, mid int64, pn, ps int) (ams []*article.Meta, count int, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &artclient.UpArtMetasReq{Mid: mid, Pn: int32(pn), Ps: int32(ps), Ip: ip}
	res, err := d.artClient.UpArtMetas(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	ams = res.Articles
	count = int(res.Count)
	return
}

// Favorites get article data from api.
func (d *Dao) Favorites(c context.Context, mid int64, pn, ps int) (favs []*artmdl.Favorite, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &artmdl.ArgFav{Mid: mid, Pn: pn, Ps: ps, RealIP: ip}
	if favs, err = d.artRPC.Favorites(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
	}
	return
}

func (d *Dao) Articles(c context.Context, aids []int64) (arts map[int64]*article.Meta, err error) {
	arg := &artclient.ArticleMetasReq{Ids: aids}
	res, err := d.artClient.ArticleMetas(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	arts = res.Res
	return
}

func (d *Dao) UpLists(c context.Context, mid int64) (lists []*article.List, count int, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &artclient.UpListsReq{Mid: mid, Ip: ip}
	res, err := d.artClient.UpLists(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	lists = res.Res
	count = int(res.Total)
	return
}
