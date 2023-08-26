package web

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	s *Service
	c = context.Background()
)

func init() {
	dir, _ := filepath.Abs("../../cmd/feed-admin-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	New(conf.Conf)
	s = New(conf.Conf)
}

func TestService_OgvList(t *testing.T) {
	convey.Convey("AllGroup", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := s.AllGroup()
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}
