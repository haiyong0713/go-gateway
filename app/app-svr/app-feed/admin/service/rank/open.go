package rank

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
	rankModel "go-gateway/app/app-svr/app-feed/admin/model/rank"
)

// 给网关用，返回所有的可见榜单详情
func (s *Service) OpenRankList(ctx context.Context, ps, pn int) (pager *rankModel.OpenListPager, err error) {
	var (
		configList []*rankModel.RankConfig
		pagerList  []*rankModel.OpenListItem
		count      int
	)
	if configList, count, err = s.dao.QueryRankConfigList(0, "", -1, 0, pn, ps); err != nil {
		log.Error("s.dao.QueryRankConfigList error(%v)", err)
		return
	}

	for _, config := range configList {
		var (
			tids       []int
			actIds     []int
			rankVideos []int64
		)

		for _, v := range string2Int64Array(config.Tids) {
			tids = append(tids, int(v))
		}

		for _, v := range string2Int64Array(config.ActIds) {
			actIds = append(actIds, int(v))
		}

		var scoreConfig []*rankModel.ScoreConfig
		//nolint:staticcheck
		err = json.Unmarshal([]byte(config.ScoreConfig), &scoreConfig)

		var description []*rankModel.Description
		//nolint:staticcheck
		err = json.Unmarshal([]byte(config.Description), &description)

		configDetail := &rankModel.RankConfigDetail{
			ID:           config.ID,
			Title:        config.Title,
			STime:        config.STime,
			ETime:        config.ETime,
			Cycle:        config.Cycle,
			PerUpdate:    config.PerUpdate,
			Tids:         tids,
			ActIds:       actIds,
			ArchiveSTime: config.ArchiveStime,
			ArchiveETime: config.ArchiveEtime,
			Cover:        config.Cover,
			Description:  description,
			HelpTips:     []string{"再接再厉呀~", "邀请好友去看视频，提高排名哦"},
		}

		var historyConfig *rankModel.RankHistoryDB
		// 判断当前时间是不是超过了今天的14:00，超过了，就用当前发榜的 HistoryId。没发榜，就用今天零点之前的最近的一个榜单
		currentHour := time.Now().Hour()
		//nolint:gomnd
		if currentHour >= 14 {
			if historyConfig, err = s.dao.FindRankHistoryConfig(config.HistoryId); err != nil {
				log.Error("s.dao.FindRankHistoryConfig historyId(%v) error(%v)", config.HistoryId, err)
				return
			}
		} else {
			if historyConfig, err = s.dao.FindLastRankHistoryConfig(); err != nil {
				log.Error("s.dao.FindLastRankHistoryConfig historyId(%v) error(%v)", config.HistoryId, err)
				return
			}
		}

		//nolint:gosimple
		for _, avid := range string2Int64Array(historyConfig.ScoreAvids) {
			rankVideos = append(rankVideos, avid)
		}

		var finalRank []*rankModel.FinalRankItem
		if historyConfig.FinalRankConfig != "" {
			var finalRankConfig []*rankModel.TMModel
			if err = json.Unmarshal([]byte(historyConfig.FinalRankConfig), &finalRankConfig); err != nil {
				log.Error("json.Unmarshal input(%v) error(%v)", historyConfig.FinalRankConfig, err)
				err = nil
			}
			for _, model := range finalRankConfig {
				//nolint:gomnd
				if model.Mode == 1 {
					avidList := string2Int64Array(historyConfig.ScoreAvids)
					end := model.Count
					if end > len(avidList) {
						end = len(avidList)
					}
					if end < 0 {
						end = 0
					}
					finalRank = append(finalRank, &rankModel.FinalRankItem{
						Position: model.Position,
						Mode:     model.Mode,
						Title:    model.Title,
						List:     avidList[0:end],
					})
				} else if model.Mode == 2 {
					finalRank = append(finalRank, &rankModel.FinalRankItem{
						Position: model.Position,
						Mode:     model.Mode,
						Title:    model.Title,
						List:     model.UidList,
					})
				}
			}
		}

		pagerList = append(pagerList, &rankModel.OpenListItem{
			ID:         config.ID,
			RankConfig: configDetail,
			RankVideos: rankVideos,
			RankState:  config.State,
			FinalRank:  finalRank,
		})
	}

	pager = &rankModel.OpenListPager{
		List: pagerList,
		Page: common.Page{
			Size:  ps,
			Num:   pn,
			Total: count,
		},
	}

	return
}
