package live

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"

	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/model/live"

	playgrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livefeed "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"

	"google.golang.org/grpc"
)

const (
	_live     = "/xlive/internal/app-interface/v2/feed/recommendFeedList"
	_rec      = "/appIndex/recommendList"
	roomAppID = "live.xroom"
)

// Dao is live dao
type Dao struct {
	client         *httpx.Client
	clientAsyn     *httpx.Client
	live           string
	rec            string
	roomRPCClient  livexroom.RoomClient
	liveDyn        livefeed.DynamicClient
	playClient     playgrpc.TopicClient
	roomGateClient roomgategrpc.XroomgateClient
}

// New live dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:     httpx.NewClient(c.HTTPClient),
		clientAsyn: httpx.NewClient(c.HTTPClientAsyn),
		live:       c.Host.ApiLiveCo + _live,
		rec:        c.Host.ApiLiveCo + _rec,
	}
	var err error
	if d.roomRPCClient, err = newLiveRoomClient(c.LiveGRPC); err != nil {
		panic(fmt.Sprintf("livexroom newLiveRoomClient error (%+v)", err))
	}
	if d.liveDyn, err = livefeed.NewClient(c.LiveFeedGRPC); err != nil {
		panic(fmt.Sprintf("livefeed NewClient error (%+v)", err))
	}
	if d.playClient, err = playgrpc.NewClient(c.LiveGRPC); err != nil {
		panic(err)
	}
	if d.roomGateClient, err = roomgategrpc.NewClientXroomgate(c.RoomGateGRPC); err != nil {
		panic(err)
	}
	return
}

func newLiveRoomClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (livexroom.RoomClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+roomAppID)
	if err != nil {
		return nil, err
	}
	return livexroom.NewRoomClient(conn), nil
}

// Live feed
func (d *Dao) Feed(c context.Context, mid int64, ak, ip string, now time.Time) (r *live.Feed, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("access_key", ak)
	var res struct {
		Code int        `json:"code"`
		Data *live.Feed `json:"data"`
	}
	if err = d.client.Get(c, d.live, ip, params, &res); err != nil {
		log.Error("Feed url(%s) error(%v)", d.live+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		log.Error("Feed url(%s) error(%v)", d.live+"?"+params.Encode(), res.Code)
		err = fmt.Errorf("feed send failed")
		return
	}
	r = res.Data
	return
}

// Recommend get live Recommend data.
func (d *Dao) Recommend(now time.Time) (r *live.Recommend, err error) {
	params := url.Values{}
	params.Set("count", "60")
	var res struct {
		Code int             `json:"code"`
		Data *live.Recommend `json:"data"`
	}
	if err = d.clientAsyn.Get(context.TODO(), d.rec, "", params, &res); err != nil { // TODO context arg, service context.TODO
		log.Error("live recommend url(%s) error(%v)", d.rec+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		log.Error("live recommend url(%s) error(%v)", d.rec+"?"+params.Encode(), res.Code)
		err = fmt.Errorf("recommend send failed")
		return
	}
	r = res.Data
	return
}

func (d *Dao) GetMultiple(c context.Context, roomIDs []int64) (info map[int64]*livexroom.Infos, err error) {
	var roomIDsFilter []int64
	for _, roomID := range roomIDs {
		if roomID != 0 {
			roomIDsFilter = append(roomIDsFilter, roomID)
		}
	}
	if len(roomIDsFilter) == 0 {
		return
	}
	var (
		arg  = &livexroom.RoomIDsReq{RoomIds: roomIDsFilter, Attrs: []string{"show", "status", "area"}}
		resp *livexroom.RoomIDsInfosResp
	)
	if resp, err = d.roomRPCClient.GetMultiple(c, arg); err != nil || resp == nil {
		log.Error("GetMultiple d.roomRPCClient.GetMultiple error(%v)", err)
		return
	}
	info = resp.List
	return
}

func (d *Dao) SessionInfoBatch(c context.Context, uidsMap map[int64][]string, playURLReq *roomgategrpc.PlayUrlReq, entryFrom []string) (map[int64]*roomgategrpc.SessionInfos, error) {
	reqUids := make(map[int64]*roomgategrpc.LiveIds)
	for k, liveIDs := range uidsMap {
		reqUids[k] = &roomgategrpc.LiveIds{LiveIds: liveIDs}
	}
	req := &roomgategrpc.SessionInfoBatchReq{
		UidLiveIds: reqUids,
		EntryFrom:  entryFrom,
		Playurl:    playURLReq,
	}
	seRly, e := d.roomGateClient.SessionInfoBatch(c, req)
	if e != nil { //错误降级处理
		return nil, e
	}
	if seRly == nil {
		return make(map[int64]*roomgategrpc.SessionInfos), nil
	}
	return seRly.List, nil
}

// GetCardInfo .
func (d *Dao) GetCardInfo(c context.Context, RoomIds []int64, mid, build int64, platform, device string, isHttps bool) (map[uint64]*livefeed.LiveCardInfo, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	rly, err := d.liveDyn.GetCardInfo(c, &livefeed.GetCardInfoReq{RoomIds: RoomIds, Uid: mid, Build: build, Platform: platform, DeviceName: device, IsHttps: isHttps, Ip: ip})
	if err != nil {
		return nil, err
	}
	if rly != nil {
		return rly.LivePlayInfo, nil
	}
	return nil, nil
}

func (d *Dao) GetListByRoomId(c context.Context, roomIDs []int64, isLive int64) (map[int64]*playgrpc.RoomList, error) {
	rly, err := d.playClient.GetListByRoomId(c, &playgrpc.GetListByRoomIdReq{RoomIds: roomIDs, Filter: &playgrpc.Filter{IsLive: isLive}})
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	res := make(map[int64]*playgrpc.RoomList)
	for _, v := range rly.List {
		if v == nil {
			continue
		}
		res[v.RoomId] = v
	}
	return res, nil
}

func (d *Dao) GetListByActId(c context.Context, actID, tabID, isLive, ps, offset int64) (*playgrpc.GetListByActIdResp, error) {
	return d.playClient.GetListByActId(c, &playgrpc.GetListByActIdReq{ActId: actID, TabId: tabID, PageSize: ps, Offset: offset, Filter: &playgrpc.Filter{IsLive: isLive}})
}

func (d *Dao) EntryRoomInfo(ctx context.Context, roomIds []int64, mid int64) (map[int64]*roomgategrpc.EntryRoomInfoResp_EntryList, error) {
	if len(roomIds) == 0 {
		return map[int64]*roomgategrpc.EntryRoomInfoResp_EntryList{}, nil
	}
	req := &roomgategrpc.EntryRoomInfoReq{
		EntryFrom: []string{"NONE"},
		RoomIds:   roomIds,
		Uid:       mid,
		Uipstr:    metadata.String(ctx, metadata.RemoteIP),
	}
	rly, err := d.roomGateClient.EntryRoomInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return map[int64]*roomgategrpc.EntryRoomInfoResp_EntryList{}, nil
	}
	return rly.List, nil
}
