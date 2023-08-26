package unicom

import (
	"context"
	"testing"

	"go-common/library/cache/memcache"
	"go-gateway/app/app-svr/app-wall/job/model/unicom"

	"github.com/smartystreets/goconvey/convey"
)

func TestUnicomkeyUserBind(t *testing.T) {
	var (
		mid = int64(0)
	)
	convey.Convey("keyUserBind", t, func(ctx convey.C) {
		p1 := keyUserBind(mid)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestUnicomkeyUserBindReceive(t *testing.T) {
	var (
		mid = int64(0)
	)
	convey.Convey("keyUserBindReceive", t, func(ctx convey.C) {
		p1 := keyUserBindReceive(mid)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestUnicomkeyUserComicReceive(t *testing.T) {
	var (
		mid = int64(0)
	)
	convey.Convey("keyUserComicReceive", t, func(ctx convey.C) {
		p1 := keyUserComicReceive(mid)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestUnicomkeyUnicom(t *testing.T) {
	var (
		usermob = ""
	)
	convey.Convey("keyUnicom", t, func(ctx convey.C) {
		p1 := keyUnicom(usermob)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestUnicomkeyUserFlow(t *testing.T) {
	var (
		key = ""
	)
	convey.Convey("keyUserFlow", t, func(ctx convey.C) {
		p1 := keyUserFlow(key)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestUnicomUserBindCache(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(27888)
	)
	convey.Convey("UserBindCache", t, func(ctx convey.C) {
		_, err := d.UserBindCache(c, mid)
		if err == memcache.ErrNotFound {
			err = nil
		}
		ctx.Convey("Then err should be nil.ub should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUnicomAddUserBindCache(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(27888888)
		ub  = &unicom.UserBind{}
	)
	convey.Convey("AddUserBindCache", t, func(ctx convey.C) {
		err := d.AddUserBindCache(c, mid, ub)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUnicomUnicomCache(t *testing.T) {
	var (
		c       = context.Background()
		usermob = ""
	)
	convey.Convey("UnicomCache", t, func(ctx convey.C) {
		_, err := d.UnicomCache(c, usermob)
		if err == memcache.ErrNotFound {
			err = nil
		}
		ctx.Convey("Then err should be nil.u should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUnicomUserFlowCache(t *testing.T) {
	var (
		c      = context.Background()
		keyStr = ""
	)
	convey.Convey("UserFlowCache", t, func(ctx convey.C) {
		err := d.UserFlowCache(c, keyStr)
		if err == memcache.ErrNotFound {
			err = nil
		}
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUnicomDeleteUserFlowCache(t *testing.T) {
	var (
		c      = context.Background()
		keyStr = ""
	)
	convey.Convey("DeleteUserFlowCache", t, func(ctx convey.C) {
		err := d.DeleteUserFlowCache(c, keyStr)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUnicomUserFlowListCache(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("UserFlowListCache", t, func(ctx convey.C) {
		_, err := d.UserFlowListCache(c)
		if err == memcache.ErrNotFound {
			err = nil
		}
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUnicomAddUserFlowListCache(t *testing.T) {
	var (
		c    = context.Background()
		list map[string]*unicom.UnicomUserFlow
	)
	convey.Convey("AddUserFlowListCache", t, func(ctx convey.C) {
		err := d.AddUserFlowListCache(c, list)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
