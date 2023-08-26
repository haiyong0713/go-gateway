package handwrite

import (
	"context"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	hwMdl "go-gateway/app/web-svr/activity/interface/model/handwrite"

	"github.com/pkg/errors"
)

const (
	// maxMemberInfoLength 一次获取用户信息的数量
	maxMemberInfoLength = 50
	// channel Length
	memberInfoChannelLength = 2
)

// AwardMemberCount 获奖用户数据
func (s *Service) AwardMemberCount(c context.Context) (res *hwMdl.AwardMemberReply, err error) {
	res = s.awardMemberReply(c)
	awardCount, err := s.handwrite.GetAwardCount(c)
	if err != nil {
		err = errors.Wrapf(err, "s.handwrite.GetAwardCount")
		log.Error("s.handwrite.GetAwardCount error(%v)", err)
		return res, ecode.ActivityWriteHandActivityMemberErr
	}
	var godCount, newCount, tiredCount int
	if awardCount != nil {
		godCount = awardCount.God
		newCount = awardCount.New
		tiredCount = awardCount.Tired
	}
	res.MoneyCount[hwMdl.GodName] = &hwMdl.AwardCountMoney{Count: godCount, Money: s.c.HandWrite.GodAllMoney}
	res.MoneyCount[hwMdl.NewName] = &hwMdl.AwardCountMoney{Count: newCount, Money: s.c.HandWrite.NewAllMoney}
	res.MoneyCount[hwMdl.TiredName] = &hwMdl.AwardCountMoney{Count: tiredCount, Money: s.c.HandWrite.TiredAllMoney}
	return res, nil
}

// awardMemberReply 返回结构体构造
func (s *Service) awardMemberReply(c context.Context) *hwMdl.AwardMemberReply {
	resMap := map[string]*hwMdl.AwardCountMoney{}
	resMap[hwMdl.GodName] = &hwMdl.AwardCountMoney{}
	resMap[hwMdl.NewName] = &hwMdl.AwardCountMoney{}
	resMap[hwMdl.TiredName] = &hwMdl.AwardCountMoney{}
	res := &hwMdl.AwardMemberReply{
		MoneyCount: resMap,
	}
	return res
}

// Rank 排行榜接口
func (s *Service) Rank(c context.Context) (*hwMdl.RankReply, error) {
	res := &hwMdl.RankReply{}
	resRank := make([]*hwMdl.RankMember, 0)
	rankList, err := s.rank.GetRank(c, hwMdl.HandWriteKey)
	if err != nil {
		err = errors.Wrapf(err, "s.rank.GetRank")
		log.Error("s.rank.GetRank error(%v)", err)
		res.Rank = resRank
		return res, ecode.ActivityWriteHandActivityMemberErr
	}
	if rankList != nil {
		mids := []int64{}
		for _, v := range rankList {
			if v != nil {
				mids = append(mids, v.Mid)
			}
		}
		if len(mids) > 0 {
			memberInfo, err := s.memberInfo(c, mids)
			if err != nil {
				return res, ecode.ActivityWriteHandActivityMemberErr
			}
			for _, v := range rankList {
				if v != nil {
					if member, ok := memberInfo[v.Mid]; ok {
						resRank = append(resRank, &hwMdl.RankMember{
							Account: member,
							Score:   v.Score,
						})
					}
				}
			}
		}
	}
	res.Rank = resRank
	return res, nil
}

// memberInfo 用户信息
func (s *Service) memberInfo(c context.Context, mids []int64) (map[int64]*hwMdl.Account, error) {
	eg := errgroup.WithContext(c)
	midsInfo := make(map[int64]*hwMdl.Account)
	channel := make(chan map[int64]*accountapi.Info, memberInfoChannelLength)
	eg.Go(func(ctx context.Context) error {
		return s.memberInfoIntoChannel(c, mids, channel)
	})
	eg.Go(func(ctx context.Context) (err error) {
		midsInfo, err = s.memberInfoOutChannel(c, channel)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.memberNickname")
		return nil, err
	}

	return midsInfo, nil
}

func (s *Service) memberInfoIntoChannel(c context.Context, mids []int64, channel chan map[int64]*accountapi.Info) error {
	var times int
	eg := errgroup.WithContext(c)
	times = len(mids) / maxMemberInfoLength
	defer close(channel)
	for index := 0; index <= times; index++ {
		i := index
		eg.Go(func(ctx context.Context) error {
			if i*maxMemberInfoLength >= len(mids) {
				return nil
			}
			start := i * maxMemberInfoLength
			reqMids := mids[start:]
			if start+maxMemberInfoLength < len(mids) {
				reqMids = mids[start : start+maxMemberInfoLength]
			}
			if len(reqMids) > 0 {
				infosReply, err := s.accClient.Infos3(ctx, &accountapi.MidsReq{Mids: reqMids})
				if err != nil || infosReply == nil {
					log.Error("s.accClient.Infos3: error(%v) batch(%d)", err, i)
					return err
				}
				channel <- infosReply.Infos
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return ecode.ActivityWriteHandMemberInfoErr
	}

	return nil
}

func (s *Service) memberInfoOutChannel(c context.Context, channel chan map[int64]*accountapi.Info) (map[int64]*hwMdl.Account, error) {
	midsInfo := make(map[int64]*hwMdl.Account)
	for item := range channel {
		for mid, value := range item {
			if value != nil {
				midsInfo[mid] = &hwMdl.Account{
					Mid:  mid,
					Name: value.Name,
					Sex:  value.Sex,
					Face: value.Face,
					Sign: value.Sign,
				}
			}
		}
	}
	return midsInfo, nil
}

// Personal 个人情况
func (s *Service) Personal(c context.Context, mid int64) (*hwMdl.PersonalReply, error) {
	eg := errgroup.WithContext(c)
	var (
		midAward   *hwMdl.MidAward
		awardCount *hwMdl.AwardCount
		midInfo    *accountapi.InfoReply
	)
	account := &hwMdl.Account{}
	res := &hwMdl.PersonalReply{}
	res.Account = account
	// 获取用户个人获奖情况
	eg.Go(func(ctx context.Context) (err error) {
		midAward, err = s.handwrite.GetMidAward(c, mid)
		if err != nil {
			log.Error("s.handwrite.GetMidAward: error(%v)", err)
			err = errors.Wrapf(err, "s.handwrite.GetMidAward")
			return err
		}
		return nil
	})
	// 获取用户信息
	eg.Go(func(ctx context.Context) (err error) {
		midInfo, err = s.accClient.Info3(ctx, &accountapi.MidReq{Mid: mid})
		if err != nil {
			log.Error("s.accClient.Info3: error(%v)", err)
			err = errors.Wrapf(err, "s.accClient.Info3")
			return err
		}
		return nil
	})
	// 获取人数
	eg.Go(func(ctx context.Context) (err error) {
		awardCount, err = s.handwrite.GetAwardCount(c)
		if err != nil {
			log.Error("s.handwrite.GetAwardCount: error(%v)", err)
			err = errors.Wrapf(err, "s.handwrite.GetAwardCount")
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return res, err
	}
	if midAward != nil && awardCount != nil {
		res.Score = midAward.Score
		res.Rank = midAward.Rank
		money := 0
		if midAward.God > 0 {
			money += (s.c.HandWrite.GodAllMoney / awardCount.God) * midAward.God
		}
		if midAward.Tired == hwMdl.AwardCanGet {
			money += s.c.HandWrite.TiredAllMoney / awardCount.Tired
		}
		if midAward.New == hwMdl.AwardCanGet {
			money += s.c.HandWrite.NewAllMoney / awardCount.New
		}
		res.Money = money
	}

	if midInfo != nil && midInfo.Info != nil {
		account.Mid = midInfo.Info.Mid
		account.Sign = midInfo.Info.Sign
		account.Sex = midInfo.Info.Sex
		account.Face = midInfo.Info.Face
		account.Name = midInfo.Info.Name
		res.Account = account
	}
	return res, nil
}

// Personal2021 个人情况
func (s *Service) Personal2021(c context.Context, mid int64) (*hwMdl.Personal2021Reply, error) {
	eg := errgroup.WithContext(c)
	var (
		midAward   *hwMdl.MidTaskAll
		awardCount *hwMdl.AwardCountNew
	)
	res := &hwMdl.Personal2021Reply{}
	// 获取用户个人获奖情况
	eg.Go(func(ctx context.Context) (err error) {
		midAward, err = s.handwrite.GetMidTask(c, mid)
		if err != nil {
			log.Error("s.handwrite.GetMidTask: error(%v)", err)
			err = errors.Wrapf(err, "s.handwrite.GetMidAward")
			return err
		}
		return nil
	})
	// 获取人数
	eg.Go(func(ctx context.Context) (err error) {
		awardCount, err = s.handwrite.GetTaskCount(c)
		if err != nil {
			log.Error("s.handwrite.GetAwardCount: error(%v)", err)
			err = errors.Wrapf(err, "s.handwrite.GetAwardCount")
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return res, err
	}
	if midAward != nil && awardCount != nil {
		var money int64
		if midAward.God > 0 {
			money += (s.c.HandWrite2021.GodAllMoney / awardCount.God) * midAward.God
		}
		if midAward.TiredLevel1 == hwMdl.AwardCanGet {
			money += s.c.HandWrite2021.Tired1Money / awardCount.TiredLevel1
		}

		if midAward.TiredLevel2 == hwMdl.AwardCanGet {
			money += s.c.HandWrite2021.Tired2Money / awardCount.TiredLevel2
		}
		if midAward.TiredLevel3 == hwMdl.AwardCanGet {
			money += s.c.HandWrite2021.Tired3Money / awardCount.TiredLevel3
		}
		res.Money = money
	}
	return res, nil
}

// AwardMemberCount2021 获奖用户数据
func (s *Service) AwardMemberCount2021(c context.Context) (res *hwMdl.AwardMemberReply, err error) {
	res = s.awardMemberReply2021(c)
	awardCount, err := s.handwrite.GetTaskCount(c)
	if err != nil {
		err = errors.Wrapf(err, "s.handwrite.GetAwardCount")
		log.Error("s.handwrite.GetAwardCount error(%v)", err)
		return res, ecode.ActivityWriteHandActivityMemberErr
	}
	var godCount, tired1Count, tired2Count, tired3Count int
	if awardCount != nil {
		godCount = int(awardCount.God)
		tired1Count = int(awardCount.TiredLevel1)
		tired2Count = int(awardCount.TiredLevel2)
		tired3Count = int(awardCount.TiredLevel3)
	}
	var godMoney, tiredMoney1, tiredMoney2, tiredMoney3 int
	if godCount > 0 {
		godMoney = int(s.c.HandWrite2021.GodAllMoney) / godCount
	}
	if tired1Count > 0 {
		tiredMoney1 = int(s.c.HandWrite2021.Tired1Money) / tired1Count
	}
	if tired2Count > 0 {
		tiredMoney2 = int(s.c.HandWrite2021.Tired2Money) / tired2Count
	}
	if tired3Count > 0 {
		tiredMoney3 = int(s.c.HandWrite2021.Tired3Money) / tired3Count
	}
	res.MoneyCount[hwMdl.GodName] = &hwMdl.AwardCountMoney{Count: godCount, Money: godMoney}
	res.MoneyCount[hwMdl.Tired1Name] = &hwMdl.AwardCountMoney{Count: tired1Count, Money: tiredMoney1}
	res.MoneyCount[hwMdl.Tired2Name] = &hwMdl.AwardCountMoney{Count: tired2Count, Money: tiredMoney2}
	res.MoneyCount[hwMdl.Tired3Name] = &hwMdl.AwardCountMoney{Count: tired3Count, Money: tiredMoney3}
	return res, nil
}

// awardMemberReply 返回结构体构造
func (s *Service) awardMemberReply2021(c context.Context) *hwMdl.AwardMemberReply {
	resMap := map[string]*hwMdl.AwardCountMoney{}
	resMap[hwMdl.GodName] = &hwMdl.AwardCountMoney{}
	resMap[hwMdl.Tired1Name] = &hwMdl.AwardCountMoney{}
	resMap[hwMdl.Tired2Name] = &hwMdl.AwardCountMoney{}
	resMap[hwMdl.Tired3Name] = &hwMdl.AwardCountMoney{}
	res := &hwMdl.AwardMemberReply{
		MoneyCount: resMap,
	}
	return res
}
