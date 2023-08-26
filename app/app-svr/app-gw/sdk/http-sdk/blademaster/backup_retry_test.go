package blademaster

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"testing"
	"time"

	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster/ab"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"

	"github.com/stretchr/testify/assert"
)

func TestReplaceURLSchema(t *testing.T) {
	URL1, _ := url.Parse("http://api.bilibili.com")
	URL2, _ := url.Parse("discovery://web.interface")
	URL3, _ := url.Parse("discovery://web.interface/")
	URL4, _ := url.Parse("discovery://web.interface/path")

	reqURL, _ := url.Parse("http://api.bilibili.com/x/web-interface/dynamic/region")

	dst1 := *reqURL
	replaceURL(&dst1, URL1)
	t.Logf("%s", dst1.String())
	assert.Equal(t, dst1.String(), "http://api.bilibili.com/x/web-interface/dynamic/region")

	dst2 := *reqURL
	replaceURL(&dst2, URL2)
	t.Logf("%s", dst2.String())
	assert.Equal(t, dst2.String(), "discovery://web.interface/x/web-interface/dynamic/region")

	dst3 := *reqURL
	replaceURL(&dst3, URL3)
	t.Logf("%s", dst3.String())
	assert.Equal(t, dst3.String(), "discovery://web.interface/")

	dst4 := *reqURL
	replaceURL(&dst4, URL4)
	t.Logf("%s", dst4.String())
	assert.Equal(t, dst4.String(), "discovery://web.interface/path")
}

func TestMatched(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "http://localhost:80", nil)
	r := &request.Request{
		HTTPRequest: httpRequest,
	}
	abT := ab.New()
	ctx := ab.NewContext(context.Background(), abT)
	r.SetContext(ctx)
	option := &BackupRetryOption{
		Ratio: 200,
	}
	setupBackupRetry(r, option)
	err := r.Send()
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Handlers)
}

func TestRatioNotMatched(t *testing.T) {
	//Ratio not match
	httpRequest, _ := http.NewRequest("GET", "http://localhost:80", nil)
	r := &request.Request{
		HTTPRequest: httpRequest,
	}
	option := &BackupRetryOption{
		Ratio: 90,
	}
	setupBackupRetry(r, option)
	err := r.Send()
	assert.NoError(t, err)
	assert.Empty(t, r.Handlers)
	//context not match
	option1 := &BackupRetryOption{
		Ratio: 200,
	}
	setupBackupRetry(r, option1)
	err = r.Send()
	assert.NoError(t, err)
	assert.Empty(t, r.Handlers)
}

func TestFalseConditionNotMatched(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "http://localhost:80", nil)
	r := &request.Request{
		HTTPRequest: httpRequest,
	}
	abT := ab.New()
	ctx := ab.NewContext(context.Background(), abT)
	r.SetContext(ctx)
	option := &BackupRetryOption{
		Ratio:                90,
		forceBackupCondition: ab.FALSE,
	}
	setupBackupRetry(r, option)
	r.Send()
	assert.Empty(t, r.Handlers)
}

func TestSendPalceHolder(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "http://localhost:80", nil)
	r := &request.Request{
		HTTPRequest: httpRequest,
	}
	abT := ab.New()
	ctx := ab.NewContext(context.Background(), abT)
	r.SetContext(ctx)
	option := &BackupRetryOption{
		Ratio:             200,
		BackupAction:      "placeholder",
		BackupPlaceholder: "testplaceholder",
	}
	setupBackupRetry(r, option)
	err := r.Send()
	assert.NoError(t, err)
	b, _ := ioutil.ReadAll(r.HTTPResponse.Body)
	assert.Equal(t, "testplaceholder", string(b))
}

func TestSendEcode(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "http://localhost:80", nil)
	r := &request.Request{
		HTTPRequest: httpRequest,
	}
	abT := ab.New()
	ctx := ab.NewContext(context.Background(), abT)
	r.SetContext(ctx)
	option := &BackupRetryOption{
		Ratio:        200,
		BackupAction: "ecode",
		BackupECode:  500,
	}
	setupBackupRetry(r, option)
	err := r.Send()
	assert.NoError(t, err)
	b, _ := ioutil.ReadAll(r.HTTPResponse.Body)
	assert.Equal(t, "{\"code\":500,\"message\":\"500\",\"ttl\":1}", string(b))
}

type testRetryer struct{}

func (t testRetryer) RetryRules(*request.Request) time.Duration {
	return 0
}

func (t testRetryer) ShouldRetry(*request.Request) bool {
	return true
}

func (t testRetryer) MaxRetries() int {
	return 2
}

func TestSendRetryBackup(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "http://api.bilibili.com/x/web-interface/dynamic/regio", nil)
	r := &request.Request{
		HTTPRequest: httpRequest,
		Retryer:     &testRetryer{},
		Operation: &request.Operation{
			Name: "test",
		},
		ClientInfo: metadata.ClientInfo{
			AppID: "test",
		},
	}
	r.HTTPRequest.Body = http.NoBody
	r.SetBufferBody([]byte{})
	abT := ab.New()
	ctx := ab.NewContext(context.Background(), abT)
	r.SetContext(ctx)
	reqURL, _ := url.Parse("discovery://web.interface/path")
	option := &BackupRetryOption{
		Ratio:        200,
		BackupAction: "retry_backup",
		backupURL:    reqURL,
	}
	setupBackupRetry(r, option)
	r.Handlers.Unmarshal.PushBack(func(req *request.Request) {
		req.Error = errors.New("test unmarshal error")
	})
	r.Handlers.AfterRetry.PushBack(func(req *request.Request) {
		if req.RetryCount < 2 {
			req.Error = nil
			req.RetryCount++
			req.Retryable = rootsdk.Bool(true)
		}
	})
	err := r.Send()
	if err == nil {
		t.Log("hope for error but got not")
	}
	assert.Equal(t, r.HTTPRequest.URL.String(), option.backupURL.String())
}

func TestSendDirectlyBackup(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "http://localhost:80", nil)
	r := &request.Request{
		HTTPRequest: httpRequest,
	}
	abT := ab.New()
	ctx := ab.NewContext(context.Background(), abT)
	r.SetContext(ctx)
	reqURL, _ := url.Parse("discovery://web.interface/path")
	option := &BackupRetryOption{
		Ratio:        200,
		BackupAction: "directly_backup",
		backupURL:    reqURL,
	}
	setupBackupRetry(r, option)
	err := r.Send()
	assert.NoError(t, err)
	assert.Equal(t, r.HTTPRequest.URL.String(), option.backupURL.String())
}

type ecodeBody struct {
	Code    int64
	Message string
	TTL     int64
}

func testRequest() *request.Request {
	info := metadata.ClientInfo{
		AppID:    "mock",
		Endpoint: "http://api.bilibili.com/x/web-interface/dynamic/regio",
	}
	s := client.New(sdk.Config{
		MaxRetries: rootsdk.Int(2),
	}, info, request.Handlers{})
	r := s.NewRequest(&request.Operation{Name: "Opreation"}, nil, nil)
	r.SetBufferBody([]byte{})
	r.HTTPRequest.Body = http.NoBody
	abT := ab.New()
	ctx := ab.NewContext(context.Background(), abT)
	r.SetContext(ctx)
	return r
}

func TestFailSendButPlaceHolder(t *testing.T) {
	r := testRequest()
	option := &BackupRetryOption{
		Ratio:             200,
		BackupAction:      "placeholder",
		BackupPlaceholder: "testbody",
	}
	setupBackupRetry(r, option)
	//Regardless of sending success or failure
	//mock send fail
	r.Handlers.Send.PushBack(func(req *request.Request) {
		req.Error = errors.New("test")
	})
	r.Handlers.Unmarshal.PushBack(func(req *request.Request) {
		if r.HTTPResponse.StatusCode != 200 {
			req.Error = errors.New("test falie")
			return
		}
	})
	err := r.Send()
	if err == nil {
		t.Log("except error but got not")
	}
	body, _ := ioutil.ReadAll(r.HTTPResponse.Body)
	assert.Equal(t, "testbody", string(body))
	assert.NotEmpty(t, r.HTTPResponse.Header)
	assert.Equal(t, 200, r.HTTPResponse.StatusCode)
}

func TestFailSendButEcode(t *testing.T) {
	r := testRequest()
	option := &BackupRetryOption{
		Ratio:        200,
		BackupAction: "ecode",
		BackupECode:  500,
	}
	setupBackupRetry(r, option)
	//mock send success
	//it was sent successfully but return 500 ecode
	r.Handlers.Send.PushBack(func(req *request.Request) {
		req.HTTPResponse = &http.Response{
			StatusCode: 200,
		}
	})
	r.Handlers.Unmarshal.PushBack(func(req *request.Request) {
		if r.HTTPResponse.StatusCode == 200 {
			req.Error = errors.New("fail to unmarshal")
		}
	})
	err := r.Send()
	if err == nil {
		t.Log("except error but got not")
	}
	body, _ := ioutil.ReadAll(r.HTTPResponse.Body)
	v := &ecodeBody{}
	err = json.Unmarshal(body, v)
	assert.Equal(t, int64(500), v.Code)
	assert.Equal(t, "500", v.Message)
	assert.Equal(t, int64(1), v.TTL)
}

func TestFailSendButRetryBackup(t *testing.T) {
	r := testRequest()
	reqURL, _ := url.Parse("discovery://web.interface/path")
	option := &BackupRetryOption{
		Ratio:        200,
		BackupAction: "retry_backup",
		backupURL:    reqURL,
	}
	setupBackupRetry(r, option)
	//mock request service failed
	r.Handlers.Send.PushBack(func(req *request.Request) {
		if req.HTTPRequest.URL.String() != reqURL.String() {
			req.Error = errors.New("test")
			return
		}
		req.Error = _errStopAttempt
	})
	//mock request success after retry_backup
	r.Handlers.Unmarshal.PushBack(func(req *request.Request) {
		r.HTTPResponse = &http.Response{
			Status:     http.StatusText(200),
			StatusCode: 200,
		}
	})
	r.Handlers.AfterRetry.PushBack(func(req *request.Request) {
		if req.RetryCount < req.MaxRetries() {
			req.Error = nil
			req.Retryable = rootsdk.Bool(true)
			req.RetryCount++
		}
	})
	err := r.Send()
	assert.NoError(t, err)
	assert.Equal(t, 2, r.RetryCount)
	assert.Equal(t, r.HTTPRequest.URL.String(), reqURL.String())
	assert.Equal(t, 200, r.HTTPResponse.StatusCode)
}

func TestDirectlyBackup(t *testing.T) {
	r := testRequest()
	reqURL, _ := url.Parse("discovery://web.interface/path")
	option := &BackupRetryOption{
		Ratio:        200,
		BackupAction: "directly_backup",
		backupURL:    reqURL,
	}
	setupBackupRetry(r, option)
	//mock request replace url success
	r.Handlers.Send.PushBack(func(req *request.Request) {
		if req.HTTPRequest.URL.String() != reqURL.String() {
			req.Error = errors.New("test")
		}
	})
	r.Handlers.Unmarshal.PushBack(func(req *request.Request) {
		r.HTTPResponse = &http.Response{
			Status:     http.StatusText(200),
			StatusCode: 200,
		}
	})
	err := r.Send()
	assert.NoError(t, err)
	assert.Equal(t, r.HTTPRequest.URL.String(), reqURL.String())
}

func TestRatio(t *testing.T) {
	testCases := []struct {
		expectPercent float64
		setRatio      int64
		delta         float64
		success       bool
	}{
		{
			float64(1),
			100,
			float64(0),
			true,
		},
		{
			float64(0),
			0,
			float64(0),
			true,
		},
		{
			float64(0.01),
			1,
			float64(0.03),
			true,
		},
		{
			float64(0.5),
			50,
			float64(0.03),
			true,
		},
		{
			float64(0.99),
			99,
			float64(0.03),
			true,
		},
		{
			float64(0.8),
			60,
			float64(0.03),
			false,
		},
		{
			float64(0.4),
			70,
			float64(0.03),
			false,
		},
	}

	for _, tcase := range testCases {
		option := &BackupRetryOption{
			Ratio: tcase.setRatio,
		}
		fbp := NewForceBackupPatcher(option)
		r := &request.Request{}
		count := 0
		for i := 0; i < 10000; i++ {
			if fbp.Matched(r) {
				count++
			}
		}
		println(count, option.Ratio)
		if tcase.success {
			if tcase.delta == float64(0) {
				assert.Equal(t, float64(count)/float64(10000), tcase.expectPercent)
				continue
			}
			if tcase.delta != float64(0) {
				assert.InDelta(t, tcase.expectPercent, float64(count)/float64(10000), tcase.delta)
			}
		}
		if !tcase.success {
			absolute := math.Abs(float64(count)/float64(10000) - tcase.expectPercent)
			assert.False(t, absolute <= tcase.delta)
		}
	}
}
