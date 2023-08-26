package question

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/model/question"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestQuestionCacheDetail(t *testing.T) {
	Convey("CacheDetail", t, func() {
		var (
			c  = context.Background()
			id = int64(0)
		)
		Convey("When everything goes positive", func() {
			res, err := d.CacheDetail(c, id)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})
	})
}

func TestQuestionAddCacheDetail(t *testing.T) {
	Convey("AddCacheDetail", t, func() {
		var (
			c   = context.Background()
			id  = int64(0)
			val = &question.Detail{}
		)
		Convey("When everything goes positive", func() {
			err := d.AddCacheDetail(c, id, val)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestQuestionDelCacheDetail(t *testing.T) {
	Convey("DelCacheDetail", t, func() {
		var (
			c  = context.Background()
			id = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.DelCacheDetail(c, id)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestQuestionCacheDetails(t *testing.T) {
	Convey("CacheDetails", t, func() {
		var (
			c   = context.Background()
			ids = []int64{}
		)
		Convey("When everything goes positive", func() {
			res, err := d.CacheDetails(c, ids)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestQuestionAddCacheDetails(t *testing.T) {
	Convey("AddCacheDetails", t, func() {
		var (
			c      = context.Background()
			values map[int64]*question.Detail
		)
		Convey("When everything goes positive", func() {
			err := d.AddCacheDetails(c, values)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestQuestionCacheLastQuesLog(t *testing.T) {
	Convey("CacheLastQuesLog", t, func() {
		var (
			c      = context.Background()
			id     = int64(0)
			baseID = int64(0)
		)
		Convey("When everything goes positive", func() {
			res, err := d.CacheLastQuesLog(c, id, baseID)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestQuestionAddCacheLastQuesLog(t *testing.T) {
	Convey("AddCacheLastQuesLog", t, func() {
		var (
			c      = context.Background()
			id     = int64(0)
			val    = &question.UserAnswerLog{}
			baseID = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.AddCacheLastQuesLog(c, id, val, baseID)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
