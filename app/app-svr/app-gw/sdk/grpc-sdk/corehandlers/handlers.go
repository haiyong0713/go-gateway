package corehandlers

import (
	"context"
	"time"

	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/request"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/sdkerr"

	"github.com/pkg/errors"
)

// Interface for matching types which also have a Len method.
//
//nolint:deadcode,unused
type lener interface {
	Len() int
}

// SendHandler is a request handler to send service request using HTTP client.
var SendHandler = request.NamedHandler{
	Name: "core.SendHandler",
	Fn: func(r *request.Request) {
		if err := doInvoke(r); err != nil {
			handleSendError(r, err)
		}
	},
}

func doInvoke(r *request.Request) error {
	return r.Operation.Invoker(r.Operation.CallContext, r.Operation.Method, r.Params, r.Data, r.Operation.CC, r.Operation.Opts...)
}

func handleSendError(r *request.Request, err error) {
	// Catch all request errors, and let the default retrier determine
	// if the error is retryable.
	r.Error = errors.WithStack(sdkerr.New(request.ErrCodeRequestError, "send request failed", err))

	// Override the error with a context canceled error, if that was canceled.
	ctx := r.Context()
	select {
	case <-ctx.Done():
		r.Error = errors.WithStack(sdkerr.New(request.CanceledErrorCode,
			"request context canceled", ctx.Err()))
		r.Retryable = rootsdk.Bool(false)
	default:
	}
}

// ValidateResponseHandler is a request handler to validate service response.
var ValidateResponseHandler = request.NamedHandler{Name: "core.ValidateResponseHandler", Fn: func(r *request.Request) {
}}

// AfterRetryHandler performs final checks to determine if the request should
// be retried and how long to delay.
var AfterRetryHandler = request.NamedHandler{
	Name: "core.AfterRetryHandler",
	Fn: func(r *request.Request) {
		// If one of the other handlers already set the retry state
		// we don't want to override it based on the service's state
		if r.Retryable == nil {
			r.Retryable = rootsdk.Bool(r.ShouldRetry(r))
		}

		if r.WillRetry() {
			r.RetryDelay = r.RetryRules(r)

			if err := SleepWithContext(r.Context(), r.RetryDelay); err != nil {
				r.Error = errors.WithStack(sdkerr.New(request.CanceledErrorCode,
					"request context canceled", err))
				r.Retryable = rootsdk.Bool(false)
				return
			}

			r.RetryCount++
			r.Error = nil
		}
	},
}

// SleepWithContext will wait for the timer duration to expire, or the context
// is canceled. Which ever happens first. If the context is canceled the Context's
// error will be returned.
//
// Expects Context to always return a non-nil error if the Done channel is closed.
func SleepWithContext(ctx context.Context, dur time.Duration) error {
	t := time.NewTimer(dur)
	defer t.Stop()

	select {
	case <-t.C:
		break
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
