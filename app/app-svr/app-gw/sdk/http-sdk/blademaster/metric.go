package blademaster

import (
	"bytes"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/stat/metric"
	"go-gateway/app/app-svr/app-gw/gateway/api"
	"go-gateway/app/app-svr/app-gw/sdk"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster/asset"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/prom"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"

	"github.com/BurntSushi/toml"
	"github.com/shirou/gopsutil/process"
)

var (
	ProxyStatusCode = prom.New().
			WithCounter(ProxyStatusCounter, []string{"location", "code"}).
			WithState("proxy_status_code_state", []string{"location", "code"})
	UpstreamStatusCode = prom.New().
				WithCounter(UpstreamStatusCounter, []string{"identifier", "code"}).
				WithState("upstream_status_code_state", []string{"identifier", "code"})
	ClientReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: "proxy_http_client",
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "http client requests duration(ms).",
		Labels:    []string{"path"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})
	ClientReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "proxy_http_client",
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "http client requests code count.",
		Labels:    []string{"path", "code"},
	})
)

const (
	UnexpectedPath        = "[unexpected-path]"
	ProxyStatusCounter    = "proxy_status_code_count"
	UpstreamStatusCounter = "upstream_status_code_count"
)

func exportMetrics(metricFamily, identifierName string) map[string]*Metrics {
	mf, ok := prom.FindMetricFamily(metricFamily)
	if !ok {
		return nil
	}
	metrics := map[string]*Metrics{}
	getTarget := func(dst string) (*Metrics, map[string]*float64) {
		target, ok := metrics[dst]
		if !ok {
			target = &Metrics{
				Name: dst,
			}
			metrics[dst] = target
		}
		codeSolt := map[string]*float64{
			"request": &target.Requests,
			"1xx":     &target.Response.Code1xx,
			"2xx":     &target.Response.Code2xx,
			"3xx":     &target.Response.Code3xx,
			"4xx":     &target.Response.Code4xx,
			"5xx":     &target.Response.Code5xx,
			"unknown": &target.Response.Unknown,
			"total":   &target.Response.Total,
		}
		return target, codeSolt
	}
	for _, m := range mf.GetMetric() {
		idLV, ok := prom.GetLabel(m, identifierName)
		if !ok {
			continue
		}
		codeLV, ok := prom.GetLabel(m, "code")
		if !ok {
			continue
		}
		_, codeSolt := getTarget(idLV.GetValue())
		solt, ok := codeSolt[codeLV.GetValue()]
		if !ok {
			continue
		}
		*solt = prom.ToFloat64(m)
	}
	return metrics
}

func asMetricSlice(in map[string]*Metrics) []*Metrics {
	out := make([]*Metrics, 0, len(in))
	for _, m := range in {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name > out[j].Name
	})
	return out
}

func (p *ProxyPass) upstreamMetrics() []*Metrics {
	metrics := exportMetrics(UpstreamStatusCounter, "identifier")
	return asMetricSlice(metrics)
}

func (p *ProxyPass) locationMetrics() []*Metrics {
	metrics := exportMetrics(ProxyStatusCounter, "location")
	return asMetricSlice(metrics)
}

func (p *ProxyPass) Metrics() *MetricsReply {
	return &MetricsReply{
		Instance: hostname(),
		Uptime:   time.Now().Unix() - uptime(),
		Pid:      os.Getpid(),
		Upstream: p.upstreamMetrics(),
		Location: p.locationMetrics(),
	}
}

func hostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func uptime() int64 {
	p, err := process.NewProcess(1)
	if err != nil {
		return 0
	}
	ctime, err := p.CreateTime()
	if err != nil {
		return 0
	}
	return ctime / 1000
}

func (p *ProxyPass) metrics(ctx *bm.Context) {
	ctx.JSON(p.Metrics(), nil)
}

func metricsPage(ctx *bm.Context) {
	page, err := asset.Asset("metric.html")
	if err != nil {
		ctx.String(500, "%v", err)
		return
	}
	ctx.Bytes(200, "text/html", page)
}

type Metrics struct {
	Name       string          `json:"name"`
	Processing float64         `json:"processing"`
	Requests   float64         `json:"requests"`
	QPS        float64         `json:"qps"`
	Response   ResponseMetrics `json:"response"`
}

// ResponseMetrics is
type ResponseMetrics struct {
	Code1xx float64 `json:"1xx"`
	Code2xx float64 `json:"2xx"`
	Code3xx float64 `json:"3xx"`
	Code4xx float64 `json:"4xx"`
	Code5xx float64 `json:"5xx"`
	Unknown float64 `json:"unknown"`
	Total   float64 `json:"total"`
}

// MetricsReply is
type MetricsReply struct {
	Instance string     `json:"instance"`
	Uptime   int64      `json:"uptime"`
	Pid      int        `json:"pid"`
	Location []*Metrics `json:"location"`
	Upstream []*Metrics `json:"upstream"`
}

func reportProxyStatusCode(pattern string, writer http.ResponseWriter) {
	code := 0
	ww, ok := writer.(*wrappedWriter)
	if ok {
		code = ww.statusCode
	}
	ProxyStatusCode.Incr(pattern, "request")
	ProxyStatusCode.Incr(pattern, "total")
	switch {
	case code >= 100 && code < 200:
		ProxyStatusCode.Incr(pattern, "1xx")
	case code >= 200 && code < 300:
		ProxyStatusCode.Incr(pattern, "2xx")
	case code >= 300 && code < 400:
		ProxyStatusCode.Incr(pattern, "3xx")
	case code >= 400 && code < 500:
		ProxyStatusCode.Incr(pattern, "4xx")
	case code >= 500 && code < 600:
		ProxyStatusCode.Incr(pattern, "5xx")
	default:
		ProxyStatusCode.Incr(pattern, "unknown")
	}
}

// ReportUpstreamAttemptHandler is
var ReportUpstreamAttemptHandler = request.NamedHandler{
	Name: "appgwsdk.blademaster.ReportUpstreamAttemptHandler",
	Fn: func(r *request.Request) {
		UpstreamStatusCode.Incr(r.SafeMetricURI(), "request")
	},
}

var ReportClientMetricsCompleteHandler = request.NamedHandler{
	Name: "appgwsdk.blademaster.ReportClientMetricsCompleteHandler",
	Fn: func(r *request.Request) {
		reportFunc := func(uri string, code string) {
			ClientReqDur.Observe(int64(r.CompleteTime.Sub(r.Time)/time.Millisecond), uri)
			ClientReqCodeTotal.Inc(uri, code)
		}
		if r.Error != nil {
			codeErr := ecode.Cause(r.Error)
			reportFunc(r.SafeMetricURI(), strconv.FormatInt(int64(codeErr.Code()), 10))
			return
		}
		reportFunc(r.SafeMetricURI(), "0")
	},
}

// ReportUpstreamCompleteHandler is
var ReportUpstreamCompleteHandler = request.NamedHandler{
	Name: "appgwsdk.blademaster.ReportUpstreamCompleteHandler",
	Fn: func(r *request.Request) {
		reportFunc := func(identifier string, code int) {
			UpstreamStatusCode.Incr(identifier, "total")
			switch {
			case code >= 100 && code < 200:
				UpstreamStatusCode.Incr(identifier, "1xx")
			case code >= 200 && code < 300:
				UpstreamStatusCode.Incr(identifier, "2xx")
			case code >= 300 && code < 400:
				UpstreamStatusCode.Incr(identifier, "3xx")
			case code >= 400 && code < 500:
				UpstreamStatusCode.Incr(identifier, "4xx")
			case code >= 500 && code < 600:
				UpstreamStatusCode.Incr(identifier, "5xx")
			default:
				UpstreamStatusCode.Incr(identifier, "unknown")
			}
		}
		if r.HTTPResponse == nil {
			reportFunc(r.SafeMetricURI(), 0)
			return
		}
		reportFunc(r.SafeMetricURI(), r.HTTPResponse.StatusCode)
	},
}

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (ww *wrappedWriter) WriteHeader(statusCode int) {
	ww.statusCode = statusCode
	ww.ResponseWriter.WriteHeader(statusCode)
}

type metricReporter struct{}

func (metricReporter) ServeHTTP(ctx *bm.Context) {
	ctx.Writer = &wrappedWriter{ResponseWriter: ctx.Writer}
}

func (p *ProxyPass) configs(ctx *bm.Context) {
	res, err := p.Configs()
	if err != nil {
		ctx.String(400, "%v", err)
		return
	}
	ctx.Bytes(200, "text/plain", res)
}

func (p *ProxyPass) Configs() ([]byte, error) {
	tomlBuf := bytes.Buffer{}
	encoder := toml.NewEncoder(&tomlBuf)
	raw := struct {
		ProxyConfig Config
	}{
		ProxyConfig: p.dupConfig(),
	}
	if err := encoder.Encode(raw); err != nil {
		log.Error("Failed to encode proxy config as toml: %+v: %+v", p.cfg, err)
		return nil, err
	}
	return tomlBuf.Bytes(), nil
}

type GatewayProfile struct {
	GatewayVersion string `json:"gateway_version"`
	SDKVersion     string `json:"sdk_version"`
	ConfigDigest   string `json:"config_digest"`
}

func (p *ProxyPass) profile(ctx *bm.Context) {
	ctx.JSON(p.Profile(), nil)
}

func (p *ProxyPass) Profile() *GatewayProfile {
	return &GatewayProfile{
		GatewayVersion: api.GatewayVersion, //todo sdk理论上不该引用 gateway 的包
		SDKVersion:     sdk.SDKVersion,
		ConfigDigest:   p.getCfgDigest(),
	}
}

func (p *ProxyPass) getCfgDigest() string {
	cfg := p.dupConfig()
	return cfg.Digest()
}
