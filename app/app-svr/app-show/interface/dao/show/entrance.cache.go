package show

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-show/interface/model/card"
)

const (
	_prefixAiChannelResKey = "aicr_"
	_reasonType            = 3
)

func aiChannelResKey(topEid int64) string {
	return _prefixAiChannelResKey + strconv.FormatInt(topEid, 10)
}

// CacheAIChannelRes get AIChannelRes info from cache.
func (d *Dao) CacheAIChannelRes(c context.Context, topEid int64) (res []*card.PopularCard, err error) {
	var (
		key  = aiChannelResKey(topEid)
		conn = d.redis.Get(c)
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
		log.Error("CacheAIChannelRes Unmarshal error(%v),id(%d)", err, topEid)
	}
	for i := 0; i < len(res); i++ {
		if res[i].Reason != "" {
			res[i].ReasonType = _reasonType
		}
	}
	log.Info("CacheAIChannelRes return %+v", &res)
	return
}
