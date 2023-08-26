package recommend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/stat/prom"
	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	"go-gateway/app/app-svr/app-feed/interface/model/sets"

	"github.com/pkg/errors"
)

const (
	_rcmd           = "/pegasus/feed/%d"
	_recommand      = "/recommand"
	_hot            = "/data/rank/reco-tmzb.json"
	_group          = "/group_changes/pegasus.json"
	_top            = "/feed/tag/top"
	_followModeList = "/data/rank/others/followmode_whitelist.json"
	_rcmdHot        = "/recommand"
)

var (
	AICodePron = prom.New().WithCounter("tianma_code", []string{"code"})
)

// Recommend is.
func (d *Dao) Recommend(c context.Context, plat int8, buvid string, mid int64, build, loginEvent, parentMode, recsysMode,
	teenagersMode, lessonsMode int, zoneID int64, group int, interest, network string, style int, column model.ColumnStatus,
	flush, count, deviceType int, avAdResource, adResource int64, autoplay, deviceName, openEvent, bannerHash, applist,
	deviceInfo, interestSelect string, resourceID, bannerExp, adExp int, mobiApp, adExtra string, pull bool, redPoint,
	inlineSound, inlineDanmu int64, now time.Time, screenWindowType int64, disableRcmd int, localBuvid, openAppURL string,
	dituiLanding, interestId int64, interestResult string, videoMode int8) (rs *feed.AIResponse, err error) {
	rs = &feed.AIResponse{}
	if mid == 0 && buvid == "" {
		return
	}
	var (
		req    *http.Request
		ip     = metadata.String(c, metadata.RemoteIP)
		uri    = fmt.Sprintf(d.rcmd, group)
		params = url.Values{}
		isPull int
	)
	if pull {
		// pagetype，表示刷新方式：1表示下拉，0表示上滑；
		isPull = 1
	}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("plat", strconv.Itoa(int(plat)))
	params.Set("build", strconv.Itoa(build))
	params.Set("login_event", strconv.Itoa(loginEvent))
	params.Set("zone_id", strconv.FormatInt(zoneID, 10))
	params.Set("interest", interest)
	params.Set("network", network)
	if column > -1 {
		params.Set("column", strconv.Itoa(int(column)))
	}
	params.Set("style", strconv.Itoa(style))
	params.Set("flush", strconv.Itoa(flush))
	params.Set("autoplay_card", autoplay)
	params.Set("parent_mode", strconv.Itoa(parentMode))
	params.Set("recsys_mode", strconv.Itoa(recsysMode))
	params.Set("device_type", strconv.Itoa(deviceType))
	params.Set("device_name", deviceName)
	params.Set("banner_hash", bannerHash)
	params.Set("open_event", openEvent)
	params.Set("interest_select", interestSelect)
	params.Set("resid", strconv.Itoa(resourceID))
	params.Set("banner_exp", strconv.Itoa(bannerExp))
	params.Set("ip", ip)
	params.Set("mobi_app", mobiApp)
	params.Set("ad_exp", strconv.Itoa(adExp))
	params.Set("ad_extra", adExtra)
	params.Set("pagetype", strconv.Itoa(isPull))
	params.Set("screen_window_type", strconv.FormatInt(screenWindowType, 10))
	if teenagersMode != 0 {
		uri = d.recommand
		params.Set("cmd", "teenager")
		// 课堂模式和ai对接，发现此参数未做处理！！！，可忽略
		params.Set("teenagers_mode", strconv.Itoa(teenagersMode))
		params.Set("request_cnt", strconv.Itoa(count))
		params.Set("from", "7")
	} else if lessonsMode != 0 {
		uri = d.recommand
		params.Set("cmd", "lesson")
		params.Set("request_cnt", strconv.Itoa(count))
		params.Set("from", "7")
	}
	if avAdResource > 0 {
		params.Set("av_ad_resource", strconv.FormatInt(avAdResource, 10))
	}
	params.Set("ad_resource", strconv.FormatInt(adResource, 10))
	params.Set("red_point", strconv.FormatInt(redPoint, 10))
	params.Set("inline_sound", strconv.FormatInt(inlineSound, 10))
	params.Set("inline_danmu", strconv.FormatInt(inlineDanmu, 10))
	if disableRcmd > 0 {
		params.Set("disable_rcmd", strconv.Itoa(disableRcmd))
	}
	params.Set("local_buvid", localBuvid)
	if openAppURLParam, ok := resolvingOpenAppURL(openAppURL); ok {
		params.Set("open_app_url_type", openAppURLParam.Type_)
		params.Set("open_app_url_id", openAppURLParam.ID)
	}
	if dituiLanding > 0 {
		params.Set("ditui_landing", strconv.FormatInt(dituiLanding, 10))
	}
	if interestId > 0 {
		params.Set("interest_id", strconv.FormatInt(interestId, 10))
	}
	if interestResult != "" {
		params.Set("interest_result", interestResult)
	}
	params.Set("video_mode", strconv.Itoa(int(videoMode)))
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
		SceneData                  *feed.SceneData `json:"scene_data"`
		FeedTopClean               int             `json:"feed_top_clean"`
		NoPreload                  int8            `json:"no_preload"`
		ManualInline               int8            `json:"manual_inline"`
		SingleGuide                int64           `json:"single_guide"`
		DislikeText                int8            `json:"dislike_text"`
		OpenSound                  int8            `json:"open_sound"`
		AutoRefreshTimeByAppear    int64           `json:"auto_refresh_time_by_appear"`
		AutoRefreshTimeByActive    int64           `json:"auto_refresh_time_by_active"`
		TriggerLoadmoreLeftLineNum int64           `json:"trigger_loadmore_left_line_num"`
		RefreshToast               string          `json:"refresh_toast"`
		IsNaviExp                  int8            `json:"is_navi_exp"`
		RefreshTopFirstToast       string          `json:"refresh_top_first_toast"`
		RefreshTopSecondToast      string          `json:"refresh_top_second_toast"`
		HistoryCacheSize           int64           `json:"history_cache_size"`
		RefreshBarType             int8            `json:"refresh_bar_type"`
		RefreshOnBack              int8            `json:"refresh_on_back"`
		SingleRcmdReason           int8            `json:"single_rcmd_reason"`
		SmallCoverWhRatio          float32         `json:"small_cover_wh_ratio"`
		VideoMode                  int8            `json:"video_mode"`
		TopRefreshLatestExp        int8            `json:"top_refresh_latest_exp"`
		ValidShowThres             int64           `json:"valid_show_thres"`
		PegasusRefreshGuidanceExp  int8            `json:"pegasus_refresh_guidance_exp"`
		SpaceEnlargeExp            int8            `json:"space_enlarge_exp"`
		IconGuidanceExp            int8            `json:"icon_guidance_exp"`
	}
	// ad_exp=1，由ai控制广告卡片的展现；
	// ad_exp=0，网关按原有逻辑控制广告卡展现；
	if adExp == 1 {
		timeout := time.Duration(d.c.HTTPDataAd.Timeout) / time.Millisecond
		params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
		if req, err = d.clientAIAd.NewRequest(http.MethodGet, uri, ip, params); err != nil {
			rs.RespCode = ecode.ServerErr.Code()
			return
		}
		req.Header.Set("AppList", applist)
		req.Header.Set("DeviceInfo", deviceInfo)
		if err = d.clientAIAd.Do(c, req, &res); err != nil {
			rs.RespCode = ecode.ServerErr.Code()
			if errors.Cause(err) == context.DeadlineExceeded { // ai timeout{
				rs.RespCode = ecode.Deadline.Code()
			}
			return
		}
	} else {
		timeout := time.Duration(d.c.HTTPData.Timeout) / time.Millisecond
		params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
		if req, err = d.client.NewRequest(http.MethodGet, uri, ip, params); err != nil {
			rs.RespCode = ecode.ServerErr.Code()
			return
		}
		req.Header.Set("AppList", applist)
		req.Header.Set("DeviceInfo", deviceInfo)
		if err = d.client.Do(c, req, &res); err != nil {
			rs.RespCode = ecode.ServerErr.Code()
			if errors.Cause(err) == context.DeadlineExceeded { // ai timeout{
				rs.RespCode = ecode.Deadline.Code()
			}
			return
		}
	}
	code := ecode.Int(res.Code)
	AICodePron.Incr(strconv.Itoa(res.Code))
	// ai广告、只是AI接口返回的code!=0这时候还是使用AI接口返回的广告信息
	rs.BizData = res.BizData.Data
	rs.AdCode = res.BizData.Code
	if !code.Equal(ecode.OK) {
		rs.RespCode = res.Code
		err = errors.Wrap(code, uri+"?"+params.Encode())
		return
	}
	rs.Items = res.Data
	rs.InterestList = res.InterestList
	rs.UserFeature = res.UserFeature
	rs.NewUser = res.NewUser
	if res.AdData != nil {
		res.AdData.ClientIP = ip
		rs.Ad = res.AdData
	}
	rs.AutoRefreshTime = res.AutoRefreshTime
	rs.DislikeExp = res.DislikeExp
	rs.SceneURI = res.SceneData.URI()
	rs.FeedTopClean = res.FeedTopClean
	rs.NoPreload = res.NoPreload
	rs.ManualInline = res.ManualInline
	rs.SingleGuide = res.SingleGuide
	rs.DislikeText = res.DislikeText
	rs.OpenSound = res.OpenSound
	rs.AutoRefreshTimeByActive = res.AutoRefreshTimeByActive
	rs.AutoRefreshTimeByAppear = res.AutoRefreshTimeByAppear
	rs.TriggerLoadmoreLeftLineNum = res.TriggerLoadmoreLeftLineNum
	rs.RefreshToast = res.RefreshToast
	rs.IsNaviExp = res.IsNaviExp
	rs.RefreshTopFirstToast = res.RefreshTopFirstToast
	rs.RefreshTopSecondToast = res.RefreshTopSecondToast
	rs.HistoryCacheSize = res.HistoryCacheSize
	rs.RefreshBarType = res.RefreshBarType
	rs.RefreshOnBack = res.RefreshOnBack
	rs.SingleRcmdReason = res.SingleRcmdReason
	rs.SmallCoverWhRatio = res.SmallCoverWhRatio
	rs.VideoMode = res.VideoMode
	rs.TopRefreshLatestExp = res.TopRefreshLatestExp
	rs.ValidShowThres = res.ValidShowThres
	rs.PegasusRefreshGuidanceExp = res.PegasusRefreshGuidanceExp
	rs.SpaceEnlargeExp = res.SpaceEnlargeExp
	rs.IconGuidanceExp = res.IconGuidanceExp
	return
}

func resolvingOpenAppURL(in string) (*feed.OpenAppURLParam, bool) {
	if in == "" {
		return nil, false
	}
	decodedValue, err := url.QueryUnescape(in)
	if err != nil {
		log.Error("Failed to decode open_app_url: %s, %+v", in, err)
		return nil, false
	}
	out := &feed.OpenAppURLParam{}
	if err = out.UnmarshalJSON([]byte(decodedValue)); err != nil {
		log.Warn("Failed to unmarshal open_app_url: %s, %+v", decodedValue, err)
		return nil, false
	}
	typeSet := sets.NewString("av", "ss")
	if out.Jump != "feed" || !typeSet.Has(out.Type_) {
		log.Warn("Unrecognized open_app_url: %+v", out)
		return nil, false
	}
	return out, true
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
		err = errors.Wrap(code, d.hot)
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
		err = errors.Wrap(code, d.top+"?"+params.Encode())
		return
	}
	aids = res.Data
	return
}

// Group is.
func (d *Dao) Group(c context.Context) (gm map[int64]int, err error) {
	err = d.clientAsyn.Get(c, d.group, "", nil, &gm)
	return
}

// FollowModeList is.
func (d *Dao) FollowModeList(c context.Context) (list map[int64]struct{}, err error) {
	var res struct {
		Code int     `json:"code"`
		Data []int64 `json:"data"`
	}
	if err = d.clientAsyn.Get(c, d.followModeList, "", nil, &res); err != nil {
		return
	}
	code := ecode.Int(res.Code)
	if !code.Equal(ecode.OK) {
		err = errors.Wrap(code, d.followModeList)
		return
	}
	b, _ := json.Marshal(&res)
	log.Warn("FollowModeList param(%s) res(%s)", b, d.followModeList)
	list = make(map[int64]struct{}, len(res.Data))
	for _, mid := range res.Data {
		list[mid] = struct{}{}
	}
	return
}

func (d *Dao) ConvergeList(c context.Context, plat int8, buvid string, mid, convergeID int64, build, displayID int, convergeParam string, convergeType int, now time.Time) (rs *ai.ConvergeInfoV2, respCode int, err error) {
	var (
		ip                    = metadata.String(c, metadata.RemoteIP)
		params                = url.Values{}
		convergeParamUnescape string
	)
	params.Set("cmd", "convergelist")
	params.Set("from", "721")
	timeout := time.Duration(d.c.HTTPData.Timeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("build", strconv.Itoa(build))
	params.Set("plat", strconv.Itoa(int(plat)))
	params.Set("request_cnt", "10")
	params.Set("converge_id", strconv.FormatInt(convergeID, 10))
	if displayID > 0 {
		params.Set("display_id", strconv.Itoa(displayID))
	}
	if convergeParamUnescape, err = url.QueryUnescape(convergeParam); err != nil {
		convergeParamUnescape = convergeParam
		err = nil
	}
	params.Set("converge_param", convergeParamUnescape)
	params.Set("converge_type", strconv.Itoa(convergeType))
	var res struct {
		Code          int             `json:"code"`
		ConvergeTitle string          `json:"converge_title"`
		ConvergeBrief string          `json:"converge_brief"`
		UserFeature   json.RawMessage `json:"user_feature"`
		Data          []*ai.Item      `json:"data"`
	}
	if err = d.client.Get(c, d.recommand, ip, params, &res); err != nil {
		respCode = ecode.ServerErr.Code()
		return
	}
	if code := ecode.Int(res.Code); !code.Equal(ecode.OK) {
		log.Error("ConvergeList url(%s) error(%v)", d.recommand+"?"+params.Encode(), err)
		respCode = res.Code
		if code != -3 {
			err = ecode.Int(res.Code)
			return
		}
		err = nil
	}
	rs = &ai.ConvergeInfoV2{
		ConvergeInfo: &ai.ConvergeInfo{
			Title: res.ConvergeTitle,
			Items: res.Data,
		},
		Desc:        res.ConvergeBrief,
		UserFeature: res.UserFeature,
	}
	return
}

// Recommend list
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
	if err = d.clientAsyn.Get(c, d.rcmdHot, "", params, &res); err != nil {
		return
	}
	if res.Code != 0 {
		err = errors.Wrapf(ecode.Int(res.Code), "recommend hot url(%s) code(%d)", d.rcmdHot, res.Code)
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
