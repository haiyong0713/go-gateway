package bws

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCachePointLevels(t *testing.T) {
	convey.Convey("CachePointLevels", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.CachePointLevels(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestDelCachePointLevels(t *testing.T) {
	convey.Convey("DelCachePointLevels", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCachePointLevels(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddCachePointLevels(t *testing.T) {
	convey.Convey("AddCachePointLevels", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(1)
			ids = []int64{1, 3}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCachePointLevels(c, bid, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawPointLevels(t *testing.T) {
	convey.Convey("RawPointLevels", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(6)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawPointLevels(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestRawPointsAward(t *testing.T) {
	convey.Convey("RawPointsAward", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(6)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawPointsAward(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestCachePointsAward(t *testing.T) {
	convey.Convey("CachePointsAward", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			plID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.CachePointsAward(c, plID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestDelCachePointsAward(t *testing.T) {
	convey.Convey("DelCachePointsAward", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			plID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCachePointsAward(c, plID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddCachePointsAward(t *testing.T) {
	convey.Convey("AddCachePointsAward", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			plID = int64(1)
			ids  = []int64{1, 2}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCachePointsAward(c, plID, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawRechargeAwards(t *testing.T) {
	convey.Convey("RawRechargeAwards", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1, 2}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawRechargeAwards(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestRawRechargeLevels(t *testing.T) {
	convey.Convey("RawRechargeLevels", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1, 2}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawRechargeLevels(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

// PointLevelsByID
func TestPointLevelsByID(t *testing.T) {
	convey.Convey("PointLevelsByID", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			bid   = int64(6)
			plids = []int64{2, 3}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.PointLevelsByID(c, bid, plids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestAddCacheSetMembers(t *testing.T) {
	convey.Convey("AddCacheSetMembers", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheSetMembers(c, "bws_p_lel_7", []int64{1, 2})
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCacheSetMembers(t *testing.T) {
	convey.Convey("CacheSetMembers", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.CacheSetMembers(c, "bws_p_lel_7")
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestDelCacheSetMembers(t *testing.T) {
	convey.Convey("DelCacheSetMembers", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheSetMembers(c, "bws_p_lel_7")
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
