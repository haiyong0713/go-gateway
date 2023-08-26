package fm

import (
	"context"
	"strconv"
	"testing"
	"time"

	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-car/job/model/fm"

	"github.com/smartystreets/goconvey/convey"
)

func Test_upsertFmSeasonDb(t *testing.T) {
	convey.Convey("Test_upsertFmSeasonDb", t, WithService(func(s *Service) {
		var (
			c      = context.Background()
			fmId   = time.Now().Unix() % 1000
			season = &fm.CommonSeason{
				Scene: fm.SceneFm,
				Fm: fm.FmSeason{
					Scene:  fm.SceneFm,
					FmType: fm.AudioSeason,
					FmId:   fmId,
					Title:  "测试FM合集-id" + strconv.FormatInt(fmId, 10),
					Cover:  "http://i0.hdslb.com/bfs/tvcover/472eb15695dd42a197e4c8f8eae6875fc3e7e5fa.jpg",
					FmList: []*fm.Item{{440117562}, {760040061}, {240029469}, {840071568}, {720064499}},
				},
			}
		)
		policy, err := s.upsertFmSeasonDb(c, season)
		convey.So(err, convey.ShouldBeNil)
		convey.So(policy, convey.ShouldEqual, railgun.MsgPolicyNormal)
	}))
}

func Test_upsertVideoSeason(t *testing.T) {
	convey.Convey("Test_upsertVideoSeason", t, WithService(func(s *Service) {
		var (
			c      = context.Background()
			fmId   = time.Now().Unix() % 1000
			season = &fm.CommonSeason{
				Scene: fm.SceneVideo,
				Video: fm.VideoSeason{
					Scene:      fm.SceneVideo,
					SeasonId:   fmId,
					Title:      "测试视频合集-id" + strconv.FormatInt(fmId, 10),
					Cover:      "http://i0.hdslb.com/bfs/tvcover/472eb15695dd42a197e4c8f8eae6875fc3e7e5fa.jpg",
					SeasonList: []*fm.Item{{960044310}, {360080117}, {880107602}, {680036510}},
				},
			}
		)
		policy, err := s.upsertVideoSeasonDb(c, season)
		convey.So(err, convey.ShouldBeNil)
		convey.So(policy, convey.ShouldEqual, railgun.MsgPolicyNormal)
	}))
}
