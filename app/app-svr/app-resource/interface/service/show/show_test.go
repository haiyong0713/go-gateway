package show

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-resource/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	s *Service
)

func WithService(f func(s *Service)) func() {
	return func() {
		f(s)
	}
}

func init() {
	dir, _ := filepath.Abs("../../cmd/app-resource-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	s = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestTabs(t *testing.T) {
	Convey("get Tabs data", t, WithService(func(s *Service) {
		res, _, _, err := s.Tabs(context.TODO(), 0, 0, 0, "xxxx", "android", "android", "hans", "oppo", 27515240)
		rr, _ := json.Marshal(res)
		fmt.Printf("%s", rr)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	}))
}
