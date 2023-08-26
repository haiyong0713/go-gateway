package service

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/esports/job/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var svf *Service

func WithService(f func(s *Service)) func() {
	return func() {
		dir, _ := filepath.Abs("../cmd/esports-job-test.toml")
		flag.Set("conf", dir)
		conf.Init()
		log.Init(conf.Conf.Log)
		svf = New(conf.Conf)
		time.Sleep(200 * time.Millisecond)
		f(svf)
	}
}

func TestService_FtpUpload(t *testing.T) {
	Convey("FtpUpload", t, WithService(func(s *Service) {
		err := s.FtpUpload()
		So(err, ShouldBeNil)
	}))
}

func TestService_SeasonValues(t *testing.T) {
	Convey("LoadSeasons", t, WithService(func(s *Service) {
		err := s.LoadSeasons()
		So(err, ShouldBeNil)
	}))
}

func TestService_TeamsValues(t *testing.T) {
	Convey("LoadTeams", t, WithService(func(s *Service) {
		err := s.LoadTeams()
		So(err, ShouldBeNil)
	}))
}

func TestService_ContestsValues(t *testing.T) {
	Convey("LoadContests", t, WithService(func(s *Service) {
		err := s.LoadContests()
		So(err, ShouldBeNil)
	}))
}

func TestService_LoadMatchs(t *testing.T) {
	Convey("LoadMatchs", t, WithService(func(s *Service) {
		err := s.LoadMatchs()
		So(err, ShouldBeNil)
	}))
}
