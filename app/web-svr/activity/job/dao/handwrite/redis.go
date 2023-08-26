package handwrite

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"
	mdl "go-gateway/app/web-svr/activity/job/model/handwrite"

	"github.com/pkg/errors"
)

// AddMidAward 添加用户获奖情况
func (d *dao) AddMidAward(c context.Context, midMap map[int64]*mdl.MidAward) (err error) {
	if len(midMap) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	for k, v := range midMap {
		var bs []byte
		if bs, err = json.Marshal(v); err != nil {
			log.Errorc(c, "AddMidAward json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(d.buildKey(awardKey, midKey, k)).Add(bs)
	}
	if _, err = redis.String(conn.Do("MSET", args...)); err != nil {
		err = errors.Wrap(err, "AddMidAward conn.Do(MSET)")
	}
	return
}

// GetMidAward 获得用户获奖情况
func (d *dao) GetMidAward(c context.Context, mid int64) (res *mdl.MidAward, err error) {
	var (
		bs   []byte
		key  = d.buildKey(awardKey, midKey, mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// SetAwardCount 获奖总人数统计
func (d *dao) SetAwardCount(c context.Context, awardCount *mdl.AwardCount) (err error) {
	var (
		bs   []byte
		key  = d.buildKey(awardCountKey)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(awardCount); err != nil {
		log.Errorc(c, "json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SET", key, bs); err != nil {
		log.Errorc(c, "SetAwardCount conn.Do(SET) key(%s) error(%v)", key, err)
	}
	return
}

// GetAwardCount 获取获奖总人数
func (d *dao) GetAwardCount(c context.Context) (res *mdl.AwardCount, err error) {
	var (
		bs   []byte
		key  = d.buildKey(awardCountKey)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// SetMidInitFans 设置用户初始fans
func (d *dao) SetMidInitFans(c context.Context, midMap map[int64]int64) (err error) {
	if len(midMap) == 0 {
		return
	}
	var (
		key    = d.buildKey(initFansKey)
		values map[string]int64
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	if values, err = redis.Int64Map(conn.Do("HGETALL", key)); err != nil {
		if err != redis.ErrNil {
			log.Errorc(c, "conn.Do(HGETALL %s) error(%v)", key, err)
			return
		}
	}
	args = args.Add(key)
	for k, v := range midMap {
		midString := strconv.FormatInt(k, 10)
		if _, ok := values[midString]; !ok {
			args = args.Add(k).Add(v)
		}
	}
	if len(args) > 1 {
		if _, err = redis.String(conn.Do("HMSET", args...)); err != nil {
			err = errors.Wrap(err, "SetMidInitFans conn.Do(HSETNX)")
		}
	}
	return
}

// GetMidInitFans 获取用户初始fans
func (d *dao) GetMidInitFans(c context.Context) (values map[string]int64, err error) {
	var (
		key = d.buildKey(initFansKey)
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	if values, err = redis.Int64Map(conn.Do("HGETALL", key)); err != nil {
		log.Errorc(c, "conn.Do(HGETALL %s) error(%v)", key, err)
	}
	return
}

// CacheActivityMember 本次参与活动的mid
func (d *dao) CacheActivityMember(c context.Context, mids []int64) (err error) {
	var (
		bs   []byte
		key  = d.buildKey(activityMid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(mids); err != nil {
		log.Errorc(c, "json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SET", key, bs); err != nil {
		log.Errorc(c, "CacheActivityMember conn.Do(SET) key(%s) error(%v)", key, err)
	}
	return
}

// GetActivityMember 获得本次参与活动的所有mid
func (d *dao) GetActivityMember(c context.Context) (res []int64, err error) {
	var (
		bs   []byte
		key  = d.buildKey(activityMid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// GetMidAward 获得用户获奖情况
func (d *dao) GetMidsAward(c context.Context, mids []int64) (res map[int64]*mdl.MidAward, err error) {
	if len(mids) == 0 {
		return
	}
	var (
		bss [][]byte

		conn = d.redis.Get(c)
	)
	defer conn.Close()
	res = map[int64]*mdl.MidAward{}
	args := redis.Args{}
	for _, v := range mids {
		args = args.Add(d.buildKey(awardKey, midKey, v))
	}
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "PlsCache conn.Do(MGET,%s) error(%v)", args, err)
		}
		return
	}
	for _, bs := range bss {
		award := &mdl.MidAward{}
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, award); err != nil {
			log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		res[award.Mid] = award
	}
	return
}

// AddMidTask 添加用户获奖情况
func (d *dao) AddMidTask(c context.Context, midMap map[int64]*mdl.MidTaskAll) (err error) {
	if len(midMap) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	var keys []string

	for k, v := range midMap {
		var bs []byte
		if bs, err = json.Marshal(v); err != nil {
			log.Errorc(c, "AddMidTask json.Marshal() error(%v)", err)
			return
		}
		key := d.buildKey(taskKey, midKey, k)
		keys = append(keys, key)
		args = args.Add(key).Add(bs)
	}

	if err = conn.Send("MSET", args...); err != nil {
		log.Errorc(c, "AddMidTask conn.Send(MSET) error(%v)", err)
		return
	}
	count := 1
	for _, v := range keys {
		count++
		if err := conn.Send("EXPIRE", v, 2592000); err != nil {
			return err
		}
	}
	if err := conn.Flush(); err != nil {
		log.Errorc(c, "AddMidTask Flush error(%v)", err)
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Errorc(c, "AddMidTask conn.Receive() error(%v)", err)
			return err
		}
	}
	return
}

// SetTaskCount 获奖总人数统计
func (d *dao) SetTaskCount(c context.Context, awardCount *mdl.AwardCountNew) (err error) {
	var (
		bs   []byte
		key  = d.buildKey(taskCountKey)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(awardCount); err != nil {
		log.Errorc(c, "json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SETEX", key, 2592000, bs); err != nil {
		log.Errorc(c, "SetTaskCount conn.Do(SET) key(%s) error(%v)", key, err)
	}
	return
}

// GetTaskCount 获取获奖总人数
func (d *dao) GetTaskCount(c context.Context) (res *mdl.AwardCountNew, err error) {
	var (
		bs   []byte
		key  = d.buildKey(taskCountKey)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}
