package bws

import (
	"context"
	"testing"

	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"github.com/smartystreets/goconvey/convey"
)

func TestBwsRawUsersMid(t *testing.T) {
	convey.Convey("RawUsersMid", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(6)
			mid = int64(27515326)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawUsersMid(c, bid, mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsRawUsersKey(t *testing.T) {
	convey.Convey("RawUsersKey", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = int64(0)
			key = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawUsersKey(c, bid, key)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsBinding(t *testing.T) {
	convey.Convey("Binding", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			loginMid = int64(0)
			p        = &bwsmdl.ParamBinding{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.Binding(c, loginMid, p)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsUserByID(t *testing.T) {
	convey.Convey("UserByID", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			keyID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.UserByID(c, keyID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestRawUsersMids(t *testing.T) {
	convey.Convey("RawUsersMids", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			bid  = int64(1)
			mids = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawUsersMids(c, bid, mids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

// RawUsersByBid
func TestRawUsersByBid(t *testing.T) {
	convey.Convey("RawUsersByBid", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = []int64{6}
			id  = int64(14)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawUsersByBid(c, bid, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

// RawUsersBids
func TestRawUsersBids(t *testing.T) {
	convey.Convey("RawUsersBids", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = []int64{3, 4, 5}
			id  = int64(14)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawUsersBids(c, bid, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestRawBidUsersMid(t *testing.T) {
	convey.Convey("RawBidUsersMid", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			bid = []int64{1, 4}
			id  = int64(908087)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawBidUsersMid(c, bid, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestRawUsersKeys(t *testing.T) {
	convey.Convey("RawUsersKeys", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			keys = []string{"test1", "test2"}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawUsersKeys(c, 1, keys)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}
