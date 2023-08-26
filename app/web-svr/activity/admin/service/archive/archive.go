package archive

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

const (
	// maxArchiveInfoLength 一次获取稿件信息的数量
	maxArchiveInfoLength = 50
	// archiveInfoChannelLength Length
	archiveInfoChannelLength = 2
	// concurrencyArchiveInfo 稿件信息服务并发量
	concurrencyArchiveInfo = 2
)

// ArchiveInfo 用户信息
func (s *Service) ArchiveInfo(c context.Context, aids []int64) (map[int64]*api.Arc, error) {
	eg := errgroup.WithContext(c)
	archiveInfo := make(map[int64]*api.Arc)
	channel := make(chan *api.ArcsReply, archiveInfoChannelLength)
	eg.Go(func(ctx context.Context) error {
		return s.archiveInfoIntoChannel(c, aids, channel)
	})
	eg.Go(func(ctx context.Context) (err error) {
		archiveInfo, err = s.archiveInfoOutChannel(c, channel)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.archiveInfoOutChannel")
		return nil, err
	}

	return archiveInfo, nil
}

func (s *Service) archiveInfoIntoChannel(c context.Context, aids []int64, channel chan *api.ArcsReply) error {
	var times int
	patch := maxArchiveInfoLength
	concurrency := concurrencyArchiveInfo
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
					reply, err := s.ArcClient.Arcs(c, &api.ArcsRequest{Aids: reqAids})
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

func (s *Service) archiveInfoOutChannel(c context.Context, channel chan *api.ArcsReply) (res map[int64]*api.Arc, err error) {
	archiveInfo := make(map[int64]*api.Arc)
	for v := range channel {
		if v == nil || v.Arcs == nil {
			continue
		}
		for _, arc := range v.Arcs {
			if arc == nil {
				continue
			}
			if arc.IsNormal() {
				archiveInfo[arc.Aid] = arc
			}
		}
	}
	return archiveInfo, nil
}
