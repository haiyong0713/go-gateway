package guess

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_UpDetailOdds(t *testing.T) {
	var (
		c          = context.Background()
		detailOdds map[int64]float64
	)
	detailOdds = make(map[int64]float64, 3)
	detailOdds[3] = 3.85
	detailOdds[4] = 2.9
	detailOdds[5] = 2.33
	convey.Convey("UpDetailOdds", t, func(ctx convey.C) {
		res, err := d.UpDetailOdds(c, detailOdds)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_GuessDetail(t *testing.T) {
	var (
		c      = context.Background()
		mainID = int64(1)
	)
	convey.Convey("GuessDetail", t, func(ctx convey.C) {
		res, err := d.GuessDetail(c, mainID)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_UpUser(t *testing.T) {
	var (
		c          = context.Background()
		mid        = int64(0)
		userIncome map[int64]float64
	)
	userIncome = make(map[int64]float64, 1)
	convey.Convey("UpUser", t, func(ctx convey.C) {
		res, err := d.UpUser(c, mid, userIncome)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_GuessFinish(t *testing.T) {
	var (
		c      = context.Background()
		mid    = int64(10000)
		mainID = int64(2)
	)
	convey.Convey("GuessFinish", t, func(ctx convey.C) {
		res, err := d.GuessFinish(c, mid, mainID)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_UpUserLog(t *testing.T) {
	var (
		c          = context.Background()
		userIncome map[int64]float64
		business   = int64(1)
	)
	userIncome = make(map[int64]float64, 1)
	userIncome[10000] = 1.5
	userIncome[100] = 2.6
	userIncome[20000] = 0
	userIncome[200] = 0
	convey.Convey("UpUserLog", t, func(ctx convey.C) {
		res, err := d.UpUserLog(c, userIncome, business)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_UserRank(t *testing.T) {
	var (
		c        = context.Background()
		business = int64(1)
		id       = int64(0)
		limit    = int64(100)
	)
	convey.Convey("UserRank", t, func(ctx convey.C) {
		res, err := d.UserRank(c, business, id, limit)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_UpUserRank(t *testing.T) {
	var (
		c        = context.Background()
		userRank map[int64]int64
	)
	userRank = make(map[int64]int64, 1)
	convey.Convey("UpUserRank", t, func(ctx convey.C) {
		res, err := d.UpUserRank(c, userRank)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_UpUserStatus(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(0)
		ids = []int64{1, 5, 6}
	)
	convey.Convey("UpUserStatus", t, func(ctx convey.C) {
		res, err := d.UpUserStatus(c, mid, ids)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
