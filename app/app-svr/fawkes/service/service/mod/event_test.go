package mod

import (
	"context"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/fawkes/service/model/mod"
)

func TestService_versionConfigChangeAction(t *testing.T) {

	monkey.PatchInstanceMethod(reflect.TypeOf(srv), "ModReleaseTrafficEstimate", func(service *Service, ctx context.Context, versionID int64, user string) (traffic *mod.Traffic, err error) {
		return &mod.Traffic{
			Pool:                    nil,
			Module:                  nil,
			Version:                 nil,
			File:                    nil,
			Patches:                 nil,
			Config:                  nil,
			Gray:                    nil,
			SetUpUserCount:          3245234,
			DownloadSizeOnlineBytes: 542643545,
			Operator:                "user",
		}, err
	})

	convey.Convey("old is nil", t, func() {
		srv.versionConfigChangeAction(&VersionConfigChangeArgs{
			Ctx: context.Background(),
			old: nil,
			new: &mod.Config{
				ID:        1,
				VersionID: 130898,
				Priority:  "low",
				AppVer:    "",
				SysVer:    "",
				Stime:     0,
				Etime:     0,
			},
			UserName: "me",
		})
	})

	convey.Convey("old not nil", t, WithService(func(s *Service) {
		s.versionConfigChangeAction(&VersionConfigChangeArgs{
			Ctx: context.Background(),
			old: &mod.Config{
				ID:        1,
				VersionID: 130898,
				Priority:  "low",
				AppVer:    "",
				SysVer:    "",
				Stime:     0,
				Etime:     0,
			},
			new: &mod.Config{
				ID:        1,
				VersionID: 130898,
				Priority:  "high",
				AppVer:    "",
				SysVer:    "",
				Stime:     0,
				Etime:     0,
			},
			UserName: "me",
		})
	}))
}

func Test_calcCost(t *testing.T) {
	b := calcCost(4.2498, mod.PriorityMiddle)
	print(b)
}

func TestService_versionPushAction(t *testing.T) {
	convey.Convey("TestService_versionPushAction", t, func() {
		srv.versionPushAction(&VersionPushArgs{
			Ctx:       context.Background(),
			VersionId: 131603,
			UserName:  "luweidan",
		})
	})
}
