package dao

import (
	"context"

	commscoregrpc "git.bilibili.co/bapis/bapis-go/community/service/score"
)

type commscoreDao struct {
	client commscoregrpc.ScoreClient
}

func (d *commscoreDao) MultiGetTargetScore(c context.Context, req *commscoregrpc.MultiGetTargetScoreReq) (*commscoregrpc.MultiGetTargetScoreResp, error) {
	rly, err := d.client.MultiGetTargetScore(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
