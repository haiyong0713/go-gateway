package dao

import (
	"context"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

type tagDao struct {
	tagGRPC taggrpc.TagRPCClient
}

func (d *tagDao) Tags(ctx context.Context, mid int64, tids []int64) (map[int64]*taggrpc.Tag, error) {
	res, err := d.tagGRPC.Tags(ctx, &taggrpc.TagsReq{Mid: mid, Tids: tids})
	if err != nil {
		return nil, err
	}
	return res.Tags, nil
}
