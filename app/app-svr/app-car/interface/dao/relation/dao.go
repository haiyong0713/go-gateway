package relation

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/conf"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

// Dao is rpc dao.
type Dao struct {
	// grpc
	relGRPC relationgrpc.RelationClient
}

// New new a relation dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.relGRPC, err = relationgrpc.NewClient(c.RelationGRPC); err != nil {
		panic(fmt.Sprintf("relationgrpc NewClientt error (%+v)", err))
	}
	return
}

// RelationsGRPC fids relations
func (d *Dao) RelationsGRPC(ctx context.Context, mid int64, fids []int64) (map[int64]*relationgrpc.FollowingReply, error) {
	var (
		arg = &relationgrpc.RelationsReq{
			Mid: mid,
			Fid: fids,
		}
	)
	followingMapReply, err := d.relGRPC.Relations(ctx, arg)
	if err != nil {
		log.Error("d.relGRPC.Relations(%v) error(%v)", arg, err)
		return nil, err
	}
	return followingMapReply.GetFollowingMap(), nil
}

// StatsGRPC fids stats
func (d *Dao) StatsGRPC(ctx context.Context, mids []int64) (map[int64]*relationgrpc.StatReply, error) {
	var (
		arg = &relationgrpc.MidsReq{Mids: mids}
	)
	statsReply, err := d.relGRPC.Stats(ctx, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return statsReply.GetStatReplyMap(), nil
}

func (d *Dao) StatGRPC(ctx context.Context, mid int64) (*relationgrpc.StatReply, error) {
	var (
		arg = &relationgrpc.MidReq{Mid: mid}
	)
	return d.relGRPC.Stat(ctx, arg)
}

func (d *Dao) Followings(c context.Context, mid int64) ([]*relationgrpc.FollowingReply, error) {
	following, err := d.relGRPC.Followings(c, &relationgrpc.MidReq{Mid: mid})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return following.GetFollowingList(), nil
}

func (d *Dao) RelationsInterrelations(ctx context.Context, mid int64, fids []int64) (map[int64]*relationgrpc.InterrelationReply, error) {
	arg := &relationgrpc.RelationsReq{
		Mid: mid,
		Fid: fids,
	}
	reply, err := d.relGRPC.Interrelations(ctx, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetInterrelationMap(), nil
}

func (d *Dao) Relation(ctx context.Context, mid, fid int64) (*relationgrpc.FollowingReply, error) {
	arg := &relationgrpc.RelationReq{
		Mid: mid,
		Fid: fid,
	}
	reply, err := d.relGRPC.Relation(ctx, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply, nil
}
