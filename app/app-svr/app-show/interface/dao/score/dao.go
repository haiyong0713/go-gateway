package score

import (
	"context"

	scoregrpc "git.bilibili.co/bapis/bapis-go/community/service/score"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-show/interface/conf"
)

type Dao struct {
	client scoregrpc.ScoreClient
}

func NewDao(cfg *conf.Config) *Dao {
	client, err := scoregrpc.NewClient(cfg.ScoreGRPC)
	if err != nil {
		panic(err)
	}
	return &Dao{client: client}
}

func (d *Dao) MultiGetTargetScore(c context.Context, req *scoregrpc.MultiGetTargetScoreReq) (*scoregrpc.MultiGetTargetScoreResp, error) {
	rly, err := d.client.MultiGetTargetScore(c, req)
	if err != nil {
		log.Error("Fail to request scoregrpc.MultiGetTargetScore, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}
