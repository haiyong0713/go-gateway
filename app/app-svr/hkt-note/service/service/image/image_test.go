package image

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/service/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPublishImgs(t *testing.T) {
	Convey("PublishImgs", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &api.PublishImgsReq{
				Mid:      216761,
				ImageIds: []int64{30, 31},
			}
		)
		res, err := s.PublishImgs(c, req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
