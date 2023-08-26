package s10

import (
	"context"
	"fmt"

	"go-gateway/app/web-svr/activity/admin/component"
	"go-gateway/app/web-svr/activity/admin/model/s10"

	"go-common/library/cache/memcache"
	"go-common/library/log"
)

const pointDetail = "s10:pd:%d"

func pointDetailkey(mid int64) string {
	return fmt.Sprintf(pointDetail, mid)
}

func (d *Dao) AddPointDetailCache(ctx context.Context, mid int64, detail []*s10.CostRecord) (err error) {
	if err = component.S10PointCostMC.Set(ctx, &memcache.Item{
		Key:        pointDetailkey(mid),
		Object:     detail,
		Flags:      memcache.FlagJSON,
		Expiration: d.pointDetailExpire,
	}); err != nil {
		log.Errorc(ctx, "d.AddPointDetailCache(mid:%d) error(%v)", mid, err)
	}
	return
}
