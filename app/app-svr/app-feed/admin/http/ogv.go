package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
)

// addOgv add ogv
func addOgv(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.SearchOgvAP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if usernameCtx, ok := c.Get("username"); ok {
		req.Person = usernameCtx.(string)
	}
	if err = searchSvc.OgvAdd(c, req); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// updateOgv update ogv
func updateOgv(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.SearchOgvUP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if usernameCtx, ok := c.Get("username"); ok {
		req.Person = usernameCtx.(string)
	}
	if err = searchSvc.OgvUpdate(c, req); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// optOgv option ogv
func optOgv(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.SearchOgvOption{}
	if err = c.Bind(req); err != nil {
		return
	}
	if usernameCtx, ok := c.Get("username"); ok {
		req.Person = usernameCtx.(string)
	}
	if err = searchSvc.OgvOpt(c, req); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// ogvList ogv list
func ogvList(c *bm.Context) {
	var (
		err   error
		pager *show.SearchOgvPager
	)
	res := map[string]interface{}{}
	req := &show.SearchOgvLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if pager, err = searchSvc.OgvList(c, req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

// openOgv .
func openOgv(c *bm.Context) {
	var (
		err   error
		pager *show.SearchOgvPager
	)
	res := map[string]interface{}{}
	req := &show.SearchOgvLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	req.Check = common.Valid
	if pager, err = searchSvc.OgvList(c, req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}
