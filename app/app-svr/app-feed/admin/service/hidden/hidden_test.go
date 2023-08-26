package hidden

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	bm "go-common/library/net/http/blademaster"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/model/hidden"

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

func TestService_HiddenList(t *testing.T) {
	Convey("HiddenList", t, WithService(func(s *Service) {
		param := &hidden.ListParam{
			Ps: 15,
			Pn: 1,
		}
		res, err := s.HiddenList(c, param)
		if err != nil {
			fmt.Println(err)
			return
		}
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}

func TestService_HiddenOpt(t *testing.T) {
	Convey("HiddenOpt", t, WithService(func(s *Service) {
		err := s.HiddenOpt(c, 1, 158747, 1, "TEST")
		if err != nil {
			fmt.Println(err)
		}
	}))
}

func TestService_HiddenSave(t *testing.T) {
	Convey("HiddenSave", t, WithService(func(s *Service) {
		arg := &hidden.HiddenSaveParam{
			SID:     1,
			RID:     3,
			Channel: "oppo",
			Stime:   xtime.Time(1574666727),
			Etime:   xtime.Time(1574666740),
			Limit:   `[{"plat":1,"build":1548,"conditions":"ne"},{"plat":5,"build":1548,"conditions":"lt"}]`,
		}
		err := s.HiddenSave(c, arg, "test", 158747)
		if err != nil {
			fmt.Println(err)
		}
	}))
}

func TestService_EntranceSearch(t *testing.T) {
	Convey("HiddenDetail", t, WithService(func(s *Service) {
		res, err := s.EntranceSearch(c, 2, 2)
		if err != nil {
			fmt.Println(err)
			return
		}
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
