package favorite

import (
	"context"

	"go-common/library/log"

	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-gateway/app/app-svr/app-channel/interface/conf"
)

// Dao is rpc dao.
type Dao struct {
	favClient favgrpc.FavoriteClient
	conf      *conf.Config
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf: c,
	}
	var err error
	if d.favClient, err = favgrpc.NewClient(c.FavoriteGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) IsFavoreds(c context.Context, mid int64, oids []int64) (res map[int64]bool, err error) {
	var (
		args  = &favgrpc.IsFavoredsReq{Typ: 2, Mid: mid, Oids: oids}
		isFav *favgrpc.IsFavoredsReply
	)
	if isFav, err = d.favClient.IsFavoreds(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = isFav.GetFaveds()
	return
}

func (d *Dao) IsFavVideos(ctx context.Context, mid int64, aids []int64) (map[int64]int8, error) {
	const (
		_typeVideo = 2
	)
	reply, err := d.favClient.IsFavoreds(ctx, &favgrpc.IsFavoredsReq{
		Typ:  _typeVideo,
		Mid:  mid,
		Oids: aids,
	})
	if err != nil {
		return nil, err
	}
	res := make(map[int64]int8, len(aids))
	for k, v := range reply.Faveds {
		if v {
			res[k] = 1
		}
	}
	return res, nil
}
