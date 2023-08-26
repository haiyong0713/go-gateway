package dynamic

import (
	"context"
	"fmt"

	"net/url"
	"strconv"
	"sync"

	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model/dynamic"

	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-common/library/sync/errgroup.v2"

	listenerChannelgrpc "git.bilibili.co/bapis/bapis-go/car-channel/interface"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dynmetadatagrpc "git.bilibili.co/bapis/bapis-go/dynamic/common/metadata"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	listenergrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"

	"github.com/pkg/errors"
)

const (
	_sound      = "sound"
	_soundAppId = 29
	_carAppId   = 33
)

type Dao struct {
	c                     *conf.Config
	dynamicClient         dyngrpc.FeedClient
	listenerClient        listenergrpc.ListenerSvrClient
	listenerChannelClient listenerChannelgrpc.CarDvDClient
	httpClient            *bm.Client
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:          c,
		httpClient: bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.dynamicClient, err = dyngrpc.NewClient(nil); err != nil {
		panic(fmt.Sprintf("dynamicC NewClient error (%+v)", err))
	}
	if d.listenerClient, err = listenergrpc.NewClient(nil); err != nil {
		panic(fmt.Sprintf("listener NewClient error (%+v)", err))
	}
	if d.listenerChannelClient, err = listenerChannelgrpc.NewClient(nil); err != nil {
		if env.DeployEnv == env.DeployEnvUat {
			log.Error("listenerChannelgrpc NewClient error (%+v)", err)
		} else {
			panic(fmt.Sprintf("listenerChannelgrpc NewClient error (%+v)", err))
		}
	}
	return d
}

func (d *Dao) DynVideoList(ctx context.Context, uid int64, updateBaseLine, assistBaseLine string, dynType []string,
	attention *dyncommongrpc.AttentionInfo, build int, platform, mobiApp, buvid, device string) (*dynamic.DynVideoListRes, error) {
	req := &dyngrpc.VideoNewReq{
		Uid:            uid,
		UpdateBaseline: updateBaseLine,
		AssistBaseline: assistBaseLine,
		TypeList:       dynType,
		AttentionInfo:  attention,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:    strconv.Itoa(build),
			Platform: platform,
			MobiApp:  mobiApp,
			Buvid:    buvid,
			Device:   device,
			Ip:       metadata.String(ctx, metadata.RemoteIP),
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
	}
	data, err := d.dynamicClient.VideoNew(ctx, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	ret := &dynamic.DynVideoListRes{}
	ret.FromVideoNew(data)
	return ret, nil
}

func (d *Dao) DynVideoHistory(ctx context.Context, uid int64, offset string, page int64, dynType []string, attention *dyncommongrpc.AttentionInfo, build int, platform, mobiApp, buvid, device string) (*dynamic.DynVideoListRes, error) {
	req := &dyngrpc.VideoHistoryReq{
		Uid:           uid,
		Offset:        offset,
		Page:          page,
		TypeList:      dynType,
		AttentionInfo: attention,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:    strconv.Itoa(build),
			Platform: platform,
			MobiApp:  mobiApp,
			Buvid:    buvid,
			Device:   device,
			Ip:       metadata.String(ctx, metadata.RemoteIP),
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
	}
	data, err := d.dynamicClient.VideoHistory(ctx, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	ret := &dynamic.DynVideoListRes{}
	ret.FromVideoHistory(data)
	return ret, nil
}

func (d *Dao) DynVideoPersonal(ctx context.Context, hostUid, uid int64, IsPreload bool, offset, build, platform, mobiApp, buvid, devide, ip, from, footprint string, attention *dyncommongrpc.AttentionInfo, dynType []string) (*dynamic.DynVideoListRes, error) {
	req := &dyngrpc.VideoPersonalReq{
		HostUid:        hostUid,
		IsPreload:      IsPreload,
		Offset:         offset,
		Uid:            uid,
		AttentionUsers: attention,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:    build,
			Platform: platform,
			MobiApp:  mobiApp,
			Buvid:    buvid,
			Device:   devide,
			Ip:       ip,
			From:     from,
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
		Footprint: footprint,
		TypeList:  dynType,
	}
	data, err := d.dynamicClient.VideoPersonal(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ret := &dynamic.DynVideoListRes{}
	ret.FromVideoPersonal(data)
	return ret, nil
}

func (d *Dao) RecommendArchives(ctx context.Context, mid int64, buvid string, build int, mobiApp, platform, device, channel string) ([]int64, error) {
	appid := _carAppId
	if channel == _sound {
		appid = _soundAppId
	}
	arg := &listenergrpc.RecommendArchivesReq{
		Mid: mid,
		Ip:  metadata.String(ctx, metadata.RemoteIP),
		Device: &dynmetadatagrpc.Device{
			AppId:    int32(appid),
			Build:    int32(build),
			Buvid:    buvid,
			MobiApp:  mobiApp,
			Platform: platform,
			Device:   device,
			Channel:  channel,
		},
		Network: &dynmetadatagrpc.Network{},
	}
	reply, err := d.listenerClient.GetRecommendArchives(ctx, arg)
	if err != nil {
		return nil, err
	}
	var aids = make([]int64, 0)
	for _, v := range reply.Archives {
		aids = append(aids, v.Aid)
	}
	return aids, nil
}

func (d *Dao) ReportPlayAction(ctx context.Context, mid int64, buvid string, aid, cid, detail int64) (bool, error) {
	arg := &listenergrpc.ReportPlayActionReq{
		Mid:    mid,
		Buvid:  buvid,
		Aid:    aid,
		Cid:    cid,
		Action: 1,
		Detail: detail,
	}
	reply, err := d.listenerClient.ReportPlayAction(ctx, arg)
	if err != nil {
		return false, err
	}
	return reply.GetSuccess(), nil
}

func (d *Dao) ChannelRecommend(ctx context.Context, pn, ps, build int, buvid string, mid, channelID int64) ([]*listenerChannelgrpc.ChannelRecommendInfo, error) {
	arg := &listenerChannelgrpc.ReqChannelRecommend{
		DeviceInfo: &listenerChannelgrpc.DeviceInfo{
			Build: int32(build),
			Buvid: buvid,
		},
		Mid:       mid,
		ChannelId: channelID,
		PageSize:  int64(ps),
	}
	reply, err := d.listenerChannelClient.ChannelRecommend(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply.GetChannel(), nil
}

func (d *Dao) ChannelFeedRecommends(ctx context.Context, pn, ps, build int, buvid string, mid int64, channelIDs []int64) (map[int64][]int64, error) {
	var (
		mutex = sync.Mutex{}
	)
	aidm := map[int64][]int64{}
	g := errgroup.WithContext(ctx)
	for _, v := range channelIDs {
		channelID := v
		g.Go(func(cc context.Context) error {
			arg := &listenerChannelgrpc.ReqChannelFeedRecommend{
				DeviceInfo: &listenerChannelgrpc.DeviceInfo{
					Build: int32(build),
					Buvid: buvid,
				},
				Mid:       mid,
				ChannelId: channelID,
				PageInfo: &listenerChannelgrpc.PageInfo{
					PageNum:  int64(pn),
					PageSize: int64(ps),
				},
			}
			reply, err := d.listenerChannelClient.ChannelFeedRecommend(cc, arg)
			if err != nil {
				return err
			}
			for _, v := range reply.GetItems() {
				mutex.Lock()
				aidm[channelID] = append(aidm[channelID], v.Id)
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return aidm, nil
}

func (d *Dao) ChannelFeedRecommend(ctx context.Context, pn, ps, build int, buvid string, mid int64, channelID int64) ([]int64, error) {
	arg := &listenerChannelgrpc.ReqChannelFeedRecommend{
		DeviceInfo: &listenerChannelgrpc.DeviceInfo{
			Build: int32(build),
			Buvid: buvid,
		},
		Mid:       mid,
		ChannelId: channelID,
		PageInfo: &listenerChannelgrpc.PageInfo{
			PageNum:  int64(pn),
			PageSize: int64(ps),
		},
	}
	reply, err := d.listenerChannelClient.ChannelFeedRecommend(ctx, arg)
	if err != nil {
		return nil, err
	}
	var aids []int64
	for _, v := range reply.Items {
		aids = append(aids, v.Id)
	}
	return aids, nil
}

func (d *Dao) ChannelRecommends(ctx context.Context, pn, ps, build int, buvid string, mid int64, channelIDs []int64) (map[int64][]*listenerChannelgrpc.ChannelRecommendInfo, error) {
	var (
		mutex = sync.Mutex{}
	)
	reply := map[int64][]*listenerChannelgrpc.ChannelRecommendInfo{}
	g := errgroup.WithContext(ctx)
	for _, v := range channelIDs {
		channelID := v
		g.Go(func(cc context.Context) error {
			crReply, err := d.ChannelRecommend(cc, pn, ps, build, buvid, mid, channelID)
			if err != nil {
				return nil
			}
			for _, v := range crReply {
				mutex.Lock()
				reply[channelID] = append(reply[channelID], v)
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) VdUpList(ctx context.Context, teenager int, uid int64, buvid string) (*dynamic.VdUpListRsp, error) {
	params := url.Values{}
	params.Set("teenagers_mode", strconv.Itoa(teenager))
	params.Set("uid", strconv.FormatInt(uid, 10))
	params.Set("buvid", buvid)
	upList := d.c.Host.VcCo + "/dynamic_svr/v0/dynamic_svr/vd_uplist"
	var ret struct {
		Code int                  `json:"code"`
		Msg  string               `json:"msg"`
		Data *dynamic.VdUpListRsp `json:"data"`
	}
	if err := d.httpClient.Get(ctx, upList, "", params, &ret); err != nil {
		log.Error("VdUpList http.Get(%v, %v, %v) error(%v)", teenager, uid, buvid, err)
		return nil, errors.WithStack(err)
	}
	if ret.Code != 0 {
		log.Errorc(ctx, "VdUpList failed to HTTP GET: %v. params: %v. code: %v. msg: %v", upList, params.Encode(), ret.Code, ret.Msg)
		return nil, errors.Wrapf(ecode.Int(ret.Code), ret.Msg)
	}
	return ret.Data, nil
}
