package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	mdlhw "go-gateway/app/web-svr/activity/job/model/handwrite"
	mdlRank "go-gateway/app/web-svr/activity/job/model/rank"
	favoriteapi "go-main/app/community/favorite/service/api"

	relationapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"github.com/pkg/errors"
)

const (
	// maxArcsLength 一次稿件服务获取稿件的数量
	maxArcsLength = 50
	// concurrencyArchiveDb 稿件服务并发量
	concurrencyArchiveDb = 2
	// maxFollowerLength 一次关系服务获取粉丝的数量
	maxFollowerLength = 50
	// concurrencyFollower 粉丝数量并发量
	concurrencyFollwer = 4
	// maxMemberInfoLength 一次获取用户信息的数量
	maxMemberInfoLength = 50
	// concurrencyMemberInfo 获取用户信息并发量
	concurrencyMemberInfo = 4
	// midAwardRecordbatch 一次记录用户获奖情况的数量
	midAwardRecordbatch = 1000
	// concurrencyMidAwardRecord 并发记录用户获奖情况
	concurrencyMidAwardRecord = 1
	// awardCanGet 可以获奖
	awardCanGet = 1
	// awardCannotGet 不可以获奖
	awardCannotGet = 0
	// fansChannelLength 用户粉丝channel长度
	fansChannelLength = 50
	// memberChannelLength 用户信息channel长度
	memberChannelLength = 50
	// favVideoType 收藏夹视频类型
	favVideoType = 2
	// favNeedFilter 收藏夹需要过滤
	favNeedFilter = 1
	// favPnStart 收藏夹第一页
	favPnStart = 1
	// favPnSize 一页请求数量
	favPnSize = 20
	// favChannelLength 收藏信息channel长度
	favChannelLength = 50
)

// HandWriteMemberScore 手书活动用户分数及获奖状态统计
func (s *Service) HandWriteMemberScore() {
	now := time.Now().Unix()
	if now < s.c.HandWrite.ActivityStart || now > s.c.HandWrite.ActivityEnd {
		return
	}
	c := context.Background()
	s.handWriteRankRunning.Lock()
	defer s.handWriteRankRunning.Unlock()
	archiveStatBatch, err := s.midArchiveInfo(c, s.c.HandWrite.Sid)
	if err != nil {
		log.Warn("s.HandWriteMemberScore get archive state error(%v)", err)
		return
	}
	mids, err := s.getAllMid(c, s.c.HandWrite.Sid)
	if err != nil {
		return
	}
	if archiveStatBatch != nil && len(*archiveStatBatch) > 0 {
		err = s.midAwardByBatch(c, archiveStatBatch, mids, s.c.HandWrite.Sid)
		if err != nil {
			log.Warn("s.midAwardByBatch set mid award error(%v)", err)
			return
		}
	}
	log.Info("HandWriteMemberScore success()")

}

// getAllMid 获取所有的mid
func (s *Service) getAllMid(c context.Context, sid int64) ([]int64, error) {
	mids := make([]int64, 0)

	midList, err := s.dao.AllDistinctMid(c, s.c.HandWrite.Sid)
	if err != nil {
		log.Warn("s.dao.AllDistinctMid get mid error(%v)", err)
		return mids, err
	}
	for _, v := range midList {
		mids = append(mids, v.Mid)
	}
	return mids, nil
}

// countMidArchive count mid archive
func (s *Service) countMidArchive(c context.Context, midArchive []*mdlRank.ArchiveStat) int {
	if len(midArchive) >= s.c.HandWrite.CountMidArchive {
		return awardCanGet
	}
	return awardCannotGet
}

// countMidArchiveCoin count mid archive coin
func (s *Service) countMidArchiveCoin(c context.Context, midArchive []*mdlRank.ArchiveStat) int {
	var archiveCount int
	for _, v := range midArchive {
		if v.Coin >= s.c.HandWrite.ArchiveCoinAward {
			archiveCount++
		}
	}
	return archiveCount
}

func (s *Service) midHistoryArchive(c context.Context, mids []int64) (map[int64]*mdlhw.MidAward, error) {
	reply, err := s.handWrite.MidListDistinct(c, mids)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]*mdlhw.MidAward, 0)
	for i := range mids {
		res[mids[i]] = &mdlhw.MidAward{New: 1, Mid: mids[i]}
	}
	for i := range reply {
		if reply[i] != nil {
			if _, ok := res[reply[i].Mid]; ok {
				res[reply[i].Mid].New = 0
			}
		}
	}
	return res, nil
}

func (s *Service) setHistoryRank(c context.Context, midScoreMap *mdlRank.MidScoreMap) error {
	historyRank, err := s.rank.GetRank(c, mdlhw.HandWriteKey)
	if err != nil {
		err = errors.Wrapf(err, "s.rank.GetRank")
		return err
	}
	if historyRank != nil {
		for _, v := range historyRank {
			if _, ok := (*midScoreMap)[v.Mid]; ok {
				(*midScoreMap)[v.Mid].History = v.Rank
			}
		}
	}
	return nil
}

// score 积分计算公式
func score(arc *mdlRank.ArchiveStat) int64 {
	if arc.View == 0 {
		return 0
	}
	return getPlayScore(arc) + getQualityScore(arc) + getTopicScore(arc)

}

// getPlayScore 获取播放分数
func getPlayScore(arc *mdlRank.ArchiveStat) int64 {
	videos := float64(arc.Videos)
	views := float64(arc.View)
	pRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", (4/(videos+3))), 64)
	aRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", ((300000+views)/(2*views))), 64)
	if aRevise > 1 {
		aRevise = 1
	}
	return int64(math.Floor(views*pRevise*aRevise + 0.5))
}

// getQualityScore 获取质量分
func getQualityScore(arc *mdlRank.ArchiveStat) int64 {
	like := float64(arc.Like)
	coin := float64(arc.Coin)
	fav := float64(arc.Fav)
	views := float64(arc.View)
	quality := like*5 + coin*10 + fav*20
	bRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", ((like*5+coin*10+fav*20)/(views+like*5+coin*10+fav*20))), 64)
	return int64(math.Floor(quality*bRevise + 0.5))
}

// getTopicScore 获取讨论分
func getTopicScore(arc *mdlRank.ArchiveStat) int64 {
	return int64((arc.Danmaku + arc.Reply)) * 20
}

// midAwardByBatch mid award count by batch
func (s *Service) midAwardByBatch(c context.Context, archiveBatch *mdlRank.ArchiveStatMap, mids []int64, sid int64) error {

	archiveBatchInfo := *archiveBatch
	midScoreMap := archiveBatchInfo.Score(score)
	err := s.setHistoryRank(c, midScoreMap)
	if err != nil {
		return err
	}
	eg := errgroup.WithContext(c)
	midScoreBatch := s.rankResult(c, midScoreMap)
	eg.Go(func(ctx context.Context) (err error) {
		// 排名计算
		return s.rankResultSave(c, midScoreBatch, sid)
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 计算获奖情况
		midRank := s.midRank(c, midScoreBatch)
		return s.awardResult(c, mids, archiveBatchInfo, midScoreMap, midRank)
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 计算初始粉丝数
		mids := make([]int64, 0)
		for mid := range archiveBatchInfo {
			mids = append(mids, mid)
		}
		return s.initFans(c, mids)
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	return nil

}

func (s *Service) rankResult(c context.Context, midScoreMap *mdlRank.MidScoreMap) *mdlRank.MidScoreBatch {
	var midScoreBatch = mdlRank.MidScoreBatch{}
	for _, v := range *midScoreMap {
		midScoreBatch.Data = append(midScoreBatch.Data, v)
	}
	midScoreBatch.TopLength = s.c.HandWrite.RankTopLength
	mdlRank.Sort(&midScoreBatch)
	return &midScoreBatch
}

// rankResultSave 排名结果保存
func (s *Service) rankResultSave(c context.Context, midScoreBatch *mdlRank.MidScoreBatch, sid int64) error {
	eg := errgroup.WithContext(c)
	// redis 存储
	eg.Go(func(ctx context.Context) (err error) {
		return s.redisRank(c, midScoreBatch)
	})
	// mysql 存储
	eg.Go(func(ctx context.Context) error {
		return s.dbRank(c, midScoreBatch, sid)
	})

	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	return nil

}

// redisRank redis rank data
func (s *Service) redisRank(c context.Context, midScoreBatch *mdlRank.MidScoreBatch) (err error) {
	rankRedis := make([]*mdlRank.Redis, 0)
	for i, v := range midScoreBatch.Data {
		if v != nil {
			rankRedis = append(rankRedis, &mdlRank.Redis{
				Mid:   v.Mid,
				Score: v.Score,
				Rank:  i + 1,
			})
		}
	}
	if len(rankRedis) > 0 {
		err = s.rank.SetRank(c, mdlhw.HandWriteKey, rankRedis)
		if err != nil {
			log.Error("s.SetRank: error(%v)", err)
			err = errors.Wrapf(err, "s.SetRank")
		}
	}
	return err
}

// dbRank db rank data
func (s *Service) dbRank(c context.Context, midScoreBatch *mdlRank.MidScoreBatch, sid int64) (err error) {
	rankDb := make([]*mdlRank.DB, 0)
	mids := make([]int64, 0)
	for _, v := range midScoreBatch.Data {
		mids = append(mids, v.Mid)
	}
	var (
		memberFansMap  map[int64]int64
		memberNickanme map[int64]string
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if memberFansMap, err = s.memberFollowerNum(c, mids); err != nil {
			log.Error("s.memberFollowerNum(%v) error(%v)", mids, err)
			return ecode.ActivityWriteHandFansErr
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if memberNickanme, err = s.memberNickname(c, mids); err != nil {
			log.Error("s.memberNickname(%v) error(%v)", mids, err)
			return ecode.ActivityWriteHandMemberErr
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	hourString := time.Now().Format("2006010215")
	hour, _ := strconv.ParseInt(hourString, 10, 64)
	for i, v := range midScoreBatch.Data {
		if v != nil {
			rank := &mdlRank.DB{
				Mid:   v.Mid,
				Score: v.Score,
				Rank:  i + 1,
				Batch: hour,
				SID:   sid,
			}
			nickName, nickNameOk := memberNickanme[v.Mid]
			if nickNameOk {
				rank.NickName = nickName
			}
			fans, fansOk := memberFansMap[v.Mid]
			if fansOk {
				rank.RemarkOrigin = mdlhw.Remark{
					Follower: fans,
				}
			}
			rankDb = append(rankDb, rank)
		}
	}
	if len(rankDb) > 0 {
		err = s.rank.BatchAddRank(c, rankDb)
		if err != nil {
			log.Error("s.rank.BatchAddRank(%v) error(%v)", mids, err)
			err = errors.Wrapf(err, "s.rank.BatchAddRank %v", mids)
		}
	}
	return err
}

// memberFollowerNum 获取用户粉丝数
func (s *Service) memberFollowerNum(c context.Context, mids []int64) (map[int64]int64, error) {
	eg := errgroup.WithContext(c)
	channel := make(chan map[int64]*relationapi.StatReply, fansChannelLength)
	var midsStates map[int64]int64
	eg.Go(func(ctx context.Context) error {
		return s.memberFansIntoChannel(c, mids, channel)
	})
	eg.Go(func(ctx context.Context) (err error) {
		midsStates = s.memberFansOutChannel(c, channel)
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, err
	}
	return midsStates, nil
}

func (s *Service) memberFansOutChannel(c context.Context, channel chan map[int64]*relationapi.StatReply) map[int64]int64 {
	midsStates := make(map[int64]int64)
	for item := range channel {
		for mid, value := range item {
			if value != nil {
				midsStates[mid] = value.Follower
			}
		}
	}
	return midsStates

}

func (s *Service) memberFansIntoChannel(c context.Context, mids []int64, channel chan map[int64]*relationapi.StatReply) error {
	var times int
	patch := maxFollowerLength
	concurrency := concurrencyFollwer
	times = len(mids) / patch / concurrency
	defer close(channel)
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
					statsReply, err := s.relationClient.Stats(ctx, &relationapi.MidsReq{Mids: reqMids})
					if err != nil || statsReply == nil {
						log.Error("s.relationClient.Stats: error(%v) batch(%d)", err, i)
						return err
					}
					channel <- statsReply.StatReplyMap
				}
				return nil
			})
			if err := eg.Wait(); err != nil {
				log.Error("eg.Wait error(%v)", err)
				return err
			}
		}

	}

	return nil
}

// memberInfo 用户信息
func (s *Service) memberNickname(c context.Context, mids []int64) (map[int64]string, error) {
	eg := errgroup.WithContext(c)
	midsNickname := make(map[int64]string)
	channel := make(chan map[int64]*accountapi.Info, memberChannelLength)
	eg.Go(func(ctx context.Context) error {
		return s.memberNicknameIntoChannel(c, mids, channel)
	})
	eg.Go(func(ctx context.Context) (err error) {
		midsNickname, err = s.memberNicknameOutChannel(c, channel)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.memberNickname")
		return nil, err
	}

	return midsNickname, nil
}

func (s *Service) memberNicknameIntoChannel(c context.Context, mids []int64, channel chan map[int64]*accountapi.Info) error {
	var times int
	patch := maxMemberInfoLength
	concurrency := concurrencyMemberInfo
	times = len(mids) / patch / concurrency
	defer close(channel)
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
			return ecode.ActivityWriteHandMemberErr
		}
	}
	return nil
}

func (s *Service) memberNicknameOutChannel(c context.Context, channel chan map[int64]*accountapi.Info) (map[int64]string, error) {
	midsNickname := make(map[int64]string)
	for item := range channel {
		for mid, value := range item {
			if value != nil {
				midsNickname[mid] = value.Name
			}
		}
	}
	return midsNickname, nil
}

// initFans 初始化粉丝设置
func (s *Service) initFans(c context.Context, mids []int64) (err error) {
	historyMid, err := s.handWrite.GetActivityMember(c)
	if err != nil {
		err = errors.Wrapf(err, "s.handWrite.GetActivityMember %v", mids)
		return err
	}
	midMap := make(map[int64]bool)
	addMid := make([]int64, 0)

	if len(historyMid) > 0 {
		for _, v := range historyMid {
			midMap[v] = true
		}
	}
	for _, v := range mids {
		if _, ok := midMap[v]; !ok {
			addMid = append(addMid, v)
		}
	}
	if len(addMid) > 0 {
		addMidsFans, err := s.memberFollowerNum(c, addMid)
		if err != nil {
			err = errors.Wrapf(err, "s.memberFollowerNum %v", mids)
			return ecode.ActivityWriteHandFansErr
		}
		err = s.handWrite.SetMidInitFans(c, addMidsFans)
		if err != nil {
			err = errors.Wrapf(err, "s.handWrite.SetMidInitFans %v", mids)
			return ecode.ActivityWriteHandFansErr
		}

	}
	return nil
}

func (s *Service) midRank(c context.Context, midScoreBatch *mdlRank.MidScoreBatch) map[int64]int {
	var midRankMap = make(map[int64]int)
	for i, v := range midScoreBatch.Data {
		midRankMap[v.Mid] = i + 1
	}
	return midRankMap
}

// awardResult 获奖结果
func (s *Service) awardResult(c context.Context, mids []int64, archiveBatchInfo mdlRank.ArchiveStatMap, midScoreMap *mdlRank.MidScoreMap, midRank map[int64]int) error {
	var times int
	patch := midAwardRecordbatch
	concurrency := concurrencyMidAwardRecord
	times = len(mids) / patch / concurrency

	var tiredCount, godCount, newCount int
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
					midAwardMap, err := s.midHistoryArchive(c, reqMids)
					if err != nil {
						log.Error("s.midHistoryArchive: error(%v)", err)
						return err
					}
					for i := range reqMids {
						countMidArchive := 0
						countMidArchiveCoin := 0

						aidInfo, ok := archiveBatchInfo[reqMids[i]]
						if ok {
							countMidArchive = s.countMidArchive(c, aidInfo)
							countMidArchiveCoin = s.countMidArchiveCoin(c, aidInfo)
						} else {
							midAwardMap[reqMids[i]].New = 0
						}
						tiredCount += countMidArchive
						godCount += countMidArchiveCoin
						if _, ok := midAwardMap[reqMids[i]]; ok {
							midAwardMap[reqMids[i]].God = countMidArchiveCoin
							midAwardMap[reqMids[i]].Tired = countMidArchive
							newCount += midAwardMap[reqMids[i]].New

						}
						if midScore, scoreOk := (*midScoreMap)[reqMids[i]]; scoreOk {
							midAwardMap[reqMids[i]].Score = midScore.Score
						}
						if rank, ok := midRank[reqMids[i]]; ok {
							midAwardMap[reqMids[i]].Rank = rank
						}
					}
					err = s.handWrite.AddMidAward(c, midAwardMap)
					if err != nil {
						err = errors.Wrapf(err, "s.handWrite.AddMidAward")
						return err
					}
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Error("eg.Wait error(%v)", err)
			return err
		}
	}

	return s.awardAllCount(c, newCount, godCount, tiredCount)
}

func (s *Service) awardAllCount(c context.Context, newCount, godCount, tiredCount int) error {
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		awardCound := &mdlhw.AwardCount{
			God:   godCount,
			New:   newCount,
			Tired: tiredCount,
		}
		return s.handWrite.SetAwardCount(c, awardCound)
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	return nil

}

// FavSyncCounterFilter 收藏夹数据同步计数过滤
func (s *Service) FavSyncCounterFilter() {
	c := context.Background()
	s.handWriteFavRunning.Lock()
	defer s.handWriteFavRunning.Unlock()
	ch := make(chan []*favoriteapi.ModelFavorite, favChannelLength)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		return s.favoritesAllIntoChannel(c, s.c.HandWrite.FavMid, s.c.HandWrite.FavMid, ch)
	})
	eg.Go(func(ctx context.Context) (err error) {
		return s.favoritesAllOutChannel(c, ch)
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	log.Info("FavSyncCounterFilter success()")
	return
}

// favoritesAll 获取收藏夹中的数据
func (s *Service) favoritesAllIntoChannel(c context.Context, mid int64, Fid int64, ch chan []*favoriteapi.ModelFavorite) error {
	batch := favPnStart
	var (
		err error
		fav *favoriteapi.FavoritesReply
	)
	defer close(ch)
	for {
		fav, err = s.favoriteClient.FavoritesAll(c, &favoriteapi.FavoritesReq{
			Tp:  favVideoType,
			Mid: s.c.HandWrite.FavMid,
			Uid: s.c.HandWrite.FavMid,
			Fid: s.c.HandWrite.FavID,
			Pn:  int32(batch),
			Ps:  favPnSize,
			// Tv:  favNeedFilter,
		})
		if err != nil {
			log.Error("s.favoriteClient.FavoritesAll: error(%v)", err)
			break
		}
		if fav.Res == nil {
			break
		}
		if len(fav.Res.List) > 0 {
			ch <- fav.Res.List
		}
		if len(fav.Res.List) < favPnSize {
			break
		}
		time.Sleep(100 * time.Microsecond)
		batch++
	}
	return err

}

// 收藏夹从channel中取出
func (s *Service) favoritesAllOutChannel(c context.Context, ch chan []*favoriteapi.ModelFavorite) (err error) {
	for v := range ch {
		aids := []int64{}
		for _, item := range v {
			if item.State == 0 {
				aids = append(aids, item.Oid)
			}
		}
		err = s.syncAidsToActPlat(c, aids)
		if err != nil {
			log.Error("s.syncAidsToActPlat: error(%v)", err)
		}
	}
	return err
}

func (s *Service) syncAidsToActPlat(c context.Context, aids []int64) error {
	values := []*actplatapi.FilterMemberInt{}
	expireTime := int64(1200)
	for _, i := range aids {
		values = append(values, &actplatapi.FilterMemberInt{Value: i, ExpireTime: expireTime})
	}
	_, err := s.actplatClient.AddFilterMemberInt(c, &actplatapi.SetFilterMemberIntReq{
		Activity: s.c.HandWrite.ActPlatActivity,
		Counter:  s.c.HandWrite.ActPlatCounter,
		Filter:   s.c.HandWrite.ActPlatFilter,
		Values:   values,
	})
	return err
}
