package dao

import (
	"testing"

	"go-gateway/app/web-svr/web/job/internal/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_WebTop(t *testing.T) {
	Convey("WebTop", t, func(convCtx C) {
		Convey("When everything goes positive", func(convCtx C) {
			data, err := d.WebTop(ctx)
			So(err, ShouldBeNil)
			So(len(data), ShouldBeGreaterThan, 0)
		})
	})
}

func TestDao_RankIndex(t *testing.T) {
	var day int64 = 3
	Convey("RankIndex", t, func(convCtx C) {
		Convey("When everything goes positive", func(convCtx C) {
			data, err := d.RankIndex(ctx, day)
			So(err, ShouldBeNil)
			So(len(data), ShouldBeGreaterThan, 0)
		})
	})
}

func TestDao_RankRecommend(t *testing.T) {
	var rid int64 = 3
	Convey("RankRecommend", t, func(convCtx C) {
		Convey("When everything goes positive", func(convCtx C) {
			data, err := d.RankRecommend(ctx, rid)
			So(err, ShouldBeNil)
			So(len(data), ShouldBeGreaterThan, 0)
		})
	})
}

func TestDao_RankRegion(t *testing.T) {
	var (
		rid      int64 = 13
		day      int64 = 3
		original int64 = 1
	)
	Convey("RankRegion", t, func(convCtx C) {
		Convey("When everything goes positive", func(convCtx C) {
			data, err := d.RankRegion(ctx, rid, day, original)
			So(err, ShouldBeNil)
			So(len(data), ShouldBeGreaterThan, 0)
		})
	})
}

func TestDao_RankList(t *testing.T) {
	var (
		typ       = model.RankListTypeRookie
		rid int64 = 0
	)
	Convey("RankList", t, func(convCtx C) {
		Convey("When everything goes positive", func(convCtx C) {
			data, err := d.RankList(ctx, typ, rid)
			So(err, ShouldBeNil)
			So(len(data.List), ShouldBeGreaterThan, 0)
		})
	})
}

func TestDao_RankListOld(t *testing.T) {
	var (
		rid int64 = 11
	)
	Convey("RankList", t, func(convCtx C) {
		Convey("When everything goes positive", func(convCtx C) {
			data, err := d.RankListOld(ctx, rid)
			So(err, ShouldBeNil)
			So(len(data.List), ShouldBeGreaterThan, 0)
		})
	})
}

func TestDao_RankTag(t *testing.T) {
	var (
		rid   int64 = 130
		tagID int64 = 23292
	)
	Convey("RankTag", t, func(convCtx C) {
		Convey("When everything goes positive", func(convCtx C) {
			data, err := d.RankTag(ctx, rid, tagID)
			So(err, ShouldBeNil)
			So(len(data), ShouldBeGreaterThan, 0)
		})
	})
}
