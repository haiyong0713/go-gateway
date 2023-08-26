package http

import (
	"strconv"
	"strings"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/history"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
)

// history list
func historyList(c *bm.Context) {
	param := &history.HisParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Pn < 1 {
		param.Pn = 1
	}
	if param.Ps > 20 || param.Ps <= 0 {
		param.Ps = 20
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	plat := model.Plat(param.MobiApp, param.Device)
	c.JSON(historySvr.List(c, param.Mid, param.Build, param.Pn, param.Ps, param.MobiApp, plat))
}

// shortAll get shorturl list
func live(c *bm.Context) {
	param := &history.LiveParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	roomIDs, err := xstr.SplitInts(param.RoomIDs)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(historySvr.Live(c, roomIDs))
}

// history list
func liveList(c *bm.Context) {
	param := &history.HisParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Pn < 1 {
		param.Pn = 1
	}
	if param.Ps > 20 || param.Ps <= 0 {
		param.Ps = 20
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	plat := model.Plat(param.MobiApp, param.Device)
	c.JSON(historySvr.LiveList(c, param, plat))
}

// history cursor
func historyCursor(c *bm.Context) {
	param := &history.HisParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Ps > 20 || param.Ps <= 0 {
		param.Ps = 20
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	plat := model.Plat(param.MobiApp, param.Device)
	res, _, err := historySvr.Cursor(c, param, 50, plat, false, nil)
	c.JSON(res, err)
}

// history del
func historyDel(c *bm.Context) {
	param := &history.DelParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var hisRes []*hisApi.ModelHistory
	for _, boid := range param.Boids {
		bo := strings.Split(boid, "_")
		//nolint:gomnd
		if len(bo) != 2 {
			log.Error("historyDel invalid param(%+v)", param)
			c.JSON(nil, ecode.RequestErr)
			return
		}
		oid, _ := strconv.ParseInt(bo[1], 10, 0)
		if oid == 0 {
			log.Error("historyDel invalid param(%+v)", param)
			c.JSON(nil, ecode.RequestErr)
			return
		}
		hisRes = append(hisRes, &hisApi.ModelHistory{
			Aid:      oid,
			Business: bo[0],
		})
	}
	c.JSON(nil, historySvr.Del(c, param.Mid, hisRes, dev))
}

// history clear
func historyClear(c *bm.Context) {
	param := &history.HisParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	dev, ok := device.FromContext(c)
	if !ok {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	c.JSON(nil, historySvr.Clear(c, param, plat, dev))
}
