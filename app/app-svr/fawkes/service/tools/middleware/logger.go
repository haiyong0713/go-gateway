package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
)

const (
	MaxPrintBodyLen = 512
	_500ms          = time.Millisecond * 500
)

type bodyLogWriter struct {
	http.ResponseWriter
	bodyBuf *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.bodyBuf.Write(b)
	return w.ResponseWriter.Write(b)
}

func Logger() bm.HandlerFunc {
	const noUser = "no_user"
	return func(c *bm.Context) {
		now := time.Now()
		ip := metadata.String(c, metadata.RemoteIP)
		req := c.Request
		path := req.URL.Path
		header := req.Header
		reqVal := url.Values{}
		if len(c.Request.Form) != 0 {
			reqVal = c.Request.Form
		} else if len(c.Request.PostForm) != 0 {
			reqVal = c.Request.PostForm
		}
		var quota float64
		if deadline, ok := c.Context.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}
		strBody := ""
		blw := bodyLogWriter{bodyBuf: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		var mid int64
		if v, ok := c.Get("mid"); ok {
			mid, _ = v.(int64)
		}
		err := c.Error
		cerr := ecode.Cause(err)
		dt := time.Since(now)
		caller := metadata.String(c, metadata.Caller)
		if caller == "" {
			caller = noUser
		}

		buvid := ""
		if dev, ok := c.Get("device"); ok {
			device, ok := dev.(*bm.Device)
			if ok {
				buvid = device.Buvid
			}
		}

		strBody = strings.Trim(blw.bodyBuf.String(), "\n")
		if len(strBody) > MaxPrintBodyLen {
			strBody = strBody[:(MaxPrintBodyLen - 1)]
		}

		lf := log.Infov
		errmsg := ""
		isSlow := dt >= _500ms
		if err != nil {
			errmsg = err.Error()
			lf = log.Errorv
			if cerr.Code() > 0 {
				lf = log.Warnv
			}
		} else {
			if isSlow {
				lf = log.Warnv
			}
		}
		lf(c,
			log.KVString("method", req.Method),
			log.KVInt64("mid", mid),
			log.KVString("ip", ip),
			log.KVString("user", caller),
			log.KVString("router", path),
			log.KVString("params", reqVal.Encode()),
			log.KVString("params_raw", fmt.Sprintf("%+v", reqVal)),
			log.KVString("header", fmt.Sprintf("%+v", header)),
			log.KVString("response", strBody),
			log.KVInt("ret", cerr.Code()),
			log.KVString("msg", cerr.Message()),
			log.KVString("stack", fmt.Sprintf("%+v", err)),
			log.KVString("err", errmsg),
			log.KVFloat64("timeout_quota", quota),
			log.KVFloat64("ts", dt.Seconds()),
			log.KVString("buvid", buvid),
			log.KVString("source", "http-fawkes-log"),
		)
	}
}
