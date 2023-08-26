package request_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	bm "go-common/library/net/http/blademaster"
	xtime "go-common/library/time"
	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	bmsdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/corehandlers"
	def "go-gateway/app/app-svr/app-gw/sdk/http-sdk/defaults"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/sdkerr"

	"github.com/stretchr/testify/assert"
)

type tempNetworkError struct {
	op     string
	msg    string
	isTemp bool
}

func (e *tempNetworkError) Temporary() bool { return e.isTemp }
func (e *tempNetworkError) Error() string {
	return fmt.Sprintf("%s: %s", e.op, e.msg)
}

var (
	// net.OpError accept, are always temporary
	errAcceptConnectionResetStub = &tempNetworkError{
		isTemp: true, op: "accept", msg: "connection reset",
	}

	// net.OpError read for ECONNRESET is not temporary.
	errReadConnectionResetStub = &tempNetworkError{
		isTemp: false, op: "read", msg: "connection reset",
	}

	// net.OpError write for ECONNRESET may not be temporary, but is treaded as
	// temporary by the SDK.
	errWriteConnectionResetStub = &tempNetworkError{
		isTemp: false, op: "write", msg: "connection reset",
	}

	// net.OpError write for broken pipe may not be temporary, but is treaded as
	// temporary by the SDK.
	errWriteBrokenPipeStub = &tempNetworkError{
		isTemp: false, op: "write", msg: "broken pipe",
	}

	// Generic connection reset error
	errConnectionResetStub = errors.New("connection reset")
)

func TestSanitizeHostForHeader(t *testing.T) {
	cases := []struct {
		url                 string
		exceptedRequestHost string
	}{
		{"https://e.s-e-1.a.com:443", "e.s-e-1.a.com"},
		{"https://e.s-e-1.a.com", "e.s-e-1.a.com"},
		{"https://localhost:9200", "localhost:9200"},
		{"http://localhost:80", "localhost"},
		{"http://localhost:8080", "localhost:8080"},
	}
	for _, c := range cases {
		r, _ := http.NewRequest("GET", c.url, nil)
		request.SanitizeHostForHeader(r)
		assert.Equal(t, r.Host, c.exceptedRequestHost)
	}
}

func TestMakeAddtoUserAgentHandler(t *testing.T) {
	fn := request.MakeAddToUserAgentHandler("name", "version", "extra1", "extra2")
	r := &request.Request{HTTPRequest: &http.Request{Header: http.Header{}}}
	r.HTTPRequest.Header.Set("User-Agent", "foo/bar")
	fn(r)
	ua := "foo/bar name/version (extra1; extra2)"
	assert.Equal(t, ua, r.HTTPRequest.Header.Get("User-Agent"))
}

func TestWithGetResponseHeader(t *testing.T) {
	r := &request.Request{}
	var val, val2 string
	r.ApplyOptions(
		request.WithGetResponseHeader("first-header", &val),
		request.WithGetResponseHeader("second-header", &val2),
	)
	r.HTTPResponse = &http.Response{
		Header: func() http.Header {
			h := http.Header{}
			h.Set("first-header", "first")
			h.Set("second-header", "second")
			return h
		}(),
	}
	r.Handlers.Complete.Run(r)
	assert.Equal(t, "first", val)
	assert.Equal(t, "second", val2)
}

func TestWithGetResponseHeaders(t *testing.T) {
	r := &request.Request{}
	var headers http.Header
	opt := request.WithGetResponseHeaders(&headers)
	r.ApplyOptions(opt)
	r.HTTPResponse = &http.Response{
		Header: func() http.Header {
			h := http.Header{}
			h.Set("a-header", "headerValue")
			return h
		}(),
	}
	r.Handlers.Complete.Run(r)
	assert.Equal(t, "headerValue", headers.Get("a-header"))
}

func TestWithDebug(t *testing.T) {
	r := &request.Request{}
	opt := request.WithDebug(true)
	r.ApplyOptions(opt)
	assert.Equal(t, true, r.Config.Debug)
}

type testHandlerPatcher struct {
	name    string
	matched bool
}

func (d *testHandlerPatcher) Name() string {
	return d.name
}

func (d *testHandlerPatcher) Matched(r *request.Request) bool {
	return d.matched
}

func (d *testHandlerPatcher) Patch(in request.Handlers) request.Handlers {
	in.Send.PushBack(func(r *request.Request) {})
	return in
}

func TestWithHandlerPatchers(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "http://localhost:80", nil)
	r := &request.Request{
		HTTPRequest: httpRequest,
	}
	handlerpatchers := &testHandlerPatcher{
		name:    "d",
		matched: true,
	}
	opt := request.WithHandlerPatchers(handlerpatchers)
	r.ApplyOptions(opt)
	err := r.Send()
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Handlers.Send)
}

func TestSetContext(t *testing.T) {
	r := &request.Request{HTTPRequest: &http.Request{}}
	r.SetContext(context.Background())
	assert.Equal(t, r.Context(), context.Background())
}

func TestRequestWillRetry_ByBody(t *testing.T) {
	info := metadata.ClientInfo{
		AppID:    "dd",
		Endpoint: "http://localhost:80",
	}
	svc := client.New(sdk.Config{}, info, def.Handlers())
	cases := []struct {
		WillRetry   bool
		HTTPMethod  string
		Body        io.ReadSeeker
		IsReqNoBody bool
	}{
		{
			WillRetry:   true,
			HTTPMethod:  "GET",
			Body:        bytes.NewReader([]byte{}),
			IsReqNoBody: true,
		},
		{
			WillRetry:   true,
			HTTPMethod:  "GET",
			Body:        bytes.NewReader(nil),
			IsReqNoBody: true,
		},
		{
			WillRetry:  true,
			HTTPMethod: "POST",
			Body:       bytes.NewReader([]byte("abc123")),
		},
		{
			WillRetry:  true,
			HTTPMethod: "POST",
			Body:       sdk.ReadSeekCloser(bytes.NewReader([]byte("abc123"))),
		},
		{
			WillRetry:   true,
			HTTPMethod:  "GET",
			Body:        sdk.ReadSeekCloser(bytes.NewBuffer(nil)),
			IsReqNoBody: true,
		},
		{
			WillRetry:   true,
			HTTPMethod:  "POST",
			Body:        sdk.ReadSeekCloser(bytes.NewBuffer(nil)),
			IsReqNoBody: true,
		},
		{
			WillRetry:  false,
			HTTPMethod: "POST",
			Body:       sdk.ReadSeekCloser(bytes.NewBuffer([]byte("abc123"))),
		},
	}
	for _, c := range cases {
		req := svc.NewRequest(&request.Operation{
			Name:       "Operation",
			HTTPMethod: c.HTTPMethod,
			HTTPPath:   "/",
		}, nil, nil)
		req.SetReaderBody(c.Body)
		req.Build()
		req.Error = fmt.Errorf("some error")
		req.Retryable = rootsdk.Bool(true)
		req.HTTPResponse = &http.Response{
			StatusCode: 500,
		}
		assert.Equal(t, c.IsReqNoBody, request.NoBody == req.HTTPRequest.Body)
		assert.Equal(t, c.WillRetry, req.WillRetry())
		assert.Equal(t, c.WillRetry, req.WillRetry())
		assert.NotEqual(t, nil, req.Error)
		assert.Equal(t, strings.Contains(req.Error.Error(), "some error"), true)
		assert.Equal(t, 0, req.RetryCount)
	}
}

func TestSetClientInfo(t *testing.T) {
	info := metadata.ClientInfo{
		AppID:    "dd",
		Endpoint: "http://localhost:80",
	}
	r := &request.Request{
		HTTPRequest: &http.Request{},
		Operation: &request.Operation{
			Name:       "a",
			HTTPMethod: "GET",
			HTTPPath:   "/",
		},
	}
	r.SetClientInfo(info)
	assert.Equal(t, r.ClientInfo, info)
}
func TestRequestDuration(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(2)
	r := &request.Request{
		Time:         t1,
		CompleteTime: t2,
	}
	var ti time.Duration
	ti = r.RequestDuration()
	assert.Equal(t, ti, time.Duration(2))
}

func unmarshal(req *request.Request) {
	defer req.HTTPResponse.Body.Close()
	if req.Data != nil {
		json.NewDecoder(req.HTTPResponse.Body).Decode(req.Data)
	}
}

func unmarshalError(req *request.Request) {
	bodyBytes, err := ioutil.ReadAll(req.HTTPResponse.Body)
	if err != nil {
		req.Error = sdkerr.New("UnmarshaleError", req.HTTPResponse.Status, err)
		return
	}
	if len(bodyBytes) == 0 {
		req.Error = sdkerr.NewRequestFailure(
			sdkerr.New("UnmarshaleError", req.HTTPResponse.Status, fmt.Errorf("empty body")),
			req.HTTPResponse.StatusCode,
			"",
		)
		return
	}
	var jsonErr jsonErrorResponse
	if err := json.Unmarshal(bodyBytes, &jsonErr); err != nil {
		req.Error = sdkerr.New("UnmarshaleError", "JSON unmarshal", err)
		return
	}
	req.Error = sdkerr.NewRequestFailure(
		sdkerr.New(jsonErr.Code, jsonErr.Message, nil),
		req.HTTPResponse.StatusCode,
		"",
	)
}

func body(str string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(str)))
}

type testData struct {
	Data string
}

type jsonErrorResponse struct {
	Code    string `json:"__type"`
	Message string `json:"message"`
}

func TestRequestRecoverRetry5xx(t *testing.T) {
	reqNum := 0
	reqs := []http.Response{
		{StatusCode: 500, Body: body(`{"__type":"UnknownError","message":"An error occurred."}`)},
		{StatusCode: 503, Body: body(`{"__type":"UnknownError","message":"An error occurred."}`)},
		{StatusCode: 200, Body: body(`{"data":"valid"}`)},
	}
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "http://endpoint",
	}
	s := client.New(sdk.Config{
		MaxRetries: rootsdk.Int(10),
	}, info, def.Handlers())
	s.Handlers.Validate.Clear()
	s.Handlers.Unmarshal.PushBack(unmarshal)
	s.Handlers.UnmarshalError.PushBack(unmarshalError)
	s.Handlers.Send.Clear()
	s.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &reqs[reqNum]
		reqNum++
	})
	out := &testData{}
	r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, out)
	r.HTTPRequest.Body = http.NoBody
	r.SetBufferBody([]byte{})
	err := r.Send()
	assert.NoError(t, err)
	assert.Equal(t, 2, r.RetryCount)
	assert.Equal(t, "valid", out.Data)
}

func TestRequestRecoverRetry4xx(t *testing.T) {
	reqNum := 0
	reqs := []http.Response{
		{StatusCode: 400, Body: body(`{"__type":"Throttling","message":"Rate exceeded."}`)},
		{StatusCode: 429, Body: body(`{"__type":"FooException","message":"Rate exceeded."}`)},
		{StatusCode: 200, Body: body(`{"data":"valid"}`)},
	}
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "http://endpoint",
	}
	s := client.New(sdk.Config{
		MaxRetries: rootsdk.Int(10),
	}, info, def.Handlers())
	s.Handlers.Validate.Clear()
	s.Handlers.Unmarshal.PushBack(unmarshal)
	s.Handlers.UnmarshalError.PushBack(unmarshalError)
	s.Handlers.Send.Clear()
	s.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &reqs[reqNum]
		reqNum++
	})
	out := &testData{}
	r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, out)
	r.HTTPRequest.Body = http.NoBody
	r.SetBufferBody([]byte{})
	err := r.Send()
	assert.NoError(t, err)
	assert.Equal(t, 2, r.RetryCount)
	assert.Equal(t, "valid", out.Data)
}

func TestRequest4xxUnretryable(t *testing.T) {
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "http://endpoint",
	}
	s := client.New(sdk.Config{
		MaxRetries: rootsdk.Int(1),
	}, info, def.Handlers())
	s.Handlers.Validate.Clear()
	s.Handlers.Unmarshal.PushBack(unmarshal)
	s.Handlers.UnmarshalError.PushBack(unmarshalError)
	s.Handlers.Send.Clear() // mock sending
	s.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &http.Response{
			StatusCode: 401,
			Body:       body(`{"__type":"SignatureDoesNotMatch","message":"Signature does not match."}`),
		}
	})
	out := &testData{}
	r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, out)
	r.HTTPRequest.Body = http.NoBody
	r.SetBufferBody([]byte{})
	err := r.Send()
	if err == nil {
		t.Fatalf("expect error, but did not get one")
	}
	asrr := err.(sdkerr.RequestFailure)
	assert.Equal(t, 401, asrr.StatusCode())
	assert.Equal(t, "SignatureDoesNotMatch", asrr.Code())
	assert.Equal(t, "Signature does not match.", asrr.Message())
	assert.Equal(t, 0, r.RetryCount)
}

func TestRequestExhaustRetries(t *testing.T) {
	reqNum := 0
	reqs := []http.Response{
		{StatusCode: 500, Body: body(`{"__type":"UnknownError","message":"An error occurred."}`)},
		{StatusCode: 500, Body: body(`{"__type":"UnknownError","message":"An error occurred."}`)},
		{StatusCode: 500, Body: body(`{"__type":"UnknownError","message":"An error occurred."}`)},
		{StatusCode: 500, Body: body(`{"__type":"UnknownError","message":"An error occurred."}`)},
	}
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "http://endpoint",
	}
	s := client.New(sdk.Config{
		MaxRetries: rootsdk.Int(2),
	}, info, def.Handlers())
	s.Handlers.Validate.Clear()
	s.Handlers.Unmarshal.PushBack(unmarshal)
	s.Handlers.UnmarshalError.PushBack(unmarshalError)
	s.Handlers.Send.Clear()
	s.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &reqs[reqNum]
		reqNum++
	})
	out := &testData{}
	r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, out)
	r.HTTPRequest.Body = http.NoBody
	r.SetBufferBody([]byte{})
	err := r.Send()
	if err == nil {
		t.Fatalf("expect error, but did not get one")
	}
	assert.Equal(t, 2, r.RetryCount)
}

func Test501NotRetrying(t *testing.T) {
	reqNum := 0
	reqs := []http.Response{
		{StatusCode: 500, Body: body(`{"__type":"UnknownError","message":"An error occurred."}`)},
		{StatusCode: 501, Body: body(`{"__type":"NotImplemented","message":"An error occurred."}`)},
		{StatusCode: 200, Body: body(`{"data":"valid"}`)},
	}
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "http://endpoint",
	}
	s := client.New(sdk.Config{
		MaxRetries: rootsdk.Int(10),
	}, info, def.Handlers())
	s.Handlers.Validate.Clear()
	s.Handlers.Unmarshal.PushBack(unmarshal)
	s.Handlers.UnmarshalError.PushBack(unmarshalError)
	s.Handlers.Send.Clear()
	s.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &reqs[reqNum]
		reqNum++
	})
	out := &testData{}
	r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, out)
	r.HTTPRequest.Body = http.NoBody
	r.SetBufferBody([]byte{})
	err := r.Send()
	if err == nil {
		t.Fatal("expect error, but got none")
	}
	serr := err.(sdkerr.Error)
	assert.Equal(t, "NotImplemented", serr.Code())
	assert.Equal(t, 1, r.RetryCount)
}

func TestIsNoBodyReader(t *testing.T) {
	cases := []struct {
		reader io.ReadCloser
		expect bool
	}{
		{ioutil.NopCloser(bytes.NewReader([]byte("abc"))), false},
		{ioutil.NopCloser(bytes.NewReader(nil)), false},
		{nil, false},
		{request.NoBody, true},
	}

	for _, c := range cases {
		assert.Equal(t, c.expect, request.NoBody == c.reader)
	}
}

func TestIsSerializationErrorRetryable(t *testing.T) {
	testCases := []struct {
		err      error
		expected bool
	}{
		{
			err:      sdkerr.New(request.ErrCodeSerialization, "foo error", nil),
			expected: false,
		},
		{
			err:      sdkerr.New("ErrFoo", "foo error", nil),
			expected: false,
		},
		{
			err:      nil,
			expected: false,
		},
		{
			err:      sdkerr.New(request.ErrCodeSerialization, "foo error", errAcceptConnectionResetStub),
			expected: true,
		},
	}

	for _, c := range testCases {
		r := &request.Request{
			Error: c.err,
		}
		assert.Equal(t, r.IsErrorRetryable(), c.expected)
	}
}

type stubSeekFail struct {
	Err error
}

func (f *stubSeekFail) Read(b []byte) (int, error) {
	return len(b), nil
}
func (f *stubSeekFail) ReadAt(b []byte, offset int64) (int, error) {
	return len(b), nil
}
func (f *stubSeekFail) Seek(offset int64, mode int) (int64, error) {
	return 0, f.Err
}

func TestRequestBodySeekFails(t *testing.T) {
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "http://endpoint",
	}
	s := client.New(sdk.Config{}, info, def.Handlers())
	s.Handlers.Validate.Clear()
	s.Handlers.Build.Clear()
	out := &testData{}
	r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, out)
	r.SetReaderBody(&stubSeekFail{
		Err: fmt.Errorf("failed to seek reader"),
	})
	err := r.Send()
	if err == nil {
		t.Fatal("expect error, but got none")
	}
	assert.NotEmpty(t, r.Error)
}

func TestRequestEndpointWithDefaultPort(t *testing.T) {
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "https://endpoint:443",
	}
	s := client.New(sdk.Config{}, info, def.Handlers())
	r := s.NewRequest(&request.Operation{
		Name:       "d",
		HTTPMethod: "GET",
		HTTPPath:   "/",
	}, nil, nil)
	r.Handlers.Validate.Clear()
	r.Handlers.ValidateResponse.Clear()
	r.Handlers.Send.Clear()
	r.Handlers.Send.PushFront(func(r *request.Request) {
		req := r.HTTPRequest
		assert.Equal(t, "endpoint", req.Host)
		assert.Equal(t, "https://endpoint:443/", req.URL.String())
	})
	err := r.Send()
	assert.NoError(t, err)
}

func TestRequestEndpointWithNonDefaultPort(t *testing.T) {
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "https://endpoint:8443",
	}
	s := client.New(sdk.Config{}, info, def.Handlers())
	r := s.NewRequest(&request.Operation{
		Name:       "d",
		HTTPMethod: "GET",
		HTTPPath:   "/",
	}, nil, nil)
	r.Handlers.Validate.Clear()
	r.Handlers.ValidateResponse.Clear()
	r.Handlers.Send.Clear()
	r.Handlers.Send.PushFront(func(r *request.Request) {
		req := r.HTTPRequest
		assert.Equal(t, "", req.Host)
		assert.Equal(t, "https://endpoint:8443/", req.URL.String())
	})
	err := r.Send()
	assert.NoError(t, err)
}

func TestRequestMarshaledEndpointWithDefaultPort(t *testing.T) {
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "https://endpoint:443",
	}
	s := client.New(sdk.Config{}, info, def.Handlers())
	r := s.NewRequest(&request.Operation{
		Name:       "d",
		HTTPMethod: "GET",
		HTTPPath:   "/",
	}, nil, nil)
	r.Handlers.Validate.Clear()
	r.Handlers.ValidateResponse.Clear()
	r.Handlers.Build.PushBack(func(r *request.Request) {
		req := r.HTTPRequest
		req.URL.Host = "d." + req.URL.Host
	})
	r.Handlers.Send.Clear()
	r.Handlers.Send.PushFront(func(r *request.Request) {
		req := r.HTTPRequest
		assert.Equal(t, "d.endpoint", req.Host)
		assert.Equal(t, "https://d.endpoint:443/", req.URL.String())
	})
	err := r.Send()
	assert.NoError(t, err)
}

func TestRequestMarshaledEndpointWithNonDefaultPort(t *testing.T) {
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "https://endpoint:8443",
	}
	s := client.New(sdk.Config{}, info, def.Handlers())
	r := s.NewRequest(&request.Operation{
		Name:       "d",
		HTTPMethod: "GET",
		HTTPPath:   "/",
	}, nil, nil)
	r.Handlers.Validate.Clear()
	r.Handlers.ValidateResponse.Clear()
	r.Handlers.Build.PushBack(func(r *request.Request) {
		req := r.HTTPRequest
		req.URL.Host = "d." + req.URL.Host
	})
	r.Handlers.Send.Clear()
	r.Handlers.Send.PushFront(func(r *request.Request) {
		req := r.HTTPRequest
		assert.Equal(t, "", req.Host)
		assert.Equal(t, "https://d.endpoint:8443/", req.URL.String())
	})
	err := r.Send()
	assert.NoError(t, err)
}

type timeoutErr struct {
	error
}

var errTimeout = sdkerr.New("foo", "bar", &timeoutErr{
	errors.New("net/http: request canceled"),
})

func (e *timeoutErr) Timeout() bool {
	return true
}

func (e *timeoutErr) Temporary() bool {
	return true
}

func TestRequestRecoverTimeoutWithNilBody(t *testing.T) {
	reqNum := 0
	reqs := []*http.Response{
		{StatusCode: 0, Body: nil},
		{StatusCode: 200, Body: body(`{"data":"valid"}`)},
	}
	errors := []error{
		errTimeout, nil,
	}
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "https://endpoint",
	}
	s := client.New(sdk.Config{}, info, def.Handlers())
	s.Handlers.Validate.Clear()
	s.Handlers.Unmarshal.PushBack(unmarshal)
	s.Handlers.UnmarshalError.PushBack(unmarshalError)
	s.Handlers.AfterRetry.Clear()
	s.Handlers.AfterRetry.PushBack(func(r *request.Request) {
		if r.Error != nil {
			r.Error = nil
			r.Retryable = rootsdk.Bool(true)
			r.RetryCount++
		}
	})
	s.Handlers.Send.Clear()
	s.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = reqs[reqNum]
		r.Error = errors[reqNum]
		reqNum++
	})
	out := &testData{}
	r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, out)
	r.HTTPRequest.Body = http.NoBody
	r.SetBufferBody([]byte{})
	err := r.Send()
	assert.NoError(t, err)
	assert.Equal(t, 1, r.RetryCount)
	assert.Equal(t, "valid", out.Data)
}

func TestRequestRecoverTimeoutWithNilResponse(t *testing.T) {
	reqNum := 0
	reqs := []*http.Response{
		nil,
		{StatusCode: 200, Body: body(`{"data":"valid"}`)},
	}
	errors := []error{
		errTimeout,
		nil,
	}
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "https://endpoint",
	}
	s := client.New(sdk.Config{}, info, def.Handlers())
	s.Handlers.Validate.Clear()
	s.Handlers.Unmarshal.PushBack(unmarshal)
	s.Handlers.UnmarshalError.PushBack(unmarshalError)
	s.Handlers.AfterRetry.Clear()
	s.Handlers.AfterRetry.PushBack(func(r *request.Request) {
		if r.Error != nil {
			r.Error = nil
			r.Retryable = rootsdk.Bool(true)
			r.RetryCount++
		}
	})
	s.Handlers.Send.Clear()
	s.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = reqs[reqNum]
		r.Error = errors[reqNum]
		reqNum++
	})
	out := &testData{}
	r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, out)
	r.HTTPRequest.Body = http.NoBody
	r.SetBufferBody([]byte{})
	err := r.Send()
	assert.NoError(t, err)
	assert.Equal(t, 1, r.RetryCount)
	assert.Equal(t, "valid", out.Data)
}

type testRetryer struct {
	shouldRetry bool
	maxRetries  int
}

func (d *testRetryer) MaxRetries() int {
	return d.maxRetries
}

// RetryRules returns the delay duration before retrying this request again
func (d *testRetryer) RetryRules(r *request.Request) time.Duration {
	return 0
}

func (d *testRetryer) ShouldRetry(r *request.Request) bool {
	return d.shouldRetry
}

func TestEnforceShouldRetryCheck(t *testing.T) {
	retryer := &testRetryer{
		shouldRetry: true, maxRetries: 3,
	}
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "https://endpoint",
	}
	s := client.New(sdk.Config{
		MaxRetries: rootsdk.Int(0),
		Retryer:    retryer,
	}, info, def.Handlers())
	s.Handlers.Validate.Clear()
	s.Handlers.Send.Swap(corehandlers.SendHandler.Name, request.NamedHandler{
		Name: "TestEnforceShouldRetryCheck",
		Fn: func(r *request.Request) {
			r.HTTPResponse = &http.Response{
				Header: http.Header{},
				Body:   ioutil.NopCloser(bytes.NewBuffer(nil)),
			}
			r.Retryable = rootsdk.Bool(true)
		},
	})
	s.Handlers.Unmarshal.PushBack(unmarshal)
	s.Handlers.UnmarshalError.PushBack(unmarshalError)
	out := &testData{}
	r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, out)
	r.HTTPRequest.Body = http.NoBody
	r.SetBufferBody([]byte{})
	err := r.Send()
	if err == nil {
		t.Fatalf("expect error, but got nil")
	}
	assert.Equal(t, 3, r.RetryCount)
}

func TestRequest_TemporaryRetry(t *testing.T) {
	done := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1024")
		w.WriteHeader(http.StatusOK)
		w.Write(make([]byte, 100))
		f := w.(http.Flusher)
		f.Flush()
		<-done
	}))
	defer server.Close()
	bmCli := bm.NewClient(&bm.ClientConfig{
		App: &bm.App{
			Key:    "53e2fa226f5ad348",
			Secret: "3cf6bd1b0ff671021da5f424fea4b04a",
		},
		Timeout: xtime.Duration(time.Second),
	})
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: server.URL,
	}
	s := client.New(sdk.Config{
		Client:     bmsdk.WrapClient(bmCli),
		MaxRetries: rootsdk.Int(1),
	}, info, def.Handlers())
	req := s.NewRequest(&request.Operation{
		Name: "name", HTTPMethod: "GET", HTTPPath: "/path",
	}, &struct{}{}, &struct{}{})
	req.Handlers.Unmarshal.PushBack(func(r *request.Request) {
		defer req.HTTPResponse.Body.Close()
		_, err := io.Copy(ioutil.Discard, req.HTTPResponse.Body)
		r.Error = sdkerr.New(request.ErrCodeSerialization, "error", err)
	})
	req.HTTPRequest.Body = http.NoBody
	req.SetBufferBody([]byte{})
	err := req.Send()
	if err == nil {
		t.Errorf("expect error, got none")
	}
	close(done)
	assert.Equal(t, 1, req.RetryCount)
}

func TestRequestThrottleRetries(t *testing.T) {
	reqNum := 0
	reqs := []http.Response{
		{StatusCode: 500, Body: body(`{"__type":"Throttling","message":"An error occurred."}`)},
		{StatusCode: 500, Body: body(`{"__type":"Throttling","message":"An error occurred."}`)},
		{StatusCode: 500, Body: body(`{"__type":"Throttling","message":"An error occurred."}`)},
		{StatusCode: 500, Body: body(`{"__type":"Throttling","message":"An error occurred."}`)},
	}
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "http://endpoint",
	}
	s := client.New(sdk.Config{}, info, def.Handlers())
	s.Handlers.Validate.Clear()
	s.Handlers.Unmarshal.PushBack(unmarshal)
	s.Handlers.UnmarshalError.PushBack(unmarshalError)
	s.Handlers.Send.Clear() // mock sending
	s.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &reqs[reqNum]
		reqNum++
	})
	r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, nil)
	r.HTTPRequest.Body = http.NoBody
	r.SetBufferBody([]byte{})
	err := r.Send()
	if err == nil {
		t.Fatalf("expect error, but did not get one")
	}
	assert.Equal(t, 1, r.RetryCount)
}

func TestRequestUserAgent(t *testing.T) {
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "http://endpoint",
	}
	s := client.New(sdk.Config{}, info, def.Handlers())
	req := s.NewRequest(&request.Operation{Name: "Operation"}, nil, &testData{})
	req.HTTPRequest.Header.Set("User-Agent", "foo/bar")
	err := req.Build()
	assert.NoError(t, err)
	expectUA := fmt.Sprintf("foo/bar %s/%s (%s; %s; %s)",
		rootsdk.SDKName, rootsdk.SDKVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	e, a := expectUA, req.HTTPRequest.Header.Get("User-Agent")
	assert.Equal(t, e, a)
}

func TestWithRetryer(t *testing.T) {
	reqNum := 0
	reqs := []http.Response{
		{StatusCode: 500, Body: body(`{"__type":"Throttling","message":"An error occurred."}`)},
		{StatusCode: 200, Body: body(`{"data":"valid"}`)},
	}
	retryer := &testRetryer{
		shouldRetry: true, maxRetries: 1,
	}
	cfg := sdk.Config{}
	cfgr := request.WithRetryer(&sdk.Config{}, retryer)
	cfg.Retryer = cfgr.Retryer
	info := metadata.ClientInfo{
		AppID:    "",
		Endpoint: "http://endpoint",
	}
	s := client.New(cfg, info, def.Handlers())
	s.Handlers.Validate.Clear()
	s.Handlers.Unmarshal.PushBack(unmarshal)
	s.Handlers.UnmarshalError.PushBack(unmarshalError)
	s.Handlers.Send.Clear()
	s.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &reqs[reqNum]
		reqNum++
	})
	req := s.NewRequest(&request.Operation{Name: "Operation"}, nil, nil)
	req.HTTPRequest.Body = http.NoBody
	req.SetBufferBody([]byte{})
	err := req.Send()
	assert.NoError(t, err)
	assert.Equal(t, 1, req.RetryCount)
}
