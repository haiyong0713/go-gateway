package common

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestService_HitSixLimit(t *testing.T) {
	convey.Convey("TestService_HitSixLimit", t, WithService(func(s *Service) {
		var (
			ctx    = context.Background()
			aidMap = map[int64]struct {
				name   string
				expect bool
			}{
				320002280: {"自见稿件", true},
				440110738: {"六禁稿件", true},
				840085592: {"普通稿件", false},
			}
		)
		for k, v := range aidMap {
			t.Run(v.name, func(t *testing.T) {
				if s.HitSixLimit(ctx, k) != v.expect {
					t.Errorf("unexpect hit res, aid:%d", k)
				}
			})
		}
	}))
}

func TestService_SixLimitFilter(t *testing.T) {
	convey.Convey("TestService_SixLimitFilter", t, WithService(func(s *Service) {
		var (
			ctx  = context.Background()
			aids = []int64{320002280, 440110738, 840085592}
		)
		t.Run("TestService_SixLimitFilter Run", func(t *testing.T) {
			convey.ShouldNotContain(s.SixLimitFilter(ctx, aids), 320002280, 440110738)
		})
	}))
}
