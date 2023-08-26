package common

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/live"
)

// LiveRooms .
func (s *Service) LiveRooms(c context.Context, ids []int64) (rooms map[int64]*live.Room, err error) {
	if rooms, err = s.liveDao.LiveRoom(c, ids); err != nil {
		log.Error("common.Lives param(%v)error %v", ids, err)
		return
	}
	if len(rooms) == 0 {
		err = fmt.Errorf("id错误，没有直播相关信息！")
		return
	}
	return
}
