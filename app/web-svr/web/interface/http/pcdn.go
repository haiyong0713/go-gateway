package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-gateway/app/web-svr/web/interface/model"

	pcdnVerifygrpc "git.bilibili.co/bapis/bapis-go/vas/pcdn/verify/service"
)

// 加入pcdn流量计划
func joinPCDN(c *bm.Context) {
	var (
		mid int64
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(nil, webSvc.JoinPCDN(c, mid))
}

// 开启/设置流量档位
func operatePCDN(c *bm.Context) {
	var (
		err error
	)
	req := &model.OperatePCDNReq{}
	if err = c.Bind(req); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		req.Mid = midStr.(int64)
	}

	c.JSON(nil, webSvc.OperatePCDN(c, req))
}

// 查询用户对pcdn的设置
func userSettings(c *bm.Context) {
	var (
		mid int64
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(webSvc.UserSettings(c, mid))
}

// 查询用户资产信息
func userAccountInfo(c *bm.Context) {
	var (
		mid int64
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	rs, err := webSvc.UserAccountInfo(c, mid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

// 兑换资产
func exchange(c *bm.Context) {
	var (
		err error
	)
	req := &model.PcdnRewardExchangeReq{}
	if err = c.Bind(req); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		req.Mid = midStr.(int64)
	}
	c.JSON(nil, webSvc.Exchange(c, req))
}

// pcdn用聚合接口
func pcdnV1(c *bm.Context) {
	var (
		err error
	)
	req := &model.PcdnV1Req{}
	if err = c.Bind(req); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		req.Mid = midStr.(int64)
	}
	rs, err := webSvc.PcdnV1(c, req)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

// pcdn上报接口
func pcdnReport(c *bm.Context) {
	var (
		err error
	)
	v := new(pcdnVerifygrpc.ReportGlobalInfo)
	if err = c.BindWith(v, binding.JSON); err != nil {
		return
	}

	c.JSON(nil, webSvc.ReportV1(c, v))
}

// 小黄条通知
func notify(c *bm.Context) {
	var (
		mid int64
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(webSvc.Notification(c, mid))
}

// 数字藏品页面
func digitialCollection(c *bm.Context) {
	var (
		mid int64
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(webSvc.DigitialCollection(c, mid))
}

// 退出pcdn
func quit(c *bm.Context) {
	var (
		mid int64
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(nil, webSvc.Quit(c, mid))
}

// pcdn 页面接口
func pacnPages(c *bm.Context) {
	var (
		mid int64
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(webSvc.PcdnPages(c, mid), nil)
}

// // 数字藏品兑换
// func digitialCollection(c *bm.Context) {
// 	var (
// 		mid int64
// 	)
// 	if midStr, ok := c.Get("mid"); ok {
// 		mid = midStr.(int64)
// 	}
// 	c.JSON(webSvc.DigitialCollection(c, mid))
// }
