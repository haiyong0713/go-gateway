package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	accClient "git.bilibili.co/bapis/bapis-go/account/service"
)

func (s *Service) GameRanks(c context.Context, cid, mid int64, pn, ps int) (*model.RankReply, error) {
	res := &model.RankReply{Page: &model.Page{Pn: pn, Ps: ps}}
	ranks, err := s.ottDao.LoadRanks(c, cid, pn, ps)
	if err != nil {
		log.Error("PlayerRanks cid(%d) err(%v)", cid, err)
		return nil, err
	}
	var (
		mids   []int64
		rank   int
		score  int
		midMap map[int64]*accClient.Card
		eg     = errgroup.WithContext(c)
	)
	if mid != 0 { // 查用户的排名
		mids = append(mids, mid)
		eg.Go(func(ctx context.Context) error {
			rankMap, err := s.ottDao.CachePlayersRank(c, cid, []int64{mid})
			if err != nil {
				log.Warn("GameRanks cid(%d) mid(%d) err(%v)", cid, mid, err)
				return nil
			}
			rank = rankMap[mid]
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			if score, err = s.ottDao.CachePlayerScore(c, cid, mid); err != nil {
				log.Warn("GameRanks cid(%d) mid(%d) err(%v)", cid, mid, err)
			}
			return nil
		})
	}
	for _, player := range ranks {
		mids = append(mids, player.Mid)
	}
	if len(mids) > 0 {
		eg.Go(func(ctx context.Context) error {
			midMap, err = s.ottDao.UserCards(c, mids)
			return err
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("GameRanks err(%v)", err)
		return nil, err
	}

	var list []*model.PlayerRankV2
	for idx, player := range ranks {
		user, ok := midMap[player.Mid]
		if !ok {
			continue
		}
		playerInfo := new(model.Player)
		playerInfo.CopyFromGRPC(user)
		list = append(list, &model.PlayerRankV2{
			Player: playerInfo,
			Score:  int(player.Score),
			Rank:   pn*ps + idx + 1,
		})
	}
	if user, ok := midMap[mid]; ok {
		if rank >= 0 && score > 0 {
			player := new(model.Player)
			player.CopyFromGRPC(user)
			res.Player = &model.PlayerRankV2{
				Player: player,
				Score:  score,
				Rank:   rank + 1,
			}
		}
	}
	res.Ranks = list
	return res, nil
}
