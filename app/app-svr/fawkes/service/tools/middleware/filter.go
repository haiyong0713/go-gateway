package middleware

import (
	"net/http"
	"net/url"

	bm "go-common/library/net/http/blademaster"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/middleware/fieldfilter"
)

type filterWriter struct {
	sl fieldfilter.Selection
	http.ResponseWriter
}

func (w filterWriter) Write(b []byte) (int, error) {
	// 拦截下写入的byte 修改后再写入response中
	mask, err := w.sl.Mask(b)
	if err != nil {
		return 0, err
	}
	return w.ResponseWriter.Write(mask)
}

// ReqFilter 只会使用给定的req字段，没有给的会被忽略
// @see fieldfilter.Compile
// example middleware.ReqFilter("app_key,build_id,gl_job_id")
func ReqFilter(req string) bm.HandlerFunc {
	return func(c *bm.Context) {
		reqc, err := fieldfilter.Compile(req)
		if err != nil {
			log.Errorc(c, "FieldFilter error %v", err)
			c.Abort()
			return
		}
		log.Infoc(c, "origin form: %v", c.Request.Form)
		maskForm := make(url.Values)
		for k := range reqc {
			maskForm.Set(k, c.Request.Form.Get(k))
		}
		c.Request.Form = maskForm
		log.Infoc(c, "current form: %v", c.Request.Form)
	}
}

// RespFilter 只会返回指定的字段 没有给的不会出现在返回值中
// 字段过滤器 @see fieldfilter.Compile
// example middleware.RespFilter("code,message,data(app_id,app_key)")
func RespFilter(resp string) bm.HandlerFunc {
	return func(c *bm.Context) {
		respc, err := fieldfilter.Compile(resp)
		if err != nil {
			log.Errorc(c, "FieldFilter error %v", err)
			c.Abort()
			return
		}
		blw := filterWriter{
			sl:             respc,
			ResponseWriter: c.Writer,
		}
		c.Writer = blw
	}
}
