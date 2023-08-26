package s10

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/stat/prom"
)

const (
	signed       = "s10:sg:%d%d%d"
	taskProgress = "s10:tg:%d%d%d"
)

func signedKey(mid int64) string {
	_, month, day := time.Now().Date()
	return fmt.Sprintf(signed, mid, int(month), day)
}

func (d *Dao) AddSignedCache(ctx context.Context, mid int64) (err error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	key := signedKey(mid)
	if err = conn.Send("SET", key, 1); err != nil {
		log.Errorc(ctx, "s10 d.dao.AddSignedCache Send (mid:%d) error(%v)", mid, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.signedExpire+(mid&63)); err != nil {
		log.Errorc(ctx, "s10 d.dao.AddSignedCache Send (mid:%d) error(%v)", mid, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "s10 d.dao.AddSignedCache Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "s10 d.dao.AddSignedCache Receive error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) SignedCache(ctx context.Context, mid int64) (int32, error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	res, err := redis.Int(conn.Do("GET", signedKey(mid)))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		prom.BusinessErrCount.Incr("s10:SignedCache")
		log.Errorc(ctx, "s10 d.dao.SignedCache(mid:%d) error(%v)", mid, err)
	}
	return int32(res), err
}

func taskProgressKey(uid int64) string {
	_, month, day := time.Now().Date()
	return fmt.Sprintf(taskProgress, uid, int(month), day)
}

func (d *Dao) AddTaskProgressCache(ctx context.Context, mid int64, res []int32) error {
	tmp, err := json.Marshal(res)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.AddTaskProgressCache(mid:%d,res:%+v) error(%v)", mid, res, err)
		return err
	}
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	key := taskProgressKey(mid)
	if err = conn.Send("SET", key, tmp); err != nil {
		log.Errorc(ctx, "s10 d.dao.AddTaskProgressCache Send (mid:%d) error(%v)", mid, err)
		return err
	}
	if err = conn.Send("EXPIRE", key, d.taskProgressExpire+(mid&63)); err != nil {
		log.Errorc(ctx, "s10 d.dao.AddTaskProgressCache Send (mid:%d) error(%v)", mid, err)
		return err
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "s10 d.dao.AddTaskProgressCache Flush error(%v)", err)
		return err
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "s10 d.dao.AddTaskProgressCache Receive error(%v)", err)
			return err
		}
	}
	return nil
}

func (d *Dao) TaskProgressCache(ctx context.Context, mid int64) ([]int32, error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	tmp, err := redis.Bytes(conn.Do("GET", taskProgressKey(mid)))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Errorc(ctx, "s10 d.dao.TaskProgressCache(mid:%d) error(%v)", mid, err)
		return nil, err
	}
	var res []int32
	if err = json.Unmarshal(tmp, &res); err != nil {
		log.Errorc(ctx, "s10 json.Unmarshal error(%v)", err)
	}
	return res, err
}
