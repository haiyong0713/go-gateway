package business

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-show/interface/conf"
	busmdl "go-gateway/app/app-svr/app-show/interface/model/business"

	"github.com/pkg/errors"
)

const (
	_sourceURI  = "/basc/api/open_api/v1/tab/source/detail"
	_productURI = "/basc/api/open_api/v1/tab/product_source/detail"
)

type Dao struct {
	c      *conf.Config
	client *httpx.Client
	//企业号-商单相关http接口
	businessSourceURL  string
	businessProduceURL string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:                  c,
		client:             httpx.NewClient(c.HTTPBusiness),
		businessSourceURL:  c.Host.Business + _sourceURI,
		businessProduceURL: c.Host.Business + _productURI,
	}
	return
}

// SourceDetail .
func (d *Dao) SourceDetail(c context.Context, sourceID string, offset, ps int64) (*busmdl.SourceReply, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("source_id", sourceID)
	params.Set("offset", strconv.FormatInt(offset, 10))
	params.Set("size", strconv.FormatInt(ps, 10))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int                 `json:"code"`
		Data *busmdl.SourceReply `json:"data"`
	}
	if err := d.client.Get(c, d.businessSourceURL, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.businessSourceURL+"?"+params.Encode())
		return nil, err
	}
	if res.Data == nil {
		return &busmdl.SourceReply{}, nil
	}
	return res.Data, nil
}

func (d *Dao) ProductDetail(c context.Context, sourceID string, offset, ps int64) (*busmdl.ProductReply, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("source_id", sourceID)
	params.Set("offset", strconv.FormatInt(offset, 10))
	params.Set("size", strconv.FormatInt(ps, 10))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int                  `json:"code"`
		Data *busmdl.ProductReply `json:"data"`
	}
	if err := d.client.Get(c, d.businessProduceURL, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.businessProduceURL+"?"+params.Encode())
		return nil, err
	}
	if res.Data == nil {
		return &busmdl.ProductReply{}, nil
	}
	return res.Data, nil
}
