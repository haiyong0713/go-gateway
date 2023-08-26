package aggregation

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/aggregation"

	"github.com/smartystreets/goconvey/convey"
)

func TestAddAggregation(t *testing.T) {
	convey.Convey("TestAddAggregation", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			param = aggregation.AggPub{
				HotTitle: "lol",
				Title:    "lol",
				SubTitle: "lol",
				Image:    "www.bilibili.com",
			}
			tagIDs = []int64{9222}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			id, err := d.AddAggregation(c, param, tagIDs)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(id, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestUpdateAggregation(t *testing.T) {
	convey.Convey("TestUpdateAggregation", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			param = aggregation.AggPub{
				HotTitle: "lol",
				Title:    "lol",
				SubTitle: "lol",
				Image:    "www.bilibili.com",
			}
			tagIDs = []int64{9222}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.UpdateAggregation(c, param, tagIDs)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAggOperate(t *testing.T) {
	convey.Convey("TestAggOperate", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AggOperate(c, 1, 1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAggList(t *testing.T) {
	convey.Convey("TestAggList", t, func(ctx convey.C) {

		var (
			c     = context.Background()
			param = &aggregation.AggListReq{
				ID: 1,
			}
			hotID = []int64{1}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.AggList(c, param, hotID)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestFindByTagIDs(t *testing.T) {
	convey.Convey("TestFindByTagIDs", t, func(ctx convey.C) {

		var (
			c      = context.Background()
			tagIDs = []int64{9222}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.FindByTagIDs(c, tagIDs)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestTagIDByID(t *testing.T) {
	convey.Convey("TestTagIDByID", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.TagIDByID(c, 1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestTagIDByName(t *testing.T) {
	convey.Convey("TestTagIDByName", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.TagIDByName(c, "lol")
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestNameByTagID(t *testing.T) {
	convey.Convey("TestNameByTagID", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			tagIDs = []int64{12131, 20215, 239855, 1102683, 1207642, 6942, 20215, 239855, 1102683, 1207642, 20215, 6942, 20215, 1102683, 10179, 6942, 239855, 239855, 1207642, 6942, 20215, 239855, 1102683, 1207642, 20215, 6942, 12131, 20215, 239855, 1102683, 1207642, 6942, 20215, 239855, 1102683, 1207642, 20215}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.NameByTagID(c, tagIDs)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestNamesByTagIDs(t *testing.T) {
	convey.Convey("TestNameByTagID", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			tagIDs = []int64{6942, 6942, 20215, 1102683, 10179, 6942, 239855, 239855, 1207642, 6942, 20215, 239855, 1102683, 1207642, 20215, 6942, 12131, 20215, 239855, 1102683, 1207642, 6942, 20215, 239855, 1102683, 1207642, 20215, 6942, 20215, 1102683, 10179, 6942, 239855, 239855, 1207642, 6942, 20215, 239855, 1102683, 1207642, 20215, 6942, 12131, 20215, 239855, 1102683, 1207642, 6942, 20215, 239855, 1102683, 1207642, 20215}
		)
		req := aggregation.FilterDupIDs(tagIDs)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.NamesByTagIDs(c, req)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestAggView(t *testing.T) {
	convey.Convey("TestAggView", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.AggView(c, 1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestHotWordCount(t *testing.T) {
	convey.Convey("TestHotWordCount", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.HotWordCount(c, "lol少女前线")
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestFindNameByID(t *testing.T) {
	convey.Convey("TestFindNameByID", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.FindNameByID(c, 1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestHwResourceByHwID(t *testing.T) {
	convey.Convey("HwResourceByHwID", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.HwResourceByHwID(c, 10000)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestAggTagAddM(t *testing.T) {
	convey.Convey("AggTagAddM", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AggTagAddM(c, 209, []int64{600})
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
