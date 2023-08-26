package blademaster

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"

	"github.com/pkg/errors"
	gjs "github.com/xeipuuv/gojsonschema"
)

type validatorBuilder func(dsn *url.URL) (Validator, error)

var _allValidatorBuilder = map[string]validatorBuilder{}

// Validator is
type Validator interface {
	String() string
	Validate(r *request.Request, body []byte) error
}

type jsonschema struct {
	schema *gjs.Schema
}

func buildJSONSchemaValidator(dsn *url.URL) (Validator, error) {
	loader := dsn.Query().Get("loader")
	ref := dsn.Query().Get("reference")
	schemaLoader, err := NewJSONLoader(loader, ref)
	if err != nil {
		return nil, err
	}
	schema, err := gjs.NewSchema(schemaLoader)
	if err != nil {
		return nil, err
	}
	return jsonschema{schema: schema}, nil
}

func (js jsonschema) String() string {
	return fmt.Sprintf("%+v", js.schema)
}

//nolint:unparam
func removeJSONP(in []byte, callback string) ([]byte, bool) {
	if callback == "" {
		return in, false
	}
	if len(in) <= len(callback) {
		return in, false
	}
	prefix := in[:len(callback)]
	if string(prefix) != callback {
		return in, false
	}
	if in[len(in)-1] != byte(')') {
		return in, false
	}
	return in[len(callback)+1 : len(in)-1], true
}

func (js jsonschema) Validate(r *request.Request, body []byte) error {
	callback := r.HTTPRequest.URL.Query().Get("callback")
	pureBody, _ := removeJSONP(body, callback)
	bodyLoader := gjs.NewStringLoader(string(pureBody))
	result, err := js.schema.Validate(bodyLoader)
	if err != nil {
		return errors.WithStack(err)
	}
	if result.Valid() {
		return nil
	}
	errorString := []string{}
	for _, err := range result.Errors() {
		errorString = append(errorString, fmt.Sprintf("  - %v", err))
	}
	return errors.Errorf("invalid body: %s:\n%s", string(pureBody), strings.Join(errorString, "\n"))
}

func AddValidator(name string, builder validatorBuilder) {
	_allValidatorBuilder[name] = builder
}

func init() {
	AddValidator("jsonschema", buildJSONSchemaValidator)
}

func BuildValidator(dsn string) (Validator, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	builder, ok := _allValidatorBuilder[u.Scheme]
	if !ok {
		return nil, errors.Errorf("no such validator: %s", u.Scheme)
	}
	return builder(u)
}

func wrapValidator(in Validator) request.NamedHandler {
	handlerFn := func(r *request.Request) {
		if r.Error != nil {
			return
		}
		// not a critical error status code
		if r.HTTPResponse.StatusCode >= 300 && r.HTTPResponse.StatusCode < 500 {
			return
		}
		r.Error = in.Validate(r, r.DataBytes)
	}
	return request.NamedHandler{Name: "appgwsdk.blademaster.ResponseBodyValidator", Fn: handlerFn}
}

func ecodeFromJSON(in []byte) (int, bool) {
	body := &struct {
		Code int `json:"code"`
	}{}
	if err := json.Unmarshal(in, body); err != nil {
		return 0, false
	}
	return body.Code, true
}

var UnmarshalECodeHandler = request.NamedHandler{
	Name: "appgwsdk.blademaster.UnmarshalECodeHandler",
	Fn: func(r *request.Request) {
		if r.Error != nil {
			return
		}
		ecodeString := r.HTTPResponse.Header.Get("Bili-Status-Code")
		if ecodeString != "" {
			ecodeInt, err := strconv.ParseInt(ecodeString, 10, 64)
			if err != nil {
				return
			}
			if ecodeInt == 0 {
				return
			}
			r.Error = errors.WithStack(ecode.Int(int(ecodeInt)))
			return
		}
		callback := r.HTTPRequest.URL.Query().Get("callback")
		pureBody, _ := removeJSONP(r.DataBytes, callback)
		ecodeInt, ok := ecodeFromJSON(pureBody)
		if ok && ecodeInt != 0 {
			r.Error = errors.WithStack(ecode.Int(int(ecodeInt)))
		}
	},
}
