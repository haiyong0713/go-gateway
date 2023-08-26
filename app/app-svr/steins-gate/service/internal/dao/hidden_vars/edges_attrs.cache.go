package hidden_vars

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

const _postfxGraphEdgeAttrs = "_atre" // graph edges' attrs

func (d *Dao) edgeAttrsCache(c context.Context, graphID int64) (attrsCache *model.EdgeAttrsCache, err error) {
	var (
		key  = edgeAttrsKey(graphID)
		conn = d.rds.Get(c)
		item []byte
	)
	defer conn.Close()
	if item, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	attrsCache = new(model.EdgeAttrsCache)
	if err = json.Unmarshal(item, &attrsCache); err != nil { // json反序列化
		log.Error("edgeAttrs Unmarshal error(%v)", err)
		attrsCache = nil
		return
	}
	return
}

func edgeAttrsKey(graphID int64) string {
	return strconv.FormatInt(graphID, 10) + _postfxGraphEdgeAttrs
}

// AddCacheEdgeAttrs .
func (d *Dao) AddCacheEdgeAttrs(c context.Context, graphID int64, edgeAttrs *model.EdgeAttrsCache) (err error) {
	if edgeAttrs == nil {
		return
	}
	var item []byte
	if item, err = json.Marshal(edgeAttrs); err != nil { // json序列化
		log.Error("edgeAttrs.Marshal error(%v)", err)
		return
	}
	key := edgeAttrsKey(graphID)
	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, item); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, edgeAttrs, err)
	}
	return

}
