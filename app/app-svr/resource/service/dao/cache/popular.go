package cache

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
)

const (
	_prefixAiChannelResKey = "aicr_"
	_gotoAv                = "av"
	_maxLen                = 200
)

func aiChannelResKey(topEid int64) string {
	return _prefixAiChannelResKey + strconv.FormatInt(topEid, 10)
}

func (d *Dao) CacheAIChannelRes(c context.Context, topEid int64) ([]int64, error) {
	var (
		key  = aiChannelResKey(topEid)
		conn = d.redisEntrance.Conn(c)
		item []byte
		list []*PopularCard
		res  []int64
		err  error
	)
	defer func() {
		conn.Close()
		if len(res) == 0 {
			log.Error("CacheAIChannelResRetrunLen0 topEid(%d) err(%+v)", topEid, err)
		}
	}()
	if item, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = ecode.NothingFound
			return nil, err
		}
		log.Error("conn.Do(GET, %s) error(%v)", key, err)
		return nil, err
	}
	if err = json.Unmarshal(item, &list); err != nil {
		log.Error("CacheAIChannelRes Unmarshal error(%v),id(%d)", err, topEid)
		return nil, err
	}
	for _, rcd := range list {
		if rcd.Type == _gotoAv {
			res = append(res, rcd.ID)
		}
	}
	if len(res) > _maxLen {
		res = res[:_maxLen]
	}
	log.Info("CacheAIChannelRes return %+v", res)
	return res, nil
}

type PopularCard struct {
	ID   int64  `json:"value"`
	Type string `json:"type"`
}
