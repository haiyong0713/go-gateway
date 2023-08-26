package client

import (
	"go-common/library/log"
	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/request"
)

// A Client implements the base client request and response handling
// used by all service clients.
type Client struct {
	request.Retryer

	config sdk.Config

	Handlers request.Handlers
}

// New will return a pointer to a new initialized service client.
func New(config sdk.Config, handlers request.Handlers, options ...func(*Client)) *Client {
	svc := &Client{
		config:   config,
		Handlers: handlers.Copy(),
	}

	switch retryer, ok := config.Retryer.(request.Retryer); {
	case ok:
		svc.Retryer = retryer
	case config.Retryer != nil:
		log.Warn("%T does not implement request.Retryer; using DefaultRetryer instead", config.Retryer)
		fallthrough
	default:
		maxRetries := rootsdk.Int64Value(config.MaxRetries)
		if config.MaxRetries == nil || maxRetries == sdk.UseServiceDefaultRetries {
			maxRetries = DefaultRetryerMaxNumRetries
		}
		svc.Retryer = DefaultRetryer{NumMaxRetries: int(maxRetries)}
	}

	svc.AddDebugHandlers()

	for _, option := range options {
		option(svc)
	}

	return svc
}

// NewRequest returns a new Request pointer for the service API
// operation and parameters.
func (c *Client) NewRequest(operation *request.Operation, params, data interface{}) *request.Request {
	return request.New(c.config, c.Handlers, c.Retryer, operation, params, data)
}

// AddDebugHandlers injects debug logging handlers into the service to log request
// debug information.
func (c *Client) AddDebugHandlers() {
	if !c.config.Debug {
		return
	}

	c.Handlers.Send.PushFrontNamed(LogGRPCRequestHandler)
	c.Handlers.Send.PushBackNamed(LogGRPCResponseHandler)
}
