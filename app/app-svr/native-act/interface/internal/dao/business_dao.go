package dao

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

const (
	_bizSourceUri  = "/basc/api/open_api/v1/tab/source/detail"         //数据源详情查询
	_bizProductUri = "/basc/api/open_api/v1/tab/product_source/detail" //商品卡数据源详情查询
)

type businessDao struct {
	host       string
	httpClient *bm.Client
}

func (d *businessDao) SourceDetail(ctx context.Context, req *model.SourceDetailReq) (*model.SourceDetailRly, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("source_id", req.SourceId)
	params.Set("offset", strconv.FormatInt(req.Offset, 10))
	params.Set("size", strconv.FormatInt(req.Size, 10))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int                    `json:"code"`
		Data *model.SourceDetailRly `json:"data"`
	}
	if err := d.httpClient.Get(ctx, d.host+_bizSourceUri, ip, params, &res); err != nil {
		log.Error("Fail to request business.SourceDetail, params=%+v error=%+v", params.Encode(), err)
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.host+_bizSourceUri+"?"+params.Encode())
		log.Error("Fail to request business.SourceDetail, error=%+v", err)
		return nil, err
	}
	if res.Data == nil {
		return nil, nil
	}
	return res.Data, nil
}

func (d *businessDao) ProductDetail(ctx context.Context, req *model.ProductDetailReq) (*model.ProductDetailRly, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("source_id", req.SourceId)
	params.Set("offset", strconv.FormatInt(req.Offset, 10))
	params.Set("size", strconv.FormatInt(req.Size, 10))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int                     `json:"code"`
		Data *model.ProductDetailRly `json:"data"`
	}
	if err := d.httpClient.Get(ctx, d.host+_bizProductUri, ip, params, &res); err != nil {
		log.Error("Fail to request business.ProductDetail, params=%+v error=%+v", params.Encode(), err)
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.host+_bizProductUri+"?"+params.Encode())
		log.Error("Fail to request business.ProductDetail, error=%+v", err)
		return nil, err
	}
	if res.Data == nil {
		return nil, nil
	}
	return res.Data, nil
}
