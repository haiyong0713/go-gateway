package source

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

const (
	// maxArchiveSnapshotLength 一次获取稿件信息的数量
	maxArchiveSnapshotLength = 1000
	// archiveInfoChannelLength Length
	archiveSnapshotChannelLength = 2
	// concurrencyArchiveInfo 稿件信息服务并发量
	concurrencyArchiveSnapshotInfo = 2
)

// ArchiveSnapshotInfo 快照稿件信息
func (s *Service) ArchiveSnapshotInfo(c context.Context, id, batch int64, attributeType int, aids []int64) (map[int64]*rankmdl.Snapshot, error) {
	eg := errgroup.WithContext(c)
	archiveInfo := make(map[int64]*rankmdl.Snapshot)
	channel := make(chan []*rankmdl.Snapshot, archiveSnapshotChannelLength)
	eg.Go(func(ctx context.Context) error {
		return s.archiveSnapshotInfoIntoChannel(c, id, batch, attributeType, aids, channel)
	})
	eg.Go(func(ctx context.Context) (err error) {
		archiveInfo, err = s.archiveSnapshotOutChannel(c, channel)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.ArchiveSnapshotInfo")
		return nil, err
	}
	return archiveInfo, nil
}

func (s *Service) archiveSnapshotInfoIntoChannel(c context.Context, id, rankBatch int64, attributeType int, aids []int64, channel chan []*rankmdl.Snapshot) error {
	var times int
	patch := maxArchiveSnapshotLength
	concurrency := concurrencyArchiveSnapshotInfo
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
					reply, err := s.rankDao.AllSnapshotByAids(c, id, reqAids, rankBatch, attributeType)
					if err != nil {
						log.Error("s.arcClient.Arcs: error(%v)", err)
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

func (s *Service) archiveSnapshotOutChannel(c context.Context, channel chan []*rankmdl.Snapshot) (res map[int64]*rankmdl.Snapshot, err error) {
	archiveInfo := make(map[int64]*rankmdl.Snapshot)
	for v := range channel {
		if v != nil {
			for _, arc := range v {

				archiveInfo[arc.AID] = arc
			}
		}
	}
	return archiveInfo, nil
}

// ArchiveSnapshotInfoByAids 快照稿件信息
func (s *Service) ArchiveSnapshotInfoByAids(c context.Context, id, batch int64, attributeType int, aids []int64) (map[int64]*rankmdl.Snapshot, error) {
	eg := errgroup.WithContext(c)
	archiveInfo := make(map[int64]*rankmdl.Snapshot)
	channel := make(chan []*rankmdl.Snapshot, archiveSnapshotChannelLength)
	eg.Go(func(ctx context.Context) error {
		return s.archiveSnapshotInfoByAidsIntoChannel(c, id, batch, attributeType, aids, channel)
	})
	eg.Go(func(ctx context.Context) (err error) {
		archiveInfo, err = s.archiveSnapshotOutChannel(c, channel)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.ArchiveSnapshotInfo")
		return nil, err
	}
	return archiveInfo, nil
}

func (s *Service) archiveSnapshotInfoByAidsIntoChannel(c context.Context, id, rankBatch int64, attributeType int, aids []int64, channel chan []*rankmdl.Snapshot) error {
	var times int
	patch := maxArchiveSnapshotLength
	concurrency := concurrencyArchiveSnapshotInfo
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
					reply, err := s.rankDao.SnapshotByAllAids(c, id, reqAids, rankBatch, attributeType)
					if err != nil {
						log.Error("s.arcClient.Arcs: error(%v)", err)
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
