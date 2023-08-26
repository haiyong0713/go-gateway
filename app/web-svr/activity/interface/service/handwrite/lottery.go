package handwrite

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"time"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	mdlhw "go-gateway/app/web-svr/activity/interface/model/handwrite"

	actPlat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"github.com/pkg/errors"
)

// AddLotteryTimes 增加抽奖次数
func (s *Service) AddLotteryTimes(c context.Context, mid int64) (*mdlhw.AddTimesReply, error) {
	// 活动时间判断
	if err := s.actTime(); err != nil {
		return nil, err
	}
	// 锁
	day := time.Now().Format("20060102")
	if err := s.handwrite.AddTimeLock(c, mid); err != nil {
		log.Error("s.handwrite.AddTimeLock(%d) error(%v)", mid, err)
		return nil, ecode.ActivityWriteHandAddtimesTooFastErr
	}
	// 账号信息验证
	if err := s.checkAccountInfo(c, mid); err != nil {
		return nil, err
	}
	// 获取当前金币数
	coin, err := s.getMidCoin(c, mid)
	if err != nil {
		return nil, ecode.ActivityWriteHandGetCoinErr
	}
	if coin < s.c.HandWrite.AwardCoinLimit {
		return nil, ecode.ActivityWriteHandCoinNotEnoughErr
	}
	orderNo := s.getOrderNo(c, mid)
	// 增加获奖次数
	if err = s.like.AddLotteryTimes(c, s.c.HandWrite.Sid, mid, 0, 7, 0, fmt.Sprint(orderNo), false); err != nil {
		log.Error("s.dao.Addtimes(%v)", err)
		return nil, err
	}
	if err := s.handwrite.AddTimesRecord(c, mid, day); err != nil {
		log.Error("s.handwrite.AddTimesRecord(%d) error(%v)", mid, err)
	}
	return nil, nil
}

func (s *Service) getOrderNo(c context.Context, mid int64) string {
	return fmt.Sprintf("%d_%v_%v", mid, "handWrite", time.Now().Format("20060102"))
}
func (s *Service) actTime() error {
	now := time.Now().Unix()
	if now < s.c.HandWrite.ActStart {
		return ecode.ActivityNotStart
	}
	if s.c.HandWrite.ActEnd < now {
		return ecode.ActivityOverEnd
	}
	return nil
}

func (s *Service) checkAccountInfo(c context.Context, mid int64) (err error) {
	var profileReply *accountapi.ProfileReply
	if profileReply, err = s.accClient.Profile3(c, &accountapi.MidReq{
		Mid: mid,
	}); err != nil {
		log.Error("accClient.Profile3(%v) error(%v)", mid, err)
		return nil
	}
	if profileReply.Profile.GetTelStatus() != 1 {
		return ecode.ActivityWriteHandTelValid
	}
	if profileReply.Profile.GetSilence() == 1 {
		return ecode.ActivityWriteHandBlocked
	}
	return
}

// getMidCoin 获取用户投币信息
func (s *Service) getMidCoin(c context.Context, mid int64) (int64, error) {
	resp, err := client.ActPlatClient.GetCounterRes(c, &actPlat.GetCounterResReq{
		Counter:  s.c.HandWrite.ActPlatCounter,
		Activity: s.c.HandWrite.ActPlatActivity,
		Mid:      mid,
		Time:     time.Now().Unix(),
	})
	if err != nil {
		log.Error("s.actPlatClient.GetCounterRes(%v) error(%v)", mid, err)
		err = errors.Wrapf(err, "s.actPlatClient.GetCounterRes %d", mid)
		return 0, err
	}
	if resp == nil || len(resp.CounterList) != 1 {
		log.Error("s.actPlatClient.GetCounterRes(%v) error(%v)", mid, err)
		err = errors.Wrapf(err, "s.actPlatClient.GetCounterRes %d return nil", mid)
		return 0, err
	}
	counter := resp.CounterList[0]
	return counter.Val, nil
}

// Coin 获取硬币数
func (s *Service) Coin(c context.Context, mid int64) (*mdlhw.CoinReply, error) {
	eg := errgroup.WithContext(c)
	day := time.Now().Format("20060102")
	var (
		coin            int64
		alreadyAddTimes string
		canAddTimes     bool
	)
	eg.Go(func(ctx context.Context) (err error) {
		if coin, err = s.getMidCoin(c, mid); err != nil {
			log.Error("s.getMidCoin(%d) error(%v)", mid, err)
			err = errors.Wrapf(err, "s.getMidCoin %d", mid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if alreadyAddTimes, err = s.handwrite.GetAddTimesRecord(c, mid, day); err != nil {
			log.Error("s.handwrite.GetAddTimesRecord(%d) error(%v)", mid, err)
			err = errors.Wrapf(err, "s.handwrite.GetAddTimesRecord %d", mid)
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, ecode.ActivityWriteHandGetCoinErr
	}
	if alreadyAddTimes == "" {
		canAddTimes = true
	}
	if coin > s.c.HandWrite.AwardCoinLimit {
		coin = s.c.HandWrite.AwardCoinLimit
	}
	return &mdlhw.CoinReply{Coin: coin, CanAddTimes: canAddTimes}, nil
}
