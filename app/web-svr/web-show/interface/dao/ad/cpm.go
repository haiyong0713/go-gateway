package ad

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/log"
	"go-common/library/xstr"
	ad "go-gateway/app/web-svr/web-show/interface/model/resource"
)

// Cpms get ads from cpm platform
func (d *Dao) Cpms(c context.Context, val *ad.CpmsRequestParam) (advert *ad.Ad, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(val.Mid, 10))
	params.Set("sid", val.Sid)
	params.Set("buvid", val.Buvid)
	params.Set("resource", xstr.JoinInts(val.Ids))
	params.Set("ip", val.IP)
	params.Set("country", val.Country)
	params.Set("province", val.Province)
	params.Set("city", val.City)
	if val.Aid > 0 && val.UpID > 0 {
		params.Set("aid", strconv.FormatInt(val.Aid, 10))
		params.Set("av_up_id", strconv.FormatInt(val.UpID, 10))
	}
	params.Set("ua", val.UserAgent)
	params.Set("from_spmid", val.FromSpmID)
	if len(val.TagIds) > 0 {
		params.Set("av_tid", xstr.JoinInts(val.TagIds))
	}
	var res struct {
		Code int    `json:"code"`
		Data *ad.Ad `json:"data"`
	}
	if err = d.httpClient.Get(c, d.cpmURL, "", params, &res); err != nil {
		log.Error("cpm url(%s) error(%v)", d.cpmURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		err = fmt.Errorf("cpm api failed(%d)", res.Code)
		log.Error("url(%s) res code(%d) or res.data(%v)", d.cpmURL+"?"+params.Encode(), res.Code, res.Data)
		return
	}
	advert = res.Data
	bs, _ := json.Marshal(advert.AdsInfo)
	log.Info("d.Cpms arg:(%+v); data:(%+v)", val, string(bs))
	return
}
