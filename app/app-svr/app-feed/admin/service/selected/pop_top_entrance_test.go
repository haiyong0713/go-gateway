package selected

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Dao_updateEntranceRank(t *testing.T) {
	Convey("Test_Dao_updateEntranceRank", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			err := svf.updateEntranceRank(c)
			So(err, ShouldBeNil)
		})
	})
}

func Test_Dao_GetAllEntrances(t *testing.T) {
	Convey("Test_Dao_GetAllEntrances", t, func() {

		Convey("When everything goes positive", func() {
			err := svf.rollbackEntranceRank()
			So(err, ShouldBeNil)
		})
	})
}
