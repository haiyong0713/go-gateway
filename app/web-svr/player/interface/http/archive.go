package http

import (
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/archive/service/api"
)

func pageList(c *bm.Context) {
	var (
		aid   int64
		err   error
		pages []*api.Page
	)
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	if pages, err = playSvr.PageList(c, aid); err != nil {
		if playSvr.SLBRetry(err) {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Error("%+v", err)
			return
		}
		c.JSON(nil, err)
		return
	}
	if len(pages) == 0 {
		c.JSON(nil, ecode.NothingFound)
		return
	}
	c.JSON(pages, nil)
}

func videoShot(c *bm.Context) {
	var (
		aid, mid int64
		err      error
	)
	v := new(struct {
		Aid   int64  `form:"aid"`
		Bvid  string `form:"bvid"`
		Cid   int64  `form:"cid"`
		Index bool   `form:"index"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	var buvid string
	if ck, err := c.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	c.JSON(playSvr.VideoShot(c, mid, aid, v.Cid, v.Index, buvid))
}

func playURLToken(c *bm.Context) {
	var (
		aid int64
		err error
	)
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
		Cid  int64  `form:"cid"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(playSvr.PlayURLToken(c, mid, aid, v.Cid))
}
