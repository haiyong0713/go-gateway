package relation

import (
	"context"

	relationapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"go-common/library/log"
	"go-gateway/app/web-svr/native-page/interface/conf"
)

type Dao struct {
	relClient relationapi.RelationClient
}

func New(cfg *conf.Config) *Dao {
	relClient, err := relationapi.NewClient(cfg.RelClient)
	if err != nil {
		panic(err)
	}
	return &Dao{relClient: relClient}
}

func (d *Dao) Attentions(c context.Context, mid int64) (*relationapi.FollowingsReply, error) {
	return d.relClient.Attentions(c, &relationapi.MidReq{Mid: mid})
}

// RelationsGRPC fids relations
func (d *Dao) RelationsGRPC(ctx context.Context, mid int64, fids []int64) (map[int64]*relationapi.FollowingReply, error) {
	var (
		arg = &relationapi.RelationsReq{
			Mid: mid,
			Fid: fids,
		}
	)
	followingMapReply, err := d.relClient.Relations(ctx, arg)
	if err != nil {
		log.Error("d.relGRPC.Relations(%v) error(%v)", arg, err)
		return nil, err
	}
	return followingMapReply.GetFollowingMap(), nil
}
