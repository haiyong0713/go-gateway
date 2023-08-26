package blademaster

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"text/template"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster/ab"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/prom"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"

	"github.com/pkg/errors"
)

var (
	ToBackupServiceCode = prom.New().
		WithCounter("http_backup_call_count", []string{"path"}).
		WithState("http_backup_call_count_state", []string{"path"})
)

var _errStopAttempt = errors.New("ForceBackupPatcher: stop attempt")

type BackupRetryOption struct {
	Ratio                int64
	ForceBackupCondition string
	forceBackupCondition ab.Condition

	BackupAction      string
	BackupPlaceholder string
	BackupECode       int64
	BackupURL         string
	backupURL         *url.URL
}

func replaceURL(dst, src *url.URL) {
	dst.Host = src.Host
	dst.Scheme = src.Scheme
	if src.Path != "" {
		dst.Path = src.Path
	}
}

func constructJSONP(body []byte, callback string) []byte {
	out := bytes.Buffer{}
	out.Write([]byte(callback))
	out.Write([]byte("("))
	out.Write(body)
	out.Write([]byte(")"))
	return out.Bytes()
}

func parseJSONPCallback(req *request.Request) string {
	rawHTTPReq := req.HTTPRequest
	params := rawHTTPReq.URL.Query()
	cb := params.Get("callback")
	cb = template.JSEscapeString(cb)
	return cb
}

func dummySuccessResponse(body []byte, req *request.Request) *http.Response {
	header := http.Header{}
	header.Set("content-type", "application/json; charset=utf-8")

	cb := parseJSONPCallback(req)
	if cb != "" {
		body = constructJSONP(body, cb)
	}
	return &http.Response{
		StatusCode: 200,
		Status:     http.StatusText(200),
		Header:     header,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}
}

func dummyECodeResponse(ec int64, req *request.Request) *http.Response {
	header := http.Header{}
	header.Set("content-type", "application/json; charset=utf-8")
	ecErr := ecode.Int(int(ec))
	body := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		TTL     int64  `json:"ttl"`
	}{
		Code:    ecErr.Code(),
		Message: ecErr.Message(),
		TTL:     1,
	}
	v, _ := json.Marshal(body)

	cb := parseJSONPCallback(req)
	if cb != "" {
		v = constructJSONP(v, cb)
	}
	return &http.Response{
		StatusCode: 200,
		Status:     http.StatusText(200),
		Header:     header,
		Body:       ioutil.NopCloser(bytes.NewReader(v)),
	}
}

func setupBackupRetry(r *request.Request, option *BackupRetryOption) {
	r.ApplyOptions(request.WithHandlerPatchers(NewForceBackupPatcher(option)))
}

// ForceBackupPatcher is
type ForceBackupPatcher struct {
	option  *BackupRetryOption
	Attempt int
}

func NewForceBackupPatcher(option *BackupRetryOption) *ForceBackupPatcher {
	return &ForceBackupPatcher{
		option:  option,
		Attempt: 1,
	}
}

// Name is
func (bp *ForceBackupPatcher) Name() string {
	return "ForceBackupPatcher"
}

// Matched is
func (bp *ForceBackupPatcher) Matched(r *request.Request) bool {
	if bp.option.Ratio == 0 {
		return false
	}
	if bp.option.Ratio < 100 && rand.Int63n(100)+1 > bp.option.Ratio {
		return false
	}
	t, ok := ab.FromContext(r.Context())
	if !ok {
		// all request should be matched if no ab environment is set
		return true
	}
	if bp.option.forceBackupCondition != nil {
		return bp.option.forceBackupCondition.Matches(t)
	}
	return true
}

// Patch is
func (bp *ForceBackupPatcher) Patch(in request.Handlers) request.Handlers {
	out := in.Copy()
	out.Send.SetFrontNamed(request.NamedHandler{
		Name: "appgwsdk.blademaster.ab.ForceBackupPatcherSendHandler",
		Fn: func(r *request.Request) {
			switch bp.option.BackupAction {
			case "placeholder":
				//nolint
				r.HTTPResponse = dummySuccessResponse([]byte(bp.option.BackupPlaceholder), r)
				r.Error = _errStopAttempt
				return
			case "ecode":
				//nolint
				r.HTTPResponse = dummyECodeResponse(bp.option.BackupECode, r)
				r.Error = _errStopAttempt
				return
			case "retry_backup":
				if r.RetryCount < bp.Attempt {
					return
				}
				replaceURL(r.HTTPRequest.URL, bp.option.backupURL)
				r.MetricURI = bp.option.BackupURL
				ToBackupServiceCode.Incr(r.Operation.HTTPPath)
				return
			case "directly_backup":
				replaceURL(r.HTTPRequest.URL, bp.option.backupURL)
				r.MetricURI = bp.option.BackupURL
				ToBackupServiceCode.Incr(r.Operation.HTTPPath)
				return
			default:
				log.Warn("Unrecognized backup action: %s", bp.option.BackupAction)
			}
		},
	})
	out.Send.AfterEachFn = func(item request.HandlerListRunItem) bool {
		if item.Request.Error == _errStopAttempt {
			item.Request.Error = nil
			return false
		}
		return request.HandlerListStopOnError(item)
	}
	out.ValidateResponse.SetFrontNamed(request.NamedHandler{
		Name: "appgwsdk.blademaster.ab.ForceBackupPatcherCompleteAttemptHandler",
		Fn: func(r *request.Request) {
			if r.Error == _errStopAttempt {
				r.Error = nil
				return
			}
		},
	})
	return out
}
