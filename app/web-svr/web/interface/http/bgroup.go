package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/model"

	bgroupgrpc "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
)

// 用mid/buvid查询是否某个人在
func memberIn(c *bm.Context) {
	var (
		err error
		rs  *bgroupgrpc.MemberInReply_MemberInReplySingle
	)
	v := &model.MemberInReq{}
	if err = c.Bind(v); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		v.Mid = midStr.(int64)
	}
	if ck, err := c.Request.Cookie("buvid3"); err == nil {
		v.Buvid = ck.Value
	}
	if rs, err = webSvc.MemberIn(c, v); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}
