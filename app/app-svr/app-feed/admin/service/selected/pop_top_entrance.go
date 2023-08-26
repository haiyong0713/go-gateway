package selected

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
)

func (s *Service) updateEntranceRank(ctx context.Context) (err error) {
	var (
		allEntrance     []*selected.PopTopEntrance
		originIndex     int
		newRankEntrance []*selected.PopTopEntrance
	)
	if allEntrance, err = s.dao.GetAllEntrances(ctx); err != nil {
		return
	}
	for i := len(allEntrance) - 1; i > 0; i-- { // 获取当前倒序的序号，用于回滚排序操作
		if s.c.WeeklySelected.RankId == allEntrance[i].ID {
			originIndex = len(allEntrance) - 1 - i
			break
		}
	}
	if originIndex == 0 {
		log.Error("updateEntranceRank not get index id:%d", s.c.WeeklySelected.RankId)
		return
	}
	for i := 0; i < len(allEntrance); i++ {
		if i == s.c.WeeklySelected.RankIndex {
			if s.c.WeeklySelected.RankId == allEntrance[i].ID { // 如果位置已经正确，直接返回
				return
			}
			newRankEntrance = append(newRankEntrance, &selected.PopTopEntrance{
				ID:   s.c.WeeklySelected.RankId,
				Rank: allEntrance[i].Rank,
			})
			newRankEntrance = append(newRankEntrance, &selected.PopTopEntrance{
				ID:   allEntrance[i].ID,
				Rank: allEntrance[i].Rank + 1,
			})
		}
		if i > s.c.WeeklySelected.RankIndex && s.c.WeeklySelected.RankId != allEntrance[i].ID {
			newRankEntrance = append(newRankEntrance, &selected.PopTopEntrance{
				ID:   allEntrance[i].ID,
				Rank: allEntrance[i].Rank + 1,
			})
		}
	}
	if len(newRankEntrance) == 0 {
		log.Error("updateEntranceRank not get need update list id:%d", s.c.WeeklySelected.RankId)
		return
	}
	if err = s.dao.AddRankCache(ctx, s.c.WeeklySelected.RankId, originIndex); err != nil {
		log.Error("updateEntranceRank AddRankCache err (%+v) originIndex (%d)", err, originIndex)
		return
	}
	if err = s.dao.UpdateEntrancesRank(ctx, newRankEntrance); err != nil {
		log.Error("updateEntranceRank UpdateEntrancesRank err (%+v) list (%+v)", err, newRankEntrance)
		return
	}
	log.Info("updateEntranceRank success originIndex (%d) updatedIndex (%d)", originIndex, s.c.WeeklySelected.RankIndex)
	return
}

func (s *Service) rollbackEntranceRank() (err error) {
	ctx := context.Background()
	var (
		allEntrance     []*selected.PopTopEntrance
		originIndex     int
		newRankEntrance []*selected.PopTopEntrance
	)
	if allEntrance, err = s.dao.GetAllEntrances(ctx); err != nil {
		return
	}
	if originIndex, err = s.dao.CacheRankCache(ctx, s.c.WeeklySelected.RankId); err != nil {
		log.Error("rollbackEntranceRank CacheRankCache err (%+v) RankId (%d)", err, s.c.WeeklySelected.RankId)
		return
	}
	originIndex = len(allEntrance) - originIndex
	for i := len(allEntrance) - 1; i > 0; i-- {
		if i > originIndex {
			newRankEntrance = append(newRankEntrance, &selected.PopTopEntrance{
				ID:   allEntrance[i].ID,
				Rank: allEntrance[i].Rank + 1,
			})
		}
		if i == originIndex {
			if s.c.WeeklySelected.RankId == allEntrance[i].ID { // 位置已经正确，不用修改
				return
			}
			newRankEntrance = append(newRankEntrance, &selected.PopTopEntrance{
				ID:   s.c.WeeklySelected.RankId,
				Rank: allEntrance[i].Rank,
			})
			newRankEntrance = append(newRankEntrance, &selected.PopTopEntrance{
				ID:   allEntrance[i].ID,
				Rank: allEntrance[i].Rank + 1,
			})
		}
	}
	if len(newRankEntrance) == 0 {
		log.Error("updateEntranceRank not get need update list id:%d", s.c.WeeklySelected.RankId)
		return
	}
	if err = s.dao.UpdateEntrancesRank(ctx, newRankEntrance); err != nil {
		log.Error("updateEntranceRank UpdateEntrancesRank err (%+v) list (%+v)", err, newRankEntrance)
		return
	}
	log.Info("updateEntranceRank success originIndex (%d) updatedIndex (%d)", originIndex, s.c.WeeklySelected.RankIndex)
	return
}
