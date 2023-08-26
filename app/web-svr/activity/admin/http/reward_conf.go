package http

import (
	"encoding/json"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	binding2 "go-common/library/net/http/blademaster/binding"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/admin/model/reward_conf"
	"strconv"
	"time"
)

func addRewardConfRouter(group *bm.RouterGroup) {
	rewardConfGroup := group.Group("/reward_conf")
	{
		rewardConfGroup.POST("/add", addOneRewardConf)
		rewardConfGroup.POST("/update", updateOneRewardConf)
		rewardConfGroup.GET("/search", searchAwardList)
	}
}

// addOneRewardConf 添加一条记录
func addOneRewardConf(ctx *bm.Context) {
	v := new(reward_conf.AddOneRewardReq)
	if err := ctx.BindWith(v, binding2.JSON); err != nil {
		return
	}
	if v.ShowTime > v.EndTime || v.EndTime < (xtime.Time)(time.Now().Unix()) {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, rewardConfSrv.AddOneRewardConf(ctx, v, Strval(username)))
}

// updateOneRewardConf 修改一条记录
func updateOneRewardConf(ctx *bm.Context) {
	v := new(reward_conf.UpdateOneRewardReq)
	if err := ctx.BindWith(v, binding2.JSON); err != nil {
		return
	}
	if v.ShowTime > v.EndTime {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, rewardConfSrv.UpdateOneRewardConf(ctx, v, Strval(username)))
}

// searchAwardList 查询
func searchAwardList(ctx *bm.Context) {
	v := new(reward_conf.SearchReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	if v.STime > v.ETime {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	if v.STime == 0 && v.ETime == 0 {
		now := time.Now().Unix()
		v.STime = xtime.Time(now)
		v.ETime = xtime.Time(now)

	}
	ctx.JSON(rewardConfSrv.SearchList(ctx, v))
}

// Strval interface转string
func Strval(value interface{}) string {
	// interface 转 string
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}
