package http

import (
	"net/http"
	"strings"
	"time"

	bm "go-common/library/net/http/blademaster"
)

const _captchaCkKey = "_dfcaptcha"

func captchaKey(ctx *bm.Context) {
	_, needJsRes := ctx.Request.Form["js"]
	var key string
	if keyCk, err := ctx.Request.Cookie(_captchaCkKey); err == nil {
		key = keyCk.Value
	}
	saveKey, res := webSvc.CaptchaKey(ctx, key, needJsRes)
	// set cookie
	http.SetCookie(ctx.Writer, &http.Cookie{Name: _captchaCkKey, Value: saveKey, Expires: time.Now().Add(time.Hour), Path: "/", Domain: ".bilibili.com", HttpOnly: true})
	ctx.Bytes(http.StatusOK, "text/html; charset=UTF-8", []byte(res))
}

func serverDate(ctx *bm.Context) {
	v := new(struct {
		Ts string `form:"ts"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var needTsRes bool
	if strings.TrimSpace(v.Ts) != "" {
		needTsRes = true
	}
	ctx.Bytes(http.StatusOK, "text/html; charset=UTF-8", []byte(webSvc.ServerDate(ctx, needTsRes)))
}
