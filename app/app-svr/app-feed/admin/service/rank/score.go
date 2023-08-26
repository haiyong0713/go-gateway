package rank

import (
	"context"
	"sort"
	"time"

	"go-common/library/log"

	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"

	rankModel "go-gateway/app/app-svr/app-feed/admin/model/rank"
)

// 获取榜单下视频列表，干预后分数和排名
//
//nolint:gocognit
func (s Service) GetRankAVShowList(ctx context.Context, rankId int) (rankAVShowList, rankAVShowListDedup []rankModel.RankDetailAVItem, err error) {
	var (
		count           int
		rankScoreList   []*rankModel.RankScore
		interventionMap map[int64]*rankModel.RankArchiveIntervention
		rankSortList    RankSortList
		rankConfig      *rankModel.RankConfigRes
		logDate         string
	)

	if originalMaxLogDate, _, dateError := s.dao.GetOriginalRankScoreListTime(rankId); dateError != nil {
		log.Error("s.dao.GetOriginalRankScoreListTime rankId(%v) error(%v)", rankId, dateError)
		logDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	} else {
		logDate = originalMaxLogDate
	}

	if rankScoreList, count, err = s.dao.GetOriginalRankScoreList(rankId, logDate); err != nil {
		log.Error("s.dao.GetOriginalRankScoreList rankId(%v) logDate(%v) error(%v)", rankId, logDate, err)
		return
	}

	if rankConfig, err = s.GetRankConfig(ctx, &rankModel.RankCommonQuery{
		Id: rankId,
	}); err != nil {
		log.Error("s.GetRankConfig rankId(%v) error(%v)", rankId, err)
		return
	}

	if count == 0 && len(rankConfig.AvManuallyList) == 0 {
		// 说明没有榜单
		return
	}

	// 拼接人工干预的稿件
	if len(rankConfig.AvManuallyList) > 0 {
		var archives *arcgrpc.ArcsReply

		if archives, err = s.arcClient.Arcs(ctx, &arcgrpc.ArcsRequest{Aids: rankConfig.AvManuallyList}); err != nil {
			log.Error("s.arcClient.Arcs error %v", err)
			err = nil
		}

		var needAddItems []*rankModel.RankScore
		rankScoreMap := map[int64]bool{}
		for _, rankScoreItem := range rankScoreList {
			rankScoreMap[rankScoreItem.Avid] = true
		}
		for _, avid := range rankConfig.AvManuallyList {
			var mid int64
			if info, ok := archives.Arcs[avid]; ok {
				mid = info.Author.Mid
			}
			// 人工添加的，如果不在已经算分的列表内，再添加，否则没必要
			if _, ok := rankScoreMap[avid]; !ok {
				needAddItems = append(needAddItems, &rankModel.RankScore{
					RankId: rankId,
					Avid:   avid,
					Mid:    mid,
				})
			}
		}
		rankScoreList = append(rankScoreList, needAddItems...)
	}

	if interventionMap, err = s.dao.GetRankInterventions([]int{rankId}); err != nil {
		log.Error("s.dao.GetRankInterventions error(%v)", err)
		return
	}

	blacklistMap := map[int64]bool{}
	for _, item := range rankConfig.Blacklist {
		blacklistMap[item.Uid] = true
	}

	avInstertedMap := map[int64]bool{}
	scoreRankIndex := 1
	for _, score := range rankScoreList {
		// 去重
		if _, ok := avInstertedMap[score.Avid]; ok {
			continue
		} else {
			avInstertedMap[score.Avid] = true
		}
		item := rankModel.RankDetailAVItem{
			User: &rankModel.UserItem{
				Uid: score.Mid,
			},
			Avid: score.Avid,
			Score: &rankModel.RankDetailScore{
				Rank:  scoreRankIndex,
				Play:  score.Play,
				Like:  score.Like,
				Coin:  score.Coin,
				Share: score.Share,
				Total: score.Play + score.Like + score.Coin + score.Share,
			},
		}

		scoreRankIndex += 1

		if intervention, ok := interventionMap[score.Avid]; ok {
			// 如果有干预，那么就把干预数据插入
			if intervention.Rank != 0 {
				item.ManualRank = intervention.Rank
			}
			item.IsHidden = intervention.IsHidden
			if intervention.Extra != nil {
				item.Score.ExtraScore = intervention.Extra
				item.Score.Total = item.Score.Total + intervention.Extra.Complete + intervention.Extra.Reduction + intervention.Extra.Creative
			}
		}

		rankSortList = append(rankSortList, item)
	}

	// 按照真实分数+干预分数，进行排序
	sort.Sort(&rankSortList)

	rankAVShowList = make([]rankModel.RankDetailAVItem, len(rankSortList))

	leftList := []rankModel.RankDetailAVItem{}
	for _, item := range rankSortList {
		// 根据指定的位置，进行重新排序
		if item.ManualRank > 0 && item.ManualRank <= len(rankAVShowList) {
			rankAVShowList[item.ManualRank-1] = item
			rankAVShowList[item.ManualRank-1].ShowRank = item.ManualRank
		} else {
			// 没有干预的稿件，先按顺序存起来
			leftList = append(leftList, item)
		}
	}

	leftIndex := 0
	showIndex := 1

	for i, item := range rankAVShowList {
		if leftIndex >= len(leftList) {
			break
		}

		//nolint:ineffassign
		ifShow := true

		// 说明这一项还没有被填充，需要从剩余稿件内填充
		if item.Avid == 0 {
			rankAVShowList[i] = leftList[leftIndex]
			leftIndex += 1
		}

		if rankAVShowList[i].IsHidden > 0 {
			rankAVShowList[i].HiddenReason = append(rankAVShowList[i].HiddenReason, "前端隐藏")
		}
		if blacklistMap[(rankAVShowList[i].User.Uid)] {
			rankAVShowList[i].HiddenReason = append(rankAVShowList[i].HiddenReason, "黑名单")
		}

		if len(rankAVShowList[i].HiddenReason) < 1 {
			ifShow = true
		} else {
			ifShow = false
		}

		if ifShow {
			rankAVShowList[i].ShowRank = showIndex
			showIndex += 1
		} else {
			// 说明黑名单内或者被隐藏了，不展示
			rankAVShowList[i].ShowRank = 0
		}
	}

	midMap := map[int64]bool{}
	rankAVShowListDedup = make([]rankModel.RankDetailAVItem, len(rankAVShowList))
	copy(rankAVShowListDedup, rankAVShowList)
	for i, item := range rankAVShowListDedup {
		// 隐藏同用户重复稿件
		if midMap[rankAVShowListDedup[i].User.Uid] {
			rankAVShowListDedup[i].HiddenReason = append(item.HiddenReason, "同用户重复稿件")
			rankAVShowListDedup[i].ShowRank = 0
		} else if item.ShowRank > 0 {
			midMap[rankAVShowListDedup[i].User.Uid] = true
		}
	}

	return
}
