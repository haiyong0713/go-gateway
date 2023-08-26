package dao

import (
	"context"

	tagGRPC "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

func (d *Dao) GetTagsByAID(aid int64) (tags []*tagGRPC.Tag, err error) {
	ctx := context.Background()
	tags = make([]*tagGRPC.Tag, 0)
	req := &tagGRPC.ArcTagsReq{
		Aid: aid,
	}
	var reply *tagGRPC.ArcTagsReply
	if reply, err = d.tagGRPCClient.ArcTags(ctx, req); err != nil {
		return
	}
	if reply != nil {
		tags = reply.Tags
	}
	return
}
