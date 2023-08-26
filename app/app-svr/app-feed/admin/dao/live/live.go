package live

import (
	"context"
	"fmt"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-feed/admin/model/live"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const room = "/room/v1/Room/rooms_for_app_player"

// LiveRoom .
func (d *Dao) LiveRoom(c context.Context, roomids []int64) (data map[int64]*live.Room, err error) {
	params := url.Values{}
	params.Set("room_ids", xstr.JoinInts(roomids))
	res := new(live.Rooms)
	if err = d.liveHTTPClient.Get(c, d.c.Host.Live+room, "", params, &res); err != nil {
		log.Error("LiveRoom Req(%v) error(%v) res(%+v)", roomids, err, res)
		return nil, fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.c.Host.Live+room+"?"+params.Encode(), err.Error())
	}
	if res.Code != ecode.OK.Code() {
		log.Error("LiveRoom Req(%d) error(%v) res(%+v)", roomids, err, res)
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.c.UserFeed.Live, d.c.Host.Live+room+"?"+params.Encode())
	}
	if res.Data == nil {
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.c.UserFeed.Live, d.c.Host.Live+room+"?"+params.Encode())
	}
	data = res.Data
	return
}
