package query

import (
	"encoding/json"
	"io/ioutil"

	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"

	"github.com/pkg/errors"
)

// UnmarshalHandler is a named request handler for unmarshaling query protocol requests
var UnmarshalHandler = request.NamedHandler{Name: "appgwsdk.query.Unmarshal", Fn: Unmarshal}

// Unmarshal unmarshals a response for an AWS Query service.
func Unmarshal(r *request.Request) {
	defer r.HTTPResponse.Body.Close()
	if !r.DataFilled() {
		return
	}
	bs, err := ioutil.ReadAll(r.HTTPResponse.Body)
	if err != nil {
		r.Error = errors.WithStack(err)
		return
	}
	if err := json.Unmarshal(bs, r.Data); err != nil {
		r.Error = errors.Errorf("%s: %s: %s", request.ErrCodeSerialization, "failed decoding Query response", err)
		return
	}
}
