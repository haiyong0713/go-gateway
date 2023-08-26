package feed

import (
	"context"

	"go-common/library/ecode"

	"github.com/pkg/errors"
)

const (
	_hot    = "/data/rank/reco-tmzb.json"
	_rcmdUp = "/x/feed/rcmd/up"
)

func (d *Dao) Hots(c context.Context) (aids []int64, err error) {
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid int64 `json:"aid"`
		} `json:"list"`
	}
	if err = d.clientAsyn.Get(c, d.hot, "", nil, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.hot)
		return
	}
	if len(res.List) == 0 {
		return
	}
	aids = make([]int64, 0, len(res.List))
	for _, list := range res.List {
		if list.Aid != 0 {
			aids = append(aids, list.Aid)
		}
	}
	return
}
