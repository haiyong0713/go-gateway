package thumbup

import (
	"context"
	"fmt"

	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-view/interface/conf"

	thumbup "git.bilibili.co/bapis/bapis-go/community/service/thumbup"

	"github.com/pkg/errors"
)

// Dao is tag dao
type Dao struct {
	thumbupGRPC thumbup.ThumbupClient
}

// New initial tag dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.thumbupGRPC, err = thumbup.NewClient(c.ThumbupClient)
	if err != nil {
		panic(fmt.Sprintf("thumbup NewClient error(%v)", err))
	}
	return
}

// Like is like view.
func (d *Dao) Like(c context.Context, mid, upMid int64, business string, messageID int64, typ thumbup.Action, withStat bool, mobiApp, device, platform string) (res *thumbup.LikeReply, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &thumbup.LikeReq{Mid: mid, UpMid: upMid, Business: business, MessageID: messageID, Action: typ, IP: ip, WithStat: withStat, MobiApp: mobiApp, Device: device, Platform: platform}
	if res, err = d.thumbupGRPC.Like(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
	}
	return
}

// HasLike user has like
func (d *Dao) HasLike(c context.Context, mid int64, business, buvid string, aid int64) (likeState thumbup.State, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	if mid > 0 {
		arg := &thumbup.HasLikeReq{Mid: mid, Business: business, MessageIds: []int64{aid}, IP: ip}
		res, err := d.thumbupGRPC.HasLike(c, arg)
		if err != nil || res == nil {
			return thumbup.State_STATE_UNSPECIFIED, err
		}
		if state, ok := res.States[aid]; ok {
			likeState = state.State
		}
	} else {
		arg := &thumbup.BuvidHasLikeReq{Buvid: buvid, Business: business, MessageIds: []int64{aid}, IP: ip}
		res, err := d.thumbupGRPC.BuvidHasLike(c, arg)
		if err != nil || res == nil {
			return thumbup.State_STATE_UNSPECIFIED, err
		}
		if state, ok := res.States[aid]; ok {
			likeState = state.State
		}
	}
	return
}

// LikeNoLogin is
func (d *Dao) LikeNoLogin(c context.Context, upMid int64, business, buvid string, messageID int64, typ thumbup.Action, withStat bool) (err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &thumbup.BuvidLikeReq{Buvid: buvid, UpMid: upMid, Business: business, MessageID: messageID, Action: typ, IP: ip}
	if _, err = d.thumbupGRPC.BuvidLike(c, arg); err != nil {
		err = errors.Wrapf(err, "BuvidLike arg(%v)", arg)
	}
	return
}

func (d *Dao) BatchHasLike(c context.Context, mid int64, business, buvid string, aids []int64) (map[int64]thumbup.State, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	res := make(map[int64]thumbup.State)
	if mid > 0 {
		arg := &thumbup.HasLikeReq{Mid: mid, Business: business, MessageIds: aids, IP: ip}
		reply, err := d.thumbupGRPC.HasLike(c, arg)
		if err != nil {
			return nil, err
		}
		for aid, v := range reply.States {
			res[aid] = v.State
		}
	} else {
		arg := &thumbup.BuvidHasLikeReq{Buvid: buvid, Business: business, MessageIds: aids, IP: ip}
		reply, err := d.thumbupGRPC.BuvidHasLike(c, arg)
		if err != nil {
			return nil, err
		}
		for aid, v := range reply.States {
			res[aid] = v.State
		}
	}
	return res, nil
}

// GetStates get like num
func (d *Dao) GetStates(c context.Context, business string, aids []int64) (res *thumbup.StatsReply, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &thumbup.StatsReq{Business: business, MessageIds: aids, IP: ip}
	if res, err = d.thumbupGRPC.Stats(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
	}
	return
}

// 点赞动效查询
func (d *Dao) GetMultiLikeAnimation(ctx context.Context, aid int64) (map[int64]*thumbup.LikeAnimation, error) {
	req := &thumbup.MultiLikeAnimationReq{
		Business:   "archive",
		MessageIds: []int64{aid},
	}
	res, err := d.thumbupGRPC.MultiLikeAnimation(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", req)
	}
	return res.LikeAnimation, nil
}
