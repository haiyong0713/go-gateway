package live

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	cardlive "go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-channel/interface/conf"
	"go-gateway/app/app-svr/app-channel/interface/model/live"

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
	client *httpx.Client
	// live
	appMRoom string
	feedList string
	card     string
}

// New new a bangumi dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		client:   httpx.NewClient(c.HTTPClient),
		appMRoom: c.Host.LiveAPI + _appMRoom,
		feedList: c.Host.LiveAPI + _feedList,
		card:     c.Host.LiveAPI + _card,
	}
	return
}

func (d *Dao) AppMRoom(c context.Context, roomids []int64, platform string) (rs map[int64]*cardlive.Room, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("platform", platform)
	for _, roomid := range roomids {
		params.Add("room_ids", strconv.FormatInt(roomid, 10))
	}
	var res struct {
		Code int `json:"code"`
		Data struct {
			List []*cardlive.Room `json:"list"`
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
		rs = make(map[int64]*cardlive.Room, len(list))
		for _, r := range list {
			rs[r.RoomID] = r
		}
	}
	return
}

func (d *Dao) FeedList(c context.Context, mid int64, pn, ps int) (fs []*live.Feed, count int, err error) {
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
			Rooms []*live.Feed `json:"rooms"`
			Count int          `json:"count"`
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

func (d *Dao) Card(c context.Context) (csm map[int64][]*cardlive.Card, err error) {
	var res struct {
		Code int                        `json:"code"`
		Data map[int64][]*cardlive.Card `json:"data"`
	}
	if err = d.client.Get(c, d.card, "", nil, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.card)
		return
	}
	csm = res.Data
	return
}
