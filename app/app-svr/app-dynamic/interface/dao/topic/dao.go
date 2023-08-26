package topic

import (
	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	dyntopic "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	topicV2 "git.bilibili.co/bapis/bapis-go/topic/service"
)

type Dao struct {
	c *conf.Config
	// 老版本的动态话题 粉板6.40之前的版本使用
	dynTopic dyntopic.TopicClient
	// 新话题服务 粉板6.40之后用
	topicV2 topicV2.TopicClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.dynTopic, err = dyntopic.NewClient(c.TopicGRPC); err != nil {
		panic(err)
	}
	if d.topicV2, err = topicV2.NewClient(c.TopicGRPC); err != nil {
		panic(err)
	}
	return d
}
