package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func ugcURL(c *bm.Context) {
	arg := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(service.PageSvc.UgcURL(c, arg.ID))
}
