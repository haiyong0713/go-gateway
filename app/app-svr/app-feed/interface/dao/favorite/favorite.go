package favorite

import (
	"context"
	"go-common/library/ecode"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/interface/conf"

	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"github.com/pkg/errors"
)

const (
	_typeVideo = 2
	_typeEp    = 24
)

// Dao is favorite dao
type Dao struct {
	// grpc
	rpcClient favgrpc.FavoriteClient
}

// New initial favorite dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.rpcClient, err = favgrpc.NewClient(c.FavClient); err != nil {
		panic(errors.WithMessage(err, "panic by favgrpc"))
	}
	return
}

// IsFavVideo is favorite
func (d *Dao) IsFavVideos(c context.Context, mid int64, aids []int64) (res map[int64]int8, err error) {
	return d.isFav(c, mid, aids, _typeVideo)
}

func (d *Dao) IsFavEp(ctx context.Context, mid int64, epids []int64) (map[int64]int8, error) {
	return d.isFav(ctx, mid, epids, _typeEp)
}

func (d *Dao) isFav(ctx context.Context, mid int64, oids []int64, typ int32) (map[int64]int8, error) {
	reply, err := d.rpcClient.IsFavoreds(ctx, &favgrpc.IsFavoredsReq{Typ: typ, Mid: mid, Oids: oids})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if reply == nil {
		return nil, ecode.NothingFound
	}
	res := make(map[int64]int8)
	for k, v := range reply.Faveds {
		if v { // 已点赞
			res[k] = 1
		}
	}
	return res, nil
}
