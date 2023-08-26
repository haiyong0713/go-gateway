package rank

import (
	"context"
	go_common_library_time "go-common/library/time"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v3"
	"testing"
	"time"

	// . "github.com/golang/mock/gomock"
	. "github.com/glycerine/goconvey/convey"
)

func TestGetNeedRankLog(t *testing.T) {
	Convey("test getNeedRankLog  success", t, WithService(func(s *Service) {
		ctx := context.Background()
		rule := make([]*rankmdl.Rule, 0)
		otherRule := make([]*rankmdl.Rule, 0)
		now := time.Now().Unix()
		rule = append(rule, &rankmdl.Rule{
			BaseID:          1,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeDay,
			UpdateScope:     rankmdl.UpdateScopeIncrement,
			Stime:           go_common_library_time.Time(now),
		}, &rankmdl.Rule{
			BaseID:          2,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeDay,
			UpdateScope:     rankmdl.UpdateScopeIncrement,
			Stime:           go_common_library_time.Time(time.Now().AddDate(0, 0, -1).Unix()),
		}, &rankmdl.Rule{
			BaseID:          3,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeDay,
			UpdateScope:     rankmdl.UpdateScopeTotal,
			Stime:           go_common_library_time.Time(now),
		}, &rankmdl.Rule{
			BaseID:          4,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeDay,
			UpdateScope:     rankmdl.UpdateScopeTotal,
			Stime:           go_common_library_time.Time(time.Now().AddDate(0, 0, -1).Unix()),
		}, &rankmdl.Rule{
			BaseID:          5,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeWeek,
			UpdateScope:     rankmdl.UpdateScopeIncrement,
			Stime:           go_common_library_time.Time(time.Now().AddDate(0, 0, -7).Unix()),
		}, &rankmdl.Rule{
			BaseID:          6,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeWeek,
			UpdateScope:     rankmdl.UpdateScopeTotal,
			Stime:           go_common_library_time.Time(time.Now().AddDate(0, 0, -14).Unix()),
		}, &rankmdl.Rule{
			BaseID:          7,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeMonth,
			UpdateScope:     rankmdl.UpdateScopeIncrement,
			Stime:           go_common_library_time.Time(time.Now().AddDate(0, -1, 0).Unix()),
		}, &rankmdl.Rule{
			BaseID:          7,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeMonth,
			UpdateScope:     rankmdl.UpdateScopeTotal,
			Stime:           go_common_library_time.Time(time.Now().AddDate(0, -2, 0).Unix()),
		})
		newLog := s.getNeedRankLog(ctx, rule)
		// fmt.Println(newLog)
		// fmt.Println(len(newLog))
		// for _, v := range newLog {
		// 	fmt.Println(v)
		// }
		So(len(newLog), ShouldEqual, len(rule))
		otherRule = append(otherRule, &rankmdl.Rule{
			BaseID:          5,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeWeek,
			UpdateScope:     rankmdl.UpdateScopeIncrement,
			Stime:           go_common_library_time.Time(time.Now().AddDate(0, 0, -1).Unix()),
		}, &rankmdl.Rule{
			BaseID:          6,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeWeek,
			UpdateScope:     rankmdl.UpdateScopeTotal,
			Stime:           go_common_library_time.Time(time.Now().AddDate(0, 0, -8).Unix()),
		}, &rankmdl.Rule{
			BaseID:          7,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeMonth,
			UpdateScope:     rankmdl.UpdateScopeIncrement,
			Stime:           go_common_library_time.Time(time.Now().AddDate(0, 0, -28).Unix()),
		}, &rankmdl.Rule{
			BaseID:          7,
			LastBatch:       1,
			UpdateFrequency: rankmdl.FrequencyTypeMonth,
			UpdateScope:     rankmdl.UpdateScopeTotal,
			Stime:           go_common_library_time.Time(time.Now().AddDate(0, 0, -47).Unix()),
		})
		newOtherLog := s.getNeedRankLog(ctx, otherRule)

		So(len(newOtherLog), ShouldEqual, 0)

	}))
}
