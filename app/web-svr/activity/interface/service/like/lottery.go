package like

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	l "go-gateway/app/web-svr/activity/interface/model/lottery"
	lott "go-gateway/app/web-svr/activity/interface/model/lottery"
	suitmdl "go-main/app/account/usersuit/service/api"

	"go-common/library/sync/errgroup.v2"

	actplat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"

	figure "git.bilibili.co/bapis/bapis-go/account/service/figure"
	spy "git.bilibili.co/bapis/bapis-go/account/service/spy"
	locationAPI "git.bilibili.co/bapis/bapis-go/community/service/location"
	vipact "git.bilibili.co/bapis/bapis-go/vip/service/activity"
	"github.com/pkg/errors"
)

const (
	_base          = 1
	_win           = 2
	_share         = 3
	_follow        = 4
	_archive       = 5
	_buy           = 6
	_other         = 7
	_customize     = 8
	_addTimes      = "add"
	_usedTimes     = "used"
	_lottery       = "抽奖消耗"
	_telValid      = 1
	_identifyValid = 2
	_vipCheck      = 1
	_monthVip      = 2
	_yearVip       = 3
	_entity        = 1
	_member        = 2
	_grant         = 3
	_coupon        = 4
	_coin          = 5
	_memberCoupon  = 6
	_otherGift     = 7
	_mc            = "1_4_1"
	_remark        = "活动抽奖所得"
	_remarkCoin    = "抽奖获得硬币"
	_vipPriority   = 1
	_likePriority  = 2
	_mustWin       = 1
	_ogv           = 9
	_fe            = 10
	_timeslike     = 11
	_timescoin     = 12
	paramsSplit    = "`||"
	caller         = "activity_lottery"
)

var lockKey = "lottery:lock:%d:%d"

// 抽奖主接口
func (s *Service) DoLottery(c context.Context, sid string, mid int64, num int, isInternal bool) (res []*l.LotteryRecordDetail, err error) {
	if _, ok := s.c.Fission.Sids[sid]; ok {
		err = xecode.AccessDenied
		log.Warn("DoLottery forbid fission sid:%s mid:%d", sid, mid)
		return
	}
	if !isInternal {
		if _, ok := s.internalLottSids[sid]; ok {
			err = xecode.AccessDenied
			log.Warn("DoLottery forbid internal sid:%s mid:%d", sid, mid)
			return
		}
	}
	var (
		lottery                                                                 *l.Lottery
		lotteryInfo                                                             *l.LotteryInfo
		lotteryTimesConf                                                        []*l.LotteryTimesConfig
		lotteryGift                                                             []*l.LotteryGift
		actionLog                                                               []*l.LotteryRecordDetail
		now                                                                     = time.Now().Unix()
		memberRly                                                               *accapi.ProfileReply
		ip                                                                      = metadata.String(c, metadata.RemoteIP)
		figureReply                                                             *figure.UserFigureReply
		userInfoReply                                                           *spy.UserInfoReply
		rate, base, win, share, follow, other, lastID, fe, timesLike, timesCoin int64
		isHigh, canLottery                                                      int
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if lottery, err = s.lottDao.Lottery(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if lotteryInfo, err = s.lottDao.LotteryInfo(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryInfo %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if lotteryTimesConf, err = s.lottDao.LotteryTimesConfig(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryTimesConfig %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if lotteryGift, err = s.lottDao.LotteryGift(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryGift %s", sid)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	if err = checkLottery(lottery, lotteryInfo, lotteryTimesConf, lotteryGift, now); err != nil {
		log.Error("DoLottery checkLottery lotteryID(%d) error(%v)", lottery.ID, err)
		return
	}
	if memberRly, err = s.accClient.Profile3(c, &accapi.MidReq{Mid: mid}); err != nil || memberRly == nil {
		log.Error("DoLottery s.accRPC.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%v)", mid, err)
		return
	}
	// 账号验证
	if err = s.judgeUserLottery(c, lotteryInfo, memberRly.Profile, ip); err != nil {
		log.Error("DoLottery s.judgeUserLottery lotteryInfo(%v) profile(%v) error(%v)", lotteryInfo, memberRly.Profile, err)
		return
	}
	if isInternal {
		orderNo := strconv.FormatInt(mid, 10) + strconv.FormatInt(now, 10)
		s.JudgeAddTimes(c, _base, 0, lottery.ID, mid, 0, lotteryTimesConf, orderNo, ip)
	}
	likeList := make([]int64, 0)
	buyList := make([]int64, 0)
	customizeList := make([]int64, 0)
	ogvList := make([]int64, 0)
	lotteryMap := make(map[string]*l.LotteryTimesConfig, len(lotteryTimesConf))
	for _, l := range lotteryTimesConf {
		switch l.Type {
		case _base:
			base = l.ID
		case _win:
			win = l.ID
		case _share:
			share = l.ID
		case _follow:
			follow = l.ID
		case _other:
			other = l.ID
		case _archive:
			likeList = append(likeList, l.ID)
		case _buy:
			buyList = append(buyList, l.ID)
		case _customize:
			customizeList = append(customizeList, l.ID)
		case _ogv:
			ogvList = append(ogvList, l.ID)
		case _fe:
			fe = l.ID
		case _timeslike:
			timesLike = l.ID
		case _timescoin:
			timesCoin = l.ID
		default:
			err = ecode.ActivityNotExist
			return
		}
		lotteryMap[rsKey(l.Type, l.ID)] = l
	}
	usedMap := make(map[string]int, len(lotteryTimesConf))
	addMap := make(map[string]int, len(lotteryTimesConf))
	all := 0
	used := 0
	egp := errgroup.WithContext(c)
	egp.Go(func(ctx context.Context) (err error) {
		// 使用次数集合
		if usedMap, err = s.lottDao.LotteryUsedTimes(ctx, lotteryTimesConf, lottery.ID, mid); err != nil {
			return
		}
		for k, v := range usedMap {
			if k == rsKey(_win, win) {
				continue
			}
			used = used + v
		}
		return
	})
	egp.Go(func(ctx context.Context) (err error) {
		// 增加次数集合
		if addMap, err = s.lottDao.LotteryAddTimes(ctx, lotteryTimesConf, lottery.ID, mid); err != nil {
			return
		}
		for k, v := range addMap {
			if k == rsKey(_win, win) {
				continue
			}
			all = all + v
		}
		return
	})
	if err = egp.Wait(); err != nil {
		log.Error("egp.Wait error(%v)", err)
		return
	}
	// 无抽奖次数
	if all-used < num {
		err = ecode.ActivityNoTimes
		return
	}
	if sid == s.c.Stupid.LotterySid {
		if _, e := s.dao.IncrWithExpire(c, fmt.Sprintf(lockKey, lottery.ID, mid), s.c.Stupid.LockExpire); e != nil {
			log.Error("Stupid failed to incr: %s, error: %+v", fmt.Sprintf(lockKey, lottery.ID, mid), e)
		}
	}
	consumeArg := &l.ConsumeArg{
		Sid:           lottery.ID,
		Mid:           mid,
		AddMap:        addMap,
		UsedMap:       usedMap,
		LotteryMap:    lotteryMap,
		Num:           num,
		Base:          base,
		Share:         share,
		Follow:        follow,
		Other:         other,
		LikeList:      likeList,
		BuyList:       buyList,
		CustomizeList: customizeList,
		OgvList:       ogvList,
		Fe:            fe,
		TimesLike:     timesLike,
		TimesCoin:     timesCoin,
	}
	if lotteryMap[rsKey(_win, win)] != nil {
		canLottery = lotteryMap[rsKey(_win, win)].Most - usedMap[rsKey(_win, win)] - 1
		// 超过中奖次数上限
		if canLottery <= 0 {
			res, err = s.lotteryWithoutCount(c, lottery, consumeArg, lotteryInfo.Coin, now, ip)
			return
		}
	}
	rate = lotteryInfo.GiftRate
	// 高优先级判断
	switch lotteryInfo.HighType {
	case _vipPriority: // 购买大会员
		for _, val := range buyList {
			if addMap[rsKey(_buy, val)] > 0 {
				rate = lotteryInfo.HighRate
				isHigh = 1
				break
			}
		}
	case _likePriority: // 投稿
		for _, val := range likeList {
			if addMap[rsKey(_archive, val)] > 0 {
				rate = lotteryInfo.HighRate
				isHigh = 1
				break
			}
		}
	}
	if rate != _mustWin {
		egroup := errgroup.WithContext(c)
		egroup.Go(func(ctx context.Context) (err error) {
			figureReply, _ = s.figureClient.UserFigure(ctx, &figure.UserFigureReq{MID: mid})
			return
		})
		egroup.Go(func(ctx context.Context) (err error) {
			userInfoReply, err = s.spyClient.UserInfo(ctx, &spy.UserInfoReq{Mid: mid})
			return
		})
		if err = egroup.Wait(); err != nil {
			log.Error("egroup.Wait error(%v)", err)
			return
		}
		// 真实分与信用分限制
		if userInfoReply.Ui == nil || figureReply == nil || userInfoReply.Ui.Score <= s.c.Lottery.SpyScore || figureReply.Percentage >= s.c.Lottery.FigerScore {
			if userInfoReply.Ui != nil && figureReply != nil {
				log.Warn("DoLottery: sid(%d) mid(%d) spy(%d) figure(%d) LawfulScore(%d) WideScore(%d) FriendlyScore(%d) BountyScore(%d) "+
					"CreativityScore(%d)", lottery.ID, mid, userInfoReply.Ui.Score, figureReply.Percentage, figureReply.LawfulScore, figureReply.WideScore, figureReply.FriendlyScore, figureReply.BountyScore, figureReply.CreativityScore)
			} else {
				log.Warn("DoLottery: sid(%d) mid(%d) spy or figure not found", lottery.ID, mid)
			}
			res, err = s.lotteryWithoutCount(c, lottery, consumeArg, lotteryInfo.Coin, now, ip)
			return
		}
	}
	// 硬币
	if lotteryInfo.Coin > 0 {
		if _, err = s.coinClient.ModifyCoins(c, &coinmdl.ModifyCoinsReq{Mid: mid, Count: float64(-lotteryInfo.Coin * num), Reason: _lottery, IP: ip}); err != nil {
			err = ecode.ActivityNotEnoughCoin
			return
		}
	}
	// 消耗抽奖次数
	insertRecord := make([]*l.InsertRecord, 0, num)
	if insertRecord, err = s.consumeLotteryTimes(c, consumeArg); err != nil {
		err = ecode.ActivityLotteryTimesFail
		return
	}
	// 奖品slice转map
	giftMap := make(map[int64]*l.LotteryGift)
	for _, lg := range lotteryGift {
		if lg.Efficient == 1 {
			giftMap[lg.ID] = lg
		}
	}
	lotteryArg := &l.LottArg{
		Sid:          lottery.ID,
		Ip:           ip,
		Mid:          mid,
		Rate:         rate,
		Lottery:      lottery,
		LotteryMap:   lotteryMap,
		GiftMap:      giftMap,
		InsertRecord: insertRecord,
		High:         isHigh,
		CanLottery:   canLottery,
		Win:          win,
	}
	gid := make([]int64, 0, num)
	// 抽奖

	if gid, err = s.Immediate(c, lotteryArg); err != nil {
		log.Error("DoLottery: s.Immediate sid(%s) mid(%d) lottery(%v) lotteryGift(%v) error(%v)", sid, mid, lottery, lotteryGift, err)
		return
	}
	// 新增记录mysql
	if int64(lottery.Ctime) > s.c.Lottery.NewLotteryTime {
		if lastID, err = s.lottDao.InsertLotteryRecardOrderNo(c, lottery.ID, insertRecord, gid, ip); err != nil {
			log.Error("DoLottery: s.dao.InsertLotteryRecard lottery.ID(%d) arg(%v) gid(%v) ip(%s) error(%v)", lottery.ID, insertRecord, gid, ip, err)
			return
		}
	} else {
		if lastID, err = s.lottDao.InsertLotteryRecard(c, lottery.ID, insertRecord, gid, ip); err != nil {
			log.Error("DoLottery: s.dao.InsertLotteryRecard lottery.ID(%d) arg(%v) gid(%v) ip(%s) error(%v)", lottery.ID, insertRecord, gid, ip, err)
			return
		}
	}

	for i, v := range gid {
		tmp := new(l.LotteryRecordDetail)
		lrd := &l.LotteryRecordDetail{Mid: mid, Num: 1, GiftID: 0, GiftName: fmt.Sprintf("未中奖%d", i), ImgURL: "", Type: -1}
		// 中奖
		if v != 0 {
			gm, ok := giftMap[v]
			if !ok || gm == nil {
				log.Warn("giftMap not found obj(%v)", giftMap)
				continue
			}
			lrd.GiftID = v
			lrd.GiftName = gm.Name
			lrd.ImgURL = gm.ImgUrl
			lrd.Type = gm.Type
		}
		lrd.Ctime = xtime.Time(now + int64(i))
		res = append(res, lrd)
		*tmp = *lrd
		tmp.Type = insertRecord[i].Type
		tmp.ID = lastID
		actionLog = append(actionLog, tmp)
	}
	//新增抽奖记录redis
	if err = s.lottDao.AddLotteryActionLog(c, lottery.ID, mid, actionLog); err != nil {
		log.Error("DoLottery s.dao.AddLotteryActionLog sid(%d) mid(%d) arg(%v) error(%v)", lottery.ID, mid, actionLog, err)
	}
	s.cache.Do(c, func(ctx context.Context) {
		s.sendLotteryAward(ctx, giftMap, gid, lottery.ID, mid, ip, lottery, lotteryInfo)
	})
	return
}

func (s *Service) lotteryOperation(c context.Context, lottery *l.Lottery, now int64, consumeArg *l.ConsumeArg, ip string) (res []*l.LotteryRecordDetail, err error) {
	var (
		lastID    int64
		actionLog []*l.LotteryRecordDetail
	)
	// 消耗抽奖次数
	insertRecord := make([]*l.InsertRecord, 0)
	if insertRecord, err = s.consumeLotteryTimes(c, consumeArg); err != nil {
		err = ecode.ActivityLotteryTimesFail
		return
	}
	// 新增记录
	gid := make([]int64, len(insertRecord))
	// 新增记录mysql
	if int64(lottery.Ctime) > s.c.Lottery.NewLotteryTime {
		if lastID, err = s.lottDao.InsertLotteryRecardOrderNo(c, lottery.ID, insertRecord, gid, ip); err != nil {
			log.Error("DoLottery: s.dao.InsertLotteryRecard lottery.ID(%d) arg(%v) gid(%v) ip(%s) error(%v)", lottery.ID, insertRecord, gid, ip, err)
			return
		}
	} else {
		if lastID, err = s.lottDao.InsertLotteryRecard(c, consumeArg.Sid, insertRecord, gid, ip); err != nil {
			log.Error("DoLottery: s.dao.InsertLotteryRecard lottery.ID(%d) arg(%v) gid(%v) ip(%s) error(%v)", consumeArg.Sid, insertRecord, gid, ip, err)
			return
		}
	}

	for i := 0; i < consumeArg.Num; i++ {
		tmp := new(l.LotteryRecordDetail)
		lrd := &l.LotteryRecordDetail{Mid: consumeArg.Mid, Num: 1, GiftID: 0, Type: -1, ImgURL: ""}
		lrd.GiftName = "未中奖" + strconv.Itoa(i)
		lrd.Ctime = xtime.Time(now + int64(i))
		res = append(res, lrd)
		*tmp = *lrd
		tmp.Type = insertRecord[i].Type
		tmp.ID = lastID
		actionLog = append(actionLog, tmp)
	}
	//新增抽奖记录redis
	if err = s.lottDao.AddLotteryActionLog(c, consumeArg.Sid, consumeArg.Mid, actionLog); err != nil {
		log.Error("DoLottery s.dao.AddLotteryActionLog sid(%d) mid:(%d) arg(%v) error(%v)", consumeArg.Sid, consumeArg.Mid, actionLog, err)
	}
	return
}

// 即时抽奖
func (s *Service) Immediate(c context.Context, arg *l.LottArg) (gift []int64, err error) {
	var (
		nowNum    int
		leastGift = new(l.LotteryGift)
		giftList  = make([]int64, 0)
	)
	if len(arg.GiftMap) == 0 { // 无进入奖池奖品，未中奖
		for i := 0; i < len(arg.InsertRecord); i++ {
			gift = append(gift, 0)
		}
		return
	}
	total := int64(0)
	randList := make([]int64, 0)
	for k, v := range arg.GiftMap {
		if v.LeastMark == 1 {
			leastGift = v
			continue
		}
		if v.Num-v.SendNum > 0 {
			giftList = append(giftList, k)
			total = total + v.Num - v.SendNum //剩余数量计算概率
			randList = append(randList, total)
		}
	}
	if total <= 0 {
		for i := 0; i < len(arg.InsertRecord); i++ {
			gift = append(gift, 0)
		}
		return
	}
	if arg.CanLottery != 0 {
		nowNum = arg.CanLottery // 可中奖次数
	}
	// 已中奖的奖品数量
	giftNum := make(map[int64]int64, len(arg.GiftMap))
	if giftNum, err = s.lottDao.CacheGiftNum(c, arg.Sid, arg.GiftMap); err != nil {
		log.Error("Immediate s.dao.CacheGiftNum arg(%v) error(%v)", arg, err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	mc := rand.Intn(5)
	for _, v := range arg.InsertRecord {
		var tmp int64
		if tmp, err = s.lottDao.CacheLotteryMcNum(c, arg.Sid, arg.High, mc); err != nil {
			log.Error("Immediate s.dao.CacheLotteryMcNum sid(%d) high(%d) mc(%d) error(%v)", arg.Sid, arg.High, mc, err)
			return
		}
		tmp = tmp + int64(v.Num)
		// 修改概率且当前概率已超过设置值
		if tmp > arg.Rate {
			tmp = arg.Rate / 2
			if err = s.lottDao.AddCacheLotteryMcNum(c, arg.Sid, arg.High, mc, tmp); err != nil {
				log.Error("Immediate s.dao.AddCacheLotteryMcNum sid(%d) high(%d) mc(%d) val(%d) error(%v)", arg.Sid, arg.High, mc, tmp, err)
				return
			}
		}
		if tmp != arg.Rate { // 未中奖
			if err = s.lottDao.AddCacheLotteryMcNum(c, arg.Sid, arg.High, mc, tmp); err != nil { // mcnum++
				log.Error("Immediate s.dao.AddCacheLotteryMcNum sid(%d) high(%d) val(%d) %d error(%v)", arg.Sid, arg.High, mc, tmp, err)
				return
			}
			if leastGift == nil { //未配置保底奖品
				gift = append(gift, 0)
				continue
			}
			// 配置保底奖品
			if arg.CanLottery == 0 { //未配置中奖上限
				if leastGift.Num-giftNum[leastGift.ID] <= 0 { // 无奖品
					gift = append(gift, 0)
					continue
				}
				gift = append(gift, leastGift.ID)
				if err = s.addTimesAndGiftNum(c, arg.Sid, arg.Mid, leastGift.ID, arg.LotteryMap[rsKey(_win, arg.Win)], 1, _usedTimes); err != nil {
					log.Error("s.addTimesAndGiftNum error(%v)", err)
					return
				}
				nowNum--
				continue
			}
			if nowNum == 0 { //中奖次数达到上限
				gift = append(gift, 0)
				continue
			}
			// 未达中奖上限
			if leastGift.Num-giftNum[leastGift.ID] <= 0 { // 无奖品
				// 无奖品，降级为未中奖
				gift = append(gift, 0)
				continue
			}
			gift = append(gift, leastGift.ID)
			if err = s.addTimesAndGiftNum(c, arg.Sid, arg.Mid, leastGift.ID, arg.LotteryMap[rsKey(_win, arg.Win)], 1, _usedTimes); err != nil {
				log.Error("s.addTimesAndGiftNum error(%v)", err)
				return
			}
			nowNum--
			continue
		}
		//中奖一
		if nowNum == 0 && arg.CanLottery != 0 { //配置了中奖上限且达到
			gift = append(gift, 0)
			continue
		}
		rand.Seed(time.Now().UnixNano())
		i := rand.Intn(2)
		if arg.Rate != 1 && (v.Type == _base || v.Type == _share || v.Type == _follow) { //外部行为
			if i == 0 {
				gift = append(gift, 0) //外部行为未中奖
				continue
			}
			//随机一个奖品
			l := rand.Int63n(total)
			// 计算l位于某个区间内
			giftIndex := countGiftSlice(l, randList)
			if arg.GiftMap[giftList[giftIndex]].Type == 1 { //实物类型
				// 海外ip检验
				var ok bool
				if ok, err = s.JudgeIpLocation(c, arg.Ip); err != nil {
					log.Error("s.JudgeIpLocation(%s) error(%v)", arg.Ip, err)
					return
				}
				if !ok {
					// 降级为未中奖
					gift = append(gift, 0)
					continue
				}
				if arg.GiftMap[giftList[giftIndex]].Num-giftNum[giftList[giftIndex]] <= 0 { // 无剩余奖品
					gift = append(gift, 0)
					continue
				}
				// 有剩余奖品
				gift = append(gift, giftList[giftIndex])
				if err = s.addTimesAndGiftNum(c, arg.Sid, arg.Mid, giftList[giftIndex], arg.LotteryMap[rsKey(_win, arg.Win)], 1, _usedTimes); err != nil {
					log.Error("s.addTimesAndGiftNum error(%v)", err)
					return
				}
				// 重置mcnum
				if err = s.lottDao.AddCacheLotteryMcNum(c, arg.Sid, arg.High, mc, 0); err != nil {
					log.Error("Immediate s.dao.AddCacheLotteryMcNum sid(%d) high(%d) mc(%d) val(%d) error(%v)", arg.Sid, arg.High, mc, 0, err)
					return
				}
				nowNum--
				continue
			}
			// 非实物类型，直接判断奖品数量
			if arg.GiftMap[giftList[giftIndex]].Num-giftNum[giftList[giftIndex]] <= 0 { // 无剩余奖品
				gift = append(gift, 0)
				continue
			}
			gift = append(gift, giftList[giftIndex])
			if err = s.addTimesAndGiftNum(c, arg.Sid, arg.Mid, giftList[giftIndex], arg.LotteryMap[rsKey(_win, arg.Win)], 1, _usedTimes); err != nil {
				log.Error("s.addTimesAndGiftNum error(%v)", err)
				return
			}
			if err = s.lottDao.AddCacheLotteryMcNum(c, arg.Sid, arg.High, mc, 0); err != nil {
				log.Error("Immediate s.dao.AddCacheLotteryMcNum sid(%d) high(%d) mc(%d) val(%d) error(%v)", arg.Sid, arg.High, mc, 0, err)
				return
			}
			nowNum--
			continue
		}
		// 内部行为，无需再随机
		l := rand.Int63n(total)

		// 计算l位于某个区间内
		giftIndex := countGiftSlice(l, randList)
		if arg.GiftMap[giftList[giftIndex]].Type == 1 { //实物类型
			// 海外ip检验
			var ok bool
			if ok, err = s.JudgeIpLocation(c, arg.Ip); err != nil {
				log.Error("s.JudgeIpLocation(%s) error(%v)", arg.Ip, err)
			}
			if !ok || err != nil {
				err = nil
				// 降级为未中奖
				gift = append(gift, 0)
				continue
			}
			if arg.GiftMap[giftList[giftIndex]].Num-giftNum[giftList[giftIndex]] <= 0 { // 无剩余奖品
				// 无剩余奖品
				gift = append(gift, 0)
				continue
			}
			gift = append(gift, giftList[giftIndex])
			if err = s.addTimesAndGiftNum(c, arg.Sid, arg.Mid, giftList[giftIndex], arg.LotteryMap[rsKey(_win, arg.Win)], 1, _usedTimes); err != nil {
				log.Error("s.addTimesAndGiftNum error(%v)", err)
				return
			}
			if err = s.lottDao.AddCacheLotteryMcNum(c, arg.Sid, arg.High, mc, 0); err != nil {
				log.Error("Immediate s.dao.AddCacheLotteryMcNum sid(%d) high(%d) mc(%d) val(%d) error(%v)", arg.Sid, arg.High, mc, 0, err)
				return
			}
			nowNum--
			continue
		}
		// 非实物奖品
		if arg.GiftMap[giftList[giftIndex]].Num-giftNum[giftList[giftIndex]] <= 0 { // 无剩余奖品
			gift = append(gift, 0)
			continue
		}
		gift = append(gift, giftList[giftIndex])
		if err = s.addTimesAndGiftNum(c, arg.Sid, arg.Mid, giftList[giftIndex], arg.LotteryMap[rsKey(_win, arg.Win)], 1, _usedTimes); err != nil {
			log.Error("s.addTimesAndGiftNum error(%v)", err)
			return
		}
		if err = s.lottDao.AddCacheLotteryMcNum(c, arg.Sid, arg.High, mc, 0); err != nil {
			log.Error("Immediate s.dao.AddCacheLotteryMcNum sid(%d) high(%d) mc(%d) val(%d) error(%v)", arg.Sid, arg.High, mc, 0, err)
			return
		}
		nowNum--
		continue
	}
	return
}

// 判断ip是否海外 ...
func (s *Service) JudgeIpLocation(c context.Context, ip string) (ok bool, err error) {
	ipInfo, err := s.locationClient.InfoComplete(c, &locationAPI.InfoCompleteReq{Addr: ip})
	if err != nil {
		log.Error("JudgeIpLocation s.locationRPC.InfoComplete ip(%v) error(%v)", ip, err)
		return
	}
	if ipInfo != nil && ipInfo.Info != nil {
		if ipInfo.Info.Country == "中国" {
			ok = true
		}
	}
	return
}

// 消耗抽奖次数
func (s *Service) consumeLotteryTimes(c context.Context, arg *l.ConsumeArg) (res []*l.InsertRecord, err error) {
	left := arg.Num
	sort := []int{_base, _share, _follow, _archive, _buy, _other, _customize, _ogv, _fe, _timeslike, _timescoin}
	for _, val := range sort {
		str := ""
		list := make([]int64, 0)
		var cid int64
		switch val {
		case _base:
			cid = arg.Base
			str = rsKey(val, cid)
		case _share:
			cid = arg.Share
			str = rsKey(val, cid)
		case _follow:
			cid = arg.Follow
			str = rsKey(val, cid)
		case _archive:
			list = arg.LikeList
		case _buy:
			list = arg.BuyList
		case _other:
			cid = arg.Other
			str = rsKey(_other, cid)
		case _customize:
			list = arg.CustomizeList
		case _ogv:
			list = arg.OgvList
		case _fe:
			cid = arg.Fe
			str = rsKey(val, cid)
		case _timeslike:
			cid = arg.TimesLike
			str = rsKey(val, cid)
		case _timescoin:
			cid = arg.TimesCoin
			str = rsKey(val, cid)
		}
		if val != _archive && val != _buy && val != _customize && val != _ogv {
			if arg.AddMap[str] == 0 || arg.UsedMap[str] >= arg.AddMap[str] {
				continue
			}
			insertRecord := make([]*l.InsertRecord, 0)
			insertRecord, left, err = s.countLotteryTimes(c, str, arg.Sid, arg.Mid, cid, left, val, arg.AddMap, arg.UsedMap, arg.LotteryMap)
			if err != nil {
				return
			}
			res = append(res, insertRecord...)
			if left == 0 {
				return
			}
			continue
		}
		for _, v := range list {
			cid = v
			str = rsKey(val, v)
			if arg.AddMap[str] == 0 && arg.UsedMap[str] >= arg.AddMap[str] {
				continue
			}
			insertRecord := make([]*l.InsertRecord, 0)
			insertRecord, left, err = s.countLotteryTimes(c, str, arg.Sid, arg.Mid, cid, left, val, arg.AddMap, arg.UsedMap, arg.LotteryMap)
			if err != nil {
				return
			}
			res = append(res, insertRecord...)
			if left == 0 {
				return
			}
		}
	}
	return
}

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
	if isOut && (actionType > 4 && actionType < 11 && actionType != _fe) {
		return
	}
	// 验证点赞投币情况
	if actionType == _timeslike || actionType == _timescoin {
		err = s.judgeLotteryCounter(c, actionType, mid, lotteryTimesConf)
		if err != nil {
			return
		}
	}
	err = s.JudgeAddTimes(c, actionType, num, lottery.ID, mid, cid, lotteryTimesConf, orderNo, ip)
	return
}

// GetIsCanAddTimes 是否可以增加抽奖次数
func (s *Service) GetIsCanAddTimes(c context.Context, sid string, id, mid int64, actionType, num int) (*l.CountStateReply, error) {
	if actionType != _timeslike && actionType != _timescoin {
		return &l.CountStateReply{State: l.TimesAddTimesStateNone}, nil
	}
	lottery, timesConfig, err := s.fetchLotteryConfig(c, sid)
	if err != nil {
		return nil, err
	}
	// 检查抽奖信息
	if err = checkTimesConf(lottery, timesConfig); err != nil {
		return &l.CountStateReply{State: l.TimesAddTimesStateNone}, nil
	}
	state, err := s.counterCanAddTimes(c, actionType, num, id, lottery.ID, mid, timesConfig)
	return &l.CountStateReply{State: state}, err
}

// counterCanAddTimes 能否增加抽奖次数
func (s *Service) counterCanAddTimes(c context.Context, actionType, num int, id, sid, mid int64, timesConfig []*l.LotteryTimesConfig) (state int, err error) {
	err = s.judgeLotteryCounter(c, actionType, mid, timesConfig)
	if err != nil {
		if xecode.EqualError(ecode.ActivityLotteryTimesNotEnough, err) {
			return l.TimesAddTimesStateNone, nil
		}
		return l.TimesAddTimesStateNone, err
	}
	incr, _, _, err := s.canAddTimes(c, actionType, num, sid, mid, id, timesConfig)
	if err != nil {
		if xecode.EqualError(ecode.ActivityLotteryAddTimesLimit, err) {
			return l.TimesAddTimesStateAlready, nil
		}
		return l.TimesAddTimesStateNone, err
	}
	if incr > 0 {
		return l.TimesAddTimesStateWait, nil
	}
	return l.TimesAddTimesStateAlready, nil
}

func (s *Service) canAddTimes(c context.Context, actionType int, num int, sid, mid, id int64, timesConfig []*l.LotteryTimesConfig) (incr int, cid int64, timesKey string, err error) {
	var (
		lotteryMap = make(map[string]*l.LotteryTimesConfig, len(timesConfig))
		times      = make(map[string]int)
	)
	for _, l := range timesConfig {
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
	if times, err = s.lottDao.LotteryAddTimes(c, timesConfig, sid, mid); err != nil {
		log.Error("JudgeAddTimes: s.dao.LotteryAddTimes sid(%d) mid(%d) error(%v)", sid, mid, err)
		return
	}
	// 增加次数达到上限
	if times[key] >= lotteryMap[key].Most {
		err = ecode.ActivityLotteryAddTimesLimit
		return
	}
	incr = lotteryMap[key].Times
	if num > 0 {
		incr = num
	}
	// 增加次数达上限
	if times[key]+incr > lotteryMap[key].Most {
		incr = lotteryMap[key].Most - times[key]
	}
	return
}

func (s *Service) judgeLotteryCounter(c context.Context, actionType int, mid int64, timesConfig []*l.LotteryTimesConfig) (err error) {
	for _, v := range timesConfig {
		if v.Type == actionType {
			return s.judgeCounterTimes(c, mid, v)
		}
	}
	return ecode.ActivityLotteryTimesTypeError
}

func (s *Service) judgeCounterTimes(c context.Context, mid int64, times *l.LotteryTimesConfig) (err error) {
	info := &l.TimesInfo{}
	err = json.Unmarshal([]byte(times.Info), info)
	if err != nil {
		return
	}
	count, err := s.getCounter(c, info.Counter, info.Activity, mid)
	if err != nil {
		return
	}
	if count >= info.Count {
		return
	}
	return ecode.ActivityLotteryTimesNotEnough
}

func (s *Service) getCounter(c context.Context, counter, activity string, mid int64) (int64, error) {

	resp, err := client.ActPlatClient.GetCounterRes(c, &actplat.GetCounterResReq{
		Counter:  counter,
		Activity: activity,
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
	count := resp.CounterList[0]
	return count.Val, nil

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

// 实物中奖添加地址接口
func (s *Service) AddLotteryAddress(c context.Context, sid string, id, mid int64) (err error) {
	var (
		addrId  int64
		val     *l.AddressInfo
		lottery *l.Lottery
	)
	if lottery, err = s.lottDao.Lottery(c, sid); err != nil {
		err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		return
	}
	if addrId, err = s.lottDao.LotteryAddr(c, lottery.ID, mid); err != nil {
		log.Error("AddLotteryAddress s.dao.LotteryAddr mid(%d) error(%v)", mid, err)
		return
	}
	if addrId == id {
		err = ecode.ActivityAddrHasAdd
		return
	}
	if addrId == 0 {
		// 校验传输的地址id是否有效
		if val, err = s.lottDao.GetMemberAddress(c, id, mid); err != nil {
			log.Error("AddLotteryAddress s.dao.GetMemberAddress id(%d) mid(%d) error(%v)", id, mid, err)
			return
		}
		if val == nil || val.ID == 0 {
			err = ecode.ActivityAddrAddFail
			return
		}
		if _, err = s.lottDao.InsertLotteryAddr(c, lottery.ID, mid, id); err != nil {
			log.Error("AddLotteryAddress s.dao.InsertLotteryAddr sid(%d) mid(%d) id(%d) error(%v)", lottery.ID, mid, id, err)
			return
		}
		s.cache.Do(c, func(c context.Context) {
			s.lottDao.AddCacheLotteryAddrCheck(c, lottery.ID, mid, id)
		})
	}
	return
}

func (s *Service) LotteryAddress(c context.Context, sid string, mid int64) (res *l.AddressInfo, err error) {
	var (
		addrId  int64
		lottery *l.Lottery
	)
	if lottery, err = s.lottDao.Lottery(c, sid); err != nil {
		err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		return
	}
	if addrId, err = s.lottDao.LotteryAddr(c, lottery.ID, mid); err != nil {
		log.Error("LotteryAddress s.dao.LotteryAddr mid(%d) error(%v)", mid, err)
		return
	}
	if addrId == 0 {
		err = ecode.ActivityAddrNotAdd
		return
	}
	if res, err = s.lottDao.GetMemberAddress(c, addrId, mid); err != nil {
		log.Error("LotteryAddress s.dao.GetMemberAddress(%d,%d) error(%v)", addrId, mid, err)
	}
	return
}

// 获取我的抽奖记录接口
func (s *Service) GetMyList(c context.Context, sid string, pn, ps int, mid int64, needAddress bool) (res *l.LotteryRecordRes, err error) {
	var (
		lottery           *l.Lottery
		lotteryGift       []*l.LotteryGift
		lotteryRecordList []*l.LotteryRecordDetail
		start             = int64((pn - 1) * ps)
		end               = start + int64(ps) - 1
		ok                bool
		count             int
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if lottery, err = s.lottDao.Lottery(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if lotteryGift, err = s.lottDao.LotteryGift(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryGift %s", sid)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	if lottery.ID == 0 {
		err = ecode.ActivityNotExist
		return
	}
	if lotteryRecordList, err = s.lotteryRecord(c, lottery.ID, mid, start, end); err != nil {
		log.Error("Failed to fetch lottery record:%d %d %+v", lottery.ID, mid, err)
		return
	}
	giftMap := make(map[int64]*l.LotteryGift)
	for _, v := range lotteryGift {
		giftMap[v.ID] = v
	}
	list := make([]*l.LotteryRecordDetail, 0)
	for _, value := range lotteryRecordList {
		// 未中奖
		if value.GiftID == 0 {
			tmp := &l.LotteryRecordDetail{ID: value.ID, Mid: value.Mid, Num: value.Num, GiftID: 0, GiftName: "未中奖", ImgURL: "", Type: value.Type, GiftType: 0, Ctime: value.Ctime}
			list = append(list, tmp)
			continue
		}
		tmp := &l.LotteryRecordDetail{ID: value.ID, Mid: value.Mid, Num: value.Num, GiftID: value.GiftID, GiftName: giftMap[value.GiftID].Name, GiftType: giftMap[value.GiftID].Type, ImgURL: giftMap[value.GiftID].ImgUrl, Type: value.Type, Ctime: value.Ctime}
		list = append(list, tmp)
	}
	// 判断是否填写过地址
	if needAddress {
		var addrId int64
		if addrId, err = s.lottDao.LotteryAddr(c, lottery.ID, mid); err != nil {
			log.Error("GetMyList s.dao.LotteryAddr id(%d) mid(%d) error(%v)", lottery.ID, mid, err)
			return
		}
		if addrId != 0 {
			ok = true
		}
	}
	count = len(lotteryRecordList)
	page := &l.Page{Num: pn, Size: ps, Total: count}
	res = &l.LotteryRecordRes{List: list, Page: page, IsAddAddress: ok}
	return
}

// 获取未抽奖次数接口
func (s *Service) GetUnusedTimes(c context.Context, sid string, mid int64) (res *l.LotteryTimesRes, err error) {
	var (
		now       = time.Now()
		addTimes  = make(map[string]int, 8)
		usedTimes = make(map[string]int, 8)
		all, used int
		win       int64
		ip        = metadata.String(c, metadata.RemoteIP)
	)
	lottery, lotteryTimesConf, err := s.fetchLotteryConfig(c, sid)
	if err != nil {
		return nil, err
	}
	if err = checkTimesConf(lottery, lotteryTimesConf); err != nil {
		return
	}
	orderNo := strconv.FormatInt(mid, 10) + strconv.FormatInt(now.Unix(), 10)
	s.JudgeAddTimes(c, _base, 0, lottery.ID, mid, 0, lotteryTimesConf, orderNo, ip)
	if sid == s.c.VipSpecial.Sid {
		check, err := s.vipActClient.LotteryTimesCheck(c, &vipact.LotteryTimesCheckReq{Mid: mid, LotteryId: sid})
		if err == nil && check != nil && check.LotteryTimes > 0 {
			orderNo := strconv.FormatInt(now.Unix(), 10) + strconv.FormatInt(mid, 10)
			actionType := _other
			if check.LotteryTimes == 3 {
				actionType = _customize
			}
			s.JudgeAddTimes(c, actionType, 0, lottery.ID, mid, 0, lotteryTimesConf, orderNo, ip)
		}
	}
	for _, val := range lotteryTimesConf {
		if val.Type == _win {
			win = val.ID
			break
		}
	}
	if addTimes, err = s.lottDao.LotteryAddTimes(c, lotteryTimesConf, lottery.ID, mid); err != nil {
		log.Error("GetUnusedTimes: s.dao.LotteryAddTimes id(%d) mid(%d) error(%v)", lottery.ID, mid, err)
		return
	}
	if usedTimes, err = s.lottDao.LotteryUsedTimes(c, lotteryTimesConf, lottery.ID, mid); err != nil {
		log.Error("GetUnusedTimes: s.dao.LotteryUsedTimes id(%d) mid(%d) error(%v)", lottery.ID, mid, err)
		return
	}
	for k, v := range addTimes {
		if k == rsKey(_win, win) {
			continue
		}
		all = all + v
	}
	for k, v := range usedTimes {
		if k == rsKey(_win, win) {
			continue
		}
		used = used + v
	}
	t := all - used
	if t < 0 {
		t = 0
	}
	res = &l.LotteryTimesRes{Times: t}
	return
}

// judgeUser judge user could lottery or not .
func (s *Service) judgeUserLottery(c context.Context, lotteryInfo *l.LotteryInfo, member *accapi.Profile, ip string) (err error) {
	if member.Silence == _silenceForbid {
		err = ecode.ActivityMemberBlocked
		return
	}
	// 账号等级限制
	if lotteryInfo.Level != 0 {
		if member.Level < int32(lotteryInfo.Level) {
			err = ecode.ActivityLotteryLevelLimit
			return
		}
	}
	// 注册时间限制
	if lotteryInfo.RegTimeStime != 0 {
		if int64(member.JoinTime) > lotteryInfo.RegTimeStime {
			err = ecode.ActivityLotteryRegisterEarlyLimit
			return
		}
	}
	if lotteryInfo.RegTimeEtime != 0 {
		if int64(member.JoinTime) < lotteryInfo.RegTimeEtime {
			err = ecode.ActivityLotteryRegisterLastLimit
			return
		}
	}
	// vip限制
	switch lotteryInfo.VipCheck {
	case _vipCheck: // vip专享
		if member.Vip.Type == 0 || member.Vip.Status != 1 {
			err = ecode.ActivityNotVip
			return
		}
	case _monthVip: // 月度大会员
		if member.Vip.Type == 0 || member.Vip.Status != 1 {
			err = ecode.ActivityNotMonthVip
			return
		}
	case _yearVip: // 年度大会员
		if member.Vip.Type != 2 || member.Vip.Status != 1 {
			err = ecode.ActivityNotYearVip
			return
		}
	}
	// 账号验证
	switch lotteryInfo.AccountCheck {
	case _telValid: // 手机验证
		if member.TelStatus != 1 {
			err = ecode.ActivityTelValid
			return
		}
	case _identifyValid: // 实名验证
		if member.Identification != 1 {
			err = ecode.ActivityIdentificationValid
			return
		}
	}
	// IP防刷
	if lotteryInfo.FsIP == 1 {
		var used int
		if used, err = s.dao.CacheIPRequestCheck(c, ip); err != nil {
			log.Error("s.dao.GetIPRequestCheck(%s) error(%v)", ip, err)
			return
		}
		if used != 0 {
			err = ecode.ActivityLotteryIPFrequence
			return
		}
		if err = s.dao.AddCacheIPRequestCheck(c, ip, 1); err != nil {
			log.Error("s.dao.SetIPRequestCheck(%s, %d) error(%v)", ip, 1, err)
			return
		}
	}
	return
}

// 中奖名单接口
func (s *Service) WinList(c context.Context, sid string, num int64, needCache bool) (res []*l.WinList, err error) {
	var (
		lottery     *l.Lottery
		lotteryGift []*l.LotteryGift
		giftList    []*l.GiftList
		membersRly  *accapi.InfosReply
		mids        []int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if lottery, err = s.lottDao.Lottery(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if lotteryGift, err = s.lottDao.LotteryGift(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryGift %s", sid)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	if lottery.ID == 0 {
		err = ecode.ActivityNotExist
		return
	}
	if len(lotteryGift) == 0 {
		err = ecode.ActivityNotConfig
		return
	}
	showGift := make([]int64, 0)
	showGiftMap := make(map[int64]*l.LotteryGift)
	for _, v := range lotteryGift {
		if v.IsShow == l.IsShow {
			showGiftMap[v.ID] = v
			showGift = append(showGift, v.ID)
		}

	}
	if giftList, err = s.lottDao.LotteryWinList(c, lottery.ID, showGift, num, needCache); err != nil {
		log.Error("WinList s.dao.LotteryWinList sid(%d) num(%d) error(%v)", lottery.ID, num, err)
		return
	}
	mids = make([]int64, 0, len(giftList))
	for _, v := range giftList {
		if v.GiftID == -1 {
			continue
		}
		mids = append(mids, v.Mid)
	}
	if len(mids) == 0 {
		return
	}
	if membersRly, err = s.accClient.Infos3(c, &accapi.MidsReq{Mids: mids}); err != nil {
		log.Error("s.accRPC.Infos3(%v) error(%v)", mids, err)
		return
	}
	res = make([]*l.WinList, 0, len(giftList))
	for _, v := range giftList {
		if v.GiftID == -1 {
			continue
		}
		gift, ok := showGiftMap[v.GiftID]
		if !ok {
			continue
		}
		v.GiftName = gift.Name
		n := &l.WinList{GiftList: v}
		if membersRly != nil {
			if val, y := membersRly.Infos[v.Mid]; y {
				n.Name = hideName(val.Name)
				n.Mid = 0
			}
		}
		res = append(res, n)
	}
	return
}

func hideName(name string) (res string) {
	if name == "" {
		return "***"
	}
	tmp := []rune(name)
	l := len(tmp)
	if l <= 3 {
		return "***" + name
	}
	for i := 0; i < 3; i++ {
		tmp[i] = '*'
	}
	res = string(tmp)
	return
}

func (s *Service) addTimesAndGiftNum(c context.Context, sid, mid, giftID int64, ltc *l.LotteryTimesConfig, val int, status string) (err error) {
	if err = s.lottDao.IncrGiftNum(c, sid, giftID); err != nil {
		log.Error("addTimesAndGiftNum s.dao.IncrGiftNum sid(%d) id(%d) error(%v)", sid, giftID, err)
		return
	}
	var ef int64
	if ef, err = s.lottDao.UpdatelotteryGiftNumSQL(c, giftID); err != nil {
		log.Error("addTimesAndGiftNum s.dao.UpdatelotteryGiftNumSQL id(%d) error(%v)", giftID, err)
		return
	}
	if ef == 0 {
		err = ecode.ActivityLotteryFail
		return
	}
	if _, err = s.lottDao.IncrTimes(c, sid, mid, ltc, val, _usedTimes); err != nil {
		log.Error("addTimesAndGiftNum s.dao.IncrTimes sid(%d) mid(%d) arg(%v) val(%d) status(%s) error(%v)", sid, mid, ltc, val, _usedTimes, err)
		return
	}
	return
}

func rsKey(actionType int, cid int64) string {
	return strconv.Itoa(actionType) + strconv.FormatInt(cid, 10)
}

func (s *Service) countLotteryTimes(c context.Context, str string, sid, mid, cid int64, left, actionType int, addMap, usedMap map[string]int, lotteryMap map[string]*l.LotteryTimesConfig) (res []*l.InsertRecord, leftTimes int, err error) {
	if usedMap[str]+left <= addMap[str] {
		if _, err = s.lottDao.IncrTimes(c, sid, mid, lotteryMap[str], left, _usedTimes); err != nil {
			log.Error("ConsumeLotteryTimes: s.dao.IncrTimes sid(%d) mid(%d) type(%d) val(%d) status(%s) error(%v)", sid, mid, actionType, left, _usedTimes, err)
			return
		}

		for i := 0; i < left; i++ {
			index := i
			iR := &l.InsertRecord{Mid: mid, Num: 1, Type: actionType, CID: cid, OrderNo: s.recordBuildOrderNo(c, index, mid, cid)}
			res = append(res, iR)
		}
		return
	}
	leftTimes = usedMap[str] + left - addMap[str]
	if _, err = s.lottDao.IncrTimes(c, sid, mid, lotteryMap[str], addMap[str]-usedMap[str], _usedTimes); err != nil {
		log.Error("ConsumeLotteryTimes: s.dao.IncrTimes sid(%d) mid(%d) type(%d) val(%d) status(%s) error(%v)", sid, mid, _base, addMap[str]-usedMap[str], _usedTimes, err)
		return
	}
	for i := 0; i < addMap[str]-usedMap[str]; i++ {
		index := i
		iR := &l.InsertRecord{Mid: mid, Num: 1, Type: actionType, CID: cid, OrderNo: s.recordBuildOrderNo(c, index, mid, cid)}
		res = append(res, iR)
	}
	return
}

func (s *Service) recordBuildOrderNo(c context.Context, index int, mid, cid int64) string {
	return fmt.Sprintf("%d_%d_%d_%d", index, mid, cid, time.Now().Unix())
}

func checkLottery(lottery *l.Lottery, info *l.LotteryInfo, config []*l.LotteryTimesConfig, gift []*l.LotteryGift, now int64) (err error) {

	if lottery.ID == 0 {
		err = ecode.ActivityNotExist
		return
	}
	if lottery.Stime.Time().Unix() > now {
		err = ecode.ActivityNotStart
		return
	}
	if lottery.Etime.Time().Unix() < now {
		err = ecode.ActivityOverEnd
		return
	}
	if info.ID == 0 || len(config) == 0 || len(gift) == 0 {
		err = ecode.ActivityNotConfig
	}
	return
}

func (s *Service) lotteryWithoutCount(c context.Context, lottery *l.Lottery, arg *l.ConsumeArg, coin int, now int64, ip string) (res []*l.LotteryRecordDetail, err error) {
	// 扣除硬币
	if coin > 0 {
		if _, err = s.coinClient.ModifyCoins(c, &coinmdl.ModifyCoinsReq{Mid: arg.Mid, Count: float64(-coin * arg.Num), Reason: _lottery, IP: ip}); err != nil {
			err = ecode.ActivityNotEnoughCoin
			return
		}
	}
	res, err = s.lotteryOperation(c, lottery, now, arg, ip)
	return
}

// sendSysMsg
func (s *Service) sendSysMsg(c context.Context, mid, sender int64, activityName, giftName, lotteryID, activityURL, addressText string, addressURL string) (err error) {
	params := make([]string, 0)
	params = append(params, activityName, giftName)
	letterParasm := &lott.LetterParam{
		RecverIDs:  []uint64{uint64(mid)},
		SenderUID:  uint64(sender),
		MsgType:    lott.MsgTypeCard,       //通知卡类型 type = 10
		NotifyCode: s.c.Lottery.NotifyCode, //通知码
		Params:     strings.Join(params, paramsSplit),
		Title:      s.c.Lottery.MessageTitle,
	}
	if activityURL != "" {
		letterParasm.JumpText = s.c.Lottery.ActivityText
		letterParasm.JumpURL = activityURL
	}
	if addressURL != "" {
		letterParasm.JumpText2 = addressText
		letterParasm.JumpURL2 = addressURL
	}
	_, err = s.lottDao.SendLetter(c, letterParasm)
	if err != nil {
		log.Errorc(c, "s.lottDao.SendLetter error(%v)", err)
	}
	return
}

func (s *Service) CardNum(c context.Context, sid string, mid int64) (res *l.LotteryCard, err error) {
	res = &l.LotteryCard{Num: 0}
	if mid <= 0 {
		return
	}
	var record *l.LotteryRecordRes
	if record, err = s.GetMyList(c, sid, 1, s.c.SpringCardAct.NumLimit, mid, true); err != nil {
		log.Error("CardNum(%s,%d) error(%v)", sid, mid, err)
		return
	}
	if record == nil || len(record.List) == 0 {
		return
	}
	card := make([]int, 6)
	for _, v := range record.List {
		switch v.GiftID {
		case s.c.SpringCardAct.CardA:
			card[0]++
			continue
		case s.c.SpringCardAct.CardB:
			card[1]++
			continue
		case s.c.SpringCardAct.CardC:
			card[2]++
			continue
		case s.c.SpringCardAct.CardD:
			card[3]++
			continue
		case s.c.SpringCardAct.CardE:
			card[4]++
			continue
		case s.c.SpringCardAct.CardF:
			card[5]++
			continue
		default:
			continue
		}
	}
	tmp := 100
	for _, v := range card {
		if tmp > v {
			tmp = v
		}
	}
	res.Num = tmp
	cardMap := map[int64]int{
		s.c.SpringCardAct.CardA: card[0],
		s.c.SpringCardAct.CardB: card[1],
		s.c.SpringCardAct.CardC: card[2],
		s.c.SpringCardAct.CardD: card[3],
		s.c.SpringCardAct.CardE: card[4],
		s.c.SpringCardAct.CardF: card[5],
	}
	res.Card = cardMap
	return
}

func countGiftSlice(randNum int64, randList []int64) (res int) {
	for i := 0; i < len(randList); i++ {
		if randNum+1 <= randList[i] {
			res = i
			break
		}
	}
	return
}

// buildAddressLink 建立抽奖地址
func (s *Service) buildAddressLink(ctx context.Context, lotteryID string, giftType int) string {
	if giftType == _entity {
		return fmt.Sprintf("%s%s", s.c.Lottery.AddressLink, lotteryID)
	}
	return ""
}

// buildAddressLink 建立抽奖地址
func (s *Service) buildCouponLink(ctx context.Context, lotteryID string, giftType int, giftID int64) string {
	if giftType == _coupon {
		return fmt.Sprintf("%s%s&gift_id=%d", s.c.Lottery.CouponLink, lotteryID, giftID)
	}
	return ""
}

func (s *Service) sendLotteryAward(ctx context.Context, giftMap map[int64]*l.LotteryGift, gid []int64, id, mid int64, ip string, lottery *l.Lottery, lotteryInfo *l.LotteryInfo) (err error) {

	for _, v := range gid {
		// 未中奖
		if v == 0 {
			continue
		}
		gm, ok := giftMap[v]
		if !ok || gm == nil {
			log.Warn("giftMap not found obj(%v)", giftMap)
			continue
		}
		switch gm.Type {
		case _entity, _otherGift: //实物类型或其他奖品类型
			if _, err = s.lottDao.InsertLotteryWin(ctx, id, v, mid, ip); err != nil {
				log.Error("DoLottery s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", id, v, mid, err)
				return
			}
			if lotteryInfo != nil && lotteryInfo.SenderID > 0 {
				s.sendSysMsg(ctx, mid, lotteryInfo.SenderID, lottery.Name, gm.Name, lottery.LotteryID, lotteryInfo.ActivityLink, s.c.Lottery.AddressText, s.buildAddressLink(ctx, lottery.LotteryID, gm.Type))
			}
		case _grant:
			if _, err = s.lottDao.InsertLotteryWin(ctx, id, v, mid, ip); err != nil {
				log.Error("sendLotteryAward s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", id, v, mid, err)
				return
			}
			uids := []int64{mid}
			source := new(l.GrantJson)
			if err = json.Unmarshal([]byte(gm.Source), &source); err != nil {
				log.Error("sendLotteryAward json.Unmarshal arg(%s) error(%v)", gm.Source, err)
				return
			}
			if _, suitErr := s.suitClient.GrantByMids(ctx, &suitmdl.GrantByMidsReq{Mids: uids, Pid: source.Pid, Expire: source.Expire}); suitErr != nil {
				log.Error("s.suitClient.GrantByMids(%d,%d,%d) error(%v)", mid, source.Pid, source.Expire, suitErr)
				return
			}
			if lotteryInfo != nil && lotteryInfo.SenderID > 0 {
				s.sendSysMsg(ctx, mid, lotteryInfo.SenderID, lottery.Name, gm.Name, lottery.LotteryID, lotteryInfo.ActivityLink, "", s.buildAddressLink(ctx, lottery.LotteryID, gm.Type))
			}
		case _coin:
			var giftID int64
			if giftID, err = s.lottDao.InsertLotteryWin(ctx, id, v, mid, ip); err != nil {
				log.Error("sendLotteryAward s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", id, v, mid, err)
				return
			}
			f, _ := strconv.ParseFloat(gm.Source, 64)
			orderID := fmt.Sprintf("%d_%d_%d_%d", mid, lottery.ID, gm.ID, giftID)
			if _, e := s.coinClient.ModifyCoins(ctx, &coinmdl.ModifyCoinsReq{Mid: mid, Count: f, Reason: _remarkCoin, IP: ip, UniqueID: orderID, Caller: caller}); e != nil {
				log.Error("sendLotteryAward need check coin.ModifyCoin mid:%d count:%v error(%v)", mid, f, e)
				return
			}
			if lotteryInfo != nil && lotteryInfo.SenderID > 0 {
				s.sendSysMsg(ctx, mid, lotteryInfo.SenderID, lottery.Name, gm.Name, lottery.LotteryID, lotteryInfo.ActivityLink, "", s.buildAddressLink(ctx, lottery.LotteryID, gm.Type))
			}
		case _coupon:
			var cdKey string
			if _, err = s.lottDao.UpdateLotteryWin(ctx, id, mid, v, ip); err != nil {
				log.Error("sendLotteryAward s.dao.UpdateLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", id, v, mid, err)
				return
			}
			if cdKey, err = s.lottDao.RawLotteryWinOne(ctx, id, mid, v); err != nil {
				log.Error("sendLotteryAward s.dao.RawLotteryWinOne id(%d) mid(%d) gift_id(%d) error(%v)", id, mid, v, err)
				return
			}
			if lotteryInfo != nil && lotteryInfo.SenderID > 0 {
				s.sendSysMsg(ctx, mid, lotteryInfo.SenderID, lottery.Name, gm.Name+"，兑换码为："+cdKey, lottery.LotteryID, lotteryInfo.ActivityLink, s.c.Lottery.CouponText, s.buildCouponLink(ctx, lottery.LotteryID, gm.Type, v))
			}
		case _memberCoupon:
			if _, err = s.lottDao.InsertLotteryWin(ctx, id, v, mid, ip); err != nil {
				log.Error("sendLotteryAward s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", id, v, mid, err)
				continue
			}
			if _, e := s.lottDao.MemberCoupon(ctx, mid, gm.Source); e != nil {
				log.Error("sendLotteryAward s.dao.MemberCoupon mid:%d token(%s) error(%v)", mid, gm.Source, e)
				continue
			}
			if lotteryInfo != nil && lotteryInfo.SenderID > 0 {
				s.sendSysMsg(ctx, mid, lotteryInfo.SenderID, lottery.Name, gm.Name, lottery.LotteryID, lotteryInfo.ActivityLink, "", s.buildAddressLink(ctx, lottery.LotteryID, gm.Type))
			}
		case _member:
			if _, err = s.lottDao.InsertLotteryWin(ctx, id, v, mid, ip); err != nil {
				log.Error("sendLotteryAward s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", id, v, mid, err)
				continue
			}
			if e := s.lottDao.MemberVip(ctx, mid, gm.Source, _remark); e != nil {
				log.Error("sendLotteryAward s.dao.MemberVip mid:%d token(%s) error(%v)", mid, gm.Source, e)
				continue
			}
			if lotteryInfo != nil && lotteryInfo.SenderID > 0 {
				s.sendSysMsg(ctx, mid, lotteryInfo.SenderID, lottery.Name, gm.Name, lottery.LotteryID, lotteryInfo.ActivityLink, "", s.buildAddressLink(ctx, lottery.LotteryID, gm.Type))
			}
		default:
			return
		}
	}
	err = s.lottDao.DeleteLotteryWinLog(ctx, id, mid)
	if err != nil {
		log.Errorc(ctx, "s.lottDao.DeleteLotteryWinLog err(%v)", err)
	}
	return
}

func (s *Service) lotteryRecord(ctx context.Context, id, mid, start, end int64) ([]*l.LotteryRecordDetail, error) {
	lotteryRecordList, err := s.lottDao.CacheLotteryActionLog(ctx, id, mid, start, end)
	if err != nil {
		return nil, err
	}
	if len(lotteryRecordList) == 0 {
		rawList, err := s.lottDao.RawLotteryUsedTimes(ctx, id, mid)
		if err != nil {
			return nil, err
		}
		s.cache.Do(ctx, func(c context.Context) {
			s.lottDao.AddCacheLotteryActionLog(c, id, mid, rawList)
		})
		rawLen := int64(len(rawList))
		if rawLen < end {
			end = rawLen
		}
		if start >= rawLen {
			return nil, nil
		}
		lotteryRecordList = rawList[start:end]
	}
	return lotteryRecordList, nil
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

func (s *Service) AddExtraTimes(ctx context.Context, sid string, mid int64) error {
	stupidStatus, err := s.StupidStatus(ctx, sid, mid)
	if err != nil {
		return err
	}
	if !stupidStatus.IsAfrican {
		return ecode.ActivityTaskPreNotCheck
	}
	if err := s.dao.RsDelNX(ctx, fmt.Sprintf(lockKey, stupidStatus.Lottery.ID, mid)); err != nil {
		return ecode.ActivityTaskAwardFailed
	}
	orderNo := strconv.FormatInt(mid, 10) + strconv.Itoa(10) + strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	if err = s.JudgeAddTimes(ctx, 10, 0, stupidStatus.Lottery.ID, mid, 0, stupidStatus.LotteryTimesConf, orderNo, metadata.String(ctx, metadata.RemoteIP)); err != nil {
		return err
	}
	if _, suitErr := s.suitClient.GrantByMids(ctx, &suitmdl.GrantByMidsReq{Mids: []int64{mid}, Pid: s.c.Stupid.Pid, Expire: s.c.Stupid.PidExpire}); suitErr != nil {
		log.Error("Failed to send grant:%d %d %d %+v", mid, s.c.Stupid.Pid, s.c.Stupid.PidExpire, suitErr)
	}
	return nil
}

// 获取我的中奖记录接口
func (s *Service) GetMyWinList(c context.Context, sid string, mid int64, needAddress bool) (res *l.LotteryRecordRes, err error) {
	var (
		lottery           *l.Lottery
		lotteryGift       []*l.LotteryGift
		lotteryRecordList []*l.LotteryRecordDetail
		ok                bool
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if lottery, err = s.lottDao.Lottery(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if lotteryGift, err = s.lottDao.LotteryGift(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryGift %s", sid)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	if lottery.ID == 0 {
		err = ecode.ActivityNotExist
		return
	}
	eg2 := errgroup.WithContext(c)
	eg2.Go(func(ctx context.Context) (err error) {
		if lotteryRecordList, err = s.lottDao.LotteryMyWinList(ctx, lottery.ID, mid); err != nil {
			log.Error("Failed to fetch lottery record:%d %d %+v", lottery.ID, mid, err)
		}
		return
	})
	// 判断是否填写过地址
	if needAddress {
		eg2.Go(func(ctx context.Context) (err error) {
			var addrId int64
			if addrId, err = s.lottDao.LotteryAddr(ctx, lottery.ID, mid); err != nil {
				log.Error("Failed to get addr: %d, %d, %+v", lottery.ID, mid, err)
				err = nil
				return
			}
			if addrId != 0 {
				ok = true
			}
			return
		})
	}
	if err = eg2.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	giftMap := make(map[int64]*l.LotteryGift)
	for _, v := range lotteryGift {
		giftMap[v.ID] = v
	}
	list := make([]*l.LotteryRecordDetail, 0)
	for _, value := range lotteryRecordList {
		if value.GiftID <= 0 || value.ID <= 0 {
			continue
		}
		gift, ok := giftMap[value.GiftID]
		if !ok {
			continue
		}
		item := &l.LotteryRecordDetail{
			ID:       value.ID,
			Mid:      value.Mid,
			Num:      value.Num,
			GiftID:   value.GiftID,
			GiftName: gift.Name,
			GiftType: gift.Type,
			ImgURL:   gift.ImgUrl,
			Type:     value.Type,
			Ctime:    value.Ctime,
		}
		list = append(list, item)
	}
	res = &l.LotteryRecordRes{List: list, IsAddAddress: ok}
	return
}
