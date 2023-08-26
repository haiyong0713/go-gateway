package dubbing

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"strconv"

	"go-gateway/app/web-svr/activity/interface/model/dubbing"
)

// GetMidDubbingScore 获取稿件积分
func (d *dao) GetMidDubbingScore(c context.Context, mid int64) (res *dubbing.MapMidDubbingScore, err error) {
	var (
		key = buildKey(midScoreKey, mid)
	)
	conn := d.redis.Get(c)
	values := make(map[string]string)
	res = &dubbing.MapMidDubbingScore{}
	res.Score = make(map[int64]*dubbing.RedisDubbing)
	defer conn.Close()
	if values, err = redis.StringMap(conn.Do("HGETALL", key)); err != nil {
		if err != redis.ErrNil {
			log.Errorc(c, "conn.Do(HGETALL %s) error(%v)", key, err)
			return
		}
	}
	for k, v := range values {
		intArc, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			continue
		}
		value := dubbing.RedisDubbing{}
		json.Unmarshal([]byte(v), &value)
		res.Score[intArc] = &value
	}
	return res, nil
}
