package http

import (
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/web/interface/model"
)

func addPwdAppeal(c *bm.Context) {
	req := &model.AddPwdAppealReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(nil, webSvc.AddPwdAppeal(c, req, mid))
}

func uploadPwdAppeal(c *bm.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("Fail to FormFile of PwdAppeal photo, error=%+v", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	key := func() string {
		if token := c.Request.Form.Get("device_token"); token != "" {
			return token
		}
		var mid int64
		if midInter, ok := c.Get("mid"); ok {
			mid = midInter.(int64)
		}
		if mid == 0 {
			return ""
		}
		return strconv.FormatInt(mid, 10)
	}()
	c.JSON(webSvc.UploadPwdAppeal(c, key, file))
}

func pwdAppealSendCaptcha(c *bm.Context) {
	req := &model.SendCaptchaReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(nil, webSvc.SendCaptcha(c, req, webSvc.PwdAppealCaptchaDao))
}
