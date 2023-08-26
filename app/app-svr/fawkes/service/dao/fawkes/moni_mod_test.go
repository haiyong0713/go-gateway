package fawkes

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/fawkes/service/model/mod"
)

func TestDao_ActiveUserCount(t *testing.T) {
	Convey("TestDao_ActiveUserCount", t, func() {
		appVer := []map[mod.Condition]int64{
			{mod.ConditionGe: 6680010, mod.ConditionLe: 6780010},
			{mod.ConditionGt: 6790010, mod.ConditionLt: 6800010},
		}
		count, err := d.ActiveUserCount(context.Background(), "w19e", appVer, time.Now(), time.Now().Add(5*time.Minute))
		So(err, ShouldBeNil)
		So(count, ShouldNotBeEmpty)
	})
}

func TestDao_ModDownloadSizeSum(t *testing.T) {
	Convey("TestDao_ModDownloadSizeSum", t, func() {
		count, err := d.ModDownloadSizeSum(context.Background(), "w19e", "pool_name", "mod_name", time.Now(), time.Now().Add(5*time.Minute))
		So(err, ShouldBeNil)
		So(count, ShouldNotBeEmpty)
	})
}
