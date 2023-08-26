package question

import (
	"context"
	xtime "go-common/library/time"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestQuestionuserLogTableName(t *testing.T) {
	Convey("userLogTableName", t, func() {
		var (
			baseID = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := userLogTableName(baseID)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestQuestionRawBases(t *testing.T) {
	Convey("RawBases", t, func() {
		var (
			c  = context.Background()
			ts xtime.Time
		)
		Convey("When everything goes positive", func() {
			data, err := d.RawBases(c, ts)
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(data)
			})
		})
	})
}

func TestQuestionRawDetail(t *testing.T) {
	Convey("RawDetail", t, func() {
		var (
			c  = context.Background()
			id = int64(0)
		)
		Convey("When everything goes positive", func() {
			data, err := d.RawDetail(c, id)
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				So(data, ShouldNotBeNil)
			})
		})
	})
}

func TestQuestionRawDetails(t *testing.T) {
	Convey("RawDetails", t, func() {
		var (
			c   = context.Background()
			ids = []int64{1, 2, 3}
		)
		Convey("When everything goes positive", func() {
			data, err := d.RawDetails(c, ids)
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(data)
			})
		})
	})
}

func TestQuestionRawLastQuesLog(t *testing.T) {
	Convey("RawLastQuesLog", t, func() {
		var (
			c      = context.Background()
			mid    = int64(0)
			baseID = int64(0)
		)
		Convey("When everything goes positive", func() {
			data, err := d.RawLastQuesLog(c, mid, baseID)
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(data)
			})
		})
	})
}

func TestQuestionRawUserLogs(t *testing.T) {
	Convey("RawUserLogs", t, func() {
		var (
			c      = context.Background()
			mid    = int64(0)
			baseID = int64(0)
			poolID = int64(0)
		)
		Convey("When everything goes positive", func() {
			list, err := d.RawUserLogs(c, mid, baseID, poolID)
			Convey("Then err should be nil.list should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(list)
			})
		})
	})
}

func TestQuestionAddUserLog(t *testing.T) {
	Convey("AddUserLog", t, func() {
		var (
			c            = context.Background()
			mid          = int64(0)
			baseID       = int64(0)
			detailID     = int64(0)
			poolID       = int64(0)
			index        = int64(0)
			questionTime = time.Now()
		)
		Convey("When everything goes positive", func() {
			lastID, err := d.AddUserLog(c, mid, baseID, detailID, poolID, index, questionTime)
			Convey("Then err should be nil.lastID should not be nil.", func() {
				So(err, ShouldBeNil)
				So(lastID, ShouldNotBeNil)
			})
		})
	})
}

func TestQuestionUpUserLog(t *testing.T) {
	Convey("UpUserLog", t, func() {
		var (
			c          = context.Background()
			isRight    = int(0)
			answerTime = time.Now()
			id         = int64(0)
			baseID     = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.UpUserLog(c, isRight, answerTime, id, baseID)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
