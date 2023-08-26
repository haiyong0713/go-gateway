package dao

import (
	"context"

	votegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
)

func (d *Dao) DoVote(ctx context.Context, req *votegrpc.DoVoteReq) (*votegrpc.DoVoteRsp, error) {
	reply, err := d.voteClient.DoVote(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
