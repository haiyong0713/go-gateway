package main

import (
	"context"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/net/netutil/breaker"
	xtime "go-common/library/time"
)

// NewHttpClient ...
func NewHttpClient(appKey, secret string) {
	_httpClient = bm.NewClient(&bm.ClientConfig{
		App: &bm.App{
			Key:    appKey,
			Secret: secret,
		},
		Dial:      xtime.Duration(time.Second),
		Timeout:   xtime.Duration(time.Second),
		KeepAlive: xtime.Duration(time.Second),
		Breaker: &breaker.Config{
			Window:  10 * xtime.Duration(time.Second),
			Sleep:   50 * xtime.Duration(time.Millisecond),
			Bucket:  10,
			Ratio:   0.5,
			Request: 100,
		},
	})
	return
}

// filterWord ...
func filterWord(c context.Context, msg string) (respData *FilterData, err error) {
	params := url.Values{}
	params.Set("area", _filterArea)
	params.Set("msg", msg)

	respValue := &filterResp{}
	ip := metadata.String(c, metadata.RemoteIP)
	if err = _httpClient.Get(c, _reqUrl, ip, params, &respValue); err != nil {
		log.Error("filterWord httpClient reqURL(%s) error(%v)", _reqUrl+"?"+params.Encode(), err)
		return
	}
	if respValue.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(respValue.Code), "filterWord failed, reqURL(%s)", _reqUrl)
		return
	}
	if respValue.Data != nil {
		respData = respValue.Data
	}
	return
}
