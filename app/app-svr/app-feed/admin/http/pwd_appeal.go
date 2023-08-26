package http

import (
	bm "go-common/library/net/http/blademaster"

	model "go-gateway/app/app-svr/app-feed/admin/model/pwd_appeal"
)

func pwdAppealList(c *bm.Context) {
	req := &model.ListReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(pwdAppealSvc.List(c, req))
}

func pwdAppealPhoto(c *bm.Context) {
	req := &model.PhotoReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	photo, err := pwdAppealSvc.Photo(c, req)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if _, err := c.Writer.Write(photo); err != nil {
		c.JSON(nil, err)
	}
}

func passPwdAppeal(c *bm.Context) {
	req := &model.PassReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	var userid int64
	if uid, ok := c.Get("uid"); ok {
		userid = uid.(int64)
	}
	c.JSON(nil, pwdAppealSvc.Pass(c, req, userid, username))
}

func rejectPwdAppeal(c *bm.Context) {
	req := &model.RejectReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	var userid int64
	if uid, ok := c.Get("uid"); ok {
		userid = uid.(int64)
	}
	c.JSON(nil, pwdAppealSvc.Reject(c, req, userid, username))
}

func exportPwdAppeal(c *bm.Context) {
	req := &model.ExportReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	rly, err := pwdAppealSvc.Export(c, req)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.Writer.Header().Set("Content-Type", "application/csv")
	c.Writer.Header().Set("Content-Disposition", "attachment;filename=appeal.csv")
	if _, err := c.Writer.Write(rly); err != nil {
		c.JSON(nil, err)
	}
}
