package service

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/web-svr/dance-taiko/interface/api"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

func (s *Service) OttGameStat(c context.Context, buvid string, gameID, mid int64, stats []*model.Stat) error {

	// 校验游戏状态
	game, err := s.ottDao.LoadGame(c, gameID)
	if err != nil {
		log.Error("日志报警 OttGameStat LoadGame_GameID %d Err %v", gameID, err)
		return err
	}
	if game.Status != model.GamePlaying { // 游戏未开始不处理数据
		log.Warn("OttGameStat GameID %d Not Started", gameID)
		return nil
	}

	statAccs := make([]*api.StatAcc, 0, len(stats))
	for _, v := range stats {
		// ott侧不减时间，由job处理
		statAccs = append(statAccs, v.GenerateAcc())
	}

	// 将数据存入redis中，由job进行消费和评分
	// 注意重复数据会被zadd去重
	if err := s.ottDao.AddCachePlayerStat(c, gameID, mid, statAccs); err != nil {
		log.Error("日志报警 OttGameStat AddCachePlayerStat_GameID(%d) Mid(%d) Err %v", gameID, mid, err)
		return err
	}

	return nil
}
