package service

import (
	"context"
	"encoding/json"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/dance-taiko/job/model"
)

const _insert = "insert"

func (s *Service) syncGame() {
	defer s.waiter.Done()
	for {
		msg, ok := <-s.danceBinlog.Messages()
		if !ok {
			log.Error("err consume s.danceBinlog.Messages")
			return
		}
		if err := msg.Commit(); err != nil {
			log.Error("syncGame err(%v)", err)
		}
		s.treatMsg(msg.Value)
	}
}

func (s *Service) treatMsg(msg json.RawMessage) {
	log.Info("Dance New Message: %s", msg)
	var (
		c       = context.Background()
		dataBus = new(model.DatabusMsg)
	)
	if err := json.Unmarshal(msg, &dataBus); err != nil {
		log.Error("msg unmarshal failed. msg(%s) err(%v)", msg, err)
		return
	}
	switch dataBus.Table {
	case model.DanceGame:
		var gameMsg = new(model.GameDatabus)
		if err := json.Unmarshal(msg, &gameMsg); err != nil {
			log.Error("msg unmarshal failed. table(%s) err(%v)", dataBus.Table, err)
			return
		}
		if err := s.handleDanceGame(c, gameMsg); err != nil {
			log.Error("s.handleDanceGame failed. gameMsg(%v) err(%v)", gameMsg, err)
		}
	case model.DancePlayers:
		if dataBus.Action == _insert { // insert时同步缓存
			var playersMsg = new(model.PlayersDatabus)
			if err := json.Unmarshal(msg, &playersMsg); err != nil {
				log.Error("msg unmarshal failed. table(%s) err(%v)", dataBus.Table, err)
				return
			}
			if err := s.handleDancePlayers(c, playersMsg); err != nil {
				log.Error("s.handleDancePlayer failed. playersMsg(%v) err(%v)", playersMsg, err)
			}
		}
	default:
		log.Error("wrong table.")
	}
}

func (s *Service) handleDanceGame(c context.Context, gameMsg *model.GameDatabus) error {
	if gameMsg.New == nil {
		log.Error("New message is Nil.")
		return nil
	}
	gameId := gameMsg.New.GameId
	s.fanout.Do(c, func(ctx context.Context) {
		if err := s.dao.AddCacheGame(ctx, gameMsg.New); err != nil {
			log.Error(" s.dao.AddCacheGame value(%v) err(%v)", gameMsg.New, err)
		}
	})
	if gameMsg.Old == nil {
		return nil
	}
	if gameMsg.Old.Status == model.GamePlaying && gameMsg.New.Status == model.GameFinished {
		var (
			eg     = errgroup.WithContext(c)
			hisMap map[int64]*model.PlayerHonor
			curMap map[int64]*model.PlayerHonor
		)
		eg.Go(func(ctx context.Context) (err error) {
			if curMap, err = s.dao.CachePlayerMap(c, gameId); err != nil {
				log.Error("handleDanceGame curPlayers filed. GameId(%d) Err(%v)", gameId, err)
			}
			return nil
		})
		eg.Go(func(ctx context.Context) (err error) {
			if hisMap, err = s.dao.PlayersMap(c, gameId); err != nil {
				log.Error("handleDanceGame hisPlayers filed. GameId(%d) Err(%v)", gameId, err)
			}
			return nil
		})
		_ = eg.Wait()

		if len(hisMap) == 0 {
			log.Error("日志报警 handleDanceGame game(%v) err(%v)", gameMsg.New, ecode.NothingFound)
			return ecode.NothingFound
		}
		var players []*model.PlayerHonor
		for mid, player := range hisMap {
			curPlayer, ok := curMap[mid]
			if !ok {
				log.Error("日志报警 handleDanceGame mid(%d) curMap wrong.", player.Mid)
				if player.Score == 0 {
					player.Score = int64(s.c.Cfg.DefaultScore) // 历史分数和当前分数都不存在，用默认分数填充
					players = append(players, player)
				}
				continue
			}
			if curPlayer.Score > player.Score { // 当前分数大于历史分数 更新
				player.Score = curPlayer.Score
				players = append(players, player)
			}
		}
		if len(players) == 0 {
			return nil
		}
		if err := s.dao.UpdatePlayers(c, gameId, players); err != nil {
			log.Error("handleDanceGame update failed.GameId(%d) Players(%v) Err(%v)", gameId, players, err)
			return err
		}
	}
	return nil
}

func (s *Service) handleDancePlayers(c context.Context, playerMsg *model.PlayersDatabus) error {
	if playerMsg.New == nil {
		log.Error("New message is Nil.")
		return nil
	}
	_ = s.fanout.Do(c, func(ctx context.Context) {
		_ = s.dao.AddCachePlayers(ctx, playerMsg.New.GameId, []model.PlayerHonor{
			{Mid: playerMsg.New.Mid, Score: playerMsg.New.Score},
		})
	})
	return nil
}
