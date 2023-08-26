package favorite

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
)

type Dao struct {
	c          *conf.Config
	grpcClient favgrpc.FavoriteClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.grpcClient, err = favgrpc.NewClient(c.FavGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) UGCSeasonRelations(c context.Context, mid int64) (*favgrpc.BatchFavsReply, error) {
	res, err := d.grpcClient.BatchFavs(c, &favgrpc.BatchFavsReq{Mid: mid, Tp: 21})
	if err != nil {
		log.Error("BatchFavs err %v", err)
		return nil, err
	}
	return res, nil
}
