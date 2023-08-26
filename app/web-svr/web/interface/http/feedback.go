package http

import (
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/model"
)

func feedback(c *bm.Context) {
	var (
		mid, aid int64
		buvid    string
		buvidCk  *http.Cookie
		midStr   interface{}
		ok       bool
		err      error
	)
	v := new(struct {
		Aid     int64  `form:"aid"`
		Bvid    string `form:"bvid"`
		TagID   int64  `form:"tag_id" validate:"min=1"`
		Content string `form:"content"`
		Browser string `form:"browser"`
		Version string `form:"version"`
		Email   string `form:"email"`
		QQ      string `form:"qq"`
		Other   string `form:"other"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Aid != 0 || v.Bvid != "" {
		if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
			c.JSON(nil, err)
			return
		}
	}
	if !model.CheckFeedTag(v.TagID) {
		log.Warn("tag_id(%d) check fail", v.TagID)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if buvidCk, err = c.Request.Cookie("buvid3"); err != nil {
		log.Warn("buvid3 is nil")
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if buvid = buvidCk.Value; buvid == "" {
		log.Warn("buvid == nil")
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midStr, ok = c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	feedParams := &model.Feedback{
		Aid:     aid,
		Mid:     mid,
		TagID:   v.TagID,
		Buvid:   buvid,
		Browser: v.Browser,
		Version: v.Version,
		Content: &model.Content{Reason: v.Content},
		Email:   v.Email,
		QQ:      v.QQ,
		Other:   v.Other,
	}
	c.JSON(nil, webSvc.Feedback(c, feedParams))
}
