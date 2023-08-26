package mod

import (
	"context"
	"reflect"
	"testing"
	"time"

	xtime "go-common/library/time"

	"github.com/bouk/monkey"
	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	"go-gateway/app/app-svr/fawkes/service/model/mod"
)

func Test_timeRange(t *testing.T) {
	Convey("", t, func() {
		config := &mod.Config{
			ID:        0,
			VersionID: 0,
			Priority:  "high",
			AppVer:    "",
			SysVer:    "",
			Stime:     xtime.Time(time.Now().Unix()),
			Etime:     xtime.Time(time.Now().Unix()),
		}

		start, end, err := timeRange(context.Background(), config, 2*time.Minute, 5*time.Minute)
		So(err, ShouldBeNil)
		So(start, ShouldNotBeEmpty)
		So(end, ShouldNotBeEmpty)
	})

}

func TestService_ModDownloadCountEstimate(t *testing.T) {
	config := &mod.Config{
		ID:        0,
		VersionID: 0,
		Priority:  "high",
		AppVer:    "",
		SysVer:    "",
		Stime:     xtime.Time(time.Now().Unix()),
		Etime:     xtime.Time(time.Now().Unix()),
	}

	gray := &mod.Gray{
		ID:             0,
		VersionID:      0,
		Strategy:       0,
		Salt:           "",
		BucketStart:    0,
		BucketEnd:      0,
		Whitelist:      "",
		WhitelistURL:   "",
		ManualDownload: false,
	}

	Convey("TestService_ModDownloadCountEstimate", t, WithService(
		func(s *Service) {
			count, err := s.ModDownloadCountEstimate(context.Background(), "w19e", config, gray, time.Now().Add(-5*time.Minute), time.Now())
			So(err, ShouldBeNil)
			So(count, ShouldNotBeEmpty)

		},
	))
}

func TestService_ModReleaseTrafficEstimate(t *testing.T) {
	Convey("TestService_ModReleaseTrafficEstimate", t, WithService(
		func(s *Service) {
			monkey.PatchInstanceMethod(reflect.TypeOf(s), "ModDownloadCountEstimate", func(service *Service, ctx context.Context, appKey string, config *mod.Config, gray *mod.Gray, start, end time.Time) (activeUsersCount float64, err error) {
				return 1111, nil
			})

			// ModDownloadSizeSum(c context.Context, appKey, poolName, modName string, startTime, endTime time.Time) (downloadSize float64, err error) {
			monkey.PatchInstanceMethod(reflect.TypeOf(s.fkDao), "ModDownloadSizeSum", func(fkdao *fawkes.Dao, c context.Context, appKey, poolName, modName string, startTime, endTime time.Time) (downloadSize float64, err error) {
				return 23333, nil
			})

			count, err := s.ModReleaseTrafficEstimate(context.Background(), 121615, "wo")
			So(err, ShouldBeNil)
			So(count, ShouldNotBeEmpty)
		},
	))
}

func TestService_Alert(t *testing.T) {
	Convey("TestService_ModReleaseTrafficEstimate", t, WithService(
		func(s *Service) {
			monkey.PatchInstanceMethod(reflect.TypeOf(s), "ModReleaseTrafficEstimate", func(service *Service, ctx context.Context, versionID int64, user string) (traffic *mod.Traffic, err error) {
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

			estimate, _ := s.ModReleaseTrafficEstimate(context.Background(), 111, "user")

			if err := s.Alert(context.Background(), estimate, mod.Release); err != nil {
				return
			}
		},
	))
}
