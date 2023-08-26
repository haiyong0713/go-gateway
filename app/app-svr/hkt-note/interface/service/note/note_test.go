package note

import (
	"context"
	"go-gateway/app/app-svr/hkt-note/interface/model/note"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNoteAdd(t *testing.T) {
	Convey("NoteAdd", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &note.NoteAddReq{
				Bvid:    "",
				Aid:     1,
				NoteId:  0,
				Title:   "newnew",
				Summary: "nnnnnew",
				Tags:    "edfff",
				Content: "testt",
			}
		)
		res, err := s.NoteAdd(c, req, 15555180)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestNoteInfo(t *testing.T) {
	Convey("NoteInfo", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &note.NoteInfoReq{
				NoteId: 178254324432906,
				Oid:    123,
			}
		)
		res, err := s.NoteInfo(c, req, 1)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestNoteList(t *testing.T) {
	Convey("NoteList", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			mid = int64(15555180)
		)
		res, err := s.NoteList(c, mid, 1, 5)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestToArcCore(t *testing.T) {
	Convey("toArcCore", t, WithService(func(s *Service) {
		res := s.dao.ToArcCore(context.Background(), 81, 0)
		So(res, ShouldNotBeNil)
	}))
}

func TestSendNoteNotify(t *testing.T) {
	cont := &note.NtNotifyMsg{
		ReplyMsg: &note.ReplyMsg{
			NoteId:  19619081043836934,
			Content: "测试测试{note:19619081043836934}前面是笔记id",
			Oid:     760086091,
			Mid:     1111121516,
		},
	}
	Convey("toArcCore", t, WithService(func(s *Service) {
		err := s.sendNoteNotify(context.Background(), cont, 1111121516)
		So(err, ShouldBeNil)
	}))
}
