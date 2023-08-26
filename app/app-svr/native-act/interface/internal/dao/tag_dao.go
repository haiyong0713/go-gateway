package dao

import (
	"context"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

type tagDao struct {
	client taggrpc.TagRPCClient
}

func (d *tagDao) Tags(c context.Context, tids []int64, mid int64) (map[int64]*taggrpc.Tag, error) {
	req := &taggrpc.TagsReq{Tids: tids, Mid: mid}
	rly, err := d.client.Tags(c, req)
	if err != nil {
		return nil, err
	}
	return rly.Tags, nil
}
