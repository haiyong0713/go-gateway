package request_test

import (
	"context"
	"errors"
	"testing"
	"time"

	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/request"

	"github.com/stretchr/testify/assert"
)

func TestSetContext(t *testing.T) {
	r := &request.Request{}
	r.SetContext(context.WithValue(context.Background(), "k", "GO"))
	assert.Equal(t, r.Context().Value("k"), "GO")
}

type testData struct {
	name string
}

func TestParamsFilled(t *testing.T) {
	r := &request.Request{}
	r.Params = &testData{
		name: "test",
	}
	b := r.ParamsFilled()
	assert.Equal(t, true, b)
	r.Params = "bilibili"
	b = r.ParamsFilled()
	assert.Equal(t, false, b)
}

func TestDataFilled(t *testing.T) {
	r := &request.Request{}
	r.Data = &testData{
		name: "test",
	}
	b := r.DataFilled()
	assert.Equal(t, true, b)
	r.Data = "bilibili"
	b = r.DataFilled()
	assert.Equal(t, false, b)
}

func TestBuildError(t *testing.T) {
	r := &request.Request{
		Operation: &request.Operation{Method: "test"},
	}
	r.Handlers.Validate.PushBack(func(req *request.Request) {
		req.Error = errors.New("test validate error")
	})
	err := r.Send()
	if err == nil {
		t.Fatalf("expect error, but got nil")
	}
	assert.NotEmpty(t, r.Error)

	r.Error = nil
	r.Handlers.Validate.Clear()
	r.Handlers.Build.PushBack(func(req *request.Request) {
		req.Error = errors.New("test build error")
	})
	err = r.Send()
	if err == nil {
		t.Fatalf("expect error, but got nil")
	}
	assert.NotEmpty(t, r.Error)
}

func TestSendRequestError(t *testing.T) {
	r := &request.Request{
		Operation: &request.Operation{Method: "test"},
		Retryer: request.TestRetryer{
			SR: true,
			MR: 1,
		},
	}
	r.Handlers.Send.PushBack(func(req *request.Request) {
		req.Error = errors.New("test Send error")
	})
	err := r.Send()
	if err == nil {
		t.Fatalf("expect error, but got nil")
	}
	assert.NotEmpty(t, r.Error)

	r.Error = nil
	r.Handlers.Send.Clear()
	r.Handlers.ValidateResponse.PushBack(func(req *request.Request) {
		req.Error = errors.New("test ValidateResponse error")
	})
	err = r.Send()
	if err == nil {
		t.Fatalf("expect error, but got nil")
	}
	assert.NotEmpty(t, r.Error)

	r.Error = nil
	r.Handlers.ValidateResponse.Clear()
	r.Handlers.Unmarshal.PushBack(func(req *request.Request) {
		req.Error = errors.New("test Unmarshal error")
	})
	err = r.Send()
	if err == nil {
		t.Fatalf("expect error, but got nil")
	}
	assert.NotEmpty(t, r.Error)
}

func TestPrepareRetry(t *testing.T) {
	r := &request.Request{
		Operation: &request.Operation{Method: "test"},
		Retryer: request.TestRetryer{
			SR: true,
			MR: 10,
		},
		RetryCount: 0,
	}
	r.Handlers.Send.PushBack(func(req *request.Request) {
		req.Error = errors.New("test Send error")
	})
	r.Handlers.AfterRetry.PushBack(func(req *request.Request) {
		req.Retryable = rootsdk.Bool(true)
		req.Error = nil
		req.RetryCount++
		if r.RetryCount == 2 {
			req.Error = errors.New("test AfterRetry error")
		}
	})
	err := r.Send()
	if err == nil {
		t.Fatalf("expect error, but got nil")
	}
	assert.NotEmpty(t, r.Error)
}

func TestWithDebug(t *testing.T) {
	r := &request.Request{}
	opt := request.WithDebug(true)
	r.ApplyOptions(opt)
	assert.Equal(t, true, r.Config.Debug)
	opt = request.WithDebug(false)
	r.ApplyOptions(opt)
	assert.Equal(t, false, r.Config.Debug)
}

type testHandlerPatcher struct {
	name    string
	matched bool
}

func (d *testHandlerPatcher) Name() string {
	return d.name
}

func (d *testHandlerPatcher) Patch(in request.Handlers) request.Handlers {
	in.Send.PushBack(func(r *request.Request) {})
	return in
}

func (d *testHandlerPatcher) Matched(r *request.Request) bool {
	return d.matched
}

func TestWithHandlerPatcher(t *testing.T) {
	r := &request.Request{}
	handlerPatchers := &testHandlerPatcher{
		name:    "d",
		matched: true,
	}
	opt := request.WithHandlerPatchers(handlerPatchers)
	r.ApplyOptions(opt)
	err := r.Send()
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Handlers.Send)
}

func TestRequestDuration(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(1)
	r := &request.Request{
		Time:         t1,
		CompleteTime: t2,
	}
	d := r.RequestDuration()
	assert.Equal(t, time.Duration(1), d)
}

func TestWillRetry(t *testing.T) {
	r := &request.Request{
		Operation: &request.Operation{Method: "test"},
		Retryer: request.TestRetryer{
			SR: true,
			MR: 1,
		},
		RetryCount: 0,
	}
	r.Handlers.Send.PushBack(func(req *request.Request) {
		req.Error = errors.New("test Send error")
	})
	r.Handlers.AfterRetry.PushBack(func(req *request.Request) {
		if req.RetryCount < 1 {
			req.Retryable = rootsdk.Bool(true)
		}
		if req.WillRetry() {
			req.RetryCount++
		}
		req.Error = nil
	})
	err := r.Send()
	if err == nil {
		t.Log("want error but got not")
	}
	assert.Equal(t, 1, r.RetryCount)
}
