package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/job/model"
)

const (
	_keyMatchOne    = "match_one"
	_keyBattleTwo   = "ba_list_%s"
	_keyBattleThree = "ba_info_%s"
)

func keyBattleTwo(matchID string) string {
	return fmt.Sprintf(_keyBattleTwo, matchID)
}

func keyBattleThree(battleString string) string {
	return fmt.Sprintf(_keyBattleThree, battleString)
}

func (d *Dao) AddCacheMatch(c context.Context, data []*model.ScoreMatch) (err error) {
	var (
		bs   []byte
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SETEX", _keyMatchOne, d.scoreLiveExpire, bs); err != nil {
		log.Error("conn.Do(SETEX, %s, %d, %d)", _keyMatchOne, d.scoreLiveExpire, bs)
	}
	return
}

func (d *Dao) MatchOne(c context.Context) (res []*model.ScoreMatch, err error) {
	var (
		bs   []byte
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", _keyMatchOne)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "conn.Do(GET,%s) error(%v)", _keyMatchOne, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// DelCacheMatch
func (d *Dao) DelCacheMatch(c context.Context) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", _keyMatchOne); err != nil {
		log.Error("DelCacheMatch conn.Do(DEL key(%s) error(%v))", _keyMatchOne, err)
	}
	return
}

func (d *Dao) AddCacheBattleList(c context.Context, matchID string, data *model.BattleList) (err error) {
	var (
		bs   []byte
		conn = d.redis.Get(c)
		key  = keyBattleTwo(matchID)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.scoreLiveExpire, bs); err != nil {
		log.Error("conn.Do(SETEX, %s, %d, %d)", key, d.scoreLiveExpire, bs)
	}
	return
}

func (d *Dao) AddCacheBattleInfo(c context.Context, battleString string, data *model.BattleInfo) (err error) {
	var (
		bs   []byte
		conn = d.redis.Get(c)
		key  = keyBattleThree(battleString)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.scoreLiveExpire, bs); err != nil {
		log.Error("conn.Do(SETEX, %s, %d, %d)", key, d.scoreLiveExpire, bs)
	}
	return
}
