package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/admin/model"
)

func notice(c *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(spcSvc.Notice(c, v.Mid))
}

func noticeUp(c *bm.Context) {
	v := new(model.NoticeUpArg)
	if err := c.Bind(v); err != nil {
		return
	}
	if uidInter, ok := c.Get("uid"); ok {
		v.UID = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		v.Uname = usernameCtx.(string)
	}
	c.JSON(nil, spcSvc.NoticeUp(c, v))
}
