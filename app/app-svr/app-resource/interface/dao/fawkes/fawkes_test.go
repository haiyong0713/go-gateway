package fawkes

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestVersions(t *testing.T) {
	Convey("should get version", t, func() {
		_, err := d.Versions(context.Background())
		Convey("Then err should not be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestUpgradConfig(t *testing.T) {
	Convey("should get upgrade", t, func() {
		_, err := d.UpgradConfig(context.Background())
		Convey("Then err should not be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestPacks(t *testing.T) {
	Convey("should get packs", t, func() {
		_, err := d.Packs(context.Background())
		Convey("Then err should not be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestPatch(t *testing.T) {
	Convey("should get patch", t, func() {
		_, err := d.Patch(context.Background())
		Convey("Then err should not be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestFilterConfig(t *testing.T) {
	Convey("should get filter config", t, func() {
		_, err := d.FilterConfig(context.Background())
		Convey("Then err should not be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestAppChannel(t *testing.T) {
	Convey("should get app channel", t, func() {
		_, err := d.AppChannel(context.Background())
		Convey("Then err should not be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestFlowConfig(t *testing.T) {
	Convey("should get flow", t, func() {
		_, err := d.FlowConfig(context.Background())
		Convey("Then err should not be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestHfUpgrade(t *testing.T) {
	Convey("Should get hotfix upgrade information", t, func() {
		_, err := d.HfUpgrade(context.Background())
		Convey("Then err should not be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}
