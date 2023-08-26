package anti_crawler

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	model "go-gateway/app/app-svr/app-feed/admin/model/anti_crawler"

	"github.com/pkg/errors"
)

// TryLock ...
func (d *Dao) TryLock(ctx context.Context, key string, timeout int32) (bool, error) {
	reply, err := redis.String(d.redis.Do(ctx, "SET", key, 1, "EX", timeout, "NX"))
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		return false, err
	}
	if reply == "OK" {
		return true, nil
	}
	return false, nil
}

// UnLock ...
func (d *Dao) UnLock(ctx context.Context, key string) error {
	_, err := d.redis.Do(ctx, "DEL", key)
	return err
}

func wListKey() string {
	return "w_list"
}

func (d *Dao) SetWListCache(ctx context.Context, data ...*model.WList) error {
	if len(data) == 0 {
		return nil
	}
	args := redis.Args{}.Add(wListKey())
	for _, v := range data {
		bs, err := json.Marshal(v)
		if err != nil {
			return err
		}
		args = args.Add(v.Value).Add(bs)
	}
	if _, err := d.redis.Do(ctx, "HMSET", args...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (d *Dao) DelWListCache(ctx context.Context, value string) error {
	key := wListKey()
	if _, err := d.redis.Do(ctx, "HDEL", key, value); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (d *Dao) WListAllCache(ctx context.Context) ([]*model.WList, error) {
	key := wListKey()
	bss, err := redis.ByteSlices(d.redis.Do(ctx, "HGETALL", key))
	if err != nil {
		return nil, err
	}
	var data []*model.WList
	for i := 1; i <= len(bss); i += 2 {
		var v *model.WList
		if err := json.Unmarshal(bss[i], &v); err != nil {
			return nil, err
		}
		data = append(data, v)
	}
	return data, nil
}

func (d *Dao) WListCache(ctx context.Context, value string) (*model.WList, error) {
	key := wListKey()
	bs, err := redis.Bytes(d.redis.Do(ctx, "HGET", key, value))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	var v *model.WList
	if err := json.Unmarshal(bs, &v); err != nil {
		return nil, err
	}
	return v, nil
}
