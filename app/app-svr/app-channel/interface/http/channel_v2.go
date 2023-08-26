package http

import (
	"net/http"
	"strconv"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-channel/interface/model"
	chmdl2 "go-gateway/app/app-svr/app-channel/interface/model/channel_v2"
)

func tab2(c *bm.Context) {
	c.JSON(channelSvcV2.Tab(c))
}

func tab3(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(channelSvcV2.Tab3(c, mid))
}

func list2(c *bm.Context) {
	var (
		params                  = c.Request.Form
		mid, ctype              int64
		build                   int
		plat                    int8
		mobiApp, device, offset string
		err                     error
	)
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	device = params.Get("device")
	plat = model.Plat(mobiApp, device)
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	ctype, _ = strconv.ParseInt(params.Get("type"), 10, 32)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	offset = params.Get("offset")
	spmid := params.Get("spmid")
	if spmid == "" {
		spmid = "traffic.discovery-channel-tab.0.0"
	}
	data, err := channelSvcV2.ChannelList(c, mid, int32(ctype), offset, plat, build, mobiApp, device, spmid)
	if err != nil {
		if status := bm.ErrorHTTPStatus(err); status != http.StatusOK {
			c.AbortWithStatus(status)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)
}

func mine(c *bm.Context) {
	var (
		params          = c.Request.Form
		mobiApp, device string
		plat            int8
		build           int
		mid             int64
		err             error
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	spmid := params.Get("spmid")
	if spmid == "" {
		spmid = "traffic.discovery-channel-tab.0.0"
	}
	device = params.Get("device")
	plat = model.Plat(mobiApp, device)
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(channelSvcV2.Mine(c, plat, build, mid, mobiApp, spmid))
}

func channelSort(c *bm.Context) {
	var (
		params        = c.Request.Form
		mid, action   int64
		stick, normal string
		err           error
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	stick = params.Get("stick")
	normal = params.Get("normal")
	if stick == "" && normal == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if action, err = strconv.ParseInt(params.Get("action"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, channelSvcV2.ChannelSort(c, mid, int32(action), stick, normal))
}

func square2(c *bm.Context) {
	var (
		params                                  = c.Request.Form
		mid                                     int64
		build, teenagersMode, autoRefresh, pn   int
		plat                                    int8
		platform, mobiApp, device, lang, offset string
		err                                     error
	)
	if platform = params.Get("platform"); platform == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if offset = params.Get("offset"); offset == "0" {
		offset = ""
	}
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	device = params.Get("device")
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat = model.Plat(mobiApp, device)
	teenagersMode, _ = strconv.Atoi(params.Get("teenagers_mode"))
	lang = params.Get("lang")
	if lang == "" {
		lang = "hans"
	}
	fromSpmid := params.Get("spmid")
	buvid := c.Request.Header.Get("Buvid")
	reqURL := c.Request.URL.String()
	// nolint:gomnd
	timeIso := time.Now().UnixNano() / 1e6
	statistics := params.Get("statistics")
	ts, _ := strconv.ParseInt(params.Get("ts"), 10, 64)
	if autoRefresh, _ = strconv.Atoi(params.Get("auto_refresh")); autoRefresh != 0 && autoRefresh != 1 {
		autoRefresh = 0
	}
	pn, _ = strconv.Atoi(params.Get("pn"))
	pn++
	paramChannel := params.Get("channel")
	c.JSON(channelSvcV2.Square(c, channelSvc, mid, timeIso, ts, build, teenagersMode, autoRefresh,
		plat, platform, mobiApp, device, lang, offset, buvid, fromSpmid, reqURL, statistics, paramChannel, pn))
}

func squareAlpha(c *bm.Context) {
	var (
		params                           = c.Request.Form
		platform, mobiApp, device, buvid string
		plat                             int8
		build, autoRefresh               int
		mid                              int64
		err                              error
	)
	if platform = params.Get("platform"); platform == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	device = params.Get("device")
	buvid = c.Request.Header.Get("Buvid")
	plat = model.Plat(mobiApp, device)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if autoRefresh, _ = strconv.Atoi(params.Get("auto_refresh")); autoRefresh != 0 && autoRefresh != 1 {
		autoRefresh = 0
	}
	// nolint:gomnd
	timeIso := time.Now().UnixNano() / 1e6
	ts, _ := strconv.ParseInt(params.Get("ts"), 10, 64)
	fromSpmid := params.Get("spmid")
	reqURL := c.Request.URL.String()
	statistics := params.Get("statistics")
	c.JSON(channelSvcV2.SquareAlpha(c, platform, mobiApp, device, buvid, plat, autoRefresh, build,
		mid, ts, timeIso, fromSpmid, reqURL, statistics))
}

func errorHTTPStatus(err error) int {
	switch ecode.Cause(err).Code() {
	case ecode.ServerErr.Code(), ecode.LimitExceed.Code(), ecode.BusinessDegrade.Code():
		return http.StatusInternalServerError
	case ecode.Deadline.Code():
		return http.StatusGatewayTimeout
	case ecode.ServiceUnavailable.Code():
		return http.StatusServiceUnavailable
	default:
		return http.StatusOK
	}
}

func detail(c *bm.Context) {
	var (
		params                = c.Request.Form
		mid, channelID, build int64
		mobiApp, device       string
		plat                  int8
		err                   error
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if channelID, err = strconv.ParseInt(params.Get("channel_id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if build, err = strconv.ParseInt(params.Get("build"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	spmid := params.Get("spmid")
	if spmid == "" {
		spmid = "traffic.new-channel-detail.0.0"
	}
	device = params.Get("device")
	plat = model.Plat(mobiApp, device)
	platform := params.Get("platform")
	externalArg := &chmdl2.ChanelDetailExternalArgs{Args: map[string]string{"source": spmid}}
	if err = c.Bind(externalArg); err != nil {
		return
	}
	reply, err := channelSvcV2.Detail(c, mid, channelID, plat, build, mobiApp, spmid, platform, externalArg)
	if err != nil {
		if status := errorHTTPStatus(err); status != http.StatusOK {
			c.AbortWithStatus(status)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(reply, nil)
}

func multiple(c *bm.Context) {
	var (
		params                                  = c.Request.Form
		channelID, mid                          int64
		build, pn                               int
		plat                                    int8
		platform, mobiApp, device, sort, offset string
		err                                     error
	)
	if platform = params.Get("platform"); platform == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if channelID, err = strconv.ParseInt(params.Get("channel_id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	device = params.Get("device")
	plat = model.Plat(mobiApp, device)
	sort = params.Get("sort")
	if sort == "" || (sort != "hot" && sort != "view" && sort != "new") {
		sort = "hot"
	}
	if offset = params.Get("offset"); offset == "0" {
		offset = ""
	}
	fromSpmid := params.Get("spmid")
	buvid := c.Request.Header.Get("Buvid")
	reqURL := c.Request.URL.String()
	// nolint:gomnd
	timeIso := time.Now().UnixNano() / 1e6
	statistics := params.Get("statistics")
	ts, _ := strconv.ParseInt(params.Get("ts"), 10, 64)
	theme := params.Get("theme")
	pn, _ = strconv.Atoi(params.Get("pn"))
	pn++
	from := params.Get("from")
	data, err := channelSvcV2.Multiple(c, channelID, mid, timeIso, ts, build, plat, platform, sort, offset, device, fromSpmid, buvid, reqURL, statistics, theme, from, mobiApp, pn)
	if err != nil {
		if status := bm.ErrorHTTPStatus(err); status != http.StatusOK {
			c.AbortWithStatus(status)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)
}

func selected(c *bm.Context) {
	var (
		params                            = c.Request.Form
		channelID, mid, cFilter           int64
		build, pn                         int
		plat                              int8
		platform, mobiApp, device, offset string
		err                               error
	)
	if platform = params.Get("platform"); platform == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if channelID, err = strconv.ParseInt(params.Get("channel_id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	device = params.Get("device")
	plat = model.Plat(mobiApp, device)
	cFilter, _ = strconv.ParseInt(params.Get("sort"), 10, 64)
	if offset = params.Get("offset"); offset == "0" {
		offset = ""
	}
	fromSpmid := params.Get("spmid")
	buvid := c.Request.Header.Get("Buvid")
	reqURL := c.Request.URL.String()
	// nolint:gomnd
	timeIso := time.Now().UnixNano() / 1e6
	statistics := params.Get("statistics")
	ts, _ := strconv.ParseInt(params.Get("ts"), 10, 64)
	theme := params.Get("theme")
	pn, _ = strconv.Atoi(params.Get("pn"))
	pn++
	from := params.Get("from")

	data, err := channelSvcV2.Selected(c, channelID, mid, timeIso, ts, int32(cFilter), build, plat, platform, offset, device, fromSpmid, buvid, reqURL, statistics, theme, from, mobiApp, pn)
	if err != nil {
		if status := bm.ErrorHTTPStatus(err); status != http.StatusOK {
			c.AbortWithStatus(status)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)
}

func rankList(c *bm.Context) {
	var (
		params = c.Request.Form
		id     int64
		offset string
		ps     int
		err    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if offset = params.Get("offset"); offset == "" {
		offset = "0"
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	c.JSON(channelSvcV2.RankList(c, id, offset, ps))
}

func share(c *bm.Context) {
	var (
		params  = c.Request.Form
		id, mid int64
		err     error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(channelSvcV2.Share(c, id, mid))
}

func red(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(channelSvcV2.Red(c, mid))
}

func channelRcmd(c *bm.Context) {
	var (
		params                           = c.Request.Form
		platform, mobiApp, device, buvid string
		plat                             int8
		build, autoRefresh               int
		mid                              int64
		err                              error
	)
	if platform = params.Get("platform"); platform == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	device = params.Get("device")
	buvid = c.Request.Header.Get("Buvid")
	plat = model.Plat(mobiApp, device)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if autoRefresh, _ = strconv.Atoi(params.Get("auto_refresh")); autoRefresh != 0 && autoRefresh != 1 {
		autoRefresh = 0
	}
	// nolint:gomnd
	timeIso := time.Now().UnixNano() / 1e6
	ts, _ := strconv.ParseInt(params.Get("ts"), 10, 64)
	fromSpmid := params.Get("spmid")
	reqURL := c.Request.URL.String()
	statistics := params.Get("statistics")
	c.JSON(channelSvcV2.Rcmd(c, platform, mobiApp, device, buvid, plat, autoRefresh, build, mid, ts, timeIso, fromSpmid, reqURL, statistics))
}

func regionList(c *bm.Context) {
	params := &chmdl2.Param{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(channelSvcV2.RegionList(c, channelSvc, params))
}

func square3(c *bm.Context) {
	var params = &chmdl2.Param{}
	if err := c.Bind(params); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.MID = midInter.(int64)
	}
	params.Buvid = c.Request.Header.Get("Buvid")
	params.ReqURL = c.Request.URL.String()
	// nolint:gomnd
	params.TimeIso = time.Now().UnixNano() / 1e6
	data, err := channelSvcV2.Square3(c, params)
	if err != nil {
		if status := bm.ErrorHTTPStatus(err); status != http.StatusOK {
			c.AbortWithStatus(status)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)
}

func channelRcmd2(c *bm.Context) {
	var params = &chmdl2.Param{}
	if err := c.Bind(params); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.MID = midInter.(int64)
	}
	params.Buvid = c.Request.Header.Get("Buvid")
	params.ReqURL = c.Request.URL.String()
	// nolint:gomnd
	params.TimeIso = time.Now().UnixNano() / 1e6
	c.JSON(channelSvcV2.Rcmd2(c, params))
}

func home(c *bm.Context) {
	var params = &chmdl2.Param{}
	if err := c.Bind(params); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.MID = midInter.(int64)
	}
	params.Buvid = c.Request.Header.Get("Buvid")
	params.ReqURL = c.Request.URL.String()
	// nolint:gomnd
	params.TimeIso = time.Now().UnixNano() / 1e6
	data, err := channelSvcV2.Home(c, params)
	if err != nil {
		if status := bm.ErrorHTTPStatus(err); status != http.StatusOK {
			c.AbortWithStatus(status)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)
}

func home2(c *bm.Context) {
	var params = &chmdl2.Param{}
	if err := c.Bind(params); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.MID = midInter.(int64)
	}
	params.Buvid = c.Request.Header.Get("Buvid")
	params.ReqURL = c.Request.URL.String()
	// nolint:gomnd
	params.TimeIso = time.Now().UnixNano() / 1e6
	data, err := channelSvcV2.Home2(c, params)
	if err != nil {
		if status := bm.ErrorHTTPStatus(err); status != http.StatusOK {
			c.AbortWithStatus(status)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)
}
