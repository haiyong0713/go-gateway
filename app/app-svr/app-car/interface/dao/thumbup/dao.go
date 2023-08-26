package thumbup

import (
	"context"
	"fmt"

	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-car/interface/conf"
	thumbup2 "go-gateway/app/app-svr/app-car/interface/model/thumbup"

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

func (d *Dao) LikeWithNotLogin(c context.Context, req *thumbup2.LikeReq) error {
	ip := metadata.String(c, metadata.RemoteIP)
	if req.Mid > 0 {
		arg := &thumbup.LikeReq{
			Mid:       req.Mid,
			UpMid:     req.UpMid,
			Business:  req.Business,
			MessageID: req.MsgId,
			Action:    req.Action,
			IP:        ip,
			WithStat:  req.WithStat,
			MobiApp:   req.MobiApp,
			Device:    req.Device,
			Platform:  req.Platform,
		}
		if _, err := d.thumbupGRPC.Like(c, arg); err != nil {
			return err
		}
	} else if req.Buvid != "" {
		arg := &thumbup.BuvidLikeReq{
			Buvid:     req.Buvid,
			UpMid:     req.UpMid,
			Business:  req.Business,
			MessageID: req.MsgId,
			Action:    req.Action,
			IP:        ip,
			MobiApp:   req.MobiApp,
			Device:    req.Device,
			Platform:  req.Platform,
		}
		if _, err := d.thumbupGRPC.BuvidLike(c, arg); err != nil {
			return err
		}
	}
	return nil
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

func (d *Dao) HasLikeBatch(c context.Context, mid int64, business, buvid string, aids []int64) (map[int64]thumbup.State, error) {
	var (
		ip  = metadata.String(c, metadata.RemoteIP)
		res = make(map[int64]thumbup.State)
	)
	if mid > 0 {
		arg := &thumbup.HasLikeReq{Mid: mid, Business: business, MessageIds: aids, IP: ip}
		reply, err := d.thumbupGRPC.HasLike(c, arg)
		if err != nil || reply == nil || len(reply.States) == 0 {
			return make(map[int64]thumbup.State), err
		}
		for k, v := range reply.States {
			res[k] = v.State
		}
	} else {
		arg := &thumbup.BuvidHasLikeReq{Buvid: buvid, Business: business, MessageIds: aids, IP: ip}
		reply, err := d.thumbupGRPC.BuvidHasLike(c, arg)
		if err != nil || reply == nil || len(reply.States) == 0 {
			return make(map[int64]thumbup.State), err
		}
		for k, v := range reply.States {
			res[k] = v.State
		}
	}
	return res, nil
}
