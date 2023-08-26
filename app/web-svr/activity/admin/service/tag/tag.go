package tag

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"github.com/pkg/errors"
)

const (
	// maxMemberInfoLength 一次获取用户信息的数量
	maxMemberInfoLength = 50
	// memberInfoChannelLength Length
	memberInfoChannelLength = 2
	// concurrencyMemberInfo 用户信息服务并发量
	concurrencyMemberInfo = 2
)

// TagInfo 用户信息
func (s *Service) TagInfo(c context.Context, tagIds []int64) (map[int64]*tagrpc.Tag, error) {
	eg := errgroup.WithContext(c)
	tagIdsInfo := make(map[int64]*tagrpc.Tag)
	channel := make(chan map[int64]*tagrpc.Tag, memberInfoChannelLength)
	eg.Go(func(ctx context.Context) error {
		return s.tagIntoChannel(c, tagIds, channel)
	})
	eg.Go(func(ctx context.Context) (err error) {
		tagIdsInfo, err = s.tagInfoOutChannel(c, channel)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.tagInfo")
		return nil, err
	}

	return tagIdsInfo, nil
}

func (s *Service) tagIntoChannel(c context.Context, tagIds []int64, channel chan map[int64]*tagrpc.Tag) error {
	var times int
	patch := maxMemberInfoLength
	concurrency := concurrencyMemberInfo
	times = len(tagIds) / patch / concurrency
	defer close(channel)
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			i := index
			b := batch
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(tagIds) {
					return nil
				}
				reqTag := tagIds[start:]
				end := start + patch
				if end < len(tagIds) {
					reqTag = tagIds[start:end]
				}
				if len(reqTag) > 0 {
					tagRes, err := s.tagRPC.Tags(c, &tagrpc.TagsReq{Tids: reqTag})
					if err != nil || tagRes == nil || tagRes.Tags == nil {
						err = errors.Wrapf(err, "s.tagRPC.TagByNames")
						return err
					}
					channel <- tagRes.Tags
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Error("eg.Wait error(%v)", err)
			return ecode.ActivityWriteHandMemberInfoErr
		}
	}
	return nil
}

func (s *Service) tagInfoOutChannel(c context.Context, channel chan map[int64]*tagrpc.Tag) (map[int64]*tagrpc.Tag, error) {
	tagInfo := make(map[int64]*tagrpc.Tag)
	for item := range channel {
		for tag, value := range item {
			if value != nil {
				tagInfo[tag] = value
			}
		}
	}
	return tagInfo, nil
}
