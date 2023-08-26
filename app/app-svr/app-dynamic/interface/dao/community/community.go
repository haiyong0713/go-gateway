package community

import (
	"context"

	"github.com/pkg/errors"

	cmtGrpc "git.bilibili.co/bapis/bapis-go/community/interface/reply"
)

func (d *Dao) DynamicFeed(ctx context.Context, mid int64, buvid string, ids []string) (*cmtGrpc.DynamicFeedReply, error) {
	req := &cmtGrpc.DynamicFeedReq{
		Ids:   ids,
		Mid:   mid,
		Buvid: buvid,
	}
	ret, err := d.cmtGrpc.DynamicFeed(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret, nil
}

func (d *Dao) SubjectInteractionStatus(ctx context.Context, req *cmtGrpc.SubjectInteractionStatusReq) (*cmtGrpc.SubjectInteractionStatusReply, error) {
	reply, err := d.cmtGrpc.SubjectInteractionStatus(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
