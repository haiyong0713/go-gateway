package request_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestResetBody_WithEmptyBody(t *testing.T) {
	r := request.Request{
		HTTPRequest: &http.Request{},
	}
	reader := strings.NewReader("")
	r.Body = reader
	r.ResetBody()
	assert.Equal(t, r.HTTPRequest.Body, http.NoBody)
}

func TestRequest_FollowPUTRedirects(t *testing.T) {
	const bodySize = 1024
	redirectHit := 0
	endpointHit := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect-me":
			u := *r.URL
			u.Path = "/endpoint"
			w.Header().Set("Location", u.String())
			w.WriteHeader(307)
			redirectHit++
		case "/endpoint":
			b := bytes.Buffer{}
			io.Copy(&b, r.Body)
			r.Body.Close()
			assert.Equal(t, bodySize, b.Len())
			endpointHit++
		default:
			t.Fatalf("unexpected endpoint used, %q", r.URL.String())
		}

	}))
	defer server.Close()
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
		AppID:    "d",
		Endpoint: server.URL,
	}
	handlers := def.Handlers()
	svc := client.New(cfg, info, handlers)
	operation := &request.Operation{
		Name:       "d",
		HTTPMethod: "GET",
		HTTPPath:   "/redirect-me",
	}
	req := svc.NewRequest(operation, nil, nil)
	req.SetReaderBody(bytes.NewReader(make([]byte, bodySize)))
	err := req.Send()
	assert.NoError(t, err)
	assert.Equal(t, 1, redirectHit)
	assert.Equal(t, 1, endpointHit)
}
