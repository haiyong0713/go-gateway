package account

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestAccountrelations3(t *testing.T) {
	var (
		c      = context.Background()
		owners = []int64{12, 13}
		mid    = int64(11)
	)
	convey.Convey("relations3", t, func(ctx convey.C) {
		relations, err := d.relations3(c, owners, mid)
		ctx.Convey("Then err should be nil.relations should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(relations, convey.ShouldNotBeNil)
		})
	})
}

func TestAccountRelations3(t *testing.T) {
	var (
		c      = context.Background()
		owners = []int64{12, 13}
		mid    = int64(11)
	)
	convey.Convey("Relations3", t, func(ctx convey.C) {
		follows := d.Relations3(c, owners, mid)
		ctx.Convey("Then follows should not be nil.", func(ctx convey.C) {
			ctx.So(follows, convey.ShouldNotBeNil)
		})
	})
}

func TestAccountIsAttention(t *testing.T) {
	var (
		c      = context.Background()
		owners = []int64{12, 13}
		mid    = int64(11)
	)
	convey.Convey("IsAttention", t, func(ctx convey.C) {
		isAtten := d.IsAttention(c, owners, mid)
		ctx.Convey("Then isAtten should not be nil.", func(ctx convey.C) {
			ctx.So(isAtten, convey.ShouldNotBeNil)
		})
	})
}

func TestAccountCard3(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(11)
	)
	convey.Convey("Card3", t, func(ctx convey.C) {
		res, err := d.Card3(c, mid)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestAccountCards3(t *testing.T) {
	var (
		c    = context.Background()
		mids = []int64{12, 13}
	)
	convey.Convey("Cards3", t, func(ctx convey.C) {
		res, err := d.Cards3(c, mids)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestAccountFollowing3(t *testing.T) {
	var (
		c     = context.Background()
		mid   = int64(11)
		owner = int64(12)
	)
	convey.Convey("Following3", t, func(ctx convey.C) {
		follow, err := d.Following3(c, mid, owner)
		ctx.Convey("Then err should be nil.follow should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(follow, convey.ShouldNotBeNil)
		})
	})
}

func TestAccountInfos3(t *testing.T) {
	var (
		c    = context.Background()
		mids = []int64{12, 13}
	)
	convey.Convey("Infos3", t, func(ctx convey.C) {
		res, err := d.Infos3(c, mids)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestAccountIsVip(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(12)
	)
	convey.Convey("IsVip", t, func(ctx convey.C) {
		res, err := d.IsVip(c, mid)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
