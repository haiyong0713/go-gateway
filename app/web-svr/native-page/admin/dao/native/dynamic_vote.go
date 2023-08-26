package native

import (
	"context"

	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	"go-common/library/log"
)

func (d *Dao) ListFeedVotes(c context.Context, req *dynvotegrpc.ListFeedVotesReq) (*dynvotegrpc.ListFeedVotesRsp, error) {
	rly, err := d.dynvoteClient.ListFeedVotes(c, req)
	if err != nil {
		log.Errorc(c, "Fail to request dynvotegrpc.ListFeedVotes, req=%+v error=%=v", req, err)
		return nil, err
	}
	return rly, nil
}
