package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"
)

const _onlineTotalURI = "/x/internal/broadcast/online/total"

// OnlineList get online list
func (d *Dao) OnlineList(c context.Context, num int64) (data []*model.OnlineAid, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("num", strconv.FormatInt(num, 10))
	var res struct {
		Code int                `json:"code"`
		Data []*model.OnlineAid `json:"data"`
	}
	if err = d.httpR.Get(c, d.onlineListURL, ip, params, &res); err != nil {
		log.Error(" d.httpR.Get.Get(%s) error(%v)", d.onlineListURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("d.httpR.Get(%s) code error(%d)", d.onlineListURL, res.Code)
		return
	}
	data = res.Data
	return
}

// OnlineTotal get online total
func (d *Dao) OnlineTotal(c context.Context) (*model.OnlineTotal, error) {
	var res struct {
		Code int                `json:"code"`
		Data *model.OnlineTotal `json:"data"`
	}
	if err := d.httpR.Get(c, d.onlineTotalURL, "", url.Values{}, &res); err != nil {
		log.Error("OnlineTotal d.httpR.Get.Get(%s) error(%v)", d.onlineTotalURL, err)
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		log.Error("OnlineTotal d.httpR.Get(%s) code error(%d)", d.onlineTotalURL, res.Code)
		return nil, ecode.Int(res.Code)
	}
	return res.Data, nil
}
