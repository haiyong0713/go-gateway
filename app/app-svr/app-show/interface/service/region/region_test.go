package region

import (
	"context"
	"flag"
	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"path/filepath"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

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
func TestRegions(t *testing.T) {
	Convey("get Regions data", t, WithService(func(s *Service) {
		res, ver, err := s.Regions(context.TODO(), 0, 11111, "", "android", "", _initlanguage)
		So(res, ShouldNotBeEmpty)
		So(ver, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	}))
}

func TestRegionsList(t *testing.T) {
	Convey("get RegionsList data", t, WithService(func(s *Service) {
		res, ver, err := s.RegionsList(context.TODO(), 0, 11111, "", "android", "", _initlanguage, "region")
		So(res, ShouldNotBeEmpty)
		So(ver, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	}))
}
