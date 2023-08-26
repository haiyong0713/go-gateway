package service

import (
	"context"
	"math"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

func (s *Service) loadExamples(c context.Context) error {
	aids, err := s.dao.AllAids(c)
	if err != nil {
		return err
	}
	if len(aids) == 0 {
		log.Warn("loadExamples No Aids to load")
		return nil
	}

	tmpMaps := make(map[int64]map[int64]float64)
	mutex := new(sync.Mutex)

	g := errgroup.WithContext(context.Background())
	for _, v := range aids {
		aid := v
		g.Go(func(c context.Context) error {
			exs, err := s.dao.PickExamples(c, aid)

			if err != nil {
				return err
			}

			lastPos := int64(0)
			lastEucli := float64(0)
			myMap := make(map[int64]float64)

			for _, stat := range exs {

				currentPos := stat.TS
				currentEucli := stat.Euclidean()

				step := (currentEucli - lastEucli) / float64(currentPos-lastPos) // 求模长平滑后的步长
				for i := lastPos; i < currentPos; i++ {                          // 逐秒平滑
					myMap[i] = model.Round(lastEucli+float64(i-lastPos)*step, 5)
				}

				lastPos = currentPos
				lastEucli = currentEucli
			}
			myMap[lastPos] = lastEucli // 最后一条数据补全

			mutex.Lock()
			tmpMaps[aid] = myMap
			mutex.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("loadExamples groups err %v", err)
		return err
	}

	s.arcExamples = tmpMaps
	return nil
}

func between(value, min, max int64) bool {
	return value >= min && value <= max
}

var normScore = map[CommentType]int64{
	_commentPerfect: 100,
	_commentGood:    70,
	_commentOk:      30,
	_commentMiss:    0,
}

func (s *Service) judge(deviation float64) (string, int64) {
	// 求评价
	comment := _commentMiss
	func() {
		if deviation > s.standardAcc[_commentPerfect] {
			comment = _commentPerfect
			return
		}
		if deviation > s.standardAcc[_commentGood] {
			comment = _commentGood
			return
		}
		if deviation > s.standardAcc[_commentOk] {
			comment = _commentOk
			return
		}
	}()

	// 求得分，最高一百
	score := int64(math.Ceil(deviation * 100))
	if s.conf.Normalizing {
		score = normScore[comment]
	}
	return string(comment), score
}

func (s *Service) GameStat(c context.Context, buvid string, gameID, mid int64, stats []*model.Stat) error {
	if len(stats) == 0 {
		log.Error("GameStat GameID %d Stats Empty", gameID)
		return ecode.RequestErr
	}

	game, err := s.dao.RawGame(c, gameID)
	if err != nil {
		log.Error("日志报警 GameStat GameID %d Err %v", gameID, err)
		return err
	}

	// 游戏未开始不采集数据
	if game.Status != model.GamePlaying {
		log.Warn("GameStat GameID %d Not Started", gameID)
		return nil
	}

	for _, v := range stats { // 收敛3位小数
		v.TreatFloat(game.Stime)
	}

	log.Warn("GameStat GameID %d, Ready To Gather? example device %v, got %s", gameID, s.conf.ExampleDevice, buvid)
	if exs := s.conf.ExampleDevice; len(exs) != 0 { // 收集数据
		if hit := func() bool {
			for _, v := range s.conf.ExampleDevice {
				if v == buvid {
					return true
				}
			}
			return false
		}(); hit {
			log.Warn("GameStat GameID %d Aid %d Mid %d Buvid %s Begin to Gather Examples", gameID, game.AID, mid)
			return s.dao.GatherExamples(c, game.AID, stats)
		}
	}

	if mid == 0 { // 上报数据必须要mid
		log.Error("日志报警 GameStat GameID %d Mid %d Not found", gameID, mid)
		return ecode.RequestErr
	}

	examples, ok := s.arcExamples[game.AID]
	if !ok {
		log.Error("日志报警 GameStat GameID %d AID %d, Example not found", gameID, game.AID)
		return ecode.ServiceUnavailable
	}

	keyFrames, ok := s.arcKeyFrames[game.AID]
	if !ok {
		log.Error("日志报警 GameStat GameID %d AID %d, KeyFrames not found", gameID, game.AID)
		return ecode.ServiceUnavailable
	}

	requestTs := stats[0].TS

	hit, keyTs := func() (bool, []int64) { // 是否命中关键帧
		var res []int64
		for _, v := range keyFrames {
			if v > requestTs && v <= requestTs+500 {
				res = append(res, v)
			}
		}
		if len(res) > 0 {
			return true, res
		}
		return false, nil
	}()

	if !hit { // 不是关键帧，不触发评分
		log.Info("GameStat RequestTS %d is not key frame", requestTs)
		return nil
	}

	// 先简单实现，每个动作点对比，求方差
	var (
		statMap     = make(map[int64]float64)
		scoreSum    int64
		commentLast string // 只取最后一次comment
	)
	for _, ts := range keyTs {
		for _, v := range stats {
			if !between(v.TS, ts-s.conf.Boundary, ts+s.conf.Boundary) {
				continue
			}
			statMap[v.TS] = v.Euclidean()
		}
		deltaSum := 0.0
		okCount := 0
		for ts, v := range statMap {
			ex, ok := examples[ts]
			if !ok {
				log.Error("日志报警 GameStat GameID %d AID %d Ts %d, Example not found", gameID, game.AID, ts)
				continue
			}
			okCount++
			deltaSum += math.Min(v, ex) / math.Max(v, ex)
		}

		deviation := 0.0
		if deltaSum == 0.0 { // 对标准数据缺失进行兼容
			deviation = 0.0
		} else {
			deviation = deltaSum / float64(okCount)
		}
		comment, score := s.judge(deviation)
		// 根据keyFrames的数量动态调整得分，使得每场游戏最高分都是5000分
		score = score * s.conf.MaxScore / (100 * int64(len(keyFrames)))
		scoreSum += score
		commentLast = comment
	}
	log.Warn("GameStat GameID %d AID %d Mid %d Ts %d, Comment %s, Score %d, cnt %d", gameID, game.AID, mid, requestTs, commentLast, scoreSum, len(statMap))

	// 增加redis中的得分，err忽略
	if err := s.dao.RedisIncrPoint(c, gameID, mid, scoreSum); err != nil {
		log.Error("日志报警 GameStat GameID %d AID %d Ts %d, RedisIncrPoint Err %v", gameID, game.AID, requestTs, err)
	}
	// 增加redis中的comment
	if err := s.dao.RedisSetComment(c, gameID, mid, commentLast); err != nil {
		log.Error("日志报警 GameStat GameID %d AID %d Ts %d, RedisSetComment Err %v", gameID, game.AID, requestTs, err)
	}

	return nil
}

func (s *Service) ArcTopRanks(c context.Context, aid int64, number int64) ([]*model.PlayerRank, error) {
	playerHonors, err := s.dao.RedisGetUserPoints(c, aid, number)
	if err != nil {
		return nil, err
	}
	if len(playerHonors) == 0 { // 当日无排行榜，返回空数据
		return []*model.PlayerRank{}, nil
	}

	mids := make([]int64, 0)
	for _, v := range playerHonors {
		mids = append(mids, v.Mid)
	}
	users, err := s.users(c, mids)
	if err != nil {
		return nil, err
	}
	res := make([]*model.PlayerRank, 0)
	for _, v := range playerHonors {
		u, ok := users[v.Mid]
		if !ok {
			continue
		}
		r := &model.PlayerRank{
			PlayerHonor: *v,
			Face:        u.Face,
			Name:        u.Name,
		}
		res = append(res, r)
	}
	return res, nil
}
