package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/model"

	"github.com/pkg/errors"
)

func NewRedis() (r *redis.Redis, cf func(), err error) {
	var (
		cfg redis.Config
		ct  paladin.Map
	)
	if err = paladin.Get("redis.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	r = redis.NewRedis(&cfg)
	cf = func() { r.Close() }
	return
}

func (d *dao) PingRedis(ctx context.Context) (err error) {
	if _, err = d.redis.Do(ctx, "SET", "ping", "pong"); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
	}
	return
}

func keyTopPhotoArc(mid int64) string {
	return fmt.Sprintf("%d_topphoto_arc", mid)
}

func (d *dao) cacheSFTopPhotoArc(mid int64) string {
	return keyTopPhotoArc(mid)
}

func (d *dao) CacheTopPhotoArc(ctx context.Context, mid int64) (*model.TopPhotoArc, error) {
	key := keyTopPhotoArc(mid)
	bs, err := redis.Bytes(d.redis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "CacheTopPhotoArc key:%s", key)
	}
	res := &model.TopPhotoArc{}
	if err = res.Unmarshal(bs); err != nil {
		return nil, errors.Wrap(err, "CacheTopPhotoArc Unmarshal")
	}
	return res, nil
}

func (d *dao) AddCacheTopPhotoArc(ctx context.Context, id int64, val *model.TopPhotoArc) error {
	if val == nil {
		return nil
	}
	key := keyTopPhotoArc(id)
	bs, err := val.Marshal()
	if err != nil {
		log.Errorc(ctx, "d.AddCacheTopPhotoArc(get key: %v) err: %+v", key, err)
		return err
	}
	if _, err = d.redis.Do(ctx, "set", key, bs, "EX", d.topPhotoArcExpire); err != nil {
		log.Errorc(ctx, "d.AddCacheTopPhotoArc(get key: %v) err: %+v", key, err)
		return err
	}
	return nil
}

func (d *dao) DelCacheTopPhotoArc(ctx context.Context, mid int64) (err error) {
	key := keyTopPhotoArc(mid)
	if _, err = d.redis.Do(ctx, "DEL", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(ctx, "d.DelCacheTopPhotoArc(key: %v) err: %+v", key, err)
		return err
	}
	return nil
}

func (d *dao) keyPrivacySetting(mid int64) string {
	return fmt.Sprintf("prvyset_%d", mid)
}

func (d *dao) privacyKey(mid int64) string {
	return fmt.Sprintf("spc_pcy_%d", mid)
}

// TODO 等全部切换后，可以删除
func (d *dao) DelCachePrivacySetting(ctx context.Context, mid int64) error {
	key := d.keyPrivacySetting(mid)
	oldKey := d.privacyKey(mid)
	_, err := d.redis.Do(ctx, "DEL", key, oldKey)
	return err
}

func (d *dao) keyLivePlaybackWhitelist() string {
	return "live_back_list"
}

func (d *dao) setCacheLivePlaybackWhitelist(ctx context.Context, data map[int64]struct{}) error {
	key := d.keyLivePlaybackWhitelist()
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = d.redis.Do(ctx, "SET", key, bs)
	return err
}
