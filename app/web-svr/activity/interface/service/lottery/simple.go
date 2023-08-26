package lottery

import (
	"context"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/tool"
	"math/rand"

	"time"

	"github.com/pkg/errors"
)

// SimpleLottery 精简版抽奖
func (s *Service) SimpleLottery(c context.Context, sid string, mid int64, risk *riskmdl.Base, num int, alreadyWinTimes int, isInternal bool) (res []*l.RecordDetail, err error) {
	// 如果不是新的sid，走原来逻辑
	log.Infoc(c, "do simple lottery sid(%v)", sid)
	var (
		lottery         *l.Lottery
		info            *l.Info
		timesConf       []*l.TimesConfig
		gift            []*l.Gift
		member          *l.MemberInfo
		memberGroup     map[int64]*l.MemberGroup
		isRiskMember    bool
		isLimitMember   bool
		isSpyMember     bool
		canWinTimes     int
		hasSpecialTimes bool
		ip              = metadata.String(c, metadata.RemoteIP)
		now             = time.Now().Unix()
		remainTimes     []l.TimesInterface
		rate            int64
		buvid           = defaultBuvid
	)
	if risk != nil && risk.IP != "" {
		ip = risk.IP
	}
	if risk != nil && risk.Buvid != "" {
		buvid = risk.Buvid
	}
	// 获取基础数据
	eg := errgroup.WithContext(c)
	// eg.Go(func(ctx context.Context) (err error) {
	lottery, info, timesConf, gift, memberGroup, err = s.getBaseLotteryInfo(c, sid)
	if err != nil {
		log.Errorc(c, "check lottery err(%v)", err)
		return
	}
	// })
	eg.Go(func(ctx context.Context) (err error) {
		if member, err = s.getMemberInfo(ctx, mid, ip, buvid); err != nil {
			err = errors.Wrapf(err, "s.getMemberInfo %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if err = s.riskMember(ctx, lottery, mid, risk); err != nil {
			if err == ecode.ActivityLotteryRiskInfo {
				isRiskMember = true
				// 如果是风险用户，返回err

			}
		}

		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}
	// 检查抽奖信息
	if err = s.checkLottery(c, lottery, info, timesConf, gift, now, isInternal); err != nil {
		log.Errorc(c, "check lottery err(%v)", err)
		return
	}
	// 用户信息验证
	if err = s.checkMemberLottery(c, info, member); err != nil {
		log.Errorc(c, "s.checkMemberLottery(%v)", err)
		return
	}

	// 获取用户新老信息
	memberNewInfo, err := s.getMemberGroupNewTime(c, member, memberGroup)
	if err != nil {
		log.Errorc(c, "s.getMemberGroupNewTime err (%v)", err)
		return
	}

	// 获取用户预约情况
	memberActionReserveInfo, err := s.getMemberGroupActionReserve(c, member, memberGroup)
	if err != nil {
		log.Errorc(c, "s.getMemberGroupNewTime err (%v)", err)
		return
	}
	// 获取用户漫画新老情况
	member, err = s.getMemberGroupCartoon(c, member, memberGroup)
	if err != nil {
		log.Errorc(c, "s.getMemberGroupNewTime err (%v)", err)
		return
	}
	tool.IncrLotteryDoCount(lottery.ID, stageDo, num)

	// 中奖概率
	rate = s.getRate(c, info, hasSpecialTimes)
	// 是否风险用户
	isSpyMember = s.checkSpyMember(c, member, info, rate)
	// 能够中奖的次数
	most := s.getMostWinTimes(c, timesConf)
	canLottery := most - alreadyWinTimes
	canWinTimes = s.checkCanWinTimes(c, canLottery, isRiskMember, isLimitMember, isSpyMember)
	log.Infoc(c, "member mid(%d) canWinTimes(%d) remainTimes (%d) hasSpecialTimes(%v)", mid, canWinTimes, remainTimes, hasSpecialTimes)
	// 获取商品
	giftList, err := s.getLotteryFilteredGift(c, lottery.ID, gift, member, memberNewInfo, memberActionReserveInfo, memberGroup)
	if err != nil {
		log.Errorc(c, "s.getLotteryFilteredGift lotteryId(%v) mid(%d) error(%v)", lottery.ID, mid, err)
		return
	}
	// 抽奖逻辑
	record, err := s.simpledrawLottery(c, lottery.ID, lottery, info, timesConf, member, giftList, num, canWinTimes, hasSpecialTimes, rate)
	if err != nil {
		return record, err
	}

	return record, err
}

// drawLottery 抽奖行为
func (s *Service) simpledrawLottery(c context.Context, sid int64, lottery *l.Lottery, info *l.Info, timesConfig []*l.TimesConfig, member *l.MemberInfo, giftList []GiftInterface, num, canWinTimes int, hasSpecialTimes bool, rate int64) ([]*l.RecordDetail, error) {
	winTimes, err := s.lotteryBucket(c, sid, num, rate, hasSpecialTimes)
	if err != nil {
		return nil, err
	}
	if canWinTimes > num {
		canWinTimes = num
	}
	if winTimes > int64(canWinTimes) {
		winTimes = int64(canWinTimes)
	}
	// 抽奖的次数-中奖次数=可获得保底奖的次数
	cangetLeastGiftNum := 0
	if int64(canWinTimes) > winTimes {
		cangetLeastGiftNum = canWinTimes - int(winTimes)
	}
	log.Infoc(c, "drawLottery initWinTimes(%d) cangetLeastGiftNum(%d)", winTimes, cangetLeastGiftNum)
	// 如果不是必中，需要把外部抽奖次数以一定概率减掉
	if rate != l.MustWinRate {
		var outTimes = num
		thisWinTimes := winTimes
		for i := 0; i < int(thisWinTimes); i++ {
			// 如果是外部抽奖且不是必得，再做一次过
			outerRand := rand.Intn(num)
			outerTimesRand := rand.Intn(outerTimesRand)
			if outerRand < outTimes && outerTimesRand == 0 && winTimes > 0 {
				winTimes--
			}
		}
	}

	choseGift := s.choseGift(c, lottery, giftList, cangetLeastGiftNum, winTimes)
	log.Infoc(c, "chose gift (%v)", choseGift)
	return s.simpleSendGift(c, sid, lottery, info, member, choseGift, num)
}

// sendGiftAndRecord 发送gift和记录
func (s *Service) simpleSendGift(c context.Context, sid int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo, giftList []GiftInterface, num int) ([]*l.RecordDetail, error) {
	var err error
	day := time.Now().Format("2006-01-02")
	// 日库存发放加1
	remainGiftList := make([]GiftInterface, 0)
	remainGift := make([]*l.Gift, 0)
	if len(giftList) > 0 {
		giftBatchList := make([]*l.Gift, 0)
		for _, v := range giftList {
			giftBatchList = append(giftBatchList, v.ResGift())
		}
		resGiftList, err := s.checkSendStore(c, lottery, member.Mid, giftBatchList, day, giftSendNum)
		if err != nil {
			log.Errorc(c, "s.checkSendStore error(%v)", err)
			return nil, err
		}
		for _, v := range resGiftList {
			remainGift = append(remainGift, v)
			gift, _ := getGiftByType(c, v)
			remainGiftList = append(remainGiftList, gift)
			giftID := v.ID
			tool.IncrLotterySendGiftCount(sid, giftID, 1)
		}
	}
	// 消耗金币
	if err = s.consumeCoin(c, num, member, info); err != nil {
		return nil, err
	}
	var record = make([]*l.RecordDetail, 0)
	// 消耗次数
	if record, err = s.simpleRecord(c, sid, lottery, member, remainGift); err != nil {
		log.Errorc(c, " s.simpleRecord", err)
		return record, err
	}
	return record, nil
}

// consumeTimes
func (s *Service) simpleRecord(c context.Context, sid int64, lottery *l.Lottery, member *l.MemberInfo, giftBatch []*l.Gift) ([]*l.RecordDetail, error) {
	recordDetail := make([]*l.RecordDetail, 0)
	now := time.Now().Unix()

	for _, v := range giftBatch {
		recordDetail = append(recordDetail, &l.RecordDetail{
			Mid:      member.Mid,
			Num:      1,
			GiftID:   v.ID,
			GiftName: v.Name,
			GiftType: v.Type,
			ImgURL:   v.ImgURL,
			Type:     v.Type,
			Ctime:    xtime.Time(now),
			Extra:    v.Extra,
		})
	}
	return recordDetail, nil
}
