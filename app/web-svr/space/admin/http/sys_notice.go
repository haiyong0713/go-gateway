package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/admin/model"
)

func addSysNotice(c *bm.Context) {
	v := &model.SysNoticeAdd{}
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, spcSvc.SysNoticeAdd(c, v))
}

func sysNoticeList(c *bm.Context) {
	v := &model.SysNoticeList{}
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(spcSvc.SysNotice(c, v))
}

func updateSysNotice(c *bm.Context) {
	v := &model.SysNoticeUp{}
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, spcSvc.SysNoticeUp(c, v))
}

func optSysNotice(c *bm.Context) {
	v := &model.SysNoticeOpt{}
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, spcSvc.SysNoticeOpt(c, v))
}

func addSysNoticeUid(c *bm.Context) {
	res := map[string]interface{}{}
	v := &model.SysNotUidAddDel{}
	if err := c.Bind(v); err != nil {
		return
	}
	if err := spcSvc.SysNoticeUidAdd(c, v); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func SysNoticeUid(c *bm.Context) {
	var (
		err   error
		value *model.SysNoticePager
	)
	res := map[string]interface{}{}
	v := &model.SysNoticeUidParam{}
	if err = c.Bind(v); err != nil {
		return
	}
	if value, err = spcSvc.SysNoticeUid(c, v); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(value, nil)
}

func delSysNoticeUid(c *bm.Context) {
	res := map[string]interface{}{}
	v := &model.SysNotUidAddDel{}
	if err := c.Bind(v); err != nil {
		return
	}
	if err := spcSvc.SysNoticeUidDel(c, v); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}
