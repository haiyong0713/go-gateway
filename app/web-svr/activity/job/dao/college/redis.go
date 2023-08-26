package college

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/job/model/college"

	"github.com/pkg/errors"
)

// SetMidPersonal 设置用户维度的排行榜结果
func (d *dao) SetMidPersonal(c context.Context, midRank []*college.Personal, version int) (err error) {
	if len(midRank) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	keys := make([]string, 0)
	for _, v := range midRank {
		var bs []byte
		if bs, err = json.Marshal(v); err != nil {
			log.Errorc(c, "SetMidPersonal json.Marshal() error(%v)", err)
			return
		}
		// 删除之前前4个版本的记录
		if version-4 > 0 {
			keys = append(keys, buildKey(version-4, midKey, v.MID))

		}
		args = args.Add(buildKey(version, midKey, v.MID)).Add(bs)
	}
	if _, err = redis.String(conn.Do("MSET", args...)); err != nil {
		err = errors.Wrap(err, "SetMidPersonal conn.Do(MSET)")
		time.Sleep(time.Second)
		if _, err = redis.String(conn.Do("MSET", args...)); err != nil {
			err = errors.Wrap(err, "SetMidPersonal conn.Do(MSET)")
			return
		}
	}
	if len(keys) > 0 {
		for _, v := range keys {
			if _, err = conn.Do("DEL", v); err != nil {
				time.Sleep(time.Second)
				log.Errorc(c, "del SetMidPersonal history version  conn.Do(DEL)")
				err = nil
			}
		}
	}
	return
}

// SetCollegeDetail set pool ids cache.
func (d *dao) SetCollegeDetail(c context.Context, college *college.Detail, version int) (err error) {
	if college == nil {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}.Add(buildKey(collegeKey, version, college.ID))
	args = args.Add(idKey).Add(college.ID)
	args = args.Add(scoreKey).Add(college.Score)
	args = args.Add(nationwideRankKey).Add(college.NationwideRank)
	args = args.Add(provinceKey).Add(college.Province)
	args = args.Add(provinceRankKey).Add(college.ProvinceRank)
	args = args.Add(tabListKey).Add(xstr.JoinInts(college.TabList))
	args = args.Add(relationKey).Add(xstr.JoinInts(college.RelationMid))
	args = args.Add(nameKey).Add(college.Name)
	// 删除之前前4个版本的记录

	if err = conn.Send("HMSET", args...); err != nil {
		log.Errorc(c, "SetCollgeDetail conn.Do(HMSET %s) error(%v)", buildKey(collegeKey, college.ID), err)
		return
	}
	if version-4 > 0 {
		if _, err = conn.Do("DEL", buildKey(collegeKey, version-4, college.ID)); err != nil {
			log.Errorc(c, "del SetCollegeDetail history version  conn.Do(DEL)")
			err = nil
		}
	}
	return
}

// SetArchiveTabArchive 本次参与活动的mid
func (d *dao) SetArchiveTabArchive(c context.Context, collegeID int64, tabID int64, aids []int64) (err error) {
	var (
		bs   []byte
		key  = buildKey(tabKey, collegeID, tabID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(aids); err != nil {
		log.Errorc(c, "json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SET", key, bs); err != nil {
		log.Errorc(c, "SetArchiveTabArchive conn.Do(SET) key(%s) error(%v)", key, err)
	}
	return
}

// GetCollegeUpdateVersion 获得本次计算的版本
func (d *dao) GetCollegeUpdateVersion(c context.Context) (res *college.Version, err error) {
	res = &college.Version{}
	var (
		key  = buildKey(versionUpdateKey)
		bs   []byte
		conn = d.redis.Get(c)
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

// SetCollegeVersion 获得本次计算的版本
func (d *dao) SetCollegeVersion(c context.Context, version *college.Version) (err error) {
	var (
		key  = buildKey(versionKey)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	var bs []byte
	if bs, err = json.Marshal(version); err != nil {
		log.Error("AddMidAward json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SET", key, bs); err != nil {
		log.Errorc(c, "SetCollegeVersion conn.Do(SET) key(%s) error(%v)", key, err)
	}
	return
}

// SetCollegeUpdateVersion 获得本次计算的版本
func (d *dao) SetCollegeUpdateVersion(c context.Context, version *college.Version) (err error) {
	var (
		key  = buildKey(versionUpdateKey)
		conn = d.redis.Get(c)
	)
	defer conn.Close()

	var bs []byte
	if bs, err = json.Marshal(version); err != nil {
		log.Error("AddMidAward json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SET", key, bs); err != nil {
		log.Errorc(c, "SetCollegeVersion conn.Do(SET) key(%s) error(%v)", key, err)
	}
	return
}
