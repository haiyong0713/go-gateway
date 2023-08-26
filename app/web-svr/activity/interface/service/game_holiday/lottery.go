package gameholiday

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"time"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	mdlgh "go-gateway/app/web-svr/activity/interface/model/game_holiday"

	actPlat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"github.com/pkg/errors"
)

const (
	stateWaitAddTimes    = 1
	stateCanAddTimes     = 2
	stateAlreadyAddTimes = 3
)

// getMidCoin 获取用户投币信息
func (s *Service) getMidLikes(c context.Context, mid int64) (int64, error) {
	resp, err := client.ActPlatClient.GetCounterRes(c, &actPlat.GetCounterResReq{
		Counter:  s.c.GameHoliday.ActPlatCounter,
		Activity: s.c.GameHoliday.ActPlatActivity,
		Mid:      mid,
		Time:     time.Now().Unix(),
	})
	if err != nil {
		log.Errorc(c, "s.actPlatClient.GetCounterRes(%v) error(%v)", mid, err)
		err = errors.Wrapf(err, "s.actPlatClient.GetCounterRes %d", mid)
		return 0, err
	}
	if resp == nil || len(resp.CounterList) != 1 {
		log.Errorc(c, "s.actPlatClient.GetCounterRes(%v) error(%v)", mid, err)
		err = errors.Wrapf(err, "s.actPlatClient.GetCounterRes %d return nil", mid)
		return 0, err
	}
	counter := resp.CounterList[0]
	return counter.Val, nil
}

// Likes 获取点赞数
func (s *Service) Likes(c context.Context, mid int64) (*mdlgh.LikesReply, error) {
	eg := errgroup.WithContext(c)
	day := time.Now().Format("20060102")
	var (
		likes           int64
		alreadyAddTimes string
		state           int
	)
	eg.Go(func(ctx context.Context) (err error) {
		if likes, err = s.getMidLikes(c, mid); err != nil {
			log.Errorc(c, "s.getMidLikes(%d) error(%v)", mid, err)
			err = errors.Wrapf(err, "s.getMidCoin %d", mid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if alreadyAddTimes, err = s.gameHoliday.GetAddTimesRecord(c, mid, day); err != nil {
			log.Errorc(c, "s.gameholiday.GetAddTimesRecord(%d) error(%v)", mid, err)
			err = errors.Wrapf(err, "s.gameholiday.GetAddTimesRecord %d", mid)
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, ecode.ActivityLikeGetErr
	}
	if alreadyAddTimes == "" {
		if likes < s.c.GameHoliday.AwardLikeLimit {
			state = stateWaitAddTimes
		} else {
			state = stateCanAddTimes
		}
	} else {
		state = stateAlreadyAddTimes
	}
	if likes > s.c.GameHoliday.AwardLikeLimit {
		likes = s.c.GameHoliday.AwardLikeLimit
	}
	return &mdlgh.LikesReply{Likes: likes, State: state}, nil
}

// AddLotteryTimes 增加抽奖次数
func (s *Service) AddLotteryTimes(c context.Context, mid int64) (*mdlgh.AddTimesReply, error) {
	// 锁
	day := time.Now().Format("20060102")
	if err := s.gameHoliday.AddTimeLock(c, mid); err != nil {
		log.Errorc(c, "s.gameHoliday.AddTimeLock(%d) error(%v)", mid, err)
		return nil, ecode.ActivityWriteHandAddtimesTooFastErr
	}
	// 账号信息验证
	if err := s.checkAccountInfo(c, mid); err != nil {
		return nil, err
	}
	// 获取当前金币数
	like, err := s.getMidLikes(c, mid)
	if err != nil {
		return nil, ecode.ActivityLikeGetErr
	}
	if like < s.c.GameHoliday.AwardLikeLimit {
		return nil, ecode.ActivityLikeNotEnoughErr
	}
	orderNo := s.getOrderNo(c, mid)
	// 增加获奖次数
	if err = s.like.AddLotteryTimes(c, s.c.GameHoliday.Sid, mid, 0, 7, 0, fmt.Sprint(orderNo), false); err != nil {
		log.Errorc(c, "s.dao.Addtimes(%v)", err)
		return nil, err
	}
	if err := s.gameHoliday.AddTimesRecord(c, mid, day); err != nil {
		log.Errorc(c, "s.gameHoliday.AddTimesRecord(%d) error(%v)", mid, err)
	}
	return nil, nil
}

func (s *Service) getOrderNo(c context.Context, mid int64) string {
	return fmt.Sprintf("%d_%v_%v", mid, "game_holiday", time.Now().Format("20060102"))
}
func (s *Service) checkAccountInfo(c context.Context, mid int64) (err error) {
	var profileReply *accountapi.ProfileReply
	if profileReply, err = s.accClient.Profile3(c, &accountapi.MidReq{
		Mid: mid,
	}); err != nil {
		log.Errorc(c, "accClient.Profile3(%v) error(%v)", mid, err)
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
