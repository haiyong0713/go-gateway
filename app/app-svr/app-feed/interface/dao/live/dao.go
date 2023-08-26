package live

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/naming/discovery"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	model "go-gateway/app/app-svr/app-feed/interface/model/live"

	liverankgrpc "git.bilibili.co/bapis/bapis-go/live/rankdb/v1"
	livegrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	"github.com/pkg/errors"
)

const (
	_appMRoom = "/xlive/internal/app-interface/v1/index/RoomsForAppIndex"
	_feedList = "/feed/v1/feed/getList"
	_card     = "/room/v1/RoomRecommend/getInfoByCardId"
)

// Dao is show dao.
type Dao struct {
	// http client
	client     *httpx.Client
	clientAsyn *httpx.Client
	// live
	appMRoom       string
	feedList       string
	card           string
	liveClient     livegrpc.XroomgateClient
	liveRankClient liverankgrpc.HotRankClient
}

// New new a bangumi dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		client:     httpx.NewClient(c.HTTPClient, httpx.SetResolver(resolver.New(nil, discovery.Builder()))),
		clientAsyn: httpx.NewClient(c.HTTPClientAsyn),
		appMRoom:   c.HostDiscovery.Live + _appMRoom,
		feedList:   c.Host.LiveAPI + _feedList,
		card:       c.Host.LiveAPI + _card,
	}
	var err error
	if d.liveClient, err = livegrpc.NewClientXroomgate(c.LiveGRPC); err != nil {
		panic(err)
	}
	if d.liveRankClient, err = liverankgrpc.NewClientHotRank(c.LiveRankGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) AppMRoom(c context.Context, roomids []int64, mid int64, platform, deviceName, accessKey, actionKey, appkey, device, mobiApp, statistics, buvid, network string, build, teenagersMode, appver, filtered, httpsUrlReq, needRoomFiter int) (rs map[int64]*live.Room, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("access_key", accessKey)
	params.Set("app_key", appkey)
	params.Set("actionKey", actionKey)
	params.Set("platform", platform)
	if mid > 0 {
		params.Set("uid", strconv.FormatInt(mid, 10))
	}
	if deviceName != "" {
		params.Set("device_name", deviceName)
	}
	params.Set("device", device)
	params.Set("mobi_app", mobiApp)
	params.Set("statistics", statistics)
	params.Set("buvid", buvid)
	params.Set("build", strconv.Itoa(build))
	// 课堂模式和直播对接后，发现此参数直播未做处理！！！，可忽略
	params.Set("teenagers_mode", strconv.Itoa(teenagersMode))
	params.Set("appver", strconv.Itoa(appver))
	params.Set("filtered", strconv.Itoa(filtered))
	params.Set("https_url_req", strconv.Itoa(httpsUrlReq))
	params.Set("network", network)
	params.Set("ip", ip)
	params.Set("need_room_filter", strconv.Itoa(needRoomFiter))
	for _, roomid := range roomids {
		params.Add("room_ids", strconv.FormatInt(roomid, 10))
	}
	var res struct {
		Code int `json:"code"`
		Data struct {
			List []*live.Room `json:"list"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.appMRoom, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.appMRoom+"?"+params.Encode())
		return
	}
	if list := res.Data.List; len(list) > 0 {
		rs = make(map[int64]*live.Room, len(list))
		for _, r := range list {
			rs[r.RoomID] = r
		}
	}
	return
}

func (d *Dao) FeedList(c context.Context, mid int64, pn, ps int) (fs []*model.Feed, count int, err error) {
	var req *http.Request
	params := url.Values{}
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	if req, err = d.client.NewRequest("GET", d.feedList, "", params); err != nil {
		return
	}
	req.Header.Set("X-BiliLive-UID", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Rooms []*model.Feed `json:"rooms"`
			Count int           `json:"count"`
		} `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.feedList+"?"+params.Encode())
		return
	}
	fs = res.Data.Rooms
	count = res.Data.Count
	return
}

func (d *Dao) Card(c context.Context) (csm map[int64][]*live.Card, err error) {
	var res struct {
		Code int                    `json:"code"`
		Data map[int64][]*live.Card `json:"data"`
	}
	if err = d.clientAsyn.Get(c, d.card, "", nil, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.card)
		return
	}
	csm = res.Data
	return
}

func (d *Dao) LiveRoomInfos(ctx context.Context, req *livegrpc.EntryRoomInfoReq) (map[int64]*livegrpc.EntryRoomInfoResp_EntryList, error) {
	reply, err := d.liveClient.EntryRoomInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply.List, nil
}

func (d *Dao) LiveHotRank(ctx context.Context, ids []int64) (map[int64]*liverankgrpc.IsInHotRankResp_HotRankData, error) {
	reply, err := d.liveRankClient.IsInHotRank(ctx, &liverankgrpc.IsInHotRankReq{
		RoomIds: ids,
	})
	if err != nil {
		return nil, err
	}
	return reply.GetList(), nil
}
