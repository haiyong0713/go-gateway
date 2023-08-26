package request

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go-common/library/log"
	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk"

	"google.golang.org/grpc"
)

const (
	// ErrCodeRequestError is an error preventing the SDK from continuing to
	// process the request.
	ErrCodeRequestError = "RequestError"

	// CanceledErrorCode is the error code that will be returned by an
	// API request that was canceled. Requests given a aws.Context may
	// return this error when canceled.
	CanceledErrorCode = "RequestCanceled"
)

// A Request is the service request to be made.
type Request struct {
	Config   sdk.Config
	Handlers Handlers

	Retryer
	AttemptTime  time.Time
	CompleteTime time.Time
	Time         time.Time
	Operation    *Operation
	Params       interface{}
	Error        error
	Data         interface{}
	RetryCount   int
	Retryable    *bool
	RetryDelay   time.Duration

	// Additional API error codes that should be retried. IsErrorRetryable
	// will consider these codes in addition to its built in cases.
	RetryErrorCodes []string
	RetryECodes     []int

	// Additional API error codes that should be retried with throttle backoff
	// delay. IsErrorThrottle will consider these codes in addition to its
	// built in cases.
	ThrottleErrorCodes []string
	ThrottleECodes     []int

	context context.Context

	built           bool
	handlerPatchers []HandlerPatcher
}

// An Operation is the service API operation to be made.
type Operation struct {
	AppID       string
	Method      string
	Invoker     grpc.UnaryInvoker
	CC          *grpc.ClientConn
	Opts        []grpc.CallOption
	CallContext context.Context
}

// New returns a new Request pointer for the service API operation and
// parameters.
//
// A Retryer should be provided to direct how the request is retried. If
// Retryer is nil, a default no retry value will be used. You can use
// NoOpRetryer in the Client package to disable retry behavior directly.
//
// Params is any value of input parameters to be the request payload.
// Data is pointer value to an object which the request's response
// payload will be deserialized to.
func New(config sdk.Config, handlers Handlers, retryer Retryer, operation *Operation, params, data interface{}) *Request {
	if retryer == nil {
		retryer = noOpRetryer{}
	}

	r := &Request{
		Config:   config,
		Handlers: handlers.Copy(),

		Retryer:   retryer,
		Time:      time.Now(),
		Operation: operation,
		Params:    params,
		Error:     nil,
		Data:      data,
	}
	return r
}

// A Option is a functional option that can augment or modify a request when
// using a WithContext API operation method.
type Option func(*Request)

func WithDebug(debug bool) Option {
	return func(r *Request) {
		r.Config.Debug = debug
	}
}

func WithHandlerPatchers(handlerPatchers ...HandlerPatcher) Option {
	return func(r *Request) {
		r.handlerPatchers = append(r.handlerPatchers, handlerPatchers...)
	}
}

// ApplyOptions will apply each option to the request calling them in the order
// the were provided.
func (r *Request) ApplyOptions(opts ...Option) {
	for _, opt := range opts {
		opt(r)
	}
}

func (r *Request) applyHandlerPatchers() {
	for _, patcher := range r.handlerPatchers {
		if patcher.Matched(r) {
			r.Handlers = patcher.Patch(r.Handlers)
		}
	}
}

// Context will always returns a non-nil context. If Request does not have a
// context aws.BackgroundContext will be returned.
func (r *Request) Context() context.Context {
	if r.context != nil {
		return r.context
	}
	return context.Background()
}

// SetContext adds a Context to the current request that can be used to cancel
// a in-flight request. The Context value must not be nil, or this method will
// panic.
//
// Unlike grpc.Request.WithContext, SetContext does not return a copy of the
// Request. It is not safe to use use a single Request value for multiple
// requests. A new Request should be created for each API operation request.
//
// Go 1.6 and below:
// The grpc.Request's Cancel field will be set to the Done() value of
// the context. This will overwrite the Cancel field's value.
//
// Go 1.7 and above:
// The grpc.Request.WithContext will be used to set the context on the underlying
// grpc.Request. This will create a shallow copy of the grpc.Request. The SDK
// may create sub contexts in the future for nested requests such as retries.
func (r *Request) SetContext(ctx context.Context) {
	if ctx == nil {
		panic("context cannot be nil")
	}
	r.context = ctx
}

// RequestDuration is
func (r *Request) RequestDuration() time.Duration {
	return r.CompleteTime.Sub(r.Time)
}

// WillRetry returns if the request's can be retried.
func (r *Request) WillRetry() (b bool) {
	return r.Error != nil && rootsdk.BoolValue(r.Retryable) && r.RetryCount < r.MaxRetries()
}

func fmtAttemptCount(retryCount, maxRetries int) string {
	return fmt.Sprintf("attempt %v/%v", retryCount, maxRetries)
}

// ParamsFilled returns if the request's parameters have been populated
// and the parameters are valid. False is returned if no parameters are
// provided or invalid.
func (r *Request) ParamsFilled() bool {
	if r.Params == nil {
		return false
	}
	v := reflect.ValueOf(r.Params)
	return v.Kind() == reflect.Ptr && v.Elem().IsValid()
}

// DataFilled returns true if the request's data for response deserialization
// target has been set and is a valid. False is returned if data is not
// set, or is invalid.
func (r *Request) DataFilled() bool {
	if r.Data == nil {
		return false
	}
	v := reflect.ValueOf(r.Data)
	return v.Kind() == reflect.Ptr && v.Elem().IsValid()
}

const (
	notRetrying = "not retrying"
)

func debugLogReqError(r *Request, stage, retryStr string, err error) {
	log.Info(fmt.Sprintf("DEBUG: %s %s/%s failed, %s, error %v",
		stage, r.Operation.AppID, r.Operation.Method, retryStr, err))
}

// Build will build the request's object so it can be signed and sent
// to the service. Build will also validate all the request's parameters.
// Any additional build Handlers set on this request will be run
// in the order they were set.
//
// The request will only be built once. Multiple calls to build will have
// no effect.
//
// If any Validate or Build errors occur the build will stop and the error
// which occurred will be returned.
func (r *Request) Build() error {
	if !r.built {
		r.Handlers.Validate.Run(r)
		if r.Error != nil {
			debugLogReqError(r, "Validate Request", notRetrying, r.Error)
			return r.Error
		}
		r.Handlers.Build.Run(r)
		if r.Error != nil {
			debugLogReqError(r, "Build Request", notRetrying, r.Error)
			return r.Error
		}
		r.built = true
	}

	return r.Error
}

// Sign will sign the request, returning error if errors are encountered.
//
// Sign will build the request prior to signing. All Sign Handlers will
// be executed in the order they were set.
func (r *Request) Sign() error {
	//nolint:errcheck
	r.Build()
	if r.Error != nil {
		debugLogReqError(r, "Build Request", notRetrying, r.Error)
		return r.Error
	}

	r.Handlers.Sign.Run(r)
	return r.Error
}

func shrinkPercent(ctx context.Context, percent float64) (context.Context, func()) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return ctx, func() {}
	}
	timeout := time.Duration(float64(time.Until(deadline)) * percent)
	return context.WithTimeout(ctx, timeout)
}

// Send will send the request, returning error if errors are encountered.
//
// Send will sign the request prior to sending. All Send Handlers will
// be executed in the order they were set.
//
// Canceling a request is non-deterministic. If a request has been canceled,
// then the transport will choose, randomly, one of the state channels during
// reads or getting the connection.
//
// readLoop() and getConn(req *Request, cm connectMethod)
// grpcs://github.com/golang/go/blob/master/src/net/grpc/transport.go
//
// Send will not close the request.Request's body.
func (r *Request) Send() error {
	r.applyHandlerPatchers()

	defer func() {
		// Regardless of success or failure of the request trigger the Complete
		// request handlers.
		r.CompleteTime = time.Now()
		r.Handlers.Complete.Run(r)
	}()

	if err := r.Error; err != nil {
		return err
	}

	for {
		r.Error = nil
		r.AttemptTime = time.Now()

		if err := r.Sign(); err != nil {
			debugLogReqError(r, "Sign Request", notRetrying, err)
			return err
		}

		r.Operation.CallContext = r.context
		if r.MaxRetries() > 0 && r.RetryCount <= 0 {
			perCallCtx, _ := shrinkPercent(r.context, 0.8)
			r.Operation.CallContext = perCallCtx
		}
		if err := r.sendRequest(); err == nil {
			return nil
		}
		r.Handlers.Retry.Run(r)
		r.Handlers.AfterRetry.Run(r)

		if r.Error != nil || !rootsdk.BoolValue(r.Retryable) {
			return r.Error
		}

		if err := r.prepareRetry(); err != nil {
			r.Error = err
			return err
		}
	}
}

func (r *Request) prepareRetry() error {
	log.Info(fmt.Sprintf("DEBUG: Retrying Request %s/%s, attempt %d",
		r.Operation.AppID, r.Operation.Method, r.RetryCount))
	return nil
}

func (r *Request) sendRequest() (sendErr error) {
	defer r.Handlers.CompleteAttempt.Run(r)

	r.Retryable = nil
	r.Handlers.Send.Run(r)
	if r.Error != nil {
		debugLogReqError(r, "Send Request",
			fmtAttemptCount(r.RetryCount, r.MaxRetries()),
			r.Error)
		return r.Error
	}

	r.Handlers.UnmarshalMeta.Run(r)
	r.Handlers.ValidateResponse.Run(r)
	if r.Error != nil {
		r.Handlers.UnmarshalError.Run(r)
		debugLogReqError(r, "Validate Response",
			fmtAttemptCount(r.RetryCount, r.MaxRetries()),
			r.Error)
		return r.Error
	}

	r.Handlers.Unmarshal.Run(r)
	if r.Error != nil {
		debugLogReqError(r, "Unmarshal Response",
			fmtAttemptCount(r.RetryCount, r.MaxRetries()),
			r.Error)
		return r.Error
	}

	return nil
}

// copy will copy a request which will allow for local manipulation of the
// request.
//
//nolint:unused
func (r *Request) copy() *Request {
	req := &Request{}
	*req = *r
	req.Handlers = r.Handlers.Copy()
	return req
}
