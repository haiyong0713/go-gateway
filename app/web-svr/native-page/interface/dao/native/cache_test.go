package native

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNativenativePageKey(t *testing.T) {
	convey.Convey("nativePageKey", t, func(convCtx convey.C) {
		var (
			id = int64(17)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := nativePageKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestNativenativeForeignKey(t *testing.T) {
	convey.Convey("nativeForeignKey", t, func(convCtx convey.C) {
		var (
			id       = int64(17)
			pageType = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := nativeForeignKey(id, pageType)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestNativenativeModuleKey(t *testing.T) {
	convey.Convey("nativeModuleKey", t, func(convCtx convey.C) {
		var (
			id = int64(104)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := nativeModuleKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestNativenativeClickKey(t *testing.T) {
	convey.Convey("nativeClickKey", t, func(convCtx convey.C) {
		var (
			id = int64(104)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := nativeClickKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestNativenativeDynamicKey(t *testing.T) {
	convey.Convey("nativeDynamicKey", t, func(convCtx convey.C) {
		var (
			id = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := nativeDynamicKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestNativenativeVideoKey(t *testing.T) {
	convey.Convey("nativeVideoKey", t, func(convCtx convey.C) {
		var (
			id = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := nativeVideoKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}
