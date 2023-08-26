package es

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	esmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/es"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/search"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"github.com/pkg/errors"
)

const (
	_searchChannel    = "/x/admin/search"
	_searchChannelSQL = `deal SELECT * FROM link_channel WHERE name LIKE "%v" AND state=%d LIMIT %d,%d`
)

func (d *Dao) SearchChannel(c context.Context, mid int64, keyword string, pn, ps, state int) (st *search.ChannelResult, tids []int64, err error) {
	var (
		req     *http.Request
		ip      = metadata.String(c, metadata.RemoteIP)
		trackID = fmt.Sprintf("%d%d", time.Now().Unix(), mid)
		sql     = fmt.Sprintf(_searchChannelSQL, keyword, state, (pn-1)*ps, ps)
	)
	params := url.Values{}
	params.Set("sql", sql)
	// new request
	if req, err = d.client.NewRequest("GET", d.searchChannel, ip, params); err != nil {
		log.Error("%v", err)
		return
	}
	var res struct {
		Code int `json:"code"`
		Data struct {
			Order  string                 `json:"order"`
			Sort   string                 `json:"sort"`
			Result []*esmdl.SearchChannel `json:"result"`
			Page   struct {
				Num   int `json:"num"`
				Size  int `json:"size"`
				Total int `json:"total"`
			} `json:"page"`
		} `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.searchChannel+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.searchChannel+"?"+params.Encode())
		return
	}
	for _, result := range res.Data.Result {
		if result == nil {
			continue
		}
		if result.CID != 0 {
			tids = append(tids, result.CID)
		}
	}
	st = &search.ChannelResult{TrackID: trackID, Pages: res.Data.Page.Num, Total: res.Data.Page.Total}
	return
}
