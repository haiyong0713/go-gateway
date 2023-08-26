package mission

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	model "go-gateway/app/web-svr/activity/interface/model/mission"
	"strconv"
	"strings"
)

const (
	roomIdsCacheKey  = "mission_room_o_%v"
	roomIdsExpire    = 15
	videoIdsCacheKey = "mission_aid_o_%v"
	videoIdsExpire   = 15
	sql4GetOperData  = `
SELECT data
FROM act_web_data
WHERE id = ?`
)

func (d *Dao) GetRoomIdsByOperSourceId(ctx context.Context, operSourceId int64) (res *model.LiveRoomList, err error) {
	res = &model.LiveRoomList{
		RoomIds: make([]int64, 0),
	}
	operRes := &model.LiveRoomListOper{}
	cacheKey := fmt.Sprintf(roomIdsCacheKey, operSourceId)
	bs, err := redis.Bytes(d.redis.Do(ctx, "GET", cacheKey))
	if err == nil {
		err = json.Unmarshal(bs, res)
	}
	if err == nil {
		return
	}
	var jsonStr string
	err = d.db.QueryRow(ctx, sql4GetOperData, operSourceId).Scan(&jsonStr)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(jsonStr), operRes)
	if err != nil {
		log.Errorc(ctx, "json.Unmarshal %v error: %v", jsonStr, err)
		err = ecode.SystemActivityConfigErr
		return
	}
	res.EntryFrom = operRes.EntryFrom
	roomIdsStr := strings.Split(operRes.RoomIds, "\n")
	if len(roomIdsStr) == 0 {
		return
	}
	var roomId int64
	for _, roomIdStr := range roomIdsStr {
		roomId, err = strconv.ParseInt(roomIdStr, 10, 64)
		if err == nil {
			res.RoomIds = append(res.RoomIds, roomId)
		}
	}
	if err == nil {
		bs, _ = json.Marshal(res)
		_, _ = d.redis.Do(ctx, "SETEX", cacheKey, roomIdsExpire, bs)
	}
	return
}

func (d *Dao) GetVideoAIdsByOperSourceId(ctx context.Context, operSourceId int64) (res []int64, err error) {
	res = make([]int64, 0)
	operRes := &model.VideoAidListOper{}
	cacheKey := fmt.Sprintf(videoIdsCacheKey, operSourceId)
	bs, err := redis.Bytes(d.redis.Do(ctx, "GET", cacheKey))
	if err == nil {
		err = json.Unmarshal(bs, &res)
	}
	if err == nil {
		return
	}
	var jsonStr string
	err = d.db.QueryRow(ctx, sql4GetOperData, operSourceId).Scan(&jsonStr)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(jsonStr), operRes)
	if err != nil {
		log.Errorc(ctx, "json.Unmarshal %v error: %v", jsonStr, err)
		err = ecode.SystemActivityConfigErr
		return
	}
	roomIdsStr := strings.Split(operRes.RoomIds, "\n")
	if len(roomIdsStr) == 0 {
		return
	}
	var aid int64
	for _, aidStr := range roomIdsStr {
		aid, err = strconv.ParseInt(aidStr, 10, 64)
		if err == nil {
			res = append(res, aid)
		}
	}
	if err == nil {
		bs, _ = json.Marshal(res)
		_, _ = d.redis.Do(ctx, "SETEX", cacheKey, roomIdsExpire, bs)
	}
	return
}
