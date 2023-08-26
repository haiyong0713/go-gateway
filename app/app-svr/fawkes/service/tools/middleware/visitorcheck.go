package middleware

import (
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_referer   = "Referer"
	_userAgent = "User-Agent"
	_caller    = "x1-bilispy-user"
)

func VisitorCheck() bm.HandlerFunc {
	return func(c *bm.Context) {
		caller := c.Request.Header.Get(_caller)
		referer := c.Request.Header.Get(_referer)
		userAgent := c.Request.Header.Get(_userAgent)

		requestURI := c.Request.RequestURI
		if !strings.HasPrefix(caller, "fawkes_") {
			log.Infoc(c, "VisitorCheck fawkes_user failed! caller=%s uri=%s", caller, requestURI)
			c.JSON(nil, ecode.Error(ecode.Unauthorized, "please make sure you have the permission."))
			c.Abort()
		}
		if !strings.Contains(referer, "fawkes.bilibili.co") {
			log.Infoc(c, "VisitorCheck fawkes_user failed! referer=%s uri=%s", referer, requestURI)
			c.JSON(nil, ecode.Error(ecode.Unauthorized, "please make sure you have the permission."))
			c.Abort()
		}
		if userAgent == "" {
			log.Infoc(c, "VisitorCheck fawkes_user failed! userAgent=%s uri=%s", userAgent, requestURI)
			c.JSON(nil, ecode.Error(ecode.Unauthorized, "please make sure you have the permission."))
			c.Abort()
		}
	}
}
