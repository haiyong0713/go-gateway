package like

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"
)

const _viewData = "view:data:%d"

func viewDataKey(sid int64) string {
	return fmt.Sprintf(_viewData, sid)
}

func (d *Dao) AddViewDataCache(ctx context.Context, sid int64, data []*like.WebData) (err error) {
	if err = d.mcLike.Set(ctx, &memcache.Item{
		Key:        viewDataKey(sid),
		Object:     data,
		Flags:      memcache.FlagJSON,
		Expiration: 864000,
	}); err != nil {
		log.Errorc(ctx, "loadOperationData AddViewDataCache d.AddViewDataCache(sid:%d) error(%v)", sid, err)
	}
	return
}
