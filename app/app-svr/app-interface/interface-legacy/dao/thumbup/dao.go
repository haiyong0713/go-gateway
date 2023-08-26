package thumbup

import (
	"context"

	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"

	"github.com/pkg/errors"
)

// Dao is tag dao
type Dao struct {
	thumbupGRPC thumbupgrpc.ThumbupClient
}

// New initial tag dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.thumbupGRPC, err = thumbupgrpc.NewClient(c.ThumbupClient); err != nil {
		panic(err)
	}
	return
}

// UserTotalLike user likes list
func (d *Dao) UserTotalLike(c context.Context, mid int64, business string, pn, ps int) (res []*thumbupgrpc.ItemRecord, count int, err error) {
	var (
		reply *thumbupgrpc.UserLikesReply
		ip    = metadata.String(c, metadata.RemoteIP)
	)
	// arg := &thumbup.ArgUserLikes{Mid: mid, Business: business, Pn: pn, Ps: ps, RealIP: ip}
	arg := &thumbupgrpc.UserLikesReq{
		Business: business,
		Pn:       int64(pn),
		Ps:       int64(ps),
		IP:       ip,
		Mid:      mid,
	}
	if reply, err = d.thumbupGRPC.UserLikes(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if reply != nil {
		res = reply.Items
		count = int(reply.Total)
	}
	return
}

// // UserLikedCounts .
func (d *Dao) UserLikedCounts(c context.Context, mid int64, Businesses []string) (counts map[string]int64, err error) {
	var reply *thumbupgrpc.UserLikedCountsReply
	arg := &thumbupgrpc.UserLikedCountsReq{Mid: mid, Businesses: Businesses}
	if reply, err = d.thumbupGRPC.UserLikedCounts(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if reply != nil {
		counts = reply.LikeCounts
	}
	return
}

func (d *Dao) HasLike(ctx context.Context, buvid string, mid int64, messageIDs []int64) (map[int64]thumbupgrpc.State, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	out := make(map[int64]thumbupgrpc.State)
	if mid > 0 {
		arg := &thumbupgrpc.HasLikeReq{
			Business:   "archive",
			MessageIds: messageIDs,
			Mid:        mid,
			IP:         ip,
		}
		reply, err := d.thumbupGRPC.HasLike(ctx, arg)
		if err != nil {
			return nil, err
		}
		for k, v := range reply.States {
			out[k] = v.State
		}
		return out, nil
	}
	arg := &thumbupgrpc.BuvidHasLikeReq{
		Business:   "archive",
		MessageIds: messageIDs,
		Buvid:      buvid,
		IP:         ip,
	}
	reply, err := d.thumbupGRPC.BuvidHasLike(ctx, arg)
	if err != nil {
		return nil, err
	}
	for k, v := range reply.States {
		out[k] = v.State
	}
	return out, nil
}
