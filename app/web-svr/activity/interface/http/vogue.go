package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/activity/interface/service"
)

func vogueState(c *bm.Context) {
	mid, _ := c.Get("mid")
	c.JSON(service.VogueSvc.State(c, mid.(int64)))
}

func voguePrizes(c *bm.Context) {
	c.JSON(service.VogueSvc.Prizes(c))
}

func vogueShare(c *bm.Context) {
	var (
		err error
		v   = new(struct {
			Token string `form:"token" validate:"required"`
		})
	)
	if err = c.Bind(v); err != nil {
		return
	}
	c.JSON(service.VogueSvc.ShareInfo(c, v.Token))
}

func vogueSelectPrizes(c *bm.Context) {
	var (
		err error
		v   = new(struct {
			PrizeId int64  `form:"prize_id" validate:"required"`
			Token   string `form:"token"`
		})
	)
	if err = c.Bind(v); err != nil {
		return
	}
	mid, _ := c.Get("mid")
	c.JSON(service.VogueSvc.SelectPrizes(c, mid.(int64), v.PrizeId, v.Token))
}

func vogueAddtimes(c *bm.Context) {
	mid, _ := c.Get("mid")
	ip := metadata.String(c, metadata.RemoteIP)
	c.JSON(service.VogueSvc.Addtimes(c, mid.(int64), ip))
}

func vogueExchange(c *bm.Context) {
	mid, _ := c.Get("mid")
	ip := metadata.String(c, metadata.RemoteIP)
	c.JSON(nil, service.VogueSvc.Exchange(c, mid.(int64), ip))
}

func vogueAddress(c *bm.Context) {
	var (
		err error
		v   = new(struct {
			AddressId int64 `form:"address_id" validate:"required"`
		})
	)
	if err = c.Bind(v); err != nil {
		return
	}
	mid, _ := c.Get("mid")
	ip := metadata.String(c, metadata.RemoteIP)
	c.JSON(nil, service.VogueSvc.Address(c, mid.(int64), v.AddressId, ip))
}

func vogueInviteList(c *bm.Context) {
	var (
		err error
		v   = new(struct {
			Id int64 `form:"id"`
		})
	)
	if err = c.Bind(v); err != nil {
		return
	}
	mid, _ := c.Get("mid")
	c.JSON(service.VogueSvc.InviteList(c, mid.(int64), v.Id))
}

func voguePrizeList(c *bm.Context) {
	c.JSON(service.VogueSvc.PrizeList(c))
}
