package test

import (
	"context"
	"net/url"
	"testing"
	"time"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/netutil/breaker"
	xtime "go-common/library/time"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	bmsdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"
	def "go-gateway/app/app-svr/app-gw/sdk/http-sdk/defaults"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/protocol/query"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"
)

func NewAccountClient() *client.Client {
	bmCli := bm.NewClient(&bm.ClientConfig{
		App: &bm.App{
			Key:    "53e2fa226f5ad348",
			Secret: "3cf6bd1b0ff671021da5f424fea4b04a",
		},
		Dial:      xtime.Duration(time.Second),
		Timeout:   xtime.Duration(time.Second),
		KeepAlive: xtime.Duration(time.Second),
		Breaker: &breaker.Config{
			Window:  10 * xtime.Duration(time.Second),
			Sleep:   50 * xtime.Duration(time.Millisecond),
			Bucket:  10,
			Ratio:   0.5,
			Request: 100,
		},
	})

	cfg := sdk.Config{
		Client: bmsdk.WrapClient(bmCli),
		Debug:  true,
		Key:    "53e2fa226f5ad348",
		Secret: "3cf6bd1b0ff671021da5f424fea4b04a",
	}
	info := metadata.ClientInfo{
		AppID:    "main.account.account-service-group1",
		Endpoint: "http://uat-api.bilibili.co",
	}
	handlers := def.Handlers()
	handlers.Build.PushBackNamed(query.BuildHandler)
	handlers.Sign.PushBackNamed(query.SignRequestHandler)
	handlers.Unmarshal.PushBackNamed(query.UnmarshalHandler)
	return client.New(cfg, info, handlers)
}

func TestRealAPICall(t *testing.T) {
	account := NewAccountClient()

	operation := &request.Operation{
		Name:       "AccountInfoV3",
		HTTPMethod: "GET",
		HTTPPath:   "/x/internal/v3/account/info",
	}
	resp := &struct {
		Code    int64                  `json:"code"`
		Message string                 `json:"message"`
		TTL     int64                  `json:"ttl"`
		Data    map[string]interface{} `json:"data"`
	}{}
	params := &struct {
		Mid int64 `queryName:"mid"`
	}{
		Mid: 2231365,
	}
	req := account.NewRequest(operation, params, resp)
	req.SetContext(context.TODO())
	req.Send()

	resp2 := &struct {
		Code    int64                  `json:"code"`
		Message string                 `json:"message"`
		TTL     int64                  `json:"ttl"`
		Data    map[string]interface{} `json:"data"`
	}{}
	params2 := url.Values{}
	params2.Set("mid", "2231365")
	req2 := account.NewRequest(operation, params2, resp2)
	req2.SetContext(context.TODO())
	req2.Send()

	t.Logf("%+v\n%+v", req.Data, req2.Data)
}
