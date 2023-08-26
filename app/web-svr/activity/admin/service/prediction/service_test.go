package prediction

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/web-svr/activity/admin/conf"
	premdl "go-gateway/app/web-svr/activity/admin/model/prediction"

	. "github.com/smartystreets/goconvey/convey"
)

var svr *Service

func init() {
	dir, _ := filepath.Abs("../../cmd/activity-admin-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	svr = New(conf.Conf)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		Reset(func() {})
		f(svr)
	}
}
func TestService_ItemAdd(t *testing.T) {
	arg := make([]*premdl.ItemAdd, 0, 101)
	for i := 0; i < 30; i++ {
		desc := fmt.Sprintf("飞猪侠%d", i)
		arg = append(arg, &premdl.ItemAdd{Sid: 16, Pid: 3, Desc: desc, Image: "", State: 1})
	}
	Convey("service test", t, WithService(func(s *Service) {
		err := s.ItemAdd(context.Background(), arg)
		So(err, ShouldBeNil)
	}))
}
