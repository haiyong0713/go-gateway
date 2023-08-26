package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"

	model "go-gateway/app/web-svr/activity/interface/model/vogue"
)

const (
	_goodsList = "act_goods_list"
	_goods     = "act_goods_%d"
)

func keyGoodsList() string {
	return _goodsList
}

func keyGoods(id int64) string {
	return fmt.Sprintf(_goods, id)
}

func (d *Dao) CacheGoodsList(c context.Context) (res []*model.Goods, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := keyGoodsList()
	var data string
	if data, err = redis.String(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET) key:%v error:%v", key, err)
		return
	}
	if err = json.Unmarshal([]byte(data), &res); err != nil {
		return nil, err
	}
	return
}

func (d *Dao) AddCacheGoodsList(c context.Context, value []*model.Goods) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := keyGoodsList()
	var data []byte
	if data, err = json.Marshal(value); err != nil {
		log.Error("d.AddCacheConfig(%+v) error(%v)", key, err)
		return err
	}
	if _, err = conn.Do("SET", key, string(data), "EX", d.goodsExpire); err != nil {
		log.Error("d.AddCacheConfig(%+v) error(%v)", key, err)
	}
	return
}

func (d *Dao) DelCacheGoodsList(c context.Context) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := keyGoodsList()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelCacheGoodsList(%v) error(%v)", key, err)
		return
	}
	return
}

func (d *Dao) CacheGoods(c context.Context, id int64) (res *model.Goods, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := keyGoods(id)
	var data string
	if data, err = redis.String(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET) key:%v error:%v", key, err)
		return
	}
	if err = json.Unmarshal([]byte(data), &res); err != nil {
		return nil, err
	}
	return
}

func (d *Dao) AddCacheGoods(c context.Context, id int64, value *model.Goods) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := keyGoods(id)
	var data []byte
	if data, err = json.Marshal(value); err != nil {
		log.Error("d.AddCacheConfig(%+v) error(%v)", key, err)
		return err
	}
	if _, err = conn.Do("SET", key, string(data), "EX", d.goodsExpire); err != nil {
		log.Error("d.AddCacheConfig(%+v) error(%v)", key, err)
	}
	return
}

func (d *Dao) DelCacheGoods(c context.Context, id int64) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := keyGoods(id)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelCacheGoodsList(%v) error(%v)", key, err)
		return
	}
	return
}
