package college

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/college"
)

// CacheGetMidCollege
func (d *dao) CacheGetMidCollege(c context.Context, mid int64) (res *college.PersonalCollege, err error) {
	var (
		key  = buildKey(midCollegeKey, mid)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		log.Errorc(c, "GetMidCollege conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// CacheSetMidCollege ...
func (d *dao) CacheSetMidCollege(c context.Context, mid int64, data *college.PersonalCollege) (err error) {
	var (
		key  = buildKey(midCollegeKey, mid)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", data, err)
		return
	}
	if err = conn.Send("SETEX", key, d.collegeMidCollegeExpire, bs); err != nil {
		log.Errorc(c, "conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.collegeMidCollegeExpire, string(bs), err)
	}
	return
}

// CacheGetCollegeDetail ...
func (d *dao) CacheGetCollegeDetail(c context.Context, collegeID int64, version int) (res *college.Detail, err error) {
	key := buildKey(collegeKey, version, collegeID)
	args := redis.Args{}.Add(key)
	args = args.Add(idKey)
	args = args.Add(scoreKey)
	args = args.Add(nationwideRankKey)
	args = args.Add(provinceKey)
	args = args.Add(provinceRankKey)
	args = args.Add(tabListKey)
	args = args.Add(relationKey)
	args = args.Add(nameKey)
	conn := d.redis.Get(c)
	defer conn.Close()
	var values [][]byte
	res = &college.Detail{}
	if values, err = redis.ByteSlices(conn.Do("HMGET", args...)); err != nil {
		log.Errorc(c, "CacheGetCollegeDetail redis.ByteSlices(%s,%v) error(%v)", key, values, err)
		return
	}
	if len(values) < len(args)-1 {
		return nil, errors.New("CacheGetCollegeDetail hmget values error")
	}
	collegeIDRedis, err := strconv.ParseInt(string(values[0]), 10, 64)
	if err != nil {
		return nil, errors.New("CacheGetCollegeDetail hmget values  collegeID error")
	}

	scoreRedis, err := strconv.ParseInt(string(values[1]), 10, 64)
	if err != nil {
		return nil, errors.New("CacheGetCollegeDetail hmget values score error")
	}
	nationwideRank, err := strconv.Atoi(string(values[2]))
	if err != nil {
		return nil, errors.New("CacheGetCollegeDetail hmget values nationwide error")
	}
	province := string(values[3])
	provinceRank, err := strconv.Atoi(string(values[4]))
	if err != nil {
		return nil, errors.New("CacheGetCollegeDetail hmget values provinceRank error")
	}
	tabList, err := xstr.SplitInts(string(values[5]))
	if err != nil {
		return nil, errors.New("CacheGetCollegeDetail hmget values tabList error")
	}
	relation, err := xstr.SplitInts(string(values[6]))
	if err != nil {
		return nil, errors.New("CacheGetCollegeDetail hmget values relation error")
	}
	name := string(values[7])
	res.ID = collegeIDRedis
	res.Score = scoreRedis
	res.NationwideRank = nationwideRank
	res.Province = province
	res.ProvinceRank = provinceRank
	res.TabList = tabList
	res.RelationMid = relation
	res.Name = name
	return res, nil
}

// GetCollegeVersion 获得本次计算的版本
func (d *dao) GetCollegeVersion(c context.Context) (res *college.Version, err error) {
	res = &college.Version{}
	var (
		key  = buildKey(versionKey)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "GetCollegePersonal conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// GetCollegePersonal 获取用户积分信息
func (d *dao) GetCollegePersonal(c context.Context, mid int64, version int) (res *college.Personal, err error) {
	var (
		key  = buildKey(version, midKey, mid)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "GetCollegePersonal conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// GetArchiveTabArchive 获取稿件信息
func (d *dao) GetArchiveTabArchive(c context.Context, collegeID int64, tabID int) (res []int64, err error) {
	var (
		key  = buildKey(tabKey, collegeID, tabID)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "GetArchiveTabArchive conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// DelCacheMidInviter 删除邀请人数统计
func (d *dao) DelCacheMidInviter(c context.Context, mid int64) (err error) {
	var (
		key  = buildKey(inviterCountKey, mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Errorc(c, "DelCacheMidInviter conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	return
}

// CacheMidInviter ...
func (d *dao) CacheMidInviter(c context.Context, mid int64) (list map[string]int64, err error) {
	list = make(map[string]int64)
	conn := d.redis.Get(c)
	defer conn.Close()
	key := buildKey(inviterCountKey, mid)
	args := redis.Args{}.Add(key)
	var tmp map[string]int64
	if tmp, err = redis.Int64Map(conn.Do("HGETALL", args...)); err != nil {
		log.Errorc(c, "CacheMidInviter redis.Ints(MGET) args(%v) error(%v)", args, err)
		return
	}
	if len(tmp) == 0 {
		return nil, redis.ErrNil
	}
	return tmp, nil
}

// AddCacheMidInviter ...
func (d *dao) AddCacheMidInviter(c context.Context, mid int64, list map[string]int64) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		conn = d.redis.Get(c)
		key  = buildKey(inviterCountKey, mid)
		args = redis.Args{}.Add(key)
	)
	for k, v := range list {
		args = args.Add(k).Add(v)
	}
	defer conn.Close()
	if err = conn.Send("HMSET", args...); err != nil {
		log.Errorc(c, "AddCacheMidInviter conn.Send(HMSET) error(%v)", err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.collegeMidCollegeExpire); err != nil {
		log.Errorc(c, "conn.Send(EXPIRE, %s) error(%v)", key, err)
		return
	}
	return
}

// MidFollow cache add mid coupon times
func (d *dao) MidFollow(c context.Context, mid int64) (err error) {
	var (
		key  = buildKey(followerKey, mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()

	if _, err = conn.Do("SET", key, 1); err != nil {
		log.Error("MidIsFollow conn.Send(SET, %s) error(%v)", key, err)
		return
	}
	return nil
}

// MidIsFollow 用户是否关注过
func (d *dao) MidIsFollow(c context.Context, mid int64) (res int, err error) {
	var (
		key  = buildKey(followerKey, mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "MidIsFollow conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	return
}
