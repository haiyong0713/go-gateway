package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_ElecShow(t *testing.T) {
	Convey("test elec ElecShow", t, WithService(func(s *Service) {
		var (
			mid int64 = 27515256
			aid int64 = 5464686
		)
		res, err := s.ElecShow(context.Background(), mid, aid, nil)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
