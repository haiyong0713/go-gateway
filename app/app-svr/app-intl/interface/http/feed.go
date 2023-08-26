package http

import (
	"strconv"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-intl/interface/model"
	"go-gateway/app/app-svr/app-intl/interface/model/feed"
)

const (
	_headerBuvid    = "Buvid"
	_headerDeviceID = "Device-ID"
)

func feedIndex(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	dvcid := header.Get(_headerDeviceID)
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
	column, ok := cdm.Columnm[param.Column]
	if !ok {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	// 兼容老的style逻辑，3为新单列
	style := int(cdm.Columnm[param.Column])
	if style == 1 {
		style = 3
	}
	// check params
	plat := model.Plat(param.MobiApp, param.Device)
	// get audit data, if check audit hit.
	if data, ok := feedSvc.Audit(c, param.MobiApp, plat, param.Build, param.Column); ok {
		c.JSON(struct {
			Item []card.Handler `json:"items"`
		}{Item: data}, nil)
		return
	}
	now := time.Now()
	// index
	data, userFeature, isRcmd, newUser, code, autoPlay, feedclean, autoPlayInfoc, err := feedSvc.Index(c, buvid, mid, plat, param, now, style)
	autoplayCard := struct {
		Column          cdm.ColumnStatus `json:"column"`
		AutoplayCard    int8             `json:"autoplay_card"`
		FeedCleanAbtest int8             `json:"feed_clean_abtest"`
	}{Column: column, AutoplayCard: autoPlay, FeedCleanAbtest: feedclean}
	if afv, ok := feedSvc.FawkesVersionCache[fawkesEnv]; ok {
		if fv, ok := afv[fawkesAppkey]; ok {
			c.Writer.Header().Set("CONFIG-V", strconv.FormatInt(fv.Config, 10))
			c.Writer.Header().Set("FF-V", strconv.FormatInt(fv.FF, 10))
		}
	}
	c.JSON(struct {
		Item   []card.Handler `json:"items"`
		Config interface{}    `json:"config"`
	}{Item: data, Config: autoplayCard}, err)
	if err != nil {
		return
	}
	// infoc
	items := make([]*ai.Item, 0, len(data))
	for _, item := range data {
		items = append(items, item.Get().Rcmd)
	}
	feedSvc.IndexInfoc(c, mid, plat, param.Build, buvid, "/x/intl/feed/index", userFeature, style, code, items, isRcmd, param.Pull, newUser, now, "", dvcid, param.Network, param.Flush, autoPlayInfoc, param.DeviceType, param.Slocale, param.Clocale, param.Timezone, param.Lang, param.SimCode)
}

func feedIndexTab(c *bm.Context) {
	var (
		id      int64
		items   []card.Handler
		isBnj   bool
		bnjDays int
		cover   string
		err     error
		mid     int64
	)
	params := c.Request.Form
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	now := time.Now()
	idStr := params.Get("id")
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	buildStr := params.Get("build")
	// check params
	build, err := strconv.Atoi(buildStr)
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
	items, cover, isBnj, bnjDays, err = feedSvc.Actives(c, id, mid, mobiApp, device, plat, build, now)
	c.JSON(struct {
		Cover   string         `json:"cover"`
		IsBnj   bool           `json:"is_bnj,omitempty"`
		BnjDays int            `json:"bnj_days,omitempty"`
		Item    []card.Handler `json:"items"`
	}{cover, isBnj, bnjDays, items}, err)
}
