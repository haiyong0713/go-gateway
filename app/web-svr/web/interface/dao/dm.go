package dao

import (
	"context"

	dmgrpc "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	"github.com/pkg/errors"
)

func (d *Dao) PostByVote(ctx context.Context, req *dmgrpc.PostByVoteReq) (*dmgrpc.PostByVoteReply, error) {
	reply, err := d.dmClient.PostByVote(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", req)
	}
	return reply, err
}
