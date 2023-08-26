package question

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestQuestionpoolKey(t *testing.T) {
	Convey("poolKey", t, func() {
		var (
			baseID = int64(0)
			poolID = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := poolKey(baseID, poolID)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestQuestionquesLimitKey(t *testing.T) {
	Convey("quesLimitKey", t, func() {
		var (
			mid    = int64(0)
			baseID = int64(0)
			day    = ""
		)
		Convey("When everything goes positive", func() {
			p1 := quesLimitKey(mid, baseID, day)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestQuestionPoolQuestionIDs(t *testing.T) {
	Convey("PoolQuestionIDs", t, func() {
		var (
			c      = context.Background()
			baseID = int64(0)
			poolID = int64(0)
			count  = int(3)
		)
		Convey("When everything goes positive", func() {
			data, err := d.PoolQuestionIDs(c, baseID, poolID, count)
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(data)
			})
		})
	})
}

func TestQuestionPoolIndexQuestionID(t *testing.T) {
	Convey("PoolIndexQuestionID", t, func() {
		var (
			c      = context.Background()
			baseID = int64(0)
			poolID = int64(0)
			index  = int64(0)
		)
		Convey("When everything goes positive", func() {
			id, err := d.PoolIndexQuestionID(c, baseID, poolID, index)
			Convey("Then err should be nil.id should not be nil.", func() {
				So(err, ShouldBeNil)
				So(id, ShouldNotBeNil)
			})
		})
	})
}

func TestQuestionIncrQuesLimit(t *testing.T) {
	Convey("IncrQuesLimit", t, func() {
		var (
			c      = context.Background()
			mid    = int64(0)
			baseID = int64(0)
			day    = ""
		)
		Convey("When everything goes positive", func() {
			err := d.IncrQuesLimit(c, mid, baseID, day)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestQuestionQuesLimit(t *testing.T) {
	Convey("QuesLimit", t, func() {
		var (
			c      = context.Background()
			mid    = int64(0)
			baseID = int64(0)
			day    = ""
		)
		Convey("When everything goes positive", func() {
			count, err := d.QuesLimit(c, mid, baseID, day)
			Convey("Then err should be nil.count should not be nil.", func() {
				So(err, ShouldBeNil)
				So(count, ShouldNotBeNil)
			})
		})
	})
}
