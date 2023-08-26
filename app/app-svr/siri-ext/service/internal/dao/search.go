package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/siri-ext/service/internal/model"

	"github.com/pkg/errors"
)

// const (
// 	// 默认zoneid 中国
// 	_defaultZoneID = 4194304
// )

// // Search app all search .
// func (d *dao) Search(c context.Context, mid int64, mobiApp, device, platform, buvid, keyword, duration, order, filtered, fromSource, recommend, parent string, plat int8, seasonNum, movieNum, upUserNum, uvLimit, userNum, userVideoLimit, biliUserNum, biliUserVideoLimit, rid, highlight, build, pn, ps, isQuery, teenagersMode, lessonsMode int, old bool, now time.Time, newPGC, flow, isNewOrder bool) (res *search.Search, code int, err error) {
// 	var (
// 		req    *http.Request
// 		ip     = metadata.String(c, metadata.RemoteIP)
// 		ipInfo *locgrpc.InfoReply
// 		zoneid int64
// 	)
// 	if ipInfo, err = d.locgrpc.Info(c, &locgrpc.InfoReq{Addr: ip}); err != nil {
// 		log.Warn("%v", err)
// 		err = nil
// 	}
// 	zoneid = _defaultZoneID
// 	if ipInfo != nil {
// 		zoneid = ipInfo.ZoneId
// 	}
// 	res = &search.Search{}
// 	params := url.Values{}
// 	params.Set("is_bvid", "1")
// 	params.Set("build", strconv.Itoa(build))
// 	params.Set("keyword", keyword)
// 	params.Set("main_ver", "v3")
// 	params.Set("highlight", strconv.Itoa(highlight))
// 	params.Set("mobi_app", mobiApp)
// 	params.Set("device", device)
// 	params.Set("userid", strconv.FormatInt(mid, 10))
// 	params.Set("tids", strconv.Itoa(rid))
// 	params.Set("page", strconv.Itoa(pn))
// 	params.Set("pagesize", strconv.Itoa(ps))
// 	params.Set("teenagers_mode", strconv.Itoa(teenagersMode))
// 	params.Set("lessons_mode", strconv.Itoa(lessonsMode))
// 	if newPGC {
// 		params.Set("media_bangumi_num", strconv.Itoa(seasonNum))
// 	} else {
// 		params.Set("bangumi_num", strconv.Itoa(seasonNum))
// 		params.Set("smerge", "1")
// 	}
// 	if (plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) {
// 		params.Set("is_new_user", "1")
// 	} else {
// 		if old {
// 			params.Set("upuser_num", strconv.Itoa(upUserNum))
// 			params.Set("uv_limit", strconv.Itoa(uvLimit))
// 		} else {
// 			params.Set("bili_user_num", strconv.Itoa(biliUserNum))
// 			params.Set("bili_user_vl", strconv.Itoa(biliUserVideoLimit))
// 		}
// 		params.Set("user_num", strconv.Itoa(userNum))
// 		params.Set("user_video_limit", strconv.Itoa(userVideoLimit))
// 		params.Set("query_rec_need", recommend)
// 	}
// 	params.Set("platform", platform)
// 	params.Set("duration", duration)
// 	params.Set("order", order)
// 	params.Set("search_type", "all")
// 	params.Set("from_source", fromSource)
// 	if filtered == "1" {
// 		params.Set("filtered", filtered)
// 	}
// 	if model.IsOverseas(plat) {
// 		params.Set("use_area", "1")
// 	} else if newPGC {
// 		params.Set("media_ft_num", strconv.Itoa(movieNum))
// 		params.Set("is_new_pgc", "1")
// 	} else {
// 		params.Set("movie_num", strconv.Itoa(movieNum))
// 	}
// 	params.Set("zone_id", strconv.FormatInt(zoneid, 10))
// 	if flow {
// 		params.Set("flow_need", "1")
// 	}
// 	params.Set("is_comic", "1")
// 	params.Set("new_live", "1")
// 	params.Set("is_twitter", "1")
// 	params.Set("app_highlight", "media_bangumi,media_ft")
// 	params.Set("card_num", "1")
// 	params.Set("is_parents", parent)
// 	params.Set("is_star", "1")
// 	params.Set("is_pgc_all", "1")
// 	params.Set("is_ticket", "1")
// 	params.Set("is_product", "1")
// 	params.Set("is_special_guide", "1")
// 	params.Set("is_article", "1")
// 	params.Set("is_tag", "1")
// 	params.Set("is_ogv", "1")
// 	params.Set("is_org_query", strconv.Itoa(isQuery))
// 	if ipInfo != nil {
// 		params.Set("user_city", ipInfo.City)
// 	}
// 	params.Set("not_default_only_video", "1")
// 	params.Set("is_esports", "1")
// 	params.Set("is_channel", "1")
// 	params.Set("is_tips", "1")
// 	// new request
// 	apiURL := d.host.Search + "/main/search"
// 	if req, err = d.httpClient.NewRequest("GET", apiURL, ip, params); err != nil {
// 		return
// 	}
// 	req.Header.Set("Buvid", buvid)
// 	if err = d.httpClient.Do(c, req, res); err != nil {
// 		return
// 	}
// 	if res.Code != ecode.OK.Code() {
// 		err = errors.Wrap(ecode.Int(res.Code), apiURL+"?"+params.Encode())
// 	}
// 	for _, flow := range res.FlowResult {
// 		flow.Change()
// 	}
// 	code = res.Code
// 	return
// }

// Suggest3 suggest data.
func (d *dao) Suggest3(ctx context.Context, arg *model.SearchSuggestReq) (*model.Suggest3, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("suggest_type", "accurate")
	params.Set("platform", arg.Platform)
	params.Set("mobi_app", arg.MobiApp)
	params.Set("clientip", ip)
	params.Set("highlight", strconv.FormatInt(arg.Highlight, 10))
	params.Set("build", strconv.FormatInt(arg.Build, 10))
	if arg.Mid != 0 {
		params.Set("userid", strconv.FormatInt(arg.Mid, 10))
	}
	params.Set("term", arg.Term)
	params.Set("sug_num", "10")
	params.Set("buvid", arg.Buvid)
	params.Set("is_detail_info", "1")
	apiURL := d.host.Search + "/main/suggest/new"
	req, err := d.httpClient.NewRequest("GET", apiURL, ip, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Buvid", arg.Buvid)
	res := &model.Suggest3{}
	if err := d.httpClient.Do(ctx, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), apiURL+"?"+params.Encode())
	}
	for _, flow := range res.Result {
		flow.SugChange()
	}
	return res, nil
}
