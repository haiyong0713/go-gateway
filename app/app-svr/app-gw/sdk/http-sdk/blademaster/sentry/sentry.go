package sentry

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	bm "go-common/library/net/http/blademaster"

	"github.com/getsentry/sentry-go"
)

const valuesKey = "sentry"

type handler struct {
	repanic         bool
	waitForDelivery bool
	timeout         time.Duration
}

type Options struct {
	// Repanic configures whether Sentry should repanic after recovery, in most cases it should be set to true,
	// as gin.Default includes it's own Recovery middleware what handles http responses.
	Repanic bool
	// WaitForDelivery configures whether you want to block the request before moving forward with the response.
	// Because Gin's default `Recovery` handler doesn't restart the application,
	// it's safe to either skip this option or set it to `false`.
	WaitForDelivery bool
	// Timeout for the event delivery requests.
	Timeout time.Duration
}

// New returns a function using default config.
func Default() bm.HandlerFunc {
	return New(Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         time.Second * 2,
	})
}

// New returns a function that satisfies gin.HandlerFunc interface
// It can be used with Use() methods.
func New(options Options) bm.HandlerFunc {
	handler := handler{
		repanic:         false,
		timeout:         time.Second * 2,
		waitForDelivery: false,
	}

	if options.Repanic {
		handler.repanic = true
	}

	if options.Timeout != 0 {
		handler.timeout = options.Timeout
	}

	if options.WaitForDelivery {
		handler.waitForDelivery = true
	}

	return handler.handle
}

func (h *handler) handle(ctx *bm.Context) {
	hub := sentry.CurrentHub().Clone()
	hub.Scope().SetRequest(ctx.Request)
	ctx.Set(valuesKey, hub)
	defer h.recoverWithSentry(hub, ctx.Request)
	ctx.Next()
}

func (h *handler) recoverWithSentry(hub *sentry.Hub, r *http.Request) {
	if err := recover(); err != nil {
		if !isBrokenPipeError(err) {
			eventID := hub.RecoverWithContext(
				context.WithValue(r.Context(), sentry.RequestContextKey, r),
				err,
			)
			if eventID != nil && h.waitForDelivery {
				hub.Flush(h.timeout)
			}
		}
		if h.repanic {
			panic(err)
		}
	}
}

// Check for a broken connection, as this is what Gin does already
func isBrokenPipeError(err interface{}) bool {
	if netErr, ok := err.(*net.OpError); ok {
		if sysErr, ok := netErr.Err.(*os.SyscallError); ok {
			if strings.Contains(strings.ToLower(sysErr.Error()), "broken pipe") ||
				strings.Contains(strings.ToLower(sysErr.Error()), "connection reset by peer") {
				return true
			}
		}
	}
	return false
}

// GetHubFromContext retrieves attached *sentry.Hub instance from gin.Context.
func GetHubFromContext(ctx *bm.Context) *sentry.Hub {
	if hub, ok := ctx.Get(valuesKey); ok {
		if hub, ok := hub.(*sentry.Hub); ok {
			return hub
		}
	}
	return nil
}
