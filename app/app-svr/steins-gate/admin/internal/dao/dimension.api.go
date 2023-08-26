package dao

import (
	"context"
	"net/http"

	"go-common/library/ecode"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/admin/internal/model"

	"github.com/pkg/errors"
)

// BvcDimension calls BVC's api to get HD dimension
func (d *Dao) BvcDimension(c context.Context, cid int64) (dimension *model.Dimension, err error) {
	var (
		resp = new(model.DimensionReply)
		req  *http.Request
	)
	if req, err = http.NewRequest("GET", d.bvcDimensionURL+d.bvcSign(cid), nil); err != nil {
		err = errors.Wrapf(ecode.Int(resp.Code), "url %s", d.bvcDimensionURL+d.bvcSign(cid))
		return
	}
	if err = d.httpVideoClient.Do(c, req, resp); err != nil {
		return
	}
	if resp.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(resp.Code), "message %s, url %s", resp.Message, d.bvcDimensionURL+d.bvcSign(cid))
		return
	}
	if len(resp.Info) == 0 {
		err = ecode.NothingFound
		return
	}
	dimension = resp.Info[0]
	return
}

func (d *Dao) ArcView(ctx context.Context, aid int64) (res *arcgrpc.SteinsGateViewReply, err error) {
	if res, err = d.arcClient.SteinsGateView(ctx, &arcgrpc.SteinsGateViewRequest{
		Aid: aid,
	}); err != nil {
		err = errors.Wrapf(err, "%v", aid)
		return
	}
	return

}
