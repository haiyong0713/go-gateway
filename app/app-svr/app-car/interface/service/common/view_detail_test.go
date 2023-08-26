package common

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
)

func TestService_viewPlaylistSerial(t *testing.T) {
	convey.Convey("TestService_viewPlaylistSerial", t, WithService(func(s *Service) {
		var (
			ctx = context.Background()
			req = &commonmdl.ViewV2SerialReq{
				Otype:    "fm_serial",
				Oid:      333,
				Mid:      8432042,
				Buvid:    "XY0432JF932N40SJ21U9S0320",
				PageNext: "",
				PagePre:  "",
				Ps:       10,
				Aid:      0,
			}
		)
		serial, err := s.viewPlaylistSerial(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(serial, convey.ShouldNotBeNil)
		marshal, _ := json.Marshal(serial)
		_, _ = convey.Printf("res:%s\n", string(marshal))
	}))
}

func Test_parseUrl(t *testing.T) {
	convey.Convey("TestService_FmLike", t, WithService(func(s *Service) {
		//got, got1 := parseUrl("https://www.bilibili.com/bangumi/play/ep459343")
		got, got1 := parseUrl("https://www.bilibili.com/bangumi/play/ss2033")
		//got, got1 := parseUrl("https://www.bilibili.com/video/BV1Aa411E7ZR?p=1&share_medium=iphone&share_plat=ios&share_session_id=8A795708-1E26-451F-A11E-BC9A94361FA6&share_source=COPY&share_tag=s_i&timestamp=1653988528&unique_k=7bzjaP7")
		fmt.Println(got)
		fmt.Println(got1)
	}))
}
