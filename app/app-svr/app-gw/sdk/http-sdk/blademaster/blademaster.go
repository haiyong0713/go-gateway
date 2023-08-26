package blademaster

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	bmsentry "go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster/sentry"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"
	def "go-gateway/app/app-svr/app-gw/sdk/http-sdk/defaults"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/sdkerr"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// ProxyPass is
type ProxyPass struct {
	sync.RWMutex
	cfg Config

	cli *client.Client
}

// ResponseProxyHandler is a named request handler for unmarshaling query protocol requests
var ResponseProxyHandler = request.NamedHandler{Name: "appgwsdk.blademaster.ResponseBodyProxy", Fn: ResponseBodyProxy}

// ResponseBodyProxy is
func ResponseBodyProxy(r *request.Request) {
	defer r.HTTPResponse.Body.Close()
	if !r.DataFilled() {
		return
	}
	dst := r.Data.(*bytes.Buffer)
	dst.Reset()
	if _, err := io.Copy(dst, r.HTTPResponse.Body); err != nil {
		r.Error = errors.WithStack(err)
	}
}

func requestQueryBuild(in url.Values) request.NamedHandler {
	return request.NamedHandler{Name: "appgwsdk.blademaster.RequestQueryBuild", Fn: func(r *request.Request) {
		r.HTTPRequest.URL.RawQuery = in.Encode()
	}}
}

type wrappedClient struct {
	*bm.Client
}

func (w wrappedClient) DoRequest(ctx context.Context, req *http.Request, metricURI string) (*http.Response, error) {
	resp, body, err := w.RawResponse(ctx, req, metricURI)
	if err != nil {
		return nil, err
	}
	resp.Body = sdk.ReadSeekCloser(bytes.NewBuffer(body))
	return resp, nil
}

// WrapClient is
func WrapClient(bmCli *bm.Client) wrappedClient {
	return wrappedClient{
		Client: bmCli,
	}
}

// New is
func New(cfg Config, sdkCfg sdk.Config, info metadata.ClientInfo) *ProxyPass {
	handlers := def.Handlers()
	handlers.Build.Clear()
	handlers.Unmarshal.PushBackNamed(ResponseProxyHandler)
	handlers.Unmarshal.PushBackNamed(UnmarshalECodeHandler)
	handlers.CompleteAttempt.PushBackNamed(ReportUpstreamAttemptHandler)
	handlers.Complete.PushBackNamed(ReportClientMetricsCompleteHandler)
	handlers.Complete.PushBackNamed(ReportUpstreamCompleteHandler)
	handlers.Retry.Clear()
	handlers.ValidateResponse.Clear()
	cli := client.New(sdkCfg, info, handlers)
	pp := &ProxyPass{cli: cli}
	if err := pp.ensureConfig(cfg); err != nil {
		panic(err)
	}
	return pp
}

func (p *ProxyPass) SetupRouter(e *bm.Engine) {
	e.Use(bmsentry.Default())
	e.Use(ABEnv{})
	e.Use(metricReporter{})
	e.NoRoute(p.ServeHTTP)
	e.GET("/_/metrics.json", p.metrics)
	e.GET("/_/metrics", metricsPage)
	e.GET("/_/configs.toml", p.configs)
	e.GET("/_/profile", p.profile)
	e.POST("/_/reload", p.reloadCfg)
}

func (p *ProxyPass) reloadCfg(ctx *bm.Context) {
	req := &struct {
		Digest         string `form:"digest" validate:"required"`
		Content        string `form:"content" validate:"required"`
		OriginalDigest string `form:"original_digest" validate:"required"`
	}{}
	if err := ctx.Bind(req); err != nil {
		return
	}

	raw := struct {
		ProxyConfig *Config
	}{}
	if _, err := toml.Decode(req.Content, &raw); err != nil {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("invalid sdk builder config: %+v", err)))
		return
	}
	if raw.ProxyConfig == nil {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "empty proxy config"))
		return
	}

	cfgDigest := raw.ProxyConfig.Digest()
	if cfgDigest != req.Digest {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("mismatched digest: %q != %q", cfgDigest, req.Digest)))
		return
	}
	if err := p.CasReload(*raw.ProxyConfig, req.OriginalDigest); err != nil {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("failed to load provided config: %+v", err)))
		return
	}

	reply := struct {
		Loaded string `json:"loaded"`
		Digest string `json:"digest"`
	}{}
	loaded := p.dupConfig()
	buf := &bytes.Buffer{}
	//nolint:errcheck
	toml.NewEncoder(buf).Encode(loaded)
	//nolint:gosimple
	reply.Loaded = string(buf.Bytes())
	reply.Digest = loaded.Digest()
	ctx.JSON(reply, nil)
	//nolint:gosimple
	return
}

func (p *ProxyPass) CasReload(cfg Config, originalDigest string) error {
	for _, pm := range cfg.DynPath {
		if err := pm.InitStatic(); err != nil {
			log.Error("Failed to load dynamic path: %+v", err)
			return err
		}
		log.Info("Parsed a new proxy rule: %+v", pm)
	}
	p.Lock()
	defer p.Unlock()
	if p.cfg.Digest() != originalDigest {
		return errors.Errorf("invalid original digest: %s, %s", p.cfg.Digest(), originalDigest)
	}
	p.cfg = cfg
	log.Info("Succeeded to reload proxy config: %+v", p.cfg)
	return nil
}

func (p *ProxyPass) Reload(cfg Config) error {
	return p.ensureConfig(cfg)
}

func (p *ProxyPass) ensureConfig(cfg Config) error {
	for _, pm := range cfg.DynPath {
		if err := pm.InitStatic(); err != nil {
			log.Error("Failed to load dynamic path: %+v", err)
			return err
		}
		log.Info("Parsed a new proxy rule: %+v", pm)
	}
	p.Lock()
	defer p.Unlock()
	p.cfg = cfg
	log.Info("Succeeded to load dynamic path: %+v", p.cfg)
	return nil
}

func copyRequestHeader(src, dst *http.Request) {
	for k, v := range src.Header {
		dst.Header[k] = v
	}
	dst.Host = src.Host // host header is seperated in field
}

func asResponse(resp *http.Response, body io.Reader, w http.ResponseWriter) error {
	hDst := w.Header()
	for k, v := range resp.Header {
		hDst[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, body); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (p *ProxyPass) dupConfig() Config {
	p.RLock()
	cfg := p.cfg
	p.RUnlock()
	return cfg
}

// MatchLongestPath is
func (p *ProxyPass) MatchLongestPath(url *url.URL) (*PathMeta, bool) {
	matched := []*PathMeta{}
	p.RLock()
	for _, p := range p.cfg.DynPath {
		if !p.matcher.Match(url) {
			continue
		}
		matched = append(matched, p)
	}
	p.RUnlock()

	if len(matched) <= 0 {
		return nil, false
	}

	sort.Slice(matched, func(i, j int) bool {
		l, r := matched[i].matcher, matched[j].matcher
		if l.Priority() < r.Priority() {
			return true
		}
		if l.Priority() == r.Priority() {
			return l.Len() > r.Len()
		}
		return false
	})
	return matched[0], true
}

func (p *ProxyPass) ServeHTTP(ctx *bm.Context) {
	defer ctx.Abort()
	pm, ok := p.MatchLongestPath(ctx.Request.URL)
	if !ok {
		ctx.String(404, "%v", errors.Errorf("no matched proxy rule: %s", ctx.Request.URL.Path))
		log.Warn("no matched proxy rule: %q", ctx.Request.URL.Path)
		reportProxyStatusCode(UnexpectedPath, ctx.Writer)
		return
	}
	defer reportProxyStatusCode(pm.Pattern, ctx.Writer)

	req := ctx.Request
	op := &request.Operation{
		Name:       fmt.Sprintf("proxy-for:%s", req.URL.Path),
		HTTPMethod: req.Method,
		HTTPPath:   req.URL.EscapedPath(),
	}
	bodyBuffer := rebuildBody(req)

	query := req.URL.Query()
	rcvBody := &bytes.Buffer{}
	r := p.cli.NewRequest(op, nil, rcvBody)
	copyRequestHeader(req, r.HTTPRequest)
	r.HTTPRequest.Header.Del("Content-Length") // force re-calculate the content length
	r.SetBufferBody(bodyBuffer.Bytes())
	r.SetContext(ctx)
	r.Handlers.Build.PushBackNamed(requestQueryBuild(query))

	// validator
	if pm.validator != nil {
		r.Handlers.ValidateResponse.PushBackNamed(wrapValidator(pm.validator))
	}

	// endpoint
	clientInfo := r.ClientInfo
	if pm.ClientInfo.Timeout > 0 {
		clientInfo.Timeout = pm.ClientInfo.Timeout
	}
	if pm.ClientInfo.Endpoint != "" {
		clientInfo.Endpoint = pm.ClientInfo.Endpoint
		clientInfo.AppID = pm.ClientInfo.Endpoint
	}
	if pm.ClientInfo.AppID != "" {
		clientInfo.AppID = pm.ClientInfo.AppID
	}
	if pm.ClientInfo.MaxRetries != nil {
		clientInfo.MaxRetries = pm.ClientInfo.MaxRetries
		r.Retryer = client.DefaultRetryer{NumMaxRetries: int(rootsdk.Int64Value(pm.ClientInfo.MaxRetries))}
	}
	//nolint:errcheck
	r.SetClientInfo(clientInfo)
	setRequestMerticURI(r, req.URL.EscapedPath())

	// signing key
	if pm.SDKConfig.Key != "" {
		r.Config.Key = pm.SDKConfig.Key
	}
	if pm.SDKConfig.Secret != "" {
		r.Config.Secret = pm.SDKConfig.Secret
	}

	// retry option
	setupBackupRetry(r, &pm.BackupRetryOption)
	// rate limiter option
	setupRateLimiter(r, pm, pm.RateLimiterOptions)

	err := r.Send()
	err = errors.Cause(err)
	switch causedErr := err.(type) {
	case nil:
	case ecode.Codes:
		// This is an exactly API error and cannot be retry to succeed.
		// Write response directly.
		r.Error = nil
	case sdkerr.Error:
		// sdkerr.Error is the sdk internal Error, should be return directly.
		origErr := causedErr.OrigErr()
		captureException(ctx, err)
		ctx.JSON(nil, origErr)
		return
	default:
		// This is an unknow error.
		// We should write error to response.
		captureException(ctx, err)
		ctx.JSON(nil, err)
		return
	}
	if r.HTTPResponse.StatusCode == 200 {
		ctx.RoutePath = req.URL.Path
	}
	if err := asResponse(r.HTTPResponse, rcvBody, ctx.Writer); err != nil {
		log.Error("Failed to write response: %+v", err)
		captureException(ctx, err)
		return
	}
	//nolint:gosimple
	return
}

func captureException(ctx *bm.Context, err error) {
	if err == nil {
		return
	}
	hub := bmsentry.GetHubFromContext(ctx)
	if hub != nil {
		hub.CaptureException(err)
	}
}

func readMultipartFile(fh *multipart.FileHeader) ([]byte, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

func rebuildMultipart(req *http.Request) *bytes.Buffer {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for name, fh := range req.MultipartForm.File {
		for _, fi := range fh {
			p, err := writer.CreateFormFile(name, fi.Filename)
			if err != nil {
				continue
			}
			bs, err := readMultipartFile(fi)
			if err != nil {
				continue
			}
			//nolint:errcheck
			p.Write(bs)
		}
	}
	for name, values := range req.MultipartForm.Value {
		for _, value := range values {
			if err := writer.WriteField(name, value); err != nil {
				continue
			}
		}
	}
	if err := writer.Close(); err != nil {
		return nil
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return body
}

func rebuildBody(req *http.Request) *bytes.Buffer {
	// GET request
	if req.Body == nil {
		return bytes.NewBuffer([]byte{})
	}

	// rebuild from multipart form
	ctype := req.Header.Get("Content-Type")
	if strings.Contains(ctype, "multipart/form-data") && req.MultipartForm != nil {
		return rebuildMultipart(req)
	}
	// rebuild from post form
	if len(req.PostForm) > 0 {
		return bytes.NewBuffer([]byte(req.PostForm.Encode()))
	}

	// copy the original body
	bodyBytes, _ := ioutil.ReadAll(req.Body)
	return bytes.NewBuffer(bodyBytes)
}

func setRequestMerticURI(r *request.Request, metricURI string) {
	if metricURI != "" {
		r.ApplyOptions(request.WithMetricURI(metricURI))
	}
}
