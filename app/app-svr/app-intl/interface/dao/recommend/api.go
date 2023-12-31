package recommend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"

	"github.com/pkg/errors"
)

const (
	_rcmd    = "/pegasus/feed/%d"
	_hot     = "/data/rank/reco-tmzb.json"
	_group   = "/group_changes/pegasus.json"
	_top     = "/feed/tag/top"
	_rcmdHot = "/recommand"
)

// Recommend is.
func (d *Dao) Recommend(c context.Context, plat int8, buvid string, mid int64, build, loginEvent int, zoneID int64, group int, interest, network string, style int, column model.ColumnStatus, flush int, autoplay string, now time.Time) (rs []*ai.Item, userFeature json.RawMessage, respCode int, newUser bool, err error) {
	if mid == 0 && buvid == "" {
		return
	}
	ip := metadata.String(c, metadata.RemoteIP)
	uri := fmt.Sprintf(d.rcmd, group)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("plat", strconv.Itoa(int(plat)))
	params.Set("build", strconv.Itoa(build))
	params.Set("login_event", strconv.Itoa(loginEvent))
	params.Set("zone_id", strconv.FormatInt(zoneID, 10))
	params.Set("interest", interest)
	params.Set("timeout", "500")
	params.Set("network", network)
	if column > -1 {
		params.Set("column", strconv.Itoa(int(column)))
	}
	params.Set("style", strconv.Itoa(style))
	params.Set("flush", strconv.Itoa(flush))
	params.Set("autoplay_card", autoplay)
	var res struct {
		Code        int             `json:"code"`
		NewUser     bool            `json:"new_user"`
		UserFeature json.RawMessage `json:"user_feature"`
		Data        []*ai.Item      `json:"data"`
	}
	if err = d.client.Get(c, uri, ip, params, &res); err != nil {
		respCode = ecode.ServerErr.Code()
		return
	}
	code := ecode.Int(res.Code)
	if !code.Equal(ecode.OK) {
		respCode = res.Code
		err = errors.Wrapf(code, "%s", uri+"?"+params.Encode())
		return
	}
	rs = res.Data
	userFeature = res.UserFeature
	newUser = res.NewUser
	return
}

// Hots is.
func (d *Dao) Hots(c context.Context) (aids []int64, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid int64 `json:"aid"`
		} `json:"list"`
	}
	if err = d.clientAsyn.Get(c, d.hot, ip, nil, &res); err != nil {
		return
	}
	code := ecode.Int(res.Code)
	if !code.Equal(ecode.OK) {
		err = errors.Wrapf(code, "%s", d.hot)
		return
	}
	for _, list := range res.List {
		if list.Aid != 0 {
			aids = append(aids, list.Aid)
		}
	}
	return
}

// TagTop is.
func (d *Dao) TagTop(c context.Context, mid, tid int64, rn int) (aids []int64, err error) {
	params := url.Values{}
	params.Set("src", "2")
	params.Set("pn", "1")
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("tag", strconv.FormatInt(tid, 10))
	params.Set("rn", strconv.Itoa(rn))
	var res struct {
		Code int     `json:"code"`
		Data []int64 `json:"data"`
	}
	if err = d.client.Get(c, d.top, "", params, &res); err != nil {
		return
	}
	code := ecode.Int(res.Code)
	if !code.Equal(ecode.OK) {
		err = errors.Wrapf(code, "%s", d.top+"?"+params.Encode())
		return
	}
	aids = res.Data
	return
}

// Group is.
func (d *Dao) Group(c context.Context) (gm map[int64]int, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	err = d.clientAsyn.Get(c, d.group, ip, nil, &gm)
	return
}

// RecommendHot list
func (d *Dao) RecommendHot(c context.Context) (rs map[int64]struct{}, err error) {
	params := url.Values{}
	params.Set("cmd", "hot")
	params.Set("from", "10")
	timeout := time.Duration(d.c.Custom.RecommendTimeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("ignore_custom", "1")
	var res struct {
		Code int `json:"code"`
		Data []struct {
			Id int64 `json:"id"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.rcmdHot, "", params, &res); err != nil {
		return
	}
	if res.Code != 0 {
		// err = errors.Wrap(err, fmt.Sprintf("code(%d)", res.Code))
		err = errors.Wrapf(ecode.Int(res.Code), "recommend url(%s) code(%d)", d.rcmdHot, res.Code)
		return
	}
	rs = map[int64]struct{}{}
	for _, l := range res.Data {
		if l.Id > 0 {
			rs[l.Id] = struct{}{}
		}
	}
	return
}
