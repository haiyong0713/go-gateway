package space

import (
	"context"

	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"

	contractgrpc "git.bilibili.co/bapis/bapis-go/community/service/contract"
)

func (s *Service) getContractResource(ctx context.Context, mid, vmid int64) (*space.ContractResource, bool, error) {
	args := &contractgrpc.ShowConfigReq{
		Mid:    mid,
		UpMid:  vmid,
		Source: contractgrpc.ShowConfigReq_SPACE,
		Common: &contractgrpc.CommonReq{},
	}
	if dev, ok := device.FromContext(ctx); ok {
		args.Common = &contractgrpc.CommonReq{
			Platform: dev.RawPlatform,
			Build:    int32(dev.Build),
			Buvid:    dev.Buvid,
			MobiApp:  dev.MobiApp(),
			Device:   dev.Device,
			Ip:       metadata.RemoteIP,
		}
	}
	reply, err := s.commDao.ContractShowConfig(ctx, args)
	if err != nil {
		log.Error("s.commDao.ContractShowConfig args=%+v error=%+v", args, err)
		return nil, false, err
	}
	if reply.IsFollowDisplay == 0 {
		return nil, false, nil
	}
	return &space.ContractResource{
		FollowShowType: reply.FollowShowType,
		ContractCard: &space.ContractCard{
			Title:    "你触发了神秘契约",
			SubTitle: "成为契约者，见证UP主的成长",
			Icon:     "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/3XtZJ8DWTR.png",
		},
		FollowButtonDecorate: &space.FollowButtonDecorate{
			WingLeft:  "https://i0.hdslb.com/bfs/activity-plat/static/2be2c5f696186bad80d4b452e4af2a76/oSdCy82Lq2.png",
			WingRight: "https://i0.hdslb.com/bfs/activity-plat/static/2be2c5f696186bad80d4b452e4af2a76/4UOrT8gNsT.png",
		},
	}, true, nil
}
