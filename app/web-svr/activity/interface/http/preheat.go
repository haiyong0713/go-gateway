package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func downloadInfo(c *bm.Context) {
	arg := new(struct {
		ID int64 `form:"id" json:"id" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(service.PreheatSvc.DownloadInfo(c, arg.ID))
}
