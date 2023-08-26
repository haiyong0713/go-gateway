package article

import (
	"context"
	accountRelationGrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"github.com/pkg/errors"
	"go-common/library/ecode"
)

func (d *Dao) AccountRelationStats(ctx context.Context, mids []int64) (rsp *accountRelationGrpc.StatsReply, err error) {
	if len(mids) == 0 {
		return &accountRelationGrpc.StatsReply{}, nil
	}
	rsp, err = d.accountRelationClient.Stats(ctx, &accountRelationGrpc.MidsReq{
		Mids: mids,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "AccountRelationStats mids(%v)", mids)
	}
	if rsp == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "AccountRelationStats mids(%v)", mids)
	}
	if rsp.StatReplyMap == nil {
		return &accountRelationGrpc.StatsReply{}, nil
	}
	return rsp, nil
}
