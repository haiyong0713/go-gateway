package fawkes

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestGitlabProjectID test GitlabProjectID.
func TestGitlabProjectID(t *testing.T) {
	convey.Convey("GitlabProjectID", t, func(convCtx convey.C) {
		var c = context.Background()
		convCtx.Convey("When everything goes correct", func(convCtx convey.C) {
			_, err := d.GitlabProjectID(c, "56ba")
			convCtx.Convey("Then err should be nil. GitlabProjectID should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestGitlabJobInfo test GitlabJobInfo.
func TestGitlabJobInfo(t *testing.T) {
	convey.Convey("GitlabJobInfo", t, func(convCtx convey.C) {
		var c = context.Background()
		convCtx.Convey("When everything goes correct", func(convCtx convey.C) {
			_, err := d.GitlabJobInfo(c, 123)
			convCtx.Convey("Then err should be nil. GitlabJobInfo should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestHotfixJobInfo test HotfixJobInfo
func TestHotfixJobInfo(t *testing.T) {
	convey.Convey("HotfixJobInfo", t, func(convCtx convey.C) {
		var c = context.Background()
		convCtx.Convey("When everything goes correct", func(convCtx convey.C) {
			_, err := d.HotfixJobInfo(c, 123)
			convCtx.Convey("Then err should be nil. HotfixJobInfo should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
