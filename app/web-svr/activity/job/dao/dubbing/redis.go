package dubbing

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/dubbing"
)

// AddArchiveScore 增加稿件积分
func (d *dao) AddArchiveScore(c context.Context, sid int64, batch int64, index int, list map[int64]int64) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		conn = d.redis.Get(c)
		key  = buildKey(archiveKey, batch, sid, index)
		args = redis.Args{}.Add(key)
	)
	defer conn.Close()
	for k, v := range list {
		args = args.Add(k).Add(v)
	}
	if err = conn.Send("HMSET", args...); err != nil {
		log.Errorc(c, "AddCacheLotteryTimes conn.Send(HMSET) error(%v)", err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.archiveScoreExpire); err != nil {
		log.Errorc(c, "conn.Send(EXPIRE, %s) error(%v)", key, err)
		return
	}
	if err := conn.Flush(); err != nil {
		log.Errorc(c, "AddCacheLotteryTimes Flush error(%v)", err)
		return err
	}
	for i := 0; i < 2; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Errorc(c, "AddCacheLotteryTimes conn.Receive() error(%v)", err)
			return err
		}
	}

	return
}

// GetArchiveScore 获取稿件积分
func (d *dao) GetArchiveScore(c context.Context, sid int64, batch int64, index int) (res map[int64]int64, err error) {
	var (
		key = buildKey(archiveKey, batch, sid, index)
	)
	conn := d.redis.Get(c)
	values := make(map[string]int64)
	res = make(map[int64]int64)
	defer conn.Close()
	if values, err = redis.Int64Map(conn.Do("HGETALL", key)); err != nil {
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
		res[intArc] = v

	}
	return res, nil
}

// SetDubbingMidScore 设置用户维度的排行榜结果
func (d *dao) SetDubbingMidScore(c context.Context, mid int64, midScore *dubbing.MapMidDubbingScore) (err error) {
	if midScore == nil || len(midScore.Score) == 0 {
		return
	}
	var (
		conn = d.redis.Get(c)
		key  = buildKey(midScoreKey, mid)
		args = redis.Args{}.Add(key)
	)
	defer conn.Close()
	for k, v := range midScore.Score {
		paramJSON, _ := json.Marshal(v)
		args = args.Add(k).Add(string(paramJSON))
	}
	if err = conn.Send("HMSET", args...); err != nil {
		log.Errorc(c, "SetDubbingMidScore conn.Send(HMSET) error(%v)", err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.archiveScoreExpire); err != nil {
		log.Errorc(c, "conn.Send(EXPIRE, %s) error(%v)", key, err)
		return
	}
	if err1 := conn.Flush(); err1 != nil {
		log.Errorc(c, "SetDubbingMidScore Flush error(%v)", err1)
		return err
	}
	for i := 0; i < 2; i++ {
		if _, err2 := conn.Receive(); err2 != nil {
			log.Errorc(c, "SetDubbingMidScore conn.Receive() error(%v)", err2)
			return err2
		}
	}
	return

}
