package dao

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/web/interface/model"
	"go-gateway/app/web-svr/web/interface/model/search"
	"go-gateway/pkg/idsafe/bvid"

	"github.com/pkg/errors"
)

const (
	_searchVer           = "v3"
	_searchWebPlatform   = "web"
	_searchXcxPlatform   = "xcx"
	_searchFromSourceXcx = "xcx_search"
	_searchUpRecType     = "up_rec"
	_searchTipDetail     = "/x/admin/feed/open/search/tips"
	// 用户如果是up主的话,附属视频个数
	_biliUserVl     = "8" // 使用昵称搜索的返回视频个数
	_userVideoLimit = "8" // 使用uid搜索的返回视频个数
)

// SearchAll search all data.
func (d *Dao) SearchAll(ctx context.Context, mid int64, arg *search.SearchAllArg, buvid, ua, typ string) (*search.Search, error) {
	params := url.Values{}
	ip := metadata.String(ctx, metadata.RemoteIP)
	var platform string
	switch arg.FromSource {
	case _searchFromSourceXcx:
		platform = _searchXcxPlatform
	default:
		platform = _searchWebPlatform
	}
	params = setSearchParam(params, mid, search.SearchTypeAll, arg.Keyword, platform, arg.FromSource, buvid, ip, arg.MobiApp)
	params.Set("duration", strconv.Itoa(arg.Duration))
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("tids", strconv.Itoa(arg.Rid))
	params.Set("is_bvid", "1")
	params.Set("from_spmid", arg.FromSpmid)
	params.Set("dynamic_offset", strconv.FormatInt(arg.DynamicOffset, 10))
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	if typ == search.WxSearchType {
		params.Set("highlight", strconv.Itoa(arg.Highlight))
		for k, v := range search.SearchDefaultArg[search.WxSearchTypeAll] {
			params.Set(k, strconv.Itoa(v))
		}
	} else {
		for k, v := range search.SearchDefaultArg[search.SearchTypeAll] {
			params.Set(k, strconv.Itoa(v))
		}
		params.Set("single_column", strconv.Itoa(arg.SingleColumn))
		params.Set("web_highlight", "media_bangumi,media_ft")
	}
	if arg.Platform == "pc" {
		params.Set("is_tips", "1")
		params.Set("is_esports", "1")
		params.Set("video_special_need", "1")
		params.Set("bili_user_vl", _biliUserVl)
		params.Set("user_video_limit", _userVideoLimit)
		if arg.PageSize > 0 {
			params.Set("video_num", strconv.FormatInt(arg.PageSize, 10))
		}
	}
	req, err := d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("browser-info", ua)
	res := &search.Search{}
	if err := d.httpSearch.Do(ctx, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.searchURL+"?"+params.Encode())
	}
	return res, nil
}

func (d *Dao) SearchTipDetail(c context.Context) (map[int64]*model.TipDetail, error) {
	params := url.Values{}
	// ids传空值，返回所有生效中卡片列表
	params.Set("ids", "")
	req, err := d.httpR.NewRequest("GET", d.searchTipDetail, "", params)
	if err != nil {
		return nil, err
	}
	var res struct {
		Code int `json:"code"`
		Data struct {
			Items []*model.TipDetail `json:"items"`
		} `json:"data"`
	}
	if err := d.httpSearch.Do(c, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.searchTipDetail+"?"+params.Encode())
		return nil, err
	}
	data := map[int64]*model.TipDetail{}
	for _, val := range res.Data.Items {
		if val == nil {
			continue
		}
		data[val.ID] = val
	}
	return data, nil
}

// SearchVideo search season data.
func (d *Dao) SearchVideo(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchVideoRes, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchTypeParam(params, mid, search.SearchTypeVideo, buvid, ip, arg)
	params.Set("duration", strconv.Itoa(arg.Duration))
	params.Set("order", arg.Order)
	params.Set("from_source", arg.FromSource)
	params.Set("tids", strconv.FormatInt(arg.Rid, 10))
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("is_bvid", "1")
	params.Set("from_spmid", arg.FromSpmid)
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	params.Set("dynamic_offset", strconv.FormatInt(arg.DynamicOffset, 10))
	for k, v := range search.SearchDefaultArg[search.SearchTypeVideo] {
		params.Set(k, strconv.Itoa(v))
	}
	if arg.PageSize > 0 {
		params.Set("pagesize", strconv.FormatInt(arg.PageSize, 10))
	}
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchVideoRes)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("SearchVideo d.httpSearch.Get(%s) error(%v)", d.searchURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchVideo d.httpSearch.Get(%s) code(%d) error", d.searchURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchBangumi search bangumi data.
func (d *Dao) SearchBangumi(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchPGCRes, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchTypeParam(params, mid, search.SearchTypeBangumi, buvid, ip, arg)
	params.Set("duration", strconv.Itoa(arg.Duration))
	params.Set("order", arg.Order)
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("web_highlight", "media_bangumi")
	params.Set("is_bvid", "1")
	params.Set("from_spmid", arg.FromSpmid)
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	params.Set("dynamic_offset", strconv.FormatInt(arg.DynamicOffset, 10))
	for k, v := range search.SearchDefaultArg[search.SearchTypeBangumi] {
		params.Set(k, strconv.Itoa(v))
	}
	if arg.PageSize > 0 {
		params.Set("pagesize", strconv.FormatInt(arg.PageSize, 10))
	}
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchPGCRes)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("SearchBangumi d.httpSearch.Get(%s) error(%v)", d.searchURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchBangumi d.httpSearch.Get(%s) code(%d) error", d.searchURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchMovie search movie data.
func (d *Dao) SearchMovie(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchPGCRes, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchTypeParam(params, mid, search.SearchTypeMovie, buvid, ip, arg)
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("web_highlight", "media_ft")
	params.Set("from_spmid", arg.FromSpmid)
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	params.Set("dynamic_offset", strconv.FormatInt(arg.DynamicOffset, 10))
	for k, v := range search.SearchDefaultArg[search.SearchTypeMovie] {
		params.Set(k, strconv.Itoa(v))
	}
	if arg.PageSize > 0 {
		params.Set("pagesize", strconv.FormatInt(arg.PageSize, 10))
	}
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchPGCRes)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("SearchPGC d.httpSearch.Get(%s) error(%v)", d.searchURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchPGC d.httpSearch.Get(%s) code(%d) error", d.searchURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchLive search live data.
func (d *Dao) SearchLive(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchTypeRes, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchTypeParam(params, mid, search.SearchTypeLive, buvid, ip, arg)
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("order", arg.Order)
	params.Set("from_spmid", arg.FromSpmid)
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	params.Set("dynamic_offset", strconv.FormatInt(arg.DynamicOffset, 10))
	for k, v := range search.SearchDefaultArg[search.SearchTypeLive] {
		params.Set(k, strconv.Itoa(v))
	}
	if arg.PageSize > 0 {
		params.Set("live_room_num", strconv.FormatInt(arg.PageSize, 10))
	}
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchTypeRes)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("SearchVideo d.httpSearch.Get(%s) error(%v)", d.searchURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchVideo d.httpSearch.Get(%s) code(%d) error", d.searchURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchLiveRoom search live data.
func (d *Dao) SearchLiveRoom(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchTypeRes, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchTypeParam(params, mid, search.SearchTypeLiveRoom, buvid, ip, arg)
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("order", arg.Order)
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	params.Set("dynamic_offset", strconv.FormatInt(arg.DynamicOffset, 10))
	for k, v := range search.SearchDefaultArg[search.SearchTypeLiveRoom] {
		params.Set(k, strconv.Itoa(v))
	}
	if arg.PageSize > 0 {
		params.Set("pagesize", strconv.FormatInt(arg.PageSize, 10))
	}
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchTypeRes)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("SearchLiveRoom d.httpSearch.Get(%s) error(%v)", d.searchURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchLiveRoom d.httpSearch.Get(%s) code(%d) error", d.searchURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchLiveUser search live user data.
func (d *Dao) SearchLiveUser(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchTypeRes, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchTypeParam(params, mid, search.SearchTypeLiveUser, buvid, ip, arg)
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("order", arg.Order)
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	params.Set("dynamic_offset", strconv.FormatInt(arg.DynamicOffset, 10))
	for k, v := range search.SearchDefaultArg[search.SearchTypeLiveUser] {
		params.Set(k, strconv.Itoa(v))
	}
	if arg.PageSize > 0 {
		params.Set("pagesize", strconv.FormatInt(arg.PageSize, 10))
	}
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchTypeRes)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("SearchVideo d.httpSearch.Get(%s) error(%v)", d.searchURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchVideo d.httpSearch.Get(%s) code(%d) error", d.searchURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchArticle search article.
func (d *Dao) SearchArticle(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchTypeRes, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchTypeParam(params, mid, search.SearchTypeArticle, buvid, ip, arg)
	params.Set("category_id", strconv.FormatInt(arg.CategoryID, 10))
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("order", arg.Order)
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	params.Set("dynamic_offset", strconv.FormatInt(arg.DynamicOffset, 10))
	for k, v := range search.SearchDefaultArg[search.SearchTypeArticle] {
		params.Set(k, strconv.Itoa(v))
	}
	if arg.PageSize > 0 {
		params.Set("pagesize", strconv.FormatInt(arg.PageSize, 10))
	}
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchTypeRes)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("SearchArticle d.httpSearch.Get(%s) error(%v)", d.searchURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchArticle d.httpSearch.Get(%s) code(%d) error", d.searchURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchSpecial search special data.
func (d *Dao) SearchSpecial(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchTypeRes, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchTypeParam(params, mid, search.SearchTypeSpecial, buvid, ip, arg)
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("vp_num", strconv.Itoa(arg.VpNum))
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	for k, v := range search.SearchDefaultArg[search.SearchTypeSpecial] {
		params.Set(k, strconv.Itoa(v))
	}
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchTypeRes)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("SearchSpecial d.httpSearch.Get(%s) error(%v)", d.searchURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchSpecial d.httpSearch.Get(%s) code(%d) error", d.searchURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchTopic search topic data.
func (d *Dao) SearchTopic(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchTypeRes, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchTypeParam(params, mid, search.SearchTypeTopic, buvid, ip, arg)
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	params.Set("dynamic_offset", strconv.FormatInt(arg.DynamicOffset, 10))
	for k, v := range search.SearchDefaultArg[search.SearchTypeTopic] {
		params.Set(k, strconv.Itoa(v))
	}
	if arg.Highlight > 0 {
		params.Set("highlight", strconv.Itoa(arg.Highlight))
	}
	if arg.PageSize > 0 {
		params.Set("pagesize", strconv.FormatInt(arg.PageSize, 10))
	}
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchTypeRes)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("SearchVideo d.httpSearch.Get(%s) error(%v)", d.searchURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchVideo d.httpSearch.Get(%s) code(%d) error", d.searchURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchUser search user data.
func (d *Dao) SearchUser(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchUserRes, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchTypeParam(params, mid, search.SearchTypeUser, buvid, ip, arg)
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("user_type", strconv.Itoa(arg.UserType))
	params.Set("bili_user_vl", strconv.Itoa(arg.BiliUserVl))
	params.Set("order_sort", strconv.Itoa(arg.OrderSort))
	params.Set("order", arg.Order)
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	params.Set("dynamic_offset", strconv.FormatInt(arg.DynamicOffset, 10))
	for k, v := range search.SearchDefaultArg[search.SearchTypeUser] {
		params.Set(k, strconv.Itoa(v))
	}
	if arg.PageSize > 0 {
		params.Set("pagesize", strconv.FormatInt(arg.PageSize, 10))
	}
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchUserRes)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("SearchUser d.httpSearch.Get(%s) error(%v)", d.searchURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchUser d.httpSearch.Get(%s) code(%d) error", d.searchURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchPhoto search photo data.
func (d *Dao) SearchPhoto(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchTypeRes, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchTypeParam(params, mid, search.SearchTypePhoto, buvid, ip, arg)
	params.Set("category_id", strconv.FormatInt(arg.CategoryID, 10))
	params.Set("page", strconv.Itoa(arg.Pn))
	params.Set("order", arg.Order)
	params.Set("is_inner", strconv.FormatInt(arg.IsInner, 10))
	for k, v := range search.SearchDefaultArg[search.SearchTypePhoto] {
		params.Set(k, strconv.Itoa(v))
	}
	if arg.Highlight > 0 {
		params.Set("highlight", strconv.Itoa(arg.Highlight))
	}
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchTypeRes)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("SearchPhoto d.httpSearch.Get(%s) error(%v)", d.searchURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchPhoto d.httpSearch.Get(%s) code(%d) error", d.searchURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchRec search recommend data.
func (d *Dao) SearchRec(c context.Context, mid int64, pn, ps int, keyword, fromSource, buvid, ua string) (res *search.SearchRec, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params = setSearchParam(params, mid, "", keyword, _searchWebPlatform, fromSource, buvid, ip, "")
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchRecURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	res = new(search.SearchRec)
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("Search d.httpSearch.Get(%s) error(%v)", d.searchRecURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("Search d.httpSearch.Do(%s) code error(%d)", d.searchRecURL, res.Code)
		err = ecode.Int(res.Code)
	}
	return
}

// SearchDefault get search default word.
func (d *Dao) SearchDefault(c context.Context, mid int64, fromSource, buvid, ua string) (data *search.SearchDefault, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("main_ver", _searchVer)
	params.Set("platform", _searchWebPlatform)
	params.Set("clientip", ip)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("search_type", "default")
	params.Set("from_source", fromSource)
	params.Set("buvid", buvid)
	// use new search default
	params.Set("is_new", "1")
	var req *http.Request
	if req, err = d.httpSearch.NewRequest(http.MethodGet, d.searchDefaultURL, ip, params); err != nil {
		return
	}
	req.Header.Set("browser-info", ua)
	var res struct {
		Code   int    `json:"code"`
		SeID   string `json:"seid"`
		Tips   string `json:"recommend_tips"`
		Result []struct {
			ID        int64  `json:"id"`
			Name      string `json:"name"`
			ShowName  string `json:"show_name"`
			Type      string `json:"type"`
			GotoType  int    `json:"goto_type"`
			GotoValue string `json:"goto_value"`
		} `json:"result"`
	}
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		log.Error("Search d.httpSearch.Get(%s) error(%v)", d.searchDefaultURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("Search d.httpSearch.Do(%s) code error(%d)", d.searchDefaultURL, res.Code)
		err = ecode.Int(res.Code)
		return
	}
	if len(res.Result) == 0 {
		err = ecode.NothingFound
		return
	}
	data = &search.SearchDefault{
		Trackid:   res.SeID,
		ID:        res.Result[0].ID,
		ShowName:  res.Result[0].ShowName,
		Name:      res.Result[0].Name,
		GotoType:  res.Result[0].GotoType,
		GotoValue: res.Result[0].GotoValue,
	}
	switch data.GotoType {
	case search.SearchDftGotoSearch:
		data.URL = model.FillURI(model.GotoSearch, data.Name, nil)
	case search.SearchDftGotoArchive:
		if aid, e := strconv.ParseInt(data.GotoValue, 10, 64); e == nil && aid > 0 {
			bvidStr, _ := bvid.AvToBv(aid)
			data.URL = model.FillURI(model.GotoBv, bvidStr, nil)
		}
	case search.SearchDftGotoArticle:
		data.URL = model.FillURI(model.GotoArticle, data.GotoValue, nil)
	case search.SearchDftGotoBangumi:
		data.URL = model.FillURI(model.GotoPGCSeason, data.GotoValue, nil)
	case search.SearchDftGotoURL:
		data.URL = model.FillURI(model.GotoURL, data.GotoValue, nil)
	}
	// 兼容小程序bug
	if fromSource == "xcx_search" {
		data.ShowName = ""
	}
	return
}

// UpRecommend .
func (d *Dao) UpRecommend(c context.Context, mid int64, arg *search.SearchUpRecArg) (rs []*search.SearchUpRecRes, trackID string, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("service_area", arg.ServiceArea)
	params.Set("rec_type", _searchUpRecType)
	params.Set("platform", arg.Platform)
	params.Set("clientip", ip)
	params.Set("pagesize", strconv.Itoa(arg.Ps))
	params.Set("buvid", arg.Buvid)
	if arg.MobiApp != "" {
		params.Set("mobi_app", arg.MobiApp)
	}
	if arg.Device != "" {
		params.Set("device", arg.Device)
	}
	if arg.Build != 0 {
		params.Set("build", strconv.FormatInt(arg.Build, 10))
	}
	if arg.ContextID != 0 {
		params.Set("context_id", strconv.FormatInt(arg.ContextID, 10))
	}
	if len(arg.MainTids) > 0 {
		params.Set("main_tids", xstr.JoinInts(arg.MainTids))
	}
	if len(arg.SubTids) > 0 {
		params.Set("sub_tids", xstr.JoinInts(arg.SubTids))
	}
	var res struct {
		Code    int                      `json:"code"`
		Trackid string                   `json:"trackid"`
		Data    []*search.SearchUpRecRes `json:"data"`
	}
	if err = d.httpSearch.Get(c, d.searchUpRecURL, ip, params, &res); err != nil {
		log.Error("UpRecommend d.httpSearch.Get(%s) error(%v)", d.searchUpRecURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("UpRecommend d.httpSearch.Do(%s) code error(%d)", d.searchUpRecURL, res.Code)
		err = ecode.Int(res.Code)
		return
	}
	rs = res.Data
	trackID = res.Trackid
	return
}

// SearchEgg search egg.
func (d *Dao) SearchEgg(c context.Context) (data []*search.SearchEgg, err error) {
	var (
		ip  = metadata.String(c, metadata.RemoteIP)
		res struct {
			Code int                 `json:"code"`
			Data []*search.SearchEgg `json:"data"`
		}
	)
	if err = d.httpSearch.Get(c, d.searchEggURL, ip, url.Values{}, &res); err != nil {
		log.Error("SearchEgg d.httpSearch.Get(%s) error(%v)", d.searchEggURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SearchEgg d.httpSearch.Do(%s) code error(%d)", d.searchEggURL, res.Code)
		err = ecode.Int(res.Code)
		return
	}
	data = res.Data
	return
}

func setSearchParam(param url.Values, mid int64, searchType, keyword, platform, fromSource, buvid, ip, mobiApp string) url.Values {
	param.Set("main_ver", _searchVer)
	if searchType != "" {
		param.Set("search_type", searchType)
	}
	param.Set("platform", platform)
	param.Set("keyword", keyword)
	param.Set("from_source", fromSource)
	param.Set("userid", strconv.FormatInt(mid, 10))
	param.Set("buvid", buvid)
	param.Set("clientip", ip)
	if mobiApp != "" {
		param.Set("mobi_app", mobiApp)
	}
	return param
}

func setSearchTypeParam(param url.Values, mid int64, searchType, buvid, ip string, arg *search.SearchTypeArg) url.Values {
	param.Set("main_ver", _searchVer)
	if searchType != "" {
		param.Set("search_type", searchType)
	}
	param.Set("platform", arg.Platform)
	param.Set("keyword", arg.Keyword)
	param.Set("from_source", arg.FromSource)
	param.Set("userid", strconv.FormatInt(mid, 10))
	param.Set("single_column", strconv.Itoa(arg.SingleColumn))
	param.Set("buvid", buvid)
	param.Set("clientip", ip)
	if arg.MobiApp != "" {
		param.Set("mobi_app", arg.MobiApp)
	}
	return param
}

func (d *Dao) Trending(c context.Context, buvid string, mid int64, limit int, isInner, zoneID int64, platform string) (*search.Hot, error) {
	params := url.Values{}
	params.Set("main_ver", "v3")
	params.Set("actionKey", "appkey")
	params.Set("limit", strconv.Itoa(limit))
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("is_inner", strconv.FormatInt(isInner, 10))
	params.Set("zone_id", strconv.FormatInt(zoneID, 10))
	params.Set("platform", platform)
	req, err := d.httpSearch.NewRequest("GET", d.trending, "", params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Buvid", buvid)
	var res *search.Hot
	if err = d.httpSearch.Do(c, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.trending+"?"+params.Encode())
		return nil, err
	}
	return res, nil
}

func (d *Dao) SearchSystemNotice(ctx context.Context) (map[int64]*model.SystemNotice, error) {
	reply := &struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List []*model.SystemNotice `json:"list"`
		} `json:"data"`
	}{}
	if err := d.httpR.Get(ctx, d.c.Host.Manager+"/x/admin/manager/search/internal/system/notice", "", nil, reply); err != nil {
		return nil, err
	}
	if reply.Code != 0 {
		return nil, errors.Errorf("invalid code: %d: %+v", reply.Code, reply)
	}
	out := make(map[int64]*model.SystemNotice, len(reply.Data.List))
	for _, v := range reply.Data.List {
		if v == nil {
			continue
		}
		v.Construct()
		out[v.Mid] = v
	}
	return out, nil
}
