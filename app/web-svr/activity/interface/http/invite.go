package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	mdl "go-gateway/app/web-svr/activity/interface/model/invite"
	"go-gateway/app/web-svr/activity/interface/service"
)

const (
	activityID = "2020college"
	// tp
	tp = 2
)

// inviteToken 生成token
func inviteToken(c *bm.Context) {
	v := new(struct {
		Source int64 `form:"source" validate:"min=1"`
	})
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	params := headers(c)
	c.JSON(service.InviteSvc.Token(c, mid, activityID, tp, v.Source, params))
}

func headers(ctx *bm.Context) *mdl.BaseInfo {
	rs := new(mdl.BaseInfo)
	request := ctx.Request
	ctx.Bind(rs)
	rs.UA = request.UserAgent()
	rs.Referer = request.Referer()
	rs.IP = metadata.String(ctx, metadata.RemoteIP)
	if rs.Buvid == "" {
		if res, err := request.Cookie("buvid3"); err == nil {
			rs.Buvid = res.Value
		}
	}
	return rs
}

// inviteBind 预保定
func inviteBind(c *bm.Context) {
	v := new(struct {
		Token string `form:"token" validate:"required"`
		Tel   string `form:"tel" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	params := &mdl.BindReq{
		ActivityID: activityID,
		Token:      v.Token,
		Tel:        v.Tel,
	}
	params.BaseInfo = headers(c)
	c.JSON(service.InviteSvc.InviteBind(c, params))
}

// inviteGetInviter 获取邀请人
func inviteGetInviter(c *bm.Context) {
	v := new(struct {
		Token string `form:"token" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.InviteSvc.Inviter(c, v.Token))
}
