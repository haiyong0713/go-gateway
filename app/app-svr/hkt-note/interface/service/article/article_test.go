package article

import (
	"context"
	"testing"

	artmdl "go-gateway/app/app-svr/hkt-note/interface/model/article"
	"go-gateway/app/app-svr/hkt-note/interface/model/note"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPubListInUser(t *testing.T) {
	Convey("PubListInUser", t, WithService(func(s *Service) {
		res, err := s.PubListInUser(context.Background(), 27515242, 1, 10)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestPubListInArc(t *testing.T) {
	Convey("PubListInArc", t, WithService(func(s *Service) {
		req := &artmdl.PubListInArcReq{
			Oid: 840016893,
			Pn:  1,
			Ps:  10,
		}
		res, err := s.PubListInArc(context.Background(), req, 0)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestPublishNoteInfo(t *testing.T) {
	Convey("PublishNoteInfo", t, WithService(func(s *Service) {
		req := &artmdl.PubNoteInfoReq{
			Cvid:   4421,
			Device: note.Device{},
		}
		res, err := s.PublishNoteInfo(context.Background(), req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
