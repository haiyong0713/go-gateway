package black

import (
	"context"
)

const (
	_blackURL = "/privatedata/reco-deny-arcs.json"
)

// Black returns blacklist of aids
func (d *Dao) Black(c context.Context) (black map[int64]struct{}, err error) {
	var res []int64
	if err = d.clientAsyn.Get(c, d.blackURL, "", nil, &res); err != nil {
		return
	}
	if len(res) == 0 {
		return
	}
	black = make(map[int64]struct{}, len(res))
	for _, aid := range res {
		black[aid] = struct{}{}
	}
	return
}
