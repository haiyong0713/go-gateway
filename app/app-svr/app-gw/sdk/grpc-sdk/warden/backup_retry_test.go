package warden

import (
	"context"
	"encoding/json"
	"testing"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/app-gw/management/api"
	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/client"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/request"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster/ab"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func testRequest() *request.Request {
	s := client.New(sdk.Config{
		MaxRetries: rootsdk.Int64(2),
	}, request.Handlers{})
	operation := &request.Operation{
		AppID:  appID,
		Method: "testOperation",
	}
	r := s.NewRequest(operation, &api.Project{}, &api.Project{})
	abT := ab.New()
	ctx := ab.NewContext(context.Background(), abT)
	r.SetContext(ctx)
	return r
}
func TestPlaceHolder(t *testing.T) {
	r := testRequest()
	option := &BackupRetryOption{
		Ratio:             100,
		BackupAction:      "placeholder",
		BackupPlaceholder: `{"project_name":"p","node":"n"}`,
	}
	setupBackupRetry(r, option)
	r.Handlers.Send.PushBack(func(req *request.Request) {
		req.Error = errors.New("test")
	})
	r.Handlers.Unmarshal.PushBack(func(req *request.Request) {
		if req.Data == nil {
			req.Error = errors.New("test failed")
			return
		}
	})
	err := r.Send()
	assert.NoError(t, err)
	d, err := json.Marshal(r.Data)
	assert.NoError(t, err)
	assert.Equal(t, string(d), `{"project_name":"p","node":"n"}`)
}

func TestEcode(t *testing.T) {
	r := testRequest()
	option := &BackupRetryOption{
		Ratio:        100,
		BackupAction: "ecode",
		BackupECode:  500,
	}
	setupBackupRetry(r, option)
	r.Handlers.Send.PushBack(func(req *request.Request) {
		req.Error = errors.New("test")
	})
	r.Handlers.Unmarshal.PushBack(func(req *request.Request) {
		if req.Data == nil {
			req.Error = errors.New("test failed")
			return
		}
	})
	err := r.Send()
	assert.Equal(t, ecode.Int(500), err)
}
