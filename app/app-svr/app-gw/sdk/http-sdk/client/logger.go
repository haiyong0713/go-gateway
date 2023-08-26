package client

import (
	"bytes"
	"io"
	"io/ioutil"

	"go-common/library/log"

	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/debug"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"
)

const logReqMsg = `DEBUG: Request %s/%s Details:
---[ REQUEST POST-SIGN ]-----------------------------
%s
-----------------------------------------------------`

const logReqErrMsg = `DEBUG ERROR: Request %s/%s:
---[ REQUEST DUMP ERROR ]-----------------------------
%s
------------------------------------------------------`

// LogHTTPRequestHandler is a SDK request handler to log the HTTP request sent
// to a service. Will include the HTTP request body if the LogLevel of the
// request matches LogDebugWithHTTPBody.
var LogHTTPRequestHandler = request.NamedHandler{
	Name: "appgwsdk.client.LogRequest",
	Fn:   logRequest,
}

func logRequest(r *request.Request) {
	logBody := r.Config.Debug
	bodySeekable := sdk.IsReaderSeekable(r.Body)

	b, err := debug.DumpRequestOut(r.HTTPRequest, logBody)
	if err != nil {
		log.Info(logReqErrMsg,
			r.ClientInfo.AppID, r.Operation.Name, err)
		return
	}

	if logBody {
		if !bodySeekable {
			r.SetReaderBody(sdk.ReadSeekCloser(r.HTTPRequest.Body))
		}
		// Reset the request body because dumpRequest will re-wrap the
		// r.HTTPRequest's Body as a NoOpCloser and will not be reset after
		// read by the HTTP client reader.
		if err := r.Error; err != nil {
			log.Info(logReqErrMsg,
				r.ClientInfo.AppID, r.Operation.Name, err)
			return
		}
	}

	log.Info(logReqMsg,
		r.ClientInfo.AppID, r.Operation.Name, string(b))
}

type teeReaderCloser struct {
	// io.Reader will be a tee reader that is used during logging.
	// This structure will read from a body and write the contents to a logger.
	io.Reader
	// Source is used just to close when we are done reading.
	Source io.ReadCloser
}

func (reader *teeReaderCloser) Close() error {
	return reader.Source.Close()
}

// LogHTTPRequestHeaderHandler is a SDK request handler to log the HTTP request sent
// to a service. Will only log the HTTP request's headers. The request payload
// will not be read.
var LogHTTPRequestHeaderHandler = request.NamedHandler{
	Name: "appgwsdk.client.LogRequestHeader",
	Fn:   logRequestHeader,
}

func logRequestHeader(r *request.Request) {
	b, err := debug.DumpRequestOut(r.HTTPRequest, false)
	if err != nil {
		log.Info(logReqErrMsg,
			r.ClientInfo.AppID, r.Operation.Name, err)
		return
	}

	log.Info(logReqMsg,
		r.ClientInfo.AppID, r.Operation.Name, string(b))
}

const logRespMsg = `DEBUG: Response %s/%s Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`

const logRespErrMsg = `DEBUG ERROR: Response %s/%s:
---[ RESPONSE DUMP ERROR ]-----------------------------
%s
-----------------------------------------------------`

// LogHTTPResponseHandler is a SDK request handler to log the HTTP response
// received from a service. Will include the HTTP response body if the LogLevel
// of the request matches LogDebugWithHTTPBody.
var LogHTTPResponseHandler = request.NamedHandler{
	Name: "appgwsdk.client.LogResponse",
	Fn:   logResponse,
}

func logResponse(r *request.Request) {
	if r.HTTPResponse == nil {
		log.Error(logRespErrMsg,
			r.ClientInfo.AppID, r.Operation.Name, "request's HTTPResponse is nil")
		return
	}

	logBody := r.Config.Debug
	buf := &bytes.Buffer{}
	if logBody {
		r.HTTPResponse.Body = &teeReaderCloser{
			Reader: io.TeeReader(r.HTTPResponse.Body, buf),
			Source: r.HTTPResponse.Body,
		}
	}

	handlerFn := func(req *request.Request) {
		if !logBody {
			return
		}
		b, err := ioutil.ReadAll(buf)
		if err != nil {
			log.Error(logRespErrMsg,
				req.ClientInfo.AppID, req.Operation.Name, err)
			return
		}
		log.Info(logRespMsg,
			req.ClientInfo.AppID, req.Operation.Name, string(b))
	}

	const handlerName = "appgwsdk.client.LogResponse.ResponseBody"

	r.Handlers.Unmarshal.SetBackNamed(request.NamedHandler{
		Name: handlerName, Fn: handlerFn,
	})
	r.Handlers.UnmarshalError.SetBackNamed(request.NamedHandler{
		Name: handlerName, Fn: handlerFn,
	})
}
