package fawkes

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	fkmdl "go-gateway/app/app-svr/fawkes/service/model"
	fkcdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
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

func TestService_verse(t *testing.T) {
	Convey("test verse", t, WithService(func(s *Service) {
		Convey("排除的系统版本包含当前版本17，不可以升级", WithService(func(s *Service) {
			var (
				system               = "17"
				build, buildID int64 = 1, 2
				upgrade              = &fkcdmdl.UpgradConfig{
					AppKey:        "android",
					Env:           "prod",
					System:        "29",
					ExcludeSystem: "17,18",
				}
				versions map[int64]*fkmdl.Version
			)
			_, pass := s.verse(system, build, buildID, upgrade, versions)
			So(pass, ShouldBeFalse)
		}))

		Convey("排除的系统版本不包含当前版本17，继续后续流程，可升级版本包含当前版本，但是未指定升级模式，不可以升级", WithService(func(s *Service) {
			var (
				system               = "17"
				build, buildID int64 = 1, 2
				upgrade              = &fkcdmdl.UpgradConfig{
					AppKey:        "android",
					Env:           "prod",
					System:        "17，29",
					ExcludeSystem: "18,19",
				}
				versions map[int64]*fkmdl.Version
			)
			_, pass := s.verse(system, build, buildID, upgrade, versions)
			So(pass, ShouldBeFalse)
		}))

		Convey("排除的系统版本为空，不会走到排除逻辑。可升级的版本不包含当前系统版本，不可以升级", WithService(func(s *Service) {
			var (
				system               = "17"
				build, buildID int64 = 1, 2
				upgrade              = &fkcdmdl.UpgradConfig{
					AppKey:        "android",
					Env:           "prod",
					System:        "29",
					ExcludeSystem: "",
				}
				versions map[int64]*fkmdl.Version
			)
			_, pass := s.verse(system, build, buildID, upgrade, versions)
			So(pass, ShouldBeFalse)
		}))

		Convey("排除的系统版本为空，不会走到排除逻辑。可升级的版本包含当前系统版本，但是未指定升级模式，不可以升级", WithService(func(s *Service) {
			var (
				system               = "17"
				build, buildID int64 = 1, 2
				upgrade              = &fkcdmdl.UpgradConfig{
					AppKey:        "android",
					Env:           "prod",
					System:        "17",
					ExcludeSystem: "",
				}
				versions map[int64]*fkmdl.Version
			)
			_, pass := s.verse(system, build, buildID, upgrade, versions)
			So(pass, ShouldBeFalse)
		}))

		Convey("排除的系统版本和可升级的版本皆为空，跳过版本校验。未指定升级模式，不可以升级", WithService(func(s *Service) {
			var (
				system               = "17"
				build, buildID int64 = 1, 2
				upgrade              = &fkcdmdl.UpgradConfig{
					AppKey:        "android",
					Env:           "prod",
					System:        "",
					ExcludeSystem: "",
				}
				versions map[int64]*fkmdl.Version
			)
			_, pass := s.verse(system, build, buildID, upgrade, versions)
			So(pass, ShouldBeFalse)
		}))

	}))
}
