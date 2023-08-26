package s10

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
)

const userFlow = "s10:sf:%d"

func userFlowKey(mid int64) string {
	return fmt.Sprintf(userFlow, mid)
}

// value:1- 哨兵；2-联调；3-移动
func (d *Dao) AddUserFlowCache(ctx context.Context, mid int64, value int32) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	ufkey := userFlowKey(mid)

	err := conn.Send("SET", ufkey, value)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Send(key:%s) error:%v", ufkey, err)
		return err
	}
	err = conn.Send("EXPIRE", ufkey, d.userFlowExpire)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Send(key:%s) error:%v", ufkey, err)
		return err
	}
	err = conn.Flush()
	if err != nil {
		log.Errorc(ctx, "s10 conn.Flush(key:%s) error:%v", ufkey, err)
		return err
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "s10 conn.Receive(key:%s) error:%v", ufkey, err)
			return err
		}
	}
	return nil
}

func (d *Dao) UserFlowCache(ctx context.Context, mid int64) (int32, error) {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	ufkey := userFlowKey(mid)
	res, err := redis.Int64(conn.Do("GET", ufkey))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", ufkey, err)
	}
	return int32(res), err
}

func (d *Dao) DelUserFlow(ctx context.Context, mid int64) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	ufkey := userFlowKey(mid)
	_, err := conn.Do("DEL", ufkey)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", ufkey, err)
		return err
	}
	return nil
}
