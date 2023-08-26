// Package query provides serialization of bilibili query requests, and responses.
package query

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/protocol/query/queryutil"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"

	"github.com/pkg/errors"
)

const (
// _appKey = "appkey"
// _appSecret = "appsecret"
// _ts = "ts"
//
//nolint:deadcode
//nolint:deadcode
//nolint:deadcode
)

// BuildHandler is a named request handler for building query protocol requests
var BuildHandler = request.NamedHandler{Name: "appgwsdk.query.Build", Fn: Build}

// Build builds a request for an AWS Query service.
func Build(r *request.Request) {
	body := url.Values{}
	if err := queryutil.Parse(body, r.Params); err != nil {
		r.Error = errors.Errorf("%s: %s: %s", request.ErrCodeSerialization, "failed encoding Query request", err)
		return
	}

	if r.HTTPRequest.Method == "GET" {
		r.HTTPRequest.Method = "GET"
		r.HTTPRequest.URL.RawQuery = body.Encode()
		return
	}

	r.HTTPRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	r.SetBufferBody([]byte(body.Encode()))
}

type signer struct {
	// Values that must be populated from the request
	Request *http.Request
	Time    time.Time
	Key     string
	Secret  string

	Query        url.Values
	stringToSign string
	signature    string

	Debug bool
}

// SignRequestHandler is a named request handler the SDK will use to sign
// service client request with using the bilibili signature.
var SignRequestHandler = request.NamedHandler{
	Name: "appgwsdk.bilibili.SignRequestHandler", Fn: SignSDKRequest,
}

// SignSDKRequest requests with signature version 2.
//
// Will sign the requests with the service config's Credentials object
// Signing is skipped if the credentials is the credentials.AnonymousCredentials
// object.
func SignSDKRequest(req *request.Request) {
	// If the request does not need to be signed ignore the signing of the
	// request if the AnonymousCredentials object is used.
	if req.Config.Key == "" || req.Config.Secret == "" {
		return
	}

	v2 := signer{
		Request: req.HTTPRequest,
		Time:    req.Time,
		Key:     req.Config.Key,
		Secret:  req.Config.Secret,
		Debug:   req.Config.Debug,
	}

	req.Error = v2.Sign()

	if req.Error != nil {
		return
	}

	if req.HTTPRequest.Method == "POST" {
		// Set the body of the request based on the modified query parameters
		req.SetStringBody(v2.Query.Encode())

		// Now that the body has changed, remove any Content-Length header,
		// because it will be incorrect
		req.HTTPRequest.ContentLength = 0
		req.HTTPRequest.Header.Del("Content-Length")
	} else {
		req.HTTPRequest.URL.RawQuery = v2.Query.Encode()
	}
}

func (v2 *signer) Sign() error {
	if v2.Request.Method == "POST" {
		// Parse the HTTP request to obtain the query parameters that will
		// be used to build the string to sign. Note that because the HTTP
		// request will need to be modified, the PostForm and Form properties
		// are reset to nil after parsing.
		//nolint:errcheck
		v2.Request.ParseForm()
		v2.Query = v2.Request.PostForm
		v2.Request.PostForm = nil
		v2.Request.Form = nil
	} else {
		v2.Query = v2.Request.URL.Query()
	}

	v2.Query.Set("appkey", v2.Key)
	if v2.Query.Get("appsecret") != "" {
		log.Warn("utils http get must not have parameter appSecret")
	}
	v2.Query.Set("ts", strconv.FormatInt(v2.Time.Unix(), 10))
	tmp := v2.Query.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	var b bytes.Buffer
	b.WriteString(tmp)
	b.WriteString(v2.Secret)

	// build the canonical string for the V2 signature
	//nolint:gosimple
	v2.stringToSign = string(b.Bytes())

	mh := md5.Sum(b.Bytes())
	v2.signature = hex.EncodeToString(mh[:])

	v2.Query.Set("sign", v2.signature)

	if v2.Debug {
		v2.logSigningInfo()
	}

	return nil
}

const logSignInfoMsg = `DEBUG: Request Signature:
---[ STRING TO SIGN ]--------------------------------
%s
---[ SIGNATURE ]-------------------------------------
%s
-----------------------------------------------------`

func (v2 *signer) logSigningInfo() {
	log.Info(logSignInfoMsg, v2.stringToSign, v2.Query.Get("sign"))
}
