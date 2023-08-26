package corehandlers

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/sdkerr"

	"github.com/pkg/errors"
)

// Interface for matching types which also have a Len method.
//
//nolint:deadcode,unused
type lener interface {
	Len() int
}

// BuildContentLengthHandler builds the content length of a request based on the body,
// or will use the HTTPRequest.Header's "Content-Length" if defined. If unable
// to determine request body length and no "Content-Length" was specified it will panic.
//
// The Content-Length will only be added to the request if the length of the body
// is greater than 0. If the body is empty or the current `Content-Length`
// header is <= 0, the header will also be stripped.
var BuildContentLengthHandler = request.NamedHandler{Name: "core.BuildContentLengthHandler", Fn: func(r *request.Request) {
	var length int64

	if slength := r.HTTPRequest.Header.Get("Content-Length"); slength != "" {
		length, _ = strconv.ParseInt(slength, 10, 64)
	} else {
		if r.Body != nil {
			var err error
			length, err = sdk.SeekerLen(r.Body)
			if err != nil {
				r.Error = errors.WithStack(sdkerr.New(request.ErrCodeSerialization, "failed to get request body's length", err))
				return
			}
		}
	}

	if length > 0 {
		r.HTTPRequest.ContentLength = length
		r.HTTPRequest.Header.Set("Content-Length", fmt.Sprintf("%d", length))
	} else {
		r.HTTPRequest.ContentLength = 0
		r.HTTPRequest.Header.Del("Content-Length")
	}
}}

var reStatusCode = regexp.MustCompile(`^(\d{3})`)

// SendHandler is a request handler to send service request using HTTP client.
var SendHandler = request.NamedHandler{
	Name: "core.SendHandler",
	Fn: func(r *request.Request) {
		sender := sendFollowRedirects

		if request.NoBody == r.HTTPRequest.Body {
			// Strip off the request body if the NoBody reader was used as a
			// place holder for a request body. This prevents the SDK from
			// making requests with a request body when it would be invalid
			// to do so.
			//
			// Use a shallow copy of the http.Request to ensure the race condition
			// of transport on Body will not trigger
			reqOrig, reqCopy := r.HTTPRequest, *r.HTTPRequest
			reqCopy.Body = nil
			r.HTTPRequest = &reqCopy
			defer func() {
				r.HTTPRequest = reqOrig
			}()
		}

		if err := sender(r); err != nil {
			handleSendError(r, err)
		}
	},
}

func sendFollowRedirects(r *request.Request) error {
	resp, err := r.Config.Client.DoRequest(r.Context(), r.HTTPRequest, r.ClientInfo.Identifier())
	if err != nil {
		return err
	}
	r.HTTPResponse = resp

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	r.DataBytes = bodyBytes
	resp.Body = sdk.ReadSeekCloser(bytes.NewBuffer(r.DataBytes))

	return nil
}

func handleSendError(r *request.Request, err error) {
	// Prevent leaking if an HTTPResponse was returned. Clean up
	// the body.
	if r.HTTPResponse != nil {
		r.HTTPResponse.Body.Close()
	}
	// Capture the case where url.Error is returned for error processing
	// response. e.g. 301 without location header comes back as string
	// error and r.HTTPResponse is nil. Other URL redirect errors will
	// comeback in a similar method.
	if e, ok := err.(*url.Error); ok && e.Err != nil {
		if s := reStatusCode.FindStringSubmatch(e.Err.Error()); s != nil {
			code, _ := strconv.ParseInt(s[1], 10, 64)
			r.HTTPResponse = &http.Response{
				StatusCode: int(code),
				Status:     http.StatusText(int(code)),
				Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
			}
			return
		}
	}
	if r.HTTPResponse == nil {
		// Add a dummy request response object to ensure the HTTPResponse
		// value is consistent.
		r.HTTPResponse = &http.Response{
			StatusCode: int(0),
			Status:     http.StatusText(int(0)),
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		}
	}

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
	if r.HTTPResponse.StatusCode == 0 || r.HTTPResponse.StatusCode >= 300 {
		// this may be replaced by an UnmarshalError handler
		r.Error = errors.WithStack(sdkerr.New("UnknownError", fmt.Sprintf("StatusCode: %d", r.HTTPResponse.StatusCode), nil))
	}
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
