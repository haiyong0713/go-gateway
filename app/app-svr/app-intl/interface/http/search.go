package http

import (
	"strconv"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/text/translate/chinese.v2"

	"go-gateway/app/app-svr/app-intl/interface/model"
)

const _keyWordLen = 50

func searchAll(c *bm.Context) {
	var (
		build  int
		mid    int64
		pn, ps int
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
	clocale := params.Get("c_locale")
	slocale := params.Get("s_locale")
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
	if recommend != "1" {
		recommend = "0"
	}
	plat := model.Plat(mobiApp, device)
	isQuery, _ := strconv.Atoi(params.Get("is_org_query"))
	searchResults, err := searchSvc.Search(c, mid, mobiApp, device, platform, buvid, keyword, duration, order, filtered,
		lang, fromSource, recommend, plat, rid, highlight, build, pn, ps, isQuery, time.Now())
	if searchResults != nil && model.IsHant(clocale, slocale) {
		for _, sr := range searchResults.Item {
			var out map[string]string
			out = chinese.Converts(c, sr.Title, sr.Desc)
			sr.Title = out[sr.Title]
			sr.Desc = out[sr.Desc]
			// 稿件推荐理由
			for _, reason := range sr.NewRecTags {
				out := chinese.Converts(c, reason.Text)
				reason.Text = out[reason.Text]
			}
			// 用户卡及下挂视频
			if sr.OfficialVerify != nil {
				out = chinese.Converts(c, sr.OfficialVerify.Desc)
				sr.OfficialVerify.Desc = out[sr.OfficialVerify.Desc]
			}
			for _, userAV := range sr.AvItems {
				if userAV != nil {
					out = chinese.Converts(c, sr.Title, sr.Desc)
					sr.Title = out[sr.Title]
					sr.Desc = out[sr.Desc]
					// 稿件推荐理由
					for _, reason := range sr.NewRecTags {
						out := chinese.Converts(c, reason.Text)
						reason.Text = out[reason.Text]
					}
				}
			}
		}
	}
	c.JSON(searchResults, err)
}

func searchByType(c *bm.Context) {
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
	filtered := params.Get("filtered")
	order := params.Get("order")
	platform := params.Get("platform")
	highlightStr := params.Get("highlight")
	categoryIDStr := params.Get("category_id")
	userTypeStr := params.Get("user_type")
	orderSortStr := params.Get("order_sort")
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
	case "2":
		typeV = "upper"
	case "6":
		typeV = "article"
	case "7":
		typeV = "season2"
	case "8":
		typeV = "movie2"
	case "9":
		typeV = "tag"
	}
	plat := model.Plat(mobiApp, device)
	qn, _ := strconv.Atoi(params.Get("qn"))
	fnver, _ := strconv.Atoi(params.Get("fnver"))
	fnval, _ := strconv.Atoi(params.Get("fnval"))
	fourk, _ := strconv.Atoi(params.Get("fourk"))
	c.JSON(searchSvc.SearchByType(c, mid, mobiApp, device, platform, buvid, typeV, keyword, filtered, order, plat, build, highlight, categoryID, userType, orderSort, pn, ps, fnver, fnval, qn, fourk, time.Now()))
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
	c.JSON(searchSvc.Suggest3(c, mid, platform, buvid, term, build, highlight, mobiApp, time.Now()), nil)
}
