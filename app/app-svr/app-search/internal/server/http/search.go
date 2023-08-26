package http

import (
	"strconv"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	"go-gateway/app/app-svr/app-search/internal/model"
	"go-gateway/app/app-svr/app-search/internal/model/search"
)

const (
	_headerBuvid = "Buvid"
	_keyWordLen  = 50
)

func searchAll(c *bm.Context) {
	var (
		build  int
		mid    int64
		pn, ps int
		data   *search.Result
		code   int
		err    error
	)
	params := c.Request.Form
	header := c.Request.Header
	// params
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	ridStr := params.Get("rid")
	keyword := params.Get("keyword")
	highlightStr := params.Get("highlight")
	lang := params.Get("lang")
	duration := params.Get("duration")
	order := params.Get("order")
	filtered := params.Get("filtered")
	platform := params.Get("platform")
	fromSource := params.Get("from_source")
	recommend := params.Get("recommend")
	parent := params.Get("parent_mode")
	adExtra := params.Get("ad_extra")
	extraWord := params.Get("extra_word")
	tidList := params.Get("tid_list")
	durationList := params.Get("duration_list")
	qvid := params.Get("qv_id")
	// header
	buvid := header.Get("Buvid")
	// check params
	if keyword == "" || len([]rune(keyword)) > _keyWordLen {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	rid, _ := strconv.Atoi(ridStr)
	highlight, _ := strconv.Atoi(highlightStr)
	// page and size
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	switch order {
	case "default", "":
		order = "totalrank"
	case "view":
		order = "click"
	case "danmaku":
		order = "dm"
	}
	if duration == "" {
		duration = "0"
	}
	if recommend == "" || recommend != "1" {
		recommend = "0"
	}
	isQuery, _ := strconv.Atoi(params.Get("is_org_query"))
	plat := model.Plat(mobiApp, device)
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	qn, _ := strconv.ParseInt(params.Get("qn"), 10, 64)
	fnver, _ := strconv.ParseInt(params.Get("fnver"), 10, 64)
	fnval, _ := strconv.ParseInt(params.Get("fnval"), 10, 64)
	fourk, _ := strconv.ParseInt(params.Get("fourk"), 10, 64)
	localTime, _ := strconv.ParseInt(params.Get("local_time"), 10, 64)
	autoPlayCard, _ := strconv.ParseInt(params.Get("auto_playcard"), 10, 64)
	// 兼容UTC-14和Etc/GMT+12,时区区间[-12,14]
	if localTime < -12 || localTime > 14 {
		localTime = 8
	}
	if data, code, err = srcSvr.Search(c, mid, mobiApp, device, platform, buvid, keyword, duration, order,
		filtered, lang, fromSource, recommend, parent, adExtra, extraWord, tidList, durationList, qvid, plat, rid, highlight, build, pn, ps, isQuery, teenagersMode, lessonsMode, qn,
		fnver, fnval, fourk, checkOld(plat, build), time.Now(), localTime, autoPlayCard); err != nil {
		if code != model.ForbidCode && code != model.NoResultCode {
			c.JSON(nil, ecode.Degrade)
			return
		}
	}
	c.JSON(data, err)
}

func searchByType(c *bm.Context) {
	var (
		build  int
		mid    int64
		pn, ps int
		typeV  string
		data   *search.TypeSearch
		code   int
		err    error
	)
	params := c.Request.Form
	header := c.Request.Header
	// params
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	sType := params.Get("type")
	keyword := params.Get("keyword")
	filtered := params.Get("filtered")
	order := params.Get("order")
	platform := params.Get("platform")
	highlightStr := params.Get("highlight")
	categoryIDStr := params.Get("category_id")
	userTypeStr := params.Get("user_type")
	orderSortStr := params.Get("order_sort")
	fourk, _ := strconv.ParseInt(params.Get("fourk"), 10, 64)
	qvid := params.Get("qv_id")
	// header
	buvid := header.Get("Buvid")
	if keyword == "" || len([]rune(keyword)) > _keyWordLen {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	userType, _ := strconv.Atoi(userTypeStr)
	orderSort, _ := strconv.Atoi(orderSortStr)
	categoryID, _ := strconv.Atoi(categoryIDStr)
	highlight, _ := strconv.Atoi(highlightStr)
	// page and size
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	switch sType {
	case "1":
		typeV = "season"
	case "2":
		typeV = "upper"
	case "3":
		typeV = "movie"
	case "4":
		typeV = "live_room"
	case "5":
		typeV = "live_user"
	case "6":
		typeV = "article"
	case "7":
		typeV = "season2"
	case "8":
		typeV = "movie2"
	case "9":
		typeV = "tag"
	case "10":
		typeV = "video"
	}
	plat := model.Plat(mobiApp, device)
	qn, _ := strconv.ParseInt(params.Get("qn"), 10, 64)
	fnver, _ := strconv.ParseInt(params.Get("fnver"), 10, 64)
	fnval, _ := strconv.ParseInt(params.Get("fnval"), 10, 64)
	if data, code, err = srcSvr.SearchByType(c, mid, mobiApp, device, platform, buvid, typeV, keyword, filtered, order, qvid, plat, build, highlight, categoryID, userType, orderSort, pn, ps, fnver, fnval, qn, fourk, checkOld(plat, build), time.Now()); err != nil {
		if code != model.ForbidCode && code != model.NoResultCode {
			c.JSON(nil, ecode.Degrade)
			return
		}
	}
	c.JSON(data, err)
}

func searchLive(c *bm.Context) {
	var (
		build  int
		mid    int64
		pn, ps int
		typeV  string
		err    error
	)
	params := c.Request.Form
	header := c.Request.Header
	// params
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	sType := params.Get("type")
	keyword := params.Get("keyword")
	order := params.Get("order")
	platform := params.Get("platform")
	qvid := params.Get("qv_id")
	// header
	buvid := header.Get("Buvid")
	if keyword == "" || len([]rune(keyword)) > _keyWordLen {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	// page and size
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	plat := model.Plat(mobiApp, device)
	// 蓝版里的直播相关功能不参与迭代
	//if !cdm.ShowLiveV2(c, config.Feature.FeatureBuildLimit.ShowLive, nil) {
	//	c.JSON(nil, nil)
	//	return
	//}
	switch sType {
	case "4":
		if (model.IsAndroid(plat) && build > search.SearchLiveAllAndroid) || (model.IsIPhone(plat) && build > search.SearchLiveAllIOS) || model.IsPad(plat) || model.IsIPhoneB(plat) {
			typeV = "live_all"
		} else {
			typeV = "live_room"
		}
	case "5":
		typeV = "live_user"
	}
	if typeV == "live_all" {
		c.JSON(srcSvr.SearchLiveAll(c, mid, mobiApp, platform, buvid, device, typeV, keyword, order, build, pn, ps))
	} else {
		c.JSON(srcSvr.SearchLive(c, mid, mobiApp, platform, buvid, device, typeV, keyword, order, qvid, build, pn, ps))
	}
}

// ip string, limit int
func hotSearch(c *bm.Context) {
	var (
		mid   int64
		build int
		limit int
		err   error
	)
	params := c.Request.Form
	header := c.Request.Header
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	platform := params.Get("platform")
	buvid := header.Get("Buvid")
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if limit, err = strconv.Atoi(params.Get("limit")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	if teenagersMode != 0 || lessonsMode != 0 {
		c.JSON(&search.Hot{}, nil)
	} else {
		data := srcSvr.HotSearch(c, buvid, mid, build, limit, mobiApp, device, platform, time.Now())
		if data == nil || data.Code != 0 || len(data.List) == 0 {
			c.JSON(&search.Hot{}, ecode.Degrade)
			return
		}
		c.JSON(data, nil)
	}
}

func trending(c *bm.Context) {
	var (
		mid   int64
		build int
		limit int
		err   error
	)
	params := c.Request.Form
	header := c.Request.Header
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	platform := params.Get("platform")
	buvid := header.Get("Buvid")
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if limit, err = strconv.Atoi(params.Get("limit")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	if teenagersMode != 0 || lessonsMode != 0 {
		c.JSON(&search.Hot{}, nil)
	} else {
		data := srcSvr.Trending(c, buvid, mid, build, limit, mobiApp, device, platform, time.Now())
		if data == nil || data.Code != 0 || len(data.List) == 0 {
			// 热搜降级条件
			c.JSON(&search.Hot{}, ecode.Degrade)
			return
		}
		c.JSON(data, nil)
	}
}

func ranking(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(search.TrendingRankingReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	data, _ := srcSvr.Ranking(ctx, mid, params)
	if data == nil || data.Code != 0 || len(data.List) == 0 {
		ctx.JSON(nil, ecode.Degrade)
		return
	}
	ctx.JSON(data, nil)
}

// suggest search suggest data.
func suggest(c *bm.Context) {
	var (
		build int
		mid   int64
		err   error
	)
	params := c.Request.Form
	header := c.Request.Header
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	term := params.Get("keyword")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	buvid := header.Get(_headerBuvid)
	c.JSON(srcSvr.Suggest(c, mid, buvid, term, build, mobiApp, device, time.Now()), nil)
}

// suggest2 search suggest data from new api.
func suggest2(c *bm.Context) {
	var (
		build int
		mid   int64
		err   error
	)
	params := c.Request.Form
	header := c.Request.Header
	mobiApp := params.Get("mobi_app")
	term := params.Get("keyword")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	buvid := header.Get(_headerBuvid)
	platform := params.Get("platform")
	device := params.Get("device")
	c.JSON(srcSvr.Suggest2(c, mid, platform, buvid, term, build, mobiApp, time.Now(), device), nil)
}

// suggest3 search suggest data from newest api.
func suggest3(c *bm.Context) {
	var (
		build int
		mid   int64
		err   error
	)
	params := c.Request.Form
	header := c.Request.Header
	mobiApp := params.Get("mobi_app")
	term := params.Get("keyword")
	device := params.Get("device")
	highlight, _ := strconv.Atoi(params.Get("highlight"))
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	buvid := header.Get(_headerBuvid)
	platform := params.Get("platform")
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	if teenagersMode != 0 || lessonsMode != 0 {
		c.JSON(&search.SuggestionResult3{}, nil)
	} else {
		c.JSON(srcSvr.Suggest3Json(c, mid, platform, buvid, term, device, build, highlight, mobiApp, time.Now()), nil)
	}
}

func checkOld(plat int8, build int) bool {
	const (
		_oldAndroid = 513000
		_oldIphone  = 6060
	)
	return (model.IsIPhone(plat) && build <= _oldIphone) || (model.IsAndroid(plat) && build <= _oldAndroid)
}

func searchUser(c *bm.Context) {
	var (
		build int
		mid   int64
		err   error
	)
	params := c.Request.Form
	header := c.Request.Header
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	platform := params.Get("platform")
	keyword := params.Get("keyword")
	filtered := params.Get("filtered")
	order := params.Get("order")
	fromSource := params.Get("from_source")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	userType, _ := strconv.Atoi(params.Get("user_type"))
	highlight, _ := strconv.Atoi(params.Get("highlight"))
	if order == "" {
		order = "totalrank"
	}
	if order != "totalrank" && order != "fans" && order != "level" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	orderSort, _ := strconv.Atoi(params.Get("order_sort"))
	if orderSort != 1 {
		orderSort = 0
	}
	if fromSource == "" {
		fromSource = "dynamic_uname"
	}
	if fromSource != "dynamic_uname" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	pn, _ := strconv.Atoi(params.Get("pn"))
	if pn < 1 {
		pn = 1
	}
	ps, _ := strconv.Atoi(params.Get("ps"))
	if ps < 1 || ps > 20 {
		ps = 20
	}
	buvid := header.Get(_headerBuvid)
	c.JSON(srcSvr.User(c, mid, buvid, mobiApp, device, platform, keyword, filtered, order, fromSource, highlight, build, userType, orderSort, pn, ps, time.Now()), nil)
}

func recommend(c *bm.Context) {
	var (
		build int
		mid   int64
		err   error
	)
	params := c.Request.Form
	header := c.Request.Header
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	platform := params.Get("platform")
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}

	from, _ := strconv.Atoi(params.Get("from"))
	show, _ := strconv.Atoi(params.Get("show"))
	buvid := header.Get("Buvid")
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	disableRcmd, _ := strconv.Atoi(params.Get("disable_rcmd"))
	if teenagersMode != 0 || lessonsMode != 0 {
		c.JSON(&search.RecommendResult{}, nil)
	} else {
		c.JSON(srcSvr.Recommend(c, mid, build, from, show, disableRcmd, buvid, platform, mobiApp, device))
	}
}

func defaultWords(c *bm.Context) {
	var (
		build int
		mid   int64
		err   error
	)
	params := c.Request.Form
	header := c.Request.Header
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	platform := params.Get("platform")
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	from, _ := strconv.Atoi(params.Get("from"))
	buvid := header.Get("Buvid")
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	loginEvent, _ := strconv.ParseInt(params.Get("login_event"), 10, 64)
	if teenagersMode != 0 || lessonsMode != 0 {
		c.JSON(&search.DefaultWords{}, nil)
	} else {
		c.JSON(srcSvr.DefaultWordsJson(c, mid, build, from, buvid, platform, mobiApp, device, loginEvent, nil))
	}
}

func recommendNoResult(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
		build  int
		mid    int64
		err    error
	)
	platform := params.Get("platform")
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	buvid := header.Get("Buvid")
	keyword := params.Get("keyword")
	if keyword == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	pn, _ := strconv.Atoi(params.Get("pn"))
	if pn < 1 {
		pn = 1
	}
	ps, _ := strconv.Atoi(params.Get("ps"))
	if ps < 1 || ps > 20 {
		ps = 20
	}
	c.JSON(srcSvr.RecommendNoResult(c, platform, mobiApp, device, buvid, keyword, build, pn, ps, mid))
}

func resource(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
		build  int
		mid    int64
		err    error
	)
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	network := params.Get("network")
	buvid := header.Get("Buvid")
	adExtra := params.Get("ad_extra")
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	if teenagersMode != 0 || lessonsMode != 0 {
		c.JSON([]struct{}{}, nil)
	} else {
		c.JSON(srcSvr.Resource(c, mobiApp, device, network, buvid, adExtra, build, plat, mid))
	}
}

func recommendTags(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(search.RecommendTagsReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(srcSvr.RecommendTags(ctx, mid, params))
}

func recommendPre(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
		build  int
		mid    int64
		err    error
	)
	platform := params.Get("platform")
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	buvid := header.Get("Buvid")
	ps, _ := strconv.Atoi(params.Get("ps"))
	if ps < 1 || ps > 20 {
		ps = 20
	}
	c.JSON(srcSvr.RecommendPre(c, platform, mobiApp, device, buvid, build, ps, mid))
}

func searchEpisodes(c *bm.Context) {
	var (
		params    = c.Request.Form
		mid, ssID int64
		err       error
	)
	if ssID, err = strconv.ParseInt(params.Get("season_id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if ssID == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(srcSvr.SearchEpisodes(c, mid, ssID))
}

func searchEpisodesNew(c *bm.Context) {
	var (
		param = new(search.EpisodesNewReq)
		err   error
	)
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(srcSvr.SearchEpsNew(c, param))
}

func searchConverge(c *bm.Context) {
	var (
		params        = c.Request.Form
		header        = c.Request.Header
		build, ps, pn int
		mid, cid      int64
		sort          string
		err           error
	)
	// params
	if cid, err = strconv.ParseInt(params.Get("card_id"), 10, 44); err != nil || cid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	trackID := params.Get("track_id")
	platform := params.Get("platform")
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	// header
	buvid := header.Get("Buvid")
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	order := params.Get("order")
	switch order {
	case "new":
		order = "pubdate"
		sort = "0"
	case "hot":
		order = "click"
		sort = "0"
	case "initial":
		order = "pubdate"
		sort = "1"
	default:
		order = "pubdate"
		sort = "0"
	}
	// page and size
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	//nolint:gomnd
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 {
		ps = 20
		//nolint:gomnd
	} else if ps > 100 {
		ps = 100
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(srcSvr.SearchConverge(c, mid, cid, trackID, platform, mobiApp, device, buvid, order, sort, plat, build, pn, ps))
}

func searchChannel(c *bm.Context) {
	var (
		params        = c.Request.Form
		header        = c.Request.Header
		mid           int64
		build, pn, ps int
		err           error
	)
	keyword := params.Get("keyword")
	if keyword == "" || len([]rune(keyword)) > _keyWordLen {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	platform := params.Get("platform")
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	// header
	buvid := header.Get("Buvid")
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	// page and size
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	plat := model.Plat(mobiApp, device)
	highlight, _ := strconv.Atoi(params.Get("highlight"))
	c.JSON(srcSvr.SearchChannel(c, keyword, platform, mobiApp, device, buvid, plat, build, pn, ps, highlight, mid))
}

func searchSquare(c *bm.Context) {
	var (
		mid   int64
		build int
		limit int
		err   error
	)
	params := c.Request.Form
	header := c.Request.Header
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	network := params.Get("network")
	platform := params.Get("platform")
	adExtra := params.Get("ad_extra")
	buvid := header.Get("Buvid")
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if limit, err = strconv.Atoi(params.Get("limit")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	from, _ := strconv.Atoi(params.Get("from"))
	show, _ := strconv.Atoi(params.Get("show"))
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	slocaleP := params.Get("s_locale")
	cLocaleP := params.Get("c_locale")
	disableRcmd, _ := strconv.Atoi(params.Get("disable_rcmd"))
	isHant := i18n.PreferTraditionalChinese(c, slocaleP, cLocaleP)
	if teenagersMode != 0 || lessonsMode != 0 {
		c.JSON(&search.IterationConverge{}, nil)
	} else {
		data, err := srcSvr.Square(c, mid, mobiApp, device, network, platform, adExtra, buvid, build, limit, from, show, disableRcmd, time.Now(), isHant)
		if err != nil {
			c.JSON(data, ecode.Degrade)
			return
		}
		c.JSON(data, nil)
	}
}

func searchChannel2(c *bm.Context) {
	var (
		params = &search.Param{}
		data   *search.ChannelResult
		err    error
	)
	if err := c.Bind(params); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.MID = midInter.(int64)
	}
	params.Buvid = c.Request.Header.Get("Buvid")
	keyword := params.Keyword
	if keyword == "" || len([]rune(keyword)) > _keyWordLen {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	// page and size
	if params.PN < 1 {
		params.PN = 1
	}
	if params.PS < 1 || params.PS > 20 {
		params.PS = 20
	}
	params.Plat = model.Plat(params.MobiApp, params.Device)
	if params.Spmid == "" {
		params.Spmid = "traffic.discovery-channel-tab.0.0"
	}
	if data, err = srcSvr.SearchChannel2(c, params); err != nil {
		c.JSON(nil, ecode.Degrade)
		return
	}
	c.JSON(data, err)
}

func searchSiri(ctx *bm.Context) {
	req := &search.SiriCommandReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if midInter, ok := ctx.Get("mid"); ok {
		req.Mid = midInter.(int64)
	}
	req.Buvid = ctx.Request.Header.Get("Buvid")
	ctx.JSON(srcSvr.ResolveCommand(ctx, req))
}
