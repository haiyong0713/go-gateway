package service

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/net/http/blademaster"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/web-svr/dance-taiko/interface/api"
	"go-gateway/app/web-svr/dance-taiko/job/model"
)

type CommentType string

const (
	// 结算周期，500ms
	_settlementInterval = 500 // ms

	_commentMiss    = CommentType("Miss")
	_commentBad     = CommentType("Bad")
	_commentGood    = CommentType("Good")
	_commentSuper   = CommentType("Super")
	_commentPerfect = CommentType("Perfect")
)

// 拦截一下panic，怕炸
func recoverFunc(f func()) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("日志报警 gameStat Panic %v", e)
		}
	}()
	f()
}

func (s *Service) gameStatDual() {

	settlement := func() {
		if s.serviceClosed {
			return
		}

		// 统计结算时间
		now := time.Now()
		defer func() {
			dt := time.Since(now)
			blademaster.MetricServerReqDur.Observe(int64(dt/time.Millisecond), "gameStat", "job")
		}()

		// 结算周期确认
		begin := time.Now().UnixNano() / int64(time.Millisecond)
		log.Info("gameStatDual counts from %d to %d", begin, begin+_settlementInterval)

		// 发版时保证一次结算流程的完整性
		s.waiter.Add(1)
		defer s.waiter.Done()

		// 获取当前正在进行中的游戏
		games, err := s.dao.GamesByStatus(context.Background(), model.GamePlaying)
		if err != nil {
			log.Error("日志报警 GamesByStatus Playing Err %v", err)
			return
		}
		if len(games) == 0 {
			return
		}
		log.Warn("gameStatDual counts %v", games)

		// 启动多个goroutine进行结算操作
		eg := errgroup.WithContext(context.Background())
		eg.GOMAXPROCS(s.c.Cfg.StatCurrency)
		for _, game := range games {
			g := game
			eg.Go(func(c context.Context) error {
				recoverFunc(func() {
					s.gameStatByOne(c, g, begin)
				})
				return nil
			})
		}

		if err := eg.Wait(); err != nil {
			log.Error("gameStatDual Err %v", err)
		}
		return
	}

	// 每500毫秒启动一次结算逻辑
	go recoverFunc(settlement)
	time.AfterFunc(_settlementInterval*time.Millisecond, func() {
		go recoverFunc(settlement)
	})
}

func (s *Service) gameStatByOne(c context.Context, game model.OttGame, begin int64) {

	eg := errgroup.WithCancel(c)

	// 【G1】1. 获取自然时间和时间时间的对应关系，及当前结算区间
	var gameGap, left, right int64
	eg.Go(func(c context.Context) error {
		var err error
		if gameGap, err = s.dao.CacheGameGap(c, game.GameId); err != nil {
			log.Error("日志报警 gameStatByOne CacheGameGap GameID %d Err %v", game.GameId, err)
			return err
		}
		if gameGap >= 2000 { // 太卡了，日志报警
			log.Warn("日志报警 gameStatByOne Gap时间过长 GameID %d，Gap %d Err %v", game.GameId, gameGap, err)
		}
		left = begin - game.Stime - gameGap
		right = left + _settlementInterval
		return nil
	})

	// 【G1】2.1 找到本周期内需要结算的关键帧
	var keyFrames []int64
	eg.Go(func(c context.Context) error {
		var err error
		if keyFrames, err = s.dao.RawKeyFrames(c, game.Aid, game.Cid); err != nil {
			log.Error("日志报警 gameStatByOne RawKeyFrames Aid %d, Cid %d, Err %v", game.Aid, game.Cid, err)
			return err
		}
		return nil
	})

	// 【G1】3. 获取当前游戏的用户信息
	var players []*model.PlayerHonor
	eg.Go(func(c context.Context) error {
		var err error
		if players, err = s.dao.RawPlayers(c, game.GameId); err != nil {
			log.Error("日志报警 gameStatByOne RawPlayers GameID %d Err %v", game.GameId, err)
			return err
		}
		if len(players) == 0 { // 无人游戏，不结算
			return fmt.Errorf("NoUser")
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return
	}
	mids := make([]int64, 0, len(players))
	for _, v := range players {
		mids = append(mids, v.Mid)
	}
	log.Warn("gameStatByOne GameID %d, PlayerMids %v", game.GameId, mids)

	// 2.2 找到本周期需要计算的关键帧：增加延迟结算时间，等待客户端上报数据；左开右闭
	frames := make([]int64, 0)
	for _, v := range keyFrames {
		if delay := v + s.c.Cfg.StatDelay; delay > left && delay <= right {
			frames = append(frames, v)
		}
	}
	if len(frames) == 0 {
		log.Info("gameStatByOne GameID %d, Aid %d, HitFrames Empty", game.GameId, game.Aid)
		return
	}
	log.Info("gameStatByOne GameID %d, Aid %d, HitFrames %v", game.GameId, game.Aid, frames)

	// 获取关键帧的最大和最小，考虑卡顿、开始时间和计算区间，一次性获取该范围内用户的游戏数据
	min, max := minAndMax(frames)
	minGame, maxGame := min-s.c.Cfg.Boundary, max+s.c.Cfg.Boundary
	minNatural, maxNatural := minGame+gameGap+game.Stime, maxGame+gameGap+game.Stime
	log.Warn("gameStatByOne GameID %d Min %d, MinGame %d, MinNatural %d, Max %d, MaxGame %d, MaxNatural %d",
		game.GameId, min, minGame, minNatural, max, maxGame, maxNatural)

	eg = errgroup.WithContext(c)
	eg.GOMAXPROCS(5) // 不知道player数量，加个限制

	// 【G2】4. 获取当前范围内的标准数据并且进行平滑操作：这里ts为游戏时间
	smoothExamples := make(map[int64]api.StatAcc, 0)
	eg.Go(func(c context.Context) error {

		// 两边多取一点，填补examples的空白
		examples, err := s.dao.PickExamples(c, game.Aid, game.Cid, minGame-s.c.Cfg.Boundary, maxGame+s.c.Cfg.Boundary)
		if err != nil {
			log.Error("日志报警 gameStatByOne PickExamples GameID %d, minGame %d, maxGame %d Err %v", game.GameId, minGame, maxGame, err)
			return err
		}
		if len(examples) <= 1 {
			log.Error("日志报警 gameStatByOne PickExamples GameID %d, minGame %d, maxGame %d Empty Example", game.GameId, minGame, maxGame)
			return fmt.Errorf("NoExample")
		}

		lastPos := int64(minGame)
		lastEucli := float64(0)

		for i := 0; i < len(examples); i++ {
			currentPos := examples[i].Ts
			currentEucli, err := examples[i].Euclidean()
			if err != nil {
				log.Error("日志报警 gameStatByOne PickExamples GameID %d, Json Err %v", game.GameId, err)
				return err
			}

			step := (currentEucli - lastEucli) / float64(currentPos-lastPos) // 求模长平滑后的步长
			for j := lastPos; j < currentPos; j++ {                          // 逐秒平滑
				smoothExamples[j] = api.StatAcc{
					Acc: lastEucli + float64(j-lastPos)*step,
					Ts:  j,
				}
			}

			lastPos = currentPos
			lastEucli = currentEucli
		}
		smoothExamples[lastPos] = api.StatAcc{ // 补齐最后一条数据
			Acc: lastEucli,
			Ts:  lastPos,
		}
		return nil
	})

	// 【G2】5. 获取时间范围内的用户上报数据
	playerStats := make(map[int64][]*api.StatAcc, len(players))
	mutex := sync.Mutex{}
	for _, player := range players {
		p := player
		eg.Go(func(c context.Context) error {
			stats, err := s.dao.PickPlayerStats(c, game.GameId, p.Mid, minNatural, maxNatural)
			if err != nil {
				log.Warn("日志报警 gameStatByOne PickPlayerStats GameID %d, Mid %d Err %v", game.GameId, p.Mid, err)
				return nil
			}
			mutex.Lock()
			playerStats[p.Mid] = stats
			mutex.Unlock()
			log.Info("gameStatByOne minNatural(%d) maxNatural(%d) GameId(%d) Mid(%d) stats %v", minNatural, maxNatural, game.GameId, p.Mid, stats)
			return nil
		})
	}

	// 【G2】6. 获取用户的combo数据
	var preCombos map[int64]int64 // key=mid, value=combo
	eg.Go(func(c context.Context) error {
		var err error
		preCombos, err = s.dao.CachePlayersCombo(c, game.GameId, mids)
		if err != nil {
			log.Warn("日志报警 gameStatByOne PickPlayerStats GameID %d CachePlayersCombo Err %v", game.GameId, err)
		}

		// 拿不到combo不阻断体验
		return nil
	})

	if err := eg.Wait(); err != nil {
		return
	}

	// 7. 得分、评价、Combo结算
	var (
		playerCombos  []model.PlayerCombo
		playerComment []model.PlayerComment
		playerScore   []model.PlayerHonor
	)
	for _, player := range players {
		stats, ok := playerStats[player.Mid]
		if !ok {
			log.Warn("日志报警 gameStatByOne playerStats GameID %d, Mid %d Can't found", game.GameId, player.Mid)
			continue
		}

		var scoreTotal int64
		var commentTotal CommentType
		for _, frame := range frames {
			deltaMin, deltaMax := 0.0, 0.0
			for _, st := range stats {
				gameTime := st.Ts - gameGap - game.Stime
				if !(gameTime >= frame-s.c.Cfg.Boundary && gameTime <= frame+s.c.Cfg.Boundary) {
					continue
				}

				ex, ok := smoothExamples[gameTime]
				if !ok {
					log.Warn("日志报警 gameStatByOne smoothExamples %d GameID %d, Mid %d Can't found", gameTime, game.GameId, player.Mid)
					continue
				}

				deltaMin += math.Min(st.Acc, ex.Acc)
				deltaMax += math.Max(st.Acc, ex.Acc)
			}

			deviation := 0.0
			if deltaMax == 0.0 { // 对标准数据缺失进行兼容
				deviation = 0.0
			} else {
				deviation = deltaMin / deltaMax
			}

			// score是累加，comment同周期内只取最后一个
			comment, score := s.judge(deviation, len(keyFrames))
			commentTotal = comment
			scoreTotal += score

			// combo是连续逻辑
			if commentTotal == _commentPerfect { // 累加combo
				preCombos[player.Mid] += 1
			} else { // combo中断
				preCombos[player.Mid] = 0
			}
		}

		playerCombos = append(playerCombos, model.PlayerCombo{
			Mid:   player.Mid,
			Combo: preCombos[player.Mid],
		})
		playerScore = append(playerScore, model.PlayerHonor{
			Mid:   player.Mid,
			Score: scoreTotal,
		})
		playerComment = append(playerComment, model.PlayerComment{
			Mid:     player.Mid,
			Comment: string(commentTotal),
		})
	}

	log.Warn("gameStatByOne GameID %d, Combos %v, Score %v, Comment %v", game.GameId, playerCombos, playerScore, playerComment)

	curGame, err := s.dao.CacheGame(c, game.GameId)
	if err != nil {
		log.Error("gameStatByOne gameId(%d) cacheGame err(%v)", game.GameId, err)
		return
	}
	if curGame.Status != model.GamePlaying {
		log.Error("gameStatByOne gameId(%d) status(%s)", game.GameId, curGame.Status)
		return
	}
	// 【G3】8. 结算数据回填redis
	eg = errgroup.WithContext(c)
	eg.Go(func(c context.Context) error {
		return s.dao.AddCacheComments(c, game.GameId, playerComment)
	})
	eg.Go(func(c context.Context) error {
		return s.dao.AddCachePlayers(c, game.GameId, playerScore)
	})
	eg.Go(func(c context.Context) error {
		return s.dao.AddCachePlayersCombo(c, game.GameId, playerCombos)
	})
	if err := eg.Wait(); err != nil {
		log.Error("日志报警 gameStatByOne AddCachePlayerData GameID %d Err %v", game.GameId, err)
		return
	}

	for _, player := range players {
		p := player
		if err := s.fanout.Do(c, func(c context.Context) {

			// 删除right-delay-boundary之前的值
			_ = s.dao.DelUnusedStats(c, game.GameId, p.Mid)
		}); err != nil {
			log.Error("日志报警 gameStatByOne Fanout GameID %d Err %v", game.GameId, err)
			return
		}
	}
}

func minAndMax(frames []int64) (min, max int64) {
	min, max = math.MaxInt64, math.MinInt64
	for _, v := range frames {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return
}

func (s *Service) judge(deviation float64, lenKeyFrames int) (CommentType, int64) {

	DevCfg := s.c.Cfg.Deviation
	ScoreCfg := s.c.Cfg.Score
	comment, score := func() (CommentType, int64) {
		if deviation > DevCfg.Perfect {
			return _commentPerfect, ScoreCfg.Perfect
		}
		if deviation > DevCfg.Super {
			return _commentSuper, ScoreCfg.Super
		}
		if deviation > DevCfg.Good {
			return _commentGood, ScoreCfg.Good
		}
		if deviation > DevCfg.Bad {
			return _commentBad, ScoreCfg.Bad
		}
		return _commentMiss, ScoreCfg.Miss
	}()

	// 非归一化逻辑下，score根据deviation得到
	if !s.c.Cfg.Normalization {
		score = int64(math.Floor(deviation * 100))
	}

	// 按照总分和帧数换算单次最高得分
	score = int64(math.Floor(s.c.Cfg.MaxScore / float64(lenKeyFrames) * float64(score) / float64(ScoreCfg.Perfect)))
	return comment, score
}
