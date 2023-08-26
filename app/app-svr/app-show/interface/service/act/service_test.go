package act

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"go-common/library/conf/paladin.v2"
	pb "go-gateway/app/app-svr/app-show/interface/api"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/model/act"

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
	dir, _ := filepath.Abs("../../cmd/app-show-test.toml")
	flag.Set("conf", dir)
	cfg, err := confInit()
	if err != nil {
		panic(err)
	}
	s = New(cfg)
	time.Sleep(time.Second)
}

func confInit() (*conf.Config, error) {
	err := paladin.Init()
	if err != nil {
		return nil, err
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("app-show.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func TestRankShow(t *testing.T) {
	Convey("get RankShow data", t, WithService(func(s *Service) {
		res, _ := s.ActIndex(context.Background(), &act.ParamActIndex{PageID: 4102, Offset: 0, Ps: 10}, 15555181)
		str, _ := json.Marshal(res)
		fmt.Printf(" res %v", string(str))
		So(res, ShouldNotBeEmpty)
	}))
}

func TestSupernatant(t *testing.T) {
	Convey("Supernatant", t, WithService(func(s *Service) {
		res, _ := s.Supernatant(context.Background(), &act.ParamSupernatant{ConfModuleID: 52915, Offset: 0, Ps: 15, LastIndex: 3, PageID: 4102}, 0)
		str, _ := json.Marshal(res)
		fmt.Printf(" res %v", string(str))
		So(res, ShouldNotBeEmpty)
	}))
}

func TestActLiked(t *testing.T) {
	Convey("get RankShow data", t, WithService(func(s *Service) {
		res, _ := s.ActLiked(context.Background(), &act.ParamActLike{Sid: 10461, Lid: 3123, Score: 3}, 15555185)
		So(res, ShouldNotBeEmpty)
	}))
}

func TestActDetail(t *testing.T) {
	Convey("get RankShow data", t, WithService(func(s *Service) {
		res, _ := s.ActDetail(context.Background(), &act.ParamActDetail{ModuleID: 5343, MobiApp: "iphone", Build: 8670})
		So(res, ShouldNotBeEmpty)
	}))
}

func TestLikeList(t *testing.T) {
	Convey("get RankShow data", t, WithService(func(s *Service) {
		res, _ := s.LikeList(context.Background(), &act.ParamLike{Sid: 10436, SortType: 1, Pn: 1, Ps: 10}, 15555185)
		So(res, ShouldNotBeEmpty)
	}))
}

func TestActTab(t *testing.T) {
	Convey("ActTab", t, WithService(func(s *Service) {
		res, _ := s.ActTab(context.Background(), &act.ParamActTab{})
		So(res, ShouldNotBeEmpty)
		str, _ := json.Marshal(res)
		fmt.Printf("%s", string(str))
	}))
}

func TestActNativeTab(t *testing.T) {
	Convey("ActNativeTab", t, WithService(func(s *Service) {
		res, _ := s.ActNativeTab(context.Background(), &pb.ActNativeTabReq{Pids: []int64{562}, Category: 1})
		So(res, ShouldNotBeEmpty)
		str, _ := json.Marshal(res)
		fmt.Printf("%s", string(str))
	}))
}

func TestMenuTab(t *testing.T) {
	Convey("MenuTab", t, WithService(func(s *Service) {
		res, _ := s.MenuTab(context.Background(), &act.ParamMenuTab{PageID: 1295, Offset: 0, Ps: 42}, 15555181)
		So(res, ShouldNotBeEmpty)
	}))
}
