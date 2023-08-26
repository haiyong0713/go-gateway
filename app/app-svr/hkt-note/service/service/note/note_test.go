package note

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/service/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNoteSize(t *testing.T) {
	Convey("NoteSize", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &api.NoteSizeReq{
				Mid:    89,
				NoteId: 0,
			}
		)
		res, err := s.NoteSize(c, req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestNoteInfo(t *testing.T) {
	Convey("NoteInfo", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &api.NoteInfoReq{
				Mid:    1,
				NoteId: 1,
			}
		)
		res, err := s.NoteInfo(c, req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestNoteList(t *testing.T) {
	Convey("NoteList", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &api.NoteListReq{
				Mid: 27515242,
				Pn:  1,
				Ps:  10,
			}
		)
		res, err := s.NoteList(c, req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestNoteCount(t *testing.T) {
	Convey("NoteCount", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &api.NoteCountReq{
				Mid: 1,
			}
		)
		res, err := s.NoteCount(c, req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestSimpleNotes(t *testing.T) {
	Convey("SimpleNotes", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &api.SimpleNotesReq{
				NoteIds: []int64{11423738877181958, 11413048156749834},
				Mid:     27515232,
			}
		)
		res, err := s.SimpleNotes(c, req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
