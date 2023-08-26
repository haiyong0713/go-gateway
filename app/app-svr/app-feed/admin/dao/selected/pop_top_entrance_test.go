package selected

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/selected"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Dao_GetAllEntrances(t *testing.T) {
	Convey("Test_Dao_GetAllEntrances", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			res, err := d.GetAllEntrances(c)
			for i, r := range res {
				fmt.Printf("%d: %+v \n", i, r)
			}
			So(err, ShouldBeNil)
		})
	})
}

func Test_Dao_UpdateEntrancesRank(t *testing.T) {
	Convey("Test_Dao_UpdateEntrancesRank", t, func() {
		var (
			c   = context.Background()
			res = []*selected.PopTopEntrance{
				{ID: 523, Rank: 3},
				{ID: 422, Rank: 5},
				{ID: 341, Rank: 2},
			}
		)
		Convey("When everything goes positive", func() {
			err := d.UpdateEntrancesRank(c, res)
			So(err, ShouldBeNil)
		})
	})
}
