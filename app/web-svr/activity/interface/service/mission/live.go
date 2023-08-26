package mission

import (
	"context"
	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/activity/interface/client"
	model "go-gateway/app/web-svr/activity/interface/model/mission"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	live "git.bilibili.co/bapis/bapis-go/live/xroom"
)

const (
	networkStateWIFI = 2
	networkStateWWAN = 1
)

func (s *Service) FetchLiveRoomByOperId(ctx context.Context, operSourceDataId int64, networkState int64, bizName string) (res []*model.LiveRoomInfo, err error) {
	operConfig, err := s.dao.GetRoomIdsByOperSourceId(ctx, operSourceDataId)
	if err != nil {
		log.Errorc(ctx, "dao.GetRoomIdsByOperSourceId:%v, err: %v", operSourceDataId, err)
		return
	}
	return s.fetchLiveRoomByRoomIds(ctx, operConfig.EntryFrom, operConfig.RoomIds, networkState, bizName)
}

func (s *Service) fetchLiveRoomByRoomIds(ctx context.Context, entryFrom string, roomIds []int64, networkState int64, bizName string) (res []*model.LiveRoomInfo, err error) {
	platform := ""
	build := int64(0)
	deviceName := ""
	userIp := metadata.String(ctx, metadata.RemoteIP)
	userMid := metadata.Int64(ctx, metadata.Mid)
	d, ok := device.FromContext(ctx)
	if ok {
		platform = d.RawPlatform
		build = d.Build
		deviceName = d.Device
	}
	network := ""
	switch networkState {
	case networkStateWIFI:
		network = "wifi"
	case networkStateWWAN:
		network = "mobile"
	default:
		network = "other"
	}
	rpcRes, err := client.LiveClient.EntryRoomInfo(ctx, &live.EntryRoomInfoReq{
		EntryFrom:  []string{entryFrom},
		RoomIds:    roomIds,
		Uid:        userMid,
		Uipstr:     userIp,
		Platform:   platform,
		Build:      build,
		DeviceName: deviceName,
		Network:    network,
		ReqBiz:     bizName,
	})
	if err != nil {
		log.Errorc(ctx, "call LiveClient.EntryRoomInfo error: %v", err)
		return
	}
	mids := make([]int64, 0, len(rpcRes.List))
	for _, roomId := range roomIds {
		tmpRoom := rpcRes.List[roomId]
		if tmpRoom != nil {
			mids = append(mids, tmpRoom.Uid)
			res = append(res, &model.LiveRoomInfo{
				RoomId:       tmpRoom.RoomId,
				RoomMid:      tmpRoom.Uid,
				RoomTitle:    tmpRoom.Title,
				RoomCover:    tmpRoom.Cover,
				JumpUrl:      tmpRoom.JumpUrl[entryFrom],
				RoomKeyFrame: tmpRoom.Keyframe,
				RoomStatus:   tmpRoom.LiveStatus,
				Online:       tmpRoom.PopularityCount,
			})
		}
	}
	accRes, err := client.AccountClient.Infos3(ctx, &accapi.MidsReq{
		Mids:   mids,
		RealIp: "", //TODO
	})
	if err != nil {
		log.Errorc(ctx, "call AccountClient.Infos3 error: %v", err)
		return
	}
	for _, room := range res {
		userInfo := accRes.Infos[room.RoomMid]
		if userInfo != nil {
			room.RoomUserName = userInfo.Name
			room.RoomUserAvatar = userInfo.Face
		}
	}

	return
}
