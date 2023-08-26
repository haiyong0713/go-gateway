package show

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-job/job/conf"
	"go-gateway/app/app-svr/app-job/job/model/show"

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
	dir, _ := filepath.Abs("../../cmd/app-job-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	s = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestPing(t *testing.T) {
	Convey("get Ping data", t, WithService(func(s *Service) {
		err := s.Ping(context.TODO())
		So(err, ShouldBeNil)
	}))
}

func TestService_TreatAIData(t *testing.T) {
	Convey("TestService_TreatAIData", t, WithService(func(s *Service) {
		err := s.treatAIData(context.Background(), []byte(`{"code":0,"num":3,"list":[{"aid":10112051,"mid":123938419,"score":469308,"desc":""},{"aid":10112049,"mid":123938419,"score":469308,"desc":""},{"aid":10112048,"mid":123938419,"score":469308,"desc":""}]}`))
		So(err, ShouldBeNil)
	}))
}

func TestService_AIAlert(t *testing.T) {
	Convey("TestService_AIAlert", t, WithService(func(s *Service) {
		err := s.alertAI("weekly_selected")
		So(err, ShouldBeNil)
	}))
}

func TestService_AlertAuditor(t *testing.T) {
	Convey("TestService_AlertAuditor", t, WithService(func(s *Service) {
		err := s.alertAuditor("weekly_selected")
		So(err, ShouldBeNil)
	}))
}

func TestService_RefArchiveHonorSend(t *testing.T) {
	Convey("RefArchiveHonorSend", t, WithService(func(s *Service) {
		err := s.sendArcHonor(context.Background(), &show.HonorMsg{
			Action: "update",
			Aid:    10113778,
			Type:   2,
			Url:    "https://www.bilibili.com/h5/weekly-recommend?num=27&navhide=1",
			Desc:   "第27期每周必看",
		})
		fmt.Println(err)
	}))
}

func TestService_HonorLink(t *testing.T) {
	Convey("TestService_HonorLink", t, WithService(func(s *Service) {
		link := s.selectedHonorURL(77)
		fmt.Println(link)
	}))
}

func TestService_loadPopularCard(t *testing.T) {
	Convey("TestService_dealSelected", t, WithService(func(s *Service) {
		s.loadPopularCard()
	}))
}
