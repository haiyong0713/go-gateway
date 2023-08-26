package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/admin/model/search"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

func searchBrandBlacklistAdd(c *bm.Context) {
	var (
		err error
		req = &search.BrandBlacklistAddReq{}
	)
	if err = c.Bind(req); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	resp, err := searchSvc.BrandBlacklistAdd(c, req)
	c.JSON(resp, err)
	//nolint:gosimple
	return
}

func searchBrandBlacklistEdit(c *bm.Context) {
	var (
		err error
		req = &search.BrandBlacklistEditReq{}
	)
	if err = c.Bind(req); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	resp, err := searchSvc.BrandBlacklistEdit(c, req)
	c.JSON(resp, err)
	//nolint:gosimple
	return
}

func searchBrandBlacklistOption(c *bm.Context) {
	var (
		err error
		req = &search.BrandBlacklistOptionReq{}
	)
	if err = c.Bind(req); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	resp, err := searchSvc.BrandBlacklistOption(c, req)
	c.JSON(resp, err)
	//nolint:gosimple
	return
}

func searchBrandBlacklistList(c *bm.Context) {
	var (
		err error
		req = &search.BrandBlacklistListReq{}
	)
	if err = c.Bind(req); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	resp, err := searchSvc.BrandBlacklistList(c, req)
	c.JSON(resp, err)
	//nolint:gosimple
	return
}

func openSearchBrandBlacklist(c *bm.Context) {
	var (
		err error
		req = &search.BrandBlacklistListReq{}
	)
	if err = c.Bind(req); err != nil {
		return
	}
	resp, err := searchSvc.OpenBrandBlacklistList(c, req)
	c.JSON(resp, err)
	//nolint:gosimple
	return
}

func getUserInfo(c *bm.Context) (uid int64, username string, err error) {
	uid, username = util.UserInfo(c)
	if username == "" {
		err = ecode.Unauthorized
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, err)
		c.Abort()
		return
	}
	return
}
