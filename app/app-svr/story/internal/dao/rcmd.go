package dao

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/stat"

	locationgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	"github.com/pkg/errors"
)

const (
	_story_ad_exp = "1" //story推荐接口广告开关，1有广告，0无广告
	_recommand    = "/recommand"
	_hot          = "/recommand"
	_storyBackup  = "/data/rank/story_zb.json"
)

func (d *dao) StoryRcmd(c context.Context, plat int8, build, pull int, buvid string, mid, aid, adResource int64, displayID int,
	storyParam, adExtra string, zone *locationgrpc.InfoReply, mobiApp, network string, feedStatus, fromAvID int64,
	fromTrackId string, disableRcmd, requestFrom int) (sv *ai.StoryView, respCode int, err error) {
	var (
		ip                 = metadata.String(c, metadata.RemoteIP)
		params             = url.Values{}
		storyParamUnescape string
	)
	params.Set("cmd", "story")
	// http接口超时时间
	timeout := int64(500)
	if clientConf, ok := d.httpClientCfg.HTTPData.URL[d.rcmd]; ok && clientConf != nil {
		timeout = int64(time.Duration(clientConf.Timeout) / time.Millisecond)
	}
	params.Set("timeout", strconv.FormatInt(timeout, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("build", strconv.Itoa(build))
	params.Set("plat", strconv.Itoa(int(plat)))
	params.Set("request_cnt", "10")
	params.Set("from_av", strconv.FormatInt(aid, 10))
	params.Set("display_id", strconv.Itoa(displayID))
	params.Set("ip", ip)
	params.Set("ad_exp", _story_ad_exp)
	params.Set("pull", strconv.Itoa(pull))
	if adExtra != "" {
		params.Set("ad_extra", adExtra)
	}
	params.Set("ad_resource", strconv.FormatInt(adResource, 10))
	params.Set("from_track_id", fromTrackId)
	if zone != nil {
		params.Set("zone_id", strconv.FormatInt(zone.ZoneId, 10))
		params.Set("country", zone.Country)
		params.Set("province", zone.Province)
		params.Set("city", zone.City)
	}
	if storyParamUnescape, err = url.QueryUnescape(storyParam); err != nil {
		storyParamUnescape = storyParam
		err = nil
	}
	params.Set("story_param", storyParamUnescape)
	params.Set("mobi_app", mobiApp)
	params.Set("network", network)
	params.Set("feed_status", strconv.FormatInt(feedStatus, 10))
	params.Set("from_av", strconv.FormatInt(fromAvID, 10))
	params.Set("disable_rcmd", strconv.Itoa(disableRcmd))
	params.Set("request_from", strconv.Itoa(requestFrom))
	var res struct {
		Code int `json:"code"`
		*ai.StoryView
	}
	if err = d.client.Get(c, d.rcmd, ip, params, &res); err != nil {
		respCode = ecode.ServerErr.Code()
		return
	}
	if code := ecode.Int(res.Code); !code.Equal(ecode.OK) {
		log.Error("StoryRcmd url(%s) code(%d) error(%v)", d.rcmd+"?"+params.Encode(), code, err)
		respCode = res.Code
		if code != -3 && code != 600 {
			err = ecode.Int(res.Code)
			return
		}
		err = nil
	}
	for _, item := range res.Data {
		if item == nil {
			continue
		}
		stat.MetricStoryAICardTotal.Inc(item.Goto, strconv.Itoa(int(plat)))
	}
	sv = res.StoryView
	return
}

func (d *dao) RecommendHot(c context.Context) (rs map[int64]struct{}, err error) {
	params := url.Values{}
	params.Set("cmd", "hot")
	params.Set("from", "10")
	timeout := time.Duration(200) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("ignore_custom", "1")
	var res struct {
		Code int `json:"code"`
		Data []struct {
			Id int64 `json:"id"`
		} `json:"data"`
	}
	if err = d.clientAsyn.Get(c, d.hot, "", params, &res); err != nil {
		return
	}
	if res.Code != 0 {
		err = errors.Wrapf(ecode.Int(res.Code), "recommend hot url(%s) code(%d)", d.hot, res.Code)
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

func (d *dao) StoryRcmdBackup(ctx context.Context) ([]int64, error) {
	var res struct {
		Code int     `json:"code"`
		List []int64 `json:"list"`
	}
	if err := d.clientAsyn.Get(ctx, d.storyBackup, "", nil, &res); err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, errors.Wrapf(ecode.Int(res.Code), "Failed to request story backup: %q: %d", d.storyBackup, res.Code)
	}
	out := make([]int64, 0, len(res.List))
	for _, aid := range res.List {
		if aid > 0 {
			out = append(out, aid)
		}
	}
	return out, nil
}
