package native

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/native-page/interface/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestNativeModuleCache(t *testing.T) {
	convey.Convey("ModuleCache", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			nid = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheModuleIDs(c, nid, 0, 1, 0)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeAddModuleCache(t *testing.T) {
	convey.Convey("AddModuleCache", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			nid = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheModuleIDs(c, nid, nil, 0)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeDeleteModuleCache(t *testing.T) {
	convey.Convey("DelModuleCache", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			nid = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DeleteModuleCache(c, nid, 0)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeActCache(t *testing.T) {
	convey.Convey("ActCache", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			moduleID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.NativeActIDs(c, moduleID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeDynamicCache(t *testing.T) {
	convey.Convey("DynamicCache", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			moduleID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.NativeDynamicIDs(c, moduleID)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddCacheNatMixIDs(t *testing.T) {
	convey.Convey("AddCacheNatMixIDs", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			moduleID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheNatMixIDs(c, moduleID, nil, 1)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCacheNatMixIDs(t *testing.T) {
	convey.Convey("CacheNatMixIDs", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			moduleID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheNatMixIDs(c, moduleID, 0, -1, 1)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestDelCacheNatMixIDs(t *testing.T) {
	convey.Convey("DelCacheNatMixIDs", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			moduleID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheNatMixIDs(c, moduleID, 1)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeAddCacheNativeSponsorsPages(t *testing.T) {
	convey.Convey("AddCacheNativeSponsorsPages", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(1100)
			ids = []*api.NativePage{{ID: 1}, {ID: 2}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheNativePagesByMids(c, mid, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
