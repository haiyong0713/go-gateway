package space

import (
	"math"
	"sort"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
)

func sortReserveRelationInfo(data []*activitygrpc.UpActReserveRelationInfo) {
	// 赛事预约（组内开始时间升序）>直播预约（组内原定开播时间升序）>首映 >稿件预约
	typeOrder := map[activitygrpc.UpActReserveRelationType]int64{
		activitygrpc.UpActReserveRelationType_ESports:  1,
		activitygrpc.UpActReserveRelationType_Live:     2,
		activitygrpc.UpActReserveRelationType_Premiere: 3,
		activitygrpc.UpActReserveRelationType_Archive:  4,
	}
	sort.Slice(data, func(i, j int) bool {
		if data[i].Type != data[j].Type {
			ci, ok := typeOrder[data[i].Type]
			if !ok {
				ci = math.MaxInt64
			}
			cj, ok := typeOrder[data[j].Type]
			if !ok {
				cj = math.MaxInt64
			}
			return ci < cj
		}
		if data[i].StartShowTime != data[j].StartShowTime {
			return data[i].StartShowTime < data[j].StartShowTime
		}
		if data[i].LivePlanStartTime != data[j].LivePlanStartTime {
			return data[i].LivePlanStartTime < data[j].LivePlanStartTime
		}
		return data[i].Sid < data[j].Sid
	})
}
