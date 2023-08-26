package defaults

import (
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/corehandlers"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/request"
)

// Handlers returns the default request handlers.
//
// Generally you shouldn't need to use this method directly, but
// is available if you need to reset the request handlers of an
// existing service client or session.
func Handlers() request.Handlers {
	var handlers request.Handlers

	handlers.Validate.AfterEachFn = request.HandlerListStopOnError
	handlers.Build.AfterEachFn = request.HandlerListStopOnError
	handlers.Send.PushBackNamed(corehandlers.SendHandler)
	handlers.AfterRetry.PushBackNamed(corehandlers.AfterRetryHandler)
	handlers.ValidateResponse.PushBackNamed(corehandlers.ValidateResponseHandler)

	return handlers
}
