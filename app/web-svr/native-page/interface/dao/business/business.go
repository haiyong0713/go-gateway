package business

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/net/metadata"

	dymdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"

	"github.com/pkg/errors"
)

const (
	_sourceURI  = "/basc/api/open_api/v1/tab/source/detail"
	_productURI = "/basc/api/open_api/v1/tab/product_source/detail"
)

// SourceDetail .
func (d *Dao) SourceDetail(c context.Context, sourceID string, offset, ps int) (*dymdl.SourceReply, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("source_id", sourceID)
	params.Set("offset", strconv.Itoa(offset))
	params.Set("size", strconv.Itoa(ps))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int                `json:"code"`
		Data *dymdl.SourceReply `json:"data"`
	}
	if err := d.client.Get(c, d.businessSourceURL, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.businessSourceURL+"?"+params.Encode())
		return nil, err
	}
	if res.Data == nil {
		return &dymdl.SourceReply{}, nil
	}
	return res.Data, nil
}

func (d *Dao) ProductDetail(c context.Context, sourceID string, offset, ps int) (*dymdl.ProductReply, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("source_id", sourceID)
	params.Set("offset", strconv.Itoa(offset))
	params.Set("size", strconv.Itoa(ps))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int                 `json:"code"`
		Data *dymdl.ProductReply `json:"data"`
	}
	if err := d.client.Get(c, d.businessProduceURL, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.businessProduceURL+"?"+params.Encode())
		return nil, err
	}
	if res.Data == nil {
		return &dymdl.ProductReply{}, nil
	}
	return res.Data, nil
}
