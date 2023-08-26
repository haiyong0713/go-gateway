package stock

import (
	"context"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/stock"
	"net/http"
	"net/url"
)

// SyncStockWithOtherPlatform 与库存使用方，同步库存信息
func (d *Dao) SyncStockWithOtherPlatform(ctx context.Context, ackUrl string, syncParam *stock.SyncParamStruct) (ok bool, err error) {
	param := url.Values{}
	param.Add("unique_id", syncParam.RetryId)
	param.Add("stock_no", syncParam.StockNo)

	fullUrl := ackUrl + "?" + param.Encode()

	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, fullUrl, nil); err != nil {
		return
	}

	var res = struct {
		Code    int    `json:"code"`
		Data    bool   `json:"data"`
		Message string `json:"message"`
	}{}
	if err = d.httpClient.Do(ctx, req, &res); err != nil {
		log.Errorc(ctx, "TppShowInfo d.httpFate.Do uri(%s) error(%v)", ackUrl, err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrapf(ecode.Int(res.Code), "subcode:%d msg:%s", res.Code, res.Message)
		return
	}

	return res.Data, nil
}
