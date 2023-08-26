package show

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/show"
)

const (
	_prefixAiChannelResKey = "aicr_"
)

func aiChannelResKey(topEid int64) string {
	return _prefixAiChannelResKey + strconv.FormatInt(topEid, 10)
}

// CacheAIChannelRes get a AIChannelRes info from cache.
func (d *Dao) CacheAIChannelRes(c context.Context, topEid int64) (res []*show.PopularCard, err error) {
	var (
		key  = aiChannelResKey(topEid)
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
	if err = json.Unmarshal(item, &res); err != nil {
		log.Error("CacheAIChannelRes Unmarshal error(%v)", err)
	}
	return
}
