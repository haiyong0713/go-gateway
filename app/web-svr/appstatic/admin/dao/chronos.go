package dao

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/web-svr/appstatic/admin/model"
)

const (
	_rawRulesKey    = "chronos_key"
	_playerRulesKey = "chronos_player"
)

// SaveRawRules 将后台提交的rules存到缓存，以备list
func (d *Dao) SaveRawRules(c context.Context, rules []*model.ChronosRule) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	str, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	if _, err = conn.Do("SET", _rawRulesKey, str); err != nil {
		return err
	}
	return nil
}

// RawRules 读取自备缓存，用于list
func (d *Dao) RawRules(c context.Context) ([]*model.ChronosRule, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	res, err := redis.Bytes(conn.Do("GET", _rawRulesKey))
	if err != nil {
		if err == redis.ErrNil {
			return make([]*model.ChronosRule, 0), nil
		}
		return nil, err
	}
	rules := make([]*model.ChronosRule, 0)
	if err = json.Unmarshal(res, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

// SavePlayerRules 将后台的rules+md5保存到缓存中，供app-player使用
func (d *Dao) SavePlayerRules(c context.Context, rules []*model.PlayerRule) error {
	str, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	eg := errgroup.WithCancel(c)
	for _, v := range d.playerRedis { // 并发保存
		red := v
		eg.Go(func(c context.Context) error {
			conn := red.Get(c)
			defer conn.Close()
			if _, err = conn.Do("SET", _playerRulesKey, str); err != nil {
				return err
			}
			return nil
		})
	}
	return eg.Wait()
}
