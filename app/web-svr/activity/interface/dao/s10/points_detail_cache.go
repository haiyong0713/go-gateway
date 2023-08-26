package s10

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/s10"
)

const pointDetail = "s10:pd:%d"

func pointDetailkey(mid int64) string {
	return fmt.Sprintf(pointDetail, mid)
}

func AddPointDetailCache(ctx context.Context, mid int64, details []*s10.CostRecord) (err error) {
	if err = component.S10PointCostMC.Set(ctx, &memcache.Item{
		Key:    pointDetailkey(mid),
		Object: details,
		Flags:  memcache.FlagJSON,
		//Expiration: service.PointDetailExpire,
	}); err != nil {
		log.Errorc(ctx, "s10 d.AddPointDetailCache(mid:%d) error(%v)", mid, err)
	}
	return
}

func PointDetailCache(ctx context.Context, mid int64) (res []*s10.CostRecord, err error) {
	if err = component.S10PointCostMC.Get(ctx, pointDetailkey(mid)).Scan(&res); err != nil {
		if err == memcache.ErrNotFound {
			return nil, nil
		}
		log.Error("s10 d.PointDetailCache(mid:%d) error(%v)", mid, err)
	}
	return
}

func DelPointDetailCache(ctx context.Context, mid int64) (err error) {
	if err = component.S10PointCostMC.Delete(ctx, pointDetailkey(mid)); err != nil {
		if err == memcache.ErrNotFound {
			return nil
		}
		log.Error("s10 d.DelPointDetailCache(mid:%d) error(%v)", mid, err)
	}
	return
}
