package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-gateway/app/web-svr/space/admin/model"
)

func bannedRcmdList(c *bm.Context) {
	p := new(model.UpRcmdBlackListSearchReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.Pn <= 0 || p.Ps < 0 || p.Ps > 100 {
		c.JSON(nil, ecode.Error(-400, "传入参数有误，请检查"))
		c.Abort()
		return
	}
	c.JSON(spcSvc.GetBannedRcmdList(c, p))
}

func bannedRcmdAdd(c *bm.Context) {
	p := new(model.UpRcmdBlackListCreateReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if len(p.Mids) == 0 {
		c.JSON(nil, ecode.Error(-400, "传入参数有误，请检查"))
		c.Abort()
		return
	}
	c.JSON(spcSvc.AddBannedRcmd(c, p))
}

func bannedRcmdDelete(c *bm.Context) {
	p := new(model.UpRcmdBlackListDeleteReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.Mid == 0 {
		c.JSON(nil, ecode.Error(-400, "传入参数有误，请检查"))
		c.Abort()
		return
	}
	c.JSON(nil, spcSvc.DeleteBannedRcmd(c, p))
}

//func bannedRcmdSearchMids(c *bm.Context) {
//	p := new(model.UserInfoSearchReq)
//	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
//		return
//	}
//	resp, err := spcSvc.SearchMidInfo(c, p)
//	c.JSON(resp, err)
//}
