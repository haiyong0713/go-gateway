package show

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestTopEntrance(t *testing.T) {
	convey.Convey("Entrances", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := dao.Entrances(context.Background())
			str, _ := json.Marshal(res)
			fmt.Println(string(str), err)
			convCtx.Convey("Then err should be nil.p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestCacheAIChannelRes(t *testing.T) {
	convey.Convey("Entrances", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := dao.CacheAIChannelRes(context.Background(), 341)
			str, _ := json.Marshal(res)
			fmt.Println(string(str), err)
			convCtx.Convey("Then err should be nil.p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
