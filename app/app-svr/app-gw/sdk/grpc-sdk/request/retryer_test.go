package request

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	sdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/sdkerr"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockTempError bool

func (e mockTempError) Error() string {
	return fmt.Sprintf("mock temporary error: %t", e.Temporary())
}
func (e mockTempError) Temporary() bool {
	return bool(e)
}

type mockEcodError struct {
	ecode int
}

func (m mockEcodError) Code() int {
	return m.ecode
}

func (m mockEcodError) Error() string {
	return "mockEcodError"
}

func TestIsErrorRetryable(t *testing.T) {
	req := &Request{}
	cases := []struct {
		Err       error
		Retryable bool
	}{
		{
			Err:       &mockEcodError{ecode: -500},
			Retryable: true,
		},
		{
			Err:       &url.Error{Err: errors.New("connection refused..a")},
			Retryable: true,
		},
		{
			Err:       &url.Error{Err: errors.New("aaaaa")},
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
		req.Error = c.Err
		assert.Equal(t, c.Retryable, req.IsErrorRetryable())
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
	//clientInfo := metadata.ClientInfo{AppID: "bilibili"}
	req := New(sdk.Config{}, Handlers{}, nil, &Operation{}, nil, nil)
	assert.Empty(t, req.Retryer)
	assert.Equal(t, 0, req.MaxRetries())
}

type TestRetryer struct {
	SR bool
	MR int
}

func (t TestRetryer) RetryRules(r *Request) time.Duration {
	return 0
}

func (t TestRetryer) ShouldRetry(r *Request) bool {
	return t.SR
}

func (t TestRetryer) MaxRetries() int {
	return t.MR
}

func TestWithRetryer(t *testing.T) {
	cfg := &sdk.Config{}
	tr := &TestRetryer{
		SR: true,
		MR: 2,
	}
	WithRetryer(cfg, tr)
	assert.Equal(t, cfg.Retryer, tr)
	WithRetryer(cfg, nil)
	assert.Equal(t, cfg.Retryer, noOpRetryer{})
}
