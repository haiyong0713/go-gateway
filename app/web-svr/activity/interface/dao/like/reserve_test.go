package like

import (
	"context"
	"strings"
	"testing"

	lmdl "go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/smartystreets/goconvey/convey"
)

func TestAddReserve(t *testing.T) {
	convey.Convey("AddReserve", t, func(convCtx convey.C) {
		var (
			c = context.Background()
			m = &lmdl.ActReserve{Sid: 10627, Mid: 1548785, Num: 1, State: 1, IPv6: []byte{}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			sid, err := d.AddReserve(c, m)
			if err != nil {
				if strings.Contains(err.Error(), "Duplicate entry") {
					err = nil
				}
			}
			convCtx.Convey("Then err should be nil.sid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Println(sid)
			})
		})
	})
}

// UpReserve
func TestUpReserve(t *testing.T) {
	convey.Convey("UpReserve", t, func(convCtx convey.C) {
		var (
			c = context.Background()
			m = &lmdl.ActReserve{Sid: 10529, Mid: 1548785, Num: 1, State: 0, IPv6: []byte{}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.UpReserve(c, m)
			convCtx.Convey("Then err should be nil.sid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCancelReserve(t *testing.T) {
	convey.Convey("CancelReserve", t, func(convCtx convey.C) {
		var (
			c = context.Background()
			m = &lmdl.ActReserve{Sid: 10629, Mid: 1548785, Num: 0, State: 0, IPv6: []byte{}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.CancelReserve(c, m)
			convCtx.Convey("Then err should be nil.sid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRawReserveOnly(t *testing.T) {
	convey.Convey("RawReserveOnly", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawReserveOnly(c, 10629, 1548785)
			convCtx.Convey("Then err should be nil.sid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%v", res)
			})
		})
	})
}
