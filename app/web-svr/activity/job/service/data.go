package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	mdlhw "go-gateway/app/web-svr/activity/job/model/handwrite"
	"go-gateway/app/web-svr/activity/job/model/like"
	mdlmail "go-gateway/app/web-svr/activity/job/model/mail"
	mdlRank "go-gateway/app/web-svr/activity/job/model/rank"

	"github.com/pkg/errors"
)

const (
	// maxScoreResultLength 一次获取用户积分长度
	maxScoreResultLength = 100
	// concurrencyMaxScoreResult 一次获取用户积分并发量
	concurrencyMaxScoreResult = 2
	// scoreResultChannelLength 积分结果channel长度
	scoreResultChannelLength = 50
	// maxRankDbMidLength 一次处理统计mid的数量
	maxRankDbMidLength = 1000
	// concurrencyMaxRankDbMid 处理统计mid并发数
	concurrencyMaxRankDbMid = 2
	// maxRankDbChannelLength 统计channel长度
	maxRankDbChannelLength = 50
	// handWorkMoneyResultFileName 手书活动-用户金额发放结果
	handWorkMoneyResultFileName = "手书活动-用户金额发放结果"
	// handWorkMidRankResultFileName 手书活动-用户积分排行结果
	handWorkMidRankResultFileName = "手书活动-用户积分排行结果"
)

// DataResult 统计数据结果
func (s *Service) DataResult() {
	c := context.Background()
	s.handWriteDataRunning.Lock()
	defer s.handWriteDataRunning.Unlock()
	var (
		midsList   []*like.Item
		awardCount *mdlhw.AwardCount
	)
	eg := errgroup.WithContext(c)
	// 获取所有参与活动的用户列表
	eg.Go(func(ctx context.Context) (err error) {
		midsList, err = s.dao.AllDistinctMid(c, s.c.HandWrite.Sid)
		if err != nil {
			log.Warn("s.dao.AllMid(%v)", err)
			return
		}
		return nil
	})
	// 获取人数
	eg.Go(func(ctx context.Context) (err error) {
		awardCount, err = s.handWrite.GetAwardCount(c)
		if err != nil {
			log.Error("s.handWrite.GetAwardCount: error(%v)", err)
			err = errors.Wrapf(err, "s.handWrite.GetAwardCount")
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	mids := []int64{}
	for _, v := range midsList {
		mids = append(mids, v.Mid)
	}
	// 统计用户获得金额
	err := s.moneyResult(c, mids, awardCount)
	if err != nil {
		log.Error("s.moneyResult error(%v)", err)
	}
	// 统计排行数据
	err2 := s.midRankResult(c, mids, s.c.HandWrite.Sid, s.c.HandWrite.RankStartBatch, s.c.HandWrite.RankLastBatch)
	if err2 != nil {
		log.Error("s.midRankResult error(%v)", err)
	}
}

// midRankResult 排行统计
func (s *Service) midRankResult(c context.Context, mids []int64, sid, startBatch, endBatch int64) error {
	var (
		rankLastBatch     map[int64]*mdlRank.DB
		memberRankTimes   map[int64]*mdlRank.MemberRankTimes
		memberRankHighest map[int64]*mdlRank.MemberRankHighest
		memberInitFans    map[int64]int64
		memberEndFans     map[int64]int64
		memberNickname    map[int64]string
	)
	eg := errgroup.WithContext(c)
	// 获取最后一次排行情况
	eg.Go(func(ctx context.Context) (err error) {
		memberNickname, err = s.memberNickname(c, mids)
		return
	})
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
	// 获取最初粉丝数
	eg.Go(func(ctx context.Context) (err error) {
		memberInitFans, err = s.getMemberInitFans(c)
		return
	})
	// 获取最终粉丝数
	eg.Go(func(ctx context.Context) (err error) {
		memberEndFans, err = s.memberFollowerNum(c, mids)
		return
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	return s.operateMemberInfoToCsv(c, mids, memberNickname, rankLastBatch, memberRankTimes, memberRankHighest, memberInitFans, memberEndFans)
}

func (s *Service) getRankListByBatch(c context.Context, sid, endBatch int64) (map[int64]*mdlRank.DB, error) {
	mapLastRank := make(map[int64]*mdlRank.DB)
	rankLastBatch, err := s.rank.GetRankListByBatch(c, sid, endBatch)
	if err != nil {
		log.Error("s.rank.GetRankListByBatch: error(%v)", err)
		err = errors.Wrapf(err, "s.rank.GetRankListByBatch")
		return nil, err
	}
	for _, v := range rankLastBatch {
		mapLastRank[v.Mid] = v
	}
	return mapLastRank, nil
}

func (s *Service) operateMemberInfoToCsv(c context.Context, mids []int64, nickNameBatch map[int64]string, rankLastBatch map[int64]*mdlRank.DB, memberRankTimes map[int64]*mdlRank.MemberRankTimes, memberRankHighest map[int64]*mdlRank.MemberRankHighest, memberInitFans map[int64]int64, memberEndFans map[int64]int64) error {
	categoryHeader := []string{"用户ID", "昵称", "最终名次", "最终积分", "最好名次", "上榜次数", "最初粉丝数", "最终粉丝数"}
	data := [][]string{}
	for _, v := range mids {
		rows := []string{}
		var midStr, nickName, lastRank, score, highest, times, fansInit, fans string
		midStr = strconv.FormatInt(v, 10)
		rankLastStruct, ok := rankLastBatch[v]
		if ok {
			lastRank = strconv.Itoa(rankLastStruct.Rank)
			score = strconv.FormatInt(rankLastStruct.Score, 10)
		}
		memberNickName, ok := nickNameBatch[v]
		if ok {
			nickName = memberNickName
		}
		rankTimes, ok := memberRankTimes[v]
		if ok {
			times = strconv.Itoa(rankTimes.Times)
		}
		rankHighest, ok := memberRankHighest[v]
		if ok {
			highest = strconv.Itoa(rankHighest.Rank)
		}
		rankInitFans, ok := memberInitFans[v]
		if ok {
			fansInit = strconv.FormatInt(rankInitFans, 10)
		}
		rankEndFans, ok := memberEndFans[v]
		if ok {
			fans = strconv.FormatInt(rankEndFans, 10)
		}
		rows = append(rows, midStr, nickName, lastRank, score, highest, times, fansInit, fans)
		data = append(data, rows)
	}
	fileName := fmt.Sprintf("%v_%v.csv", handWorkMidRankResultFileName, time.Now().Format("200601021504"))
	err := s.createCsvAndSend(c, s.c.HandWrite.FilePath, fileName, handWorkMidRankResultFileName, categoryHeader, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) getMemberInitFans(c context.Context) (map[int64]int64, error) {
	memberInitFans := make(map[int64]int64)
	memberInitFansStr, err := s.handWrite.GetMidInitFans(c)
	if err != nil {
		err = errors.Wrapf(err, "s.handWrite.GetMidInitFans")
		log.Error("s.handWrite.GetMidInitFans: error(%v) ", err)
	}
	for k, v := range memberInitFansStr {
		mid, _ := strconv.ParseInt(k, 10, 64)
		memberInitFans[mid] = v
	}
	return memberInitFans, nil
}

// countMidRankTimes 统计用户上榜次数
func (s *Service) countMidRankTimes(c context.Context, sid, startBatch, endBatch int64, mids []int64) (map[int64]*mdlRank.MemberRankTimes, error) {
	ch := make(chan []*mdlRank.MemberRankTimes, maxRankDbChannelLength)
	var mapMidTimes map[int64]*mdlRank.MemberRankTimes
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		return s.countMidTimesIntoChannel(c, sid, startBatch, endBatch, mids, ch)
	})
	eg.Go(func(ctx context.Context) (err error) {
		mapMidTimes = s.countMidTimesOutChannel(c, ch)
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, err
	}
	return mapMidTimes, nil

}

func (s *Service) countMidTimesIntoChannel(c context.Context, sid, startBatch, endBatch int64, mids []int64, ch chan []*mdlRank.MemberRankTimes) error {
	var times int
	patch := maxRankDbMidLength
	concurrency := concurrencyMaxRankDbMid
	times = len(mids) / patch / concurrency
	defer close(ch)
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			i := index
			b := batch
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
					reply, err := s.rank.GetMemberRankTimes(c, sid, startBatch, endBatch, reqMids)
					if err != nil || reply == nil {
						err = errors.Wrapf(err, "s.handWrite.GetMidsAward")
						log.Error("s.handWrite.GetMidsAward: error(%v) batch(%d)", err, i)
						return err
					}
					ch <- reply
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Error("eg.Wait error(%v)", err)
			return err
		}

	}

	return nil

}

func (s *Service) countMidTimesOutChannel(c context.Context, channel chan []*mdlRank.MemberRankTimes) map[int64]*mdlRank.MemberRankTimes {
	midTimes := make(map[int64]*mdlRank.MemberRankTimes)
	for ch := range channel {
		for _, v := range ch {
			midTimes[v.Mid] = v
		}
	}
	return midTimes
}

// countMidRankHighest 统计用户最好成绩
func (s *Service) countMidRankHighest(c context.Context, sid, startBatch, endBatch int64, mids []int64) (map[int64]*mdlRank.MemberRankHighest, error) {
	ch := make(chan []*mdlRank.MemberRankHighest, maxRankDbChannelLength)
	var mapMidHighest map[int64]*mdlRank.MemberRankHighest
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		return s.countMidHighestIntoChannel(c, sid, startBatch, endBatch, mids, ch)
	})
	eg.Go(func(ctx context.Context) (err error) {
		mapMidHighest = s.countMidHighestOutChannel(c, ch)
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, err
	}
	return mapMidHighest, nil

}

func (s *Service) countMidHighestIntoChannel(c context.Context, sid, startBatch, endBatch int64, mids []int64, ch chan []*mdlRank.MemberRankHighest) error {
	var times int
	patch := maxRankDbMidLength
	concurrency := concurrencyMaxRankDbMid
	times = len(mids) / patch / concurrency
	defer close(ch)
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			i := index
			b := batch
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if i*maxRankDbMidLength >= len(mids) {
					return nil
				}
				if start >= len(mids) {
					return nil
				}
				reqMids := mids[start:]
				end := start + patch
				if end < len(mids) {
					reqMids = mids[start:end]
				}
				if len(reqMids) > 0 {
					reply, err := s.rank.GetMemberHighest(c, sid, startBatch, endBatch, reqMids)
					if err != nil || reply == nil {
						err = errors.Wrapf(err, "s.handWrite.GetMemberHighest")
						log.Error("s.handWrite.GetMemberHighest: error(%v) batch(%d)", err, i)
						return err
					}
					ch <- reply
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Error("eg.Wait error(%v)", err)
			return err
		}
	}

	return nil
}

func (s *Service) countMidHighestOutChannel(c context.Context, channel chan []*mdlRank.MemberRankHighest) map[int64]*mdlRank.MemberRankHighest {
	midHighest := make(map[int64]*mdlRank.MemberRankHighest)
	for ch := range channel {
		for _, v := range ch {
			midHighest[v.Mid] = v
		}
	}
	return midHighest
}

// moneyResult 统计用户获得金额
func (s *Service) moneyResult(c context.Context, mids []int64, awardCount *mdlhw.AwardCount) error {
	ch := make(chan map[int64]*mdlhw.MidAward, scoreResultChannelLength)
	var midsAward []*mdlhw.MidAward
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		return s.moneyResultIntoChannel(c, mids, ch)
	})
	eg.Go(func(ctx context.Context) (err error) {
		midsAward = s.moneyResultOutChannel(c, ch)
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	return s.moneyResultCsv(c, midsAward, awardCount)
}

func (s *Service) moneyResultCsv(c context.Context, midAward []*mdlhw.MidAward, awardCount *mdlhw.AwardCount) error {
	categoryHeader := []string{"用户ID", "神仙模式", "爆肝模式", "新人福利", "最终金额:单位（分）"}
	data := [][]string{}
	for _, v := range midAward {
		money := 0
		if v.God > 0 && awardCount.God > 0 {
			money += (s.c.HandWrite.GodAllMoney / awardCount.God) * v.God
		}
		if v.Tired == awardCanGet && awardCount.Tired > 0 {
			money += s.c.HandWrite.TiredAllMoney / awardCount.Tired
		}
		if v.New == awardCanGet && awardCount.New > 0 {
			money += s.c.HandWrite.NewAllMoney / awardCount.New
		}

		if money > 0 {
			rows := []string{}
			rows = append(rows, fmt.Sprintf("%d", v.Mid), fmt.Sprintf("%d", v.God), fmt.Sprintf("%d", v.Tired), fmt.Sprintf("%d", v.New), fmt.Sprintf("%d", money))
			data = append(data, rows)
		}
	}
	fileName := fmt.Sprintf("%v_%v.csv", handWorkMoneyResultFileName, time.Now().Format("200601021504"))

	err := s.createCsvAndSend(c, s.c.HandWrite.FilePath, fileName, handWorkMoneyResultFileName, categoryHeader, data)
	if err != nil {
		return err
	}
	return nil
}

// createCsvAndSend 创建csv文件并发送
func (s *Service) createCsvAndSend(c context.Context, filePath, fileName string, subject string, categoryHeader []string, data [][]string) error {
	err := os.MkdirAll(filePath, 0755)
	if err != nil {
		log.Error("os.MkdirAll error(%v)", err)
		return err
	}
	f, err := os.Create(filePath + fileName)
	if err != nil {
		err = errors.Wrapf(err, "s.createCsv")
		log.Error("s.createCsv: error(%v) fileName %v", err, fileName)
	}
	defer os.RemoveAll(filePath)
	defer f.Close()
	f.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(f)
	w.Write(categoryHeader)
	w.WriteAll(data) //写入数据
	err = s.mailFile(c, subject, mdlmail.TypeTextHTML, &mdlmail.Attach{Name: fileName, File: filePath + fileName})
	if err != nil {
		log.Error("s.mailFile: error(%v) fileName %v", err, fileName)

	}
	return nil
}

// mailFile 邮件发送
func (s *Service) mailFile(c context.Context, subject string, mailType mdlmail.Type, attach *mdlmail.Attach) error {
	base := &mdlmail.Base{
		Host:    s.c.Handwrite2021.Mail.Host,
		Port:    s.c.Handwrite2021.Mail.Port,
		Address: s.c.Handwrite2021.Mail.Address,
		Pwd:     s.c.Handwrite2021.Mail.Pwd,
		Name:    s.c.Handwrite2021.Mail.Name,
	}
	toAddress := []*mdlmail.Address{}
	for _, v := range s.c.Handwrite2021.Mail.ToAddress {
		toAddress = append(toAddress, &mdlmail.Address{
			Address: v.Address,
			Name:    v.Name,
		})
	}
	ccAddress := []*mdlmail.Address{}
	for _, v := range s.c.Handwrite2021.Mail.CcAddress {
		ccAddress = append(ccAddress, &mdlmail.Address{
			Address: v.Address,
			Name:    v.Name,
		})
	}
	bccAddress := []*mdlmail.Address{}
	for _, v := range s.c.Handwrite2021.Mail.BccAddresses {
		bccAddress = append(bccAddress, &mdlmail.Address{
			Address: v.Address,
			Name:    v.Name,
		})
	}
	mail := &mdlmail.Mail{
		ToAddresses:  toAddress,
		CcAddresses:  ccAddress,
		BccAddresses: bccAddress,
		Subject:      subject,
		Type:         mailType,
	}
	return s.SendMail(c, mail, base, attach)
}

// moneyResultIntoChannel 金额统计结果进入channel
func (s *Service) moneyResultIntoChannel(c context.Context, mids []int64, ch chan map[int64]*mdlhw.MidAward) error {
	var times int
	patch := maxScoreResultLength
	concurrency := concurrencyMaxScoreResult
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
					reply, err := s.handWrite.GetMidsAward(ctx, reqMids)
					if err != nil || reply == nil {
						log.Error("s.handWrite.GetMidsAward: error(%v) batch(%d)", err, i)
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

// moneyResultOutChannel channel中统计出金额
func (s *Service) moneyResultOutChannel(c context.Context, channel chan map[int64]*mdlhw.MidAward) []*mdlhw.MidAward {
	midAward := []*mdlhw.MidAward{}
	for ch := range channel {
		for _, v := range ch {
			midAward = append(midAward, v)
		}
	}
	return midAward
}
