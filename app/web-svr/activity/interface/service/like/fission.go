package like

import (
	"context"
	"fmt"
	"strconv"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	l "go-gateway/app/web-svr/activity/interface/model/lottery"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

func (s *Service) FissionDoLottery(c context.Context, sid string, mid int64) (res []*l.LotteryRecordDetail, err error) {
	var (
		lottery                                                           *l.Lottery
		lotteryInfo                                                       *l.LotteryInfo
		lotteryTimesConf                                                  []*l.LotteryTimesConfig
		lotteryGift                                                       []*l.LotteryGift
		base, win, share, follow, other, fe, lastID, timesLike, timesCoin int64
		canLottery                                                        int
	)
	caller := metadata.String(c, metadata.Caller)
	if caller != s.c.Fission.Caller {
		err = xecode.AccessDenied
		log.Warn("FissionDoLottery Not Authorised Caller %s", caller)
		return
	}
	if _, ok := s.c.Fission.Sids[sid]; !ok {
		err = xecode.AccessDenied
		log.Warn("FissionDoLottery sid:%s not found", sid)
		return
	}
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
		log.Error("FissionDoLottery eg.Wait error(%v)", err)
		return
	}
	nowTs := time.Now().Unix()
	if err = checkLottery(lottery, lotteryInfo, lotteryTimesConf, lotteryGift, nowTs); err != nil {
		log.Error("FissionDoLottery checkLottery lotteryID(%d) error(%v)", lottery.ID, err)
		return
	}
	ip := metadata.String(c, metadata.RemoteIP)
	memberRly, err := s.accClient.Profile3(c, &accapi.MidReq{Mid: mid})
	if err != nil || memberRly == nil {
		log.Error("FissionDoLottery s.accRPC.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%v)", mid, err)
		return
	}
	// 账号验证
	if err = s.judgeUserLottery(c, lotteryInfo, memberRly.Profile, ip); err != nil {
		log.Error("FissionDoLottery s.judgeUserLottery lotteryInfo(%v) profile(%v) error(%v)", lotteryInfo, memberRly.Profile, err)
		return
	}
	orderNo := strconv.FormatInt(mid, 10) + strconv.FormatInt(nowTs, 10)
	// 新增每日抽奖机会
	if err = s.JudgeAddTimes(c, _other, 0, lottery.ID, mid, 0, lotteryTimesConf, orderNo, ip); err != nil {
		log.Error("FissionDoLottery s.JudgeAddTimes mid:%d sid:%s error(%v)", mid, sid, err)
		err = nil
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
		log.Error("FissionDoLottery egp.Wait error(%v)", err)
		return
	}
	// 无抽奖次数
	if all-used < 1 {
		err = ecode.ActivityNoTimes
		return
	}
	consumeArg := &l.ConsumeArg{
		Sid:           lottery.ID,
		Mid:           mid,
		AddMap:        addMap,
		UsedMap:       usedMap,
		LotteryMap:    lotteryMap,
		Num:           1,
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
			res, err = s.lotteryWithoutCount(c, lottery, consumeArg, lotteryInfo.Coin, nowTs, ip)
			return
		}
	}
	rate := lotteryInfo.GiftRate
	// 消耗抽奖次数
	insertRecord := make([]*l.InsertRecord, 0, 1)
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
		High:         0,
		CanLottery:   canLottery,
		Win:          win,
	}
	gid := make([]int64, 0, 1)
	// 抽奖
	if gid, err = s.Immediate(c, lotteryArg); err != nil {
		log.Error("FissionDoLottery: s.Immediate sid(%s) mid(%d) lottery(%v) lotteryGift(%v) error(%v)", sid, mid, lottery, lotteryGift, err)
		return
	}
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
	var actionLog []*l.LotteryRecordDetail
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
		lrd.Ctime = xtime.Time(nowTs + int64(i))
		res = append(res, lrd)
		*tmp = *lrd
		tmp.Type = insertRecord[i].Type
		tmp.ID = lastID
		actionLog = append(actionLog, tmp)
	}
	//新增抽奖记录redis
	if err = s.lottDao.AddLotteryActionLog(c, lottery.ID, mid, actionLog); err != nil {
		log.Error("FissionDoLottery s.dao.AddLotteryActionLog sid(%d) mid(%d) arg(%v) error(%v)", lottery.ID, mid, actionLog, err)
	}
	s.cache.Do(c, func(ctx context.Context) {
		s.sendLotteryAward(ctx, giftMap, gid, lottery.ID, mid, ip, lottery, lotteryInfo)
	})
	return
}

func (s *Service) FissionUpLotteryNum(c context.Context, sid string, incrNum int64) (af int64, err error) {
	caller := metadata.String(c, metadata.Caller)
	if caller != s.c.Fission.UpCaller {
		err = xecode.AccessDenied
		log.Warn("FissionUpLotteryNum Not Authorised Caller %s", caller)
		return
	}
	if _, ok := s.c.Fission.Sids[sid]; !ok {
		err = xecode.AccessDenied
		log.Warn("FissionUpLotteryNum sid:%s not found", sid)
		return
	}
	if af, err = s.lottDao.UpdateGiftNum(c, sid, incrNum); err != nil {
		log.Error("FissionUpLotteryNum UpdateGiftNum sid:%s incrNum:%d error(%v)", sid, incrNum, err)
	}
	return
}
