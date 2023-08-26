package reward_conf

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/cost"
)

func awardListByDateKey(activityId string, date string) string {
	return fmt.Sprintf("%s_award_list_%s", activityId, date)
}

// CacheAwardList get data from redis
func (d *dao) CacheAwardList(c context.Context, id string, date string) (res []*cost.AwardConfigDataDB, err error) {
	key := awardListByDateKey(id, date)
	reply, err1 := redis.Bytes(d.redis.Do(c, "GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheAwardList(get key: %v) err: %+v", key, err)
		return
	}
	res = []*cost.AwardConfigDataDB{}
	err = json.Unmarshal(reply, &res)
	if err != nil {
		log.Errorc(c, "d.CacheAwardList(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheAwardList Set data to redis
func (d *dao) AddCacheAwardList(c context.Context, id string, val string, value []*cost.AwardConfigDataDB) (err error) {
	if len(val) == 0 {
		return
	}
	key := awardListByDateKey(id, val)
	var bs []byte
	if bs, err = json.Marshal(value); err != nil {
		log.Errorc(c, "AddCacheAwardList json.Marshal(%v) error (%v)", value, err)
		return
	}
	expire := d.awardListExpire
	if value == nil {
		expire = 300
	}
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheAwardList(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheAwardListByDate delete data from redis
func (d *dao) DelCacheAwardListByDate(c context.Context, id string, date string) (err error) {
	key := awardListByDateKey(id, date)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheAwardListByDate(get key: %v) err: %+v", key, err)
		return
	}
	return
}
