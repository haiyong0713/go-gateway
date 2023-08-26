package client

import (
	"encoding/json"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/request"
)

const logReqMsg = `DEBUG: Request %s/%s Details:
---[ REQUEST POST-SIGN ]-----------------------------
%s
-----------------------------------------------------`

//nolint:deadcode
// const logReqErrMsg = `DEBUG ERROR: Request %s/%s:
// ---[ REQUEST DUMP ERROR ]-----------------------------
// %s
// ------------------------------------------------------`

// LogGRPCRequestHandler is a SDK request handler to log the GRPC request sent
// to a service. Will include the GRPC request body if the LogLevel of the
// request matches LogDebugWithGRPCBody.
var LogGRPCRequestHandler = request.NamedHandler{
	Name: "appgwsdk.warden.client.LogRequest",
	Fn:   logRequest,
}

func logRequest(r *request.Request) {
	b, _ := json.Marshal(r.Params)
	log.Info(logReqMsg,
		r.Operation.AppID, r.Operation.Method, string(b))
}

// LogGRPCRequestHeaderHandler is a SDK request handler to log the GRPC request sent
// to a service. Will only log the GRPC request's headers. The request payload
// will not be read.
var LogGRPCRequestHeaderHandler = request.NamedHandler{
	Name: "appgwsdk.warden.client.LogRequestHeader",
	Fn:   logRequestHeader,
}

func logRequestHeader(r *request.Request) {
	b, _ := json.Marshal(r.Context)
	log.Info(logReqMsg,
		r.Operation.AppID, r.Operation.Method, string(b))
}

const logRespMsg = `DEBUG: Response %s/%s Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`

const logRespErrMsg = `DEBUG ERROR: Response %s/%s:
---[ RESPONSE DUMP ERROR ]-----------------------------
%s
-----------------------------------------------------`

// LogGRPCResponseHandler is a SDK request handler to log the GRPC response
// received from a service. Will include the GRPC response body if the LogLevel
// of the request matches LogDebugWithGRPCBody.
var LogGRPCResponseHandler = request.NamedHandler{
	Name: "appgwsdk.warden.client.LogResponse",
	Fn:   logResponse,
}

func logResponse(r *request.Request) {
	if r.Data == nil {
		log.Error(logRespErrMsg,
			r.Operation.AppID, r.Operation.Method, "request's GRPCResponse is nil")
		return
	}

	logBody := r.Config.Debug

	handlerFn := func(req *request.Request) {
		if !logBody {
			return
		}
		b, _ := json.Marshal(req.Data)
		log.Info(logRespMsg,
			req.Operation.AppID, req.Operation.Method, string(b))
	}

	const handlerName = "appgwsdk.warden.client.LogResponse.ResponseBody"

	r.Handlers.Unmarshal.SetBackNamed(request.NamedHandler{
		Name: handlerName, Fn: handlerFn,
	})
	r.Handlers.UnmarshalError.SetBackNamed(request.NamedHandler{
		Name: handlerName, Fn: handlerFn,
	})
}
