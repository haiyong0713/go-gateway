package middleware

import (
	"strconv"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/stat/metric"
)

const (
	serverNamespace = "fawkes"
)

var (
	MetricServerStatusCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "http",
		Name:      "status_code_total",
		Help:      "http server requests code count.",
		Labels:    []string{"path", "code"},
	})

	MetricServerDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "http",
		Name:      "duration_ms",
		Help:      "client requests duration(ms).",
		Labels:    []string{"path", "code"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500},
	})
)

func Moni() bm.HandlerFunc {
	return func(c *bm.Context) {
		now := time.Now()

		c.Next()

		path := c.Request.URL.Path
		err := c.Error
		cerr := ecode.Cause(err)
		MetricServerStatusCodeTotal.Inc(path, strconv.Itoa(cerr.Code()))
		MetricServerDur.Observe(int64(time.Since(now)/time.Millisecond), path, strconv.Itoa(cerr.Code()))
	}
}
