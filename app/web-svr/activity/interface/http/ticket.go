package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func ticketSign(c *bm.Context) {
	v := new(struct {
		Ticket string `form:"ticket" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midI, _ := c.Get("mid")
	mid := midI.(int64)
	c.JSON(service.LikeSvc.TicketSign(c, mid, v.Ticket))
}
