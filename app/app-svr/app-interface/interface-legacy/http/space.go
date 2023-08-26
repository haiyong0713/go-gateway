package http

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	acm "go-gateway/app/app-svr/app-interface/interface-legacy/model/account"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
	intlmdl "go-gateway/app/app-svr/app-intl/interface/model"
	xecode "go-gateway/ecode"
	"go-gateway/pkg/idsafe/bvid"
)

const (
	_crop      = "@750w_250h_1c"
	_cropBuild = 9150
	// search
	_headerBuvid = "Buvid"
	_keyWordLen  = 50
)

type userAct struct {
	Client   string `json:"client"`
	Buvid    string `json:"buvid"`
	Mid      int64  `json:"mid"`
	Time     int64  `json:"time"`
	From     string `json:"from"`
	Build    string `json:"build"`
	ItemID   string `json:"item_id"`
	ItemType string `json:"item_type"`
	Action   string `json:"action"`
	ActionID string `json:"action_id"`
	Extra    string `json:"extra"`
}

//nolint:bilirailguncheck
func spaceAll(c *bm.Context) {
	var (
		mid    int64
		vmid   int64
		build  int
		pn, ps int
		err    error
	)
	params := c.Request.Form
	header := c.Request.Header
	mobiApp := params.Get("mobi_app")
	platform := params.Get("platform")
	device := params.Get("device")
	buildStr := params.Get("build")
	name := params.Get("name")
	// check params
	if vmid, _ = strconv.ParseInt(params.Get("vmid"), 10, 64); vmid < 1 && name == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	fromViewAid, _ := strconv.ParseInt(params.Get("from_view_aid"), 10, 64)
	if build, err = strconv.Atoi(buildStr); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	plat := model.Plat(mobiApp, device)
	teenagersMode, _ := strconv.Atoi(params.Get("teenagers_mode"))
	lessonsMode, _ := strconv.Atoi(params.Get("lessons_mode"))
	buvid := header.Get("Buvid")
	adExtra := params.Get("ad_extra")
	spmid := params.Get("spmid")
	fromSpmid := params.Get("from_spmid")
	network := params.Get("network")
	filtered := params.Get("filtered")
	clocale := params.Get("clocale")
	slocale := params.Get("slocale")
	slocaleP := params.Get("s_locale")
	cLocaleP := params.Get("c_locale")
	isHant := (model.IsOverseas(plat) && intlmdl.IsHant(clocale, slocale)) || (i18n.PreferTraditionalChinese(c, slocaleP, cLocaleP))
	space, err := spaceSvr.Space(c, mid, vmid, plat, build, pn, ps, teenagersMode, lessonsMode, fromViewAid, platform, device, mobiApp, name, time.Now(), buvid, network, adExtra, spmid, fromSpmid, filtered, isHant)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if model.IsIPhone(plat) && space.Space != nil && space.Space.ImgURL != "" && build <= _cropBuild { // iPhone老版本
		space.Space.ImgURL = space.Space.ImgURL + _crop
	}
	space.Relation = compRealtion(space.Relation, mobiApp, build, device)
	c.JSON(space, nil)
	// for ai big data
	//nolint:errcheck
	userActPub.Send(context.TODO(), strconv.FormatInt(mid, 10), &userAct{
		Client:   mobiApp,
		Buvid:    buvid,
		Mid:      mid,
		Time:     time.Now().Unix(),
		From:     params.Get("from"),
		Build:    buildStr,
		ItemID:   space.Card.Mid,
		ItemType: "mid",
		Action:   "space",
	})
}

func playGame(c *bm.Context) {
	var (
		mid int64
		arg = new(struct {
			Vmid     int64  `form:"vmid" validate:"min=1"`
			Pn       int    `form:"pn" default:"1" validate:"min=1"`
			Ps       int    `form:"ps" default:"15" validate:"min=1,max=20"`
			Platform string `form:"platform" validate:"required"`
		})
	)
	if err := c.Bind(arg); err != nil {
		return
	}
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	c.JSON(spaceSvr.PlayGamesSub(c, mid, arg.Vmid, nil, arg.Pn, arg.Ps, arg.Platform), nil)
}

func upArchive(c *bm.Context) {
	var (
		mid           int64
		pn, ps, build int
		err           error
	)
	params := c.Request.Form
	// check params
	vmid, _ := strconv.ParseInt(params.Get("vmid"), 10, 64)
	if vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	if build, err = strconv.Atoi(buildStr); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	order := params.Get("order")
	if order == "" {
		order = space.ArchiveNew
	}
	clocale := params.Get("clocale")
	slocale := params.Get("slocale")
	slocaleP := params.Get("s_locale")
	clocaleP := params.Get("c_locale")
	isHant := (model.IsOverseas(plat) && intlmdl.IsHant(clocale, slocale)) || (i18n.PreferTraditionalChinese(c, slocaleP, clocaleP))
	res, err := spaceSvr.UpArcs(c, mobiApp, device, mid, vmid, pn, ps, build, plat, true, order, isHant, nil)
	if err != nil {
		log.Error("%+v", err)
		if !ecode.EqualError(ecode.NothingFound, err) {
			err = ecode.Degrade
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(res, nil)
}

func upArchiveCursor(c *bm.Context) {
	param := &space.ArchiveCursorParam{}
	if err := c.Bind(param); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if param.Order == "" {
		param.Order = space.ArchiveNew
	}
	if param.Ps < 1 || param.Ps > 20 {
		param.Ps = 20
	}
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	isHant := model.IsOverseas(plat) && i18n.PreferTraditionalChinese(c, param.SLocaleP, param.CLocaleP)
	isIpad := model.IsIPad(plat)
	res, err := spaceSvr.UpArcsCursor(c, param, mid, isHant, isIpad)
	if err != nil {
		log.Error("%+v", err)
		if !ecode.EqualError(ecode.NothingFound, err) {
			err = ecode.Degrade
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(res, nil)
}

func upSeries(c *bm.Context) {
	params := &space.SeriesParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	if params.Vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if params.Ps < 1 || params.Ps > 20 {
		params.Ps = 20
	}
	plat := model.Plat(params.MobiApp, params.Device)
	isHant := (model.IsOverseas(plat) && intlmdl.IsHant(params.Clocale, params.Slocale)) || (i18n.PreferTraditionalChinese(c, params.SLocaleP, params.CLocaleP))
	seriesReply, err := spaceSvr.UpSeries(c, params.Vmid, params.SeriesId, params.Ps, params.Next, params.Sort, true, isHant)
	c.JSON(seriesReply, err)
}

func seasonView(ctx *bm.Context) {
	params := &space.SeasonArchiveParam{}
	if err := ctx.Bind(params); err != nil {
		return
	}
	plat := model.Plat(params.MobiApp, params.Device)
	isHant := (model.IsOverseas(plat) && intlmdl.IsHant(params.Clocale, params.Slocale)) || i18n.PreferTraditionalChinese(ctx, params.SLocaleP, params.CLocaleP)
	ctx.JSON(spaceSvr.SeasonArchiveList(ctx, isHant, params))
}

func upComic(c *bm.Context) {
	var (
		pn, ps, build int
		err           error
	)
	params := c.Request.Form
	// check params
	vmid, _ := strconv.ParseInt(params.Get("vmid"), 10, 64)
	if vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	if build, err = strconv.Atoi(buildStr); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	c.JSON(spaceSvr.UpComics(c, vmid, pn, ps, build, time.Now(), plat, true), nil)
}

func subComic(c *bm.Context) {
	var (
		mid           int64
		pn, ps, build int
		err           error
	)
	params := c.Request.Form
	// check params
	vmid, _ := strconv.ParseInt(params.Get("vmid"), 10, 64)
	if vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 24 {
		ps = 24
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	if build, err = strconv.Atoi(buildStr); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	plat := model.Plat(mobiApp, device)
	c.JSON(spaceSvr.SubComics(c, mid, vmid, nil, pn, ps, build, plat), nil)
}

func upSeason(c *bm.Context) {
	var (
		pn, ps int64
		build  int
		err    error
	)
	params := c.Request.Form
	// check params
	vmid, _ := strconv.ParseInt(params.Get("vmid"), 10, 64)
	if vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, _ = strconv.ParseInt(params.Get("pn"), 10, 64); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.ParseInt(params.Get("ps"), 10, 64); ps < 1 || ps > 20 {
		ps = 20
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	if build, err = strconv.Atoi(buildStr); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	c.JSON(spaceSvr.UpSeasons(c, vmid, pn, ps, build, time.Now(), plat, mobiApp), nil)
}

func myinfo(c *bm.Context) {
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	build, _ := strconv.Atoi(params.Get("build"))
	plat := model.Plat(mobiApp, device)
	mid, err := authSvc.AuthToken(c)
	if err != nil {
		shouldChangeError := false
		if (model.IsIPhone(plat) && build > config.LoginBuild.Iphone) || (model.IsAndroidPick(plat) && build >= config.LoginBuild.Android) ||
			(model.IsAndroidB(plat) && build >= config.LoginBuild.AndroidB) || (model.IsIPhoneB(plat) && build >= config.LoginBuild.IphoneB) ||
			(model.IsIPadHD(plat) && build >= config.LoginBuild.IpadHD) {
			shouldChangeError = true
		}
		if shouldChangeError && ecode.EqualError(ecode.NoLogin, err) {
			c.JSON(nil, xecode.UserLoginInvalid)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(accSvr.Myinfo(c, mid, mobiApp))
}

func mine(c *bm.Context) {
	params := &space.MineParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.Mid = midInter.(int64)
	}
	buvid := c.Request.Header.Get(_headerBuvid)
	plat := model.Plat(params.MobiApp, params.Device)
	if model.IsOverseas(plat) {
		params.Lang = "hant"
	}
	c.JSON(accSvr.Mine(c, params.Mid, params.Platform, params.Lang, params.Filtered, params.Channel, params.MobiApp, params.Build, params.TeenagersMode, params.LessonsMode, plat, params.Device, params.SLocale, params.CLocale, buvid, params.BiliLinkNew))
}

func mineIpad(c *bm.Context) {
	params := &space.MineParam{}
	if err := c.Bind(params); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.Mid = midInter.(int64)
	}
	buvid := c.Request.Header.Get(_headerBuvid)
	plat := model.Plat(params.MobiApp, params.Device)
	c.JSON(accSvr.MineIpad(c, params.Mid, params.MobiApp, params.Platform, params.Lang, params.Filtered, params.Channel, params.SLocale, params.CLocale, params.Build, params.TeenagersMode, params.LessonsMode, plat, buvid))
}

func configSet(c *bm.Context) {
	params := &struct {
		Buvid        string `form:"buvid" validate:"required"`
		AdSpecial    int    `form:"ad_special"`
		SensorAccess int    `form:"sensor_access"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	c.JSON(nil, accSvr.ConfigSet(c, mid, params.Buvid, params.AdSpecial, params.SensorAccess))
}

func upArticle(c *bm.Context) {
	var (
		pn, ps int
	)
	params := c.Request.Form
	// check params
	vmid, _ := strconv.ParseInt(params.Get("vmid"), 10, 64)
	if vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	c.JSON(spaceSvr.UpArticles(c, vmid, pn, ps), nil)
}

func contribute(c *bm.Context) {
	var (
		vmid, mid int64
		build     int
		pn, ps    int
		err       error
	)
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	// check params
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if vmid, _ = strconv.ParseInt(params.Get("vmid"), 10, 64); vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	plat := model.Plat(mobiApp, device)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spaceSvr.Contribute(c, plat, build, vmid, pn, ps, time.Now(), mobiApp, device, mid))
}

func contribution(c *bm.Context) {
	var (
		vmid, mid    int64
		build        int
		maxID, minID int64
		size         int
		err          error
	)
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	maxIDStr := params.Get("max_id")
	minIDStr := params.Get("min_id")
	// check params
	if build, err = strconv.Atoi(params.Get("build")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if vmid, _ = strconv.ParseInt(params.Get("vmid"), 10, 64); vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if maxIDStr != "" {
		if maxID, err = strconv.ParseInt(maxIDStr, 10, 64); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if minIDStr != "" {
		if minID, err = strconv.ParseInt(minIDStr, 10, 64); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if size, _ = strconv.Atoi(params.Get("size")); size < 1 || size > 20 {
		size = 20
	}
	cursor, err := model.NewCursor(maxID, minID, size)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		log.Error("%+v", err)
		return
	}
	plat := model.Plat(mobiApp, device)
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spaceSvr.Contribution(c, plat, build, vmid, cursor, time.Now(), mobiApp, device, mid))
}

func upContribute(c *bm.Context) {
	var (
		vmid                   int64
		attrs                  *space.Attrs
		items                  []*space.Item
		isCooperation, isComic bool
		err                    error
	)
	params := c.Request.Form
	vmidStr := params.Get("vmid")
	attrsStr := params.Get("attrs")
	itemsStr := params.Get("items")
	// check params
	if vmid, _ = strconv.ParseInt(vmidStr, 10, 64); vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if err = json.Unmarshal([]byte(attrsStr), &attrs); err != nil {
		c.JSON(nil, ecode.RequestErr)
		log.Error("json.Unmarshal(%s) error(%v)", attrsStr, err)
		return
	}
	if err = json.Unmarshal([]byte(itemsStr), &items); err != nil {
		c.JSON(nil, ecode.RequestErr)
		log.Error("json.Unmarshal(%s) error(%v)", itemsStr, err)
		return
	}
	isCooperation, _ = strconv.ParseBool(params.Get("is_cooperation"))
	isComic, _ = strconv.ParseBool(params.Get("is_comic"))
	c.JSON(spaceSvr.AddContribute(c, vmid, attrs, items, isCooperation, isComic), nil)
}

func bangumi(c *bm.Context) {
	var (
		mid    int64
		pn, ps int
	)
	params := c.Request.Form
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	// check params
	vmid, _ := strconv.ParseInt(params.Get("vmid"), 10, 64)
	if vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	c.JSON(spaceSvr.Bangumi(c, mid, vmid, nil, pn, ps), nil)
}

func community(c *bm.Context) {
	var (
		pn, ps int
	)
	params := c.Request.Form
	ak := params.Get("access_key")
	platform := params.Get("platform")
	// check params
	vmid, _ := strconv.ParseInt(params.Get("vmid"), 10, 64)
	if vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	c.JSON(spaceSvr.Community(c, vmid, pn, ps, ak, platform), nil)
}

func coinCancel(c *bm.Context) {
	var (
		mid int64
		arg = new(struct {
			Aid int64 `form:"aid" validate:"min=1"`
		})
	)
	if err := c.BindWith(arg, binding.Form); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(nil, spaceSvr.CoinCancel(c, arg.Aid, mid))
}

func coinArc(c *bm.Context) {
	var (
		mid    int64
		pn, ps int
	)
	params := c.Request.Form
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	// check params
	vmid, _ := strconv.ParseInt(params.Get("vmid"), 10, 64)
	if vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	clocale := params.Get("c_locale")
	slocale := params.Get("s_locale")
	isHant := (model.IsOverseas(plat) && intlmdl.IsHant(clocale, slocale)) || i18n.PreferTraditionalChinese(c, slocale, clocale)
	c.JSON(spaceSvr.CoinArcs(c, mid, vmid, nil, pn, ps, isHant, mobiApp, device), nil)
}

func likeArc(c *bm.Context) {
	var (
		mid    int64
		pn, ps int
	)
	params := c.Request.Form
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	// check params
	vmid, _ := strconv.ParseInt(params.Get("vmid"), 10, 64)
	if vmid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn, _ = strconv.Atoi(params.Get("pn")); pn < 1 {
		pn = 1
	}
	if ps, _ = strconv.Atoi(params.Get("ps")); ps < 1 || ps > 20 {
		ps = 20
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	clocale := params.Get("c_locale")
	slocale := params.Get("s_locale")
	isHant := (model.IsOverseas(plat) && intlmdl.IsHant(clocale, slocale)) || i18n.PreferTraditionalChinese(c, slocale, clocale)
	c.JSON(spaceSvr.LikeArcs(c, mid, vmid, nil, pn, ps, isHant, mobiApp, device), nil)
}

func upperRecmd(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	v := new(space.FollowParam)
	if err := c.Bind(v); err != nil {
		return
	}
	plat := model.Plat(v.MobiApp, v.Device)
	buvid := c.Request.Header.Get(_headerBuvid)
	c.JSON(spaceSvr.UpperRecmd(c, plat, v.Platform, v.MobiApp, v.Device, buvid, v.Build, mid, v.Vmid))
}

func report(c *bm.Context) {
	params := c.Request.Form
	ak := params.Get("access_key")
	reason := params.Get("reason")
	mid, _ := strconv.ParseInt(params.Get("mid"), 10, 64)
	if mid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, spaceSvr.Report(c, mid, reason, ak))
}

func clips(c *bm.Context) {
	c.JSON(space.ClipList{Item: []*space.Item{}}, nil)
}

func albums(c *bm.Context) {
	c.JSON(space.AlbumList{Item: []*space.Item{}}, nil)
}

// checkPay pay movie or bangumi
func compRealtion(rel int, mobiApp string, build int, _ string) (r int) {
	const (
		_upAndroid = 505000
		_banIphone = 5550
		_banIPad   = 10450
	)
	switch mobiApp {
	case "android", "android_G":
		if build < _upAndroid && rel == -1 {
			return -999
		}
	case "iphone", "iphone_G":
		if build < _banIphone && rel == -1 {
			return -999
		}
	case "ipad", "ipad_G":
		if build <= _banIPad && rel == -1 {
			return -999
		}
	}
	return rel
}

func spaceSearch(ctx *bm.Context) {
	params := ctx.Request.Form
	keyword := params.Get("keyword")
	isTitle, _ := strconv.Atoi(params.Get("is_title"))
	highlight, _ := strconv.Atoi(params.Get("highlight"))
	vmid, _ := strconv.ParseInt(params.Get("vmid"), 10, 64)
	if vmid < 1 {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	if keyword == "" || len([]rune(keyword)) > _keyWordLen {
		ctx.JSON(nil, ecode.RequestErr)
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
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	isIpad := model.IsIPad(plat)
	res, err := srcSvr.SpaceSearch(ctx, vmid, keyword, highlight == 1, isTitle == 1, isIpad, pn, ps)
	if err != nil {
		log.Error("%+v", err)
		if !ecode.EqualError(ecode.NothingFound, err) {
			err = ecode.Degrade
		}
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(res, nil)
}

func attentionMark(c *bm.Context) {
	var mid int64
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	c.JSON(nil, spaceSvr.AttentionMark(c, mid))
}

func photoMallList(c *bm.Context) {
	var mid int64
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := &space.PhotoTopParm{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(spaceSvr.PhotoMallList(c, params, mid))
}

func photoTopSet(c *bm.Context) {
	var mid int64
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := &space.PhotoTopParm{}
	if err := c.Bind(params); err != nil {
		return
	}
	// type av
	var err error
	//nolint:gomnd
	switch params.Type {
	case 1:
		params.Oid, err = strconv.ParseInt(params.ID, 10, 64)
	case 2:
		if strings.HasPrefix(params.ID, "BV1") {
			params.Oid, err = bvid.BvToAv(params.ID)
		} else {
			params.Oid, err = strconv.ParseInt(params.ID, 10, 64)
		}
	}
	if err != nil || params.Oid <= 0 {
		log.Warn("photoTopSet params:%+v not allowed", params)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, spaceSvr.PhotoTopSet(c, params, mid))
}

func photoArcList(c *bm.Context) {
	v := new(struct {
		Pn       int    `form:"pn" default:"1" validate:"min=1"`
		Ps       int    `form:"ps" default:"10" validate:"min=1,max=20"`
		MobiApp  string `form:"mobi_app"`
		Device   string `form:"device"`
		Platform string `form:"platform"`
		Build    int    `form:"build"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	plat := model.Plat(v.MobiApp, v.Device)
	var mid int64
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	buvid := c.Request.Header.Get("Buvid")
	c.JSON(spaceSvr.PhotoArcList(c, v.MobiApp, v.Platform, v.Device, buvid, mid, plat, v.Build, v.Pn, v.Ps))
}

func nftSettingButton(c *bm.Context) {
	var req = &acm.NFTSettingButtonReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInt, ok := c.Get("mid"); ok {
		mid = midInt.(int64)
	}
	req.Mid = mid
	dev, _ := device.FromContext(c)
	req.MobiApp = dev.RawMobiApp
	c.JSON(accSvr.NFTSettingButton(c, req))
}
