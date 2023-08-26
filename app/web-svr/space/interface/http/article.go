package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/interface/conf"
	"go-gateway/app/web-svr/space/interface/model"
)

func article(c *bm.Context) {
	var (
		mid int64
		ok  bool
		err error
	)
	params := c.Request.Form
	midStr := params.Get("mid")
	pnStr := params.Get("pn")
	psStr := params.Get("ps")
	sortStr := params.Get("sort")
	if mid, err = strconv.ParseInt(midStr, 10, 64); err != nil || mid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	pn, err := strconv.ParseInt(pnStr, 10, 32)
	if err != nil || pn < 1 {
		pn = 1
	}
	ps, err := strconv.ParseInt(psStr, 10, 32)
	if err != nil || ps < 1 || ps > int64(conf.Conf.Rule.MaxArticlePs) {
		ps = int64(conf.Conf.Rule.MaxArticlePs)
	}
	var sort int
	if sortStr != "" {
		if sort, ok = model.ArticleSortType[sortStr]; !ok {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	} else {
		sort = 0
	}
	c.JSON(spcSvc.Article(c, mid, int32(pn), int32(ps), int32(sort)))
}
