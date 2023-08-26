package dao

import (
	"context"
	"go-common/library/log"
	"sync"

	"go-common/component/metadata/device"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-card/interface/model/card/story"

	contractgrpc "git.bilibili.co/bapis/bapis-go/community/service/contract"
)

// 老粉卡片配置
func (d *dao) ContractShowConfig(ctx context.Context, aids []int64, mid int64) (map[int64]*story.ContractResource, error) {
	res := make(map[int64]*story.ContractResource)
	mu := sync.Mutex{}
	reqCommon := &contractgrpc.CommonReq{}
	if dev, ok := device.FromContext(ctx); ok {
		reqCommon = &contractgrpc.CommonReq{
			Platform: dev.RawPlatform,
			Build:    int32(dev.Build),
			Buvid:    dev.Buvid,
			MobiApp:  dev.MobiApp(),
			Device:   dev.Device,
			Ip:       metadata.RemoteIP,
		}
	}
	eg := errgroup.WithContext(ctx)
	for i := range aids {
		aid := aids[i]
		eg.Go(func(ctx context.Context) error {
			req := &contractgrpc.ShowConfigReq{
				Mid:    mid,
				Aid:    aid,
				Source: contractgrpc.ShowConfigReq_STORY,
				Common: reqCommon,
			}
			reply, err := d.contractClient.ShowConfig(ctx, req)
			if err != nil {
				log.Error("d.ContractShowConfig req:%+v, err:%v", req, err)
				return nil
			}
			if reply.IsFollowDisplay == 0 && reply.IsInteractDisplay == 0 && reply.IsTripleDisplay == 0 {
				return nil
			}
			mu.Lock()
			res[aid] = &story.ContractResource{
				IsFollowDisplay:   reply.IsFollowDisplay,
				IsInteractDisplay: reply.IsInteractDisplay,
				IsTripleDisplay:   reply.IsTripleDisplay,
				ContractCard: &story.ContractCard{
					Title:    `成为UP主的“老粉”`,
					SubTitle: "助力UP成长，让更多人发现TA",
				},
			}
			mu.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}
