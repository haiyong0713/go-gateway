package article

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/service/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPublishListInArc(t *testing.T) {
	Convey("PublishListInArc", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &api.NoteListReq{
				Oid:     840016893,
				OidType: 0,
				Pn:      1,
				Ps:      10,
			}
		)
		res, err := s.PublishListInArc(c, req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestPublishListInUser(t *testing.T) {
	Convey("PublishListInUser", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &api.NoteListReq{
				Mid: 27515232,
				Pn:  1,
				Ps:  10,
			}
		)
		res, err := s.PublishListInUser(c, req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestPublishNoteInfo(t *testing.T) {
	Convey("PublishNoteInfo", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &api.PublishNoteInfoReq{
				Cvid: 4420,
			}
		)
		res, err := s.PublishNoteInfo(c, req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestSimpleArticles(t *testing.T) {
	Convey("SimpleArticles", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &api.SimpleArticlesReq{
				Cvids: []int64{4494, 4466},
			}
		)
		res, err := s.SimpleArticles(c, req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
