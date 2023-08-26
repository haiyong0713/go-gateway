package like

import (
	"context"
	"fmt"
	"strconv"
	"time"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	bbqtaskapi "git.bilibili.co/bapis/bapis-go/bbq/task"
	actPlat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"github.com/pkg/errors"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	mdlgh "go-gateway/app/web-svr/activity/interface/model/game_holiday"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_tpExtend            = 1
	_tpBase              = 2
	_tpArchive1          = 3
	_tpArchive2          = 4
	_tpLight             = 5
	_tpBcut              = 6
	_tpWinSN             = 10
	_otherAction         = 7
	stateWaitAddTimes    = 1
	stateCanAddTimes     = 2
	stateAlreadyAddTimes = 3
	_oneDecimalFmt       = "%.1f"
)

func (s *Service) loadContributionAwards() {
	ctx := context.Background()
	awards, err := s.dao.RawContriAwards(ctx)
	if err != nil {
		log.Errorc(ctx, "loadContributionAwards s.dao.RawAwards error(%+v)", err)
		return
	}
	if len(awards) > 0 {
		s.contributionAwards = awards
	}
}

func (s *Service) ArcInfo(ctx context.Context, mid int64) (res *like.ArchiveInfo, err error) {
	res = &like.ArchiveInfo{
		ViewCounts: &like.ViewCounts{},
		Awards:     &like.AwardsFinish{},
	}
	contribution, _ := s.getUserContribution(ctx, mid)
	if contribution != nil && contribution.Mid > 0 {
		res.IsJoin = true
	}
	if len(s.contributionAwards) == 0 {
		return
	}
	// 格式化金额
	s.formatArcAward(contribution, s.contributionAwards, res)
	return
}

func (s *Service) HaveMoney(c context.Context) {
	var (
		id       int64
		err      error
		res      *like.ArchiveInfo
		mids     []int64
		errCount int64
	)
	ctx := context.Background()
	mids = make([]int64, 0, 0)
	for {
		var data []int64
		data, id, err = s.dao.ContributionMids(ctx, id)
		if err != nil {
			errCount++
			if errCount > 3 {
				break
			}
			log.Errorc(ctx, "HaveMoney s.dao.ContributionMids id(%d) error(%+v)", id, err)
			time.Sleep(time.Second)
			continue
		}
		mids = append(mids, data...)
		if len(data) < 1000 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	for _, mid := range mids {
		if res, err = s.ArcInfo(ctx, mid); err != nil {
			log.Errorc(ctx, "HaveMoney s.ArcInfo mid(%d) error(%+v)", mid, err)
			continue
		}
		if _, err = s.dao.UpUserMoney(ctx, res.HaveMoney, mid); err != nil {
			log.Errorc(ctx, "HaveMoney s.dao.UpUserMoney mid(%+v) error(%+v)", mid, err)
		}
	}
}

func (s *Service) getUserContribution(ctx context.Context, mid int64) (res *like.ContributionUser, err error) {
	if mid == 0 {
		return
	}
	if res, err = s.dao.UserContribution(ctx, mid); err != nil {
		log.Error("s.getUserContribution(%d) error(%v)", mid, err)
	}
	return
}

func (s *Service) LightBcutInfo(ctx context.Context, mid int64) (res *like.LightBcut, err error) {
	res = &like.LightBcut{
		Lights: &like.LightBcutFinish{},
		Bcuts:  &like.LightBcutFinish{},
	}
	contribution, _ := s.getUserContribution(ctx, mid)
	if contribution != nil && contribution.Mid > 0 {
		res.IsJoin = true
	}
	if len(s.contributionAwards) == 0 {
		return
	}
	// 格式化完成视频
	s.formatLightBcut(ctx, mid, contribution, s.contributionAwards, res)
	if mid > 0 {
		s.lightVideoPartIn(ctx, mid)
	}
	return
}

func (s *Service) formatArcAward(contribution *like.ContributionUser, contriAwards []*like.ContriAwards, res *like.ArchiveInfo) {
	var extendMoney int64
	for _, award := range contriAwards {
		switch award.AwardType {
		case _tpExtend:
			res.ViewCounts.CurrentViews = award.CurrentViews
			res.ViewCounts.TargetViews = award.Views
			if award.Views > 0 {
				res.ViewCounts.ViewPercent = (float64(award.CurrentViews) / float64(award.Views)) * float64(100)
				if res.ViewCounts.ViewPercent > 100 {
					res.ViewCounts.ViewPercent = 100
				} else {
					res.ViewCounts.ViewPercent, _ = strconv.ParseFloat(fmt.Sprintf(_oneDecimalFmt, res.ViewCounts.ViewPercent), 64)
				}
			}
			if award.CurrentViews >= award.Views {
				extendMoney = award.SplitMoney
			}
		case _tpBase:
			res.ViewCounts.Money = award.SplitMoney + extendMoney
			if contribution != nil && contribution.UpArchives >= award.UpArchives && contribution.Likes >= award.Likes {
				res.Awards.BaseFinished = true
				if award.SplitPeople > 0 {
					res.HaveMoney += (float64(award.SplitMoney) + float64(extendMoney)) / float64(award.SplitPeople)
				}
			}
		case _tpWinSN:
			if s.c.S10Contribution.IsWinSN == 1 && contribution != nil && contribution.SnUpArchives >= award.SnUpArchives && contribution.SnLikes >= award.SnLikes {
				res.Awards.SnFinished = true
				if award.SplitPeople > 0 {
					res.HaveMoney += float64(award.SplitMoney) / float64(award.SplitPeople)
				}
			}
		default:
			if contribution != nil && contribution.Views >= award.Views {
				if award.AwardType == _tpArchive1 {
					res.Awards.OneFinished = true
					if award.SplitPeople > 0 {
						res.HaveMoney += (float64(award.SplitMoney)) / float64(award.SplitPeople)
					}
				} else if award.AwardType == _tpArchive2 {
					res.Awards.TwoFinished = true
					if award.SplitPeople > 0 {
						res.HaveMoney += (float64(award.SplitMoney)) / float64(award.SplitPeople)
					}
				}
			}
		}
	}
	res.HaveMoney, _ = strconv.ParseFloat(fmt.Sprintf(_oneDecimalFmt, res.HaveMoney), 64)
}

func (s *Service) formatLightBcut(ctx context.Context, mid int64, contribution *like.ContributionUser, contriAwards []*like.ContriAwards, res *like.LightBcut) {
	for _, award := range contriAwards {
		switch award.AwardType {
		case _tpLight:
			lightCount, err := s.dao.CacheLightCount(ctx, mid)
			if err != nil {
				err = nil
			}
			// 空缓存
			if lightCount == -1 {
				res.Lights.MyFinish = 0
				continue
			}
			if lightCount > 0 {
				res.Lights.MyFinish = lightCount
				continue
			}
			ProgressRly, err := s.bbqtaskClient.ActivityUserProgress(ctx, &bbqtaskapi.ActivityUserProgressReq{Mid: uint64(mid)})
			if err != nil {
				log.Errorc(ctx, "s.bbqtaskClient.ActivityUserProgress(%d) error(%+v)", mid, err)
				continue
			}
			if ProgressRly != nil {
				lightCount = int64(ProgressRly.Progress)
			}
			if lightCount > 3 {
				lightCount = 3
			}
			res.Lights.MyFinish = lightCount
			if err = s.dao.AddCacheLightCount(ctx, mid, lightCount); err != nil {
				log.Errorc(ctx, "s.dao.AddCacheLightCount mid(%d) error(%+v)", mid, err)
			}
		case _tpBcut:
			if contribution != nil {
				res.Bcuts.MyFinish = contribution.Bcuts
			}
		default:
		}
	}
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
			log.Errorc(c, "S10Contribution s.getMidLikes(%d) error(%v)", mid, err)
			err = errors.Wrapf(err, "s.getMidCoin %d", mid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if alreadyAddTimes, err = s.dao.GetAddTimesRecord(c, mid, _otherAction, day); err != nil {
			log.Errorc(c, "S10Contribution s.dao.GetAddTimesRecord(%d) error(%v)", mid, err)
			err = errors.Wrapf(err, "s.dao.GetAddTimesRecord %d", mid)
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait S10Contribution error(%v)", err)
		return nil, ecode.ActivityLikeGetErr
	}
	if alreadyAddTimes == "" {
		if likes < s.c.S10Contribution.AwardLikeLimit {
			state = stateWaitAddTimes
		} else {
			state = stateCanAddTimes
		}
	} else {
		state = stateAlreadyAddTimes
	}
	if likes > s.c.S10Contribution.AwardLikeLimit {
		likes = s.c.S10Contribution.AwardLikeLimit
	}
	return &mdlgh.LikesReply{Likes: likes, State: state}, nil
}

// getMidLikes 获取用户点赞信息
func (s *Service) getMidLikes(c context.Context, mid int64) (int64, error) {
	resp, err := client.ActPlatClient.GetCounterRes(c, &actPlat.GetCounterResReq{
		Counter:  s.c.S10Contribution.ActPlatCounter,
		Activity: s.c.S10Contribution.ActPlatActivity,
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

// AddContriLotteryTimes 增加抽奖次数
func (s *Service) AddContriLotteryTimes(c context.Context, mid int64, actionType int) (*mdlgh.AddTimesReply, error) {
	// 锁
	day := time.Now().Format("20060102")
	if err := s.dao.AddTimeLock(c, mid); err != nil {
		log.Errorc(c, "S10Contribution s.dao.AddTimeLock(%d) error(%v)", mid, err)
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
	if like < s.c.S10Contribution.AwardLikeLimit {
		return nil, ecode.ActivityLikeNotEnoughErr
	}
	orderNo := s.getOrderNo(mid)
	// 增加获奖次数
	if err = s.AddLotteryTimes(c, s.c.S10Contribution.LotteryID, mid, 0, actionType, 0, fmt.Sprint(orderNo), false); err != nil {
		log.Errorc(c, "S10Contribution s.AddLotteryTimes lotteryID(%d) mid(%d) actionType(%d)(%v)", s.c.S10Contribution.LotteryID, mid, actionType, err)
		return nil, err
	}
	if err := s.dao.AddTimesRecord(c, mid, actionType, day); err != nil {
		log.Errorc(c, "S10Contribution s.dao.AddTimesRecord(%d) error(%v)", mid, err)
	}
	return nil, nil
}

func (s *Service) getOrderNo(mid int64) string {
	return fmt.Sprintf("%d_%s_%s", mid, "s10_contribution", time.Now().Format("20060102"))
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

func (s *Service) TotalRank(ctx context.Context, dt string) (res *like.TotalRank, err error) {
	var top string
	if dt == "" {
		dt = time.Now().Format("20060102")
	}
	res = &like.TotalRank{
		Date: dt,
	}
	key := fmt.Sprintf("contri:%d:%s", s.c.S10Contribution.Sid, dt)
	top, err = s.dao.RsGet(ctx, key)
	res.Top = top
	return
}

func (s *Service) lightVideoPartIn(ctx context.Context, mid int64) (err error) {
	var join int64
	if join, err = s.dao.CacheLightVideoJoin(ctx, mid); err != nil {
		log.Errorc(ctx, "lightVideoPartIn s.dao.CacheLightVideoJoin mid(%d) error(%+v)", mid, err)
		return
	}
	if join > 0 {
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = s.bbqtaskClient.ActivityTakePartIn(ctx, &bbqtaskapi.ActivityTakePartInReq{Mid: uint64(mid)}); err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(ctx, "lightVideoPartIn s.bbqtaskClient.ActivityTakePartIn(%d) error(%+v)", mid, err)
	} else {
		if err = s.dao.AddCacheLightVideoJoin(ctx, mid); err != nil {
			log.Errorc(ctx, "lightVideoPartIn s.dao.AddCacheLightVideoJoin mid(%d) error(%+v)", mid, err)
		}
	}
	return err
}
