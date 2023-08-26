package dao

import (
	"context"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/ugc-season/service/api"
)

const (
	_prefixSeason     = "si_" // season info
	_prefixSeasonStat = "ss_" // season stat
	_prefixView       = "v_"  // view cache
)

func seasonKey(sid int64) string {
	return _prefixSeason + strconv.FormatInt(sid, 10)
}

func seasonStatKey(sid int64) string {
	return _prefixSeasonStat + strconv.FormatInt(sid, 10)
}

func viewKey(sid int64) string {
	return _prefixView + strconv.FormatInt(sid, 10)
}

// AddSeasonCache set season into cache.
func (d *Dao) AddSeasonCache(c context.Context, s *api.Season) error {
	bs, err := s.Marshal()
	if err != nil {
		log.Error("season.Marshal error(%+v)", err)
		return err
	}
	key := seasonKey(s.ID)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, bs); err != nil {
		log.Error("conn.Do(SET, %s) error(%+v)", key, err)
		return err
	}
	return nil
}

// AddSeasonsCache is
func (d *Dao) AddSeasonsCache(c context.Context, seasons map[int64]*api.Season) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	for _, s := range seasons {
		bs, err := s.Marshal()
		if err != nil {
			log.Error("season.Marshal error(%+v)", err)
			continue
		}
		if err := conn.Send("SET", seasonKey(s.ID), bs); err != nil {
			log.Error("conn.Send(SET, %s) error(%+v)", seasonKey(s.ID), err)
			return err
		}
	}
	if err := conn.Flush(); err != nil {
		log.Error("conn.Flush error(%+v)", err)
		return err
	}
	return nil
}

// AddViewCache set view into cache
func (d *Dao) AddViewCache(c context.Context, sid int64, bs []byte) error {
	key := viewKey(sid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("SET", key, bs); err != nil {
		log.Error("conn.Do(SET, %s) error(%+v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) AddViewCaches(c context.Context, v []*api.View) error {
	var (
		keys        []string
		argsRecords = redis.Args{}
		conn        = d.redis.Get(c)
	)
	defer conn.Close()
	for _, v := range v {
		bs, err := v.Marshal()
		if err != nil {
			log.Error("view.Marshal error(%+v)", err)
			continue
		}
		key := viewKey(v.Season.ID)
		keys = append(keys, key)
		argsRecords = argsRecords.Add(key).Add(bs)
	}
	if _, err := conn.Do("MSET", argsRecords...); err != nil {
		log.Error("conn.Do(MSET) keys(%+v) err(%+v)", keys, err)
		return err
	}
	return nil
}

// DelSeasonCache is
func (d *Dao) DelSeasonCache(c context.Context, sid int64) error {
	key := seasonKey(sid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("conn.Do(DEL, %s) error(%+v)", key, err)
		return err
	}
	return nil
}

// DelViewCache is
func (d *Dao) DelViewCache(c context.Context, sid int64) error {
	key := viewKey(sid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("conn.Do(DEL, %s) error(%+v)", key, err)
		return err
	}
	return nil
}
