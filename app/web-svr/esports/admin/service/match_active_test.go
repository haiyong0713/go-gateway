package service

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/esports/admin/mock"

	liveRoom "git.bilibili.co/bapis/bapis-go/live/xroom"

	"github.com/golang/mock/gomock"
)

func TestLiveRoomBiz(t *testing.T) {
	t.Run("test 0 roomID", zeroRoomID)
	t.Run("test matched live room", matchedLiveRoom)
	t.Run("no matched live room", noMatchedLiveRoom)
}

func zeroRoomID(t *testing.T) {
	if isLiveRoomValid(context.Background(), []int64{0}) {
		t.Errorf("live roomID(0) should as invalid")
	}
}

func noMatchedLiveRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	roomID := int64(8888)
	req := new(liveRoom.RoomIDsReq)
	{
		req.RoomIds = []int64{roomID}
		req.Attrs = []string{"show", "status"}
	}
	res := new(liveRoom.RoomIDsInfosResp)
	{
		m := make(map[int64]*liveRoom.Infos)
		res.List = m
	}

	mockClient := mock.NewMockRoomClient(ctrl)
	mockClient.EXPECT().GetMultiple(ctx, req).Return(res, nil)

	//liveRoomClient = mockClient
	if isLiveRoomValid(ctx, []int64{roomID}) {
		t.Errorf("liveRoomInfo should not matched")
	}
}

func matchedLiveRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	roomID := int64(8888)
	req := new(liveRoom.RoomIDsReq)
	{
		req.RoomIds = []int64{roomID}
		req.Attrs = []string{"show", "status"}
	}
	res := new(liveRoom.RoomIDsInfosResp)
	{
		m := make(map[int64]*liveRoom.Infos)
		info := new(liveRoom.Infos)
		{
			status := new(liveRoom.RoomStatusInfo)
			{
				status.LockStatus = 0
			}

			showInfo := new(liveRoom.RoomShowInfo)
			{
				showInfo.Title = "test"
			}

			info.Show = showInfo
			info.Status = status
		}
		m[roomID] = info

		res.List = m
	}

	mockClient := mock.NewMockRoomClient(ctrl)
	mockClient.EXPECT().GetMultiple(ctx, req).Return(res, nil)

	//liveRoomClient = mockClient
	if !isLiveRoomValid(ctx, []int64{roomID}) {
		t.Errorf("liveRoomInfo should matched")
	}
}
