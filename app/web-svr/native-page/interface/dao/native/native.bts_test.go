package native

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNativeNativePages(t *testing.T) {
	convey.Convey("NativePages", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{17, 9999}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.NativePages(c, ids)
			fmt.Printf("%v", res)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeNativeForeigns(t *testing.T) {
	convey.Convey("NativeForeigns", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			ids      = []int64{10357}
			pageType = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.NativeForeigns(c, ids, pageType)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeNativeModules(t *testing.T) {
	convey.Convey("NativeModules", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{104, 105, 106}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.NativeModules(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeNativeClicks(t *testing.T) {
	convey.Convey("NativeClicks", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{104, 105, 106}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.NativeClicks(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeNativeDynamics(t *testing.T) {
	convey.Convey("NativeDynamics", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{105, 102, 103}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.NativeDynamics(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeNativeVideos(t *testing.T) {
	convey.Convey("NativeVideos", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1, 2, 78, 90}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.NativeVideos(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativeMixtures(t *testing.T) {
	convey.Convey("NativeMixtures", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1, 2, 78, 90}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.NativeMixtures(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// NativeTabModules
func TestNativeTabModules(t *testing.T) {
	convey.Convey("NativeTabModules", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1, 2, 9999}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.NativeTabModules(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestNativeTabSort(t *testing.T) {
	convey.Convey("NativeTabSort", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.NativeTabSort(c, 2)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestCacheNativeTabSort(t *testing.T) {
	convey.Convey("CacheNativeTabSort", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.CacheNativeTabSort(c, 2)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestCacheNativeTabModules(t *testing.T) {
	convey.Convey("CacheNativeTabModules", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{2}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.CacheNativeTabModules(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestModuleIDs(t *testing.T) {
	convey.Convey("ModuleIDs", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.ModuleIDs(c, 4, 1, 0, -1)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestNatMixIDs(t *testing.T) {
	convey.Convey("NatMixIDs", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.NatMixIDs(c, 1, 4, 2, 3)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestNatAllMixIDs(t *testing.T) {
	convey.Convey("NatAllMixIDs", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.NatAllMixIDs(c, 1, 2, 4)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestPartPids(t *testing.T) {
	convey.Convey("PartPids", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.PartPids(c, 81, 2, 3)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestNativePart(t *testing.T) {
	convey.Convey("NativePart", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.NativePart(c, []int64{699})
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestCacheNativePart(t *testing.T) {
	convey.Convey("CacheNativePart", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.CacheNativePart(c, []int64{6})
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestNativeTabBind(t *testing.T) {
	convey.Convey("NativeTabBind", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.NativeTabBind(c, []int64{198, 197, 196}, 3)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestNativeTabs(t *testing.T) {
	convey.Convey("NativeTabs", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reT, err := d.NativeTabs(c, []int64{2})
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestCacheTs(t *testing.T) {
	convey.Convey("CacheNativePart", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("NtTsPages", func(convCtx convey.C) {
			reT, err := d.NtTsPages(c, []int64{1, 2, 3, 9999})
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
		convCtx.Convey("NtTsUIDs", func(convCtx convey.C) {
			reT, err := d.NtTsUIDs(c, 2773, 1, 3)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
		convCtx.Convey("NtTsOnlineIDs", func(convCtx convey.C) {
			reT, err := d.NtTsOnlineIDs(c, 223, 1, 2)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
		convCtx.Convey("NtTsModuleIDs", func(convCtx convey.C) {
			reT, err := d.NtTsModuleIDs(c, 1, 1, 2)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
		convCtx.Convey("NtTsModulesExt", func(convCtx convey.C) {
			reT, err := d.NtTsModulesExt(c, []int64{1, 2})
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
		convCtx.Convey("NtPidToTsID", func(convCtx convey.C) {
			reT, err := d.NtPidToTsID(c, 1)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
		convCtx.Convey("NtPidToTsIDs", func(convCtx convey.C) {
			reT, err := d.NtPidToTsIDs(c, []int64{1, 2, 3})
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(reT)
				fmt.Printf("%s", str)
			})
		})
	})
}
