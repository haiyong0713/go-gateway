package fawkes

import (
	"context"
	"testing"

	xhttp "go-common/library/net/http/blademaster"

	v2 "git.bilibili.co/bapis/bapis-go/push/service/broadcast/v2"
	. "github.com/smartystreets/goconvey/convey"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
)

func TestPushBuvids(t *testing.T) {
	httpClient := xhttp.NewClient(d.c.HTTPClient)
	b := BroadcastProxyHTTP{
		httpClient: httpClient,
	}

	in := v2.PushBuvidsReq{
		Opts:                 nil,
		Msg:                  nil,
		Buvids:               nil,
		Token:                "",
		Expired:              0,
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}

	Convey("", t, func() {
		_, err := b.PushBuvids(context.Background(), &in)
		if err != nil {
			return
		}
	})
}

func TestUnmarshal(t *testing.T) {
	m := make(map[string]interface{})
	m["msg"] = 11111
	m["session"] = []string{"hhhhh"}
	m["pushed"] = 1

	Convey("", t, func() {
		var a int
		a = 1
		var b int64
		b = 1
		So(appmdl.AppServerZone_Abroad, ShouldEqual, a)
		So(appmdl.AppServerZone_Abroad, ShouldEqual, b)

	})
}
