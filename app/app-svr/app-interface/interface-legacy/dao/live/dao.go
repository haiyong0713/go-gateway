package live

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/live"

	livexfans "git.bilibili.co/bapis/bapis-go/live/xfansmedal"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"

	"github.com/pkg/errors"
)

const (
	_live        = "/AppRoom/getRoomInfo"
	_appMRoom    = "/xlive/internal/app-interface/v1/index/RoomsForAppIndex"
	_statusInfo  = "/room/v1/Room/get_status_info_by_uids"
	_visibleInfo = "/rc/v1/Glory/get_visible"
	_usersInfo   = "/user/v3/User/getMultiple"
	_LiveByRID   = "/room/v2/Room/get_by_ids"
	_liveCenter  = "/xlive/internal/app-ucenter/v1/liveCenter/liveSection"
)

// Dao is space dao
type Dao struct {
	client      *httpx.Client
	live        string
	appMRoom    string
	statusInfo  string
	visibleInfo string
	userInfo    string
	liveByRID   string
	liveCenter  string
	// grpc
	rpcClient      livexfans.AnchorClient
	roomRPCClient  livexroom.RoomClient
	roomGateClient livexroomgate.XroomgateClient
	fansUserClient livexfans.UserClient
}

// New initial space dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:      httpx.NewClient(c.HTTPLive),
		live:        c.Host.APILiveCo + _live,
		appMRoom:    c.Host.APILiveCo + _appMRoom,
		statusInfo:  c.Host.APILiveCo + _statusInfo,
		visibleInfo: c.Host.APILiveCo + _visibleInfo,
		userInfo:    c.Host.APILiveCo + _usersInfo,
		liveByRID:   c.Host.APILiveCo + _LiveByRID,
		liveCenter:  c.Host.APILiveCo + _liveCenter,
	}
	var err error
	if d.rpcClient, err = newClient(c.LiveGRPC); err != nil {
		panic(fmt.Sprintf("livexfans newClient error (%+v)", err))
	}
	if d.roomRPCClient, err = newLiveRoomClient(c.LiveGRPC); err != nil {
		panic(fmt.Sprintf("livexroom newLiveRoomClient error (%+v)", err))
	}
	if d.roomGateClient, err = livexroomgate.NewClientXroomgate(c.LiveGRPC); err != nil {
		panic(fmt.Sprintf("livexroomgate NewClientXroomgate error (%+v)", err))
	}
	if d.fansUserClient, err = newFansUserClient(c.LiveGRPC); err != nil {
		panic(fmt.Sprintf("livexfans newFansUserClient error (%+v)", err))
	}
	return
}

// Live is space live data.
func (d *Dao) Live(c context.Context, mid int64, platform string) (live json.RawMessage, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("platform", platform)
	var res struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err = d.client.Get(c, d.live, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.live+"?"+params.Encode())
		return
	}
	live = res.Data
	return
}

// AppMRoom for live
func (d *Dao) AppMRoom(c context.Context, roomids []int64, platform string) (rs map[int64]*live.Room, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("platform", platform)
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

// StatusInfo for live
func (d *Dao) StatusInfo(c context.Context, mids []int64) (status map[int64]*live.Status, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	for _, mid := range mids {
		params.Add("uids[]", strconv.FormatInt(mid, 10))
	}
	params.Set("filter_offline", "1")
	var res struct {
		Code int                    `json:"code"`
		Data map[int64]*live.Status `json:"data"`
	}
	if err = d.client.Get(c, d.statusInfo, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.statusInfo+"?"+params.Encode())
		return
	}
	status = res.Data
	return
}

// Glory for live search
func (d *Dao) Glory(c context.Context, uid int64) (glory []*live.Glory, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(uid, 10))
	var res struct {
		Code int           `json:"code"`
		Data []*live.Glory `json:"data"`
	}
	if err = d.client.Get(c, d.visibleInfo, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.visibleInfo+"?"+params.Encode())
		return
	}
	glory = res.Data
	return
}

// UserInfo for live search
func (d *Dao) UserInfo(c context.Context, uids []int64) (userInfo map[int64]map[string]*live.Exp, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	for _, uid := range uids {
		params.Set("uids[]", strconv.FormatInt(uid, 10))
	}
	params.Set("attributes[]", "exp")
	var res struct {
		Code int                            `json:"code"`
		Data map[int64]map[string]*live.Exp `json:"data"`
	}
	if err = d.client.Get(c, d.userInfo, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.userInfo+"?"+params.Encode())
		return
	}
	userInfo = res.Data
	return
}

// LiveCenter for mine live center
// 直播中心接口文档：https://info.bilibili.co/pages/viewpage.action?pageId=53625998#id-%E7%9B%B4%E6%92%AD%E4%B8%AD%E5%BF%83%E6%89%A9%E5%B1%95%E6%8A%80%E6%9C%AF%E6%96%B9%E6%A1%88-%E8%8D%89%E7%A8%BF-%E6%8E%A5%E5%8F%A3(%E7%BB%99%E4%B8%BB%E7%AB%99%E7%94%A8)
func (d *Dao) LiveCenter(c context.Context, uid int64, build int, platform string) (int64, error) {
	var (
		err    error
		ip     = metadata.String(c, metadata.RemoteIP)
		params = url.Values{}
	)
	params.Set("uid", strconv.FormatInt(uid, 10))
	params.Set("platform", platform)
	params.Set("build", strconv.Itoa(build))
	var res struct {
		Code int `json:"code"`
		Data struct {
			LiveCenter struct {
				Status        int32 `json:"status"`
				FirstLiveTime int64 `json:"first_live_time"`
			} `json:"live_center"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.liveCenter, ip, params, &res); err != nil {
		return 0, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.liveCenter+"?"+params.Encode())
		return 0, err
	}
	liveCenter := res.Data.LiveCenter
	if liveCenter.Status != 1 {
		return 0, nil
	}
	return liveCenter.FirstLiveTime, nil
}
