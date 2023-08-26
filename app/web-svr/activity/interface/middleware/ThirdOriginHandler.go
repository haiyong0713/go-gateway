package middleware

import (
	bm "go-common/library/net/http/blademaster"
)

type thirdPartOriginHandler struct{}

func (handler *thirdPartOriginHandler) ServeHTTP(ctx *bm.Context) {
	path := ctx.Request.URL.Path
	switch path {
	case "/x/activity/invite/bind", "/x/activity/invite/inviter", "/x/activity/spring2021/get_card", "/x/activity/spring2021/bind":
		ctx.Request.Header.Set("Origin", "")
	}
}

func ThirdPartOriginHandler() bm.Handler {
	return new(thirdPartOriginHandler)
}
