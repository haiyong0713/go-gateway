package search

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
	"go-common/library/naming/discovery"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"
	appdynamicgrpc "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/search"
	arcmiddle "go-gateway/app/app-svr/archive/middleware/v1"

	dynamicFeed "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	upgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"

	"github.com/pkg/errors"
)

const (
	_rcmd     = "/query/recommend"
	_suggest3 = "/main/suggest/new"
	_upper    = "/main/recommend"
	_space    = "/space/search/v2"
)

// Dao is search dao
type Dao struct {
	c        *conf.Config
	client   *httpx.Client
	rcmd     string
	suggest3 string
	upper    string
	space    string

	appDynamicClient  appdynamicgrpc.DynamicClient
	dynamicFeedClient dynamicFeed.FeedClient
	upClient          upgrpc.UpArchiveClient
}

// New initial search dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:        c,
		client:   httpx.NewClient(c.HTTPSearch, httpx.SetResolver(resolver.New(nil, discovery.Builder()))),
		rcmd:     c.HostDiscovery.Search + _rcmd,
		suggest3: c.HostDiscovery.Search + _suggest3,
		upper:    c.HostDiscovery.Search + _upper,
		space:    c.HostDiscovery.Search + _space,
	}
	var err error
	if d.dynamicFeedClient, err = dynamicFeed.NewClient(c.FeedClient); err != nil {
		panic(err)
	}
	if d.upClient, err = upgrpc.NewClient(c.UpArcGRPC); err != nil {
		panic(err)
	}
	if d.appDynamicClient, err = appdynamicgrpc.NewClient(c.AppDynamicGRPC); err != nil {
		panic(err)
	}
	return
}

// Suggest3 suggest data.
func (d *Dao) Suggest3(c context.Context, mid int64, platform, buvid, term, device string, build, highlight int, mobiApp string, now time.Time) (res *search.Suggest3, err error) {
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
	if req, err = d.client.NewRequest("GET", d.suggest3, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	res = &search.Suggest3{}
	if err = d.client.Do(c, req, &res); err != nil {
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

// DefaultWords is default words search data.
func (d *Dao) DefaultWords(c context.Context, mid int64, build, from int, buvid, platform, mobiApp, device string, loginEvent int64, extParam *search.DefaultWordsExtParam) (res *search.DefaultWords, err error) {
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
	// if (model.IsAndroid(plat) && build > d.c.SearchBuildLimit.DefaultWordJumpAndroid) || (model.IsIPhone(plat) && build >= d.c.SearchBuildLimit.DefaultWordJumpIOS) {
	if (plat == model.PlatAndroid && build > d.c.SearchBuildLimit.DefaultWordJumpAndroid) ||
		(plat == model.PlatIPhone && build >= d.c.SearchBuildLimit.DefaultWordJumpIOS) ||
		(plat == model.PlatAndroidI && build > d.c.SearchBuildLimit.DefaultWordJumpAndroidI) ||
		(model.IsAndroidHD(plat)) {
		params.Set("is_new", "1")
	}
	req, err := d.client.NewRequest("GET", d.rcmd, ip, params)
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
	if err = d.client.Do(c, req, &rcmdRes); err != nil {
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

func (d *Dao) ArcPassedSearch(ctx context.Context, vmid int64, keyword string, highlight bool, kwFields []upgrpc.KwField, order upgrpc.SearchOrder, sort string, pn, ps int64, isIpad bool) (*upgrpc.ArcPassedSearchReply, error) {
	var without []upgrpc.Without
	if !isIpad {
		without = append(without, upgrpc.Without_no_space)
	}
	req := &upgrpc.ArcPassedSearchReq{
		Mid:       vmid,
		Keyword:   keyword,
		Pn:        pn,
		Ps:        ps,
		Highlight: highlight,
		KwFields:  kwFields,
		Order:     order,
		Sort:      sort,
		Without:   without,
	}
	return d.upClient.ArcPassedSearch(ctx, req)
}

// Space get space search.
func (d *Dao) Space(c context.Context, mobiApp, platform, device, keyword, group, order, fromSource, buvid string, plat int8, build, rid, isTitle, highlight, pn, ps int, vmid, mid, attrNot int64, now time.Time) (res *search.Space, err error) {
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
	if req, err = d.client.NewRequest("GET", d.space, ip, params); err != nil {
		return
	}
	req.Header.Set("Buvid", buvid)
	// do
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.space+"?"+params.Encode())
		log.Error("%v", err)
	}
	return
}

func (d *Dao) DynamicSearch(ctx context.Context, mid, vmid int64, keyword string, pn, ps int64) (dynamicIDs []int64, searchWords []string, total int64, err error) {
	req := &dynamicFeed.PersonalSearchReq{
		Keywords: keyword,
		Pn:       int32(pn),
		Ps:       int32(ps),
		Mid:      mid,
		UpId:     vmid,
	}
	reply, err := d.dynamicFeedClient.PersonalSearch(ctx, req)
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

func (d *Dao) DynamicDetail(ctx context.Context, mid int64, dynamicIDs []int64, searchWords []string, playerArgs *arcmiddle.PlayerArgs, dev device.Device, ip string, net network.Network) (map[int64]*appdynamicgrpc.DynamicItem, error) {
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

// Follow picks upper recommend data from search API
func (d *Dao) Follow(c context.Context, platform, mobiApp, device, buvid string, build int, mid, vmid int64) (ups []*search.Upper, trackID string, err error) {
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
	if err = d.client.Get(c, d.upper, ip, params, &res); err != nil {
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
