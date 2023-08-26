package like

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikeredisKey(t *testing.T) {
	convey.Convey("redisKey", t, func(ctx convey.C) {
		var (
			key = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := redisKey(key)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeRsGet(t *testing.T) {
	convey.Convey("RsGet", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RsGet(c, key)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeRiGet(t *testing.T) {
	convey.Convey("RiGet", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RiGet(c, key)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeRsSetNX(t *testing.T) {
	convey.Convey("RsSetNX", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			key    = ""
			expire = int32(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RsSetNX(c, key, expire)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeRb(t *testing.T) {
	convey.Convey("Rb", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.Rb(c, key)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeIncr(t *testing.T) {
	convey.Convey("Incr", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.Incr(c, key)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeIncrby(t *testing.T) {
	convey.Convey("Incrby", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.Incrby(c, key)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestRawEsLikesIDs(t *testing.T) {
	convey.Convey("RawEsLikesIDs", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10523)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res := d.RawEsLikesIDs(c, sid, 0, 1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.Print(res)
			})
		})
	})
}

func TestRsNXGet(t *testing.T) {
	convey.Convey("RsNXGet", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = "suits_c_k_27515254_123456"
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.RsNXGet(c, key)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(res, convey.ShouldNotBeNil)
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestIncrCacheReserveTotal(t *testing.T) {
	convey.Convey("IncrCacheReserveTotal", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.IncrCacheReserveTotal(c, 10829, 1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCacheReservesTotal(t *testing.T) {
	convey.Convey("CacheReservesTotal", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.CacheReservesTotal(c, []int64{10829, 10629})
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

func TestAddCacheReservesTotal(t *testing.T) {
	convey.Convey("AddCacheReservesTotal", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddCacheReservesTotal(c, map[int64]int64{10829: 12, 10928: 11})
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
