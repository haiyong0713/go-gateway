package show

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"go-common/library/conf/paladin.v2"

	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/model"

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

func TestIndex(t *testing.T) {
	Convey("get Index data", t, WithService(func(s *Service) {
		res := s.Index(context.TODO(), 0, model.PlatIPhone, 0, "", "", "", "", "", "iphone", "phone", _initlanguage, "", false, time.Now())
		So(res, ShouldNotBeEmpty)
	}))
}

func TestChange(t *testing.T) {
	Convey("get Change data", t, WithService(func(s *Service) {
		res := s.Change(context.TODO(), 1, 1, 1, 1, "", "", "", "", "")
		So(res, ShouldNotBeEmpty)
	}))
}

func TestBangumiChange(t *testing.T) {
	Convey("get BangumiChange data", t, WithService(func(s *Service) {
		res := s.BangumiChange(context.TODO(), 1, 1)
		So(res, ShouldNotBeEmpty)
	}))
}

func TestFeedIndex(t *testing.T) {
	Convey("get FeedIndex data", t, WithService(func(s *Service) {
		res := s.FeedIndex(context.TODO(), 1, 1, 1, 1, 1, "", "", "", "", time.Now())
		So(res, ShouldNotBeEmpty)
	}))
}

func TestCardSet(t *testing.T) {
	// cardSetChange
	Convey("get FeedIndex data", t, WithService(func(s *Service) {
		cardm, aids, upid := s.cardSetChange(context.TODO(), 97, 98, 99)
		str, _ := json.Marshal(cardm)
		Println(string(str))
		Println(aids, upid)
	}))
}

func TestFeedIndex2(t *testing.T) {
	Convey("get FeedIndex data", t, WithService(func(s *Service) {
		res, ver, conf, err := s.FeedIndex2(context.TODO(), 1, 20, 0, 15, 1, 1, "", "", "", "", time.Now())
		println(ver)
		println(conf)
		println(err)
		So(res, ShouldNotBeEmpty)
	}))
}

func TestSerieShow(t *testing.T) {
	Convey("get SerieShow data", t, WithService(func(s *Service) {
		res, _ := s.SerieShow(context.TODO(), "weekly_selected", 29, 0, "", "")
		So(res, ShouldNotBeEmpty)
	}))
}

func TestFeedIndexSvideo(t *testing.T) {
	Convey("TestFeedIndexSvideo", t, WithService(func(s *Service) {
		res, err := s.FeedIndexSvideo(context.TODO(), 341, 12)
		fmt.Println(res)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	}))
}

func TestAggrSvideo(t *testing.T) {
	Convey("TestAggrSvideo", t, WithService(func(s *Service) {
		res, err := s.AggrSvideo(context.TODO(), 188, 0)
		fmt.Println(res)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	}))
}

func TestAllSeries(t *testing.T) {
	Convey("get AllSeries data", t, WithService(func(s *Service) {
		res, _ := s.AllSeries(context.TODO(), "weekly_selected")
		So(res, ShouldNotBeEmpty)
	}))
}
