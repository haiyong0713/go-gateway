package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	accClient "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/web-svr/dance-taiko/interface/ecode"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	qrcode "github.com/skip2/go-qrcode"
)

func (s *Service) GameCreate(c context.Context, aid, cid int64) (*model.GameCreateReply, error) {
	gameId, err := s.ottDao.CreateGame(c, aid, cid)
	if err != nil {
		log.Error("GameCreate aid(%d) cid(%d) err(%v)", aid, cid, err)
		return nil, err
	}
	return &model.GameCreateReply{GameId: gameId}, nil
}

func (s *Service) GameStart(c context.Context, gameId int64) (*model.GameJoinReply, error) {
	var (
		eg      = errgroup.WithContext(c)
		game    *model.OttGame
		players []*model.PlayerHonor
		err     error
	)
	eg.Go(func(ctx context.Context) error {
		game, err = s.ottDao.LoadGame(c, gameId)
		return err
	})
	eg.Go(func(ctx context.Context) error {
		players, err = s.ottDao.LoadPlayers(c, gameId)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Error("GameStart Err(%v)", err)
		return nil, xecode.GameStatusErr
	}
	if game.Status != model.GameJoining || len(players) == 0 {
		log.Error("GameStart game(%v) err(%v)", game, xecode.GameStatusErr)
		return nil, xecode.GameStatusErr
	}
	if err = s.ottDao.StartGame(c, gameId); err != nil {
		log.Error("GameStart gameId(%d) err(%v)", gameId, err)
		return nil, xecode.GameStatusErr
	}
	if err = s.ottDao.DelCacheGame(c, gameId); err != nil { // 删除缓存，强制回源
		log.Error("GameStart cache failed.gameId(%d) err(%v)", gameId, err)
	}
	return &model.GameJoinReply{ServerTime: s.timeNow()}, nil
}

func (s *Service) GameJoin(c context.Context, gameId, mid int64) (*model.GameJoinReply, error) {
	var (
		eg      = errgroup.WithContext(c)
		game    *model.OttGame
		players []*model.PlayerHonor
	)
	eg.Go(func(ctx context.Context) (err error) {
		game, err = s.ottDao.LoadGame(c, gameId)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		players, err = s.ottDao.LoadPlayers(c, gameId)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		_, err = s.user(c, mid) // 验证用户是否存在
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("GameJoin Err(%v)", err)
		return nil, err
	}
	serverTime := s.timeNow()
	for _, player := range players { // 之前加入过的用户 可直接再次加入
		if mid == player.Mid {
			return &model.GameJoinReply{ServerTime: serverTime}, nil
		}
	}
	if game.Status != model.GameJoining {
		return nil, xecode.GameStatusErr
	}
	if len(players) >= s.conf.OttCfg.PlayersMax {
		return nil, xecode.PlayerBeyound
	}
	if err := s.ottDao.AddPlayer(c, gameId, mid); err != nil { // 更新db
		log.Error("GameJoin gameId(%d) mid(%d)", gameId, mid)
		return nil, err
	}
	return &model.GameJoinReply{ServerTime: serverTime}, nil
}

func (s *Service) GameFinish(c context.Context, gameId int64) error {
	game, err := s.ottDao.LoadGame(c, gameId)
	if err != nil {
		log.Error("GameFinish gameId(%d) err(%v)", gameId, err)
		return err
	}
	if game.Status != model.GamePlaying {
		log.Error("GameFinish game(%v) err(%v)", game, err)
		return xecode.GameStatusErr
	}
	if err := s.ottDao.FinishGame(c, gameId); err != nil {
		log.Error("GameFinish gameId(%d) err(%v)", gameId, err)
		return err
	}
	if err := s.ottDao.DelCacheGame(c, gameId); err != nil { // 删除缓存，强制回源
		log.Error("GameFinish gameId(%d) err(%v)", gameId, err)
	}
	// 排名结算 错误忽略，不影响其他数据显示
	if err := s.updatePlayersRank(c, gameId, game.Cid); err != nil {
		log.Error("GameFinish gameId(%d) err(%v)", gameId, err)
	}
	return nil
}

func (s *Service) updatePlayersRank(c context.Context, gameId, cid int64) error {
	var (
		players = make([]*model.PlayerHonor, 0)
		err     error
	)
	if players, err = s.ottDao.CachePlayer(c, gameId); err != nil { // 缓存失败，查db获取mid，用默认分数填充
		log.Error("updatePlayersRank gameId(%d), err(%v)", gameId, err)
		players, err = s.ottDao.RawPlayers(c, gameId)
		if err != nil {
			log.Error("updatePlayersRank gameId(%d) err(%v)", gameId, err)
			return err
		}
		for _, player := range players {
			players = append(players, &model.PlayerHonor{Mid: player.Mid, Score: int64(s.conf.OttCfg.DefaultScore)})
		}
	}
	if err = s.ottDao.AddCacheRank(c, cid, players); err != nil {
		log.Error("updatePlayersRank cid(%d) err(%v)", cid, err)
		return err
	}
	return nil
}

func (s *Service) GameStatus(c context.Context, gameId, playTime, nalTime int64) (*model.GameStatusReply, error) {
	game, err := s.ottDao.LoadGame(c, gameId)
	if err != nil {
		log.Error("GameStatus gameId(%d) err(%v)", gameId, err)
		return nil, err
	}
	// 如果进度时间落后自然时间，落cache
	if playTime < nalTime {
		if err := s.ottDao.AddCacheGameGap(c, gameId, nalTime-playTime); err != nil {
			log.Error("GameStatus err(%v)", err)
		}
	}

	res := &model.GameStatusReply{GameStatus: game.Status}
	switch game.Status {
	case model.GameJoining:
		playStatus, _, err := s.loadPlayerCards(c, gameId)
		if err != nil {
			log.Error("GameStatus gameId(%v) err(%v)", gameId, err)
			return nil, err
		}
		if len(playStatus) == 0 {
			return res, nil
		}
		res.PlayerStatus = playStatus
	case model.GameFinished:
		playStatus, mids, err := s.loadPlayerCards(c, gameId)
		if err != nil {
			log.Error("GameStatus gameId(%v) err(%v)", gameId, err)
			return nil, err
		}
		if len(playStatus) == 0 {
			log.Error("GameStatus gameId(%v) err(%v)", gameId, xecode.PlayerErr)
			return nil, xecode.PlayerErr
		}
		sort.Slice(playStatus, func(i, j int) bool {
			return playStatus[i].Points > playStatus[j].Points
		})
		rankMap, err := s.ottDao.CachePlayersRank(c, game.Cid, mids)
		if err != nil {
			log.Warn("GameStatus mids(%v) err(%v)", mids, err)
			break
		}
		for _, player := range playStatus {
			if rankMap[player.Mid] >= 0 {
				player.GlobalRank = rankMap[player.Mid] + 1
			}
		}
		res.PlayerStatus = playStatus
	case model.GamePlaying:
		players, err := s.ottDao.CachePlayer(c, gameId)
		if err != nil {
			log.Error("GameStatus gameId(%v) err(%v)", gameId, err)
			return nil, err
		}
		if players == nil {
			log.Error("GameStatus gameId(%v) err(%v)", gameId, xecode.PlayerErr)
			return nil, xecode.PlayerErr
		}

		var mids []int64
		for _, player := range players {
			mids = append(mids, player.Mid)
		}
		var (
			comboMap   map[int64]int
			userMap    map[int64]*accClient.Card
			commentMap map[int64]string
			eg         = errgroup.WithCancel(c)
		)
		if len(mids) > 0 {
			eg.Go(func(ctx context.Context) error {
				if commentMap, err = s.ottDao.CachePlayerComment(ctx, gameId); err != nil {
					return err
				}
				if err := s.ottDao.DelPlayerComment(ctx, gameId); err != nil { // 保证一个comment只展示一次
					log.Error("GameStatus s.ottDao.DelPlayerComment id(%d) err(%v)", gameId, err)
				}
				return nil
			})
			eg.Go(func(ctx context.Context) error {
				comboMap, err = s.ottDao.CachePlayersCombo(ctx, gameId, mids)
				return err
			})
			eg.Go(func(ctx context.Context) error {
				userMap, err = s.ottDao.UserCards(ctx, mids)
				return err
			})
			if err = eg.Wait(); err != nil {
				log.Error("GameStatus Err(%v)", err)
				return nil, err
			}
		}

		for _, p := range players {
			user, ok := userMap[p.Mid]
			if !ok {
				continue
			}
			player := new(model.Player)
			player.CopyFromGRPC(user)
			comment, _ := commentMap[p.Mid]
			combo, _ := comboMap[p.Mid]
			res.PlayerStatus = append(res.PlayerStatus, &model.PlayerInfo{
				Player:      player,
				LastComment: comment,
				Points:      int(p.Score),
				ComboTimes:  combo,
			})
		}
	default:
		log.Error("GameStatus game(%v)", game)
		return nil, xecode.GameStatusErr
	}
	return res, nil
}

func (s *Service) loadPlayerCards(c context.Context, gameId int64) ([]*model.PlayerInfo, []int64, error) {
	players, err := s.ottDao.LoadPlayers(c, gameId)
	if err != nil {
		log.Error("loadPlayerCards gameId(%d) err(%v)", gameId, err)
		return nil, nil, err
	}
	var (
		mids []int64
		res  = make([]*model.PlayerInfo, 0)
	)
	for _, player := range players {
		mids = append(mids, player.Mid)
	}
	if len(mids) == 0 {
		return res, nil, nil
	}
	playerMap, err := s.ottDao.UserCards(c, mids)
	if err != nil {
		log.Error("loadPlayerCards mids(%d) err(%v)", mids, err)
		return nil, nil, err
	}
	for _, player := range players {
		info, ok := playerMap[player.Mid]
		if !ok {
			log.Error("loadPlayerCards mid(%d) err(%v)", player.Mid, err)
			return nil, nil, err
		}
		var playerInfo = new(model.Player)
		playerInfo.CopyFromGRPC(info)
		res = append(res, &model.PlayerInfo{Player: playerInfo, Points: int(player.Score)})
	}
	return res, mids, nil
}

func (s *Service) GamePkgUpload(c context.Context, fileType string, body io.Reader) error {
	location, err := s.dao.BfsUpload(c, fileType, body)
	if err != nil {
		log.Error("GamePkgUpload err(%v)", err)
		return err
	}
	log.Info("GamePkgUpload url(%s)", location)
	err = s.ottDao.AddCacheGamePkg(c, location)
	if err != nil {
		log.Error("GamePkgUpload location(%s) err(%v)", location, err)
		return err
	}
	return nil
}

func (s *Service) LoadQRCode(c context.Context, id int64) (*model.QRCodeReply, error) {
	res := &model.QRCodeReply{
		Msg: s.conf.OttCfg.QRCodeMsg,
	}
	url, err := s.ottDao.CacheGameQRCode(c, id)
	if err == nil && len(url) > 0 {
		res.QRCode = url
		return res, nil
	}
	pkgUrl, err := s.ottDao.CacheGamePkg(c)
	if err != nil {
		log.Error("LoadQRCode err(%v)", err)
		return nil, err
	}
	url = fmt.Sprintf(s.conf.OttCfg.QRCodeUrl, pkgUrl, id)
	png, err := qrcode.Encode(url, qrcode.Medium, s.conf.OttCfg.QRCodeSize) // 生成二维码
	if err != nil {
		log.Error("LoadQRCode url(%s) err(%v)", url, err)
		return nil, err
	}
	fileType := http.DetectContentType(png)
	location, err := s.dao.BfsUpload(c, fileType, bytes.NewReader(png)) // 上传bfs
	if err != nil {
		log.Error("LoadQRCode fileType(%s) err(%v)", fileType, err)
		return nil, err
	}
	s.fanout.Do(c, func(ctx context.Context) {
		if err := s.ottDao.AddCacheQRCode(ctx, id, location); err != nil {
			log.Error("LoadQRCode id(%d) err(%v)", id, err)
		}
	})
	res.QRCode = location
	return res, nil
}

func (s *Service) GameRestart(c context.Context, gameId int64) (*model.GameJoinReply, error) {
	var (
		eg      = errgroup.WithCancel(c)
		game    *model.OttGame
		players []*model.PlayerHonor
	)
	eg.Go(func(ctx context.Context) (err error) {
		game, err = s.ottDao.LoadGame(c, gameId)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		players, err = s.ottDao.LoadPlayers(c, gameId)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("GameRestart gameId(%d) err(%v)", gameId, err)
		return nil, xecode.RestartErr
	}
	if len(players) == 0 || game.Status != model.GameFinished {
		log.Error("GameRestart game(%v) failed.", game)
		return nil, xecode.RestartErr
	}
	if err := s.ottDao.StartGame(c, gameId); err != nil {
		log.Error("GameRestart gameId(%d) err(%v)", gameId, err)
		return nil, xecode.GameStatusErr
	}
	var mids []int64
	for _, player := range players {
		mids = append(mids, player.Mid)
		player.Score = 0 // 重置用户的得分
	}
	// 删除缓存
	eg = errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) error {
		return s.ottDao.DelCaches(c, gameId, mids)
	})
	eg.Go(func(ctx context.Context) error {
		return s.ottDao.AddCachePLayer(c, gameId, players)
	})

	if err := eg.Wait(); err != nil {
		log.Error("日志报警 GameRestart cache failed.GameId(%d) Err(%v)", gameId, err)
		return nil, xecode.GameStatusErr
	}
	log.Info("Game(%v) restart.", game)
	return &model.GameJoinReply{ServerTime: s.timeNow()}, nil
}

func (s *Service) timeNow() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
