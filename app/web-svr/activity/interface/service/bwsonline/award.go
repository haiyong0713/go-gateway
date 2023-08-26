package bwsonline

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/client"
	"strconv"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"

	garbapi "git.bilibili.co/bapis/bapis-go/garb/service"
)

func (s *Service) AwardPackageList(ctx context.Context, mid, bid int64) ([]*bwsonline.AwardPackageDetail, error) {
	packIDs, err := s.dao.AwardPackageList(ctx, bid)
	if err != nil {
		log.Errorc(ctx, "AwardPackageList AwardPackageList error:%v", err)
		return nil, err
	}
	if len(packIDs) == 0 {
		return []*bwsonline.AwardPackageDetail{}, nil
	}
	packages, err := s.dao.AwardPackageByIDs(ctx, packIDs)
	if err != nil {
		log.Errorc(ctx, "AwardPackageList AwardPackageByIDs error:%v", err)
		return nil, err
	}
	var awardIDs, packageIDs []int64
	for _, v := range packages {
		if v == nil || len(v.AwardIds) == 0 {
			continue
		}
		packageIDs = append(packageIDs, v.ID)
		awardIDs = append(awardIDs, v.AwardIds...)
	}
	awardIDs = filterIDs(awardIDs)
	if len(awardIDs) == 0 {
		return []*bwsonline.AwardPackageDetail{}, nil
	}
	var userPackage map[int64]map[int64]struct{}
	if mid > 0 {
		userPackage = func() map[int64]map[int64]struct{} {
			userPackages, packErr := s.dao.UserPackage(ctx, mid)
			if packErr != nil {
				log.Errorc(ctx, "AwardPackageList UserPackage mid:%d error:%v", mid, packErr)
				return nil
			}
			userPackageMap := make(map[int64]map[int64]struct{}, len(userPackages))
			for _, v := range userPackages {
				packageInfo, ok := userPackageMap[v.ID]
				if !ok {
					packageInfo = make(map[int64]struct{})
				}
				for _, id := range v.AwardIds {
					packageInfo[id] = struct{}{}
				}
				userPackageMap[v.ID] = packageInfo
			}
			return userPackageMap
		}()
	}
	awards, err := s.dao.AwardByIDs(ctx, awardIDs)
	if err != nil {
		log.Errorc(ctx, "AwardPackageList s.dao.AwardByIDs ids:%v error:%v", awardIDs, err)
		return nil, err
	}
	var res []*bwsonline.AwardPackageDetail
	for _, packID := range packIDs {
		v, ok := packages[packID]
		if !ok || v == nil {
			continue
		}
		tmp := &bwsonline.AwardPackageDetail{
			AwardPackage: v,
		}
		funcAwardOwned := func(int64) int64 {
			return 0
		}
		if _, ok := userPackage[v.ID]; ok {
			tmp.Owned = bwsonline.AwardPackageOwned
			funcAwardOwned = func(awardID int64) int64 {
				if _, ok := userPackage[v.ID][awardID]; ok {
					return 1
				}
				return 0
			}
		}
		for _, id := range v.AwardIds {
			if award, ok := awards[id]; ok && award != nil {
				owned := funcAwardOwned(id)
				tmp.Items = append(tmp.Items, &bwsonline.AwardPackageItem{
					Award: award,
					Owned: owned,
				})
				if award.TypeId == bwsonline.AwardTypeDress {
					tmp.Total++
					if owned == 1 {
						tmp.Awarded++
					}
				}
			}
		}
		res = append(res, tmp)
	}
	return res, nil
}

func (s *Service) MyAwardList(ctx context.Context, mid, bid int64) ([]*bwsonline.UserAward, error) {
	userAward, err := s.dao.UserAward(ctx, mid, bid)
	if err != nil {
		log.Errorc(ctx, "TicketReward UserAward mid:%d error:%v", mid, err)
		return nil, err
	}
	if len(userAward) == 0 {
		return []*bwsonline.UserAward{}, nil
	}
	var awardIDs []int64
	for _, item := range userAward {
		if item == nil {
			continue
		}
		awardIDs = append(awardIDs, item.ID)
	}
	awards, err := s.dao.AwardByIDs(ctx, awardIDs)
	if err != nil {
		log.Errorc(ctx, "AwardPackageList s.dao.AwardByIDs ids:%v error:%v", awardIDs, err)
		return nil, err
	}
	var res []*bwsonline.UserAward
	for _, item := range userAward {
		if item == nil {
			continue
		}
		if award, ok := awards[item.ID]; !ok || award == nil {
			continue
		}
		res = append(res, &bwsonline.UserAward{Award: awards[item.ID], State: item.State})
	}
	return res, nil
}

func (s *Service) AwardPackageReward(ctx context.Context, mid, id, bid int64) ([]*bwsonline.Award, error) {
	data, err := s.dao.AwardPackage(ctx, id)
	if err != nil {
		log.Errorc(ctx, "AwardPackageReward AwardPackage:%d error:%v", id, err)
		return nil, err
	}
	if data == nil || data.TypeId != bwsonline.PackageTypeAward || len(data.AwardIds) == 0 || data.Price == 0 || data.Bid != bid {
		return nil, ecode.BwsOnlinePackageNoAward
	}
	curr, err := s.dao.UserCurrency(ctx, mid, bid)
	if err != nil {
		log.Errorc(ctx, "AwardPackageReward UserCurrency mid:%d error:%v", mid, err)
		return nil, err
	}
	if curr[bwsonline.CurrTypeCoin] < data.Price {
		return nil, ecode.BwsOnlineCoinLow
	}
	awards, err := s.dao.AwardByIDs(ctx, data.AwardIds)
	if err != nil {
		log.Errorc(ctx, "AwardPackageReward AwardByIDs awardIDs(%v) error:%v", data.AwardIds, err)
		return nil, err
	}
	var (
		addAwardIDs   []int64
		addDressIDs   []int64
		rewardAwards  []*bwsonline.Award
		dressAwardMap = make(map[int64]*bwsonline.Award)
	)
	for _, id := range data.AwardIds {
		if v, ok := awards[id]; ok {
			if v == nil {
				continue
			}
			if v.TypeId == bwsonline.AwardTypeDress {
				addDressID, _ := strconv.ParseInt(v.Token, 10, 64)
				addDressIDs = append(addDressIDs, addDressID)
				dressAwardMap[addDressID] = v
				continue
			}
			addAwardIDs = append(addAwardIDs, v.ID)
		}
	}
	var noDressIDs []int64
	if len(addDressIDs) > 0 {
		var userDress []*bwsonline.UserDress
		userDress, err = s.dao.UserDress(ctx, mid)
		if err != nil {
			log.Errorc(ctx, "AwardPackageReward UserDress mid:%d error:%v", mid, err)
			return nil, err
		}
		userDressMap := make(map[int64]*bwsonline.UserDress, len(userDress))
		for _, v := range userDress {
			if v == nil {
				continue
			}
			userDressMap[v.DressId] = v
		}
		for _, dressID := range addDressIDs {
			if _, ok := userDressMap[dressID]; !ok {
				noDressIDs = append(noDressIDs, dressID)
			}
		}
	}
	if len(noDressIDs) == 0 {
		return nil, ecode.BwsOnlineAwardAll
	}
	var noAddAwardIDs []int64
	if userAward, err := s.dao.UserAward(ctx, mid, bid); err != nil {
		log.Errorc(ctx, "AwardPackageReward UserAward mid:%d bid:%d error:%v", mid, bid, err)
		return nil, err
	} else {
		userAwardMap := make(map[int64]struct{})
		for _, award := range userAward {
			userAwardMap[award.ID] = struct{}{}
		}
		for _, awardID := range addAwardIDs {
			if _, ok := userAwardMap[awardID]; !ok {
				noAddAwardIDs = append(noAddAwardIDs, awardID)
				rewardAwards = append(rewardAwards, awards[awardID])
			}
		}
	}

	if err = s.upUserCurrency(ctx, mid, bwsonline.CurrTypeCoin, 0, -data.Price, bid); err != nil {
		log.Errorc(ctx, "AwardPackageReward upUserCurrency mid:%d id:%d price:%d error:%v", mid, id, data.Price, err)
		return nil, err
	}
	if _, err = s.dao.AddUserAwardPackage(ctx, mid, id, append(noAddAwardIDs, dressAwardMap[noDressIDs[0]].ID)); err != nil {
		log.Errorc(ctx, "AwardPackageReward AddUserAwardPackage mid:%d id:%d error:%v", mid, id, err)
		return nil, err
	}
	if len(noAddAwardIDs) > 0 {
		if _, err = s.dao.AddUserAward(ctx, mid, noAddAwardIDs, bid); err != nil {
			log.Errorc(ctx, "AwardPackageReward AddUserAward mid:%d awardIDs:%v error:%v", mid, data.AwardIds, err)
			return nil, err
		}
	}
	if _, err = s.dao.DressAdd(ctx, mid, []int64{noDressIDs[0]}); err != nil {
		log.Errorc(ctx, "AwardPackageReward DressAdd mid:%d dressID:%d error:%v", mid, noDressIDs[0], err)
		return nil, err
	}
	if noDressIDs[0] > 0 {
		dressInfo, err := s.dao.Dress(ctx, noDressIDs[0])
		if err != nil {
			log.Errorc(ctx, "AwardPackageReward dressInfo dressID:%d error:%v", noDressIDs[0], err)
		}
		if dressInfo != nil {
			rewardAwards = append(rewardAwards, &bwsonline.Award{
				Title:  dressInfo.Title,
				Image:  dressInfo.Image,
				TypeId: bwsonline.AwardTypeDress,
			})
		}
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		s.dao.DelCacheUserPackage(ctx, mid)
		s.dao.DelCacheUserAward(ctx, mid, bid)
		s.dao.DelCacheUserDress(ctx, mid)
	})
	return rewardAwards, nil
}

func (s *Service) TicketReward(ctx context.Context, mid, id, bid int64) error {
	userAward, err := s.dao.UserAward(ctx, mid, bid)
	if err != nil {
		log.Errorc(ctx, "TicketReward UserAward mid:%d error:%v", mid, err)
		return err
	}
	awardState, ok := func() (int64, bool) {
		for _, v := range userAward {
			if v == nil {
				continue
			}
			if v.ID == id {
				return v.State, true
			}
		}
		return 0, false
	}()
	if !ok {
		return ecode.BwsOnlineNotReward
	}
	if awardState == bwsonline.HadReward {
		return ecode.BwsOnlineAwardUsed
	}
	award, err := s.dao.Award(ctx, id)
	if err != nil {
		log.Errorc(ctx, "TicketReward Award id:%d error:%v", id, err)
		return err
	}
	if award == nil {
		return xecode.NothingFound
	}
	if _, err = s.dao.UpUserAward(ctx, mid, id); err != nil {
		log.Errorc(ctx, "TicketReward UpUserAward mid:%d id:%d error:%v", mid, id, err)
		return err
	}
	// send award
	if err = func() error {
		switch award.TypeId {
		case bwsonline.AwardTypeBBQ:
			pendentID, _ := strconv.ParseInt(award.Token, 10, 64)
			_, err = client.GarbClient.GrantByBiz(ctx, &garbapi.GrantByBizReq{Mids: []int64{mid}, Ids: []int64{pendentID}, AddSecond: award.Expire})
			return err
		case bwsonline.AwardTypeVip:
			_, err = s.lottDao.MemberCoupon(ctx, mid, award.Token)
			return err
		case bwsonline.AwardTypeSuit:
			{
				suitID, _ := strconv.ParseInt(award.Token, 10, 64)
				_, err = client.GarbClient.GrantSuit(ctx, &garbapi.GrantSuitReq{
					Mids:      []int64{mid},
					SuitID:    suitID,
					AddSecond: 7 * 86400,
					Token:     fmt.Sprintf("bws:%d:%d:%d", mid, bid, id),
					Business:  "bws乐园兑换装扮",
				})
				return err
			}
		default:
			log.Errorc(ctx, "TicketReward award:%+v type error", award)
			return ecode.BwsOnlineNotRewardType
		}
	}(); err != nil {
		log.Errorc(ctx, "TicketReward Send Award mid:%d award:%v error:%v", mid, award, err)
		return err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		s.dao.DelCacheUserAward(ctx, mid, bid)
	})
	return nil
}

func (s *Service) ReserveAward(ctx context.Context, mid int64) error {
	if _, err := s.dao.DressAdd(ctx, mid, s.c.BwsOnline.ReserveAward); err != nil {
		log.Errorc(ctx, "ReserveAward s.dao.DressAdd(ctx, %d, %v) err[%v]", mid, s.c.BwsOnline.ReserveAward, err)
		return err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		s.dao.DelCacheUserDress(ctx, mid)
	})
	return nil
}
