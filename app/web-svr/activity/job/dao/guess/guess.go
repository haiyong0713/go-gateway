package guess

import (
	"context"
	"math"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	espmdl "go-gateway/app/web-svr/esports/interface/model"

	"github.com/pkg/errors"
)

const (
	_contestList = "/x/internal/esports/matchs/list"
	_firstPn     = 1
	_maxPs       = 50
)

// OidList get esports id.
func (d *Dao) OidList(c context.Context, sid int64) (rs []int64, err error) {
	var (
		count int
		tmpRs []int64
	)
	if rs, count, err = d.contestList(c, sid, _firstPn, _maxPs); err != nil {
		log.Error("d.contestList sid(%d) pn(%d) error(%+v)", sid, _firstPn, err)
		return
	}
	for i := 2; i <= count; i++ {
		if tmpRs, _, err = d.contestList(c, sid, i, _maxPs); err != nil {
			log.Error("d.contestList sid(%d) pn(%d) error(%+v)", sid, i, err)
			return
		}
		if len(tmpRs) > 0 {
			rs = append(rs, tmpRs...)
		}
	}
	log.Warn("OidList sid(%d) count(%d)", sid, len(rs))
	return
}

func (d *Dao) contestList(c context.Context, sid int64, pn, ps int) (rs []int64, count int, err error) {
	var (
		res struct {
			Code int `json:"code"`
			Data struct {
				List []*espmdl.Contest `json:"list"`
				Page struct {
					Num   int `json:"num"`
					Size  int `json:"size"`
					Total int `json:"total"`
				}
			} `json:"data"`
		}
	)
	params := url.Values{}
	params.Set("sids", strconv.FormatInt(sid, 10))
	params.Set("pn", strconv.Itoa(pn))
	params.Set("ps", strconv.Itoa(ps))
	if err = d.client.Get(c, d.contestsURL, "", params, &res); err != nil {
		err = errors.Wrap(err, "contestList:d.client.Get error")
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.contestsURL+"?"+params.Encode())
		return
	}
	if len(res.Data.List) == 0 {
		return
	}
	for _, contest := range res.Data.List {
		rs = append(rs, contest.ID)
	}
	count = int(math.Ceil(float64(res.Data.Page.Total) / float64(_maxPs)))
	return
}
