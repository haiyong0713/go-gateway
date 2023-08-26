package funny

import (
	"context"
	"errors"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"strconv"
)

// SetTask1Data 写redis
func (d *dao) SetTask1Data(c context.Context, num int) (err error) {
	var (
		key = buildKey("task1")
	)
	conn := d.redis.Get(c)

	defer conn.Close()
	if _, err = redis.String(conn.Do("SET", key, strconv.Itoa(num))); err != nil {
		if err != redis.ErrNil {
			return errors.New(fmt.Sprintf("set SetTask1Num conn.Do(SET %s %v) error(%v)", key, num, err))
		}
	}
	log.Infoc(c, "set SetTask1Num Succ %v %v", key, num)

	return nil
}

// SetTask2Data 写redis
func (d *dao) SetTask2Data(c context.Context, num int) (err error) {
	var (
		key = buildKey("task2")
	)
	conn := d.redis.Get(c)

	defer conn.Close()
	if _, err = redis.String(conn.Do("SET", key, strconv.Itoa(num))); err != nil {
		if err != redis.ErrNil {
			return errors.New(fmt.Sprintf("set SetTask2Num conn.Do(SET %s %v) error(%v)", key, num, err))
		}
	}

	log.Infoc(c, "set SetTask2Num Succ %v %v", key, num)

	return nil
}
