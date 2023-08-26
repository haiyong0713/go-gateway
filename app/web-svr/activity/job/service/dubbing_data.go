package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	mdlmail "go-gateway/app/web-svr/activity/job/model/mail"
	mdlRank "go-gateway/app/web-svr/activity/job/model/rank"
	mdltask "go-gateway/app/web-svr/activity/job/model/task"

	"github.com/pkg/errors"
)

const (
	// dubbingMidTaskLength 一次获取用户任务情况数量
	dubbingMidTaskLength = 100
	// dubbingConcurrencyMax 并发量
	dubbingConcurrencyMax = 2
	// dubbingChannelLength 配音结果channel长度
	dubbingChannelLength = 50
	// dubbingTaskResultFileName 配音任务完成情况
	dubbingTaskResultFileName = "配音任务完成情况"
	// dubbingRankResultFileName 配音排名结果
	dubbingRankResultFileName = "配音排名结果"
)

var dubbingDataCtx context.Context

func dubbingCtxInit() {
	dubbingDataCtx = trace.SimpleServerTrace(context.Background(), "dubbing data")
}

// DubbingData 鬼畜统计数据结果
func (s *Service) DubbingData() {
	s.dubbingDataRunning.Lock()
	defer s.dubbingDataRunning.Unlock()
	dubbingCtxInit()
	var (
		mids  []int64
		count int64
	)
	eg := errgroup.WithContext(dubbingDataCtx)
	// 获取所有参与活动的用户列表
	eg.Go(func(ctx context.Context) (err error) {
		mids, err = s.getAllMids(dubbingDataCtx, s.c.Dubbing.ChildSid)
		if err != nil {
			log.Warn("s.dao.AllDistinctMidBySids(%v)", err)
			return
		}
		return nil
	})
	// 获取人数
	eg.Go(func(ctx context.Context) (err error) {
		count, err = s.getActivityTaskCount(dubbingDataCtx, s.c.Dubbing.TaskID)
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
	err := s.dubbingMoneyResult(dubbingDataCtx, s.c.Dubbing.Sid, s.c.Dubbing.TaskID, mids)
	if err != nil {
		log.Error("s.moneyResult error(%v)", err)
	}
	// 统计排行数据
	err2 := s.dubbingMidRankResult(dubbingDataCtx, mids, s.c.Dubbing.Sid, s.c.Dubbing.StartBatch, s.c.Dubbing.LastBatch)
	if err2 != nil {
		log.Errorc(dubbingDataCtx, "s.midRankResult error(%v)", err)
	}
	log.Infoc(dubbingDataCtx, "s.getActivityTaskCount() count(%d)", count)
}

// dubbingMidRankResult 排行统计
func (s *Service) dubbingMidRankResult(c context.Context, mids []int64, sid, startBatch, endBatch int64) error {
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
	return s.dubbingOperateMemberInfoToCsv(c, mids, rankLastBatch, memberRankTimes, memberRankHighest)
}

func (s *Service) dubbingOperateMemberInfoToCsv(c context.Context, mids []int64, rankLastBatch map[int64]*mdlRank.DB, memberRankTimes map[int64]*mdlRank.MemberRankTimes, memberRankHighest map[int64]*mdlRank.MemberRankHighest) error {
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
	fileName := fmt.Sprintf("%v_%v.csv", dubbingRankResultFileName, time.Now().Format("200601021504"))
	err := s.dubbingCreateCsvAndSend(c, s.c.Dubbing.FilePath, fileName, dubbingRankResultFileName, categoryHeader, data)
	if err != nil {
		return err
	}
	return nil
}

// dubbingMoneyResult 统计用户获得金额
func (s *Service) dubbingMoneyResult(c context.Context, sid, taskID int64, mids []int64) error {
	ch := make(chan map[int64]*mdltask.UserState, dubbingChannelLength)
	var midsAward []*mdltask.UserState
	eg := errgroup.WithContext(c)
	now := time.Now().Unix()
	eg.Go(func(ctx context.Context) (err error) {
		if now > s.c.Dubbing.ActivityEnd.Unix() {
			return s.dubbingMoneyResultIntoChannel(c, sid, taskID, mids, dbGetData, ch)
		}
		return s.dubbingMoneyResultIntoChannel(c, sid, taskID, mids, redisGetData, ch)
	})
	eg.Go(func(ctx context.Context) (err error) {
		midsAward = s.dubbingMoneyResultOutChannel(c, mids, ch)
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	return s.dubbingMoneyResultCsv(c, midsAward)
}

// dubbingMoneyResultIntoChannel 金额统计结果进入channel
func (s *Service) dubbingMoneyResultIntoChannel(c context.Context, sid, taskID int64, mids []int64, fromRedis bool, ch chan map[int64]*mdltask.UserState) error {
	var times int
	patch := dubbingMidTaskLength
	concurrency := dubbingConcurrencyMax
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

func (s *Service) dubbingMoneyResultCsv(c context.Context, userState []*mdltask.UserState) error {
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
		everyCanGet = s.c.Dubbing.Money / count
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
	fileName := fmt.Sprintf("%v_%v.csv", dubbingTaskResultFileName, time.Now().Format("200601021504"))
	err := s.dubbingCreateCsvAndSend(c, s.c.Dubbing.FilePath, fileName, dubbingTaskResultFileName, categoryHeader, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) dubbingCreateCsvAndSend(c context.Context, filePath, fileName string, subject string, categoryHeader []string, data [][]string) error {
	base := &mdlmail.Base{
		Host:    s.c.Dubbing.Mail.Host,
		Port:    s.c.Dubbing.Mail.Port,
		Address: s.c.Dubbing.Mail.Address,
		Pwd:     s.c.Dubbing.Mail.Pwd,
		Name:    s.c.Dubbing.Mail.Name,
	}
	toAddress := []*mdlmail.Address{}
	for _, v := range s.c.Dubbing.Mail.ToAddress {
		toAddress = append(toAddress, &mdlmail.Address{
			Address: v.Address,
			Name:    v.Name,
		})
	}
	ccAddress := []*mdlmail.Address{}
	for _, v := range s.c.Dubbing.Mail.CcAddress {
		ccAddress = append(ccAddress, &mdlmail.Address{
			Address: v.Address,
			Name:    v.Name,
		})
	}
	bccAddress := []*mdlmail.Address{}
	for _, v := range s.c.Dubbing.Mail.BccAddresses {
		bccAddress = append(bccAddress, &mdlmail.Address{
			Address: v.Address,
			Name:    v.Name,
		})
	}
	return s.activityCreateCsvAndSend(c, filePath, fileName, subject, base, toAddress, ccAddress, bccAddress, categoryHeader, data)
}

// dubbingMoneyResultOutChannel channel中统计出金额
func (s *Service) dubbingMoneyResultOutChannel(c context.Context, mids []int64, channel chan map[int64]*mdltask.UserState) []*mdltask.UserState {
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
