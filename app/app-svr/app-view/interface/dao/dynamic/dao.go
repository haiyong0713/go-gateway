package dynamic

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-view/interface/conf"

	votegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
)

type Dao struct {
	voteClient votegrpc.VoteSvrClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.voteClient, err = votegrpc.NewClient(c.VoteClient); err != nil {
		panic(fmt.Sprintf("VoteClient not found err(%v)", err))
	}
	return
}

func (d *Dao) Vote(ctx context.Context, param *votegrpc.DoVoteReq) (*votegrpc.DoVoteRsp, error) {
	reply, err := d.voteClient.DoVote(ctx, param)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
