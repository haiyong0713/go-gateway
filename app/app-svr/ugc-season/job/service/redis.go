package service

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/time"
	"go-gateway/app/app-svr/ugc-season/service/api"
)

const (
	_prefixUpper    = "up_%d"
	_prefixUpperCnt = "uc_%d"
	_upperCacheTime = 86400 * 7
)

func upKey(mid int64) string {
	return fmt.Sprintf(_prefixUpper, mid)
}

func upperNoSeasonKey(mid int64) string {
	return fmt.Sprintf(_prefixUpperCnt, mid)
}

// AddUpperSeason is
func (s *Service) AddUpperSeason(c context.Context, sid, mid int64, maxPtime time.Time) (err error) {
	conn := s.redis.Get(c)
	defer conn.Close()
	var (
		exist bool
	)
	if exist, err = redis.Bool(conn.Do("EXPIRE", upKey(mid), _upperCacheTime)); err != nil {
		log.Error("redis.Bool(EXPIRE, %d) error(%v)", mid, err)
		return
	}
	if _, err = conn.Do("DEL", upperNoSeasonKey(mid)); err != nil {
		log.Error("conn.Do(DEL, %d) error(%v)", mid, err)
		return
	}
	if !exist {
		// list不存在，调用service list方法初始化
		log.Warn("upper(%d) season list cache not exist", mid)
		if _, err = s.seasonClient.UpperList(c, &api.UpperListRequest{Mid: mid, PageNum: 1, PageSize: 1}); err != nil {
			log.Error("season.UpperList(%d) error(%v)", mid, err)
			//season.show=1,season.state=-6的情况下，up下还查不到投稿信息，404错误先忽略不重试
			if ecode.EqualError(ecode.NothingFound, err) {
				err = nil
			}
			return
		}
		return
	}
	if _, err = conn.Do("ZADD", upKey(mid), maxPtime, sid); err != nil {
		log.Error("conn.Do(ZADD, %s, %d, %d) error(%v)", upKey(mid), maxPtime, sid, err)
		return
	}
	return
}

// DelUpperSeason is
func (s *Service) DelUpperSeason(c context.Context, sid, mid int64) (err error) {
	conn := s.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("ZREM", upKey(mid), sid); err != nil {
		log.Error("conn.Do(ZREM, %d, %d) error(%v)", mid, sid, err)
		return
	}
	return
}
