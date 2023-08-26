package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/web/interface/conf"
)

func articleList(c *bm.Context) {
	var (
		rid, mid int64
		aids     []int64
		sort     int64
		err      error
	)
	param := c.Request.Form
	pnStr := param.Get("pn")
	psStr := param.Get("ps")
	ridStr := param.Get("rid")
	aidsStr := param.Get("aids")
	sortStr := param.Get("sort")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	pn, err := strconv.ParseInt(pnStr, 10, 32)
	if err != nil || pn < 1 {
		pn = 1
	}
	ps, err := strconv.ParseInt(psStr, 10, 32)
	if err != nil || ps < 1 || ps > int64(conf.Conf.Rule.MaxArtPageSize) {
		ps = int64(conf.Conf.Rule.MaxArtPageSize)
	}
	if ridStr != "" {
		if rid, err = strconv.ParseInt(ridStr, 10, 64); err != nil || rid < 0 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if aidsStr != "" {
		if aids, err = xstr.SplitInts(aidsStr); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if sortStr != "" {
		if sort, err = strconv.ParseInt(sortStr, 10, 32); err != nil || sort < 0 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	c.JSON(webSvc.ArticleList(c, rid, mid, int32(sort), int32(pn), int32(ps), aids))
}

func articleUpList(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.ArticleUpList(c, mid))
}

func categories(c *bm.Context) {
	c.JSON(webSvc.Categories(c))
}

func newCount(c *bm.Context) {
	var (
		count, pubTime int64
		err            error
	)
	pubTimeStr := c.Request.Form.Get("pubtime")
	if pubTime, err = strconv.ParseInt(pubTimeStr, 10, 64); err != nil || pubTime < 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if count, err = webSvc.NewCount(c, pubTime); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(struct {
		NewCount int64 `json:"new_count"`
	}{NewCount: count}, nil)
}

func upMoreArts(c *bm.Context) {
	var (
		aid int64
		err error
	)
	aidStr := c.Request.Form.Get("aid")
	if aid, err = strconv.ParseInt(aidStr, 10, 64); err != nil || aid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(webSvc.UpMoreArts(c, aid))
}
