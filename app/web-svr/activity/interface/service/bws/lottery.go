package bws

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"

	"github.com/pkg/errors"
)

// Lottery get lottery account.
func (s *Service) Lottery(c context.Context, bid, loginMid, aid int64, day string) (data *bwsmdl.LotteryUser, err error) {
	var (
		mid     int64
		accData *accapi.InfoReply
	)
	if _, ok := s.lotteryMids[loginMid]; !ok {
		err = ecode.ActivityNotLotteryAdmin
		return
	}
	if _, ok := s.lotteryAids[aid]; !ok {
		err = ecode.ActivityNotLotteryAchieve
		return
	}
	if _, err = s.Achievement(c, &bwsmdl.ParamID{Bid: bid, ID: aid}); err != nil {
		return
	}
	if mid, err = s.dao.CacheLotteryMid(c, aid, day); err != nil || mid == 0 {
		err = ecode.ActivityLotteryFail
		return
	}
	log.Warnc(c, "Lottery bid(%d) loginMid(%d) aid(%d) lotteryMid(%d)", bid, loginMid, aid, mid)
	data = &bwsmdl.LotteryUser{Mid: mid}
	if accData, err = s.accClient.Info3(c, &accapi.MidReq{Mid: mid}); err != nil {
		log.Errorc(c, "Lottery s.accRPC.Info3(%d) error(%v)", mid, err)
		err = nil
		return
	}
	if accData != nil && accData.Info != nil {
		data = &bwsmdl.LotteryUser{Mid: mid, Name: accData.Info.Name, Face: accData.Info.Face}
	}
	return
}

// LotteryCheck .
func (s *Service) LotteryCheck(c context.Context, mid, aid int64, day string) (data []int64, err error) {
	if !s.isAdmin(mid) {
		err = ecode.ActivityNotAdmin
		return
	}
	return s.dao.CacheLotteryMids(c, aid, day)
}

// LotteryV2 .
func (s *Service) LotteryV1(c context.Context, bid, awardID, typ int64) (data []*bwsmdl.LotteryUser, err error) {
	var (
		lotteryUser []*bwsmdl.LotteryCache
		mids        []int64
		reply       *accapi.InfosReply
	)
	// 特殊抽奖
	if typ != 0 {
		lotteryUser, err = s.dao.CacheLotterySpec(c, bid)
	} else {
		lotteryUser, err = s.dao.CacheLotteryV1(c, bid, awardID)
	}
	if err != nil {
		err = ecode.ActivityLotteryFail
		return
	}
	for _, v := range lotteryUser {
		if v.Mid > 0 {
			mids = append(mids, v.Mid)
		}
	}
	if len(mids) == 0 {
		err = ecode.ActivityLotteryFail
		return
	}
	if reply, err = s.accClient.Infos3(c, &accapi.MidsReq{Mids: mids}); err != nil {
		err = nil
	}
	for _, v := range lotteryUser {
		item := &bwsmdl.LotteryUser{Mid: v.Mid, AchieveRank: v.Rank}
		if reply != nil {
			if info, ok := reply.Infos[v.Mid]; ok {
				item.Name = info.Name
				item.Face = info.Face
			}
		}
		data = append(data, item)
	}
	return
}

func (s *Service) userLotteryLog(ctx context.Context, userToken string) (offline []*bwsmdl.LotteryLog, online []*bwsmdl.LotteryLog, err error) {
	userAwards, err := s.dao.UserAward(ctx, userToken)
	if err != nil {
		return nil, nil, err
	}
	for _, v := range userAwards {
		if v == nil || v.ID <= 0 || v.AwardId <= 0 {
			continue
		}
		if award, ok := s.bwsAllAwards[v.AwardId]; ok && award != nil {
			switch award.IsOnline {
			case 0:
				offline = append(offline, &bwsmdl.LotteryLog{
					ID:      v.ID,
					AwardID: v.AwardId,
					State:   v.State,
					Title:   award.Title,
					Stage:   award.Stage,
					Intro:   award.Intro,
					Image:   award.Image,
					Amount:  1,
					Ctime:   v.Ctime,
					Mtime:   v.Mtime,
				})
			case 1: //前期线上活动奖励
				online = append(online, &bwsmdl.LotteryLog{
					ID:      v.ID,
					AwardID: v.AwardId,
					State:   v.State,
					Title:   award.Title,
					Stage:   award.Stage,
					Intro:   award.Intro,
					Image:   award.Image,
					Amount:  1,
					Ctime:   v.Ctime,
					Mtime:   v.Mtime,
				})
			default:
				log.Warnc(ctx, "userLotteryLog unexpected isOnline award:%+v", award)
				continue
			}
		}
	}
	return offline, online, nil
}

func (s *Service) Lottery2020(ctx context.Context, mid int64, bid int64) (*bwsmdl.Award, error) {
	memberRly, err := s.accClient.Profile3(ctx, &accapi.MidReq{Mid: mid})
	if err != nil || memberRly == nil || memberRly.Profile == nil {
		err = errors.Wrapf(err, "s.accRPC.Profile3(c,&accmdl.ArgMid{Mid:%d})", mid)
		return nil, err
	}

	userToken, err := s.midToKey(ctx, bid, mid)
	if err != nil {
		return nil, err
	}
	_, dayStr := todayDate()

	_, lotteryTimes, err := s.UserLotteryTimes(ctx, bid, mid, dayStr)
	if err != nil || lotteryTimes < s.c.Bws.LotteryUsed {
		log.Errorc(ctx, "Lottery2020 UserLotteryTimes userToken:%s error:%v", dayStr, err)
		return nil, ecode.ActivityNoTimes
	}
	if len(s.bwsAllAwards) == 0 {
		return nil, ecode.BwsNoAward
	}
	var hitSpecialAward bool
	rand.Seed(time.Now().Unix())
	if i := rand.Intn(100); i < s.c.Bws.Bws20Rand {
		hitSpecialAward = true
	}

	hitAward := func() *bwsmdl.Award {
		var tmpAward *bwsmdl.Award
		if hitSpecialAward {
			if i := rand.Intn(100); i < s.c.Bws.Bws20Rand1 {
				for _, v := range s.bwsAllAwards {
					if v != nil && v.IsOnline == 0 && v.ID == s.c.Bws.StockAwardID {
						tmpAward = v
						break
					}
				}
			}

		}
		if tmpAward != nil && tmpAward.Stock != 0 {
			return tmpAward
		}
		for _, v := range s.bwsAllAwards {
			if v != nil && v.IsOnline == 0 && v.ID != s.c.Bws.StockAwardID && v.Stock != 0 {

				tmpAward = v
				break
			}
		}
		return tmpAward
	}()
	if hitAward == nil {
		log.Warnc(ctx, "Lottery2020 userToken:%s fail", userToken)
		return nil, ecode.ActivityLotteryFail
	}
	if hitAward.Stock != -1 {
		stock, err := s.dao.RawAwardStock(ctx, hitAward.ID)
		if err != nil {
			log.Errorc(ctx, "Lottery2020 RawAwardStock awardID:%d error:%v", hitAward.ID, err)
			return nil, err
		}
		if stock <= 0 {
			// choose other unlimited award
			var tmpAward *bwsmdl.Award
			for _, v := range s.bwsAllAwards {
				if v != nil && v.IsOnline == 0 && v.ID != hitAward.ID && v.Stock == -1 {
					tmpAward = v
					break
				}
			}
			if tmpAward == nil {
				log.Warnc(ctx, "Lottery2020 userToken:%s fail", userToken)
				return nil, ecode.ActivityLotteryFail
			}
			hitAward = tmpAward
		}
	}
	if hitAward.Stock != -1 {
		// decr stock
		affected, err := s.dao.DecrAwardStock(ctx, hitAward.ID)
		if err != nil {
			log.Errorc(ctx, "Lottery2020 DecrAwardStock fail award:%d error:%v", hitAward.ID, err)
			return nil, err
		}
		// choose other unlimited award
		if affected <= 0 {
			var tmpAward *bwsmdl.Award
			for _, v := range s.bwsAllAwards {
				if v != nil && v.IsOnline == 0 && v.ID != hitAward.ID && v.Stock == -1 {
					tmpAward = v
					break
				}
			}
			if tmpAward == nil {
				log.Warnc(ctx, "Lottery2020 userToken:%s fail", userToken)
				return nil, ecode.ActivityLotteryFail
			}
			hitAward = tmpAward
		}
	}
	// 使用抽奖次数
	err = s.UpdateUserDetail(ctx, mid, bid, s.c.Bws.LotteryUsed, 0, dayStr, 0, nil, false, nil, "", userToken)
	if err != nil {
		log.Errorc(ctx, "s.UpdateUserDetail() err(%v)", err)
		return nil, err
	}
	state := bwsmdl.AwardStateInit
	if _, err = s.dao.AddUserAward(ctx, userToken, hitAward.ID, state); err != nil {
		log.Errorc(ctx, "Lottery2020 AddUserAward usarderToken:%s award:%d error:%v", userToken, hitAward.ID, err)
		return nil, err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		retry(func() error {
			return s.dao.DelCacheUserAward(ctx, userToken)
		})
		retry(func() error {
			return s.dao.DelCacheUserLotteryTimes(ctx, userToken)
		})
	})
	return hitAward, nil
}

func (s *Service) sendCoupon(c context.Context, orderID string, ip string, mid int64, name string) (err error) {
	err = s.lotDao.SendVipBuyCoupon(c, ip, s.c.Bws.VipBuyToken, s.c.Lottery.VipBuy.SourceActivityID, orderID, name, s.c.Lottery.VipBuy.SourceID, mid)
	if err != nil {
		log.Errorc(c, "s.SendVipBuyCoupon(%d,%s,%s,%s,%s) error(%v)", mid, ip, s.c.Bws.VipBuyToken, orderID, name, err)
		err = fmt.Errorf("s.SendVipBuyCoupon(%d,%s,%s,%s,%s) error(%v)", mid, ip, s.c.Bws.VipBuyToken, orderID, name, err)
		return err
	}
	return nil
}

func (s *Service) AwardList(_ context.Context) ([]*bwsmdl.Award, error) {
	var res []*bwsmdl.Award
	for _, v := range s.bwsAllAwards {
		if v != nil && v.IsOnline == 0 {
			res = append(res, v)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res, nil
}

// AwardSend 奖品发放
func (s *Service) AwardSend(ctx context.Context, bid, loginMid, mid, id, awardID int64, key string) error {
	awardInfo, ok := s.bwsAllAwards[awardID]
	if !ok || awardInfo == nil {
		log.Errorc(ctx, "AwardSend award not exist awardID:%d", awardID)
		return xecode.NothingFound
	}
	if awardInfo.Owner != loginMid && !s.isAdmin(loginMid) {
		return ecode.ActivityNotOwner
	}
	userToken := key
	var err error
	if userToken == "" {
		userToken, err = s.midToKey(ctx, bid, mid)
		if err != nil {
			return err
		}
	}
	userAwards, err := s.dao.UserAward(ctx, userToken)
	if err != nil {
		log.Errorc(ctx, "AwardSend UserAward userToken:%s error:%v", userToken, err)
		return err
	}
	if len(userAwards) == 0 {
		return ecode.ActivityNoAward
	}
	var currAward *bwsmdl.UserAward
	for _, v := range userAwards {
		if v != nil && v.ID == id {
			currAward = v
		}
	}
	if currAward == nil {
		return ecode.ActivityNoAward
	}
	if currAward.AwardId != awardID {
		return ecode.ActivityAwardNotExpected
	}
	if currAward.State == bwsmdl.AwardStateFinish {
		return ecode.ActivityHasAward
	}
	preState := bwsmdl.AwardStateInit
	if awardID == s.c.Bws.StockAwardID {
		if currAward.State != bwsmdl.AwardStatePending {
			return ecode.BawAwardNeedCheck
		}
		preState = bwsmdl.AwardStatePending
	}
	affected, err := s.dao.UpUserAward(ctx, id, userToken, bwsmdl.AwardStateFinish, preState)
	if err != nil {
		log.Warn("AwardSend UpUserAward userToken:%s error:%+v", userToken, err)
		return err
	}
	if affected <= 0 {
		log.Warn("AwardSend UpUserAward userToken:%s no affected", userToken)
		return nil
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		retry(func() error {
			return s.dao.DelCacheUserAward(ctx, userToken)
		})
	})
	return nil
}

// OfflineAwardSend 线下奖品发放
func (s *Service) OfflineAwardSend(ctx context.Context, bid, loginMid, id, awardID int64) error {
	awardInfo, ok := s.bwsAllAwards[awardID]
	if !ok || awardInfo == nil {
		log.Errorc(ctx, "OfflineAwardSend awardID:%d not found", awardID)
		return xecode.NothingFound
	}
	if awardInfo.ID != s.c.Bws.StockAwardID && awardInfo.ID != s.c.Bws.StockAwardID2 {
		return xecode.RequestErr
	}
	userToken, err := s.midToKey(ctx, bid, loginMid)
	if err != nil {
		return err
	}
	userAwards, err := s.dao.UserAward(ctx, userToken)
	if err != nil {
		log.Errorc(ctx, "OfflineAwardSend UserAward userToken:%s error:%v", userToken, err)
		return err
	}
	if len(userAwards) == 0 {
		return ecode.ActivityNoAward
	}
	var currAward *bwsmdl.UserAward
	for _, v := range userAwards {
		if v != nil && v.ID == id {
			currAward = v
		}
	}
	if currAward == nil {
		return ecode.ActivityNoAward
	}
	if currAward.AwardId != awardID {
		return ecode.ActivityAwardNotExpected
	}
	if currAward.State == bwsmdl.AwardStateFinish || currAward.State == bwsmdl.AwardStatePending {
		return ecode.ActivityHasAward
	}
	preState := bwsmdl.AwardStateInit
	affected, err := s.dao.UpUserAward(ctx, id, userToken, bwsmdl.AwardStatePending, preState)
	if err != nil {
		log.Errorc(ctx, "OfflineAwardSend UpUserAward userToken:%s error:%+v", userToken, err)
		return err
	}
	if affected <= 0 {
		log.Errorc(ctx, "OfflineAwardSend UpUserAward userToken:%s no affected", userToken)
		return nil
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		retry(func() error {
			return s.dao.DelCacheUserAward(ctx, userToken)
		})
	})
	return nil
}

// GetStore 获取实物奖库存
func (s *Service) GetStore(ctx context.Context) (int64, error) {
	if s.bwsAllAwards != nil {
		for _, v := range s.bwsAllAwards {
			if v.ID == s.c.Bws.StockAwardID {
				return v.Stock, nil
			}
		}
	}
	return 0, ecode.ActivityBwsStockErr
}

func (s *Service) loadAllAwards() {
	ctx := context.Background()
	awards, err := s.dao.RawAwardList(ctx)
	if err != nil {
		log.Errorc(ctx, "loadAllAwards RawMainTask error:%v", err)
		return
	}
	s.bwsAllAwards = awards
}

func (s *Service) AddUserAward(ctx context.Context, userToken string, awardID int64, state string) (giftID int64, err error) {
	if giftID, err = s.dao.AddUserAward(ctx, userToken, awardID, state); err != nil {
		log.Errorc(ctx, "AddUserAward usarderToken:%s award:%d error:%v", userToken, awardID, err)
	}
	return
}
