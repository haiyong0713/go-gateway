package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	model "go-gateway/app/app-svr/app-feed/admin/model/card"
)

func addNavigationCard(c *bm.Context) {
	var (
		err error
		req = &model.AddNavigationCardReq{}
	)
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	c.JSON(resourceCardSvc.AddNavigationCard(c, req))
	//nolint:gosimple
	return
}

func updateNavigationCard(c *bm.Context) {
	var (
		err error
		req = &model.UpdateNavigationCardReq{}
	)
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	c.JSON(resourceCardSvc.UpdateNavigationCard(c, req))
	//nolint:gosimple
	return
}

func deleteNavigationCard(c *bm.Context) {
	var (
		err error
		req = &model.DeleteNavigationCardReq{}
	)
	if err = c.BindWith(req, binding.Form); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	c.JSON(resourceCardSvc.DeleteNavigationCard(c, req))
	//nolint:gosimple
	return
}

func queryNavigationCard(c *bm.Context) {
	var (
		err error
		req = &model.QueryNavigationCardReq{}
	)
	if err = c.BindWith(req, binding.Form); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	c.JSON(resourceCardSvc.QueryNavigationCard(c, req))
	//nolint:gosimple
	return
}

func listNavigationCard(c *bm.Context) {
	var (
		err error
		req = &model.ListNavigationCardReq{}
	)
	if err = c.BindWith(req, binding.Form); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	c.JSON(resourceCardSvc.ListNavigationCard(c, req))
	//nolint:gosimple
	return
}

// content card
func addContentCard(c *bm.Context) {
	var (
		err error
		req = &model.AddContentCardReq{}
	)
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	c.JSON(resourceCardSvc.AddContentCard(c, req))
	//nolint:gosimple
	return
}

func updateContentCard(c *bm.Context) {
	var (
		err error
		req = &model.UpdateContentCardReq{}
	)
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	c.JSON(resourceCardSvc.UpdateContentCard(c, req))
	//nolint:gosimple
	return
}

func deleteContentCard(c *bm.Context) {
	var (
		err error
		req = &model.DeleteContentCardReq{}
	)
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	c.JSON(resourceCardSvc.DeleteContentCard(c, req))
	//nolint:gosimple
	return
}

func queryContentCard(c *bm.Context) {
	var (
		err error
		req = &model.QueryContentCardReq{}
	)
	if err = c.BindWith(req, binding.Form); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	c.JSON(resourceCardSvc.QueryContentCard(c, req))
	//nolint:gosimple
	return
}

func listContentCard(c *bm.Context) {
	var (
		err error
		req = &model.ListContentCardReq{}
	)
	if err = c.BindWith(req, binding.Form); err != nil {
		return
	}
	if req.Uid, req.Username, err = getUserInfo(c); err != nil {
		return
	}
	c.JSON(resourceCardSvc.ListContentCard(c, req))
	//nolint:gosimple
	return
}
