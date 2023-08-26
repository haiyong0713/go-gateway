package request

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/sdkerr"

	"github.com/stretchr/testify/assert"
)

type mockTempError bool

func (e mockTempError) Error() string {
	return fmt.Sprintf("mock temporary error: %t", e.Temporary())
}
func (e mockTempError) Temporary() bool {
	return bool(e)
}

func TestIsErrorRetryable(t *testing.T) {
	cases := []struct {
		Err       error
		Retryable bool
	}{
		{
			Err:       &url.Error{Err: errors.New("connection refused..a")},
			Retryable: true,
		},
		{
			Err:       &url.Error{Err: errors.New("aaaaa")},
			Retryable: true,
		},
		{
			Err:       sdkerr.New(ErrCodeSerialization, "temporary error", mockTempError(true)),
			Retryable: true,
		},
		{
			Err:       sdkerr.New(ErrCodeSerialization, "temporary error", mockTempError(false)),
			Retryable: false,
		},
		{
			Err:       sdkerr.New(ErrCodeSerialization, "some error", errors.New("blah")),
			Retryable: true,
		},
		{
			Err:       sdkerr.New("connection refused", "connection refused", nil),
			Retryable: false,
		},
		{
			Err:       sdkerr.New(ErrCodeRequestError, "some error", nil),
			Retryable: true,
		},
		{
			Err:       nil,
			Retryable: false,
		},
	}
	for _, c := range cases {
		retryable := IsErrorRetryable(c.Err)
		assert.Equal(t, c.Retryable, retryable)
	}
}

func TestRequestThrottling(t *testing.T) {
	req := Request{}
	cases := []struct {
		ecode string
	}{
		{
			ecode: "ProvisionedThroughputExceededException",
		},
		{
			ecode: "ThrottledException",
		},
		{
			ecode: "Throttling",
		},
		{
			ecode: "ThrottlingException",
		},
		{
			ecode: "RequestLimitExceeded",
		},
		{
			ecode: "RequestThrottled",
		},
		{
			ecode: "TooManyRequestsException",
		},
		{
			ecode: "PriorRequestNotComplete",
		},
		{
			ecode: "TransactionInProgressException",
		},
		{
			ecode: "EC2ThrottledException",
		},
	}
	for _, c := range cases {
		req.Error = sdkerr.New(c.ecode, "", nil)
		assert.Equal(t, true, req.IsErrorThrottle())
	}
}

func TestRequest_NilRetryer(t *testing.T) {
	clientInfo := metadata.ClientInfo{Endpoint: "http://mock.bilibili.com"}
	req := New(sdk.Config{}, clientInfo, Handlers{}, nil, &Operation{}, nil, nil)
	assert.Empty(t, req.Retryer)
	assert.Equal(t, 0, req.MaxRetries())
}

func TesturlError(t *testing.T) {
	urlErr := &url.Error{
		Err: errors.New("connection refused"),
	}
	retr := IsErrorRetryable(urlErr)
	assert.Equal(t, retr, true)
}
