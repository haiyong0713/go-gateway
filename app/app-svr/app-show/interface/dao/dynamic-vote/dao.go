package dynamic_vote

import (
	"context"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-show/interface/conf"
)

type Dao struct {
	client dynvotegrpc.VoteSvrClient
}

func NewDao(cfg *conf.Config) *Dao {
	client, err := dynvotegrpc.NewClient(cfg.DynvoteGRPC)
	if err != nil {
		panic(err)
	}
	return &Dao{client: client}
}

func (d *Dao) ListFeedVotes(c context.Context, req *dynvotegrpc.ListFeedVotesReq) (*dynvotegrpc.ListFeedVotesRsp, error) {
	rly, err := d.client.ListFeedVotes(c, req)
	if err != nil {
		log.Errorc(c, "Fail to request dynvotegrpc.ListFeedVotes, req=%+v error=%=v", req, err)
		return nil, err
	}
	return rly, nil
}

func (d *Dao) ListFeedVote(c context.Context, voteID, mid int64) (*dyncommongrpc.VoteInfo, error) {
	rly, err := d.ListFeedVotes(c, &dynvotegrpc.ListFeedVotesReq{Uid: mid, VoteIds: []int64{voteID}})
	if err != nil {
		return nil, err
	}
	return rly.GetVoteInfos()[voteID], nil
}

func (d *Dao) DoVote(c context.Context, req *dynvotegrpc.DoVoteReq) (*dynvotegrpc.DoVoteRsp, error) {
	rly, err := d.client.DoVote(c, req)
	if err != nil {
		log.Errorc(c, "Fail to request dynvotegrpc.DoVote, req=%+v error=%=v", req, err)
		return nil, err
	}
	return rly, nil
}
