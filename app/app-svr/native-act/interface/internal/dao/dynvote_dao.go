package dao

import (
	"context"

	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
)

type dynvoteDao struct {
	client dynvotegrpc.VoteSvrClient
}

func (d *dynvoteDao) ListFeedVotes(c context.Context, req *dynvotegrpc.ListFeedVotesReq) (*dynvotegrpc.ListFeedVotesRsp, error) {
	rly, err := d.client.ListFeedVotes(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *dynvoteDao) DoVote(c context.Context, req *dynvotegrpc.DoVoteReq) (*dynvotegrpc.DoVoteRsp, error) {
	rly, err := d.client.DoVote(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
