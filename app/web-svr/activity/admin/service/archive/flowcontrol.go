package archive

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"

	flowcontrolapi "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

const (
	// aidFlowControlSize 一次获取稿件信息的数量
	aidFlowControlSize = 30
	// archiveInfoChannelLength Length
	aidFlowControlChannelLength = 2
	// aidFlowControlconcurrency 稿件信息服务并发量
	aidFlowControlconcurrency = 2
	// archiveBusinessID 稿件
	archiveBusinessID = 1
	// activityRank 排行source
	activityRank = "rank_activity"
)

// ArchiveFlowControl 稿件封禁信息
func (s *Service) ArchiveFlowControl(c context.Context, aids []int64) (map[int64]*flowcontrolapi.FlowCtlInfoReply, error) {
	eg := errgroup.WithContext(c)
	archiveInfo := make(map[int64]*flowcontrolapi.FlowCtlInfoReply)
	channel := make(chan *flowcontrolapi.FlowCtlInfosReply, aidFlowControlChannelLength)
	eg.Go(func(ctx context.Context) error {
		return s.archiveFlowControlIntoChannel(c, aids, channel)
	})
	eg.Go(func(ctx context.Context) (err error) {
		archiveInfo, err = s.archiveFlowControlOutChannel(c, channel)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.archiveInfoOutChannel")
		return nil, err
	}

	return archiveInfo, nil
}

func (s *Service) archiveFlowControlIntoChannel(c context.Context, aids []int64, channel chan *flowcontrolapi.FlowCtlInfosReply) error {
	var times int
	patch := aidFlowControlSize
	concurrency := aidFlowControlconcurrency
	times = len(aids) / patch / concurrency
	defer close(channel)
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			i := index
			b := batch
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(aids) {
					return nil
				}
				reqAids := aids[start:]
				end := start + patch
				if end < len(aids) {
					reqAids = aids[start:end]
				}
				if len(reqAids) > 0 {
					reply, err := s.flowcontrolClient.Infos(c, &flowcontrolapi.FlowCtlInfosReq{Oids: reqAids, BusinessId: archiveBusinessID, Source: activityRank})
					if err != nil {
						log.Error("s.ArcClient.Arcs: error(%v)", err)
						return err
					}
					channel <- reply
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Error("eg.Wait error(%v)", err)
			return ecode.ActivityRemixArchiveInfoErr
		}
	}
	return nil
}

func (s *Service) archiveFlowControlOutChannel(c context.Context, channel chan *flowcontrolapi.FlowCtlInfosReply) (res map[int64]*flowcontrolapi.FlowCtlInfoReply, err error) {
	archiveInfo := make(map[int64]*flowcontrolapi.FlowCtlInfoReply)
	for v := range channel {
		if v == nil || v.ForbiddenItemMap == nil {
			continue
		}
		for aid, arc := range v.ForbiddenItemMap {
			if arc == nil {
				continue
			}
			archiveInfo[aid] = arc
		}
	}
	return archiveInfo, nil
}
