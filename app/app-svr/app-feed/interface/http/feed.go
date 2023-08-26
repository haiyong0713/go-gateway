package http

import (
	"strconv"
	"strings"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/stat"
	"go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/app/app-svr/app-feed/interface/model/dislike"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
)

const (
	_headerBuvid      = "Buvid"
	_headerDeviceID   = "Device-ID"
	_headerAppList    = "AppList"
	_headerDeviceInfo = "DeviceInfo"
	_blankCardType    = "_blank"
)

func feedIndex(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	header := c.Request.Header
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	// get params
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	platform := params.Get("platform")
	network := params.Get("network")
	buildStr := params.Get("build")
	idxStr := params.Get("idx")
	pullStr := params.Get("pull")
	styleStr := params.Get("style")
	loginEventStr := params.Get("login_event")
	openEvent := params.Get("open_event")
	bannerHash := params.Get("banner_hash")
	adExtra := params.Get("ad_extra")
	interest := params.Get("interest")
	flushStr := params.Get("flush")
	autoplayCard, _ := strconv.Atoi(params.Get("autoplay_card"))
	accessKey := params.Get("access_key")
	actionKey := params.Get("actionKey")
	appkey := params.Get("appkey")
	statistics := params.Get("statistics")
	// check params
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	style, _ := strconv.Atoi(styleStr)
	flush, _ := strconv.Atoi(flushStr)
	// get audit data, if check audit hit.
	is, ok := feedSvc.Audit(c, mid, mobiApp, device, plat, build)
	if ok {
		c.JSON(is, nil)
		return
	}
	buvid := header.Get(_headerBuvid)
	dvcid := header.Get(_headerDeviceID)
	// page
	idx, err := strconv.ParseInt(idxStr, 10, 64)
	if err != nil || idx < 0 {
		idx = 0
	}
	// pull default
	pull, err := strconv.ParseBool(pullStr)
	if err != nil {
		pull = true
	}
	// login event
	loginEvent, err := strconv.Atoi(loginEventStr)
	if err != nil {
		loginEvent = 0
	}
	now := time.Now()
	// index
	data, userFeature, isRcmd, newUser, code, feedclean, autoPlayInfoc, info, err := feedSvc.Index(c, mid, plat, build, buvid, network, mobiApp, device, platform, openEvent, loginEvent, idx, pull, now, bannerHash, adExtra, interest, style, flush, autoplayCard, accessKey, actionKey, appkey, statistics)
	res := map[string]interface{}{
		"data": data,
		"config": map[string]interface{}{
			"feed_clean_abtest": feedclean,
		},
	}
	c.JSONMap(res, err)
	if err != nil {
		return
	}
	// infoc
	items := make([]*ai.Item, 0, len(data))
	cardTypes := make(map[int]string, len(data))
	cardGotos := make(map[int]string, len(data))
	for index, item := range data {
		items = append(items, item.AI)
		cardTypes[index] = _blankCardType
		cardGotos[index] = _blankCardType
	}
	feedSvc.IndexInfoc(c, mid, plat, build, buvid, "/x/feed/index", userFeature, style, code, items, isRcmd, pull, newUser, now, "", dvcid, network, flush, autoPlayInfoc, 0, info, nil, "", "", 0, 0, nil, nil, nil, openEvent, "", nil, nil, cardTypes,
		nil, nil, mobiApp, nil, cardGotos, nil, nil)
}

func feedUpper(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	midInter, _ := c.Get("mid")
	mid = midInter.(int64)
	// get params
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	pnStr := params.Get("pn")
	psStr := params.Get("ps")
	// check params
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	// check page
	// check page
	pn, err := strconv.Atoi(pnStr)
	if err != nil || pn < 1 {
		pn = 1
	}
	ps, err := strconv.Atoi(psStr)
	//nolint:gomnd
	if err != nil || ps < 1 {
		ps = 20
	} else if ps > 100 {
		ps = 100
	}
	plat := model.Plat(mobiApp, device)
	now := time.Now()
	uas, _ := feedSvc.Upper(c, mid, plat, build, pn, ps, now)
	data := map[string]interface{}{}
	if len(uas) != 0 {
		data["item"] = uas
	} else {
		data["item"] = []struct{}{}
	}
	if model.IsBlueByMobiApp(mobiApp) {
		c.JSON(data, nil)
		return
	}
	uls, count := feedSvc.UpperLive(c, mid)
	if len(uls) != 0 {
		data["live"] = struct {
			Item  []*feed.Item `json:"item"`
			Count int          `json:"count"`
			Conut int          `json:"conut"`
		}{uls, count, count}
	}
	c.JSON(data, nil)
}

func feedUpperArchive(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	midInter, _ := c.Get("mid")
	mid = midInter.(int64)
	// get params
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	pnStr := params.Get("pn")
	psStr := params.Get("ps")
	// check params
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	// check page
	pn, err := strconv.Atoi(pnStr)
	if err != nil || pn < 1 {
		pn = 1
	}
	ps, err := strconv.Atoi(psStr)
	//nolint:gomnd
	if err != nil || ps < 1 {
		ps = 20
	} else if ps > 200 {
		ps = 200
	}
	plat := model.Plat(mobiApp, device)
	now := time.Now()
	uas, _ := feedSvc.UpperArchive(c, mid, plat, build, pn, ps, now)
	data := map[string]interface{}{}
	if len(uas) != 0 {
		data["item"] = uas
	} else {
		data["item"] = []struct{}{}
	}
	c.JSON(data, nil)
}

func feedUpperBangumi(c *bm.Context) {
	// var mid int64
	// params := c.Request.Form
	// midInter, _ := c.Get("mid")
	// mid = midInter.(int64)
	// // get params
	// mobiApp := params.Get("mobi_app")
	// device := params.Get("device")
	// buildStr := params.Get("build")
	// pnStr := params.Get("pn")
	// psStr := params.Get("ps")
	// // check params
	// build, err := strconv.Atoi(buildStr)
	// if err != nil {
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// // check page
	// pn, err := strconv.Atoi(pnStr)
	// if err != nil || pn < 1 {
	// 	pn = 1
	// }
	// ps, err := strconv.Atoi(psStr)
	// if err != nil || ps < 1 {
	// 	ps = 20
	// } else if ps > 200 {
	// 	ps = 200
	// }
	// plat := model.Plat(mobiApp, device)
	// now := time.Now()
	// uas, _ := feedSvc.UpperBangumi(c, mid, plat, build, pn, ps, now)
	// data := map[string]interface{}{}
	// if len(uas) != 0 {
	// 	data["item"] = uas
	// } else {
	// 	data["item"] = []struct{}{}
	// }
	// c.JSON(data, nil)
	c.JSON(nil, ecode.NothingFound)
}

func feedUpperArticle(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	midInter, _ := c.Get("mid")
	mid = midInter.(int64)
	// get params
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	pnStr := params.Get("pn")
	psStr := params.Get("ps")
	// check params
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	// check page
	pn, err := strconv.Atoi(pnStr)
	if err != nil || pn < 1 {
		pn = 1
	}
	ps, err := strconv.Atoi(psStr)
	//nolint:gomnd
	if err != nil || ps < 1 {
		ps = 20
	} else if ps > 200 {
		ps = 200
	}
	plat := model.Plat(mobiApp, device)
	now := time.Now()
	uas, _ := feedSvc.UpperArticle(c, mid, plat, build, pn, ps, now)
	data := map[string]interface{}{}
	if len(uas) != 0 {
		data["item"] = uas
	} else {
		data["item"] = []struct{}{}
	}
	c.JSON(data, nil)
}

func feedUnreadCount(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	midInter, _ := c.Get("mid")
	mid = midInter.(int64)
	// get params
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	// check params
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	total, feedCount, articleCount := feedSvc.UnreadCount(c, mid, plat, build, time.Now())
	c.JSON(struct {
		Total   int `json:"total"`
		Count   int `json:"count"`
		Article int `json:"article"`
	}{total, feedCount, articleCount}, nil)
}

func feedDislike(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	header := c.Request.Header
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	gt := params.Get("goto")
	if gt == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	id, _ := strconv.ParseInt(params.Get("id"), 10, 64)
	reasonID, _ := strconv.ParseInt(params.Get("reason_id"), 10, 64)
	cmreasonID, _ := strconv.ParseInt(params.Get("cm_reason_id"), 10, 64)
	feedbackID, _ := strconv.ParseInt(params.Get("feedback_id"), 10, 64)
	upperID, _ := strconv.ParseInt(params.Get("mid"), 10, 64)
	rid, _ := strconv.ParseInt(params.Get("rid"), 10, 64)
	tagID, _ := strconv.ParseInt(params.Get("tag_id"), 10, 64)
	adcb := params.Get("ad_cb")
	fromspmid := params.Get("from_spmid")
	frommodule := params.Get("from_module")
	disableRcmd, _ := strconv.ParseInt(params.Get("disable_rcmd"), 10, 64)
	fromAvid, _ := strconv.ParseInt(params.Get("from_avid"), 10, 64)
	fromTypeStr := params.Get("from_type")
	fromType := dislike.FeedbackFromType[fromTypeStr]
	//ios客户端在51版本传的是int，但是在其他版本传的是字符串，所以要兼容51版本
	if plat == model.PlatIPhone && (build >= 65100000 && build < 65200000) {
		fromType, _ = strconv.ParseInt(fromTypeStr, 10, 64)
	}
	materialId, _ := strconv.ParseInt(params.Get("material_id"), 10, 64)
	reportData := params.Get("report_data")
	buvid := header.Get(_headerBuvid)
	if buvid == "" && mid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, feedSvc.Dislike(c, mid, id, buvid, gt, reasonID, cmreasonID, feedbackID, upperID, rid, tagID, adcb,
		fromspmid, frommodule, time.Now(), disableRcmd, fromAvid, fromType, materialId, reportData))
}

func feedDislikeCancel(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	header := c.Request.Header
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	gt := params.Get("goto")
	if gt == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	id, _ := strconv.ParseInt(params.Get("id"), 10, 64)
	reasonID, _ := strconv.ParseInt(params.Get("reason_id"), 10, 64)
	cmreasonID, _ := strconv.ParseInt(params.Get("cm_reason_id"), 10, 64)
	feedbackID, _ := strconv.ParseInt(params.Get("feedback_id"), 10, 64)
	upperID, _ := strconv.ParseInt(params.Get("mid"), 10, 64)
	rid, _ := strconv.ParseInt(params.Get("rid"), 10, 64)
	tagID, _ := strconv.ParseInt(params.Get("tag_id"), 10, 64)
	adcb := params.Get("ad_cb")
	fromspmid := params.Get("from_spmid")
	frommodule := params.Get("from_module")
	fromAvid, _ := strconv.ParseInt(params.Get("from_avid"), 10, 64)
	fromTypeStr := params.Get("from_type")
	fromType := dislike.FeedbackFromType[fromTypeStr]
	buvid := header.Get(_headerBuvid)
	closeRcmd, _ := strconv.ParseInt(params.Get("close_rcmd"), 10, 64)
	//ios客户端在51版本传的是int，但是在其他版本传的是字符串，所以要兼容51版本
	if plat == model.PlatIPhone && (build >= 65100000 && build < 65200000) {
		fromType, _ = strconv.ParseInt(fromTypeStr, 10, 64)
	}
	materialId, _ := strconv.ParseInt(params.Get("material_id"), 10, 64)
	reportData := params.Get("report_data")
	if buvid == "" && mid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, feedSvc.DislikeCancel(c, mid, id, buvid, gt, reasonID, cmreasonID, feedbackID, upperID, rid,
		tagID, adcb, fromspmid, frommodule, time.Now(), closeRcmd, fromAvid, fromType, materialId, reportData))
}

func feedUpperRecent(c *bm.Context) {
	// var mid int64
	// params := c.Request.Form
	// if midInter, ok := c.Get("mid"); ok {
	// 	mid = midInter.(int64)
	// }
	// aidStr := params.Get("param")
	// aid, err := strconv.ParseInt(aidStr, 10, 64)
	// if err != nil {
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// upperStr := params.Get("vmid")
	// upperID, err := strconv.ParseInt(upperStr, 10, 64)
	// if err != nil {
	// 	c.JSON(nil, ecode.RequestErr)
	// 	return
	// }
	// c.JSON(struct {
	// 	Item []*feed.Item `json:"item"`
	// }{feedSvc.UpperRecent(c, mid, upperID, aid, time.Now())}, nil)
	c.JSON(nil, ecode.NothingFound)
}

func feedIndexTab(c *bm.Context) {
	var (
		id      int64
		items   []*feed.Item
		isBnj   bool
		bnjDays int
		cover   string
		err     error
		mid     int64
	)
	header := c.Request.Header
	params := c.Request.Form
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	now := time.Now()
	idStr := params.Get("id")
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	network := params.Get("network")
	platform := params.Get("platform")
	buildStr := params.Get("build")
	// check params
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	if id, _ = strconv.ParseInt(idStr, 10, 64); id <= 0 {
		c.JSON(struct {
			Tab []*operate.Menu `json:"tab"`
		}{feedSvc.Menus(c, plat, build, now)}, nil)
		return
	}
	buvid := header.Get(_headerBuvid)
	accessKey := params.Get("access_key")
	actionKey := params.Get("actionKey")
	statistics := params.Get("statistics")
	appkey := params.Get("appkey")
	items, cover, isBnj, bnjDays, err = feedSvc.Actives(c, id, mid, platform, now, mobiApp, buvid, device, accessKey, actionKey, appkey, statistics, network, build)
	c.JSON(struct {
		Cover   string       `json:"cover"`
		IsBnj   bool         `json:"is_bnj,omitempty"`
		BnjDays int          `json:"bnj_days,omitempty"`
		Item    []*feed.Item `json:"item"`
	}{cover, isBnj, bnjDays, items}, err)
}

func feedIndex2(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	dvcid := header.Get(_headerDeviceID)
	applist := header.Get(_headerAppList)
	deviceInfo := header.Get(_headerDeviceInfo)
	fawkesAppkey := header.Get("App-key")
	fawkesEnv := header.Get("Env")
	if fawkesEnv == "" {
		fawkesEnv = "prod"
	}
	param := &feed.IndexParam{}
	// get params
	if err := c.Bind(param); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	_, ok := cdm.Columnm[param.Column]
	if !ok {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	dev, ok := device.FromContext(c)
	if ok {
		param.Ua = dev.UserAgent
	}
	// 兼容老的style逻辑，3为新单列
	style := int(cdm.Columnm[param.Column])
	if style == 1 {
		style = 3
	}
	now := time.Now()
	defer func() {
		userI, _ := c.Get("user")
		// user
		user, ok := userI.(string)
		if !ok || user == "" {
			user = "no_user"
		}
		dt := time.Since(now)
		if param.LoginEvent == 1 || param.LoginEvent == 2 {
			bm.MetricServerReqDur.Observe(int64(dt/time.Millisecond), "x/v2/feed/index/coldstart", user)
		}
	}()
	// check params
	plat := model.Plat(param.MobiApp, param.Device)
	// get audit data, if check audit hit.
	if data, ok := feedSvc.Audit2(c, mid, param.MobiApp, param.Device, plat, param.Build, param.Column); ok {
		c.JSON(struct {
			Item []card.Handler `json:"items"`
		}{Item: data}, nil)
		return
	}
	// index
	data, config, infc, info, err := feedSvc.Index2(c, buvid, mid, plat, param, style, applist, deviceInfo, now)
	if afv, ok := feedSvc.FawkesVersionCache[fawkesEnv]; ok {
		if fv, ok := afv[fawkesAppkey]; ok {
			c.Writer.Header().Set("CONFIG-V", strconv.FormatInt(fv.Config, 10))
			c.Writer.Header().Set("FF-V", strconv.FormatInt(fv.FF, 10))
		}
	}
	c.JSON(struct {
		Item   []card.Handler `json:"items"`
		Config *feed.Config   `json:"config"`
	}{Item: data, Config: config}, err)
	if err != nil {
		return
	}
	statResponseCard(data, param.Column, plat)
	setSessionRecordResponse(c, data)
	// 隐私弹窗模式
	if param.PrivacyDisagreeMode == 1 {
		return
	}
	feedSvc.FeedAppListProduce(c, param, mid, buvid, applist)
	// infoc
	items := make([]*ai.Item, 0, len(data))
	cardTypes := make(map[int]string, len(data))
	cardGotos := make(map[int]string, len(data))
	for index, item := range data {
		handleItemCardGotoInfoc(item, infc, index)
		handleItemCardTypeInfoc(item, infc, index)
		handleItemBadgeInfoc(item, infc)
		items = append(items, item.Get().Rcmd)
		cardTypes[index] = string(item.Get().CardType)
		cardGotos[index] = string(item.Get().CardGoto)
	}
	feedSvc.IndexInfoc(c, mid, plat, param.Build, buvid, "/x/feed/index", infc.UserFeature, style, infc.Code, items,
		infc.IsRcmd, param.Pull, infc.NewUser, now, "", dvcid, param.Network, param.Flush, infc.AutoPlayInfoc,
		param.DeviceType, info, infc.IsGifCover, infc.BannerHash, param.BannerHash, param.LoginEvent, infc.AdCode, infc.AdError,
		infc.AdPos, infc.AdPkCode, param.OpenEvent, param.IsMelloi, infc.PendantMap, infc.SubGotoMap, cardTypes,
		infc.AiTunnelOidMap, infc.AiBangumiRcmdOgvInfoMap, param.MobiApp, infc.DiscardReason, cardGotos, infc.GameBadge,
		infc.BadgeMap)
}

func feedIndexTab2(c *bm.Context) {
	var (
		id      int64
		items   []card.Handler
		isBnj   bool
		bnjDays int
		cover   string
		err     error
		mid     int64
	)
	header := c.Request.Header
	params := c.Request.Form
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	now := time.Now()
	idStr := params.Get("id")
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	network := params.Get("network")
	// check params
	build, err := strconv.Atoi(buildStr)
	platform := params.Get("platform")
	// check params
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	if id, _ = strconv.ParseInt(idStr, 10, 64); id <= 0 {
		c.JSON(struct {
			Tab []*operate.Menu `json:"tab"`
		}{feedSvc.Menus(c, plat, build, now)}, nil)
		return
	}
	buvid := header.Get(_headerBuvid)
	accessKey := params.Get("access_key")
	actionKey := params.Get("actionKey")
	appkey := params.Get("appkey")
	statistics := params.Get("statistics")
	items, cover, isBnj, bnjDays, err = feedSvc.Actives2(c, id, mid, mobiApp, platform, device, plat, build, now, accessKey, actionKey, appkey, statistics, buvid, network)
	c.JSON(struct {
		Cover   string         `json:"cover"`
		IsBnj   bool           `json:"is_bnj,omitempty"`
		BnjDays int            `json:"bnj_days,omitempty"`
		Item    []card.Handler `json:"items"`
	}{cover, isBnj, bnjDays, items}, err)
}

func feedIndexConverge(c *bm.Context) {
	var (
		mid   int64
		title string
		cover string
		uri   string
	)
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &feed.ConvergeParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	now := time.Now()
	data, converge, _, _, _, err := feedSvc.Converge(c, mid, plat, param, buvid, now)
	if converge != nil {
		title = converge.Title
		cover = converge.Cover
		uri = converge.URI
	}
	c.JSON(struct {
		Items []card.Handler `json:"items"`
		Title string         `json:"title"`
		Cover string         `json:"cover,omitempty"`
		Param string         `json:"param,omitempty"`
		URI   string         `json:"uri,omitempty"`
	}{Items: data, Title: title, Cover: cover, Param: strconv.FormatInt(param.ID, 10), URI: uri}, err)
}

func feedIndexAvConverge(c *bm.Context) {
	var (
		mid   int64
		title string
	)
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &feed.ConvergeParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	now := time.Now()
	data, converge, _, err := feedSvc.AvConverge(c, mid, plat, buvid, param, now)
	if converge != nil {
		title = converge.Title
	}
	c.JSON(struct {
		Items []card.Handler `json:"items"`
		Title string         `json:"title"`
	}{Items: data, Title: title}, err)
}

func statResponseCard(data []card.Handler, cs cdm.ColumnStatus, plat int8) {
	rowType := stat.BuildRowType(cs, plat)
	for _, handler := range data {
		base := handler.Get()
		if base == nil {
			continue
		}
		stat.MetricResponseCardTotal.Inc(rowType, string(base.CardGoto), string(base.CardType),
			descFromCardTypeAndCardGoto(string(base.CardType), string(base.CardGoto)))
	}
}

func descFromCardTypeAndCardGoto(cardType, cardGoto string) string {
	switch cardGoto {
	case "av":
		switch cardType {
		case "small_cover_v2":
			return "ugc小卡"
		case "large_cover_v1":
			return "ugc大卡"
		case "large_cover_single_v9":
			return "ugc-inline"
		default:
			return "ugc默认卡"
		}
	case "bangumi":
		switch cardType {
		case "small_cover_v2", "ogv_small_cover":
			return "ogv小卡-bangumi"
		case "large_cover_single_v7":
			return "bangumi-inline"
		case "large_cover_v1":
			return "ogv大卡-bangumi"
		}
		return "ogv默认卡"
	case "pgc":
		switch cardType {
		case "small_cover_v2":
			return "ogv小卡-pgc"
		case "large_cover_v7", "large_cover_single_v7":
			return "pgc-inline"
		}
		return "ogv默认卡"
	case "live":
		switch cardType {
		case "small_cover_v2", "small_cover_v9":
			return "直播小卡"
		case "large_cover_v1":
			return "直播大卡"
		case "large_cover_v8", "large_cover_single_v8":
			return "直播inline"
		}
		return "直播小卡"
	case "special_s":
		return "运营特殊小卡"
	case "articla_s":
		return "专栏卡"
	case "big_tunnel":
		return "订阅大卡"
	case "new_tunnel":
		return "订阅小卡"
	case "banner":
		return "banner"
	case "game":
		return "游戏小卡"
	case "inline_av", "inline_av_v2":
		return "ugc-inline"
	case "inline_pgc":
		return "pgc-inline"
	case "inline_live":
		return "live-inline"
	case "ad_inline_av":
		return "广告ugc-inline"
	case "ad_inline_ogv":
		return "广告ogv-inline"
	case "ad_inline_live":
		return "广告live-inline"
	default:
	}
	switch cardType {
	case "cm_v1", "cm_v2", "cm_double_v9", "cm_single_v9", "cm_double_v7", "cm_single_v7", "cm_single_v1":
		return "广告卡"
	default:
	}
	return ""
}

func handleItemCardGotoInfoc(item card.Handler, infoc *feed.Infoc, index int) {
	const (
		_tunnelSubGotoOgvRcmd = "ogv_rcmd"
	)
	if infoc == nil {
		return
	}
	switch item.Get().CardGoto {
	case cdm.CardGotoNewTunnel:
		if infoc.SubGotoMap == nil {
			infoc.SubGotoMap = make(map[int][]string)
		}
		if infoc.AiTunnelOidMap == nil {
			infoc.AiTunnelOidMap = make(map[int][]string)
		}
		if handler := item.(*card.UniversalNotifyTunnelV1); handler != nil {
			for _, tunnelItemV1 := range handler.Items {
				// subgoto上报
				infoc.SubGotoMap[index] = append(infoc.SubGotoMap[index], tunnelItemV1.SubGoto)
				// oid上报
				switch tunnelItemV1.SubGoto {
				case _tunnelSubGotoOgvRcmd:
					infoc.AiTunnelOidMap[index] = append(infoc.AiTunnelOidMap[index], constructInfocOids(item.Get().Rcmd.MsgIDs, tunnelItemV1.Param)...)
				default:
				}
			}
		}
	case cdm.CardGotoBigTunnel:
		if infoc.SubGotoMap == nil {
			infoc.SubGotoMap = make(map[int][]string)
		}
		if handler := item.(*card.UniversalNotifyTunnelLargeV1); handler != nil {
			// subgoto上报
			infoc.SubGotoMap[index] = append(infoc.SubGotoMap[index], handler.Item.SubGoto)
			// oid上报
			switch handler.Item.SubGoto {
			case _tunnelSubGotoOgvRcmd:
				if infoc.AiTunnelOidMap == nil {
					infoc.AiTunnelOidMap = make(map[int][]string)
				}
				infoc.AiTunnelOidMap[index] = constructInfocOids(item.Get().Rcmd.MsgIDs, handler.Item.Param)
			default:
			}
		}
	default:
	}
}

func constructInfocOids(msgIDs, param string) []string {
	var out []string
	slots := strings.Split(msgIDs, ",")
	for _, slot := range slots {
		oids := strings.Split(slot, "|")
		for _, oidStr := range oids {
			if param == oidStr {
				out = append(out, oids...)
				break
			}
		}
	}
	return out
}

func handleItemCardTypeInfoc(item card.Handler, infoc *feed.Infoc, id int) {
	if infoc == nil {
		return
	}
	switch item.Get().CardType {
	case cdm.SmallCoverV9:
		handler := item.(*card.SmallCoverV9)
		if handler == nil || handler.LeftCoverBadgeNewStyle == nil {
			return
		}
		id := handler.Get().Rcmd.ID
		if infoc.PendantMap == nil {
			infoc.PendantMap = map[int64]string{id: handler.LeftCoverBadgeNewStyle.Text}
			return
		}
		infoc.PendantMap[id] = handler.LeftCoverBadgeNewStyle.Text
	case cdm.SmallCoverV1:
		handler := item.(*card.SmallCoverV1)
		if handler == nil {
			return
		}
		ogvInfo := &feed.BangumiRcmdInfoc{
			SeasonId: handler.SeasonId,
			Epid:     handler.Epid,
		}
		if infoc.AiBangumiRcmdOgvInfoMap == nil {
			infoc.AiBangumiRcmdOgvInfoMap = map[int]*feed.BangumiRcmdInfoc{id: ogvInfo}
			return
		}
		infoc.AiBangumiRcmdOgvInfoMap[id] = ogvInfo
	case cdm.SmallCoverV4:
		handler := item.(*card.SmallCoverV4)
		if handler == nil {
			return
		}
		ogvInfo := &feed.BangumiRcmdInfoc{
			SeasonId: handler.SeasonId,
			Epid:     handler.Epid,
		}
		if infoc.AiBangumiRcmdOgvInfoMap == nil {
			infoc.AiBangumiRcmdOgvInfoMap = map[int]*feed.BangumiRcmdInfoc{id: ogvInfo}
			return
		}
		infoc.AiBangumiRcmdOgvInfoMap[id] = ogvInfo
	case cdm.SmallCoverV10:
		handler := item.(*card.SmallCoverV10)
		if handler == nil || handler.LeftCoverBadgeNewStyle == nil {
			return
		}
		id := handler.Get().Rcmd.ID
		if infoc.GameBadge == nil {
			infoc.GameBadge = map[int64]string{id: "1"}
			return
		}
		infoc.GameBadge[id] = "1"
	default:
	}
}

func handleItemBadgeInfoc(item card.Handler, infoc *feed.Infoc) {
	if infoc == nil {
		return
	}
	switch item.Get().CardType {
	case cdm.SmallCoverV2:
		handler := item.(*card.SmallCoverV2)
		if handler == nil || handler.BadgeStyle == nil {
			return
		}
		id := handler.Get().Rcmd.ID
		if infoc.BadgeMap == nil {
			infoc.BadgeMap = map[int64]string{id: handler.BadgeStyle.Text}
			return
		}
		infoc.BadgeMap[id] = handler.BadgeStyle.Text
	case cdm.OgvSmallCover:
		handler := item.(*card.OgvSmallCover)
		if handler == nil || handler.BadgeStyle == nil {
			return
		}
		id := handler.Get().Rcmd.ID
		if infoc.BadgeMap == nil {
			infoc.BadgeMap = map[int64]string{id: handler.BadgeStyle.Text}
			return
		}
		infoc.BadgeMap[id] = handler.BadgeStyle.Text
	default:
	}
}

func verticalTab(ctx *bm.Context) {
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &feed.VerticalTabParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	param.Buvid = buvid
	param.Plat = model.Plat(param.MobiApp, param.Device)
	if v, ok := ctx.Get("mid"); ok {
		param.Mid = v.(int64)
	}
	ctx.JSON(feedSvc.VerticalTab(ctx, param))
}

func verticalTabTag(ctx *bm.Context) {
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &feed.VerticalTagParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	param.Buvid = buvid
	param.Plat = model.Plat(param.MobiApp, param.Device)
	if v, ok := ctx.Get("mid"); ok {
		param.Mid = v.(int64)
	}
	ctx.JSON(feedSvc.VerticalTag(ctx, param))
}

func feedIndexInterest(ctx *bm.Context) {
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	var mid int64
	if v, ok := ctx.Get("mid"); ok {
		mid = v.(int64)
	}
	result, err := feedSvc.IndexInterest(ctx, mid, buvid)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(map[string]interface{}{"interest_choose": result}, nil)
}
