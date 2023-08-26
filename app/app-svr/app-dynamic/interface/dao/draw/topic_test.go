package dao

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoGetHotTopicTopK(t *testing.T) {
	Convey("GetHotTopicTopK", t, func() {
		var (
			ctx = context.Background()
			k   = int(10)
		)
		Convey("When everything goes positive", func() {
			topics, err := d.GetHotTopicTopK(ctx, k)
			Convey("Then err should be nil.topics should not be nil.", func() {
				So(err, ShouldBeNil)
				So(topics, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchTopic(t *testing.T) {
	Convey("SearchTopic", t, func() {
		var (
			ctx      = context.Background()
			word     = "bili"
			page     = int(0)
			pageSize = int(20)
		)
		Convey("When everything goes positive", func() {
			topics, hasMore, err := d.SearchTopic(ctx, word, page, pageSize)
			Convey("Then err should be nil.topics should not be nil.", func() {
				So(err, ShouldBeNil)
				So(topics, ShouldNotBeNil)
				So(hasMore, ShouldBeTrue)
			})
		})
	})
}
