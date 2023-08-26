package service

import (
	"context"
	"encoding/binary"
	"hash/crc32"
	"sort"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/pkg/idsafe/bvid"

	"github.com/pkg/errors"
)

func decodeScore(in int32, validation uint32) (int32, bool) {
	//nolint:gomnd
	score := in ^ 19951021 // 质数
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(in))
	return score, crc32.ChecksumIEEE(bs) == validation
}

// RankList is
func (s *Service) RankList(ctx context.Context, req *api.RankListReq) (*api.RankListReply, error) {
	func() {
		if req.Bvid == "" {
			return
		}
		aid, err := bvid.BvToAv(req.Bvid)
		if err != nil {
			log.Error("Failed to parse bvid: %q: %+v", req.Bvid, err)
			return
		}
		req.Aid = aid
	}()
	if req.Aid <= 0 {
		return nil, errors.WithStack(ecode.RequestErr)
	}
	items, err := s.dao.RankList(ctx, req)
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})
	reply := &api.RankListReply{
		List: make([]*api.RankItem, 0, len(items)),
		CurrentUser: &api.RankItem{
			Mid:    req.CurrentMid,
			Ranked: false,
		},
	}
	if len(items) <= 0 {
		if req.CurrentMid <= 0 {
			reply.CurrentUser = nil
		}
		return reply, nil
	}
	// user last score
	func() {
		lastScore, err := s.dao.GetScore(ctx, req.CurrentMid, req.Aid, req.Cid)
		if err != nil {
			log.Error("Failed to get user last score: %+v: %+v", req, err)
			return
		}
		reply.CurrentUser.Score = lastScore.Score
	}()

	mids := make([]int64, 0, len(items))
	for _, item := range items {
		mids = append(mids, item.Mid)
	}
	if req.CurrentMid > 0 {
		mids = append(mids, req.CurrentMid)
	}
	infos, err := s.accDao.Infos3(ctx, mids)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		rankItem := &api.RankItem{
			Mid:   item.Mid,
			Score: item.Score,
		}
		info, ok := infos[rankItem.Mid]
		if !ok {
			log.Warn("Failed to get user info: %d", rankItem.Mid)
			continue
		}
		rankItem.Name = info.Name
		rankItem.Face = info.Face
		rankItem.Sign = info.Sign
		reply.List = append(reply.List, rankItem)
	}
	currentUser, ok := infos[req.CurrentMid]
	if ok {
		reply.CurrentUser.Name = currentUser.Name
		reply.CurrentUser.Face = currentUser.Face
		reply.CurrentUser.Sign = currentUser.Sign
	}
	if int64(len(reply.List)) > req.Size_ {
		reply.List = reply.List[:req.Size_]
	}
	setRankStatus(reply.List)
	rankMap := makeRankListMap(reply.List)
	currentUserRank, ok := rankMap[req.CurrentMid]
	if ok {
		reply.CurrentUser = currentUserRank
	}
	return reply, nil
}

// RankScoreSubmit is
func (s *Service) RankScoreSubmit(ctx context.Context, req *api.RankScoreSubmitReq) error {
	func() {
		if req.Bvid == "" {
			return
		}
		aid, err := bvid.BvToAv(req.Bvid)
		if err != nil {
			log.Error("Failed to parse bvid: %q: %+v", req.Bvid, err)
			return
		}
		req.Aid = aid
	}()
	if req.Aid <= 0 {
		return errors.WithStack(ecode.RequestErr)
	}
	score, ok := decodeScore(req.Score, req.Validation)
	if !ok {
		return errors.WithStack(ecode.RequestErr)
	}
	req.Score = score

	lastScore, err := s.dao.GetScore(ctx, req.CurrentMid, req.Aid, req.Cid)
	if err != nil {
		return err
	}
	if req.Score <= lastScore.Score {
		log.Warn("Skip to update user score: %+v to %d", lastScore, req.Score)
		return nil
	}
	return s.dao.RankScoreUpdate(ctx, req)
}

func setRankStatus(in []*api.RankItem) {
	for i, item := range in {
		item.Ranked = true
		item.Ranking = int64(i + 1)
	}
	//nolint:gosimple
	return
}

func makeRankListMap(in []*api.RankItem) map[int64]*api.RankItem {
	out := make(map[int64]*api.RankItem, len(in))
	for _, i := range in {
		out[i.Mid] = i
	}
	return out

}
