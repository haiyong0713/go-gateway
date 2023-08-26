package show

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-job/job/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestAddCacheAIData(t *testing.T) {
	convey.Convey("AddCacheAIData", t, func(convCtx convey.C) {
		var (
			ctx = context.Background()
		)
		maps := make(map[int64][]show.PopAIChannelResource)
		list1 := []show.PopAIChannelResource{}
		list1 = append(list1, show.PopAIChannelResource{
			RID:        960010833,
			TagId:      "0",
			Goto:       "av",
			FromType:   "recommend",
			Desc:       "",
			CornerMark: 0,
		})
		list1 = append(list1, show.PopAIChannelResource{
			RID:        880070772,
			TagId:      "0",
			Goto:       "av",
			FromType:   "recommend",
			Desc:       "",
			CornerMark: 0,
		})
		list1 = append(list1, show.PopAIChannelResource{
			RID:        400092665,
			TagId:      "0",
			Goto:       "av",
			FromType:   "recommend",
			Desc:       "",
			CornerMark: 0,
		})
		list1 = append(list1, show.PopAIChannelResource{
			RID:        440086947,
			TagId:      "0",
			Goto:       "av",
			FromType:   "recommend",
			Desc:       "",
			CornerMark: 0,
		})
		list1 = append(list1, show.PopAIChannelResource{
			RID:        520025078,
			TagId:      "0",
			Goto:       "av",
			FromType:   "recommend",
			Desc:       "",
			CornerMark: 0,
		})
		maps[280] = list1
		list2 := []show.PopAIChannelResource{}
		list2 = append(list1, show.PopAIChannelResource{
			RID:        400016089,
			TagId:      "0",
			Goto:       "av",
			FromType:   "recommend",
			Desc:       "",
			CornerMark: 0,
		})
		list2 = append(list1, show.PopAIChannelResource{
			RID:        520027562,
			TagId:      "0",
			Goto:       "av",
			FromType:   "recommend",
			Desc:       "",
			CornerMark: 0,
		})
		list2 = append(list1, show.PopAIChannelResource{
			RID:        920042133,
			TagId:      "0",
			Goto:       "av",
			FromType:   "recommend",
			Desc:       "",
			CornerMark: 0,
		})
		list2 = append(list1, show.PopAIChannelResource{
			RID:        600073898,
			TagId:      "0",
			Goto:       "av",
			FromType:   "recommend",
			Desc:       "",
			CornerMark: 0,
		})
		list2 = append(list1, show.PopAIChannelResource{
			RID:        360091969,
			TagId:      "0",
			Goto:       "av",
			FromType:   "recommend",
			Desc:       "",
			CornerMark: 0,
		})
		maps[281] = list2
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheAIData(ctx, maps)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddEntranceCache(t *testing.T) {
	convey.Convey("AddEntranceCache", t, func(convCtx convey.C) {
		var (
			ctx = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddEntranceCache(ctx)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetAllEntranceIds(t *testing.T) {
	convey.Convey("GetAllEntranceIds", t, func(convCtx convey.C) {
		var (
			ctx = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.GetAllEntranceIds(ctx)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}
