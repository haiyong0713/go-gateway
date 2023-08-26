package summer_camp

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/summer_camp"
)

// CacheGetCourseList get all course from cache.
func (d *dao) CacheGetCourseList(ctx context.Context) (res []*summer_camp.DBCourseCamp, err error) {
	var (
		key = buildKey(courseListKey)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redis.Do(ctx, "GET", key)); err != nil {
		log.Errorc(ctx, "CacheGetCourseList conn.Do(GET key(%v)) error(%v)", key, err)
		if err == redis.ErrNil {
			err = nil
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "CacheGetCourseList json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return

}

// CacheSetCourseList set all course into cache.
func (d *dao) CacheSetCourseList(ctx context.Context, data []*summer_camp.DBCourseCamp) (err error) {
	var (
		key = buildKey(courseListKey)
		bs  []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(ctx, "CacheSetCourseList json.Marshal(%v) error (%v)", data, err)
		return
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.courseListExpire, bs); err != nil {
		log.Errorc(ctx, "CacheSetCourseList conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.courseListExpire, string(bs), err)
	}
	return
}

// CacheGetUserCourseByID get user one course from cache.
func (d *dao) CacheGetUserCourseByID(ctx context.Context, mid int64, courseId int64) (res *summer_camp.DBUserCourse, err error) {
	keyStr := fmt.Sprintf(userCourseByIdKey, mid, courseId)
	var (
		key = buildKey(keyStr)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redis.Do(ctx, "GET", key)); err != nil {
		log.Errorc(ctx, "CacheGetUserCourseByID conn.Do(GET key(%v)) error(%v)", key, err)
		if err == redis.ErrNil {
			err = nil
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "CacheGetUserCourseByID json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return

}

// CacheSetUserCourseByID set user one course from cache.
func (d *dao) CacheSetUserCourseByID(ctx context.Context, mid int64, courseId int64, data *summer_camp.DBUserCourse) (err error) {
	keyStr := fmt.Sprintf(userCourseByIdKey, mid, courseId)
	var (
		key = buildKey(keyStr)
		bs  []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(ctx, "CacheSetUserCourseByID json.Marshal(%v) error (%v)", data, err)
		return
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.courseListExpire, bs); err != nil {
		log.Errorc(ctx, "CacheSetUserCourseByID conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.courseListExpire, string(bs), err)
	}
	return
}

// CacheDelUserCourseByID 删除单个
func (d *dao) CacheDelUserCourseByID(c context.Context, mid int64, courseId int64) (err error) {
	keyStr := fmt.Sprintf(userCourseByIdKey, mid, courseId)
	var (
		key = buildKey(keyStr)
	)
	if _, err := d.redis.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "CacheDelUserCourseByID conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return
}

// CacheGetUserCourseList get user all course from cache.
func (d *dao) CacheGetUserCourseList(ctx context.Context, mid int64) (res []*summer_camp.DBUserCourse, err error) {
	keyStr := fmt.Sprintf(userCourseKey, mid)
	var (
		key = buildKey(keyStr)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redis.Do(ctx, "GET", key)); err != nil {
		log.Errorc(ctx, "CacheGetUserCourseList conn.Do(GET key(%v)) error(%v)", key, err)
		if err == redis.ErrNil {
			err = nil
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "CacheGetUserCourseList json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return

}

// CacheSetUserCourseList set user all course into cache.
func (d *dao) CacheSetUserCourseList(ctx context.Context, mid int64, data []*summer_camp.DBUserCourse) (err error) {
	keyStr := fmt.Sprintf(userCourseKey, mid)
	var (
		key = buildKey(keyStr)
		bs  []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(ctx, "CacheSetUserCourseList json.Marshal(%v) error (%v)", data, err)
		return
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.courseListExpire, bs); err != nil {
		log.Errorc(ctx, "CacheSetUserCourseList conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.courseListExpire, string(bs), err)
	}
	return
}

// CacheDelUserCourseList 删除
func (d *dao) CacheDelUserCourseList(c context.Context, mid int64) (err error) {
	keyStr := fmt.Sprintf(userCourseKey, mid)
	var (
		key = buildKey(keyStr)
	)
	if _, err := d.redis.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "CacheDelUserCourseList conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return
}
