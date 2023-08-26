package cpm

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/resource/service/model"

	location "git.bilibili.co/bapis/bapis-go/community/service/location"
)

// CpmsAPP get ads from cpm platform.
func (d *Dao) CpmsAPP(c context.Context, aid, mid, splashID int64, build int, resource, mobiApp, device, buvid, network, openEvent, adExtra string, ip *location.InfoReply) (adr *model.ADRequest, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("build", strconv.Itoa(build))
	params.Set("buvid", buvid)
	params.Set("resource", resource)
	params.Set("mobi_app", mobiApp)
	params.Set("ip", ip.Addr)
	if aid != 0 {
		params.Set("aid", strconv.FormatInt(aid, 10))
	}
	if device != "" {
		params.Set("device", device)
	}
	if ip.Country != "" {
		params.Set("country", ip.Country)
	}
	if ip.Province != "" {
		params.Set("province", ip.Province)
	}
	if ip.City != "" {
		params.Set("city", ip.City)
	}
	if network != "" {
		params.Set("network", network)
	}
	if openEvent != "" {
		params.Set("open_event", openEvent)
	}
	if adExtra != "" {
		params.Set("ad_extra", adExtra)
	}
	if splashID != 0 {
		params.Set("top_view_id", strconv.FormatInt(splashID, 10))
	}
	var res struct {
		Code int              `json:"code"`
		Data *model.ADRequest `json:"data"`
		Msg  string           `json:"message"`
	}
	if err = d.httpClient.Get(c, d.cpmAppURL, ip.Addr, params, &res); err != nil {
		log.Error("CpmsAPP url(%s) error(%v)", d.cpmAppURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		err = fmt.Errorf("CpmsAPP api failed(%d)", res.Code)
		log.Error("CpmsApp url(%s) res code(%d) or res.data(%v)", d.cpmAppURL+"?"+params.Encode(), res.Code, res.Data)
		return
	}
	adr = res.Data
	return
}
