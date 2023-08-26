package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	ng "go-gateway/app/app-svr/app-feed/interface-ng/internal/model"
	model "go-gateway/app/app-svr/app-feed/interface/model/live"

	"github.com/pkg/errors"
)

const (
	_appMRoom = "/xlive/internal/app-interface/v1/index/RoomsForAppIndex"
	_feedList = "/feed/v1/feed/getList"
	_card     = "/room/v1/RoomRecommend/getInfoByCardId"
)

type liveConfig struct {
	Host string
}

type liveDao struct {
	client *bm.Client
	cfg    liveConfig
}

func (d *liveDao) AppMRoom(ctx context.Context, req *ng.AppMRoomReq) (map[int64]*live.Room, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("access_key", req.AccessKey)
	params.Set("app_key", req.Appkey)
	params.Set("actionKey", req.ActionKey)
	params.Set("platform", req.Platform)
	if req.Mid > 0 {
		params.Set("uid", strconv.FormatInt(req.Mid, 10))
	}
	if req.DeviceName != "" {
		params.Set("device_name", req.DeviceName)
	}
	params.Set("device", req.Device)
	params.Set("mobi_app", req.MobiApp)
	params.Set("statistics", req.Statistics)
	params.Set("buvid", req.Buvid)
	params.Set("build", strconv.Itoa(req.Build))
	// 课堂模式和直播对接后，发现此参数直播未做处理！！！，可忽略
	params.Set("teenagers_mode", strconv.Itoa(req.TeenagersMode))
	params.Set("appver", strconv.Itoa(req.Appver))
	params.Set("filtered", strconv.Itoa(req.Filtered))
	params.Set("https_url_req", strconv.Itoa(req.HttpsUrlReq))
	params.Set("network", req.Network)
	params.Set("need_room_filter", strconv.Itoa(req.NeedRoomFilter))
	for _, roomid := range req.RoomIds {
		params.Add("room_ids", strconv.FormatInt(roomid, 10))
	}
	var res struct {
		Code int `json:"code"`
		Data struct {
			List []*live.Room `json:"list"`
		} `json:"data"`
	}
	if err := d.client.Get(ctx, d.cfg.Host+_appMRoom, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.New(d.cfg.Host + _appMRoom + "?" + params.Encode())
	}
	rs := make(map[int64]*live.Room, len(res.Data.List))
	for _, r := range res.Data.List {
		rs[r.RoomID] = r
	}
	return rs, nil
}

func (d *liveDao) FeedList(c context.Context, mid int64, pn, ps int) ([]*model.Feed, int, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	req, err := d.client.NewRequest("GET", d.cfg.Host+_feedList, "", params)
	if err != nil {
		return nil, 0, err
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
		return nil, 0, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.New(d.cfg.Host + _feedList + "?" + params.Encode())
		return nil, 0, err
	}
	return res.Data.Rooms, res.Data.Count, nil
}

func (d *liveDao) Card(ctx context.Context) (map[int64][]*live.Card, error) {
	var res struct {
		Code int                    `json:"code"`
		Data map[int64][]*live.Card `json:"data"`
	}
	if err := d.client.Get(ctx, d.cfg.Host+_card, metadata.String(ctx, metadata.RemoteIP), nil, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrapf(ecode.Int(res.Code), "%s", d.cfg.Host+_card)
	}
	return res.Data, nil
}
