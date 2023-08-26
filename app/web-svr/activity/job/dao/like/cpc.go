package like

import (
	"context"
	"go-common/library/net/netutil"
	"go-common/library/retry"
)

const cpc100PVCacheKey = "cpc100_pv"
const cpc100TVCacheKey = "cpc100_tv"

func (d *Dao) CpcSetPV(ctx context.Context, pv int64) (err error) {
	conn := d.redis.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()
	err = retry.WithAttempts(ctx, "CpcSetPV", 5, netutil.DefaultBackoffConfig, func(c context.Context) (err error) {
		_, err = conn.Do("SETEX", cpc100PVCacheKey, 5184000, pv)
		return
	})
	return
}

func (d *Dao) CpcSetTopicView(ctx context.Context, topicView int64) (err error) {
	conn := d.redis.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()
	err = retry.WithAttempts(ctx, "CpcSetTopicView", 5, netutil.DefaultBackoffConfig, func(c context.Context) (err error) {
		_, err = conn.Do("SETEX", cpc100TVCacheKey, 5184000, topicView)
		return
	})
	return
}
