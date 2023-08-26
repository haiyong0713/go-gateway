package vogue

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/activity/ecode"
	l "go-gateway/app/web-svr/activity/interface/model/lottery"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

const (
	_follow   = 4
	_addTimes = "add"
	_fe       = 10
)

// 增加抽奖次数接口
func (s *Service) AddLotteryTimes(c context.Context, sid string, mid, cid int64, actionType, num int, orderNo string, isOut bool) (err error) {
	var (
		ip = metadata.String(c, metadata.RemoteIP)
	)
	lottery, lotteryTimesConf, err := s.fetchLotteryConfig(c, sid)
	if err != nil {
		return err
	}
	if err = checkTimesConf(lottery, lotteryTimesConf); err != nil {
		return
	}
	if isOut && (actionType > 4 && actionType != _fe) {
		return
	}
	err = s.JudgeAddTimes(c, actionType, num, lottery.ID, mid, cid, lotteryTimesConf, orderNo, ip)
	return
}

func rsKey(actionType int, cid int64) string {
	return strconv.Itoa(actionType) + strconv.FormatInt(cid, 10)
}

// 增加次数判断
func (s *Service) JudgeAddTimes(c context.Context, actionType, num int, sid int64, mid, id int64, lotteryTimesConf []*l.LotteryTimesConfig, orderNo, ip string) (err error) {
	var (
		lotteryMap = make(map[string]*l.LotteryTimesConfig, len(lotteryTimesConf))
		times      = make(map[string]int)
		cid        int64
	)
	for _, l := range lotteryTimesConf {
		if l.Type == actionType {
			cid = l.ID
		}
		lotteryMap[rsKey(l.Type, l.ID)] = l
	}
	if actionType > _follow && id > 0 {
		cid = id
	}
	key := rsKey(actionType, cid)
	// 未配置或配置行为增加次数为0
	if lotteryMap[key] == nil || lotteryMap[key].Times == 0 {
		return
	}
	if times, err = s.lottDao.LotteryAddTimes(c, lotteryTimesConf, sid, mid); err != nil {
		log.Error("JudgeAddTimes: s.dao.LotteryAddTimes sid(%d) mid(%d) error(%v)", sid, mid, err)
		return
	}
	// 增加次数达到上限
	if times[key] >= lotteryMap[key].Most {
		err = ecode.ActivityLotteryAddTimesLimit
		return
	}
	incr := lotteryMap[key].Times
	if num > 0 {
		incr = num
	}
	// 增加次数达上限
	if times[key]+incr > lotteryMap[key].Most {
		incr = lotteryMap[key].Most - times[key]
	}
	if _, err = s.lottDao.InsertLotteryAddTimes(c, sid, mid, actionType, incr, cid, ip, orderNo); err != nil {
		log.Error("JudgeAddTimes: s.dao.InsertLotteryAddTimes sid(%d) mid(%d) type(%d) val(%d) ip(%s) orderNo(%s) error(%v)", sid, mid, actionType, incr, ip, orderNo, err)
		return
	}
	if _, err = s.lottDao.IncrTimes(c, sid, mid, lotteryMap[key], incr, _addTimes); err != nil {
		log.Error("JudgeAddTimes: s.dao.IncrTimes sid(%d) mid(%d) val(%d) type(%s) error(%v)", sid, mid, incr, _addTimes, err)
	}
	return
}

func checkTimesConf(lottery *l.Lottery, timesConf []*l.LotteryTimesConfig) error {
	now := time.Now().Unix()
	if lottery.ID == 0 {
		return ecode.ActivityNotExist
	}
	if lottery.Stime.Time().Unix() > now {
		return ecode.ActivityNotStart
	}
	if lottery.Etime.Time().Unix() < now {
		return ecode.ActivityOverEnd
	}
	if len(timesConf) == 0 {
		return ecode.ActivityNotConfig
	}
	return nil
}

func (s *Service) fetchLotteryConfig(ctx context.Context, sid string) (*l.Lottery, []*l.LotteryTimesConfig, error) {
	var (
		lottery          *l.Lottery
		lotteryTimesConf []*l.LotteryTimesConfig
	)
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if lottery, err = s.lottDao.Lottery(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if lotteryTimesConf, err = s.lottDao.LotteryTimesConfig(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryTimesConfig %s", sid)
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Error("Failed to fetch lottery: %+v", err)
		return nil, nil, err
	}
	return lottery, lotteryTimesConf, nil
}
