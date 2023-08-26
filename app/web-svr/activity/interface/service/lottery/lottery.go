package lottery

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/tool"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"

	figure "git.bilibili.co/bapis/bapis-go/account/service/figure"
	spy "git.bilibili.co/bapis/bapis-go/account/service/spy"
	locationAPI "git.bilibili.co/bapis/bapis-go/community/service/location"
	passportinfoapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	silverbulletapi "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	vipinfoapi "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"

	"github.com/pkg/errors"
)

const (
	lotteryStrategy = "lottery_strategy"
	bucketNum       = 5
	giftSendNum     = 1
	mc              = "1_4_1"
	qpsLimitMax     = 1
	defaultBuvid    = "activity_lottery_zi199TossPl4c9dw1"
	// outerTimesRand 外部获得的抽奖 再随机一次
	outerTimesRand = 2
	paramsSplit    = "`||"
	stageDo        = "stage_do"
	stageGetGift   = "stage_chose_gift"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}
func (s *Service) doNewLottery(c context.Context, sid string) (bool, error) {
	lottery, err := s.base(c, sid)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		return false, err
	}
	if lottery == nil {
		return false, ecode.ActivityNotExist
	}

	if int64(lottery.Ctime) < s.c.Lottery.NewLotteryTime {
		return false, nil
	}
	if int64(lottery.Ctime) > s.c.Lottery.NewLotteryTime100 {
		return true, nil
	}
	if int64(lottery.Ctime) > s.c.Lottery.NewLotteryTime75 {
		if lottery.ID%3 == 1 || lottery.ID%3 == 2 {
			return true, nil
		}
		return false, nil
	}
	if int64(lottery.Ctime) > s.c.Lottery.NewLotteryTime50 {
		if lottery.ID%2 == 1 {
			return true, nil
		}
		return false, nil
	}

	if lottery.ID%3 == 1 {
		return true, nil
	}
	return false, nil
}

// DoLottery 抽奖
func (s *Service) DoLottery(c context.Context, sid string, mid int64, risk *riskmdl.Base, num int, isInternal bool, orderNo string) (res []*l.RecordDetail, err error) {
	check, err := s.PreCheck(c, sid)
	if err != nil {
		return
	}
	if check != nil {
		err = check.check(c, s, mid)
		if err != nil {
			return
		}
	}
	// 如果不是新的sid，走原来逻辑
	log.Infoc(c, "do new lottery sid(%v)", sid)
	// if err = s.qpsLimit(c, mid, qpsLimitMax); err != nil {
	// 	log.Errorc(c, "mid(%d) sid(%s) qps error error(%v)", mid, sid, err)
	// 	return nil, err
	// }
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
				// 如果是风险用户，不提示错误，消耗抽奖次数并不中奖
				err = nil
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

	// 获取抽奖次数
	usedTimes, addTimes, err := s.getMidLotteryTimes(c, timesConf, lottery.ID, mid)
	if err != nil {
		log.Errorc(c, "s.getMidLotteryTimes err (%v)", err)
		return
	}
	// 验证使用次数
	canWinTimes, remainTimes, hasSpecialTimes, err = s.checkUsedTimes(c, timesConf, num, usedTimes, addTimes, info.HighType)
	if err != nil {
		log.Errorc(c, "s.checkUsedTimes(%v,%d,%d,%d) error(%v)", timesConf, lottery.ID, mid, num, err)
		return
	}
	tool.IncrLotteryDoCount(lottery.ID, stageDo, num)
	// 中奖概率
	rate = s.getRate(c, info, hasSpecialTimes)
	// 是否风险用户
	isSpyMember = s.checkSpyMember(c, member, info, rate)
	// 能够中奖的次数
	canWinTimes = s.checkCanWinTimes(c, canWinTimes, isRiskMember, isLimitMember, isSpyMember)
	thisTimesConsume, newNum := s.getConsumeTimes(c, remainTimes, num)
	log.Infoc(c, "member mid(%d) canWinTimes(%d) remainTimes (%d) hasSpecialTimes(%v)", mid, canWinTimes, remainTimes, hasSpecialTimes)
	// 获取商品
	giftList, err := s.getLotteryFilteredGift(c, lottery.ID, gift, member, memberNewInfo, memberActionReserveInfo, memberGroup)
	if err != nil {
		log.Errorc(c, "s.getLotteryFilteredGift lotteryId(%v) mid(%d) error(%v)", lottery.ID, mid, err)
		return
	}
	// 抽奖逻辑
	record, err := s.drawLottery(c, lottery.ID, lottery, info, timesConf, member, giftList, thisTimesConsume, newNum, canWinTimes, hasSpecialTimes, rate, orderNo)
	if err != nil {
		return record, err
	}
	if check != nil {
		err = check.checkError(c, err)
	}
	return record, err
}

// qpsLimit ...
func (s *Service) qpsLimit(c context.Context, mid int64, maxLimit int64) error {
	limit, err := s.lottery.CacheQPSLimit(c, mid)
	if err != nil {
		log.Errorc(c, "s.lottery.CacheQPSLimit(%d) error(%v)", mid, err)
		return err
	}
	if limit > maxLimit {
		return ecode.ActivityQPSLimitErr
	}
	return nil
}

// InitLottery ...
func (s *Service) InitLottery(c context.Context, sid string) (bool, error) {
	for _, v := range s.c.Lottery.NewLotterSid {
		if v == sid {
			return true, nil
		}
	}
	return s.doNewLottery(c, sid)
}

// getMemberGroupActionReserve 根据配置获取用户新老情况
func (s *Service) getMemberGroupActionReserve(c context.Context, member *l.MemberInfo, memberGroup map[int64]*l.MemberGroup) (*l.MemberIdsInfo, error) {
	var err error
	var memberIdsInfo = &l.MemberIdsInfo{}
	memberIdsInfo.Info = make(map[int64]bool)
	ids := make([]int64, 0)
	g := &l.GroupAction{}
	if memberGroup != nil {
		for _, group := range memberGroup {
			for _, v := range group.Group {
				if v.GroupType == l.GroupTypeAction {
					err := g.Init(v)
					if err != nil {
						continue
					}
					if g.Action == l.GroupTypeActionReserve {
						ids = g.Ids
					}
				}
			}
		}
		if len(ids) > 0 {
			eg := errgroup.WithContext(c)
			for _, v := range ids {
				id := v
				eg.Go(func(ctx context.Context) (err error) {
					newR, e := s.likedao.ReserveOnly(ctx, id, member.Mid)
					if e != nil {
						log.Errorc(ctx, "s.dao.ReserveOnly(%v,%d) error(%v)", id, member.Mid, e)
						return nil
					}
					if newR != nil && newR.State == 1 {
						memberIdsInfo.Set(id, true)
					}
					return nil
				})
			}
		}
	}

	return memberIdsInfo, err

}

// getMemberGroupCartoon 根据配置获取用户新老情况
func (s *Service) getMemberGroupCartoon(c context.Context, member *l.MemberInfo, memberGroup map[int64]*l.MemberGroup) (*l.MemberInfo, error) {
	var err error
	var memberIdsInfo = &l.MemberIdsInfo{}
	memberIdsInfo.Info = make(map[int64]bool)
	if memberGroup != nil {
		for _, group := range memberGroup {
			for _, v := range group.Group {
				if v.GroupType == l.GroupTypeCartoon {
					isRookie, err := s.lottery.ComicsIsRookie(c, member.Mid)
					if err != nil {
						log.Errorc(c, "s.lottery.ComicsIsRookie err(%v)", err)
						return member, nil
					}
					if isRookie == l.IsRookie {
						member.IsCartoonNew = true
					}
				}
			}
		}

	}
	return member, err

}

// getMemberGroupNewTime 根据配置获取用户新老情况
func (s *Service) getMemberGroupNewTime(c context.Context, member *l.MemberInfo, memberGroup map[int64]*l.MemberGroup) (*l.MemberNewInfo, error) {
	var err error
	var memberNewInfo = &l.MemberNewInfo{}
	memberNewInfo.Info = make(map[int64]bool)
	period := make(map[int64]struct{})
	groupNewMember := make([]*l.GroupNewMember, 0)
	if memberGroup != nil {
		for _, group := range memberGroup {
			for _, v := range group.Group {
				if v.GroupType == l.GroupTypeNewMember {
					g := &l.GroupNewMember{}
					g.Init(v)
					if _, ok := period[g.Period]; ok {
						continue
					}
					groupNewMember = append(groupNewMember, g)
					period[g.Period] = struct{}{}
				}
			}
		}
	}
	if len(groupNewMember) > 0 {
		eg := errgroup.WithContext(c)
		for _, v := range groupNewMember {
			period := v.Period
			eg.Go(func(ctx context.Context) (err error) {
				var infoReply *passportinfoapi.CheckFreshUserReply
				if infoReply, err = s.passportClient.CheckFreshUser(c, &passportinfoapi.CheckFreshUserReq{Mid: member.Mid, Buvid: member.DeviceID, Period: period}); err != nil || infoReply == nil {
					log.Errorc(c, "s.passportClient.CheckFreshUser(%d) infoReply(%v) error(%v)", member.Mid, infoReply, err)
					err = errors.Wrapf(err, "s.passportClient.CheckFreshUser %d", member.Mid)
					return err
				}
				log.Infoc(c, "s.passportClient.CheckFreshUser(%d, %s, %d,) infoReply(%v)", member.Mid, member.DeviceID, period, infoReply.IsNew)
				memberNewInfo.Set(period, infoReply.IsNew)
				return nil
			})
		}
		if err = eg.Wait(); err != nil {
			log.Errorc(c, "eg.Wait error(%v)", err)
			return nil, err
		}
	}
	return memberNewInfo, err

}

// consumeCoin 抽奖消耗硬币
func (s *Service) consumeCoin(c context.Context, num int, member *l.MemberInfo, info *l.Info) error {
	if info.Coin > 0 {
		if _, err := s.coinClient.ModifyCoins(c, &coinmdl.ModifyCoinsReq{Mid: member.Mid, Count: float64(-info.Coin * num), Reason: l.CoinConsume, IP: member.IP}); err != nil {
			return ecode.ActivityNotEnoughCoin
		}
	}
	return nil
}

func (s *Service) checkCanWinTimes(c context.Context, canWinTimes int, isRiskMember, isLimitMember, isSpyMember bool) int {
	if isRiskMember || isLimitMember || isSpyMember {
		//消耗抽奖次数，不发奖
		canWinTimes = 0
	}
	log.Infoc(c, " isRiskMember(%v) isLimitMember(%v) isSpyMember (%v)", isRiskMember, isLimitMember, isSpyMember)
	return canWinTimes
}

func (s *Service) getRate(c context.Context, info *l.Info, hasSpecialTimes bool) int64 {
	rate := info.GiftRate
	if hasSpecialTimes {
		rate = info.HighRate
	}
	return rate
}

// checkSpyMember 验证信用分
func (s *Service) checkSpyMember(c context.Context, member *l.MemberInfo, info *l.Info, rate int64) bool {
	if rate == l.MustWinRate {
		return false
	}
	if member.SpyScore <= int32(info.SpyScore) || int64(member.Percentage) >= info.FigureScore {
		return true
	}
	return false

}

// drawLottery 抽奖行为
func (s *Service) drawLottery(c context.Context, sid int64, lottery *l.Lottery, info *l.Info, timesConfig []*l.TimesConfig, member *l.MemberInfo, giftList []GiftInterface, thisTimesConsume []l.TimesInterface, num, canWinTimes int, hasSpecialTimes bool, rate int64, orderNo string) ([]*l.RecordDetail, error) {
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
		var outTimes int
		for _, v := range thisTimesConsume {
			// 记录外部抽奖次数
			if !v.IsInternal() {
				times := v.Record()
				outTimes += times.Num

			}
		}
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
	return s.sendGiftAndRecord(c, sid, lottery, info, timesConfig, member, choseGift, num, thisTimesConsume, orderNo)
}

// sendGiftAndRecord 发送gift和记录
func (s *Service) sendGiftAndRecord(c context.Context, sid int64, lottery *l.Lottery, info *l.Info, timesConfig []*l.TimesConfig, member *l.MemberInfo, giftList []GiftInterface, num int, thisTimesConsume []l.TimesInterface, orderNo string) ([]*l.RecordDetail, error) {
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
		// 真实抽中的奖品个数
	}

	// 消耗金币
	if err = s.consumeCoin(c, num, member, info); err != nil {
		return nil, err
	}
	var record = make([]*l.RecordDetail, 0)
	// 消耗次数
	if record, err = s.consumeTimes(c, sid, lottery, member, timesConfig, thisTimesConsume, remainGift, orderNo); err != nil {
		return nil, err
	}
	// 发放奖品
	s.cache.SyncDo(c, func(ctx context.Context) {
		for _, v := range remainGiftList {
			if err = v.Init(); err != nil {
				log.Errorc(ctx, "gift v.Init(),gift(%v) error(%v)", v, err)
				continue
			}
			err = v.Send(ctx, s, sid, info.SenderID, lottery, info, member)
			if err != nil {
				log.Errorc(ctx, "gift v.Send(),gift(%v) error(%v)", v, err)
				content := fmt.Sprintf("sid:%s,giftid:%d,mid:%d,ip:%s,err:%v", lottery.LotteryID, v.ResGift().ID, member.Mid, member.IP, err)
				err = s.wechatdao.SendWeChat(ctx, s.c.Wechat.PublicKey, "[抽奖]", content, "zhangtinghua")
				if err != nil {
					log.Errorc(ctx, " s.wechatdao.SendWeChat(%v)", err)
				}
				continue
			}
			log.Infoc(ctx, "gift v.Send(),gift(%v) success", v)
		}
		err = s.lottery.DeleteLotteryWinLog(ctx, sid, member.Mid)
		if err != nil {
			log.Errorc(ctx, "s.lottery.DeleteLotteryWinLog err(%v)", err)
		}
	})
	return record, nil
}

func (s *Service) recordBuildOrderNo(c context.Context, index int, round int, mid, cid int64, orderNo string) string {
	if orderNo != "" {
		return fmt.Sprintf("%d_%d@%s", index, round, orderNo)
	}
	return fmt.Sprintf("%d_%d_%d_%d", index, mid, cid, time.Now().Unix())

}

// SupplymentWin 不降频
func (s *Service) SupplymentWin(c context.Context, sid string, mid int64, giftID int64, ip string) error {
	lottery, info, _, gift, _, err := s.getBaseLotteryInfo(c, sid)
	if err != nil {
		log.Errorc(c, "s.getBaseLotteryInfo sid(%s) err(%v)", sid, err)
		return err
	}
	member, err := s.getMemberInfo(c, mid, ip, "")
	if err != nil {
		return err
	}
	var giftInterface GiftInterface
	for _, v := range gift {
		if v.ID == giftID {
			giftInterface, _ = getGiftByType(c, v)
			giftInterface.Init()
			return giftInterface.Send(c, s, lottery.ID, info.SenderID, lottery, info, member)
		}
	}
	return nil
}

// consumeTimes
func (s *Service) consumeTimes(c context.Context, sid int64, lottery *l.Lottery, member *l.MemberInfo, timesConfig []*l.TimesConfig, thisTimesConsume []l.TimesInterface, giftBatch []*l.Gift, orderNo string) ([]*l.RecordDetail, error) {
	recordBatch := make([]*l.InsertRecord, 0)
	recordDetail := make([]*l.RecordDetail, 0)
	recordDetailRedis := make([]*l.RecordDetail, 0)
	giftIds := make([]int64, 0)
	now := time.Now().Unix()
	for _, v := range giftBatch {
		giftIds = append(giftIds, v.ID)
	}
	var round int
	for _, v := range thisTimesConsume {
		record := v.Record()
		round++
		for index := 0; index < record.Num; index++ {

			i := index
			recordBatch = append(recordBatch, &l.InsertRecord{
				Mid:     member.Mid,
				Num:     1,
				Type:    record.Type,
				CID:     record.CID,
				OrderNo: s.recordBuildOrderNo(c, i, round, member.Mid, record.CID, orderNo),
			})
		}
		recordDetailRedis = append(recordDetailRedis, &l.RecordDetail{
			Mid:   member.Mid,
			Num:   record.Num,
			Type:  record.Type,
			CID:   record.CID,
			Ctime: xtime.Time(now),
		})
	}
	if len(giftBatch) < len(recordBatch) {
		giftIDDiff := len(recordBatch) - len(giftBatch)
		for i := 0; i < giftIDDiff; i++ {
			giftIds = append(giftIds, 0)
		}
	}

	if len(giftBatch) > 0 {
		var winCID int64
		for _, v := range timesConfig {
			// 记录win的cid
			if v.Type == l.TimesWinType {
				winCID = v.ID
			}
		}
		for _, v := range giftBatch {
			recordDetailRedis = append(recordDetailRedis, &l.RecordDetail{
				Mid:    member.Mid,
				Num:    1,
				Type:   l.TimesWinType,
				CID:    winCID,
				GiftID: v.ID,
				Ctime:  xtime.Time(now),
			})
		}
	}

	var i int
	extra := make(map[string]string)
	for k, v := range recordBatch {
		if k >= len(giftBatch) {
			recordDetail = append(recordDetail, &l.RecordDetail{
				Mid:      member.Mid,
				Num:      1,
				GiftID:   0,
				GiftType: 0,
				GiftName: fmt.Sprintf("未中奖%d", i),
				ImgURL:   "",
				Type:     v.Type,
				Ctime:    xtime.Time(now),
				Extra:    extra,
			})
			i++
			continue
		}
		recordDetail = append(recordDetail, &l.RecordDetail{
			Mid:      member.Mid,
			Num:      1,
			GiftID:   giftBatch[k].ID,
			GiftName: giftBatch[k].Name,
			GiftType: giftBatch[k].Type,
			ImgURL:   giftBatch[k].ImgURL,
			Type:     v.Type,
			Ctime:    xtime.Time(now),
			Extra:    giftBatch[k].Extra,
		})
	}
	var err error
	if int64(lottery.Ctime) > s.c.Lottery.NewLotteryTime {
		_, err = s.lottery.InsertLotteryRecardOrderNo(c, sid, recordBatch, giftIds, member.IP)

	} else {
		_, err = s.lottery.InsertLotteryRecard(c, sid, recordBatch, giftIds, member.IP)
	}
	if err != nil {
		log.Errorc(c, "s.lottery.InsertLotteryRecard(%d,%v,%v,%s) error(%v)", sid, recordBatch, giftIds, member.IP, err)
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = ecode.ActivityLotteryDuplicateErr
		}
		return nil, err
	}
	// redis incr usedtimes
	recordRedis := s.operateLotteryTimes(c, timesConfig, recordDetailRedis)
	if err = s.lottery.IncrTimes(c, sid, member.Mid, recordRedis, l.UsedTimesKey); err != nil {
		log.Errorc(c, "s.lottery.IncrTimes(%d,%d,%v,%s)", sid, member.Mid, recordRedis, l.UsedTimesKey)
		return nil, err
	}
	//新增抽奖记录redis
	if err = s.lottery.DeleteLotteryActionLog(c, sid, member.Mid); err != nil {
		log.Errorc(c, "DoLottery s.lottery.AddLotteryActionLog sid(%d) mid(%d) arg(%v) error(%v)", sid, member.Mid, recordDetail, err)
		return nil, err
	}
	return recordDetail, nil
}

// getLeastGift 获得必得奖品
func (s *Service) getLeastGift(c context.Context, giftList []GiftInterface) GiftInterface {
	for _, v := range giftList {
		if v.ResGift().LeastMark == l.GiftLeastMark {
			return v
		}
	}
	return nil
}

// isLeastGift 是否保底奖
func (s *Service) isLeastGift(c context.Context, gift GiftInterface) bool {
	if gift.ResGift().LeastMark == l.GiftLeastMark {
		return true
	}
	return false
}

// getLimitMinNum 单奖品可获得上限
func getLimitMinNum(gift GiftInterface) int64 {
	g := gift.ResGift()
	var num = g.Num - g.SendNum
	if g.DayNum != nil {
		for k, v := range g.DayNum {
			if v < num && v != 0 {
				if sendNum, ok := g.DaySendNum[k]; ok {
					if v-sendNum > 0 && v-sendNum < num {
						num = v - sendNum
					}
				}
				if otherSendNum, ok := g.OtherSendNum[k]; ok {
					if v-otherSendNum > 0 && v-otherSendNum < num {
						num = v - otherSendNum
					}
				}
			}
		}
	}
	return num
}

// choseGift 选择商品
func (s *Service) choseGift(c context.Context, lottery *l.Lottery, giftList []GiftInterface, canGetLeastGiftNum int, canGetGiftNum int64) (gift []GiftInterface) {
	log.Infoc(c, "choseGift canGetLeastGiftNum(%d) canGetGiftNum(%d)", canGetLeastGiftNum, canGetGiftNum)
	gift = make([]GiftInterface, 0)
	if canGetGiftNum+int64(canGetLeastGiftNum) == 0 {
		return
	}
	intGetNum := int(canGetGiftNum)
	probabilityList := make([]int64, 0)
	noLeastGift := make([]GiftInterface, 0)
	var allProbability int64
	for _, v := range giftList {
		if !s.isLeastGift(c, v) {
			noLeastGift = append(noLeastGift, v)
		}
	}
	for _, v := range noLeastGift {
		probability := v.ResGift().Probability * v.ResGift().Num
		allProbability += probability
		probabilityList = append(probabilityList, allProbability)
	}
	for _, v := range noLeastGift {
		var p float64
		probability := v.ResGift().Probability * v.ResGift().Num
		if allProbability > 0 {
			p = float64(probability) / float64(allProbability)
			tool.IncrLotterySendGiftProbability(lottery.ID, v.ResGift().ID, p)
		}
	}
	if allProbability != 0 {
		giftRand := make([]int64, 0, 1000)
		for intGetNum > 0 {
			curRand := rand.Int63n(allProbability) + 1
			giftRand = append(giftRand, curRand)
			if len(giftRand) > 1000 {
				break
			}
			giftMapNum := make(map[int64]int64)
			find := true
			for i, probability := range probabilityList {
				if curRand <= probability {
					giftLimitNum := getLimitMinNum(noLeastGift[i])
					if giftLimitNum > 0 {
						if n, ok := giftMapNum[noLeastGift[i].ResGift().ID]; ok {
							if n >= giftLimitNum {
								find = false
								break
							}
						}
						giftMapNum[noLeastGift[i].ResGift().ID]++
					}
					gift = append(gift, noLeastGift[i])
					break
				}
			}
			if find {
				intGetNum--
			}
		}
	}
	leastGift := s.getLeastGift(c, giftList)
	if leastGift != nil {
		for i := 0; i < canGetLeastGiftNum; i++ {
			gift = append(gift, leastGift)
		}
	}
	tool.IncrLotteryDoCount(lottery.ID, stageGetGift, len(gift))
	return
}

// lotteryBucket 中奖桶,返回中几次奖
func (s *Service) lotteryBucket(c context.Context, sid int64, num int, rate int64, hasSpecialTimes bool) (winTimes int64, err error) {
	var isHigh = 0
	if hasSpecialTimes {
		isHigh = 1
	}
	if rate == l.MustWinRate {
		winTimes = int64(num)
		return
	}
	bucket := make([]int64, bucketNum)
	bucketIDMap := make(map[int]struct{})
	for i := 0; i < num; i++ {
		buckedID := rand.Intn(bucketNum)
		bucket[buckedID]++
		bucketIDMap[buckedID] = struct{}{}
	}
	for bucketID := range bucketIDMap {
		var tmp int64
		if tmp, err = s.lottery.CacheLotteryMcNum(c, sid, isHigh, bucketID); err != nil {
			log.Errorc(c, "Immediate s.dao.CacheLotteryMcNum sid(%d) high(%d) mc(%d) error(%v)", sid, isHigh, bucketID, err)
			return
		}
		var bucketHeight int64
		bucketHeight = tmp + bucket[bucketID]
		// 若原本tmp比rate多，则不中奖
		if tmp > rate {
			// 高度折半
			bucketHeight = tmp / 2
		}
		// 高度正好等于比例
		if bucketHeight == rate {
			winTimes++
			// 高度清0
			bucketHeight = 0
		}
		// 如果bucketHeight< rateInt64，直接更新bucketHeight
		if tmp <= rate && bucketHeight > rate {
			winTimes += bucketHeight / rate
			bucketHeight = bucketHeight % rate
		}
		if err = s.lottery.AddCacheLotteryMcNum(c, sid, isHigh, bucketID, bucketHeight); err != nil {
			log.Errorc(c, "Immediate s.dao.AddCacheLotteryMcNum sid(%d) high(%d) mc(%d) val(%d) error(%v)", sid, isHigh, bucketID, bucketHeight, err)
			return
		}

	}
	return
}

// getConsumeTimes 获取将要消耗的次数，一次获得的抽奖次数可能是多个，所以需要判断本次消耗多少个
func (s *Service) getConsumeTimes(c context.Context, remainTimes []l.TimesInterface, num int) (consumeTimes []l.TimesInterface, newNum int) {
	consumeTimes = make([]l.TimesInterface, 0)
	newNum = num
	for _, v := range remainTimes {
		times := v
		record := times.Record()
		thisTimesNum := record.Num
		// 额外赠送次数
		if record.Type == l.TimesAdditionalType {
			newRecord, err := l.GetTimesByType(record)
			if err != nil {
				continue
			}
			consumeTimes = append(consumeTimes, newRecord)
			newNum += record.Num
			if err != nil {
				continue
			}
			continue
		}
		if num == 0 {
			break
		}
		if thisTimesNum <= num {
			newRecord, err := l.GetTimesByType(record)
			if err != nil {
				continue
			}
			consumeTimes = append(consumeTimes, newRecord)
			num -= thisTimesNum
			continue
		}
		record.Num = num
		newRecord, err := l.GetTimesByType(record)
		if err != nil {
			continue
		}
		consumeTimes = append(consumeTimes, newRecord)
		break
	}
	return
}

// getLotteryFilteredGift 获取用户可中奖商品
func (s *Service) getLotteryFilteredGift(c context.Context, sid int64, giftList []*l.Gift, memberInfo *l.MemberInfo, memberNewInfo *l.MemberNewInfo, memberActionInfo *l.MemberIdsInfo, memberGroup map[int64]*l.MemberGroup) (res []GiftInterface, err error) {
	res = make([]GiftInterface, 0)
	efficientGiftList := make([]*l.Gift, 0)
	for _, v := range giftList {
		if v.Efficient == l.Efficient {
			efficientGiftList = append(efficientGiftList, v)
		}
	}
	remainGiftList, err := s.filterGiftStore(c, sid, memberInfo.Mid, efficientGiftList)
	if err != nil {
		return
	}
	newGiftList, err := s.filterGiftMemberGroup(c, remainGiftList, memberGroup, memberInfo, memberNewInfo, memberActionInfo)
	if err != nil {
		return
	}
	for _, v := range newGiftList {
		gift, err := getGiftByType(c, v)
		if err != nil {
			continue
		}
		gift.Init()
		if err = gift.Check(memberInfo); err != nil {
			log.Errorc(c, "gift.Check(%v);error(%v)", memberInfo, err)
			continue
		}
		res = append(res, gift)
	}
	log.Infoc(c, "giftList efficient(%v) filterStore(%v) memberGroup(%v) checkGift(%v)", efficientGiftList, remainGiftList, newGiftList, res)

	return
}

// filterGiftMemberGroup 过滤商品用户组
func (s *Service) filterGiftMemberGroup(c context.Context, giftList []*l.Gift, memberGroup map[int64]*l.MemberGroup, member *l.MemberInfo, memberNewInfo *l.MemberNewInfo, memberActionInfo *l.MemberIdsInfo) (res []*l.Gift, err error) {
	res = make([]*l.Gift, 0)
	giftMemberGroup := s.getGiftMemberGroup(c, giftList, memberGroup)
	giftFilterIds := make(map[int64]struct{})
	for giftID, groupBatch := range giftMemberGroup {
		if groupBatch == nil || len(groupBatch) == 0 {
			giftFilterIds[giftID] = struct{}{}
		}
		for _, v := range groupBatch {
			err = s.memberConformToMemberGroup(c, member, memberNewInfo, memberActionInfo, v)
			if err != nil {
				log.Errorc(c, "member can not comform to member group member(%v) giftID(%d) memberNewInfo(%v) memberActionInfo(%v) groupBatch(%v) error(%v)", member, giftID, memberNewInfo, memberActionInfo, v, err)
				continue
			}
			if err == nil {
				giftFilterIds[giftID] = struct{}{}
				break
			}
		}
	}
	for _, v := range giftList {
		if _, ok := giftFilterIds[v.ID]; ok {
			res = append(res, v)
		}
	}
	return res, nil
}

// memberConformToMemberGroup 用户是否满足标准
func (s *Service) memberConformToMemberGroup(c context.Context, member *l.MemberInfo, memberNewInfo *l.MemberNewInfo, memberActionInfo *l.MemberIdsInfo, memberGroup *l.MemberGroup) error {
	if memberGroup == nil {
		return nil
	}
	for _, v := range memberGroup.Group {
		group, err := l.GetMemberGroup(v)
		if err != nil {
			return err
		}
		err = group.Init(v)
		if err != nil {
			return err
		}
		err = group.Check(member, memberNewInfo, memberActionInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) getGiftMemberGroup(c context.Context, giftList []*l.Gift, memberGroup map[int64]*l.MemberGroup) (giftMemberGroup map[int64][]*l.MemberGroup) {
	giftMemberGroup = make(map[int64][]*l.MemberGroup)
	for _, v := range giftList {
		giftMemberGroupBatch := make([]*l.MemberGroup, 0)
		for _, memberGroupID := range v.MemberGroup {
			memberGroupData, ok := memberGroup[memberGroupID]
			if !ok {
				log.Errorc(c, "memberGroup no get,memberGroupID (%d) Sid(%v)", memberGroupID, v.Sid)
				continue
			}
			giftMemberGroupBatch = append(giftMemberGroupBatch, memberGroupData)
		}
		giftMemberGroup[v.ID] = giftMemberGroupBatch
	}
	return
}

// filterGiftStore
func (s *Service) filterGiftStore(c context.Context, sid, mid int64, giftList []*l.Gift) (remainGiftList []*l.Gift, err error) {
	day := time.Now().Format("2006-01-02")
	resGiftList, err := s.getLatestGiftStore(c, sid, mid, giftList, day)
	if err != nil {
		log.Errorc(c, "s.getLatestGiftStore(%d,%v,%s) error(%v)", sid, giftList, day, err)
		return
	}
	remainGiftList = make([]*l.Gift, 0)
	for _, v := range resGiftList {
		if v.CheckStore(c) != nil {
			continue
		}
		remainGiftList = append(remainGiftList, v)
	}
	return remainGiftList, nil
}

// checkUsedTimes 验证用户使用次数
func (s *Service) checkUsedTimes(c context.Context, timesConfig []*l.TimesConfig, num int, usedTimes, addTimes map[string]int, highType int) (canWinTimes int, remainRecordTimes []l.TimesInterface, hasSpecialTimes bool, err error) {
	remainTimes, winTimes, remainRecordTimes, err := s.getMidLotteryTimesAndWinTimes(c, timesConfig, usedTimes, addTimes, int64(num))
	if err != nil {
		return 0, nil, false, err
	}
	if remainTimes < num {
		return 0, nil, false, ecode.ActivityNoTimes
	}
	most := s.getMostWinTimes(c, timesConfig)
	canLottery := most - winTimes
	hasSpecialTimes = s.addTimesHasSpecial(c, addTimes, highType)
	// 超过中奖次数上限
	return canLottery, remainRecordTimes, hasSpecialTimes, nil
}

// addTimesHasSpecial 用户是否获取特殊的抽奖次数
func (s *Service) addTimesHasSpecial(c context.Context, addTimesMap map[string]int, highType int) (hasSpecialTimes bool) {
	if highType == 0 {
		return false
	}
	for key, v := range addTimesMap {
		keyName := strings.Split(key, "_")
		if len(keyName) == 3 && v > 0 {
			typeInt, err := strconv.Atoi(keyName[0])
			if err != nil {
				continue
			}
			if highType == l.HightTypeArchive && typeInt == l.TimesArchiveType {
				return true
			}
			if highType == l.HightTypeBuyVip && typeInt == l.TimesBuyVipType {
				return true
			}
		}
	}
	return false
}

func (s *Service) getMidLotteryTimes(c context.Context, timesConfig []*l.TimesConfig, id int64, mid int64) (usedTimes, addTimes map[string]int, err error) {
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if usedTimes, err = s.getMidUsedTimes(c, timesConfig, id, mid); err != nil {
			return err
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if addTimes, err = s.getMidAddTimes(c, timesConfig, id, mid); err != nil {
			return err
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}
	return
}

// getMidLotteryTimes 获取用户可抽奖次数
func (s *Service) getMidLotteryTimesAndWinTimes(c context.Context, timesConfig []*l.TimesConfig, usedTimesMap, addTimesMap map[string]int, thisTimesConsume int64) (canUsedTimes, alreadyWinTimes int, remainTimes []l.TimesInterface, err error) {
	baseUsedTimes, usedOtherTimes, winTimes := s.getLotteryTimes(c, timesConfig, usedTimesMap)
	usedTimes := baseUsedTimes + usedOtherTimes
	baseAddTimes, addOtherTimes, _ := s.getLotteryTimes(c, timesConfig, addTimesMap)
	addTimes := baseAddTimes + addOtherTimes
	remainTimes, sendNum := s.buildRemainTimes(c, timesConfig, usedTimesMap, addTimesMap, thisTimesConsume)
	return addTimes - usedTimes + sendNum, winTimes + sendNum, remainTimes, nil
}

func (s *Service) buildRemainTimes(c context.Context, timesConfig []*l.TimesConfig, usedTimesMap, addTimesMap map[string]int, thisTimesConsume int64) (remainTimes []l.TimesInterface, sendNum int) {
	remainTimes = make([]l.TimesInterface, 0)
	_, _, day := s.getTodayTime()
	for _, v := range timesConfig {
		// 额外赠送次数
		if v.Type == l.TimesAdditionalType {
			info := &l.ConsumeInfo{}
			err := json.Unmarshal([]byte(v.Info), info)
			if err != nil {
				continue
			}
			if thisTimesConsume == info.Consume {
				times, err := l.GetTimesByType(&l.RecordDetail{
					Num:  info.Send,
					Type: v.Type,
					CID:  v.ID,
				})
				sendNum += info.Send
				if err != nil {
					log.Errorc(c, "l.GetTimesByType (%d) error(%v)", v.Type, err)
					continue
				}
				remainTimes = append(remainTimes, times)
			}
			continue
		}
		if v.Type == l.TimesWinType {
			continue
		}
		var key = getRecordKeyRedis(getRecordKey(v.Type, v.ID), "0")
		if v.AddType == l.DailyAddType {
			key = getRecordKeyRedis(getRecordKey(v.Type, v.ID), day)
		}

		usedTimes := usedTimesMap[key]
		addTimes := addTimesMap[key]
		if addTimes-usedTimes > 0 {
			times, err := l.GetTimesByType(&l.RecordDetail{
				Num:  addTimes - usedTimes,
				Type: v.Type,
				CID:  v.ID,
			})
			if err != nil {
				log.Errorc(c, "l.GetTimesByType (%d) error(%v)", v.Type, err)
				continue
			}
			remainTimes = append(remainTimes, times)
		}
	}
	return
}

func getRecordKey(actionType int, cid int64) string {
	return fmt.Sprintf("%d_%d", actionType, cid)
}

func getRecordKeyRedis(recordKey, day string) string {
	return fmt.Sprintf("%s_%s", recordKey, day)
}

func (s *Service) getMostWinTimes(c context.Context, timesConfig []*l.TimesConfig) (most int) {
	for _, v := range timesConfig {
		if v.Type == l.TimesWinType {
			return v.Times
		}
	}
	return
}

func (s *Service) getLotteryTimes(c context.Context, timesConfig []*l.TimesConfig, userTimes map[string]int) (baseTimes, times, winTimes int) {
	_, _, day := s.getTodayTime()
	for _, v := range timesConfig {
		if v.Type == l.TimesAdditionalType {
			continue
		}
		if v.Type == l.TimesBaseType {
			var key = getRecordKeyRedis(getRecordKey(v.Type, v.ID), "0")
			if v.AddType == l.DailyAddType {
				key = getRecordKeyRedis(getRecordKey(v.Type, v.ID), day)
			}
			baseTimes = userTimes[key]
			continue
		}
		if v.Type == l.TimesWinType {
			var key = getRecordKeyRedis(getRecordKey(v.Type, v.ID), "0")
			if v.AddType == l.DailyAddType {
				key = getRecordKeyRedis(getRecordKey(v.Type, v.ID), day)
			}
			winTimes = userTimes[key]
			continue
		}
		var key = getRecordKeyRedis(getRecordKey(v.Type, v.ID), "0")
		if v.AddType == l.DailyAddType {
			key = getRecordKeyRedis(getRecordKey(v.Type, v.ID), day)
		}
		times += userTimes[key]
	}
	return
}

// operateLotteryTimes
func (s *Service) operateLotteryTimes(c context.Context, timesConfig []*l.TimesConfig, record []*l.RecordDetail) (res map[string]int) {
	// 配置的次数
	var dayTypeTimes = make(map[string]*l.TimesConfig)
	var allActivityTypeTimes = make(map[string]*l.TimesConfig)
	var winKey string
	var winDaily bool
	for _, v := range timesConfig {
		if v.Type == l.TimesWinType {
			winKey = getRecordKey(v.Type, v.ID)
			if v.AddType == l.DailyAddType {
				winDaily = true
			}
			continue
		}
		if v.AddType == l.DailyAddType {
			dayTypeTimes[getRecordKey(v.Type, v.ID)] = v
			continue
		}
		allActivityTypeTimes[getRecordKey(v.Type, v.ID)] = v
	}
	res = make(map[string]int)
	_, _, today := s.getTodayTime()
	var dayWinTimes int
	var allWinTimes int

	for _, v := range record {
		if v.GiftID > 0 {
			allWinTimes++
		}
		key := getRecordKey(v.Type, v.CID)
		day := v.Ctime.Time().Format("2006-01-02")
		if day == today && v.GiftID > 0 {
			dayWinTimes++
		}
		// 处理每日过期次数
		if _, ok := dayTypeTimes[key]; ok {
			day := v.Ctime.Time().Format("2006-01-02")
			newKey := getRecordKeyRedis(key, day)
			res[newKey] = res[newKey] + v.Num
			continue
		}
		if _, ok := allActivityTypeTimes[key]; ok {
			newKey := getRecordKeyRedis(key, "0")
			res[newKey] = res[newKey] + v.Num
			continue
		}
	}
	newWinKey := getRecordKeyRedis(winKey, "0")
	res[newWinKey] = allWinTimes
	if winDaily {
		newWinKey := getRecordKeyRedis(winKey, today)
		res[newWinKey] = dayWinTimes
	}
	return
}

func (s *Service) getTodayTime() (xtime.Time, xtime.Time, string) {
	nowT := time.Now().Format("2006-01-02")
	timeTemplate := "2006-01-02 15:04:05"
	start, _ := time.ParseInLocation(timeTemplate, nowT+" 00:00:00", time.Local)
	dayStart := start.Unix()
	end, _ := time.ParseInLocation(timeTemplate, nowT+" 23:59:59", time.Local)
	dayEnd := end.Unix()
	return xtime.Time(dayStart), xtime.Time(dayEnd), nowT
}

// getMidUsedTimes 获取用户使用次数
func (s *Service) getMidUsedTimes(c context.Context, lotteryTimesConfig []*l.TimesConfig, id int64, mid int64) (recordMap map[string]int, err error) {
	recordMap, err = s.lottery.CacheLotteryTimes(c, id, mid, l.UsedTimesKey)
	if err != nil {
		log.Errorc(c, "s.lottery.CacheLotteryTimes(%d,%d,%v,%s) error(%v)", id, mid, lotteryTimesConfig, l.UsedTimesKey, err)
	}
	if recordMap != nil {
		cache.MetricHits.Inc("LotteryUsedTimes")
		return
	}
	cache.MetricMisses.Inc("LotteryUsedTimes")
	res, err := s.lottery.RawLotteryUsedTimes(c, id, mid)
	if err != nil {
		return
	}
	recordMap = s.operateLotteryTimes(c, lotteryTimesConfig, res)
	s.cache.Do(c, func(ctx context.Context) {
		s.lottery.AddCacheLotteryTimes(ctx, id, mid, l.UsedTimesKey, recordMap)
	})
	return
}

// getMidAddTimes 获取增加的次数
func (s *Service) getMidAddTimes(c context.Context, lotteryTimesConfig []*l.TimesConfig, id int64, mid int64) (recordMap map[string]int, err error) {
	var record = make([]*l.RecordDetail, 0)
	var now = xtime.Time(time.Now().Unix())
	// 增加基础抽奖次数
	_, _, day := s.getTodayTime()
	var key string
	var num int
	for _, v := range lotteryTimesConfig {
		if v.Type == l.TimesBaseType {
			record = append(record, &l.RecordDetail{
				ID:    v.ID,
				Mid:   mid,
				Num:   v.Times,
				Type:  v.Type,
				CID:   v.ID,
				Ctime: now,
			})
			num = v.Times
			key = getRecordKeyRedis(getRecordKey(v.Type, v.ID), "0")
			if v.AddType == l.DailyAddType {
				key = getRecordKeyRedis(getRecordKey(v.Type, v.ID), day)
			}

		}
	}

	recordMap, err = s.lottery.CacheLotteryTimes(c, id, mid, l.AddTimesKey)
	if err != nil {
		log.Errorc(c, "s.lottery.CacheLotteryTimes(%d,%d,%s) error(%v)", id, mid, l.AddTimesKey, err)
	}
	if recordMap != nil {
		if record, ok := recordMap[key]; ok && record == num {
			cache.MetricHits.Inc("LotteryAddTimes")
			return
		}
	}
	cache.MetricMisses.Inc("LotteryAddTimes")
	addTimes, err := s.lottery.RawLotteryAddTimes(c, id, mid)
	if err != nil {
		return
	}

	if addTimes != nil {
		for _, v := range addTimes {
			record = append(record, &l.RecordDetail{
				ID:    v.ID,
				Mid:   v.Mid,
				Num:   v.Num,
				Type:  v.Type,
				Ctime: v.Ctime,
				CID:   v.CID,
			})
		}
	}
	recordMap = s.operateLotteryTimes(c, lotteryTimesConfig, record)
	s.cache.Do(c, func(c context.Context) {
		s.lottery.AddCacheLotteryTimes(c, id, mid, l.AddTimesKey, recordMap)
	})
	return recordMap, nil
}

// checkLottery 检查抽奖基础信息
func (s *Service) checkLottery(c context.Context, lottery *l.Lottery, info *l.Info, config []*l.TimesConfig, gift []*l.Gift, now int64, isInternal bool) (err error) {
	if !isInternal && lottery.IsInternal == l.IsInternal {
		return ecode.ActivityLotteryIsInternalError
	}
	err = s.checkTimesConf(c, lottery, config)
	if err != nil {
		return err
	}
	if info.ID == 0 || len(config) == 0 || len(gift) == 0 {
		return ecode.ActivityNotConfig
	}
	return nil
}

// checkTimesConf
func (s *Service) checkTimesConf(c context.Context, lottery *l.Lottery, timesConf []*l.TimesConfig) error {
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

// getBaseLotteryInfo 获取基础抽奖信息
func (s *Service) getBaseLotteryInfo(c context.Context, sid string) (lottery *l.Lottery, info *l.Info, timesConf []*l.TimesConfig, gift []*l.Gift, memberGroup map[int64]*l.MemberGroup, err error) {
	// 获取基础信息
	// eg.Go(func(ctx context.Context) (err error) {
	lottery, err = s.base(c, sid)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		return
	}
	// 获取相关配置
	info, err = s.lotteryInfo(c, sid)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.LotteryInfo %s", sid)
		return
	}
	// 获取次数配置
	timesConf, err = s.timesConfigAndSort(c, sid)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.LotteryTimesConfig %s", sid)
		return
	}
	// 获取商品信息
	gift, err = s.gift(c, sid)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.LotteryGift %s", sid)
		return
	}
	// 获取用户组信息
	memberGroup, err = s.getMemberGroupMap(c, sid)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.MemberGroup %s", sid)
		return
	}
	return
}

// riskMember 风险用户
func (s *Service) riskMember(c context.Context, lottery *l.Lottery, mid int64, risk *riskmdl.Base) error {
	if risk != nil {
		lotteryRisk := &riskmdl.Lottery{
			Base:     *risk,
			RecordID: lottery.ID,
			RcLevel:  1,
			SID:      lottery.LotteryID,
			Author:   lottery.Author,
			Name:     lottery.Name,
			Stime:    lottery.Stime.Time().Format("2006-01-02 15:04:05"),
			Etime:    lottery.Etime.Time().Format("2006-01-02 15:04:05"),
		}
		var bs []byte
		bs, _ = json.Marshal(lotteryRisk)
		riskReply, err := client.SilverbulletClient.RuleCheck(c, &silverbulletapi.RuleCheckReq{
			Scene:    risk.Action,
			EventCtx: string(bs),
			EventTs:  int64(risk.EsTime),
		})
		if err != nil || riskReply == nil || riskReply.Decisions == nil {
			err = errors.Wrapf(err, "s.silverbulletClient.RuleCheck %d", mid)
			log.Errorc(c, "s.silverbulletClient.RuleCheck(%d) riskReply(%v) error(%v)", mid, riskReply, err)
			return nil
		}
		// 风险命中判断
		if riskReply == nil || riskReply.Decisions == nil || len(riskReply.Decisions) == 0 {
			return nil
		}
		riskMap := riskReply.Decisions[0]
		if riskMap == riskmdl.Reject {
			return ecode.ActivityLotteryRiskInfo
		}
		return nil
	}
	log.Infoc(c, "risk params is nil,none riskMember(%d,%v)", mid, &risk)
	return nil
}

// judgeUser judge user could lottery or not .
func (s *Service) checkMemberLottery(c context.Context, info *l.Info, member *l.MemberInfo) (err error) {
	if err = member.IsSilence(); err != nil {
		return
	}
	// 账号等级限制
	if info.Level != 0 {
		if err = member.LevelLimit(info.Level); err != nil {
			return
		}
	}
	// 注册时间限制
	if info.RegTimeStime != 0 {
		if err = member.RegStimeLimit(info.RegTimeStime); err != nil {
			return
		}
	}
	if info.RegTimeEtime != 0 {
		if err = member.RegEtimeLimit(info.RegTimeEtime); err != nil {
			return
		}
	}
	// vip限制
	if err = member.VipCheck(info.VipCheck); err != nil {
		return
	}
	// 账号验证
	if err = member.AccountCheck(info.AccountCheck); err != nil {
		return
	}
	// IP防刷
	if info.FsIP == 1 {
		var used int
		if used, err = s.lottery.CacheIPRequestCheck(c, member.IP); err != nil {
			log.Errorc(c, "s.dao.GetIPRequestCheck(%s) error(%v)", member.IP, err)
			return
		}
		if used != 0 {
			err = ecode.ActivityLotteryIPFrequence
			return
		}
		if err = s.lottery.AddCacheIPRequestCheck(c, member.IP, 1); err != nil {
			log.Errorc(c, "s.dao.SetIPRequestCheck(%s, %d) error(%v)", member.IP, 1, err)
			return
		}
	}
	return
}

// 判断ip是否海外
func (s *Service) judgeIPLocationgo(c context.Context, ip string) (ok bool, err error) {
	ipInfo, err := s.locationClient.InfoComplete(c, &locationAPI.InfoCompleteReq{Addr: ip})
	if err != nil {
		log.Errorc(c, "JudgeIpLocation s.locationRPC.InfoComplete ip(%v) error(%v)", ip, err)
		return
	}
	if ipInfo != nil && ipInfo.Info != nil {
		if ipInfo.Info.Country == "中国" {
			ok = true
		}
	}
	return
}

// getMemberInfo 用户信息
func (s *Service) getMemberInfo(c context.Context, mid int64, ip string, buvid string) (lotteryMember *l.MemberInfo, err error) {
	lotteryMember = &l.MemberInfo{}
	var (
		memberRly     *accapi.ProfileReply
		infoReply     *vipinfoapi.InfoReply
		figureReply   *figure.UserFigureReply
		userInfoReply *spy.UserInfoReply
		ipOk          bool
		percentage    int32
		spyScore      int32
	)
	eg := errgroup.WithContext(c)
	// 用户信息
	eg.Go(func(ctx context.Context) (err error) {
		if memberRly, err = s.accClient.Profile3(c, &accapi.MidReq{Mid: mid}); err != nil || memberRly == nil || memberRly.Profile == nil {
			err = errors.Wrapf(err, "s.accRPC.Profile3(c,&accmdl.ArgMid{Mid:%d})", mid)
			return
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if infoReply, err = s.vipInfoClient.Info(c, &vipinfoapi.InfoReq{Mid: mid}); err != nil || infoReply == nil || infoReply.Res == nil {
			err = errors.Wrapf(err, "s.vipInfoClient.Info(c,&vipinfoapi.InfoReq{Mid:%d})", mid)
			return
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if ipOk, err = s.judgeIPLocationgo(c, ip); err != nil {
			log.Errorc(c, "judgeIPLocationgo(c, %s) err(%v)", ip, err)
			err = nil
			return
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		figureReply, err = s.figureClient.UserFigure(ctx, &figure.UserFigureReq{MID: mid})
		if err != nil || figureReply == nil {
			err = errors.Wrapf(err, "s.figureClient.UserFigure(%d)", mid)
			return nil
		}
		percentage = figureReply.Percentage
		log.Infoc(c, "lottery user figureReply score mid(%d), percentage(%d)", mid, percentage)

		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		userInfoReply, err = s.spyClient.UserInfo(ctx, &spy.UserInfoReq{Mid: mid})
		if err != nil || userInfoReply.Ui == nil {
			err = errors.Wrapf(err, "s.spyClient.UserInfo(%d)", mid)
			return nil
		}
		spyScore = userInfoReply.Ui.Score
		log.Infoc(c, "lottery user spy score mid(%d), score(%d)", mid, spyScore)
		return
	})

	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}
	lotteryMember = &l.MemberInfo{}
	profile := memberRly.Profile
	vipInfo := infoReply.Res
	lotteryMember = &l.MemberInfo{
		Mid:            mid,
		Silence:        profile.Silence,
		Name:           profile.Name,
		Level:          profile.Level,
		JoinTime:       profile.JoinTime,
		VipType:        profile.Vip.Type,
		VipStatus:      profile.Vip.Status,
		TelStatus:      profile.TelStatus,
		Identification: profile.Identification,
		NeverVip:       vipInfo.IsNeverVip(),
		MonthVip:       vipInfo.IsMonth(),
		AnnualVip:      vipInfo.IsAnnual(),
		IP:             ip,
		ValidIP:        ipOk,
		SpyScore:       spyScore,
		Percentage:     percentage,
		DeviceID:       buvid,
	}
	return
}

// lottery 抽奖信息
func (s *Service) base(c context.Context, sid string) (res *l.Lottery, err error) {
	res, err = s.lottery.CacheLottery(c, sid)
	if err != nil {
		log.Errorc(c, "s.lottery.CacheLottery(%s) err(%v)", sid, err)
	}
	if res != nil && err == nil {
		cache.MetricHits.Inc("Lottery")
		return res, nil
	}
	cache.MetricMisses.Inc("Lottery")
	res, err = s.lottery.RawLottery(c, sid)
	if err != nil {
		return
	}
	s.cache.Do(c, func(c context.Context) {
		s.lottery.AddCacheLottery(c, sid, res)
	})
	return
}

// lotteryInfo get data from cache if miss will call source method, then add to cache.
func (s *Service) lotteryInfo(c context.Context, sid string) (res *l.Info, err error) {
	res, err = s.lottery.CacheLotteryInfo(c, sid)
	if err != nil {
		log.Errorc(c, "lotteryInfo s.lottery.CacheLotteryInfo(%v) error(%v)", sid, err)
	}
	if res != nil && err == nil {
		cache.MetricHits.Inc("LotteryInfo")
		return
	}
	cache.MetricMisses.Inc("LotteryInfo")
	res, err = s.lottery.RawLotteryInfo(c, sid)
	if err != nil {
		return
	}
	s.cache.Do(c, func(c context.Context) {
		s.lottery.AddCacheLotteryInfo(c, sid, res)
	})
	return
}

func (s *Service) timesConfigAndSort(c context.Context, sid string) (res []*l.TimesConfig, err error) {
	times, err := s.timesConfig(c, sid)
	res = make([]*l.TimesConfig, 0)
	sortInt := []int{l.TimesAdditionalType, l.TimesBaseType, l.TimesWinType, l.TimesShareType, l.TimesFollowType, l.TimesArchiveType, l.TimesBuyVipType, l.TimesOtherType, l.TimesCustomizeType, l.TimesOGVType, l.TimesFeType, l.TimesLikeType, l.TimesCoinType, l.TimesActType, l.TimesActPointType}
	for _, v := range sortInt {
		for _, t := range times {
			if t.Type == v {
				res = append(res, t)
			}
		}
	}
	return res, nil
}

// timesConfig get data from cache if miss will call source method, then add to cache.
func (s *Service) timesConfig(c context.Context, sid string) (res []*l.TimesConfig, err error) {
	res, err = s.lottery.CacheLotteryTimesConfig(c, sid)
	if err != nil {
		log.Errorc(c, "lotteryInfo s.lottery.CacheLotteryTimesConfig(%v) error(%v)", sid, err)
	}
	if res != nil && err == nil {
		cache.MetricHits.Inc("LotteryTimesConfig")
		return
	}
	cache.MetricMisses.Inc("LotteryTimesConfig")
	res, err = s.lottery.RawLotteryTimesConfig(c, sid)
	if err != nil || len(res) == 0 {
		return
	}
	s.cache.Do(c, func(c context.Context) {
		s.lottery.AddCacheLotteryTimesConfig(c, sid, res)
	})
	return
}

// Gift ...
func (s *Service) Gift(c context.Context, sid string) (res []*l.Gift, err error) {

	return s.gift(c, sid)
}

// Gift ...
func (s *Service) GiftRes(c context.Context, sid string) (res []*l.GiftRes, err error) {
	res = make([]*l.GiftRes, 0)
	gift, err := s.gift(c, sid)
	if err != nil {
		return
	}
	for _, v := range gift {
		if v.Efficient == l.Efficient {
			res = append(res, &l.GiftRes{
				ID:     v.ID,
				Sid:    v.Sid,
				ImgURL: v.ImgURL,
				Type:   v.Type,
				Name:   v.Name,
			})
		}
	}
	return
}

func (s *Service) gift(c context.Context, sid string) (res []*l.Gift, err error) {
	res = make([]*l.Gift, 0)
	res, err = s.lottery.CacheLotteryGift(c, sid)
	if err != nil {
		err = nil
	}
	if len(res) != 0 {
		cache.MetricHits.Inc("LotteryGift")
		return
	}
	cache.MetricMisses.Inc("LotteryGift")
	resDao, err := s.lottery.RawLotteryGift(c, sid)

	if err != nil || len(resDao) == 0 {
		return
	}
	res, err = s.dbGiftTurnToGift(c, resDao)
	if err != nil {
		return
	}
	s.cache.Do(c, func(c context.Context) {
		s.lottery.AddCacheLotteryGift(c, sid, res)
	})
	return
}

// dbGiftTurnToGift
func (s *Service) dbGiftTurnToGift(c context.Context, giftDB []*l.GiftDB) (res []*l.Gift, err error) {
	res = make([]*l.Gift, 0)
	for _, v := range giftDB {
		dayNum := make(map[string]int64)
		err := json.Unmarshal([]byte(v.DayNum), &dayNum)
		if err != nil {
			log.Errorc(c, "dbGiftTurnToGift dayNum (%s) can not turn to int error(%v)", v.DayNum, err)
		}
		memberGroupStr := strings.Split(v.MemberGroup, ",")
		var memberGroup = make([]int64, 0)
		if len(memberGroupStr) > 0 && memberGroupStr[0] != "" {
			for _, v := range memberGroupStr {
				group, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					log.Errorc(c, "dbGiftTurnToGift groupMember str(%s) can not turn to int error(%v)", v, err)
					continue
				}
				memberGroup = append(memberGroup, group)
			}
		}
		extra := make(map[string]string)
		if v.Extra != "" {
			err := json.Unmarshal([]byte(v.Extra), &extra)
			if err != nil {
				log.Errorc(c, "v.Extra (%s)json error %v", v.Extra, err)
				continue
			}
		}
		gift := &l.Gift{
			ID:             v.ID,
			Sid:            v.Sid,
			Name:           v.Name,
			Num:            v.Num,
			Type:           v.Type,
			Source:         v.Source,
			ImgURL:         v.ImgURL,
			IsShow:         v.IsShow,
			LeastMark:      v.LeastMark,
			MessageTitle:   v.MessageTitle,
			MessageContent: v.MessageContent,
			SendNum:        v.SendNum,
			Efficient:      v.Efficient,
			State:          v.State,
			MemberGroup:    memberGroup,
			DayNum:         dayNum,
			Probability:    v.Probability,
			Params:         v.Params,
			Extra:          extra,
		}
		res = append(res, gift)
	}
	return
}

// buildAddressLink 建立抽奖地址
func (s *Service) buildAddressLink(ctx context.Context, lotteryID string, giftType int, giftID int64) string {
	if giftType == giftEntity {
		return fmt.Sprintf("%s%s", s.c.Lottery.AddressLink, lotteryID)
	}
	if giftType == giftCoupon {
		return fmt.Sprintf("%s%s&gift_id=%d", s.c.Lottery.CouponLink, lotteryID, giftID)

	}
	return ""
}

func (s *Service) buildAddressText(ctx context.Context, giftType int) string {
	if giftType == giftEntity {
		return s.c.Lottery.AddressText
	}
	if giftType == giftCoupon {
		return s.c.Lottery.CouponText

	}
	return ""
}

func (s *Service) getGiftString(c context.Context, giftName, giftOther string) string {
	if giftOther == "" {
		return giftName
	}
	return fmt.Sprintf("%s,%s", giftName, giftOther)
}
func (s *Service) sendSysMsg(c context.Context, mid int64, lottery *l.Lottery, info *l.Info, gift *l.Gift, senderID int64, giftOther string) (err error) {
	if senderID > 0 && gift != nil {
		params := make([]string, 0)
		params = append(params, lottery.Name, s.getGiftString(c, gift.Name, giftOther))
		letterParasm := &l.LetterParam{
			RecverIDs:  []uint64{uint64(mid)},
			SenderUID:  uint64(senderID),
			MsgType:    l.MsgTypeCard,          //通知卡类型 type = 10
			NotifyCode: s.c.Lottery.NotifyCode, //通知码
			Params:     strings.Join(params, paramsSplit),
			Title:      s.c.Lottery.MessageTitle,
		}
		if info.ActivityLink != "" {
			letterParasm.JumpText = s.c.Lottery.ActivityText
			letterParasm.JumpURL = info.ActivityLink
		}
		addressURL := s.buildAddressLink(c, lottery.LotteryID, gift.Type, gift.ID)
		if addressURL != "" {
			letterParasm.JumpText2 = s.buildAddressText(c, gift.Type)
			letterParasm.JumpURL2 = addressURL
		}
		_, err = s.lottery.SendLetter(c, letterParasm)
	}

	return
}
