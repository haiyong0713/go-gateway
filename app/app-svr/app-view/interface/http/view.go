package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	gwecode "go-gateway/app/app-svr/app-card/ecode"
	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/share"
	"go-gateway/app/app-svr/app-view/interface/model/view"
	viewService "go-gateway/app/app-svr/app-view/interface/service/view"
	resource "go-gateway/app/app-svr/resource/service/model"
	mainEcode "go-gateway/ecode"
	"go-gateway/pkg/idsafe/bvid"
)

const (
	_headerBuvid = "Buvid"
	// 投币引导类型
	coinTypeNone   = "none"
	coinTypeShare  = "share"
	coinTypeFollow = "follow"
)

type CoinAddGuide struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}

var (
	_rate    = []int64{564, 1328, 2592, 4192}
	_formats = map[int]string{1: "mp4", 2: "hdmp4", 3: "flv", 4: "flv"}
	_dislike = []*view.Dislike{
		{
			ID:   5,
			Name: "标题党/封面党",
		},
		{
			ID:   6,
			Name: "内容质量差",
		},
		{
			ID:   7,
			Name: "内容/封面令人不适",
		},
		{
			ID:   8,
			Name: "营销广告",
		},
	}
)

type viewErr struct {
	CustomConfig *viewApi.CustomConfig `json:"custom_config,omitempty"`
}

// viewIndex view handler
func viewIndex(c *bm.Context) {
	var (
		mid, aid   int64
		err        error
		parentMode int
	)
	params := c.Request.Form
	header := c.Request.Header
	// get params
	bvID := params.Get("bvid")
	withoutCharge := params.Get("without_charge")
	aidStr := params.Get("aid")
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	from := params.Get("from")
	trackid := params.Get("trackid")
	network := params.Get("network")
	adExtra := params.Get("ad_extra")
	parentModeStr := params.Get("parent_mode")
	parentMode, _ = strconv.Atoi(parentModeStr)
	spmid := params.Get("spmid")
	fromSpmid := params.Get("from_spmid")
	platform := params.Get("platform")
	filtered := params.Get("filtered")
	slocale := params.Get("s_locale")
	clocale := params.Get("c_locale")
	fawkesAppkey := header.Get("App-key")
	fawkesEnv := header.Get("Env")
	if fawkesEnv == "" {
		fawkesEnv = "prod"
	}
	// check params
	if aidStr == "" && bvID == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if aidStr != "" && aidStr != "0" {
		if aid, err = strconv.ParseInt(aidStr, 10, 64); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	} else if bvID != "" {
		if aid, err = bvid.BvToAv(bvID); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if aid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	dev, _ := device.FromContext(c)
	plat := model.PlatNew(mobiApp, dev.Device)
	buvid := header.Get("Buvid")
	cdnIP := header.Get("X-Cache-Server-Addr")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	autoplay, _ := strconv.Atoi(params.Get("autoplay"))
	now := time.Now()
	// view
	ip := metadata.String(c, metadata.RemoteIP)
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	disableRcmdMode, _ := strconv.Atoi(params.Get("close_rcmd"))
	isMelloi := params.Get("is_melloi")
	viewSvr.ViewInfoc(mid, int(plat), trackid, aidStr, ip, model.PathView, buildStr, buvid, from, now, err, autoplay, spmid, fromSpmid, network, isMelloi)
	data, err := viewSvr.ViewHttp(c, mid, aid, plat, build, parentMode, autoplay, teenagersMode, lessonsMode, disableRcmdMode, mobiApp, dev.Device, buvid, cdnIP, network, adExtra, from, spmid, fromSpmid, trackid, platform, filtered, withoutCharge, isMelloi, dev.Brand, slocale, clocale)
	if err != nil {
		if ecode.EqualError(gwecode.AppViewForRetry, err) {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if ecode.EqualError(ecode.NothingFound, err) && viewSvr.HasCustomConfig(c, aid) {
			c.JSON(&viewErr{CustomConfig: &viewApi.CustomConfig{RedirectUrl: viewSvr.NothingFoundUrl(aid)}}, err)
			return
		}
		c.JSON(nil, err)
		return
	}
	data.Dislikes = _dislike
	// dislike v2
	disableRcmd, _ := strconv.Atoi(params.Get("close_rcmd"))
	data.DislikeReasons(c, cfg.Feature, mobiApp, dev.Device, build, disableRcmd)
	// dislike v2
	compMeta(data, build)
	if mid == 0 && data.Duration > 360 {
		data.Paster, _ = viewSvr.Paster(c, plat, resource.VdoAdsTypeNologin, aidStr, strconv.Itoa(int(data.TypeID)), buvid)
	}
	if afv, ok := viewSvr.FawkesVersionCache[fawkesEnv]; ok {
		if fv, ok := afv[fawkesAppkey]; ok {
			c.Writer.Header().Set("CONFIG-V", strconv.FormatInt(fv.Config, 10))
			c.Writer.Header().Set("FF-V", strconv.FormatInt(fv.FF, 10))
		}
	}
	c.JSON(data, nil)
	viewSvr.RelateInfoc(mid, aid, int(plat), buildStr, buvid, ip, model.PathView, data.ReturnCode, data.UserFeature,
		from, "", data.Relates, now, data.IsRec, autoplay, data.PlayParam, trackid, model.PageTypeRelate,
		fromSpmid, spmid, data.PvFeature, data.TabInfo, isMelloi, nil, 0)
}

// viewPage view page handler.
func viewPage(c *bm.Context) {
	var (
		mid, aid int64
		build    int
		err      error
	)
	params := c.Request.Form
	header := c.Request.Header
	// get params
	bvID := params.Get("bvid")
	aidStr := params.Get("aid")
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	from := params.Get("from")
	trackid := params.Get("trackid")
	spmid := params.Get("spmid")
	fromSpmid := params.Get("from_spmid")
	network := params.Get("network")
	slocale := params.Get("s_locale")
	clocale := params.Get("c_locale")
	platform := params.Get("platform")
	// check params
	if aidStr != "" && aidStr != "0" {
		if aid, err = strconv.ParseInt(aidStr, 10, 64); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	} else if bvID != "" {
		if aid, err = bvid.BvToAv(bvID); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if aid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if build, err = strconv.Atoi(buildStr); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.PlatNew(mobiApp, device)
	buvid := header.Get("Buvid")
	cdnIP := header.Get("X-Cache-Server-Addr")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	autoplay, _ := strconv.Atoi(params.Get("autoplay"))
	ip := metadata.String(c, metadata.RemoteIP)
	now := time.Now()
	// view page
	viewSvr.ViewInfoc(mid, int(plat), trackid, aidStr, ip, model.PathViewPage, buildStr, buvid, from, now, err, autoplay, spmid, fromSpmid, network, "")
	vp, extra, err := viewSvr.ArcView(c, aid, 0, "", "", "", plat)
	if err != nil {
		if ecode.EqualError(gwecode.AppViewForRetry, err) {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(nil, err)
		return
	}
	data, err := viewSvr.ViewPage(c, mid, plat, build, mobiApp, device, cdnIP, false, buvid, slocale, clocale, vp, "", spmid, platform, 0, extra)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	viewService.HideArcAttribute(data.Arc)
	compMeta(data, build)
	c.JSON(data, nil)
}

// videoShot video shot .
func videoShot(c *bm.Context) {
	params := &view.VideoShotParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(viewSvr.Shot(c, params.AID, params.CID, params.MobiApp, params.Build))
}

func shareIcon(c *bm.Context) {
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	var mid, aid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	plat := model.PlatNew(mobiApp, device)
	aid, _ = strconv.ParseInt(params.Get("aid"), 10, 64)
	buvid := c.Request.Header.Get("Buvid")
	build, _ := strconv.ParseInt(params.Get("build"), 10, 64)
	c.JSON(viewSvr.ShareIcon(c, mid, aid, build, plat, buvid), nil)
}

// addShare add a share.
func addShare(c *bm.Context) {
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	platform := params.Get("platform")
	buvid := c.Request.Header.Get("Buvid")
	from := params.Get("from")
	build := params.Get("build")
	buildInt, _ := strconv.ParseInt(build, 10, 64)
	shareChnStr := params.Get("share_channel")
	if shareChnStr == "" {
		shareChnStr = model.ShareDefaultStr
	}
	// check
	aidStr := params.Get("aid")
	aid, _ := strconv.ParseInt(aidStr, 10, 64)
	if aid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	share, isReport, upID, toast, err := viewSvr.AddShare(c, aid, mid, buildInt, shareChnStr, metadata.String(c, metadata.RemoteIP), buvid)
	c.JSON(struct {
		Aid   int64  `json:"aid"`
		Count int64  `json:"count"`
		Toast string `json:"toast,omitempty"`
	}{aid, share, toast}, err)
	if err != nil {
		return
	}
	sendUserAct(c, mid, upID, aid, mobiApp, buvid, from, build, aidStr, "av", "share", device, platform, "", false, isReport)
}

//nolint:gomnd
func addCoin(c *bm.Context) {
	var (
		mid, aid, upID int64
		avType         int64
		actLike        = "cointolike"
		actCoin        = "coin"
	)
	params := c.Request.Form
	// check
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	platform := params.Get("platform")
	appkey := params.Get("appkey")
	buvid := c.Request.Header.Get("Buvid")
	from := params.Get("from")
	build := params.Get("build")
	aidStr := params.Get("aid")
	upIDStr := params.Get("upid")
	multiStr := params.Get("multiply")
	selectLikeStr := params.Get("select_like")
	spmid := params.Get("spmid")
	fromSpmid := params.Get("from_spmid")
	trackid := params.Get("track_id")
	goto_ := params.Get("goto")
	selectLike, _ := strconv.Atoi(selectLikeStr)
	aid, _ = strconv.ParseInt(aidStr, 10, 64)
	if aid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	upID, _ = strconv.ParseInt(upIDStr, 10, 64)
	multiply, err := strconv.ParseInt(multiStr, 10, 64)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if avType, _ = strconv.ParseInt(params.Get("avtype"), 10, 64); avType == 0 {
		avType = 1
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	prompt, likeReport, err := viewSvr.AddCoin(c, aid, mid, upID, avType, multiply, selectLike, buvid, platform, c.Request.URL.Path, appkey, c.Request.UserAgent(), build, mobiApp, device)
	// 投币引导逻辑
	var (
		guideType  = coinTypeNone
		guideTitle string
	)
	if bucket := int(mid % 10); bucket >= 0 && bucket <= 4 {
		guideType = coinTypeShare
		guideTitle = cfg.Resource.Coin.Title
	} else if prompt {
		guideType = coinTypeFollow
	}
	c.JSON(struct {
		Prompt bool          `json:"prompt,omitempty"`
		Like   bool          `json:"like"`
		Guide  *CoinAddGuide `json:"guide"`
	}{
		Prompt: prompt,
		Like:   likeReport,
		Guide: &CoinAddGuide{
			Type:  guideType,
			Title: guideTitle,
		},
	}, err)
	if err != nil && !ecode.EqualError(mainEcode.SilverBulletLikeReject, err) {
		return
	}
	// for ai big data
	var itemType = "archive"
	if avType == 2 {
		itemType = "article"
	}
	extraStr := extraJson(spmid, fromSpmid, trackid, goto_)
	sendUserAct(c, mid, upID, aid, mobiApp, buvid, from, build, aidStr, itemType, actCoin, device, platform, extraStr, ecode.EqualError(mainEcode.SilverBulletLikeReject, err), true)
	if selectLike == 1 {
		sendUserAct(c, mid, upID, aid, mobiApp, buvid, from, build, aidStr, "av", actLike, device, platform, extraStr, ecode.EqualError(mainEcode.SilverBulletLikeReject, err), likeReport)
	}
}

func like(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	// check
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	platform := params.Get("platform")
	appkey := params.Get("appkey")
	buvid := c.Request.Header.Get("Buvid")
	from := params.Get("from")
	build := params.Get("build")
	aidStr := params.Get("aid")
	ogvTypeStr := params.Get("ogv_type")
	ogvType, _ := strconv.ParseInt(ogvTypeStr, 10, 64)
	aid, _ := strconv.ParseInt(aidStr, 10, 64)
	spmid := params.Get("spmid")
	fromSpmid := params.Get("from_spmid")
	trackid := params.Get("track_id")
	goto_ := params.Get("goto")
	if aid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	like, err := strconv.Atoi(params.Get("like"))
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if like != 0 && like != 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	upperID, toast, err := viewSvr.Like(c, aid, mid, int8(like), ogvType, buvid, platform, c.Request.URL.Path, appkey, c.Request.UserAgent(), build, mobiApp, device)
	c.JSON(struct {
		Toast string `json:"toast"`
	}{Toast: toast}, err)
	if err != nil && !ecode.EqualError(mainEcode.SilverBulletLikeReject, err) {
		return
	}
	action := "like"
	if like == 1 {
		action = "like_cancel"
	}
	extraStr := extraJson(spmid, fromSpmid, trackid, goto_)
	// for ai big data
	sendUserAct(c, mid, upperID, aid, mobiApp, buvid, from, build, aidStr, "av", action, device, platform, extraStr, ecode.EqualError(mainEcode.SilverBulletLikeReject, err), true)
}

func dislike(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	platform := params.Get("platform")
	buvid := c.Request.Header.Get("Buvid")
	from := params.Get("from")
	build := params.Get("build")
	// check
	aidStr := params.Get("aid")
	spmid := params.Get("spmid")
	fromSpmid := params.Get("from_spmid")
	trackid := params.Get("track_id")
	goto_ := params.Get("goto")
	aid, _ := strconv.ParseInt(aidStr, 10, 64)
	if aid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	dislike, err := strconv.Atoi(params.Get("dislike"))
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if dislike != 0 && dislike != 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	upperID, err := viewSvr.Dislike(c, aid, mid, int8(dislike), mobiApp, device, platform)
	c.JSON(nil, err)
	if err != nil {
		return
	}
	// for ai big data
	action := "dislike"
	if dislike == 1 {
		action = "dislike_cancel"
	}
	extraStr := extraJson(spmid, fromSpmid, trackid, goto_)
	sendUserAct(c, mid, upperID, aid, mobiApp, buvid, from, build, aidStr, "av", action, device, platform, extraStr, false, true)
}

// adDislike ad dislike
// nolint:bilirailguncheck
func adDislike(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	header := c.Request.Header
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	gt := params.Get("goto")
	id, _ := strconv.ParseInt(params.Get("id"), 10, 64)
	reasonID, _ := strconv.ParseInt(params.Get("reason_id"), 10, 64)
	cmreasonID, _ := strconv.ParseInt(params.Get("cm_reason_id"), 10, 64)
	rid, _ := strconv.ParseInt(params.Get("rid"), 10, 64)
	tagID, _ := strconv.ParseInt(params.Get("tag_id"), 10, 64)
	adcb := params.Get("ad_cb")
	buvid := header.Get("Buvid")
	if buvid == "" && mid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	// for ad data
	err := dislikePub.Send(context.TODO(), strconv.FormatInt(mid, 10), &cmDislike{
		ID:         id,
		Buvid:      buvid,
		Goto:       gt,
		Mid:        mid,
		ReasonID:   reasonID,
		CMReasonID: cmreasonID,
		UpperID:    0,
		Rid:        rid,
		TagID:      tagID,
		ADCB:       adcb,
	})
	c.JSON(nil, err)
}

func compMeta(v *view.View, build int) {
	if build == 5800 || build == 508000 {
		for _, vp := range v.Pages {
			metas := make([]*view.Meta, 0, 4)
			for i, r := range _rate {
				meta := &view.Meta{
					Quality: i + 1,
					Size:    int64(float64(r*v.Duration) * 1.1 / 8.0),
					Format:  _formats[i+1],
				}
				metas = append(metas, meta)
			}
			vp.Metas = metas
		}
	}
}

// vipPlayURL get big-member token.
func vipPlayURL(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	params := c.Request.Form
	aid, _ := strconv.ParseInt(params.Get("aid"), 10, 64)
	if aid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	cid, _ := strconv.ParseInt(params.Get("cid"), 10, 64)
	if aid == 0 || cid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(viewSvr.VipPlayURL(c, aid, cid, mid))
}

// follow check if follow.
func follow(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	params := c.Request.Form
	vmid, err := strconv.ParseInt(params.Get("vmid"), 10, 64)
	if err != nil || vmid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(viewSvr.Follow(c, vmid, mid))
}

func upperRecmd(c *bm.Context) {
	var (
		mid    int64
		vmid   int64
		header = c.Request.Header
		params = c.Request.Form
	)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	platform := params.Get("platform")
	buildStr := params.Get("build")
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if vmid, err = strconv.ParseInt(params.Get("vmid"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.PlatNew(mobiApp, device)
	buvid := header.Get(_headerBuvid)
	data, err := viewSvr.UpperRecmd(c, plat, platform, mobiApp, device, buvid, build, mid, vmid)
	c.JSON(data, err)
}

func likeTriple(c *bm.Context) {
	var (
		mid       int64
		actTriple = "triplelike"
	)
	params := &view.TripleParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	buvid := c.Request.Header.Get("Buvid")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	triple, isReport, err := viewSvr.LikeTriple(c, params.AID, mid, buvid, params.Platform, c.Request.URL.Path, params.Appkey, c.Request.UserAgent(), params.Build, params.MobiApp, params.Device)
	c.JSON(triple, err)
	if err != nil && !ecode.EqualError(mainEcode.SilverBulletLikeReject, err) {
		return
	}
	extraStr := extraJson(params.Spmid, params.FromSpmid, params.TrackID, params.Goto)
	// for ai big data
	sendUserAct(c, mid, triple.UpID, params.AID, params.MobiApp, buvid, params.From, params.Build, strconv.FormatInt(params.AID, 10), "av", actTriple, params.Device, params.Platform, extraStr, ecode.EqualError(mainEcode.SilverBulletLikeReject, err), isReport)
}

// nolint:bilirailguncheck
func sendUserAct(ctx *bm.Context, mid, uid, aid int64, mobiApp, buvid, from, build, itemID, itemType, action, device, platform, extra string, isRisk, isReport bool) {
	isRiskInt := "0"
	if isRisk {
		isRiskInt = "1"
	}
	// send topic to AI
	_ = userActPub.Send(context.TODO(), strconv.FormatInt(mid, 10), &userAct{
		Client:   mobiApp,
		Buvid:    buvid,
		Mid:      mid,
		Time:     time.Now().Unix(),
		From:     from,
		Build:    build,
		ItemID:   itemID,
		ItemType: itemType,
		Action:   action,
		IsRisk:   isRiskInt,
	})
	if !isReport {
		return
	}
	// 报000078行为日志
	client := mobiApp
	if platform == "ios" {
		client = "iphone"
		if device == "pad" {
			client = "ipad"
		}
	} else if platform == "android" {
		client = "android"
		if mobiApp == "android_tv_yst" {
			client = "ott"
		}
	}
	infocItemType := itemType
	if action == "coin" && itemType == "av" {
		infocItemType = "archive"
	}
	sid := ""
	if csid, err := ctx.Request.Cookie("sid"); err == nil {
		sid = csid.Value
	}
	viewSvr.InfocV2(view.UserActInfoc{
		Buvid:    buvid,
		Build:    build,
		Client:   client,
		Ip:       metadata.String(ctx, metadata.RemoteIP),
		Uid:      uid,
		Aid:      aid,
		Mid:      mid,
		Sid:      sid,
		Url:      ctx.Request.URL.Path,
		From:     from,
		ItemID:   itemID,
		ItemType: infocItemType,
		Action:   action,
		Ua:       ctx.Request.UserAgent(),
		Ts:       strconv.FormatInt(time.Now().Unix(), 10),
		IsRisk:   isRiskInt,
		Extra:    extra,
	})
}

func material(c *bm.Context) {
	params := &view.MaterialParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(viewSvr.Material(c, params))
}

// shareClick when share click. 影响稿件计数
func shareClick(c *bm.Context) {
	params := &view.ShareParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	if params.AID <= 0 && params.OID <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	buvid := c.Request.Header.Get("Buvid")
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if params.OID <= 0 { // 兼容6.9版本前只传aid的情况
		params.OID = params.AID
		params.Type = model.ShareTypeAV
	} else if params.Type == model.ShareTypeAV { // 上报时需要保留aid因此做该转化
		params.AID = params.OID
	}
	var sid string
	if csid, err := c.Request.Cookie("sid"); err == nil {
		sid = csid.Value
	}
	_, _, err := viewSvr.ShareClick(c, params, mid, buvid, c.Request.UserAgent(), c.Request.URL.Path, c.Request.Referer(), sid)
	c.JSON(struct{}{}, err)
	if err != nil && !ecode.EqualError(mainEcode.SilverBulletLikeReject, err) {
		return
	}
}

// shareComplete when share complete 上报用
func shareComplete(c *bm.Context) {
	params := &view.ShareParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	buvid := c.Request.Header.Get("Buvid")
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if params.OID <= 0 { // 兼容6.9版本前只传aid的情况
		params.OID = params.AID
		params.Type = model.ShareTypeAV
	}
	toast, err := viewSvr.ShareComplete(c, params, mid, buvid)
	c.JSON(struct {
		Toast string `json:"toast,omitempty"`
	}{toast}, err)
}

func likeNoLogin(c *bm.Context) {
	params := &view.LikeNoLoginParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	buvid := c.Request.Header.Get("Buvid")
	if params.Like != 0 && params.Like != 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	upperID, res, err := viewSvr.LikeNoLogin(c, params.Aid, params.OgvType, params.Like, buvid, params.Platform, c.Request.URL.Path, params.Appkey, c.Request.UserAgent(), params.Build)
	c.JSON(res, err)
	if err != nil && !ecode.EqualError(mainEcode.SilverBulletLikeReject, err) {
		return
	}
	extraStr := extraJson(params.Spmid, params.FromSpmid, params.TrackID, params.Goto)
	action := params.Action
	// 当前已点赞状态本次操作为取消点赞
	if params.Like == 1 {
		action = "like_cancel"
	}
	// for ai big data
	sendUserAct(c, 0, upperID, params.Aid, params.MobiApp, buvid, params.From, params.Build, strconv.FormatInt(params.Aid, 10), "av", action, params.Device, params.Platform, extraStr, ecode.EqualError(mainEcode.SilverBulletLikeReject, err), true)
}

func extraJson(spmid, fromSpmid, trackid, goto_ string) string {
	return fmt.Sprintf(`{"spmid":"%s","from_spmid":"%s","track_id":"%s","goto":"%s"}`, spmid, fromSpmid, trackid, goto_)
}

func videoOnline(c *bm.Context) {
	params := &view.VideoOnlineParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	dev, _ := device.FromContext(c)
	params.Buvid = dev.Buvid
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	params.Mid = mid
	c.JSON(viewSvr.VideoOnline(c, params))
}

func videoDownload(ctx *bm.Context) {
	params := &viewApi.ShortFormVideoDownloadReq{}
	if err := ctx.Bind(params); err != nil {
		return
	}
	dev, _ := device.FromContext(ctx)
	params.Buvid = dev.Buvid
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	params.Mid = mid
	_, tfType := dev.TrafficFree()
	videoDownloadReq := &view.VideoDownloadReq{
		ShortFormVideoDownloadReq: params,
		TfType:                    tfType,
	}
	ctx.JSON(viewSvr.VideoDownload(ctx, videoDownloadReq))
}

func dmVote(ctx *bm.Context) {
	req := &view.DmVoteReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	midInter, _ := ctx.Get("mid")
	req.Mid = midInter.(int64)
	req.Buvid = ctx.Request.Header.Get("Buvid")
	ctx.JSON(viewSvr.DmVote(ctx, req))
}

func arcStat(ctx *bm.Context) {
	req := &view.StatReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(viewSvr.GetStatSrv(ctx, req))
}

//nolint:bilirailguncheck
func adDislikeCancel(c *bm.Context) {
	var mid int64
	params := c.Request.Form
	header := c.Request.Header
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	gt := params.Get("goto")
	id, _ := strconv.ParseInt(params.Get("id"), 10, 64)
	reasonID, _ := strconv.ParseInt(params.Get("reason_id"), 10, 64)
	cmreasonID, _ := strconv.ParseInt(params.Get("cm_reason_id"), 10, 64)
	rid, _ := strconv.ParseInt(params.Get("rid"), 10, 64)
	tagID, _ := strconv.ParseInt(params.Get("tag_id"), 10, 64)
	adcb := params.Get("ad_cb")
	buvid := header.Get("Buvid")
	if buvid == "" && mid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	// for ad data
	err := dislikePub.Send(context.TODO(), strconv.FormatInt(mid, 10), &cmDislike{
		ID:         id,
		Buvid:      buvid,
		Goto:       gt,
		Mid:        mid,
		ReasonID:   reasonID,
		CMReasonID: cmreasonID,
		UpperID:    0,
		Rid:        rid,
		TagID:      tagID,
		ADCB:       adcb,
		State:      1, //撤销
	})
	c.JSON(nil, err)
}

func addUserAction(ctx *bm.Context) {
	param := struct {
		Build    string `form:"build"`
		MobiApp  string `form:"mobi_app"`
		Uid      int64  `form:"uid"`
		Action   string `form:"action"`
		ItemID   string `form:"item_id"`
		ItemType string `form:"item_type"`
		Extra    string `form:"extra"`
	}{}
	if err := ctx.Bind(&param); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := ctx.Request.Header
	buvid := header.Get(_headerBuvid)
	viewSvr.InfocV2(view.UserActInfoc{
		Buvid:    buvid,
		Build:    param.Build,
		Client:   param.MobiApp,
		Ip:       metadata.String(ctx, metadata.RemoteIP),
		Uid:      param.Uid,
		Mid:      mid,
		Url:      ctx.Request.URL.Path,
		ItemID:   param.ItemID,
		ItemType: param.ItemType,
		Action:   param.Action,
		Ua:       ctx.Request.UserAgent(),
		Ts:       strconv.FormatInt(time.Now().Unix(), 10),
		Extra:    param.Extra,
	})
	ctx.JSON(nil, nil)
}

// shareInfo 分享落地页查询稿件信息
func shareInfo(c *bm.Context) {
	params := &share.InfoParam{}
	if err := c.Bind(params); err != nil {
		log.Error("shareInfo params(%+v), err(%+v)", params, err)
		return
	}

	if params.Bvid == "" {
		log.Error("shareInfo params(%+v), bad params", params)
		c.JSON(nil, ecode.RequestErr)
		return
	}

	c.JSON(viewSvr.ShareInfo(c, params))
}
