package account

import (
	"context"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"

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

// MemberInfo 用户信息
func (s *Service) MemberInfo(c context.Context, mids []int64) (map[int64]*accountapi.Info, error) {
	eg := errgroup.WithContext(c)
	midsInfo := make(map[int64]*accountapi.Info)
	channel := make(chan map[int64]*accountapi.Info, memberInfoChannelLength)
	eg.Go(func(ctx context.Context) error {
		return s.memberInfoIntoChannel(c, mids, channel)
	})
	eg.Go(func(ctx context.Context) (err error) {
		midsInfo, err = s.memberInfoOutChannel(c, channel)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.memberInfo")
		return nil, err
	}

	return midsInfo, nil
}

func (s *Service) memberInfoIntoChannel(c context.Context, mids []int64, channel chan map[int64]*accountapi.Info) error {
	var times int
	patch := maxMemberInfoLength
	concurrency := concurrencyMemberInfo
	times = len(mids) / patch / concurrency
	defer close(channel)
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			i := index
			b := batch
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(mids) {
					return nil
				}
				reqMids := mids[start:]
				end := start + patch
				if end < len(mids) {
					reqMids = mids[start:end]
				}
				if len(reqMids) > 0 {
					infosReply, err := s.AccClient.Infos3(ctx, &accountapi.MidsReq{Mids: reqMids})
					if err != nil || infosReply == nil {
						log.Error("s.AccClient.Infos3: error(%v) batch(%d)", err, i)
						return err
					}
					channel <- infosReply.Infos
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

func (s *Service) memberInfoOutChannel(c context.Context, channel chan map[int64]*accountapi.Info) (map[int64]*accountapi.Info, error) {
	midsInfo := make(map[int64]*accountapi.Info)
	for item := range channel {
		for mid, value := range item {
			if value != nil {
				midsInfo[mid] = value
			}
		}
	}
	return midsInfo, nil
}
