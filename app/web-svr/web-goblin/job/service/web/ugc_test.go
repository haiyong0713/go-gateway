package web

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/web-goblin/job/model/web"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_UgcIncrement(t *testing.T) {
	Convey("test UgcIncrement", t, WithService(func(s *Service) {
		var (
			err error
			ctx = context.Background()
			v   = &web.ArcMsg{
				Action: "add",
				Table:  "archive",
				New:    nil,
			}
		)
		err = s.UgcIncrement(ctx, v)
		So(err, ShouldBeNil)
	}))
}
