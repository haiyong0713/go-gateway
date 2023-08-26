package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/interface/model"
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

func (d *Dao) CacheBattleList(c context.Context, matchID string) (res *model.BattleList, err error) {
	var (
		bs   []byte
		conn = d.redis.Get(c)
		key  = keyBattleTwo(matchID)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

func (d *Dao) CacheBattleInfo(c context.Context, battleString string) (res *model.BattleInfo, err error) {
	var (
		bs   []byte
		conn = d.redis.Get(c)
		key  = keyBattleThree(battleString)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}
