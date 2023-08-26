package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/admin/model"
)

// updateOfficial .
func updateOfficial(c *bm.Context) {
	var err error
	v := &model.SpaceOfficial{}
	if err = c.Bind(v); err != nil {
		return
	}
	if err = spcSvc.UpdateOfficial(v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "更新失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// addOfficial .
func addOfficial(c *bm.Context) {
	var err error
	v := &model.SpaceOfficial{}
	if err = c.Bind(v); err != nil {
		return
	}
	if err = spcSvc.AddOfficial(v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// delOfficial .
func delOfficial(c *bm.Context) {
	var (
		err error
	)
	req := &struct {
		ID int64 `form:"id" validate:"min=1"`
	}{}
	if err = c.Bind(req); err != nil {
		return
	}
	if err = spcSvc.DeleteOfficial(req.ID); err != nil {
		res := map[string]interface{}{}
		res["message"] = "删除失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// officialIndex .
func officialIndex(c *bm.Context) {
	req := &model.SpaceOfficialParam{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(spcSvc.Official(req))
}
