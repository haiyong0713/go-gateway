package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"sort"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/dance-taiko/interface/api"
	xecode "go-gateway/app/web-svr/dance-taiko/interface/ecode"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	account "git.bilibili.co/bapis/bapis-go/account/service"

	"github.com/golang/protobuf/ptypes/empty"
)

func (s *Service) Current(ctx context.Context, req *empty.Empty) (resp *api.CurrentResp, err error) {
	curGame, err := s.dao.CurrentGame(context.Background())
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			curGame = &model.Game{
				GameID: -1,
				Status: model.GameFinished,
			}
		} else {
			return nil, ecode.ServiceUnavailable
		}
	}
	var filePath string
	filePath, err = s.dao.RedisGetFilePath(ctx)
	if err != nil {
		log.Error("日志报警 GameID %d", curGame.GameID)
		err = nil
	}
	resp = &api.CurrentResp{
		GameId:   curGame.GameID,
		FilePath: filePath,
	}
	return
}

func (s *Service) Create(ctx context.Context, req *api.CreateReq) (*api.CreateResp, error) {
	curGame, err := s.dao.CurrentGame(context.Background())
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			curGame = &model.Game{
				Status: model.GameFinished,
			}
		} else {
			return nil, ecode.ServiceUnavailable
		}
	}
	if curGame.Status != model.GameFinished {
		return nil, ecode.MethodNotAllowed
	}
	var gameId int64
	if s.conf.BwsCfg.EnableBws && req.Experiential == "false" { // 本场游戏不是体验场
		s.dao.BwsReset(ctx) // 预先清除所有房间，确保房间开启成功
		id, err := s.dao.BwsCreateRoom(ctx)
		if err != nil {
			log.Error("Create aid(%d) err(%v)", req.Aid, err)
			return nil, ecode.ServiceUnavailable
		}
		gameId = int64(id)
	}
	if gameId == 0 { // 没有gameId 用当前时间戳代替
		gameId = time.Now().Unix()
		if err := s.dao.AddRedisExperiment(ctx, gameId); err != nil {
			log.Error("AddRedisExperiment failed. err %v", err) // 忽略错误，设置日志报警
		}
	}
	if err = s.dao.CreateGame(ctx, req.Aid, gameId); err != nil {
		return nil, ecode.ServiceUnavailable
	}
	resp := &api.CreateResp{
		GameId: gameId,
	}
	return resp, nil

}

func (s *Service) Start(ctx context.Context, req *api.StartReq) (resp *empty.Empty, err error) {
	curGame, err := s.dao.CurrentGame(context.Background())
	if err != nil {
		return nil, ecode.MethodNotAllowed
	}
	if curGame.GameID != req.GameId {
		return nil, ecode.Unauthorized
	}
	if curGame.Status != model.GameJoining {
		return nil, ecode.MethodNotAllowed
	}
	expri, _ := s.dao.RedisExperiment(ctx, req.GameId)
	if s.conf.BwsCfg.EnableBws && !expri { // 非体验场走bws逻辑
		if err = s.dao.BwsStartGame(ctx, int(req.GameId)); err != nil {
			return nil, ecode.MethodNotAllowed
		}
	}
	err = s.dao.StartGame(ctx, req.GameId, model.GamePlaying)
	return
}

func (s *Service) Status(ctx context.Context, req *api.StatusReq) (resp *api.StatusResp, err error) {
	curGame, err := s.dao.CurrentGame(context.Background())
	if err != nil {
		return nil, ecode.MethodNotAllowed
	}
	if curGame.GameID != req.GameId {
		return nil, ecode.Unauthorized
	}

	resp = &api.StatusResp{
		GameDtatus:   curGame.Status,
		PlayerStatus: []*api.PlayerStatus{},
	}
	var (
		players  map[int]string
		points   map[int64]int64
		comments map[int64]string
	)
	if players, err = s.dao.RedisGetJoinedPlayers(ctx, req.GameId); err != nil {
		return
	}
	if points, err = s.dao.RedisGetPoints(ctx, req.GameId); err != nil {
		return
	}
	if comments, err = s.dao.RedisGetComments(ctx, req.GameId); err != nil {
		return
	}

	for p := 1; p <= len(players); p++ {
		pJson, ok := players[p]
		if !ok {
			log.Error("position:%d not existed", p)
			continue
		}
		player := model.Player{}
		if err = json.Unmarshal([]byte(pJson), &player); err != nil {
			log.Error("json:%s unmarshal failed", pJson)
			err = nil
			continue
		}

		pStatus := &api.PlayerStatus{
			Mid:  player.Mid,
			Name: player.Name,
			Face: player.Face,
		}
		if v, ok := points[player.Mid]; ok {
			pStatus.Points = v
		}
		if v, ok := comments[player.Mid]; ok {
			pStatus.LastComment = v
		}
		resp.PlayerStatus = append(resp.PlayerStatus, pStatus)
	}

	return
}

func (s *Service) Join(ctx context.Context, req *api.JoinReq) (resp *api.JoinResp, err error) {
	curGame, err := s.dao.CurrentGame(context.Background())
	if err != nil {
		return nil, ecode.MethodNotAllowed
	}
	if curGame.GameID != req.GameId {
		return nil, ecode.Unauthorized
	}
	expri, _ := s.dao.RedisExperiment(ctx, req.GameId)
	var players map[int]string
	players, err = s.dao.RedisGetJoinedPlayers(ctx, req.GameId)
	if err != nil {
		return nil, ecode.ServiceUnavailable
	}

	for pos, str := range players {
		player := model.Player{}
		if err = json.Unmarshal([]byte(str), &player); err != nil {
			log.Error("json:%s unmarshal failed", str)
			err = nil
			continue
		}
		if player.Mid == req.Mid { // 曾经加入过的玩家
			resp = &api.JoinResp{
				ServerTime: time.Now().UnixNano() / int64(time.Millisecond),
				Position:   int64(pos),
			}
			return
		}
	}

	if len(players) >= 20 {
		return nil, ecode.LimitExceed
	}

	if curGame.Status != model.GameJoining { // 只有在joining时候才接受新玩家，其他场景需要玩家已加入过
		return nil, ecode.MethodNotAllowed
	}

	var (
		playerJson []byte
		card       *account.Card
		eg         = errgroup.WithContext(ctx)
		playInfo   *model.BwsPlayInfo
	)
	eg.Go(func(ctx context.Context) (err error) {
		card, err = s.user(ctx, req.Mid)
		return err
	})
	if s.conf.BwsCfg.EnableBws && !expri { // 非体验场走bws逻辑
		eg.Go(func(ctx context.Context) (err error) {
			playInfo, err = s.dao.BwsMidInfo(ctx, req.Mid, int(req.GameId))
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("Join mid(%d) game(%v) err(%v)", req.Mid, curGame, err)
		return nil, ecode.ServiceUnavailable
	}
	if playInfo != nil && (!playInfo.Valid || playInfo.Energy < s.conf.BwsCfg.NeedEnergy) {
		log.Error("Join mid(%d) game(%v) playInfo(%v)", req.Mid, curGame, playInfo)
		return nil, xecode.PlayerErr
	}
	p := &model.Player{
		Mid:  req.Mid,
		Face: card.Face,
		Name: card.Name,
	}
	position := len(players) + 1

	if playerJson, err = json.Marshal(p); err != nil {
		return
	}
	err = s.dao.RedisJoin(ctx, req.GameId, position, string(playerJson))
	if err != nil {
		return nil, ecode.ServiceUnavailable
	}
	if s.conf.BwsCfg.EnableBws {
		if err := s.dao.BwsJoinRoom(ctx, int(curGame.GameID), req.Mid); err != nil {
			log.Error("Join gameId(%d) mid(%d) err(%v)", curGame.GameID, req.Mid, err)
		}
	}
	resp = &api.JoinResp{
		ServerTime: time.Now().UnixNano() / int64(time.Millisecond),
		Position:   int64(position),
	}
	return
}

func (s *Service) Finish(ctx context.Context, req *api.FinishReq) (resp *empty.Empty, err error) {
	curGame, err := s.dao.CurrentGame(context.Background())
	if err != nil {
		return nil, ecode.MethodNotAllowed
	}
	if curGame.GameID != req.GameId {
		return nil, ecode.Unauthorized
	}
	if err = s.dao.UpdateGameStatus(ctx, req.GameId, model.GameFinished); err != nil {
		log.Error("日志报警 gameID %d UpdateGameStatus, err %v", req.GameId, err)
		return nil, err
	}
	expri, _ := s.dao.RedisExperiment(ctx, req.GameId)
	// 这个错误是忽略的
	if e := func() error {
		points, err := s.dao.RedisGetPoints(ctx, req.GameId)
		if err != nil {
			return err
		}
		if err := s.dao.RedisSetUserPoints(ctx, curGame.AID, points); err != nil {
			return err
		}
		if !s.conf.BwsCfg.EnableBws && expri { // 体验场不做星结算
			return nil
		}
		gameId, _ := s.dao.RedisGetGame(ctx, int(curGame.GameID))
		if int64(gameId) == curGame.GameID { // 请求过restart的房间，结算星
			var players []*model.BwsPlayResult
			for mid, point := range points {
				players = append(players, &model.BwsPlayResult{
					Mid:   mid,
					Score: point,
				})
			}
			sort.Slice(players, func(i, j int) bool { // 对分数进行排序
				return players[i].Score > players[j].Score
			})
			for index, player := range players {
				if index < 10 {
					player.Star = 3
				} else if index < 15 {
					player.Star = 2
				} else {
					player.Star = 1
				}
			}
			if err := s.dao.BwsEndGame(ctx, int(curGame.GameID), players); err != nil { // 上报
				return err
			}
		}
		return nil
	}(); e != nil {
		log.Error("日志报警 gameID %d RedisTopTen Logic, err %v", req.GameId, err)
	}

	return
}

func (s *Service) Upload(ctx context.Context, fileType string, body io.Reader) (location string, err error) {
	location, err = s.dao.BfsUpload(ctx, fileType, body)
	if err == nil {
		err = s.dao.RedisSetFilePath(ctx, location)
		if err != nil {
			log.Error("日志报警 Upload set redis err:%v", err)
		}
	} else {
		log.Error("日志报警 Upload location err:%v", err)
	}
	return
}

func (s *Service) ReStart(ctx context.Context, req *api.ReStartReq) (resp *empty.Empty, err error) {
	curGame, err := s.dao.CurrentGame(context.Background())
	if err != nil {
		return nil, ecode.MethodNotAllowed
	}
	if curGame.GameID != req.GameId {
		return nil, ecode.Unauthorized
	}
	if curGame.Status != model.GameFinished {
		return nil, ecode.MethodNotAllowed
	}
	err = s.dao.RedisDelPoints(ctx, curGame.GameID)
	err = s.dao.StartGame(ctx, req.GameId, model.GamePlaying)
	s.dao.RedisSetGame(ctx, int(curGame.GameID)) // 把游戏加入缓存
	return
}
