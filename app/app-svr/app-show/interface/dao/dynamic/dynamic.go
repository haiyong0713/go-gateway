package dynamic

import (
	"context"
	"errors"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/archive/service/api"
	dynarc "go-gateway/app/web-svr/dynamic/service/model"
	dynrpc "go-gateway/app/web-svr/dynamic/service/rpc/client"
)

// Dao is rpc dao.
type Dao struct {
	// dynamic rpc
	dynRpc *dynrpc.Service
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// dynamic rpc
		dynRpc: dynrpc.New(c.DynamicRPC),
	}
	return
}

// regionDynamic
func (d *Dao) RegionDynamic(ctx context.Context, rid, pn, ps int) (res []*api.Arc, aids []int64, err error) {
	arg := &dynarc.ArgRegion3{
		RegionID: int32(rid),
		Pn:       pn,
		Ps:       ps,
	}
	var as *dynarc.DynamicArcs3
	if as, err = d.dynRpc.RegionArcs3(ctx, arg); err != nil {
		log.Error("d.dynRpc.RegionArcs(%v) error(%v)", arg, err)
		return
	}
	if as != nil {
		res = as.Archives
		for _, a := range res {
			if a == nil {
				err = errors.New("RegionDynamic a is nil")
				return
			}
			aids = append(aids, a.Aid)
		}
	}
	return
}
