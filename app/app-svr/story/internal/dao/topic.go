package dao

import (
	"context"

	topicgrpc "git.bilibili.co/bapis/bapis-go/topic/service"
)

func (d *dao) TopicStory(ctx context.Context, arg *topicgrpc.VideoStoryReq) (*topicgrpc.VideoStoryRsp, error) {
	reply, err := d.topicClient.VideoStory(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
