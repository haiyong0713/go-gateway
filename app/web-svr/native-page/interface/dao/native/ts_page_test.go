package native

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go-gateway/app/web-svr/native-page/interface/api"
)

func TestPages(t *testing.T) {
	convey.Convey("TestRawNtTsPages", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("TsPageSave", func(convCtx convey.C) {
			rly, err := d.TsPageSave(c, &api.NativeTsPage{Pid: 1, Title: "title", ForeignID: 118, State: 1})
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Printf("RawNtPidToTsIDs %v", rly)
			})
		})
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			rly, err := d.RawNtTsPages(c, []int64{1, 2, 3})
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(rly)
				fmt.Printf("RawNtTsPages %s", string(str))
			})
		})
		convCtx.Convey("RawNtPidToTsID", func(convCtx convey.C) {
			rly, err := d.RawNtPidToTsID(c, 1)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Printf("RawNtPidToTsID %v", rly)
			})
		})
		convCtx.Convey("RawNtPidToTsIDs", func(convCtx convey.C) {
			rly, err := d.RawNtPidToTsIDs(c, []int64{1, 2, 3, 4})
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Printf("RawNtPidToTsIDs %v", rly)
			})
		})
	})
}
