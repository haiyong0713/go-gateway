package hidden_vars

import (
	"context"

	"go-common/library/log"
	"go-common/library/stat/prom"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

// edgeAttrsByGraph 专用于隐藏变量，是一个graph中所有包含attribute的边的缓存
func (d *Dao) edgeAttrsByGraph(c context.Context, graphID int64) (attrs *model.EdgeAttrsCache, err error) {
	var cached = true
	if attrs, err = d.edgeAttrsCache(c, graphID); err != nil {
		log.Error("%+v", err)
		err = nil
		cached = false
	}
	if attrs != nil {
		prom.CacheHit.Incr("EdgeAttrs")
		return
	}
	prom.CacheMiss.Incr("EdgeAttrs")
	if attrs, err = d.RawGraphEdgeAttrs(c, graphID); err != nil {
		log.Error("%+v", err)
		return
	}
	if !cached {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		//nolint:errcheck
		d.AddCacheEdgeAttrs(c, graphID, attrs)
	})
	return

}
