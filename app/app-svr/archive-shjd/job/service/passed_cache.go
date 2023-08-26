package service

import (
	"context"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/time"

	"go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"
)

const (
	_prefixUpCnt = "uc_"
	_prefixUpPas = "up_"
)

func upCntKey(mid int64) string {
	return _prefixUpCnt + strconv.FormatInt(mid, 10)
}

func upPasKey(mid int64) string {
	return _prefixUpPas + strconv.FormatInt(mid, 10)
}

func (s *Service) delUpperPassedCache(c context.Context, aid, mid int64) (err error) {
	key := upPasKey(mid)
	conn := s.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("ZREM", key, aid); err != nil {
		return err
	}
	if err = s.setUpperCountCache(c, mid); err != nil {
		return err
	}
	return nil
}

func (s *Service) zcardUpperCnt(c context.Context, mid int64) (cnt int64, err error) {
	key := upPasKey(mid)
	conn := s.rds.Get(c)
	defer conn.Close()
	if cnt, err = redis.Int64(conn.Do("ZCARD", key)); err != nil {
		return 0, err
	}
	return cnt, nil
}

func (s *Service) setUpperCountCache(c context.Context, mid int64) (err error) {
	cnt, err := s.zcardUpperCnt(c, mid)
	if err != nil {
		return err
	}
	if cnt == 0 {
		cnt, err = s.dao.UpCount(c, mid)
		if err != nil {
			return err
		}
	}
	// nolint:gomnd
	expireTime := 86400 * 30
	if cnt == 0 {
		expireTime = 86400
	}
	key := upCntKey(mid)
	conn := s.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SETEX", key, expireTime, cnt); err != nil {
		return err
	}
	return nil
}

func (s *Service) addUpperPassed(c context.Context, aid int64) (err error) {
	arc, _, err := s.dao.Archive(c, aid)
	if err != nil {
		return errors.Wrap(err, "s.dao.Archive")
	}
	if arc == nil {
		log.Warn("addUpperPassed not exist aid(%d)", aid)
		return nil
	}
	if arc.AttrVal(api.AttrBitIsPUGVPay) == api.AttrYes || arc.AttrValV2(api.AttrBitV2NoPublic) == api.AttrYes {
		log.Warn("delUpperPassedCache aid(%d) mid(%d) isPUGV or NoPublic", aid, arc.Author.Mid)
		return s.delUpperPassedCache(c, aid, arc.Author.Mid)
	}
	ok, err := s.expireUpperPassedCache(c, arc.Author.Mid)
	if err != nil {
		return errors.Wrap(err, "s.expireUpperPassedCache")
	}
	if ok {
		if err = s.addUpperPassedCache(c, arc.Author.Mid, []int64{aid}, []time.Time{arc.PubDate}, []int64{int64(arc.Copyright)}); err != nil {
			return errors.Wrap(err, "s.addUpperPassedCache")
		}
		if err = s.setUpperCountCache(c, arc.Author.Mid); err != nil {
			return errors.Wrap(err, "s.setUpperCountCache")
		}
		return nil
	}
	alls, ptimes, copyrights, err := s.dao.RawUpperPassed(c, arc.Author.Mid)
	if err != nil {
		return errors.Wrap(err, "s.dao.UpperPassed")
	}
	if len(alls) == 0 {
		return nil
	}
	if err = s.addUpperPassedCache(c, arc.Author.Mid, alls, ptimes, copyrights); err != nil {
		return errors.Wrap(err, "s.addUpperPassedCache")
	}
	if err = s.setUpperCountCache(c, arc.Author.Mid); err != nil {
		return errors.Wrap(err, "s.setUpperCountCache")
	}
	return nil
}

func (s *Service) addUpperPassedCache(c context.Context, mid int64, aids []int64, pTimes []time.Time, copyrights []int64) (err error) {
	key := upPasKey(mid)
	conn := s.rds.Get(c)
	defer conn.Close()
	for k, aid := range aids {
		score := int64(pTimes[k]<<2) | copyrights[k]
		if err = conn.Send("ZADD", key, score, aid); err != nil {
			return err
		}
	}
	if err = conn.Flush(); err != nil {
		return err
	}
	for i := 0; i < len(aids); i++ {
		if _, err = conn.Receive(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) expireUpperPassedCache(c context.Context, mid int64) (ok bool, err error) {
	// nolint:gomnd
	expireTime := 86400 * 30
	conn := s.rds.Get(c)
	defer conn.Close()
	key := upPasKey(mid)
	if ok, err = redis.Bool(conn.Do("EXPIRE", key, expireTime)); err != nil {
		log.Error("conn.Do(EXPIRE, %s, %d) error(%+v)", key, expireTime, err)
		return false, err
	}
	return ok, nil
}
