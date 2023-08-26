package common

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/model/common"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	c   = context.Background()
	bmC *bm.Context
)

var (
	svr *Service
)

func init() {
	dir, _ := filepath.Abs("../../cmd/feed-admin-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	svr = New(conf.Conf)
	time.Sleep(time.Second)
	bmC = &bm.Context{}

}

func WithService(f func(s *Service)) func() {
	return func() {
		f(svr)
	}
}

func TestService_FilterMatch(t *testing.T) {
	Convey("test service filterMatch", t, WithService(func(s *Service) {
		param := &common.Log{
			Type:      12,
			Starttime: "2019-07-01 00:00:00",
			Endtime:   "2019-07-31 23:59:59",
			Pn:        20,
			Ps:        1,
			Uname:     "",
		}
		res, err := s.LogAction(c, param)
		if err != nil {
			fmt.Println(err)
			return
		}
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
