package http

import (
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-show/interface/model"
)

func dailyID(c *bm.Context) {
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	mobiApp = model.MobiAPPBuleChange(mobiApp)
	device := params.Get("device")
	buildStr := params.Get("build")
	pnStr := params.Get("pn")
	psStr := params.Get("ps")
	dailyIDStr := params.Get("daily_id")
	dailyID, err := strconv.Atoi(dailyIDStr)
	if err != nil {
		log.Error("dailyID(%s) error(%v)", dailyIDStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(mobiApp, device)
	pn, err := strconv.Atoi(pnStr)
	if err != nil || pn < 1 {
		pn = 1
	}
	ps, err := strconv.Atoi(psStr)
	if err != nil || ps > 60 || ps <= 0 {
		ps = 60
	}
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		log.Error("build(%s) error(%v)", buildStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	data := dailySvc.Daily(c, plat, build, dailyID, pn, ps)
	returnJSON(c, data, nil)
}

// 经确认双端均已下线代码，且监控无流量
func columnList(c *bm.Context) {
	returnJSON(c, nil, ecode.NothingFound)
}

// 经确认双端均已下线代码，且监控无流量
func category(c *bm.Context) {
	returnJSON(c, nil, ecode.NothingFound)
}
