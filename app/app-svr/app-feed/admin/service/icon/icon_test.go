package icon

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
	"go-gateway/app/app-svr/app-feed/admin/model/icon"

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

func TestService_IconList(t *testing.T) {
	Convey("IconList", t, WithService(func(s *Service) {
		param := &icon.ListParam{
			Ps: 15,
			Pn: 1,
		}
		res, err := s.IconList(c, param)
		if err != nil {
			fmt.Println(err)
			return
		}
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}

func TestService_IconDetail(t *testing.T) {
	Convey("IconDetail", t, WithService(func(s *Service) {
		res, err := s.IconDetail(c, 3)
		if err != nil {
			fmt.Println(err)
			return
		}
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}

func TestService_IconOpt(t *testing.T) {
	Convey("IconOpt", t, WithService(func(s *Service) {
		err := s.IconOpt(c, 3, 158747, 1, "TEST")
		if err != nil {
			fmt.Println(err)
		}
	}))
}

func TestService_IconSave(t *testing.T) {
	Convey("IconSave", t, WithService(func(s *Service) {
		arg := &icon.IconSaveParam{
			ID:          1,
			Module:      `[{"plat":0,"oid":1},{"plat":1,"oid":2}]`,
			Icon:        "icon.icon.",
			GlobalRed:   0,
			EffectGroup: 3,
			Stime:       xtime.Time(1574666727),
			Etime:       xtime.Time(1574666740),
		}
		err := s.IconSave(c, arg, "test", 158747)
		if err != nil {
			fmt.Println(err)
		}
	}))
}

func TestService_IconModule(t *testing.T) {
	Convey("IconModule", t, WithService(func(s *Service) {
		res, err := s.IconModule(c, 1)
		if err != nil {
			fmt.Println(err)
			return
		}
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
