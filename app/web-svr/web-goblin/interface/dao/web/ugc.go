package web

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go-common/library/database/elastic"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	webmdl "go-gateway/app/web-svr/web-goblin/interface/model/web"
)

const (
	_ugcIncre      = "web_goblin"
	_rankRegionURL = "all_region-%d-%d"
)

// UgcIncre ugc increment .
func (d *Dao) UgcIncre(ctx context.Context, pn, ps int, start, end int64) (res []*webmdl.SearchAids, err error) {
	var (
		startStr, endStr string
		rs               struct {
			Result []*webmdl.SearchAids `json:"result"`
		}
	)
	startStr = time.Unix(start, 0).Format("2006-01-02 15:04:05")
	endStr = time.Unix(end, 0).Format("2006-01-02 15:04:05")
	r := d.ela.NewRequest(_ugcIncre).WhereRange("mtime", startStr, endStr, elastic.RangeScopeLoRo).Fields("aid").Fields("action").Index(_ugcIncre).Pn(pn).Ps(ps)
	if err = r.Scan(ctx, &rs); err != nil {
		log.Error("r.Scan error(%v)", err)
		return
	}
	res = rs.Result
	return
}

// Ranking get data from bigdata .
func (d *Dao) Ranking(c context.Context, rid, day int) (res []*webmdl.NewArchive, err error) {
	var (
		params   = url.Values{}
		remoteIP = metadata.String(c, metadata.RemoteIP)
	)
	var rs struct {
		Code int                  `json:"code"`
		List []*webmdl.NewArchive `json:"list"`
	}
	if err = d.httpR.RESTfulGet(c, d.rankURL, remoteIP, params, &rs, fmt.Sprintf(_rankRegionURL, rid, day)); err != nil {
		log.Error("d.httpR.RESTfulGet() rid(%d) day(%d) error(%v)", rid, day, err)
		return
	}
	if rs.Code != ecode.OK.Code() {
		log.Error("d.httpR.RESTfulGet() rid(%d) day(%d) error code(%d)", rid, day, rs.Code)
		err = ecode.Int(rs.Code)
		return
	}
	res = rs.List
	return
}
