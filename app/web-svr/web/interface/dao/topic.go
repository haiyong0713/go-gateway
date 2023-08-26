package dao

import (
	"context"

	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

func (d *Dao) BatchResTopicByType(ctx context.Context, req *topicsvc.BatchResTopicByTypeReq) (*topicsvc.BatchResTopicByTypeRsp, error) {
	return d.TopicGRPC.BatchResTopicByType(ctx, req)
}

func (d *Dao) TopicGeneralFeedList(ctx context.Context, req *topicsvc.GeneralFeedListReq) (*topicsvc.GeneralFeedListRsp, error) {
	return d.TopicGRPC.GeneralFeedList(ctx, req)
}

func (d *Dao) ListDyns(ctx context.Context, req *dyntopicgrpc.ListDynsReq) (*dyntopicgrpc.ListDynsRsp, error) {
	return d.dynTopicClient.ListDyns(ctx, req)
}
