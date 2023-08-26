package anticrawler

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler/model"
)

// Report is used to report access log of anti-crawler.
func Report() bm.HandlerFunc {
	return report(_antiCrawler.send, httpFilter())
}

// report is used to report access log of anti-crawler.
func report(send Send, filter Filter) bm.HandlerFunc {
	return func(ctx *bm.Context) {
		if send == nil {
			ctx.Next()
			return
		}
		writer := &antiCrawlerWriter{}
		writer.response = ctx.Writer
		ctx.Writer = writer
		ctx.Next()
		req := ctx.Request
		bodyFunc := func(req *http.Request) (bodyBytes []byte) {
			if len(req.PostForm) != 0 {
				bodyBytes = []byte(req.PostForm.Encode())
				return
			}
			if req.Body != nil {
				bodyBytes, _ = ioutil.ReadAll(req.Body)
				req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			}
			return
		}
		var mid int64
		if v, ok := ctx.Get("mid"); ok {
			mid, _ = v.(int64)
		}
		var buvid string
		if dev, ok := ctx.Get("device"); ok {
			device, ok := dev.(*device.Device)
			if ok {
				buvid = device.Buvid
			}
		}
		host := func() string {
			host := req.URL.Host
			if host != "" {
				return host
			}
			return os.Getenv("APP_ID")
		}()
		reqHeader, _ := json.Marshal(req.Header)
		respHeader, _ := json.Marshal(writer.Header())
		sample := random()
		if filter != nil && filter(ctx) {
			sample = -1
		}
		data := &model.InfocMsg{
			Mid:            mid,
			Buvid:          buvid,
			Host:           host,
			Path:           req.URL.Path,
			Body:           string(bodyFunc(req)),
			Method:         req.Method,
			Header:         string(reqHeader),
			Query:          req.Form.Encode(),
			Referer:        req.Referer(),
			IP:             metadata.String(ctx, metadata.RemoteIP),
			Ctime:          time.Now().Unix(),
			ResponseHeader: string(respHeader),
			ResponseBody:   writer.body.String(),
			Sample:         sample,
		}
		if err := send(ctx, data); err != nil {
			log.Error("failed to send data error: %+v", err)
		}
	}
}

type antiCrawlerWriter struct {
	response http.ResponseWriter

	status int
	body   bytes.Buffer
}

func (w *antiCrawlerWriter) Header() http.Header { return w.response.Header() }

func (w *antiCrawlerWriter) WriteHeader(code int) {
	w.status = code
	w.response.WriteHeader(code)
}

func (w *antiCrawlerWriter) Write(data []byte) (size int, err error) {
	// write origin response
	w.body.Write(data)
	return w.response.Write(data)
}

// httpFilter return mid filter.
func httpFilter() Filter {
	return func(c context.Context) bool {
		ctx, ok := c.(*bm.Context)
		if !ok {
			return false
		}
		var mid int64
		if v, ok := ctx.Get("mid"); ok {
			mid, _ = v.(int64)
		}
		var buvid string
		if d, ok := device.FromContext(ctx); ok {
			buvid = d.Buvid
		}
		return wList(mid, buvid)
	}
}
