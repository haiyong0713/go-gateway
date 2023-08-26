package menu

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
	"go-gateway/app/app-svr/app-feed/admin/model/menu"

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

func TestService_MenuTabList(t *testing.T) {
	Convey("MenuTabList", t, WithService(func(s *Service) {
		param := &menu.ListParam{
			Ps:    15,
			Pn:    1,
			TabID: 1,
		}
		res, err := s.MenuTabList(c, param)
		if err != nil {
			fmt.Println(err)
			return
		}
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}

func TestService_MenuTabOperate(t *testing.T) {
	Convey("MenuTabOperate", t, WithService(func(s *Service) {
		err := s.MenuTabOperate(c, 1, 158747, 1, "TEST")
		if err != nil {
			fmt.Println(err)
		}
	}))
}

func TestService_MenuTabSave(t *testing.T) {
	Convey("MenuTabSave", t, WithService(func(s *Service) {
		arg := &menu.TabSaveParam{
			ID:              5,
			TabID:           3,
			Type:            0,
			Attribute:       1,
			InactiveType:    2,
			InactiveIcon:    "http://uat-i0.hdslb.com/bfs/archive/3c1915c1e09946c4449142d7aa2bf3c0a05fc4cf.jpg_80x50.jpg",
			Inactive:        0,
			ActiveIcon:      "http://uat-i0.hdslb.com/bfs/archive/3c1915c1e09946c4449142d7aa2bf3c0a05fc4cf.jpg_80x50.jpg",
			ActiveType:      3,
			Active:          1,
			FontColor:       "#F8E71C",
			BarColor:        1,
			Stime:           xtime.Time(1571133040),
			Etime:           xtime.Time(1575133040),
			Limit:           `[{"type":0,"plat":1,"build":1548,"conditions":"ne"},{"type":0,"plat":5,"build":1548,"conditions":"lt"}]`,
			TabColorBegin:   1,
			TabTopColor:     "#F8E72C",
			TabBottomColor:  "#F8E72C",
			PageTopColor:    "#F8E72C",
			PageBottomColor: "#F8E72C",
			BgImage1:        "http://uat-i0.hdslb.com/bfs/archive/3c1915c1e09946c4449142d7aa2bf3c0a05fc4cf.jpg_80x50.jpg",
			BgImage2:        "http://uat-i0.hdslb.com/bfs/archive/3c1915c1e09946c4449142d7aa2bf3c0a05fc4cf.jpg_80x50.jpg",
		}
		res, err := s.MenuTabSave(c, arg, "test", 158747)
		if err != nil {
			fmt.Println(err)
		}
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
