package topic

import (
	"context"
	"strconv"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-channel/interface/conf"

	dynamicCommon "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dynamicTopic "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

type Dao struct {
	c            *conf.Config
	topicClient  dynamicTopic.TopicClient
	newTopicGRPC topicsvc.TopicClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.topicClient, err = dynamicTopic.NewClient(c.DynamicTopicGRPC); err != nil {
		panic(err)
	}
	if d.newTopicGRPC, err = topicsvc.NewClient(c.NewTopicGRPC); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) HasCreatedTopic(ctx context.Context, params *topicsvc.HasCreatedTopicReq) (*topicsvc.HasCreatedTopicRsp, error) {
	return d.newTopicGRPC.HasCreatedTopic(ctx, params)
}

func (d *Dao) HotNewTopics(ctx context.Context, params *topicsvc.HotNewTopicsReq) (*topicsvc.HotNewTopicsRsp, error) {
	return d.newTopicGRPC.HotNewTopics(ctx, params)
}

func (d *Dao) RcmdTopicsBigCard(c context.Context, mid int64, build int, platform, mobiApp, Device, fromSpmid, version, buvid string) ([]*dynamicTopic.HotListDetail, error) {
	resTmp, err := d.topicClient.RcmdTopicsBigCard(c, &dynamicTopic.RcmdTopicsBigCardReq{
		Mid: mid,
		MetaData: &dynamicCommon.CmnMetaData{
			Build:     strconv.Itoa(build),
			Platform:  platform,
			MobiApp:   mobiApp,
			Device:    Device,
			FromSpmid: fromSpmid,
			Version:   version,
			Buvid:     buvid,
		},
		TraceId: "",
	})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return resTmp.GetTopics(), nil
}
