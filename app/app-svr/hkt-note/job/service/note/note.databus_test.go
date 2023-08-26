package note

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/job/model/note"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConsumeNoteNotifyMsg(t *testing.T) {
	Convey("consumeNoteNotifyMsg", t, WithService(func(s *Service) {
		s.consumeNoteNotifyMsg()
	}))
}

func TestTreatNoteAuditMsg(t *testing.T) {
	Convey("consumeNoteNotifyMsg", t, WithService(func(s *Service) {
		s.treatNoteAuditMsg(context.TODO(), &note.NtAddMsg{
			NoteId: 124343200,
			Mid:    0,
		})
	}))
}

func TestTreatNotePubNotifyMsg(t *testing.T) {
	Convey("treatNotePubNotifyMsg", t, WithService(func(s *Service) {
		s.treatNotePubNotifyMsg(context.TODO(), &note.NtPubMsg{
			Mid:      27515242,
			NoteId:   9970684826484742,
			ContLen:  22,
			Title:    "埋点自动化勿操作1m30s",
			Summary:  "测试一下发布",
			Oid:      600100329,
			OidType:  0,
			ArcCover: "http://i1.hdslb.com/bfs/archive/6de59d2f7dc7273663363c1aa9acc2557cbf6597.jpg",
			Cvid:     0,
			Mtime:    0,
		})
	}))
}
