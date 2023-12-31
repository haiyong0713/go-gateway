package dao

import (
	"context"
	"net/url"
	"strconv"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/space/ecode"
	"go-gateway/app/web-svr/space/interface/model"
)

const (
	_shopURI     = "/mall-shop/merchant/enter/service/shop/info"
	_shopLinkURI = "/mall-shop/merchant/enter/service/shop/get"
)

// ShopInfo get shop info data for pc.
func (d *Dao) ShopInfo(c context.Context, mid int64) (data *model.ShopInfo, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Shop *model.ShopInfo `json:"shop"`
		} `json:"data"`
	}
	if err = d.httpR.Get(context.Background(), d.shopURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil { // TODO 电商接口traceid报错，先兼容屏蔽
		log.Error("ShopInfo(%s) mid(%d) error(%v)", d.shopURL+params.Encode(), mid, err)
		return
	}
	if res.Code != xecode.OK.Code() {
		log.Error("ShopInfo(%s) mid(%d) code(%d) error", d.shopURL+params.Encode(), mid, res.Code)
		err = ecode.SpaceNoShop
		return
	}
	data = res.Data.Shop
	return
}

// ShopLink only get simply data for h5.
func (d *Dao) ShopLink(c context.Context, mid int64, platform int) (data *model.ShopLinkInfo, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("type", strconv.Itoa(platform))
	var res struct {
		Code int                 `json:"code"`
		Data *model.ShopLinkInfo `json:"data"`
	}
	if err = d.httpR.Get(context.Background(), d.shopLinkURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil { // TODO 电商接口traceid报错，先兼容屏蔽
		log.Error("ShopLink(%s) mid(%d) error(%v)", d.shopLinkURL+params.Encode(), mid, err)
		return
	}
	if res.Code != xecode.OK.Code() {
		log.Error("ShopLink(%s) mid(%d) code(%d) error", d.shopLinkURL+params.Encode(), mid, res.Code)
		err = ecode.SpaceNoShop
		return
	}
	data = res.Data
	return
}
