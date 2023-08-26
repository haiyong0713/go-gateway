package native

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestModules(t *testing.T) {
	convey.Convey("TestRawNtTsPages", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		//convCtx.Convey("TsModuleSave", func(convCtx convey.C) {
		//	err := d.TsModuleSave(c, []*dymmdl.NativeTsModuleExt{{Category: 19, TsID: 1, State: 1, Rank: 1, Meta: "https://www.bilibili.com", Width: 100, Length: 22, PType: 1, Ukey: "image"}, {Category: 1, TsID: 1, State: 1, Rank: 2, Remark: "remark", PType: 1, Ukey: "remark"}}, 1)
		//	convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
		//		convCtx.So(err, convey.ShouldBeNil)
		//	})
		//})
		convCtx.Convey("RawNtTsModulesExt", func(convCtx convey.C) {
			rly, err := d.RawNtTsModulesExt(c, []int64{1, 2, 3})
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(rly)
				fmt.Printf("RawNtTsPages %s", string(str))
			})
		})
		convCtx.Convey("NtTsModuleIDSearch", func(convCtx convey.C) {
			rly, err := d.NtTsModuleIDSearch(c, 2)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Printf("RawNtPidToTsID %v", rly)
			})
		})
	})
}
