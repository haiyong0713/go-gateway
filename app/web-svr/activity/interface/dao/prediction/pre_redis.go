package prediction

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"github.com/pkg/errors"
)

// preList prediction list key.
func preListKey(sid int64) string {
	return fmt.Sprintf("pre_l_s_%d", sid)
}

func itemListKey(pid int64) string {
	return fmt.Sprintf("pre_it_l_i_%d", pid)
}

// PreList .
func (d *Dao) PreList(c context.Context, sid int64) (list []int64, err error) {
	var (
		key  = preListKey(sid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if list, err = redis.Int64s(conn.Do("SMEMBERS", key)); err != nil {
		if err == redis.ErrNil {
			log.Error("PreList:data need to reload sid:%d", sid)
			err = nil
		} else {
			err = errors.Wrap(err, "redis.Int64s(SMEMBERS)")
		}
	}
	return
}

// ItemRandMember .
func (d *Dao) ItemRandMember(c context.Context, pid int64, count int) (ids []int64, err error) {
	var (
		key  = itemListKey(pid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if ids, err = redis.Int64s(conn.Do("SRANDMEMBER", key, count)); err != nil {
		if err == redis.ErrNil {
			log.Error("ItemRandMember:data need to reload pid(%d)", pid)
			err = nil
		} else {
			err = errors.Wrap(err, "redis.Int64s(SRANDMEMBER)")
		}
	}
	return
}

// AddPreSet .
func (d *Dao) AddPreSet(c context.Context, ids []int64, sid int64) (err error) {
	var (
		key  = preListKey(sid)
		conn = d.redis.Get(c)
		args = make([]interface{}, 0, len(ids)+1)
	)
	defer conn.Close()
	if len(ids) == 0 {
		return
	}
	args = append(args, key)
	for _, v := range ids {
		args = append(args, v)
	}
	if _, err = conn.Do("SADD", args...); err != nil {
		log.Error("AddPreSet:conn.Do(SADD) error(%v)", err)
	}
	return
}

// DelPreSet .
func (d *Dao) DelPreSet(c context.Context, ids []int64, sid int64) (err error) {
	var (
		key  = preListKey(sid)
		conn = d.redis.Get(c)
		args = make([]interface{}, 0, len(ids)+1)
	)
	defer conn.Close()
	if len(ids) == 0 {
		return
	}
	args = append(args, key)
	for _, v := range ids {
		args = append(args, v)
	}
	if _, err = conn.Do("SREM", args...); err != nil {
		log.Error("DelPreSet:conn.Do(SREM) error(%v)", err)
	}
	return
}

// DelItemPreSet .
func (d *Dao) DelItemPreSet(c context.Context, ids []int64, pid int64) (err error) {
	var (
		key  = itemListKey(pid)
		conn = d.redis.Get(c)
		args = make([]interface{}, 0, len(ids)+1)
	)
	defer conn.Close()
	if len(ids) == 0 {
		return
	}
	args = append(args, key)
	for _, v := range ids {
		args = append(args, v)
	}
	if _, err = conn.Do("SREM", args...); err != nil {
		log.Error("DelItemPreSet:conn.Do(SREM) error(%v)", err)
	}
	return
}

// AddItemPreSet .
func (d *Dao) AddItemPreSet(c context.Context, ids []int64, pid int64) (err error) {
	var (
		key  = itemListKey(pid)
		conn = d.redis.Get(c)
		args = make([]interface{}, 0, len(ids)+1)
	)
	defer conn.Close()
	if len(ids) == 0 {
		return
	}
	args = append(args, key)
	for _, v := range ids {
		args = append(args, v)
	}
	if _, err = conn.Do("SADD", args...); err != nil {
		log.Error("AddItemPreSet:conn.Do(SADD) error(%v)", err)
	}
	return
}
