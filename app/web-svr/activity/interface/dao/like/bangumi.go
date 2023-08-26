package like

import (
	"context"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const (
	_epPlayURI = "/pgc/internal/dynamic/v3/ep/list"
)

func (d *Dao) EpPlayer(c context.Context, epIDs []int64) (epPlayer map[int64]*lmdl.EpPlayer, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("ep_ids", xstr.JoinInts(epIDs))
	params.Set("ip", ip)
	var res struct {
		Code   int                      `json:"code"`
		Result map[int64]*lmdl.EpPlayer `json:"result"`
	}
	if err = d.client.Get(c, d.epPlayURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.epPlayURL+"?"+params.Encode())
		return
	}
	epPlayer = res.Result
	return
}
