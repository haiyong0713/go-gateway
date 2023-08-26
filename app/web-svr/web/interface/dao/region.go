package dao

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/web/interface/model"

	"github.com/pkg/errors"
)

func regionListKey() string {
	return "region_list"
}

func (d *Dao) RegionList(ctx context.Context) (map[string][]*model.Region, error) {
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	key := regionListKey()
	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "GET,%v", key)
	}
	var res map[string][]*model.Region
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}
