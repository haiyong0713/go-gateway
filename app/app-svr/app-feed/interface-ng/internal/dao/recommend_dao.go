package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	feedngmdl "go-gateway/app/app-svr/app-feed/interface-ng/internal/model"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"

	"github.com/pkg/errors"
)

type recommendConfig struct {
	DataDiscoveryHost string
	BigDataHost       string
	DataHost          string
}

type recommendDao struct {
	client *bm.Client
	cfg    recommendConfig
}

const _hot = "/data/rank/reco-tmzb.json"

func (d *recommendDao) Hots(ctx context.Context) (sets.Int64, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid int64 `json:"aid"`
		} `json:"list"`
	}
	if err := d.client.Get(ctx, d.cfg.DataHost+_hot, ip, nil, &res); err != nil {
		return nil, err
	}
	code := ecode.Int(res.Code)
	if !code.Equal(ecode.OK) {
		return nil, errors.Wrap(code, d.cfg.DataHost+_hot)
	}
	out := sets.Int64{}
	for _, list := range res.List {
		if list.Aid != 0 {
			out.Insert(list.Aid)
		}
	}
	return out, nil
}

func (d *recommendDao) Recommend(ctx context.Context, req *feedngmdl.RecommendReq) (*feed.AIResponse, error) {
	aiResponse := &feed.AIResponse{}
	if req.Mid == 0 && req.Buvid == "" {
		return nil, errors.Errorf("empty mid and buvid")
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	uri := fmt.Sprintf(d.cfg.DataDiscoveryHost+"/pegasus/feed/%d", req.Group)
	params := url.Values{}
	isPull := 0
	if req.Pull {
		// pagetype，表示刷新方式：1表示下拉，0表示上滑；
		isPull = 1
	}
	params.Set("mid", strconv.FormatInt(req.Mid, 10))
	params.Set("buvid", req.Buvid)
	params.Set("plat", strconv.Itoa(int(req.Plat)))
	params.Set("build", strconv.Itoa(req.Build))
	params.Set("login_event", strconv.Itoa(req.LoginEvent))
	params.Set("zone_id", strconv.FormatInt(req.ZoneID, 10))
	params.Set("interest", req.Interest)
	params.Set("network", req.Network)
	if req.Column > -1 {
		params.Set("column", strconv.Itoa(int(req.Column)))
	}
	params.Set("style", strconv.Itoa(req.Style))
	params.Set("flush", strconv.Itoa(req.Flush))
	params.Set("autoplay_card", req.AutoPlay)
	params.Set("parent_mode", strconv.Itoa(req.ParentMode))
	params.Set("recsys_mode", strconv.Itoa(req.RecsysMode))
	params.Set("device_type", strconv.Itoa(req.DeviceType))
	params.Set("device_name", req.DeviceName)
	params.Set("banner_hash", req.BannerHash)
	params.Set("open_event", req.OpenEvent)
	params.Set("interest_select", req.InterestSelect)
	params.Set("resid", strconv.Itoa(req.ResourceID))
	params.Set("banner_exp", strconv.Itoa(req.BannerExp))
	params.Set("ip", ip)
	params.Set("mobi_app", req.MobiApp)
	params.Set("ad_exp", strconv.Itoa(req.AdExp))
	params.Set("ad_extra", req.AdExtra)
	params.Set("pagetype", strconv.Itoa(isPull))
	if req.TeenagersMode != 0 {
		uri = d.cfg.DataDiscoveryHost + "/recommand"
		params.Set("cmd", "teenager")
		// 课堂模式和ai对接，发现此参数未做处理！！！，可忽略
		params.Set("teenagers_mode", strconv.Itoa(req.TeenagersMode))
		params.Set("request_cnt", strconv.Itoa(req.IndexCount))
		params.Set("from", "7")
	} else if req.LessonsMode != 0 {
		uri = d.cfg.DataDiscoveryHost + "/recommand"
		params.Set("cmd", "lesson")
		params.Set("request_cnt", strconv.Itoa(req.IndexCount))
		params.Set("from", "7")
	}
	if req.AvAdResource > 0 {
		params.Set("av_ad_resource", strconv.FormatInt(req.AvAdResource, 10))
	}
	params.Set("ad_resource", strconv.FormatInt(req.AdResource, 10))
	params.Set("red_point", strconv.FormatInt(req.RedPoint, 10))
	params.Set("inline_sound", strconv.FormatInt(req.InlineSound, 10))
	params.Set("inline_danmu", strconv.FormatInt(req.InlineDanmu, 10))
	var res struct {
		Code            int             `json:"code"`
		NewUser         bool            `json:"new_user"`
		UserFeature     json.RawMessage `json:"user_feature"`
		Data            []*ai.Item      `json:"data"`
		AdData          *cm.Ad          `json:"ad_data"`
		DislikeExp      int             `json:"dislike_exp"`
		InterestList    []*ai.Interest  `json:"interest_list"`
		AutoRefreshTime int64           `json:"auto_refresh_time"`
		BizData         struct {
			Code int         `json:"code"`
			Data *ai.BizData `json:"data"`
		} `json:"biz_data"`
		SceneData    *feed.SceneData `json:"scene_data"`
		FeedTopClean int             `json:"feed_top_clean"`
		NoPreload    int8            `json:"no_preload"`
		ManualInline int8            `json:"manual_inline"`
	}
	// ad_exp=1，由ai控制广告卡片的展现；
	// ad_exp=0，网关按原有逻辑控制广告卡展现；
	timeout := 350
	if req.AdExp == 1 {
		timeout = 500
	}
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	clientReq, err := d.client.NewRequest(http.MethodGet, uri, ip, params)
	if err != nil {
		return nil, err
	}
	clientReq.Header.Set("AppList", req.AppList)
	clientReq.Header.Set("DeviceInfo", req.DeviceInfo)
	if err = d.client.Do(ctx, clientReq, &res); err != nil {
		return nil, err
	}
	code := ecode.Int(res.Code)
	aiResponse.BizData = res.BizData.Data
	aiResponse.AdCode = res.BizData.Code
	if !code.Equal(ecode.OK) {
		aiResponse.RespCode = res.Code
		return aiResponse, errors.Wrap(code, uri+"?"+params.Encode())
	}
	aiResponse.Items = res.Data
	aiResponse.InterestList = res.InterestList
	aiResponse.UserFeature = res.UserFeature
	aiResponse.NewUser = res.NewUser
	if res.AdData != nil {
		res.AdData.ClientIP = ip
		aiResponse.Ad = res.AdData
	}
	aiResponse.AutoRefreshTime = res.AutoRefreshTime
	aiResponse.DislikeExp = res.DislikeExp
	aiResponse.SceneURI = res.SceneData.URI()
	aiResponse.FeedTopClean = res.FeedTopClean
	aiResponse.NoPreload = res.NoPreload
	aiResponse.ManualInline = res.ManualInline

	return aiResponse, nil
}

func (d *recommendDao) Group(ctx context.Context) (map[int64]int, error) {
	var groupMap map[int64]int
	if err := d.client.Get(ctx, d.cfg.BigDataHost+_group, "", nil, &groupMap); err != nil {
		return nil, err
	}
	return groupMap, nil
}

const (
	_group          = "/group_changes/pegasus.json"
	_followModeList = "/data/rank/others/followmode_whitelist.json"
)

func (d *recommendDao) FollowModeSet(c context.Context) (sets.Int64, error) {
	var res struct {
		Code int     `json:"code"`
		Data []int64 `json:"data"`
	}
	if err := d.client.Get(c, d.cfg.DataHost+_followModeList, "", nil, &res); err != nil {
		return nil, err
	}
	code := ecode.Int(res.Code)
	if !code.Equal(ecode.OK) {
		return nil, errors.Wrap(code, d.cfg.DataHost+_followModeList)
	}
	return sets.NewInt64(res.Data...), nil
}
