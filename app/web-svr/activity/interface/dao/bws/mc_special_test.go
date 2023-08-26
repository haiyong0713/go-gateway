package bws

import (
	"context"
	"testing"

	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"github.com/smartystreets/goconvey/convey"
)

func TestAddCacheUsersMids(t *testing.T) {
	convey.Convey("AddCacheUsersMids", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(2)
			val = map[int64]*bwsmdl.Users{2: {ID: 2, Mid: 2, Key: "123456", Ctime: 1561954074, Mtime: 1561954074, Bid: 2}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUsersMids(c, id, val)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCacheUsersMids(t *testing.T) {
	convey.Convey("CacheUsersMids", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			id   = int64(2)
			mids = []int64{1, 2}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.CacheUsersMids(c, id, mids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestAchieveReloadSet(t *testing.T) {
	convey.Convey("AchieveReloadSet", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.AchieveReloadSet(c, 1, 1548787774, 10)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestSetNXCache(t *testing.T) {
	convey.Convey("setNXCache", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.setNXCache(c, "b_te_tes8", 10)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestUsersMids(t *testing.T) {
	convey.Convey("UsersMids", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.UsersMids(c, 1, []int64{0})
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestCacheUsersKeys(t *testing.T) {
	convey.Convey("CacheUsersKeys", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.CacheUsersKeys(c, 1, []string{})
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}

func TestAddCacheUsersKeys(t *testing.T) {
	convey.Convey("AddCacheUsersKeys", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUsersKeys(c, 1, map[string]*bwsmdl.Users{"8af2ec0295e86c33": {ID: 1, Mid: 1}})
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestUsersKeys(t *testing.T) {
	convey.Convey("UsersKeys", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.UsersKeys(c, 1, []string{})
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list)
			})
		})
	})
}
