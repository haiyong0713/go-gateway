package mark

import (
	"context"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"github.com/pkg/errors"
)

// mark代表用户对剧情树的评分, evaluation代表稿件的总和评分（aid维度）
const (
	_postFixMark       = "mk_"
	_postfixEvaluation = "_ev"
)

func markKey(aid, mid int64) string {
	return strconv.FormatInt(mid, 10) + "_" + _postFixMark + strconv.FormatInt(aid, 10)
}

func evaluationKey(aid int64) string {
	return strconv.FormatInt(aid, 10) + _postfixEvaluation
}

// CacheMark .
func (d *Dao) CacheMark(c context.Context, aid int64, mid int64) (res int64, err error) {
	var (
		key  = markKey(aid, mid)
		conn = d.rds.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	return
}

// AddCacheKMark .
func (d *Dao) AddCacheMark(c context.Context, aid int64, mid int64, val int64) (err error) {
	key := markKey(aid, mid)
	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, val); err != nil {
		log.Error("conn.Do(SETEX, %s, %+v) error(%v)", key, val, err)
	}
	return
}

// CacheMark .
func (d *Dao) CacheEvaluation(c context.Context, aid int64) (res int64, err error) {
	var (
		key  = evaluationKey(aid)
		conn = d.rds.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	return
}

// AddCacheKMark .
func (d *Dao) AddCacheEvaluation(c context.Context, aid int64, val int64) (err error) {
	key := evaluationKey(aid)
	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, val); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, val, err)
		return
	}
	return
}

func (d *Dao) CacheEvaluations(c context.Context, IDs []int64) (res map[int64]int64, err error) {
	var (
		conn  = d.rds.Get(c)
		args  = redis.Args{}
		items []int64
	)
	defer conn.Close()
	for _, aid := range IDs {
		args = args.Add(evaluationKey(aid))
	}
	if items, err = redis.Int64s(conn.Do("MGET", args...)); err != nil {
		log.Error("MGet conn.Do(mget) error(%v) args(%+v)", err, args)
		return
	}
	res = make(map[int64]int64, len(IDs))
	for k, item := range items {
		if item == 0 { // omg
			continue
		}
		res[IDs[k]] = item
	}
	return
}

func (d *Dao) AddCacheEvaluations(c context.Context, evaluations map[int64]int64) (err error) {
	var (
		key         string
		keys        []string
		argsRecords = redis.Args{}
		conn        = d.rds.Get(c)
	)
	defer conn.Close()
	for k, evaluation := range evaluations {
		if evaluation == 0 { // ignore
			continue
		}
		key = evaluationKey(k)
		keys = append(keys, key)
		argsRecords = argsRecords.Add(key).Add(evaluation)
	}
	if _, err = conn.Do("MSET", argsRecords...); err != nil {
		err = errors.Wrapf(err, "conn.Do(MSET) keys:%+v error", keys)
	}
	return
}

func (d *Dao) AddCacheMarks(c context.Context, marks map[int64]int64, mid int64) (err error) {
	var (
		key         string
		keys        []string
		argsRecords = redis.Args{}
		conn        = d.rds.Get(c)
	)
	defer conn.Close()
	for k, mark := range marks {
		if mark == 0 { // ignore
			continue
		}
		key = markKey(k, mid)
		keys = append(keys, key)
		argsRecords = argsRecords.Add(key).Add(mark)
	}
	if _, err = conn.Do("MSET", argsRecords...); err != nil {
		err = errors.Wrapf(err, "conn.Do(MSET) keys:%+v error", keys)
	}
	return
}

func (d *Dao) CacheMarks(c context.Context, IDs []int64, mid int64) (res map[int64]int64, err error) {
	var (
		conn  = d.rds.Get(c)
		args  = redis.Args{}
		items []int64
	)
	defer conn.Close()
	for _, aid := range IDs {
		args = args.Add(markKey(aid, mid))
	}
	if items, err = redis.Int64s(conn.Do("MGET", args...)); err != nil {
		log.Error("MGet conn.Do(mget) error(%v) args(%+v)", err, args)
		return
	}
	res = make(map[int64]int64, len(IDs))
	for k, item := range items {
		if item == 0 { // omg
			continue
		}
		res[IDs[k]] = item
	}
	return

}
