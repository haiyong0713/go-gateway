package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model/page"
)

func pageList(c *bm.Context) {
	req := new(page.ReqPageList)
	if err := c.Bind(req); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(pageSvc.PageList(c, req))
}
