package blademaster

import (
	"fmt"
	"testing"

	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"

	"github.com/stretchr/testify/assert"
)

func TestResponseNil(t *testing.T) {
	req := request.New(
		sdk.Config{},
		metadata.ClientInfo{
			AppID:    "test",
			Endpoint: "http://localhost:8000",
		},
		testHandlers(),
		nil,
		&request.Operation{
			Name:       "Operation",
			HTTPMethod: "GET",
			HTTPPath:   "/",
		},
		struct{}{}, nil,
	)
	req.Error = fmt.Errorf("error")
	req.Retryable = rootsdk.Bool(true)
	err := req.Send()
	assert.NotEqual(t, nil, err)
	assert.Empty(t, req.HTTPResponse)
}

func testHandlers() request.Handlers {
	var handlers request.Handlers

	handlers.CompleteAttempt.PushBackNamed(ReportUpstreamAttemptHandler)
	handlers.Complete.PushBackNamed(ReportUpstreamAttemptHandler)

	return handlers
}
