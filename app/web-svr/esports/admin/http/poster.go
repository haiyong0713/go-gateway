package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	pb "go-gateway/app/web-svr/esports/admin/api"
)

func createPoster(c *bm.Context) {
	p := new(pb.CreatePosterReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	p.CreatedBy = userInfo(c).Name

	if p.Order == 0 || p.BgImage == "" || p.ContestID == 0 {
		c.JSON(nil, ecode.Error(-400, "必填项未填写，请检查"))
		c.Abort()
		return
	}

	if contest, err := esSvc.ContestInfo(c, p.ContestID); err != nil {
		log.Error("create poster error 0: %s", err.Error())
		c.JSON(nil, err)
		c.Abort()
		return
	} else if contest == nil {
		c.JSON(nil, ecode.Error(-400, "赛事ID不合法，请检查"))
		c.Abort()
		return
	}

	err := esSvc.CreatePoster(c, p)
	c.JSON(nil, err)
}

func editPoster(c *bm.Context) {
	p := new(pb.EditPosterReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	p.CreatedBy = userInfo(c).Name

	if p.Id == 0 || p.Order == 0 || p.BgImage == "" || p.ContestID == 0 {
		c.JSON(nil, ecode.Error(-400, "必填项未填写，请检查"))
		c.Abort()
		return
	}

	if contest, err := esSvc.ContestInfo(c, p.ContestID); err != nil {
		log.Error("create poster error 0: %s", err.Error())
		c.JSON(nil, err)
		c.Abort()
		return
	} else if contest == nil {
		c.JSON(nil, ecode.Error(-400, "赛事ID不合法，请检查"))
		c.Abort()
		return
	}

	err := esSvc.EditPoster(c, p)
	c.JSON(nil, err)
}

func togglePoster(c *bm.Context) {
	p := new(pb.TogglePosterReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.Id == 0 || (p.OnlineStatus != 0 && p.OnlineStatus != 1) {
		c.JSON(nil, ecode.Error(-400, "必填项未填写，请检查"))
		c.Abort()
		return
	}
	err := esSvc.TogglePoster(c, p)
	c.JSON(nil, err)
}

func centerPoster(c *bm.Context) {
	p := new(pb.CenterPosterReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.Id == 0 || (p.IsCenteral != 0 && p.IsCenteral != 1) {
		c.JSON(nil, ecode.Error(-400, "必填项未填写，请检查"))
		c.Abort()
		return
	}
	err := esSvc.CenterPoster(c, p)
	c.JSON(nil, err)
}

func deletePoster(c *bm.Context) {
	p := new(pb.DeletePosterReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.Id == 0 {
		c.JSON(nil, ecode.Error(-400, "有必填项未填写，请检查"))
		c.Abort()
		return
	}

	err := esSvc.DeletePoster(c, p)
	c.JSON(nil, err)
}

func getPosterList(c *bm.Context) {
	p := new(pb.GetPosterListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	if p.PageSize < 0 || p.PageNum < 0 {
		c.JSON(nil, ecode.Error(-400, "分页配置有问题，请检查"))
		c.Abort()
		return
	}
	if p.PageSize == 0 {
		p.PageSize = 20
	}
	if p.PageNum == 0 {
		p.PageNum = 1
	}
	resp, err := esSvc.GetPosterList(c, p)
	c.JSON(resp, err)
}

func getEffectivePosterList(c *bm.Context) {
	resp, err := esSvc.GetEffectivePosterList(c)
	c.JSON(resp, err)
}
