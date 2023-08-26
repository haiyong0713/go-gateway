package service

import (
	"context"
	"strconv"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/web/interface/model"

	accApi "git.bilibili.co/bapis/bapis-go/account/service"

	dmgrpc "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	votegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
)

func (s *Service) DmVote(ctx context.Context, req *model.DmVoteReq) (*model.DmVoteReply, error) {
	var (
		vote *votegrpc.DoVoteRsp
		card *accApi.Card
	)
	g := errgroup.WithCancel(ctx)
	g.Go(func(ctx context.Context) error {
		param := &votegrpc.DoVoteReq{
			VoteId:   req.VoteID,
			Votes:    []int32{req.Vote},
			VoterUid: req.Mid,
		}
		var err error
		vote, err = s.dao.DoVote(ctx, param)
		return err
	})
	g.Go(func(ctx context.Context) error {
		var err error
		card, err = s.dao.Card3(ctx, req.Mid)
		if err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	reply := &model.DmVoteReply{
		Vote: &model.VoteReply{
			UID:  vote.GetUid(),
			Type: vote.GetType(),
			Vote: req.Vote,
		},
	}
	if card.GetLevel() <= 0 {
		return reply, nil
	}
	param := &dmgrpc.PostByVoteReq{
		Progress: req.Progress,
		Aid:      req.AID,
		Cid:      req.CID,
		Mid:      req.Mid,
		Msg:      strconv.FormatInt(int64(req.Vote), 10),
		Platform: "web",
		Buvid:    req.Buvid,
	}
	dmReply, err := s.dao.PostByVote(ctx, param)
	if err != nil {
		log.Error("%+v", err)
		return reply, nil
	}
	reply.Dm = &model.DmReply{
		DmID:      dmReply.GetDmid(),
		DmIDStr:   dmReply.GetDmidStr(),
		Visible:   dmReply.GetVisible(),
		Action:    dmReply.GetAction(),
		Animation: dmReply.GetAnimation(),
	}
	return reply, nil
}
