package region

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	dynGRPC "go-gateway/app/web-svr/dynamic/service/api/v1"
)

// regionDynamic
func (d *Dao) RegionDynamic(ctx context.Context, rid int64, pn, ps int) ([]*api.Arc, error) {
	arg := &dynGRPC.RegionArcs3Req{
		Rid: rid,
		Pn:  int64(pn),
		Ps:  int64(ps),
	}
	as, err := d.dynClient.RegionArcs3(ctx, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return as.GetArcs(), nil
}

func (d *Dao) RanksArcs(ctx context.Context, rid, pn, ps int64) ([]*api.Arc, error) {
	arg := &dynGRPC.RegAllReq{
		Rid: rid,
		Pn:  pn,
		Ps:  ps,
	}
	reply, err := d.dynClient.RegAllArcs(ctx, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetArchives(), nil
}
