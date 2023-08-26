package selected

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Dao_GetCronJobLock(t *testing.T) {
	Convey("Test_Dao_GetCronJobLock", t, func() {
		var (
			c           = context.Background()
			cronJobName = "publishWeekly"
		)
		Convey("When everything goes positive", func() {
			lock, err := d.GetCronJobLock(c, cronJobName)
			So(err, ShouldBeNil)
			So(lock, ShouldBeTrue)
		})
	})
}
