package client

import (
	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"

	"go-common/library/log"
)

// A Client implements the base client request and response handling
// used by all service clients.
type Client struct {
	request.Retryer
	metadata.ClientInfo

	Config   sdk.Config
	Handlers request.Handlers
}

// New will return a pointer to a new initialized service client.
func New(cfg sdk.Config, info metadata.ClientInfo, handlers request.Handlers, options ...func(*Client)) *Client {
	svc := &Client{
		Config:     cfg,
		ClientInfo: info,
		Handlers:   handlers.Copy(),
	}

	switch retryer, ok := cfg.Retryer.(request.Retryer); {
	case ok:
		svc.Retryer = retryer
	case cfg.Retryer != nil:
		log.Warn("%T does not implement request.Retryer; using DefaultRetryer instead", cfg.Retryer)
		fallthrough
	default:
		maxRetries := rootsdk.IntValue(cfg.MaxRetries)
		if cfg.MaxRetries == nil || maxRetries == sdk.UseServiceDefaultRetries {
			maxRetries = DefaultRetryerMaxNumRetries
		}
		svc.Retryer = DefaultRetryer{NumMaxRetries: maxRetries}
	}

	svc.AddDebugHandlers()

	for _, option := range options {
		option(svc)
	}

	return svc
}

// NewRequest returns a new Request pointer for the service API
// operation and parameters.
func (c *Client) NewRequest(operation *request.Operation, params interface{}, data interface{}) *request.Request {
	return request.New(c.Config, c.ClientInfo, c.Handlers, c.Retryer, operation, params, data)
}

// AddDebugHandlers injects debug logging handlers into the service to log request
// debug information.
func (c *Client) AddDebugHandlers() {
	if !c.Config.Debug {
		return
	}

	c.Handlers.Send.PushFrontNamed(LogHTTPRequestHandler)
	c.Handlers.Send.PushBackNamed(LogHTTPResponseHandler)
}
