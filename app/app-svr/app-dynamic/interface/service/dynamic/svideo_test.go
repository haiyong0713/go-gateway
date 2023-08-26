package dynamic

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/app-dynamic/interface/api"
	dymdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"
)

func TestService_SVideo(t *testing.T) {
	var (
		req    = &api.SVideoReq{Oid: 199, Type: 3}
		mid    = int64(88895133)
		header = &dymdl.Header{}
		vm     = &dymdl.VideoMate{}
	)
	Convey("SVideo", t, func() {
		res, err := s.SVideo(context.TODO(), req, mid, header, vm, 0)
		fmt.Printf("%+v", res)
		So(err, ShouldBeNil)
	})
}
