package topic

import (
	"context"

	topiccommon "git.bilibili.co/bapis/bapis-go/topic/common"
	"go-gateway/app/app-svr/app-feed/interface/conf"

	topicgrpc "git.bilibili.co/bapis/bapis-go/topic/service"
)

type Dao struct {
	topicClient topicgrpc.TopicClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{}
	var err error
	if d.topicClient, err = topicgrpc.NewClient(c.TopicClient); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) TopicStory(ctx context.Context, arg *topicgrpc.VideoStoryReq) (*topicgrpc.VideoStoryRsp, error) {
	reply, err := d.topicClient.VideoStory(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) GetVideoAttachedTopics(ctx context.Context, aids []int64) (map[int64]*topicgrpc.TopicIconInfo, error) {
	reply, err := d.topicClient.BatchResourceTopics(ctx, &topicgrpc.BatchResourceTopicsReq{
		Resources: convertAidToTopicResource(aids),
		MetaData: &topiccommon.MetaDataCtrl{
			From: "topic_from_story_mode",
		},
	})
	if err != nil {
		return nil, err
	}
	out := make(map[int64]*topicgrpc.TopicIconInfo)
	for _, v := range reply.GetResTopics() {
		out[v.GetResId()] = v.GetTopic()
	}
	return out, nil
}

func convertAidToTopicResource(aids []int64) []*topicgrpc.Resource {
	out := make([]*topicgrpc.Resource, 0, len(aids))
	for _, aid := range aids {
		out = append(out, &topicgrpc.Resource{
			ResId:   aid,
			ResType: 1,
		})
	}
	return out
}
