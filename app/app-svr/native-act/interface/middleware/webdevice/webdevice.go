package webdevice

import (
	"context"
	"regexp"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
)

type Webdevice struct {
	Platform  string `json:"platform" form:"platform" validate:"required"`
	MobiApp   string
	Buvid     string
	UserAgent string
}

type webdeviceKey struct{}

func NewContext(c context.Context, dev Webdevice) context.Context {
	return context.WithValue(c, webdeviceKey{}, dev)
}

func FromContext(c context.Context) (Webdevice, bool) {
	dev, ok := c.Value(webdeviceKey{}).(Webdevice)
	return dev, ok
}

func BindWebdevice() bm.HandlerFunc {
	return func(ctx *bm.Context) {
		dev := &Webdevice{}
		if err := ctx.BindWith(dev, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
			log.Error("Fail to bind Webdevice, form=%+v error=%+v", ctx.Request.Form, err)
			return
		}
		dev.MobiApp = mobiApp(ctx)
		dev.Buvid = buvid(ctx)
		dev.UserAgent = ctx.Request.UserAgent()
		ctx.Context = NewContext(ctx.Context, *dev)
	}
}

func mobiApp(c *bm.Context) string {
	ua := c.Request.Header.Get("user-agent")
	if ua == "" {
		return ""
	}
	if matched, err := regexp.Match("iPhone", []byte(ua)); err == nil && matched {
		return "iphone"
	}
	if matched, err := regexp.Match("android", []byte(ua)); err == nil && matched {
		return "android"
	}
	return ""
}

func buvid(c *bm.Context) string {
	if cookie, err := c.Request.Cookie("Buvid"); err == nil {
		return cookie.Value
	}
	if cookie, err := c.Request.Cookie("buvid3"); err == nil {
		return cookie.Value
	}
	return ""
}
