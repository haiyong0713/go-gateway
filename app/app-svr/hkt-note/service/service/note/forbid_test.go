package note

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/service/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoadFeaPolitics(t *testing.T) {
	Convey("loadFeaPolitics", t, WithService(func(s *Service) {
		s.loadFeaPolitics()
	}))
}

func TestArcsForbid(t *testing.T) {
	Convey("ArcsForbid", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &api.ArcsForbidReq{
				Aids: []int64{880014573, 10318707, 10318733},
			}
		)
		res, err := s.ArcsForbid(c, req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
