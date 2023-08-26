package bws

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestBwsUsersMid(t *testing.T) {
	convey.Convey("UsersMid", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			key = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.UsersMid(c, bid, key)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsUsersKey(t *testing.T) {
	convey.Convey("UsersKey", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			bid  = int64(0)
			ukey = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.UsersKey(c, bid, ukey)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsPoints(t *testing.T) {
	convey.Convey("Points", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Points(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsBwsPoints(t *testing.T) {
	convey.Convey("BwsPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.BwsPoints(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsBwsSign(t *testing.T) {
	convey.Convey("BwsSign", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.BwsSign(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsSigns(t *testing.T) {
	convey.Convey("Signs", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			pid = int64(54)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Signs(c, pid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAchievements(t *testing.T) {
	convey.Convey("Achievements", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(6)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Achievements(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsUserAchieves(t *testing.T) {
	convey.Convey("UserAchieves", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(6)
			key = "ba271e537e71a9bb"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.UserAchieves(c, bid, key)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsUserPoints(t *testing.T) {
	convey.Convey("UserPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(6)
			key = "4a765e5fed1f9280"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.UserPoints(c, bid, key)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsUserLockPoints(t *testing.T) {
	convey.Convey("UserLockPoints", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			bid      = int64(6)
			lockType = int64(1)
			ukey     = "4457568c828b253d"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.UserLockPoints(c, bid, lockType, ukey)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAchieveCounts(t *testing.T) {
	convey.Convey("AchieveCounts", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			day = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.AchieveCounts(c, bid, day)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsRechargeLevels(t *testing.T) {
	convey.Convey("RechargeLevels", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{134}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RechargeLevels(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsRechargeAwards(t *testing.T) {
	convey.Convey("RechargeAwards", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{134}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RechargeAwards(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsPointLevels(t *testing.T) {
	convey.Convey("PointLevels", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.PointLevels(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsPointsAward(t *testing.T) {
	convey.Convey("PointsAward", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			plID = int64(134)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.PointsAward(c, plID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsCompositeAchievesPoint(t *testing.T) {
	convey.Convey("CompositeAchievesPoint", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			mids = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CompositeAchievesPoint(c, mids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsActFields(t *testing.T) {
	convey.Convey("ActFields", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.ActFields(c, bid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}
