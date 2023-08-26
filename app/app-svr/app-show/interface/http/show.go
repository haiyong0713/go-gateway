package http

import (
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-show/interface/model"
	popularmodel "go-gateway/app/app-svr/app-show/interface/model/popular"
	"go-gateway/app/app-svr/app-show/interface/model/selected"
	"go-gateway/app/app-svr/app-show/interface/model/show"
)

const (
	_headerBuvid     = "Buvid"
	_cookieBuvid3    = "buvid3"
	_headerTfIsp     = "X-Tf-Isp"
	_weeklySelected  = "weekly_selected"
	_headerUserAgent = "user-agent"
)

func shows(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
	)
	// get params
	mobiApp := params.Get("mobi_app")
	mobiApp = model.MobiAPPBuleChange(mobiApp)
	buildStr := params.Get("build")
	channel := params.Get("channel")
	ak := params.Get("access_key")
	// check params
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		log.Error("strconv.Atoi(%s) error(%v)", buildStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	// get audit data, if check audit hit.
	ss, ok := showSvc.Audit(c, mobiApp, plat, build, mid, device)
	if ok {
		returnJSON(c, ss, nil)
		return
	}
	network := params.Get("network")
	ip := metadata.String(c, metadata.RemoteIP)
	buvid := header.Get(_headerBuvid)
	// display
	ss = showSvc.Display(c, mid, plat, build, buvid, channel, ip, ak, network, mobiApp, device, "hans", "", false, time.Now())
	returnJSON(c, ss, nil)
	// infoc
	if len(ss) == 0 {
		return
	}
	showSvc.Infoc(mid, plat, buvid, ip, "/x/v2/show", ss[0].Body, time.Now())
}

func showsRegion(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
	)
	// get params
	mobiApp := params.Get("mobi_app")
	mobiApp = model.MobiAPPBuleChange(mobiApp)
	buildStr := params.Get("build")
	channel := params.Get("channel")
	ak := params.Get("access_key")
	// check params
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		log.Error("strconv.Atoi(%s) error(%v)", buildStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	// get audit data, if check audit hit.
	ss, ok := showSvc.Audit(c, mobiApp, plat, build, mid, device)
	if !ok {
		ip := metadata.String(c, metadata.RemoteIP)
		buvid := header.Get(_headerBuvid)
		network := params.Get("network")
		// display
		language := params.Get("lang")
		ss = showSvc.RegionDisplay(c, mid, plat, build, buvid, channel, ip, ak, network, mobiApp, device, language, "", false, time.Now())
	}
	res := map[string]interface{}{
		"data": ss,
	}
	returnDataJSON(c, res, 25, nil)
}

func showsIndex(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
	)
	// get params
	mobiApp := params.Get("mobi_app")
	mobiApp = model.MobiAPPBuleChange(mobiApp)
	buildStr := params.Get("build")
	channel := params.Get("channel")
	ak := params.Get("access_key")
	// check params
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		log.Error("strconv.Atoi(%s) error(%v)", buildStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	// get audit data, if check audit hit.
	ss, ok := showSvc.Audit(c, mobiApp, plat, build, mid, device)
	if !ok {
		ip := metadata.String(c, metadata.RemoteIP)
		buvid := header.Get(_headerBuvid)
		network := params.Get("network")
		// display
		language := params.Get("lang")
		adExtra := params.Get("ad_extra")
		ss = showSvc.Index(c, mid, plat, build, buvid, channel, ip, ak, network, mobiApp, device, language, adExtra, false, time.Now())
	}
	res := map[string]interface{}{
		"data": ss,
	}
	returnDataJSON(c, res, 25, nil)
}

func showTemps(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
	)
	// get params
	mobiApp := params.Get("mobi_app")
	// check params
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	ip := metadata.String(c, metadata.RemoteIP)
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	// display
	data := showSvc.Display(c, mid, plat, 0, header.Get(_headerBuvid), "", ip, "", "wifi", mobiApp, device, "hans", "", true, time.Now())
	returnJSON(c, data, nil)
}

func showChange(c *bm.Context) {
	params := c.Request.Form
	header := c.Request.Header
	// get params
	mobiApp := params.Get("mobi_app")
	mobiApp = model.MobiAPPBuleChange(mobiApp)
	randStr := params.Get("rand")
	buildStr := params.Get("build")
	// check params
	rand, err := strconv.Atoi(randStr)
	if err != nil {
		log.Error("strconv.Atoi(%s) error(%v)", randStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if rand < 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	// normal data
	ip := metadata.String(c, metadata.RemoteIP)
	buvid := header.Get(_headerBuvid)
	network := params.Get("network")
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	build, _ := strconv.Atoi(buildStr)
	// change
	sis := showSvc.Change(c, mid, build, plat, rand, buvid, ip, network, mobiApp, device)
	returnJSON(c, sis, nil)
	// infoc
	showSvc.Infoc(mid, plat, buvid, ip, "/x/v2/show/change", sis, time.Now())
}

// showRegionChange
func showRegionChange(c *bm.Context) {
	// params := c.Request.Form
	// mobiApp := params.Get("mobi_app")
	// mobiApp = model.MobiAPPBuleChange(mobiApp)
	// device := params.Get("device")
	// buildStr := params.Get("build")
	// plat := model.Plat(mobiApp, device)
	// // get params
	// randStr := params.Get("rand")
	// // check params
	// rand, err := strconv.Atoi(randStr)
	// if err != nil {
	// 	log.Error("strconv.Atoi(%s) error(%v)", randStr, err)
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// if rand < 0 {
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// ridStr := params.Get("rid")
	// rid, err := strconv.Atoi(ridStr)
	// if err != nil {
	// 	log.Error("ridStr(%s) error(%v)", ridStr, err)
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// build, _ := strconv.Atoi(buildStr)
	// data := showSvc.RegionChange(c, rid, rand, plat, build, mobiApp)
	// returnJSON(c, data, nil)
	c.JSON(nil, ecode.NothingFound)
}

// showBangumiChange
func showBangumiChange(c *bm.Context) {
	// params := c.Request.Form
	// mobiApp := params.Get("mobi_app")
	// mobiApp = model.MobiAPPBuleChange(mobiApp)
	// device := params.Get("device")
	// plat := model.Plat(mobiApp, device)
	// // get params
	// randStr := params.Get("rand")
	// // check params
	// rand, err := strconv.Atoi(randStr)
	// if err != nil {
	// 	log.Error("strconv.Atoi(%s) error(%v)", randStr, err)
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// if rand < 0 {
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// data := showSvc.BangumiChange(c, rand, plat)
	// returnJSON(c, data, nil)
	c.JSON(nil, ecode.NothingFound)
}

// showArticleChange
func showArticleChange(c *bm.Context) {
	// data := []*show.Item{}
	// returnJSON(c, data, nil)
	c.JSON(nil, ecode.NothingFound)
}

// showDislike
func showDislike(c *bm.Context) {
	// var (
	// 	params = c.Request.Form
	// 	header = c.Request.Header
	// 	mid    int64
	// )
	// // get params
	// mobiApp := params.Get("mobi_app")
	// mobiApp = model.MobiAPPBuleChange(mobiApp)
	// device := params.Get("device")
	// plat := model.Plat(mobiApp, device)
	// if midInter, ok := c.Get("mid"); ok {
	// 	mid = midInter.(int64)
	// }
	// idStr := params.Get("id")
	// gt := params.Get("goto")
	// if !model.IsGoto(gt) {
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// // normal data
	// ip := metadata.String(c, metadata.RemoteIP)
	// buvid := header.Get(_headerBuvid)
	// // parse id
	// id, err := strconv.ParseInt(idStr, 10, 64)
	// if err != nil {
	// 	log.Error("strconv.Atoi(%s) error(%v)", idStr, err)
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// // change
	// si := showSvc.Dislike(c, mid, plat, id, buvid, mobiApp, device, gt, ip)
	// if si == nil {
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// returnJSON(c, si, nil)
	// // infoc
	// showSvc.Infoc(mid, plat, buvid, ip, "/x/v2/show/change/dislike", []*show.Item{si}, time.Now())
	c.JSON(nil, ecode.NothingFound)
}

// nolint:gomnd
func showWidget(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	buildStr := params.Get("build")
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		log.Error("strconv.Atoi(%s) error(%v)", buildStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var data []*show.Item
	if ss, ok := showSvc.AuditChild(c, mobiApp, plat, build, mid, device); ok {
		if len(ss) > 3 {
			data = ss[:3]
		} else {
			data = ss
		}
		returnJSON(c, data, nil)
		return
	}
	data = showSvc.Widget(c, plat)
	returnJSON(c, data, nil)
}

// show live change
func showLiveChange(c *bm.Context) {
	// params := c.Request.Form
	// // get params
	// randStr := params.Get("rand")
	// ak := params.Get("access_key")
	// rand, err := strconv.Atoi(randStr)
	// if err != nil {
	// 	log.Error("strconv.Atoi(%s) error(%v)", randStr, err)
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// if rand < 0 {
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// var mid int64
	// if midInter, ok := c.Get("mid"); ok {
	// 	mid = midInter.(int64)
	// }
	// // change
	// ip := metadata.String(c, metadata.RemoteIP)
	// data, err := showSvc.LiveChange(c, mid, ak, ip, rand, time.Now())
	// returnJSON(c, data, err)
	c.JSON(nil, ecode.NothingFound)
}

// popular hot tab popular
func popular(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
		mid    int64
	)
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	idxStr := params.Get("idx")
	loginEventStr := params.Get("login_event")
	lastParam := params.Get("last_param")
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	loginEvent, err := strconv.Atoi(loginEventStr)
	if err != nil {
		loginEvent = 0
	}
	idx, err := strconv.ParseInt(idxStr, 10, 64)
	if err != nil || idx < 0 {
		idx = 0
	}
	now := time.Now()
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	buvid := header.Get(_headerBuvid)
	// get audit data, if check audit hit.
	data, ok := showSvc.AuditFeed(c, mobiApp, plat, build, mid, device)
	if !ok {
		data = showSvc.FeedIndex(c, mid, idx, plat, build, loginEvent, lastParam, mobiApp, device, buvid, now)
	}
	c.JSON(data, nil)
}

// popular hot tab popular
func popular2(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
		mid    int64
	)
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	idxStr := params.Get("idx")
	entranceIdStr := params.Get("entrance_id")
	loginEventStr := params.Get("login_event")
	lastParam := params.Get("last_param")
	spmid := params.Get("spmid")
	build, _ := strconv.Atoi(buildStr)
	plat := model.Plat(mobiApp, device)
	loginEvent, err := strconv.Atoi(loginEventStr)
	if err != nil {
		loginEvent = 0
	}
	idx, err := strconv.ParseInt(idxStr, 10, 64)
	if err != nil || idx < 0 {
		idx = 0
	}
	now := time.Now()
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	entranceId, _ := strconv.ParseInt(entranceIdStr, 10, 64)
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		buvid = params.Get("buvid") // if H5, use param to pass buvid
	}
	// get audit data, if check audit hit.
	data, ok := showSvc.AuditFeed2(c, mobiApp, plat, build, mid, device)
	var (
		ver    string
		config *show.HotConfig
	)
	if !ok {
		log.Warn("popular_index api(%s) mid(%d) buvid(%s) mobi_app(%s) device(%s) idx(%d) lastparam(%s)", "/x/v2/show/popular/index", mid, buvid, mobiApp, device, idx, lastParam)
		data, ver, config, err = showSvc.FeedIndex2(c, mid, idx, entranceId, plat, build, loginEvent, mobiApp, device, buvid, spmid, now)
	}
	c.JSONMap(map[string]interface{}{"data": data, "ver": ver, "config": config}, err)
}

func selectedSerie(c *bm.Context) {
	param := new(selected.SelectedParam)
	if err := c.Bind(param); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(showSvc.SerieShow(c, param.Type, param.Number, mid, param.MobiApp, param.Device))
}

func series(c *bm.Context) {
	param := new(struct {
		Type string `form:"type" validate:"required"`
	})
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(showSvc.AllSeries(c, param.Type))
}

func addFav(c *bm.Context) {
	rawMid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	mid := rawMid.(int64)
	param := new(struct {
		Type string `form:"type" validate:"required"`
	})
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Type != _weeklySelected {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, showSvc.AddFav(c, mid))
}

func delFav(c *bm.Context) {
	rawMid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	mid := rawMid.(int64)
	param := new(struct {
		Type string `form:"type" validate:"required"`
	})
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Type != _weeklySelected {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, showSvc.DelFav(c, mid))
}

func checkFav(c *bm.Context) {
	rawMid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	mid := rawMid.(int64)
	param := new(struct {
		Type string `form:"type" validate:"required"`
	})
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Type != _weeklySelected {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(showSvc.CheckFav(c, mid))
}

func precious(c *bm.Context) {
	params := &struct {
		Style   int64  `form:"style"`
		Mid     int64  `form:"-"`
		MobiApp string `form:"mobi_app"`
		Device  string `form:"device"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	mid, ok := c.Get("mid")
	if ok {
		params.Mid = mid.(int64)
	}
	c.JSON(showSvc.Precious(c, params.Style, params.Mid, params.MobiApp, params.Device))
}

func aggregation(c *bm.Context) {
	var (
		param = new(struct {
			HotwordID int64  `form:"hotword_id" validate:"required"`
			MobiApp   string `form:"mobi_app"`
			Device    string `form:"device"`
		})
		request = c.Request
		mid     int64
		err     error
	)
	if err = c.Bind(param); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	buvid := request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := request.Cookie(_cookieBuvid3)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	data, aggsc, err := showSvc.Aggregation(c, param.HotwordID, mid, param.MobiApp, param.Device)
	c.JSON(data, err)
	if err == nil {
		showSvc.InfocAggregation(param.HotwordID, mid, buvid, "/x/v2/show/popular/aggregation", aggsc, time.Now())
	}
}

func preciousSubAdd(ctx *bm.Context) {
	mid, ok := ctx.Get("mid")
	if !ok {
		ctx.JSON(nil, ecode.NoLogin)
		return
	}
	ctx.JSON(nil, showSvc.PreciousSubAdd(ctx, mid.(int64)))
}

func preciousSubDel(ctx *bm.Context) {
	mid, ok := ctx.Get("mid")
	if !ok {
		ctx.JSON(nil, ecode.NoLogin)
		return
	}
	ctx.JSON(nil, showSvc.PreciousSubDel(ctx, mid.(int64)))
}

func popularArchive(ctx *bm.Context) {
	req := &popularmodel.PopularArchiveRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(showSvc.PurePopularArchive(ctx, req))
}
