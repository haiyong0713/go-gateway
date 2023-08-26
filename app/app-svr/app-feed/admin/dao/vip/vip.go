package vip

import (
	"context"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
)

const showURL = "/api/ticket/project/getcard"

// Show .
func (d *Dao) Show(c context.Context, ids []int64) (rs map[int64]*show.Shopping, err error) {
	params := url.Values{}
	params.Set("id", xstr.JoinInts(ids))
	var res struct {
		Code int              `json:"errno"`
		Data []*show.Shopping `json:"data"`
	}
	url := d.c.Host.Vip + showURL
	if err = d.vipHTTPClient.Get(c, url, "", params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("Show url(%s) res code(%d)", url+params.Encode(), res.Code)
		return
	}
	rs = make(map[int64]*show.Shopping, len(res.Data))
	for _, r := range res.Data {
		rs[r.ID] = r
	}
	return
}
