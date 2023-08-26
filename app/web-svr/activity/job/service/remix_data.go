package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	mdlmail "go-gateway/app/web-svr/activity/job/model/mail"
	mdlRank "go-gateway/app/web-svr/activity/job/model/rank"
	mdltask "go-gateway/app/web-svr/activity/job/model/task"

	"github.com/pkg/errors"
)

const (
	// remixMidTaskLength 一次获取用户任务情况数量
	remixMidTaskLength = 100
	// concurrencyMax 并发量
	concurrencyMax = 2
	// remixChannelLength 鬼畜结果channel长度
	remixChannelLength = 50
	// redisGetData 从redis中获取数据
	redisGetData = true
	// dbGetData 从db中获取数据
	dbGetData = false
	// remixTaskResultFileName 鬼畜任务完成情况
	remixTaskResultFileName = "鬼畜任务完成情况"
	// remixRankResultFileName 鬼畜排名结果
	remixRankResultFileName = "鬼畜排名结果"
)

// RemixDataResult 鬼畜统计数据结果
func (s *Service) RemixDataResult() {
	c := context.Background()
	s.remixDataRunning.Lock()
	defer s.remixDataRunning.Unlock()
	var (
		mids  []int64
		count int64
	)
	eg := errgroup.WithContext(c)
	// 获取所有参与活动的用户列表
	eg.Go(func(ctx context.Context) (err error) {
		mids, err = s.getAllMids(c, s.c.Remix.ChildSid)
		if err != nil {
			log.Warn("s.dao.AllDistinctMidBySids(%v)", err)
			return
		}
		return nil
	})
	// 获取人数
	eg.Go(func(ctx context.Context) (err error) {
		count, err = s.getActivityTaskCount(c, s.c.Remix.TaskID)
		if err != nil {
			log.Error("s.getActivityTaskCount: error(%v)", err)
			err = errors.Wrapf(err, "s.getActivityTaskCount")
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	// 统计用户获得金额
	err := s.remixMoneyResult(c, s.c.Remix.Sid, s.c.Remix.TaskID, mids)
	if err != nil {
		log.Error("s.moneyResult error(%v)", err)
	}
	// 统计排行数据
	err2 := s.remixMidRankResult(c, mids, s.c.Remix.Sid, s.c.Remix.StartBatch, s.c.Remix.LastBatch)
	if err2 != nil {
		log.Error("s.midRankResult error(%v)", err)
	}
	log.Infoc(c, "s.getActivityTaskCount() count(%d)", count)
}

// midRankResult 排行统计
func (s *Service) remixMidRankResult(c context.Context, mids []int64, sid, startBatch, endBatch int64) error {
	var (
		rankLastBatch     map[int64]*mdlRank.DB
		memberRankTimes   map[int64]*mdlRank.MemberRankTimes
		memberRankHighest map[int64]*mdlRank.MemberRankHighest
	)
	eg := errgroup.WithContext(c)
	// 获取最后一次排行情况
	eg.Go(func(ctx context.Context) (err error) {
		rankLastBatch, err = s.getRankListByBatch(c, sid, endBatch)
		return
	})
	// 获取mid上榜次数
	eg.Go(func(ctx context.Context) (err error) {
		memberRankTimes, err = s.countMidRankTimes(c, sid, startBatch, endBatch, mids)
		return
	})
	// 获取最好名次
	eg.Go(func(ctx context.Context) (err error) {
		memberRankHighest, err = s.countMidRankHighest(c, sid, startBatch, endBatch, mids)
		return
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	return s.remixOperateMemberInfoToCsv(c, mids, rankLastBatch, memberRankTimes, memberRankHighest)
}

func (s *Service) remixOperateMemberInfoToCsv(c context.Context, mids []int64, rankLastBatch map[int64]*mdlRank.DB, memberRankTimes map[int64]*mdlRank.MemberRankTimes, memberRankHighest map[int64]*mdlRank.MemberRankHighest) error {
	categoryHeader := []string{"用户ID", "最终名次", "最终积分", "最好名次", "上榜次数"}
	data := [][]string{}
	for _, v := range mids {
		rows := []string{}
		var midStr, lastRank, score, highest, times string
		midStr = strconv.FormatInt(v, 10)
		rankLastStruct, ok := rankLastBatch[v]
		if ok {
			lastRank = strconv.Itoa(rankLastStruct.Rank)
			score = strconv.FormatInt(rankLastStruct.Score, 10)
		}
		rankTimes, ok := memberRankTimes[v]
		if ok {
			times = strconv.Itoa(rankTimes.Times)
		}
		rankHighest, ok := memberRankHighest[v]
		if ok {
			highest = strconv.Itoa(rankHighest.Rank)
		}

		rows = append(rows, midStr, lastRank, score, highest, times)
		data = append(data, rows)
	}
	fileName := fmt.Sprintf("%v_%v.csv", remixRankResultFileName, time.Now().Format("200601021504"))
	err := s.remixCreateCsvAndSend(c, s.c.Remix.FilePath, fileName, remixRankResultFileName, categoryHeader, data)
	if err != nil {
		return err
	}
	return nil
}

// moneyResult 统计用户获得金额
func (s *Service) remixMoneyResult(c context.Context, sid, taskID int64, mids []int64) error {
	ch := make(chan map[int64]*mdltask.UserState, remixChannelLength)
	var midsAward []*mdltask.UserState
	eg := errgroup.WithContext(c)
	now := time.Now().Unix()
	eg.Go(func(ctx context.Context) (err error) {
		if now > s.c.Remix.ActivityEnd.Unix() {
			return s.remixMoneyResultIntoChannel(c, sid, taskID, mids, dbGetData, ch)
		}
		return s.remixMoneyResultIntoChannel(c, sid, taskID, mids, redisGetData, ch)
	})
	eg.Go(func(ctx context.Context) (err error) {
		midsAward = s.remixMoneyResultOutChannel(c, mids, ch)
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	return s.remixMoneyResultCsv(c, midsAward)
}

// moneyResultIntoChannel 金额统计结果进入channel
func (s *Service) remixMoneyResultIntoChannel(c context.Context, sid, taskID int64, mids []int64, fromRedis bool, ch chan map[int64]*mdltask.UserState) error {
	var times int
	patch := remixMidTaskLength
	concurrency := concurrencyMax
	times = len(mids) / patch / concurrency
	defer close(ch)
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			b := batch
			i := index
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(mids) {
					return nil
				}
				reqMids := mids[start:]
				end := start + patch
				if end < len(mids) {
					reqMids = mids[start:end]
				}
				if len(reqMids) > 0 {
					if fromRedis {
						reply, err := s.taskUserStateByRedis(ctx, taskID, reqMids)
						if err != nil || reply == nil {
							log.Error("s.taskUserStateByRedis: error(%v) sid(%d), taskID(%d) mids(%v)", err, sid, taskID, reqMids)
							return err
						}
						ch <- reply
						return nil
					}
					reply, err := s.taskUserState(ctx, sid, taskID, reqMids)
					if err != nil || reply == nil {
						log.Error("s.taskUserState: error(%v) sid(%d), taskID(%d) mids(%v)", err, sid, taskID, reqMids)
						return err
					}
					ch <- reply
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Error("eg.Wait error(%v)", err)
			return ecode.ActivityWriteHandMemberErr
		}
	}
	return nil
}

func (s *Service) remixMoneyResultCsv(c context.Context, userState []*mdltask.UserState) error {
	categoryHeader := []string{"用户ID", "任务完成", "最终金额:单位（分）"}
	data := [][]string{}
	var canGet int64
	var count int64
	var everyCanGet int64
	for _, v := range userState {
		if v.Finish == mdltask.HasFinish {
			count++
		}
	}
	if count > 0 {
		everyCanGet = s.c.Remix.Money / count
	}
	for _, v := range userState {
		rows := []string{}
		if v.Finish != mdltask.HasFinish {
			canGet = 0
		}
		if v.Finish == mdltask.HasFinish {
			canGet = everyCanGet
		}
		rows = append(rows, fmt.Sprintf("%d", v.MID), fmt.Sprintf("%d", v.Finish), fmt.Sprintf("%d", canGet))
		data = append(data, rows)

	}
	fileName := fmt.Sprintf("%v_%v.csv", remixTaskResultFileName, time.Now().Format("200601021504"))
	err := s.remixCreateCsvAndSend(c, s.c.Remix.FilePath, fileName, remixTaskResultFileName, categoryHeader, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) remixCreateCsvAndSend(c context.Context, filePath, fileName string, subject string, categoryHeader []string, data [][]string) error {
	base := &mdlmail.Base{
		Host:    s.c.Remix.Mail.Host,
		Port:    s.c.Remix.Mail.Port,
		Address: s.c.Remix.Mail.Address,
		Pwd:     s.c.Remix.Mail.Pwd,
		Name:    s.c.Remix.Mail.Name,
	}
	toAddress := []*mdlmail.Address{}
	for _, v := range s.c.Remix.Mail.ToAddress {
		toAddress = append(toAddress, &mdlmail.Address{
			Address: v.Address,
			Name:    v.Name,
		})
	}
	ccAddress := []*mdlmail.Address{}
	for _, v := range s.c.Remix.Mail.CcAddress {
		ccAddress = append(ccAddress, &mdlmail.Address{
			Address: v.Address,
			Name:    v.Name,
		})
	}
	bccAddress := []*mdlmail.Address{}
	for _, v := range s.c.Remix.Mail.BccAddresses {
		bccAddress = append(bccAddress, &mdlmail.Address{
			Address: v.Address,
			Name:    v.Name,
		})
	}
	return s.activityCreateCsvAndSend(c, filePath, fileName, subject, base, toAddress, ccAddress, bccAddress, categoryHeader, data)
}

// remixMoneyResultOutChannel channel中统计出金额
func (s *Service) remixMoneyResultOutChannel(c context.Context, mids []int64, channel chan map[int64]*mdltask.UserState) []*mdltask.UserState {
	allMidUserState := make(map[int64]*mdltask.UserState)
	res := make([]*mdltask.UserState, 0)
	for ch := range channel {
		for _, v := range ch {
			allMidUserState[v.MID] = v
		}
	}
	for _, v := range mids {
		if userState, ok := allMidUserState[v]; ok {
			res = append(res, userState)
		}
		res = append(res, &mdltask.UserState{MID: v})
	}
	return res
}
