package relation

import (
	"context"
	"fmt"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-view/interface/conf"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"

	"github.com/pkg/errors"
)

type Dao struct {
	c *conf.Config
	// grpc
	relGRPC relationgrpc.RelationClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.relGRPC, err = relationgrpc.NewClient(c.RelationGRPC); err != nil {
		panic(fmt.Sprintf("relationgrpc NewClientt error (%+v)", err))
	}
	return
}

// Close close resource.
func (d *Dao) Close() {
}

// Prompt prompt
func (d *Dao) Prompt(c context.Context, mid, vmid int64, btype int8) (prompt bool, err error) {
	arg := &relationgrpc.PromptReq{Mid: mid, Fid: vmid, Btype: int32(btype)}
	reply, err := d.relGRPC.Prompt(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	prompt = reply.GetSuccess()
	return
}

// Stat stat
func (d *Dao) Stat(c context.Context, mid int64) (stat *relationgrpc.StatReply, err error) {
	if stat, err = d.relGRPC.Stat(c, &relationgrpc.MidReq{Mid: mid}); err != nil {
		err = errors.Wrapf(err, "%v", mid)
		return
	}
	return
}

// StatsGRPC fids stats
func (d *Dao) StatsGRPC(ctx context.Context, mids []int64) (res map[int64]*relationgrpc.StatReply, err error) {
	var (
		arg        = &relationgrpc.MidsReq{Mids: mids}
		statsReply *relationgrpc.StatsReply
	)
	if statsReply, err = d.relGRPC.Stats(ctx, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = statsReply.StatReplyMap
	return
}

// Relation .
func (d *Dao) Relation(ctx context.Context, mid, fid int64) (rlyReply *relationgrpc.FollowingReply, err error) {
	var (
		arg = &relationgrpc.RelationReq{Mid: mid, Fid: fid}
	)
	if rlyReply, err = d.relGRPC.Relation(ctx, arg); err != nil {
		log.Error("d.relGRPC.Relation(%v) error(%v)", arg, err)
	}
	return
}

// Interrelations
func (d *Dao) Interrelations(ctx context.Context, mid int64, fids []int64) (res map[int64]*relationgrpc.InterrelationReply, err error) {
	var (
		arg = &relationgrpc.RelationsReq{
			Mid: mid,
			Fid: fids,
		}
		reply *relationgrpc.InterrelationMapReply
	)
	if reply, err = d.relGRPC.Interrelations(ctx, arg); err != nil {
		err = errors.Wrapf(err, "d.relGRPC.Interrelations(%v)", arg)
		return
	}
	res = reply.InterrelationMap
	return
}
