package test

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	bm "go-common/library/net/http/blademaster"
	xtime "go-common/library/time"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	bmsdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"
	def "go-gateway/app/app-svr/app-gw/sdk/http-sdk/defaults"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"

	"github.com/stretchr/testify/assert"
)

const (
	StatusOk = 200
	ep       = "http://www.baidu.com:80"
)

func NewBaiduClient() *client.Client {
	bmCli := bm.NewClient(&bm.ClientConfig{
		App: &bm.App{
			Key:    "53e2fa226f5ad348",
			Secret: "3cf6bd1b0ff671021da5f424fea4b04a",
		},
		Timeout: xtime.Duration(time.Second),
	})
	cfg := sdk.Config{
		Client: bmsdk.WrapClient(bmCli),
		Debug:  true,
	}
	info := metadata.ClientInfo{
		AppID:    "dd",
		Endpoint: ep,
	}
	handlers := def.Handlers()
	return client.New(cfg, info, handlers)
}

func NewBaiduRequest(name string, httpMethod string, httpPath string) *request.Request {
	BaiduClient := NewBaiduClient()
	operation := &request.Operation{
		Name:       name,
		HTTPMethod: httpMethod,
		HTTPPath:   httpPath,
	}
	return BaiduClient.NewRequest(operation, nil, nil)
}
func TestSendRequestToBaidu(t *testing.T) {
	req := NewBaiduRequest("baidu", "GET", "/")
	err := req.Build()
	assert.NoError(t, err)
	req.SetContext(context.TODO())
	err = req.Send()
	assert.NoError(t, err)
	rep := req.HTTPResponse
	code := rep.StatusCode
	assert.Equal(t, StatusOk, code)
	_, err = ioutil.ReadAll(rep.Body)
	assert.NoError(t, err)
}

func TestSanitizeHostForHeader(t *testing.T) {
	req := NewBaiduRequest("baidu", "GET", "/")
	t.Log(req.HTTPRequest.URL.Host, req.HTTPRequest.URL.Scheme)
	request.SanitizeHostForHeader(req.HTTPRequest)
	reqSanitizeHost := req.HTTPRequest.Host
	assert.Equal(t, "www.baidu.com", reqSanitizeHost)
}
