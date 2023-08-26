package steins

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoGraphShow(t *testing.T) {
	convey.Convey("GraphShow", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			aid     = int64(0)
			graphID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.GraphShow(c, aid, graphID)
			convCtx.Convey("Then err should be nil.data should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(data, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoGraphInfo(t *testing.T) {
	convey.Convey("GraphInfo", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(10114549)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			a, err := d.GraphInfo(c, aid)
			fmt.Println(a)
			convCtx.Convey("Then err should be nil.a should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(a, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoGraphInfoPreview(t *testing.T) {
	convey.Convey("GraphInfoPreview", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			a, err := d.GraphInfoPreview(c, aid)
			convCtx.Convey("Then err should be nil.a should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(a, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaosetGraphAllCache(t *testing.T) {
	convey.Convey("setGraphAllCache", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			aid     = int64(10113448)
			graphID = int64(4)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.setGraphAllCache(c, aid, graphID, nil)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_GraphInfos(t *testing.T) {
	convey.Convey("setGraphAllCache", t, func(convCtx convey.C) {
		var c = context.Background()
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res := d.GraphInfos(c, []int64{10111819,
				10111160,
				10113611,
				10113601,
				10113593})
			str, _ := json.Marshal(res)
			fmt.Println(string(str))
		})
	})
}
