package bplus

import (
	"context"

	"go-common/library/log"

	dynccommon "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dynctopic "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
)

func (d *Dao) TopicStats(ctx context.Context, ids []int64) (map[int64]*dynccommon.TopicStats, error) {
	reply, err := d.topicClient.BatchGetStats(ctx, &dynctopic.BatchGetStatsReq{TopicIds: ids})
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return reply.GetStats(), nil
}
