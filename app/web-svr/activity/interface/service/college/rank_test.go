package college

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProvinceRank(t *testing.T) {
	Convey("test award member count success", t, WithService(func(s *Service) {

		_, err := s.ProvinceRank(context.Background(), 1, 1, 2)
		So(err, ShouldBeNil)

	}))

}
