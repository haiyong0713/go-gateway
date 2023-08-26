package archive

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/view"

	"github.com/pkg/errors"
)

const (
	_realteURL     = "/recsys/related"
	_commercialURL = "/x/internal/creative/arc/commercial"
	_relateRecURL  = "/recommand"
)

// RelateAids get relate by aid
func (d *Dao) RelateAids(c context.Context, aid int64) (aids []int64, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("key", strconv.FormatInt(aid, 10))
	var res struct {
		Code int `json:"code"`
		Data []*struct {
			Value string `json:"value"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.realteURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.realteURL+"?"+params.Encode())
		return
	}
	if len(res.Data) != 0 {
		if aids, err = xstr.SplitInts(res.Data[0].Value); err != nil {
			err = errors.Wrap(err, res.Data[0].Value)
		}
	}
	return
}

// NewRelateAids relate online recommend 在线实时推荐
func (d *Dao) NewRelateAidsV2(c context.Context, recommendReq *view.RecommendReq) (res *view.RelateResV2, returnCode string, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(recommendReq.Mid, 10))
	params.Set("cmd", recommendReq.Cmd)
	params.Set("timeout", "500")
	params.Set("plat", strconv.Itoa(int(recommendReq.Plat)))
	params.Set("build", strconv.Itoa(recommendReq.Build))
	params.Set("buvid", recommendReq.Buvid)
	params.Set("from_av", strconv.FormatInt(recommendReq.Aid, 10))
	params.Set("source_page", recommendReq.SourcePage)
	params.Set("parent_mode", strconv.Itoa(recommendReq.ParentMode))
	params.Set("auto_play", strconv.Itoa(recommendReq.AutoPlay))
	params.Set("from_trackid", recommendReq.TrackId)
	params.Set("tabid", recommendReq.TabId)
	params.Set("zone_id", strconv.FormatInt(recommendReq.ZoneId, 10))
	params.Set("in_activity", strconv.Itoa(recommendReq.IsAct))
	params.Set("copyright", strconv.Itoa(int(recommendReq.Copyright)))
	//cmd = related
	params.Set("request_cnt", "40")
	params.Set("need_dalao", "1")
	params.Set("from", "2")
	//广告+商单参数
	params.Set("ad_exp", strconv.Itoa(int(recommendReq.AdExp)))
	params.Set("is_ad", strconv.Itoa(int(recommendReq.IsAd)))
	params.Set("is_commercial", strconv.Itoa(int(recommendReq.IsCommercial)))
	params.Set("ad_resource", recommendReq.AdResource)
	params.Set("ip", ip)
	params.Set("mobi_app", recommendReq.MobileApp)
	params.Set("ad_extra", recommendReq.AdExtra)
	params.Set("av_rid", strconv.Itoa(int(recommendReq.AvRid)))
	params.Set("av_pid", strconv.Itoa(int(recommendReq.AvPid)))
	params.Set("av_tid", recommendReq.AvTid)
	params.Set("av_up_id", strconv.FormatInt(recommendReq.AvUpId, 10))
	params.Set("from_spmid", recommendReq.FromSpmid)
	params.Set("spmid", recommendReq.Spmid)
	params.Set("request_type", recommendReq.RequestType)
	params.Set("device", recommendReq.Device)
	params.Set("version", recommendReq.PageVersion)
	params.Set("network", recommendReq.Network)
	params.Set("ad_from", recommendReq.AdFrom)
	params.Set("ad_tab", strconv.FormatBool(recommendReq.AdTab))
	params.Set("rec_style", strconv.Itoa(recommendReq.RecStyle))
	params.Set("disable_rcmd", strconv.Itoa(recommendReq.DisableRcmd))
	params.Set("device_type", strconv.FormatInt(recommendReq.DeviceType, 10))
	params.Set("session_id", recommendReq.SessionId)
	if recommendReq.PageIndex > 0 {
		params.Set("page_index", strconv.FormatInt(recommendReq.PageIndex, 10))
	}
	params.Set("in_feed_play", strconv.Itoa(int(recommendReq.InfeedPlay)))
	params.Set("is_arc_pay", strconv.Itoa(int(recommendReq.IsArcPay)))
	params.Set("is_free_watch", strconv.Itoa(int(recommendReq.IsFreeWatch)))
	params.Set("is_up_blue", strconv.Itoa(int(recommendReq.IsUpBlue)))
	//refresh
	params.Set("refresh_type", strconv.FormatInt(int64(recommendReq.RefreshType), 10))
	params.Set("refresh_num", strconv.FormatInt(int64(recommendReq.RefreshNum), 10))
	if err = d.httpAiClient.Get(c, d.relateRecURL, ip, params, &res); err != nil {
		log.Error("d.relateRecURL err(%+v) url(%s)", err, d.relateRecURL+"?"+params.Encode())
		returnCode = "500"
		return
	}
	returnCode = strconv.Itoa(res.Code)
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.relateRecURL+"?"+params.Encode())
		return
	}
	return
}

// NewRelateAids relate online recommend 在线实时推荐
func (d *Dao) NewRelateAids(c context.Context, aid, mid, zoneID int64, build, parentMode, autoplay, isAct int, buvid, sourcePage, trackid, cmd, tabid string, plat int8, pageVersion, fromSpmid string) (res *view.RelateRes, returnCode string, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("cmd", cmd)
	params.Set("timeout", "200")
	params.Set("plat", strconv.Itoa(int(plat)))
	params.Set("build", strconv.Itoa(build))
	params.Set("buvid", buvid)
	params.Set("from_av", strconv.FormatInt(aid, 10))
	params.Set("source_page", sourcePage)
	params.Set("parent_mode", strconv.Itoa(parentMode))
	params.Set("auto_play", strconv.Itoa(autoplay))
	params.Set("from_trackid", trackid)
	params.Set("tabid", tabid)
	params.Set("zone_id", strconv.FormatInt(zoneID, 10))
	params.Set("in_activity", strconv.Itoa(isAct))
	params.Set("version", pageVersion)
	params.Set("from_spmid", fromSpmid)
	if cmd == model.RelateCmd {
		params.Set("request_cnt", "40")
		params.Set("need_dalao", "1")
		params.Set("from", "2")
	} else if cmd == model.RelateTabCmd {
		params.Set("request_cnt", "20")
		params.Set("need_dalao", "0")
		params.Set("from", "24")
	}
	log.Info("dalao url(%s)", d.relateRecURL+"?"+params.Encode())
	if err = d.client.Get(c, d.relateRecURL, ip, params, &res); err != nil {
		returnCode = "500"
		return
	}
	returnCode = strconv.Itoa(res.Code)
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.relateRecURL+"?"+params.Encode())
		return
	}
	return
}

// NewRelateAids relate online recommend 在线实时推荐
func (d *Dao) ContinuousPlayRelate(c context.Context, recommendReq *view.RecommendReq) (res *view.RelateResV2, returnCode string, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(recommendReq.Mid, 10))
	params.Set("cmd", recommendReq.Cmd)
	params.Set("timeout", "500")
	params.Set("plat", strconv.Itoa(int(recommendReq.Plat)))
	params.Set("build", strconv.Itoa(recommendReq.Build))
	params.Set("buvid", recommendReq.Buvid)
	params.Set("from_av", strconv.FormatInt(recommendReq.Aid, 10))
	params.Set("source_page", recommendReq.SourcePage)
	params.Set("from_trackid", recommendReq.TrackId)
	params.Set("zone_id", strconv.FormatInt(recommendReq.ZoneId, 10))
	params.Set("ip", ip)
	params.Set("mobi_app", recommendReq.MobileApp)
	params.Set("network", recommendReq.Network)
	params.Set("display_id", strconv.FormatInt(recommendReq.DisplayId, 10))
	if err = d.httpAiClient.Get(c, d.relateRecURL, ip, params, &res); err != nil {
		log.Error("d.relateRecURL err(%+v) url(%s)", err, d.relateRecURL+"?"+params.Encode())
		returnCode = "500"
		return
	}
	returnCode = strconv.Itoa(res.Code)
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.relateRecURL+"?"+params.Encode())
		return
	}
	return
}
