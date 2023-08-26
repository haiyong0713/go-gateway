package dao

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoNewDB(t *testing.T) {
	Convey("NewDB", t, func() {
		Convey("When everything goes positive", func() {
			db, err := NewDB()
			Convey("Then err should be nil.db should not be nil.", func() {
				So(err, ShouldBeNil)
				So(db, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoDB(t *testing.T) {
	Convey("DB", t, func() {
		Convey("When everything goes positive", func() {
			p1 := d.DB()
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}
