package dao

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaokeyRegion(t *testing.T) {
	convey.Convey("keyRegion", t, func(ctx convey.C) {
		var (
			rid = int64(0)
			pn  = int64(0)
			ps  = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyRegion(rid, pn, ps)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyRegionTag(t *testing.T) {
	convey.Convey("keyRegionTag", t, func(ctx convey.C) {
		var (
			tagID = int64(0)
			rid   = int64(0)
			pn    = int64(0)
			ps    = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyRegionTag(tagID, rid, pn, ps)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetRegionBakCache(t *testing.T) {
	convey.Convey("SetRegionBakCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			rid = int64(0)
			pn  = int64(0)
			ps  = int64(0)
			rs  = &model.DynamicBvArcs{}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetRegionBakCache(c, rid, pn, ps, rs)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoRegionBakCache(t *testing.T) {
	convey.Convey("RegionBakCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			rid = int64(0)
			pn  = int64(0)
			ps  = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			rs, err := d.RegionBakCache(c, rid, pn, ps)
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(rs, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetRegionTagBakCache(t *testing.T) {
	convey.Convey("SetRegionTagBakCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			tagID = int64(0)
			rid   = int64(0)
			pn    = int64(0)
			ps    = int64(0)
			rs    = &model.DynamicBvArcs{}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetRegionTagBakCache(c, tagID, rid, pn, ps, rs)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoRegionTagBakCache(t *testing.T) {
	convey.Convey("RegionTagBakCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			tagID = int64(0)
			rid   = int64(0)
			pn    = int64(0)
			ps    = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			rs, err := d.RegionTagBakCache(c, tagID, rid, pn, ps)
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(rs, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetRegionsBakCache(t *testing.T) {
	convey.Convey("SetRegionsBakCache", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		rs := map[int32][]*model.BvArc{1111: {{Arc: &api.Arc{Aid: 1111}}}}
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetRegionsBakCache(c, rs)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoRegionsBakCache(t *testing.T) {
	convey.Convey("RegionsBakCache", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			rs, err := d.RegionsBakCache(c)
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(rs, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaosetBakCache(t *testing.T) {
	convey.Convey("setBakCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = ""
			rs  = &model.DynamicBvArcs{}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.setBakCache(c, key, rs)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaobakCache(t *testing.T) {
	convey.Convey("bakCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			rs, err := d.bakCache(c, key)
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(rs, convey.ShouldNotBeNil)
			})
		})
	})
}
