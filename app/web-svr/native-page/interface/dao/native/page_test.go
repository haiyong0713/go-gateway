package native

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/web-svr/native-page/interface/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestNativeRawNativePages(t *testing.T) {
	convey.Convey("RawNativePages", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1, 17}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawNativePages(c, ids)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNative(t *testing.T) {
	convey.Convey("TestNative", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.RawNativeForeigns(c, []int64{1025}, 2)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("NtTsOnlineIDsSearch", func(convCtx convey.C) {
			rly, err := d.NtTsOnlineIDsSearch(c, 223)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(rly)
				fmt.Printf("%s", str)
			})
		})
		convCtx.Convey("NtTsOnlineIDsSearch", func(convCtx convey.C) {
			rly, err := d.NtTsUIDsSearch(c, 223)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				for _, v := range rly {
					fmt.Printf("%d,%d,end ", v.ID, v.Mtime)
				}
			})
		})
		//RawTitleSearch
		convCtx.Convey("RawNatTagIDExist", func(convCtx convey.C) {
			id, err := d.RawNatTagIDExist(c, 18674)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Printf("%d", id)
			})
		})
		convCtx.Convey("PageSave", func(convCtx convey.C) {
			id, err := d.PageSave(c, &api.NativePage{Title: "title", ForeignID: 11568, Type: 1, State: 1, RelatedUid: 223, FromType: 1, BgColor: "#uuyyie"})
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Printf("%d", id)
			})
		})
		convCtx.Convey("PageColorUpdate", func(convCtx convey.C) {
			err := d.PageUpdate(c, &api.NativePage{ID: 1, BgColor: "#cccccc"})
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("PageBind", func(convCtx convey.C) {
			err := d.PageBind(c, &api.NativePage{Title: "title", State: 0, ID: 1, ForeignID: 1126, Type: 2})
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
