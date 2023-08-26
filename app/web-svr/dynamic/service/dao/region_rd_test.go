package dao

import (
	"context"
	"testing"
	"time"

	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/dynamic/service/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaofmtAllKey(t *testing.T) {
	convey.Convey("fmtAllKey", t, func(ctx convey.C) {
		var (
			rid     = int32(0)
			pubDate = ""
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			p1 := fmtAllKey(rid, pubDate)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaofmtOriginKey(t *testing.T) {
	convey.Convey("fmtOriginKey", t, func(ctx convey.C) {
		var (
			rid     = int32(32)
			pubDate = "201801"
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			p1 := fmtOriginKey(rid, pubDate)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaofmtEveryRegKey(t *testing.T) {
	convey.Convey("fmtEveryRegKey", t, func(ctx convey.C) {
		var (
			rid = int32(0)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			p1 := fmtEveryRegKey(rid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaofmtEveryOrRegKey(t *testing.T) {
	convey.Convey("fmtEveryOrRegKey", t, func(ctx convey.C) {
		var (
			rid = int32(0)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			p1 := fmtEveryOrRegKey(rid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddRegionArcCache(t *testing.T) {
	convey.Convey("AddRegionArcCache", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			rid  = int32(0)
			reid = int32(0)
			arc  = &arcmdl.RegionArc{}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.AddRegionArcCache(c, rid, reid, arc)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoAllRegion(t *testing.T) {
	convey.Convey("AllRegion", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			param = []*model.ResKey{}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			_, err := d.AllRegion(c, param)
			ctx.Convey("Then err should be nil.aids should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoRegionKeyCount(t *testing.T) {
	convey.Convey("RegionKeyCount", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = ""
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			_, err := d.RegionKeyCount(c, key)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoRegCount(t *testing.T) {
	convey.Convey("RegCount", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = ""
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			count, err := d.RegCount(c, key)
			ctx.Convey("Then err should be nil.count should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(count, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoDelArcCache(t *testing.T) {
	convey.Convey("DelArcCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			rid   = int32(0)
			reid  = int32(0)
			param = &arcmdl.RegionArc{}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.DelArcCache(c, rid, reid, param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaosaveRegCount(t *testing.T) {
	convey.Convey("saveRegCount", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			rid = int32(0)
			arc = &arcmdl.RegionArc{}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.saveRegCount(c, rid, arc)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoPushFail(t *testing.T) {
	convey.Convey("PushFail", t, func(ctx convey.C) {
		var (
			c = context.Background()
			a = interface{}(0)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PushFail(c, a)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoPopFail(t *testing.T) {
	convey.Convey("PopFail", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			bt, err := d.PopFail(c)
			ctx.Convey("Then err should be nil.bt should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(bt, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoRegionCnt(t *testing.T) {
	convey.Convey("RegionCnt", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			t   = time.Now()
			min = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Unix()
			max = t.Unix()
			rid = []int32{127}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.RegionCnt(c, rid, min, max)
			ctx.Convey("Then err should be nil.bt should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoRecentRegArc(t *testing.T) {
	convey.Convey("TestDaoRecentRegArc", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			start = 0
			end   = 9
			rid   = int32(0)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.RecentRegArc(c, rid, start, end)
			ctx.Convey("Then err should be nil.bt should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoRecentAllRegArcCnt(t *testing.T) {
	convey.Convey("TestDaoRecentAllRegArcCnt", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.RecentAllRegArcCnt(c)
			ctx.Convey("Then err should be nil.bt should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
