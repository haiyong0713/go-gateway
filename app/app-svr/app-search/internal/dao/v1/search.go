package v1

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"

	errgroupv2 "go-common/library/sync/errgroup.v2"

	cdm "go-gateway/app/app-svr/app-card/interface/model"
	appdynamicgrpc "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-interface/interface-legacy/middleware/stat"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-search/internal/model/search"
	arcmiddle "go-gateway/app/app-svr/archive/middleware/v1"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	managersearch "git.bilibili.co/bapis/bapis-go/ai/search/mgr/interface"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	dynamicFeed "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	"github.com/pkg/errors"
)

const (
	// 默认zoneid 中国
	_defaultZoneID = 4194304
)

const (
	_oldAndroid = 514000
	_oldIOS     = 6090

	_searchCodeLimitAndroid = 5250001
	_searchCodeLimitIPhone  = 6680
)

func (d *dao) DynamicSearch(ctx context.Context, mid, vmid int64, keyword string, pn, ps int64) (dynamicIDs []int64, searchWords []string, total int64, err error) {
	req := &dynamicFeed.PersonalSearchReq{
		Keywords: keyword,
		Pn:       int32(pn),
		Ps:       int32(ps),
		Mid:      mid,
		UpId:     vmid,
	}
	reply, err := d.dynamicClient.PersonalSearch(ctx, req)
	if err != nil {
		return nil, nil, 0, err
	}
	var ids []int64
	for _, val := range reply.GetDynamics() {
		if val == nil {
			continue
		}
		ids = append(ids, val.DynamicId)
	}
	return ids, reply.GetTokens(), int64(reply.GetTotal()), nil
}

func (d *dao) DynamicDetail(ctx context.Context, mid int64, dynamicIDs []int64, searchWords []string, playerArgs *arcmiddle.PlayerArgs, dev device.Device, ip string, net network.Network) (map[int64]*appdynamicgrpc.DynamicItem, error) {
	req := &appdynamicgrpc.DynSpaceSearchDetailsReq{
		DynamicIds:  dynamicIDs,
		SearchWords: searchWords,
		LocalTime:   8,
		PlayerArgs:  playerArgs,
		MobiApp:     dev.RawMobiApp,
		Device:      dev.Device,
		Buvid:       dev.Buvid,
		Build:       dev.Build,
		Mid:         mid,
		Platform:    dev.RawPlatform,
		Ip:          ip,
		NetType:     appdynamicgrpc.NetworkType(net.Type),
		TfType:      appdynamicgrpc.TFType(net.TF),
	}
	reply, err := d.appDynamicClient.DynSpaceSearchDetails(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply.GetItems(), nil
}

// Search app all search .
//
//nolint:gocognit
func (d *dao) Search(c context.Context, mid int64, mobiApp, device, platform, buvid, keyword, duration, order, filtered, fromSource, recommend, parent, adExtra, extraWord, tidList, durationList, qvid string, plat int8, seasonNum, movieNum, upUserNum, uvLimit, userNum, userVideoLimit, biliUserNum, biliUserVideoLimit, rid, highlight, build, pn, ps, isQuery, teenagersMode, lessonsMode int,
	old, isOgvExpNewUser bool, now time.Time, newPGC, flow, isNewOrder bool, autoPlayCard int64) (res *search.Search, code int, err error) {
	var (
		req    *http.Request
		ip     = metadata.String(c, metadata.RemoteIP)
		ipInfo *locgrpc.InfoReply
		zoneid int64
	)
	if ipInfo, err = d.locationClient.Info(c, &locgrpc.InfoReq{Addr: ip}); err != nil {
		log.Warn("%v", err)
		err = nil
	}
	zoneid = _defaultZoneID
	if ipInfo != nil {
		zoneid = ipInfo.ZoneId
	}
	res = &search.Search{}
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("build", strconv.Itoa(build))
	params.Set("keyword", keyword)
	params.Set("main_ver", "v3")
	params.Set("highlight", strconv.Itoa(highlight))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("tids", strconv.Itoa(rid))
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("teenagers_mode", strconv.Itoa(teenagersMode))
	params.Set("lessons_mode", strconv.Itoa(lessonsMode))
	params.Set("auto_playcard", strconv.FormatInt(autoPlayCard, 10))
	params.Set("qv_id", qvid)
	if newPGC {
		params.Set("media_bangumi_num", strconv.Itoa(seasonNum))
	} else {
		params.Set("bangumi_num", strconv.Itoa(seasonNum))
		params.Set("smerge", "1")
	}
	if (plat == model.PlatIPad && build >= search.SearchNewIPad) ||
		(plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) ||
		model.IsAndroidHD(plat) {
		params.Set("is_new_user", "1")
		// iPad HD 3.16 之后也支持用户卡了
		if (plat == model.PlatIpadHD && build >= 31600000) ||
			model.IsAndroidHD(plat) ||
			(model.IsIPad(plat) && build >= 63100000) {
			params.Set("bili_user_num", strconv.Itoa(biliUserNum))
			params.Set("bili_user_vl", strconv.Itoa(10))
			params.Set("user_num", strconv.Itoa(userNum))
			params.Set("user_video_limit", strconv.Itoa(10))
		}
	} else {
		if old {
			params.Set("upuser_num", strconv.Itoa(upUserNum))
			params.Set("uv_limit", strconv.Itoa(uvLimit))
		} else {
			params.Set("bili_user_num", strconv.Itoa(biliUserNum))
			params.Set("bili_user_vl", strconv.Itoa(biliUserVideoLimit))
		}
		params.Set("user_num", strconv.Itoa(userNum))
		params.Set("user_video_limit", strconv.Itoa(userVideoLimit))
		if !d.c.Switch.SearchRecommend {
			params.Set("query_rec_need", recommend)
		}
	}
	params.Set("platform", platform)
	params.Set("duration", duration)
	params.Set("order", order)
	params.Set("search_type", "all")
	params.Set("from_source", fromSource)
	params.Set("extra_word", extraWord)
	if filtered == "1" {
		params.Set("filtered", filtered)
	}
	if model.IsOverseas(plat) {
		params.Set("use_area", "1")
	} else if newPGC {
		params.Set("media_ft_num", strconv.Itoa(movieNum))
		params.Set("is_new_pgc", "1")
	} else {
		params.Set("movie_num", strconv.Itoa(movieNum))
	}
	params.Set("zone_id", strconv.FormatInt(zoneid, 10))
	if flow {
		params.Set("flow_need", "1")
	}
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.ComicAndroid) ||
		(model.IsIPhone(plat) && build > d.c.SearchBuildLimit.ComicIOS) ||
		model.IsIPhoneB(plat) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_comic", "1")
	}
	if (model.IsAndroid(plat) && !model.IsAndroidB(plat) && build > search.SearchLiveAllAndroid) ||
		(model.IsIPhone(plat) && build > search.SearchLiveAllIOS) ||
		(plat == model.PlatIPad && build >= search.SearchNewIPad) ||
		(plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) ||
		model.IsAndroidHD(plat) {
		params.Set("new_live", "1")
	}
	if (model.IsAndroid(plat) && build > search.SearchTwitterAndroid) ||
		(model.IsIPhone(plat) && build > search.SearchTwitterIOS) ||
		model.IsIPhoneB(plat) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_twitter", "1")
	}
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.PGCHighLightAndroid) ||
		(model.IsIPhone(plat) && build > d.c.SearchBuildLimit.PGCHighLightIOS) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("app_highlight", "media_bangumi,media_ft")
	}
	if (model.IsAndroid(plat) && build >= search.SearchConvergeAndroid) ||
		(model.IsIPhone(plat) && build >= search.SearchConvergeIOS) ||
		model.IsIPhoneB(plat) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("card_num", "1")
	}
	params.Set("is_parents", parent)
	if (model.IsAndroid(plat) && build > search.SearchStarAndroid) ||
		(model.IsIPhone(plat) && build > search.SearchStarIOS) ||
		model.IsIPhoneB(plat) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_star", "1")
	}
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.PGCALLAndroid) ||
		(model.IsIPhone(plat) && build > d.c.SearchBuildLimit.PGCALLIOS) ||
		model.IsIPhoneB(plat) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_pgc_all", "1")
	}
	if (model.IsAndroid(plat) && build > search.SearchTicketAndroid) ||
		(model.IsIPhone(plat) && build > search.SearchTicketIOS) ||
		model.IsIPhoneB(plat) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_ticket", "1")
	}
	if (model.IsAndroid(plat) && build > search.SearchProductAndroid) ||
		(model.IsIPhone(plat) && build > search.SearchProductIOS) ||
		model.IsIPhoneB(plat) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_product", "1")
	}
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) ||
		(model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_special_guide", "1")
	}
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SearchArticleAndroid) ||
		(model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SearchArticleIOS) ||
		model.IsIPhoneB(plat) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_article", "1")
	}
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.ChannelAndroid) ||
		(model.IsIPhone(plat) && build > d.c.SearchBuildLimit.ChannelIOS) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_tag", "1")
	}
	if (mobiApp == "iphone" && device == "phone" && build > 8910 || mobiApp == "android" && build > 5485000) &&
		rid == 0 && order == "totalrank" && (duration == "" || duration == "0") && (durationList == "" || durationList == "0") && (filtered == "" || filtered == "0") && teenagersMode == 0 && (parent == "0" || parent == "") {
		params.Set("is_ogv", "1")
	}
	params.Set("is_org_query", strconv.Itoa(isQuery))
	if ipInfo != nil {
		params.Set("user_city", ipInfo.City)
	}
	if isNewOrder {
		params.Set("not_default_only_video", "1")
	}
	if (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.ESportsIOS) ||
		(model.IsAndroid(plat) && build > d.c.SearchBuildLimit.ESportsAndroid) ||
		(model.IsAndroidHD(plat) && build <= 1000000) ||
		(model.IsIPadHD(plat) && build >= 32500000) ||
		(model.IsIPadPink(plat) && build >= 64100000) {
		params.Set("is_esports", "1")
	}
	if (model.IsIPadHD(plat) && build >= 32600000) ||
		(model.IsIPadPink(plat) && build >= 64300000) {
		params.Set("activity", "1")
	}
	if (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialCardIOS) ||
		(model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialCardAndroid) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_channel", "1")
	}
	if (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.TipsCardIOS) ||
		(model.IsAndroid(plat) && build > d.c.SearchBuildLimit.TipsCardAndroid) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_tips", "1")
	}
	if (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.ADCardIOS) ||
		(model.IsAndroid(plat) && build > d.c.SearchBuildLimit.ADCardAndroid) {
		params.Set("is_ad", "1")
		params.Set("ad_extra", adExtra)
		if model.IsIPhone(plat) {
			params.Set("ad_resource", "4220")
		}
		if model.IsAndroid(plat) {
			params.Set("ad_resource", "4225")
		}
	}
	if (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.UserInlineLiveIOS) ||
		(model.IsAndroid(plat) && build > d.c.SearchBuildLimit.UserInlineLiveAndroid) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_bili_user_live", "1")
	}
	if (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.FlowInlineCardIOS) ||
		(model.IsAndroid(plat) && build > d.c.SearchBuildLimit.FlowInlineCardAndroid) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_live_room_inline", "1")
		params.Set("is_ugc_inline", "1")
	}
	if (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.FlowOGVInlineCardIOS) ||
		(model.IsAndroid(plat) && build > d.c.SearchBuildLimit.FlowOGVInlineCardAndroid) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_ogv_inline", "1")
	}
	if (model.IsIPhone(plat) && build > 65099999) || (model.IsAndroid(plat) && build > 6509999) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		params.Set("is_top_game", "1")
	}
	if (model.IsIPhone(plat) && build > 65699999) || (model.IsAndroid(plat) && build > 6569999) {
		params.Set("is_sports", "1")
		params.Set("is_pedia_card_inline", "1")
	}
	if (model.IsIPhone(plat) && build > 66699999) || (model.IsAndroid(plat) && build > 6669999) {
		if isOgvExpNewUser {
			params.Set("ogv_inline_new_user", "1")
		}
	}
	if (model.IsIPhone(plat) && build > 66999999) || (model.IsAndroid(plat) && build > 6699999) {
		params.Set("tid_list", tidList)
		params.Set("duration_list", durationList)
		params.Set("new_search_exp_group", "2") // 全量实验组2
	}
	if (model.IsIPhone(plat) && build > 66999999) || (model.IsAndroid(plat) && build > 6699999) {
		params.Set("video_fulltext", "1")
	}
	if (model.IsIPhone(plat) && build > 67199999) || (model.IsAndroid(plat) && build >= 6760000) {
		params.Set("is_collection_card", "1")
	}
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	if err = d.searchClient.Do(c, req, res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if (model.IsAndroid(plat) && build > _searchCodeLimitAndroid) ||
			(model.IsIPhone(plat) && build > _searchCodeLimitIPhone) ||
			(plat == model.PlatIPad && build >= search.SearchNewIPad) ||
			(plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) ||
			model.IsIPhoneB(plat) ||
			model.IsAndroidHD(plat) {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		} else if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
	}
	for _, flow := range res.FlowResult {
		flow.Change()
	}
	code = res.Code
	return
}

// Season search season data.
func (d *dao) Season(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered string, plat int8, build, pn, ps int, now time.Time) (st *search.TypeSearch, code int, err error) {
	var (
		req    *http.Request
		ip     = metadata.String(c, metadata.RemoteIP)
		ipInfo *locgrpc.InfoReply
		zoneid int64
	)
	if ipInfo, err = d.locationClient.Info(c, &locgrpc.InfoReq{Addr: ip}); err != nil {
		log.Warn("%+v", err)
		err = nil
	}
	zoneid = _defaultZoneID
	if ipInfo != nil {
		zoneid = ipInfo.ZoneId
	}
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("main_ver", "v3")
	params.Set("platform", platform)
	params.Set("build", strconv.Itoa(build))
	params.Set("keyword", keyword)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("func", "search")
	params.Set("smerge", "1")
	params.Set("search_type", "bangumi")
	params.Set("source_type", "0")
	if filtered == "1" {
		params.Set("filtered", filtered)
	}
	if model.IsOverseas(plat) {
		params.Set("use_area", "1")
	}
	params.Set("zone_id", strconv.FormatInt(zoneid, 10))
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) {
		params.Set("is_special_guide", "1")
	}
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code  int               `json:"code"`
		SeID  string            `json:"seid"`
		Pages int               `json:"numPages"`
		List  []*search.Bangumi `json:"result"`
	}
	// do
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if (model.IsAndroid(plat) && build > _searchCodeLimitAndroid) || (model.IsIPhone(plat) && build > _searchCodeLimitIPhone) || (plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) || model.IsIPhoneB(plat) {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		} else if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		code = res.Code
		return
	}
	items := make([]*search.Item, 0, len(res.List))
	for _, v := range res.List {
		si := &search.Item{}
		if (model.IsAndroid(plat) && build <= _oldAndroid) || (model.IsIPhone(plat) && build <= _oldIOS) {
			si.FromSeason(v, model.GotoBangumi)
		} else {
			si.FromSeason(v, model.GotoBangumiWeb)
		}
		items = append(items, si)
	}
	st = &search.TypeSearch{TrackID: res.SeID, Pages: res.Pages, Items: items}
	return
}

// Upper search upper data.
//
//nolint:gocognit
func (d *dao) Upper(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered, order, qvid string, biliUserVL, highlight, build, userType, orderSort, pn, ps int, old bool, now time.Time, notices map[int64]*search.SystemNotice) (st *search.TypeSearch, code int, err error) {
	var (
		req                   *http.Request
		plat                  = model.Plat(mobiApp, device)
		avids, roomIDs, uids  []int64
		apm                   map[int64]*arcgrpc.ArcPlayer
		accCards              map[int64]*accgrpc.Card
		entryRoom             map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
		ip                    = metadata.String(c, metadata.RemoteIP)
		isBlue, isNewDuration bool
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("main_ver", "v3")
	params.Set("keyword", keyword)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("highlight", strconv.Itoa(highlight))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("func", "search")
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("smerge", "1")
	params.Set("platform", platform)
	params.Set("build", strconv.Itoa(build))
	params.Set("qv_id", qvid)
	if old {
		params.Set("search_type", "upuser")
	} else {
		params.Set("search_type", "bili_user")
		params.Set("bili_user_vl", strconv.Itoa(biliUserVL))
		params.Set("user_type", strconv.Itoa(userType))
		params.Set("order_sort", strconv.Itoa(orderSort))
	}
	params.Set("order", order)
	params.Set("source_type", "0")
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) {
		params.Set("is_special_guide", "1")
	}
	if filtered == "1" {
		params.Set("filtered", filtered)
	}
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code   int            `json:"code"`
		SeID   string         `json:"seid"`
		Pages  int            `json:"numPages"`
		ExpStr string         `json:"exp_str"`
		QvId   string         `json:"qv_id"`
		List   []*search.User `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if (model.IsAndroid(plat) && build > _searchCodeLimitAndroid) || (model.IsIPhone(plat) && build > _searchCodeLimitIPhone) || (plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) || model.IsIPhoneB(plat) {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		} else if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		code = res.Code
		return
	}
	items := make([]*search.Item, 0, len(res.List))
	for _, v := range res.List {
		for _, vr := range v.Res {
			avids = append(avids, vr.Aid)
		}
		if cdm.ShowLive(mobiApp, device, build) {
			roomIDs = append(roomIDs, v.RoomID)
		}
		uids = append(uids, v.Mid)
	}
	g, ctx := errgroup.WithContext(c)
	if len(avids) != 0 {
		g.Go(func() (err error) {
			if apm, err = d.Arcs(ctx, avids, mobiApp, device, mid); err != nil {
				log.Error("Upper %+v", err)
				err = nil
			}
			return
		})
	}
	if len(uids) > 0 {
		g.Go(func() (err error) {
			if accCards, err = d.Cards3(ctx, uids); err != nil {
				log.Error("accDao.Cards Owners %+v, Err %+v", uids, err)
				err = nil
			}
			return
		})
	}
	if len(roomIDs) > 0 {
		g.Go(func() (err error) {
			entryReq := &livexroomgate.EntryRoomInfoReq{
				EntryFrom: []string{model.DefaultLiveEntry},
				RoomIds:   roomIDs,
				Uid:       mid,
				Uipstr:    metadata.String(c, metadata.RemoteIP),
				Platform:  platform,
				Build:     int64(build),
				Network:   "other",
			}
			if entryRoom, err = d.EntryRoomInfo(c, entryReq); err != nil {
				log.Error("Failed to get entry room info: %+v: %+v", req, err)
				err = nil
				return
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	if !cdm.ShowLive(mobiApp, device, build) {
		isBlue = true
	}
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.VideoDurationAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.VideoDurationIOS) {
		isNewDuration = true
	}
	for _, v := range res.List {
		si := &search.Item{}
		si.FromUpUser(v, accCards[v.Mid], apm, entryRoom[v.RoomID], isBlue, isNewDuration, notices)
		items = append(items, si)
	}
	st = &search.TypeSearch{
		TrackID: res.SeID,
		Pages:   res.Pages,
		Items:   items,
		QvId:    res.QvId,
		ExpStr:  res.ExpStr,
	}
	return
}

// MovieByType search movie data from api .
func (d *dao) MovieByType(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered string, plat int8, build, pn, ps int, now time.Time) (st *search.TypeSearch, code int, err error) {
	var (
		req    *http.Request
		avids  []int64
		apm    map[int64]*arcgrpc.ArcPlayer
		ip     = metadata.String(c, metadata.RemoteIP)
		ipInfo *locgrpc.InfoReply
		zoneid int64
	)
	if ipInfo, err = d.locationClient.Info(c, &locgrpc.InfoReq{Addr: ip}); err != nil {
		log.Warn("%v", err)
		err = nil
	}
	zoneid = _defaultZoneID
	if ipInfo != nil {
		zoneid = ipInfo.ZoneId
	}
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("keyword", keyword)
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("platform", platform)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("build", strconv.Itoa(build))
	params.Set("main_ver", "v3")
	params.Set("search_type", "pgc")
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("order", "totalrank")
	if filtered == "1" {
		params.Set("filtered", filtered)
	}
	if model.IsOverseas(plat) {
		params.Set("use_area", "1")
	}
	params.Set("zone_id", strconv.FormatInt(zoneid, 10))
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) {
		params.Set("is_special_guide", "1")
	}
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code  int             `json:"code"`
		SeID  string          `json:"seid"`
		Pages int             `json:"numPages"`
		List  []*search.Movie `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if (model.IsAndroid(plat) && build > _searchCodeLimitAndroid) || (model.IsIPhone(plat) && build > _searchCodeLimitIPhone) || (plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) || model.IsIPhoneB(plat) {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		} else if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		code = res.Code
		return
	}
	items := make([]*search.Item, 0, len(res.List))
	for _, v := range res.List {
		if v.Type == "movie" {
			avids = append(avids, v.Aid)
		}
	}
	if len(avids) != 0 {
		if apm, err = d.Arcs(c, avids, mobiApp, device, mid); err != nil {
			log.Error("RecommendNoResult %+v", err)
			err = nil
		}
	}
	for _, v := range res.List {
		si := &search.Item{}
		si.FromMovie(v, apm)
		items = append(items, si)
	}
	st = &search.TypeSearch{TrackID: res.SeID, Pages: res.Pages, Items: items}
	return
}

// LiveByType search by diff type
func (d *dao) LiveByType(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered, order, sType, qvid string, plat int8, build, pn, ps int, now time.Time) (st *search.TypeSearch, code int, err error) {
	var (
		req       *http.Request
		ip        = metadata.String(c, metadata.RemoteIP)
		ipInfo    *locgrpc.InfoReply
		roomIDs   []int64
		entryRoom map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
		zoneid    int64
	)
	if ipInfo, err = d.locationClient.Info(c, &locgrpc.InfoReq{Addr: ip}); err != nil {
		log.Warn("%v", err)
		err = nil
	}
	zoneid = _defaultZoneID
	if ipInfo != nil {
		zoneid = ipInfo.ZoneId
	}
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("keyword", keyword)
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("platform", platform)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("build", strconv.Itoa(build))
	params.Set("main_ver", "v3")
	params.Set("search_type", sType)
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("order", order)
	params.Set("qv_id", qvid)
	if filtered == "1" {
		params.Set("filtered", filtered)
	}
	if model.IsOverseas(plat) {
		params.Set("use_area", "1")
	}
	params.Set("zone_id", strconv.FormatInt(zoneid, 10))
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) {
		params.Set("is_special_guide", "1")
	}
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code   int            `json:"code"`
		SeID   string         `json:"seid"`
		Pages  int            `json:"numPages"`
		ExpStr string         `json:"exp_str,omitempty"`
		QvId   string         `json:"qv_id"`
		List   []*search.Live `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if (model.IsAndroid(plat) && build > _searchCodeLimitAndroid) || (model.IsIPhone(plat) && build > _searchCodeLimitIPhone) || (plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) || model.IsIPhoneB(plat) {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		} else if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		code = res.Code
		return
	}
	for _, v := range res.List {
		roomIDs = append(roomIDs, v.RoomID)
	}
	if len(roomIDs) > 0 {
		entryReq := &livexroomgate.EntryRoomInfoReq{
			EntryFrom: []string{model.DefaultLiveEntry},
			RoomIds:   roomIDs,
			Uid:       mid,
			Uipstr:    metadata.String(c, metadata.RemoteIP),
			Platform:  platform,
			Build:     int64(build),
			Network:   "other",
		}
		if entryRoom, err = d.EntryRoomInfo(c, entryReq); err != nil {
			log.Error("Failed to get entry room info: %+v: %+v", req, err)
			err = nil
		}
	}
	items := make([]*search.Item, 0, len(res.List))
	for _, v := range res.List {
		si := &search.Item{}
		si.FromLive(v, entryRoom[v.RoomID], nil)
		items = append(items, si)
	}
	st = &search.TypeSearch{TrackID: res.SeID, Pages: res.Pages, Items: items, ExpStr: res.ExpStr, KeyWord: keyword, QvId: res.QvId}
	return
}

// Live search for live
func (d *dao) Live(c context.Context, mid int64, keyword, mobiApp, platform, buvid, device, order, sType, qvid string, build, pn, ps int) (st *search.TypeSearch, err error) {
	var (
		req            *http.Request
		plat           = model.Plat(mobiApp, device)
		roomIDs, upIDs []int64
		entryRoom      map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
		accCards       map[int64]*accgrpc.Card
		ip             = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("keyword", keyword)
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("build", strconv.Itoa(build))
	params.Set("main_ver", "v3")
	params.Set("search_type", sType)
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("device", device)
	params.Set("order", order)
	params.Set("qv_id", qvid)
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) {
		params.Set("is_special_guide", "1")
	}
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code   int            `json:"code"`
		SeID   string         `json:"seid"`
		ExpStr string         `json:"exp_str"`
		QvId   string         `json:"qv_id"`
		Pages  int            `json:"numPages"`
		List   []*search.Live `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if (model.IsAndroid(plat) && build > _searchCodeLimitAndroid) || (model.IsIPhone(plat) && build > _searchCodeLimitIPhone) || (plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) || model.IsIPhoneB(plat) {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		} else if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		return
	}
	eg, ctx := errgroup.WithContext(c)
	if !model.IsAndroid(plat) || (model.IsAndroid(plat) && build > search.LiveBroadcastTypeAndroid) {
		for _, v := range res.List {
			roomIDs = append(roomIDs, v.RoomID)
			upIDs = append(upIDs, v.UID)
		}
		if len(roomIDs) > 0 {
			eg.Go(func() (err error) {
				entryReq := &livexroomgate.EntryRoomInfoReq{
					EntryFrom: []string{model.DefaultLiveEntry},
					RoomIds:   roomIDs,
					Uid:       mid,
					Uipstr:    metadata.String(c, metadata.RemoteIP),
					Platform:  platform,
					Build:     int64(build),
					Network:   "other",
				}
				if entryRoom, err = d.EntryRoomInfo(ctx, entryReq); err != nil {
					log.Error("Failed to get entry room info: %+v: %+v", req, err)
					return nil
				}
				return nil
			})
		}
		if len(upIDs) > 0 {
			eg.Go(func() (err error) {
				if accCards, err = d.Cards3(ctx, upIDs); err != nil {
					log.Error("s.accDao.Cards3 err(%+v)", err)
					return nil
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Error("%+v", err)
		}
	}
	items := make([]*search.Item, 0, len(res.List))
	for _, v := range res.List {
		si := &search.Item{}
		si.FromLive2(v, entryRoom[v.RoomID], accCards[v.UID])
		items = append(items, si)
	}
	st = &search.TypeSearch{TrackID: res.SeID, Pages: res.Pages, Items: items, ExpStr: res.ExpStr, KeyWord: keyword, QvId: res.QvId}
	return
}

// LiveAll search for live version > 5.28

func (d *dao) LiveAll(c context.Context, mid int64, keyword, mobiApp, platform, buvid, device, order, sType string, build, pn, ps int) (st *search.TypeSearchLiveAll, err error) {
	var (
		req            *http.Request
		plat           = model.Plat(mobiApp, device)
		roomIDs, upIDs []int64
		entryRoom      map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
		accCards       map[int64]*accgrpc.Card
		ip             = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("keyword", keyword)
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("build", strconv.Itoa(build))
	params.Set("main_ver", "v3")
	params.Set("search_type", sType)
	params.Set("page", strconv.Itoa(pn))
	params.Set("live_room_num", strconv.Itoa(ps))
	params.Set("device", device)
	params.Set("order", order)
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) {
		params.Set("is_special_guide", "1")
	}
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code     int    `json:"code"`
		SeID     string `json:"seid"`
		ExpStr   string `json:"exp_str"`
		Pages    int    `json:"numPages"`
		PageInfo *struct {
			Master *search.Live `json:"live_master"`
			Room   *search.Live `json:"live_room"`
		} `json:"pageinfo"`
		List *struct {
			Master []*search.Live `json:"live_master"`
			Room   []*search.Live `json:"live_room"`
		} `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if (model.IsAndroid(plat) && build > _searchCodeLimitAndroid) || (model.IsIPhone(plat) && build > _searchCodeLimitIPhone) || model.IsPad(plat) || model.IsIPhoneB(plat) {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		} else if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		return
	}
	st = &search.TypeSearchLiveAll{
		TrackID: res.SeID,
		ExpStr:  res.ExpStr,
		KeyWord: keyword,
		Pages:   res.Pages,
		Master:  &search.TypeSearch{},
		Room:    &search.TypeSearch{},
	}
	if res.PageInfo != nil {
		if res.PageInfo.Master != nil {
			st.Master.Pages = res.PageInfo.Master.Pages
			st.Master.Total = res.PageInfo.Master.Total
		}
		if res.PageInfo.Room != nil {
			st.Room.Pages = res.PageInfo.Room.Pages
			st.Room.Total = res.PageInfo.Room.Total
		}
	}
	if res.List != nil {
		eg, ctx := errgroup.WithContext(c)
		if !model.IsAndroid(plat) || (model.IsAndroid(plat) && build > search.LiveBroadcastTypeAndroid) {
			for _, v := range res.List.Master {
				roomIDs = append(roomIDs, v.RoomID)
				upIDs = append(upIDs, v.UID)
			}
			for _, v := range res.List.Room {
				roomIDs = append(roomIDs, v.RoomID)
				upIDs = append(upIDs, v.UID)
			}
			if len(roomIDs) > 0 {
				eg.Go(func() (err error) {
					entryReq := &livexroomgate.EntryRoomInfoReq{
						EntryFrom: []string{model.DefaultLiveEntry},
						RoomIds:   roomIDs,
						Uid:       mid,
						Uipstr:    metadata.String(c, metadata.RemoteIP),
						Platform:  platform,
						Build:     int64(build),
						Network:   "other",
					}
					if entryRoom, err = d.EntryRoomInfo(ctx, entryReq); err != nil {
						log.Error("Failed to get entry room info: %+v: %+v", req, err)
						return nil
					}
					return nil
				})
			}
			if len(upIDs) > 0 {
				eg.Go(func() (err error) {
					if accCards, err = d.Cards3(ctx, upIDs); err != nil {
						log.Error("s.accDao.Cards3 err(%+v)", err)
						return nil
					}
					return nil
				})
			}
			if err := eg.Wait(); err != nil {
				log.Error("%+v", err)
			}
		}
		st.Master.Items = make([]*search.Item, 0, len(res.List.Master))
		for _, v := range res.List.Master {
			si := &search.Item{}
			extFunc := []func(*search.Item){}
			extFunc = append(extFunc, search.WithLiveParentArea(mobiApp, build))
			si.FromLiveMaster(v, entryRoom[v.RoomID], accCards[v.UID], extFunc...)
			st.Master.Items = append(st.Master.Items, si)
		}
		st.Room.Items = make([]*search.Item, 0, len(res.List.Room))
		for _, v := range res.List.Room {
			si := &search.Item{}
			si.FromLive2(v, entryRoom[v.RoomID], accCards[v.UID])
			st.Room.Items = append(st.Room.Items, si)
		}
	}
	return
}

// ArticleByType search article.
func (d *dao) ArticleByType(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered, order, sType, qvid string, plat int8, categoryID, build, highlight, pn, ps int, now time.Time) (st *search.TypeSearch, code int, err error) {
	var (
		req    *http.Request
		ip     = metadata.String(c, metadata.RemoteIP)
		ipInfo *locgrpc.InfoReply
		zoneid int64
	)
	if ipInfo, err = d.locationClient.Info(c, &locgrpc.InfoReq{Addr: ip}); err != nil {
		log.Warn("%v", err)
		err = nil
	}
	zoneid = _defaultZoneID
	if ipInfo != nil {
		zoneid = ipInfo.ZoneId
	}
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("keyword", keyword)
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("platform", platform)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("build", strconv.Itoa(build))
	params.Set("main_ver", "v3")
	params.Set("highlight", strconv.Itoa(highlight))
	params.Set("search_type", sType)
	params.Set("category_id", strconv.Itoa(categoryID))
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("order", order)
	params.Set("qv_id", qvid)
	if filtered == "1" {
		params.Set("filtered", filtered)
	}
	if model.IsOverseas(plat) {
		params.Set("use_area", "1")
	}
	params.Set("zone_id", strconv.FormatInt(zoneid, 10))
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) {
		params.Set("is_special_guide", "1")
	}
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code  int               `json:"code"`
		SeID  string            `json:"seid"`
		Pages int               `json:"numPages"`
		QvId  string            `json:"qv_id"`
		List  []*search.Article `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		code = res.Code
		return
	}
	items := make([]*search.Item, 0, len(res.List))
	for _, v := range res.List {
		si := &search.Item{}
		si.FromArticle(v, nil)
		items = append(items, si)
	}
	st = &search.TypeSearch{TrackID: res.SeID, Pages: res.Pages, Items: items, QvId: res.QvId}
	return
}

// HotSearch is hot words search data.
func (d *dao) HotSearch(c context.Context, buvid string, mid int64, build, limit, zoneId int, mobiApp, device, platform string, now time.Time) (res *search.Hot, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("main_ver", "v3")
	params.Set("actionKey", "appkey")
	params.Set("limit", strconv.Itoa(limit))
	params.Set("build", strconv.Itoa(build))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("platform", platform)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("zone_id", strconv.Itoa(zoneId))
	req, err := d.searchClient.NewRequest("GET", d.hot, ip, params)
	if err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.hot+"?"+params.Encode())
	}
	return
}

func (d *dao) Trending(c context.Context, buvid string, mid int64, build, limit, zoneId int, mobiApp, device, platform string, now time.Time, isRanking bool) (res *search.Hot, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("main_ver", "v3")
	params.Set("actionKey", "appkey")
	params.Set("limit", strconv.Itoa(limit))
	params.Set("build", strconv.Itoa(build))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("platform", platform)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("zone_id", strconv.Itoa(zoneId))
	if isRanking {
		params.Set("hotword_list", "1")
	}
	req, err := d.searchClient.NewRequest("GET", d.trending, ip, params)
	if err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		stat.MetricSearchAiMainFailed.Inc(d.trending, strconv.Itoa(res.Code))
		err = errors.Wrap(ecode.Int(res.Code), d.trending+"?"+params.Encode())
	}
	return
}

// Suggest suggest data.
//
//nolint:gocognit
func (d *dao) Suggest(c context.Context, mid int64, buvid, term string, build int, mobiApp, device string, now time.Time) (res *search.Suggest, err error) {
	plat := model.Plat(mobiApp, device)
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("main_ver", "v4")
	params.Set("func", "suggest")
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("build", strconv.Itoa(build))
	params.Set("mobi_app", mobiApp)
	params.Set("bangumi_acc_num", "3")
	params.Set("special_acc_num", "3")
	params.Set("topic_acc_num", "3")
	params.Set("upuser_acc_num", "1")
	params.Set("tag_num", "10")
	params.Set("special_num", "10")
	params.Set("bangumi_num", "10")
	params.Set("upuser_num", "3")
	params.Set("suggest_type", "accurate")
	params.Set("term", term)
	res = &search.Suggest{}
	if err = d.searchClient.Get(c, d.suggest, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if (model.IsAndroid(plat) && build > _searchCodeLimitAndroid) || (model.IsIPhone(plat) && build > _searchCodeLimitIPhone) || (plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) || model.IsIPhoneB(plat) {
			err = errors.Wrap(ecode.Int(res.Code), d.suggest+"?"+params.Encode())
		} else if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.suggest+"?"+params.Encode())
		}
		return
	}
	if res.ResultBs == nil {
		return
	}
	switch v := res.ResultBs.(type) {
	case []interface{}:
		return
	case map[string]interface{}:
		if acc, ok := v["accurate"]; ok {
			if accm, ok := acc.(map[string]interface{}); ok && accm != nil {
				res.Result.Accurate.UpUser = accm["upuser"]
				res.Result.Accurate.Bangumi = accm["bangumi"]
			}
		}
		if tag, ok := v["tag"]; ok {
			if tags, ok := tag.([]interface{}); ok {
				for _, t := range tags {
					if tm, ok := t.(map[string]interface{}); ok && tm != nil {
						if v, ok := tm["value"]; ok {
							if vs, ok := v.(string); ok {
								res.Result.Tag = append(res.Result.Tag, &struct {
									Value string `json:"value,omitempty"`
								}{vs})
							}
						}
					}
				}
			}
		}
	}
	return
}

// Suggest2 suggest data.
func (d *dao) Suggest2(c context.Context, mid int64, platform, buvid, term string, build int, mobiApp string, now time.Time) (res *search.Suggest2, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("main_ver", "v4")
	params.Set("suggest_type", "accurate")
	params.Set("platform", platform)
	params.Set("mobi_app", mobiApp)
	params.Set("clientip", ip)
	params.Set("build", strconv.Itoa(build))
	if mid != 0 {
		params.Set("userid", strconv.FormatInt(mid, 10))
	}
	params.Set("term", term)
	params.Set("sug_num", "10")
	params.Set("buvid", buvid)
	res = &search.Suggest2{}
	if err = d.searchClient.Get(c, d.suggest, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.suggest+"?"+params.Encode())
	}
	return
}

// Suggest3 suggest data.
func (d *dao) Suggest3(c context.Context, mid int64, platform, buvid, term, device string, build, highlight int, mobiApp string, now time.Time) (res *search.Suggest3, err error) {
	var (
		req  *http.Request
		plat = model.Plat(mobiApp, device)
		ip   = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("suggest_type", "accurate")
	params.Set("platform", platform)
	params.Set("mobi_app", mobiApp)
	params.Set("clientip", ip)
	params.Set("highlight", strconv.Itoa(highlight))
	params.Set("build", strconv.Itoa(build))
	if mid != 0 {
		params.Set("userid", strconv.FormatInt(mid, 10))
	}
	params.Set("term", term)
	params.Set("sug_num", "10")
	params.Set("buvid", buvid)
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SugDetailAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SugDetailIOS) {
		params.Set("is_detail_info", "1")
	}
	if req, err = d.searchClient.NewRequest("GET", d.suggest3, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	res = &search.Suggest3{}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.suggest3+"?"+params.Encode())
		return
	}
	for _, flow := range res.Result {
		flow.SugChange()
	}
	return
}

//nolint:gocognit
func (d *dao) Season2(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, qvid string, highlight, build, pn, ps int, fnver, fnval, qn, fourk int64) (st *search.TypeSearch, code int, err error) {
	var (
		req       *http.Request
		plat      = model.Plat(mobiApp, device)
		ip        = metadata.String(c, metadata.RemoteIP)
		seasonIDs []int64
		bangumis  map[string]*search.Card
		sepReqs   []*pgcsearch.SeasonEpReq
		seasonEps map[int32]*pgcsearch.SearchCardProto
		medias    map[int32]*pgcsearch.SearchMediaProto
		ipInfo    *locgrpc.InfoReply
		zoneid    int64
	)
	if ipInfo, err = d.locationClient.Info(c, &locgrpc.InfoReq{Addr: ip}); err != nil {
		log.Warn("%v", err)
		err = nil
	}
	zoneid = _defaultZoneID
	if ipInfo != nil {
		zoneid = ipInfo.ZoneId
	}
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("main_ver", "v3")
	params.Set("platform", platform)
	params.Set("build", strconv.Itoa(build))
	params.Set("keyword", keyword)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("search_type", "media_bangumi")
	params.Set("order", "totalrank")
	params.Set("highlight", strconv.Itoa(highlight))
	params.Set("app_highlight", "media_bangumi")
	params.Set("zone_id", strconv.FormatInt(zoneid, 10))
	params.Set("qv_id", qvid)
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.PGCALLAndroid) || (model.IsIPhone(plat) && build >= d.c.SearchBuildLimit.PGCALLIOS) || model.IsIPhoneB(plat) {
		params.Set("is_pgc_all", "1")
	}
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) {
		params.Set("is_special_guide", "1")
	}
	if (model.IsAndroid(plat) && build > 6640000) || (model.IsIPhone(plat) && build > 66400000) {
		params.Set("is_recommend", "1")
	}
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code              int             `json:"code"`
		SeID              string          `json:"seid"`
		Total             int             `json:"numResults"`
		Pages             int             `json:"numPages"`
		ExpStr            string          `json:"exp_str"`
		QvId              string          `json:"qv_id"`
		ResultIsRecommend int             `json:"result_is_recommend"`
		List              []*search.Media `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if (model.IsAndroid(plat) && build > _searchCodeLimitAndroid) || (model.IsIPhone(plat) && build > _searchCodeLimitIPhone) || (plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) || model.IsIPhoneB(plat) {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		} else if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		code = res.Code
		return
	}
	for _, v := range res.List {
		seasonIDs = append(seasonIDs, v.SeasonID)
		if v.Canplay() {
			sepReqs = append(sepReqs, v.BuildPgcReq())
		}
	}
	if len(seasonIDs) > 0 {
		if bangumis, err = d.BangumiCard(c, mid, seasonIDs); err != nil {
			log.Error("Season2 %+v", err)
			err = nil
		}
		var isWithPlayURL bool
		if (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.TypeSearchWithPlayURLIOS) || (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.TypeSearchWithPlayURLAndroid) {
			isWithPlayURL = true
		}
		if len(sepReqs) > 0 {
			if seasonEps, medias, err = d.SearchPGCCards(c, sepReqs, keyword, mobiApp, device, platform, mid, fnver, fnval, qn, fourk, int64(build), isWithPlayURL); err != nil {
				log.Error("bangumiDao SearchPGCCards %v", err)
				err = nil
			}
		}
	}
	items := make([]*search.Item, 0, len(res.List))
	isOgvExpNewUser := d.CheckNewDeviceAndUser(c, mid, buvid, model.NewUserOgvExperimentPeriod)
	for _, v := range res.List {
		si := &search.Item{}
		switch v.Type {
		case search.TypeRecommendTips:
			si.FromRecommendTips(v)
		default:
			var extFunc []func(*search.Item)
			if isOgvExpNewUser {
				extFunc = append(extFunc, search.WithOgvNewUserUpdateBadges(c, v, seasonEps))
			}
			si.FromMediaPgcCard(v, "", model.GotoBangumi, bangumis, seasonEps, medias, d.c.Cfg.PgcSearchCard, model.IsPad(plat), extFunc...) // 新增参数指是否ipad垂搜
		}
		items = append(items, si)
	}
	st = &search.TypeSearch{
		TrackID:           res.SeID,
		Pages:             res.Pages,
		Total:             res.Total,
		Items:             items,
		ExpStr:            res.ExpStr,
		QvId:              res.QvId,
		ResultIsRecommend: res.ResultIsRecommend,
	}
	return
}

// MovieByType2 search new movie data from api .
//
//nolint:gocognit
func (d *dao) MovieByType2(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, qvid string, highlight, build, pn, ps int, fnver, fnval, qn, fourk int64) (st *search.TypeSearch, code int, err error) {
	var (
		req       *http.Request
		plat      = model.Plat(mobiApp, device)
		ip        = metadata.String(c, metadata.RemoteIP)
		seasonIDs []int64
		sepReqs   []*pgcsearch.SeasonEpReq
		seasonEps map[int32]*pgcsearch.SearchCardProto
		bangumis  map[string]*search.Card
		medias    map[int32]*pgcsearch.SearchMediaProto
		ipInfo    *locgrpc.InfoReply
		zoneid    int64
	)
	if ipInfo, err = d.locationClient.Info(c, &locgrpc.InfoReq{Addr: ip}); err != nil {
		log.Warn("%v", err)
		err = nil
	}
	zoneid = _defaultZoneID
	if ipInfo != nil {
		zoneid = ipInfo.ZoneId
	}
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("keyword", keyword)
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("platform", platform)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("build", strconv.Itoa(build))
	params.Set("main_ver", "v3")
	params.Set("search_type", "media_ft")
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("order", "totalrank")
	params.Set("highlight", strconv.Itoa(highlight))
	params.Set("app_highlight", "media_ft")
	params.Set("zone_id", strconv.FormatInt(zoneid, 10))
	params.Set("qv_id", qvid)
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.PGCALLAndroid) || (model.IsIPhone(plat) && build >= d.c.SearchBuildLimit.PGCALLIOS) || model.IsIPhoneB(plat) {
		params.Set("is_pgc_all", "1")
	}
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) {
		params.Set("is_special_guide", "1")
	}
	if (model.IsAndroid(plat) && build > 6640000) || (model.IsIPhone(plat) && build > 66400000) {
		params.Set("is_recommend", "1")
	}
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code   int             `json:"code"`
		SeID   string          `json:"seid"`
		Total  int             `json:"numResults"`
		Pages  int             `json:"numPages"`
		ExpStr string          `json:"exp_str"`
		QvId   string          `json:"qv_id"`
		List   []*search.Media `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if (model.IsAndroid(plat) && build > _searchCodeLimitAndroid) || (model.IsIPhone(plat) && build > _searchCodeLimitIPhone) || (plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) || model.IsIPhoneB(plat) {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		} else if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		code = res.Code
		return
	}
	for _, v := range res.List {
		seasonIDs = append(seasonIDs, v.SeasonID)
		if v.Canplay() {
			sepReqs = append(sepReqs, v.BuildPgcReq())
		}
	}
	if len(seasonIDs) > 0 {
		if bangumis, err = d.BangumiCard(c, mid, seasonIDs); err != nil {
			log.Error("MovieByType2 %+v", err)
			err = nil
		}
		var isWithPlayURL bool
		if (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.TypeSearchWithPlayURLIOS) || (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.TypeSearchWithPlayURLAndroid) {
			isWithPlayURL = true
		}
		if seasonEps, medias, err = d.SearchPGCCards(c, sepReqs, keyword, mobiApp, device, platform, mid, fnver, fnval, qn, fourk, int64(build), isWithPlayURL); err != nil {
			log.Error("bangumiDao SearchPGCCards %v", err)
			err = nil
		}
	}
	items := make([]*search.Item, 0, len(res.List))
	isOgvExpNewUser := d.CheckNewDeviceAndUser(c, mid, buvid, model.NewUserOgvExperimentPeriod)
	for _, v := range res.List {
		si := &search.Item{}
		switch v.Type {
		case search.TypeRecommendTips:
			si.FromRecommendTips(v)
		default:
			var extFunc []func(*search.Item)
			if isOgvExpNewUser {
				extFunc = append(extFunc, search.WithOgvNewUserUpdateBadges(c, v, seasonEps))
			}
			si.FromMediaPgcCard(v, "", model.GotoMovie, bangumis, seasonEps, medias, d.c.Cfg.PgcSearchCard, model.IsPad(plat), extFunc...) // 新增参数指是否ipad垂搜
		}
		items = append(items, si)
	}
	st = &search.TypeSearch{
		TrackID: res.SeID,
		Pages:   res.Pages,
		Total:   res.Total,
		Items:   items,
		QvId:    res.QvId,
		ExpStr:  res.ExpStr,
	}
	return
}

// User search user data.
func (d *dao) User(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered, order, fromSource string, highlight, build, userType, orderSort, pn, ps int, now time.Time) (user []*search.User, err error) {
	var (
		req  *http.Request
		plat = model.Plat(mobiApp, device)
		ip   = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("platform", platform)
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("build", strconv.Itoa(build))
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("keyword", keyword)
	params.Set("highlight", strconv.Itoa(highlight))
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("main_ver", "v3")
	params.Set("func", "search")
	params.Set("smerge", "1")
	params.Set("source_type", "0")
	params.Set("search_type", "bili_user")
	params.Set("user_type", strconv.Itoa(userType))
	params.Set("order", order)
	params.Set("order_sort", strconv.Itoa(orderSort))
	params.Set("from_source", fromSource)
	if filtered == "1" {
		params.Set("filtered", filtered)
	}
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code  int            `json:"code"`
		SeID  string         `json:"seid"`
		Pages int            `json:"numPages"`
		List  []*search.User `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if (model.IsAndroid(plat) && build > _searchCodeLimitAndroid) || (model.IsIPhone(plat) && build > _searchCodeLimitIPhone) || (plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) || model.IsIPhoneB(plat) {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		} else if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		return
	}
	user = res.List
	return
}

// Recommend is recommend search data.
func (d *dao) Recommend(c context.Context, mid int64, build, from, show, disableRcmd int, buvid, platform, mobiApp, device string) (res *search.RecommendResult, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("main_ver", "v3")
	params.Set("platform", platform)
	params.Set("mobi_app", mobiApp)
	params.Set("clientip", ip)
	params.Set("device", device)
	params.Set("build", strconv.Itoa(build))
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("search_type", "guess")
	params.Set("req_source", strconv.Itoa(from))
	params.Set("show_area", strconv.Itoa(show))
	params.Set("disable_rcmd", strconv.Itoa(disableRcmd))
	req, err := d.searchClient.NewRequest("GET", d.rcmd, ip, params)
	if err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var rcmdRes struct {
		Code      int    `json:"code,omitempty"`
		SeID      string `json:"seid,omitempty"`
		Tips      string `json:"recommend_tips,omitempty"`
		NumResult int    `json:"numResult,omitempty"`
		Resutl    []struct {
			ID   int64  `json:"id,omitempty"`
			Name string `json:"name,omitempty"`
			Type string `json:"type,omitempty"`
			Pos  int    `json:"pos,omitempty"`
		} `json:"result,omitempty"`
		ExpStr string `json:"exp_str,omitempty"`
	}
	if err = d.searchClient.Do(c, req, &rcmdRes); err != nil {
		return
	}
	if rcmdRes.Code != ecode.OK.Code() {
		if rcmdRes.Code != model.ForbidCode {
			err = errors.Wrap(ecode.Int(rcmdRes.Code), d.rcmd+"?"+params.Encode())
		}
		return
	}
	res = &search.RecommendResult{
		TrackID: rcmdRes.SeID,
		Title:   rcmdRes.Tips,
		Pages:   rcmdRes.NumResult,
		ExpStr:  rcmdRes.ExpStr,
	}
	for _, v := range rcmdRes.Resutl {
		item := &search.Item{}
		item.ID = v.ID
		item.Param = strconv.Itoa(int(v.ID))
		item.Title = v.Name
		item.Type = v.Type
		item.Position = v.Pos
		res.Items = append(res.Items, item)
	}
	return
}

// DefaultWords is default words search data.
func (d *dao) DefaultWords(c context.Context, mid int64, build, from int, buvid, platform, mobiApp, device string, loginEvent int64, extParam *search.DefaultWordsExtParam) (res *search.DefaultWords, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	plat := model.Plat(mobiApp, device)
	params := url.Values{}
	params.Set("main_ver", "v3")
	params.Set("platform", platform)
	params.Set("mobi_app", mobiApp)
	params.Set("clientip", ip)
	params.Set("device", device)
	params.Set("build", strconv.Itoa(build))
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("search_type", "default")
	params.Set("req_source", strconv.Itoa(from))
	params.Set("login_event", strconv.FormatInt(loginEvent, 10))
	if extParam != nil {
		params.Set("tab", extParam.Tab)
		params.Set("event_id", extParam.EventId)
		params.Set("avid", extParam.Avid)
		params.Set("query", extParam.Query)
		params.Set("an", strconv.FormatInt(extParam.An, 10))
		params.Set("is_fresh", strconv.FormatInt(extParam.IsFresh, 10))
		params.Set("disable_rcmd", strconv.FormatInt(extParam.DisableRcmd, 10))
	}
	if (plat == model.PlatAndroid && build > d.c.SearchBuildLimit.DefaultWordJumpAndroid) ||
		(plat == model.PlatIPhone && build >= d.c.SearchBuildLimit.DefaultWordJumpIOS) ||
		(plat == model.PlatAndroidI && build > d.c.SearchBuildLimit.DefaultWordJumpAndroidI) ||
		(model.IsAndroidHD(plat)) {
		params.Set("is_new", "1")
	}
	req, err := d.searchClient.NewRequest("GET", d.rcmd, ip, params)
	if err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var rcmdRes struct {
		Code      int    `json:"code,omitempty"`
		SeID      string `json:"seid,omitempty"`
		Tips      string `json:"recommend_tips,omitempty"`
		NumResult int    `json:"numResult,omitempty"`
		ShowFront int    `json:"show_front,omitempty"`
		Resutl    []struct {
			ID        int64  `json:"id,omitempty"`
			Name      string `json:"name,omitempty"`
			ShowName  string `json:"show_name,omitempty"`
			Type      string `json:"type,omitempty"`
			GotoType  int    `json:"goto_type,omitempty"`
			GotoValue string `json:"goto_value,omitempty"`
			ModuleID  int64  `json:"module_id,omitempty"`
		} `json:"result,omitempty"`
		Trackid string `json:"trackid,omitempty"`
		ExpStr  string `json:"exp_str,omitempty"`
	}
	if err = d.searchClient.Do(c, req, &rcmdRes); err != nil {
		return
	}
	if rcmdRes.Code != ecode.OK.Code() {
		if rcmdRes.Code != model.ForbidCode {
			err = errors.Wrap(ecode.Int(rcmdRes.Code), d.rcmd+"?"+params.Encode())
		}
		return
	}
	res = &search.DefaultWords{}
	if len(rcmdRes.Resutl) == 0 {
		return
	}
	for _, v := range rcmdRes.Resutl {
		res.Trackid = rcmdRes.SeID
		res.Param = strconv.Itoa(int(v.ID))
		res.Show = v.ShowName
		res.Word = v.Name
		res.ShowFront = rcmdRes.ShowFront
		res.ExpStr = rcmdRes.ExpStr
		switch v.GotoType {
		case search.DefaultWordTypeArchive:
			res.Goto = model.GotoAv
			res.Value = v.GotoValue
			res.URI = model.FillURI(res.Goto, res.Value, nil)
		case search.DefaultWordTypeArticle:
			res.Goto = model.GotoArticle
			res.Value = v.GotoValue
			res.URI = model.FillURI(res.Goto, res.Value, nil)
		case search.DefaultWordTypePGC:
			res.Goto = model.GotoEP
			res.Value = v.GotoValue
			res.URI = model.FillURI(model.GotoPGC, res.Value, nil)
		case search.DefaultWordTypeURL:
			res.Goto = model.GotoWeb
			res.URI = model.FillURI(res.Goto, v.GotoValue, nil)
		}
		v.GotoType = 0
		v.GotoValue = ""
	}
	return
}

// RecommendNoResult is no result recommend search data.
func (d *dao) RecommendNoResult(c context.Context, platform, mobiApp, device, buvid, keyword string, build, pn, ps int, mid int64) (res *search.NoResultRcndResult, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("main_ver", "v3")
	params.Set("platform", platform)
	params.Set("mobi_app", mobiApp)
	params.Set("clientip", ip)
	params.Set("device", device)
	params.Set("build", strconv.Itoa(build))
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("search_type", "video")
	params.Set("keyword", keyword)
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	req, err := d.searchClient.NewRequest("GET", d.rcmdNoResult, ip, params)
	if err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var (
		resTmp      *search.NoResultRcmd
		avids       []int64
		apm         map[int64]*arcgrpc.ArcPlayer
		cooperation bool
	)
	if err = d.searchClient.Do(c, req, &resTmp); err != nil {
		return
	}
	if resTmp.Code != ecode.OK.Code() {
		if resTmp.Code != model.ForbidCode {
			err = errors.Wrap(ecode.Int(resTmp.Code), d.rcmdNoResult+"?"+params.Encode())
		}
		return
	}
	res = &search.NoResultRcndResult{TrackID: resTmp.Trackid, Title: resTmp.RecommendTips, Pages: resTmp.NumResults}
	for _, v := range resTmp.Result {
		avids = append(avids, v.ID)
	}
	if len(avids) != 0 {
		if apm, err = d.Arcs(c, avids, mobiApp, device, mid); err != nil {
			log.Error("RecommendNoResult %+v", err)
			err = nil
		}
	}
	items := make([]*search.Item, 0, len(resTmp.Result))
	for _, v := range resTmp.Result {
		ri := &search.Item{}
		ri.FromVideo(v, apm[v.ID], cooperation, false, false, false, false, false, "", nil)
		items = append(items, ri)
	}
	res.Items = items
	return
}

// Channel for search channel
func (d *dao) Channel(c context.Context, mid int64, keyword, mobiApp, platform, buvid, device, order, sType string, build, pn, ps, highlight int) (st *search.TypeSearch, code int, err error) {
	var (
		req  *http.Request
		plat = model.Plat(mobiApp, device)
		ip   = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("keyword", keyword)
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("build", strconv.Itoa(build))
	params.Set("main_ver", "v3")
	params.Set("search_type", sType)
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("device", device)
	params.Set("order", order)
	params.Set("highlight", strconv.Itoa(highlight))
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) {
		params.Set("is_special_guide", "1")
	}
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code  int               `json:"code"`
		SeID  string            `json:"seid"`
		Pages int               `json:"numPages"`
		List  []*search.Channel `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		code = res.Code
		return
	}
	items := make([]*search.Item, 0, len(res.List))
	for _, v := range res.List {
		si := &search.Item{}
		apm := make(map[int64]*arcgrpc.ArcPlayer)
		si.FromChannel(v, apm, nil, "type_search", order)
		items = append(items, si)
	}
	st = &search.TypeSearch{TrackID: res.SeID, Pages: res.Pages, Items: items}
	return
}

// RecommendPre search at pre-page
func (d *dao) RecommendPre(c context.Context, platform, mobiApp, device, buvid string, build, ps int, mid int64) (res *search.RecommendPreResult, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("main_ver", "v3")
	params.Set("platform", platform)
	params.Set("mobi_app", mobiApp)
	params.Set("clientip", ip)
	params.Set("device", device)
	params.Set("build", strconv.Itoa(build))
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("search_type", "discover_page")
	params.Set("pagesize", strconv.Itoa(ps))
	req, err := d.searchClient.NewRequest("GET", d.pre, ip, params)
	if err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var (
		resTmp    *search.RecommendPre
		avids     []int64
		avm       map[int64]*arcgrpc.Arc
		seasonIDs []int32
		bangumis  map[int32]*seasongrpc.CardInfoProto
	)
	if err = d.searchClient.Do(c, req, &resTmp); err != nil {
		return
	}
	if resTmp.Code != ecode.OK.Code() {
		if resTmp.Code != model.ForbidCode {
			err = errors.Wrap(ecode.Int(resTmp.Code), d.pre+"?"+params.Encode())
		}
		return
	}
	for _, v := range resTmp.Result {
		for _, vv := range v.List {
			if vv.Type == "video" {
				avids = append(avids, vv.ID)
			} else if vv.Type == "pgc" {
				seasonIDs = append(seasonIDs, int32(vv.ID))
			}
		}
	}
	g, ctx := errgroup.WithContext(c)
	if len(avids) != 0 {
		g.Go(func() (err error) {
			if avm, err = d.Archives(ctx, avids, mobiApp, device, mid); err != nil {
				log.Error("RecommendPre avids(%v) error(%v)", avids, err)
				err = nil
			}
			return
		})
	}
	if len(seasonIDs) > 0 {
		g.Go(func() (err error) {
			if bangumis, err = d.SeasonCards(c, seasonIDs); err != nil {
				log.Error("RecommendPre seasonIDs(%v) error(%v)", seasonIDs, err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	res = &search.RecommendPreResult{TrackID: resTmp.Trackid, Total: resTmp.NumResult}
	items := make([]*search.Item, 0, len(resTmp.Result))
	for _, v := range resTmp.Result {
		rs := &search.Item{Title: v.Query}
		for _, vv := range v.List {
			if vv.Type == "video" {
				if a, ok := avm[vv.ID]; ok {
					r := &search.Item{}
					r.FromRcmdPre(vv.ID, a, nil)
					rs.Item = append(rs.Item, r)
				}
			} else if vv.Type == "pgc" {
				if b, ok := bangumis[int32(vv.ID)]; ok {
					r := &search.Item{}
					r.FromRcmdPre(vv.ID, nil, b)
					rs.Item = append(rs.Item, r)
				}
			}
		}
		items = append(items, rs)
	}
	res.Items = items
	return
}

// Video search new archive data.
func (d *dao) Video(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, order string, highlight, build, pn, ps int) (st *search.TypeSearch, code int, err error) {
	var (
		req         *http.Request
		ip          = metadata.String(c, metadata.RemoteIP)
		plat        = model.Plat(mobiApp, device)
		avids       []int64
		apm         map[int64]*arcgrpc.ArcPlayer
		cooperation bool
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("main_ver", "v3")
	params.Set("platform", platform)
	params.Set("build", strconv.Itoa(build))
	params.Set("keyword", keyword)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("search_type", "video")
	params.Set("order", "totalrank")
	params.Set("highlight", strconv.Itoa(highlight))
	if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.SpecialerGuideAndroid) || (model.IsIPhone(plat) && build > d.c.SearchBuildLimit.SpecialerGuideIOS) {
		params.Set("is_special_guide", "1")
	}
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code  int             `json:"code"`
		SeID  string          `json:"seid"`
		Total int             `json:"numResults"`
		Pages int             `json:"numPages"`
		List  []*search.Video `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		code = res.Code
		return
	}
	for _, v := range res.List {
		avids = append(avids, v.ID)
	}
	if len(avids) > 0 {
		if apm, err = d.Arcs(c, avids, mobiApp, device, mid); err != nil {
			log.Error("Upper %+v", err)
			err = nil
		}
	}
	items := make([]*search.Item, 0, len(res.List))
	for _, v := range res.List {
		si := &search.Item{}
		if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.CardOptimizeAndroid) ||
			(model.IsIPhone(plat) && build > d.c.SearchBuildLimit.CardOptimizeIPhone) ||
			(plat == model.PlatIpadHD && build > d.c.SearchBuildLimit.CardOptimizeIpadHD) {
			si.FromVideo(v, apm[v.ID], cooperation, false, false, true, false, false, order, nil)
		} else {
			si.FromVideo(v, apm[v.ID], cooperation, false, false, false, false, false, order, nil)
		}
		items = append(items, si)
	}
	st = &search.TypeSearch{TrackID: res.SeID, Pages: res.Pages, Total: res.Total, Items: items}
	return
}

// Follow picks upper recommend data from search API
func (d *dao) Follow(c context.Context, platform, mobiApp, device, buvid string, build int, mid, vmid int64) (ups []*search.Upper, trackID string, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("platform", platform)
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("clientip", ip)
	params.Set("build", strconv.Itoa(build))
	params.Set("buvid", buvid)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("context_id", strconv.FormatInt(vmid, 10))
	params.Set("rec_type", "up_rec")
	params.Set("pagesize", "20")
	params.Set("service_area", "space_suggest")
	var res struct {
		Code    int             `json:"code"`
		TrackID string          `json:"trackid"`
		Msg     string          `json:"msg"`
		Data    []*search.Upper `json:"data"`
	}
	if err = d.searchClient.Get(c, d.upper, ip, params, &res); err != nil {
		return
	}
	code := ecode.Int(res.Code)
	if !code.Equal(ecode.OK) {
		err = errors.Wrap(code, d.upper+"?"+params.Encode())
		return
	}
	ups = res.Data
	trackID = res.TrackID
	return
}

// Converge search Converge.
func (d *dao) Converge(c context.Context, _, cid int64, trackID, platform, mobiApp, device, buvid, order, sort string, plat int8, build, pn, ps int) (st *search.ResultConverge, err error) {
	var (
		req *http.Request
		ip  = metadata.String(c, metadata.RemoteIP)
		res *search.Converge
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("main_ver", "v3")
	params.Set("search_type", "card_content")
	params.Set("platform", platform)
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("build", strconv.Itoa(build))
	params.Set("keyword", strconv.FormatInt(cid, 10))
	params.Set("refer_seid", trackID)
	params.Set("order", order)
	params.Set("order_sort", sort)
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		return
	}
	var userItems, videoItems []*search.Item
	for _, v := range res.Result.User {
		si := &search.Item{}
		si.FromConverge2(v, nil)
		userItems = append(userItems, si)
	}
	for _, v := range res.Result.Video {
		si := &search.Item{}
		si.FromConverge2(nil, v)
		videoItems = append(videoItems, si)
	}
	st = &search.ResultConverge{
		TrackID:    res.SeID,
		Pages:      res.Pages,
		Total:      res.Total,
		UserItems:  userItems,
		VideoItems: videoItems,
		ExpStr:     res.ExpStr,
	}
	return
}

// Space get space search.
func (d *dao) Space(c context.Context, mobiApp, platform, device, keyword, group, order, fromSource, buvid string, plat int8, build, rid, isTitle, highlight, pn, ps int, vmid, mid, attrNot int64, now time.Time) (res *search.Space, err error) {
	var (
		req      *http.Request
		ip       = metadata.String(c, metadata.RemoteIP)
		attrPugv = int64(1 << 30) // pugv付费attribute=30位
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("clientip", ip)
	params.Set("device", device)
	params.Set("build", strconv.Itoa(build))
	params.Set("keyword", keyword)
	params.Set("search_type", "sub_video")
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("highlight", strconv.Itoa(highlight))
	params.Set("mid", strconv.FormatInt(vmid, 10))
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("attr_not", strconv.FormatInt(attrPugv|attrNot, 10))
	if fromSource != "" {
		params.Set("from_source", fromSource)
	}
	if rid > 0 {
		params.Set("tid", strconv.Itoa(rid))
	}
	if group != "" {
		params.Set("group", group)
	}
	params.Set("additional_ranks", "-6")
	if order != "" {
		params.Set("order", order)
	}
	if isTitle != 0 {
		params.Set("is_title", "1")
	}
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.space, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	// do
	if err = d.searchClient.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.space+"?"+params.Encode())
		log.Error("%v", err)
	}
	return
}

// Channel for search channel
func (d *dao) ChannelNew(c context.Context, mid int64, keyword, mobiApp, platform, buvid, device string, build, pn, ps, highlight int) (st *search.ChannelResult, tids []int64, err error) {
	var (
		req *http.Request
		ip  = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("keyword", keyword)
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("build", strconv.Itoa(build))
	params.Set("main_ver", "v3")
	params.Set("search_type", "tag")
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("device", device)
	params.Set("highlight", strconv.Itoa(highlight))
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.main, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code   int               `json:"code"`
		SeID   string            `json:"seid"`
		Pages  int               `json:"numPages"`
		Total  int               `json:"numResults"`
		ExpStr string            `json:"exp_str"`
		List   []*search.Channel `json:"result"`
	}
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		if res.Code != model.ForbidCode && res.Code != model.NoResultCode {
			err = errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
		}
		return
	}
	for _, v := range res.List {
		if v != nil && v.TagID != 0 {
			tids = append(tids, v.TagID)
		}
	}
	st = &search.ChannelResult{TrackID: res.SeID, Pages: res.Pages, Total: res.Total, ExpStr: res.ExpStr}
	return
}

// SearchTips for manager search tips
func (d *dao) SearchTips(c context.Context) (map[int64]*search.SearchTips, error) {
	var (
		req *http.Request
		ip  = metadata.String(c, metadata.RemoteIP)
		err error
	)
	params := url.Values{}
	// ids传空值，返回所有生效中卡片列表
	params.Set("ids", "")
	// new request
	if req, err = d.searchClient.NewRequest("GET", d.searchTips, ip, params); err != nil {
		return nil, err
	}
	var res struct {
		Code int `json:"code"`
		Data struct {
			Items []*search.SearchTips `json:"items"`
		} `json:"data"`
	}
	// do
	if err = d.searchClient.Do(c, req, &res); err != nil {
		return nil, err
	}
	// = 0 返回成功
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.searchTips+"?"+params.Encode())
		return nil, err
	}
	tipsMap := make(map[int64]*search.SearchTips)
	for _, tips := range res.Data.Items {
		if tips == nil {
			continue
		}
		// Status 状态：2 手动下线；1 上线中；0 未上线
		if tips.Status == 1 {
			tipsMap[tips.Id] = tips
		}
	}
	return tipsMap, nil
}

func (d *dao) GetEsportConfigs(ctx context.Context, req *managersearch.GetEsportConfigsReq) (*managersearch.GetEsportConfigsResp, error) {
	return d.managersearch.GetEsportConfigs(ctx, req)
}

func (d *dao) CheckNewDeviceAndUser(ctx context.Context, mid int64, buvid, periods string) bool {
	var (
		isNewBuvid, isNewMid bool
	)
	eg := errgroupv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if buvid == "" {
			isNewBuvid = true
			return nil
		}
		isNewBuvid = d.CheckRegTime(ctx, &accgrpc.CheckRegTimeReq{Buvid: buvid, Periods: periods})
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if mid == 0 {
			isNewMid = true // 未登录时判定为新用户
			return nil
		}
		isNewMid = d.CheckRegTime(ctx, &accgrpc.CheckRegTimeReq{Mid: mid, Periods: periods})
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("CheckNewDeviceAndUser() eg.Wait err: %+v", err)
		return false
	}
	return isNewBuvid && isNewMid
}
