package native

import (
	"context"
	v1 "go-gateway/app/web-svr/native-page/interface/api"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNativeCacheNativePages(t *testing.T) {
	convey.Convey("CacheNativePages", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{2}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheNativePages(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeAddCacheNativePages(t *testing.T) {
	convey.Convey("AddCacheNativePages", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			values = map[int64]*v1.NativePage{2: {ID: 2}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheNativePages(c, values)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeDelCacheNativePages(t *testing.T) {
	convey.Convey("DelCacheNativePages", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{2}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheNativePages(c, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeCacheNativeForeign(t *testing.T) {
	convey.Convey("CacheNativeForeign", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			id       = int64(1032)
			pageType = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheNativeForeign(c, id, pageType)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeAddCacheNativeForeign(t *testing.T) {
	convey.Convey("AddCacheNativeForeign", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			id       = int64(1035)
			val      = int64(2)
			pageType = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheNativeForeign(c, id, val, pageType)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeCacheNativeForeigns(t *testing.T) {
	convey.Convey("CacheNativeForeigns", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			ids      = []int64{1035}
			pageType = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheNativeForeigns(c, ids, pageType)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeAddCacheNativeForeigns(t *testing.T) {
	convey.Convey("AddCacheNativeForeigns", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			values   = map[int64]int64{1035: 2}
			pageType = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheNativeForeigns(c, values, pageType)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeDelCacheNativeForeign(t *testing.T) {
	convey.Convey("DelCacheNativeForeign", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			id       = int64(1035)
			pageType = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheNativeForeign(c, id, pageType)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeCacheNativeModules(t *testing.T) {
	convey.Convey("CacheNativeModules", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{105, 107, 201}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheNativeModules(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeAddCacheNativeModules(t *testing.T) {
	convey.Convey("AddCacheNativeModules", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			values = map[int64]*v1.NativeModule{2: {ID: 2}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheNativeModules(c, values)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeDelCacheNativeModules(t *testing.T) {
	convey.Convey("DelCacheNativeModules", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{2}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheNativeModules(c, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeCacheNativeClicks(t *testing.T) {
	convey.Convey("CacheNativeClicks", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1, 103, 105}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheNativeClicks(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeAddCacheNativeClicks(t *testing.T) {
	convey.Convey("AddCacheNativeClicks", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			values = map[int64]*v1.NativeClick{1: {ID: 1}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheNativeClicks(c, values)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeDelCacheNativeClicks(t *testing.T) {
	convey.Convey("DelCacheNativeClicks", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheNativeClicks(c, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeCacheNativeDynamics(t *testing.T) {
	convey.Convey("CacheNativeDynamics", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheNativeDynamics(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeAddCacheNativeDynamics(t *testing.T) {
	convey.Convey("AddCacheNativeDynamics", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			values = map[int64]*v1.NativeDynamicExt{1: {ID: 1}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheNativeDynamics(c, values)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeDelCacheNativeDynamics(t *testing.T) {
	convey.Convey("DelCacheNativeDynamics", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheNativeDynamics(c, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeCacheNativeVideos(t *testing.T) {
	convey.Convey("CacheNativeVideos", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{102, 104, 105}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheNativeVideos(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeAddCacheNativeVideos(t *testing.T) {
	convey.Convey("AddCacheNativeVideos", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			values = map[int64]*v1.NativeVideoExt{1: {ID: 1}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheNativeVideos(c, values)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeDelCacheNativeVideos(t *testing.T) {
	convey.Convey("DelCacheNativeVideos", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheNativeVideos(c, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddCacheNativeMixtures(t *testing.T) {
	convey.Convey("AddCacheNativeMixtures", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = map[int64]*v1.NativeMixtureExt{1: {ID: 1}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheNativeMixtures(c, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCacheNativeMixtures(t *testing.T) {
	convey.Convey("CacheNativeMixtures", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CacheNativeMixtures(c, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDelCacheNativeMixtures(t *testing.T) {
	convey.Convey("DelCacheNativeMixtures", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheNativeMixtures(c, []int64{1})
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
