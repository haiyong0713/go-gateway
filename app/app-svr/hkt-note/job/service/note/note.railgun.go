package note

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/hkt-note/job/model/note"
)

func (s *Service) initNoteAuditRailGun(cfg *railgun.DatabusV1Config) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(nil, s.noteAuditRailGunUnpack, s.noteAuditRailGunDo)
	g := railgun.NewRailGun("NoteAuditNotify", nil, inputer, processor)
	s.noteAuditRailGun = g
	g.Start()
}

func (s *Service) noteAuditRailGunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	log.Warn("initNoteAuditRailGun msg=%s", msg.Payload())
	m := &note.NtAuditNotify{}
	if err := json.Unmarshal(msg.Payload(), &m); err != nil {
		log.Error("noteError initNoteAuditRailGun msg(%s) err(%+v)", msg.Payload(), err)
		return nil, err
	}
	rs := m.Content
	if rs == nil {
		log.Error("noteError initNoteAuditRailGun msg(%s) empty", msg.Payload())
		return nil, nil
	}
	return &railgun.SingleUnpackMsg{
		Group: time.Now().Unix(),
		Item:  rs,
	}, nil
}

func (s *Service) noteAuditRailGunDo(_ context.Context, item interface{}) railgun.MsgPolicy {
	rs := item.(*note.NtAddMsg)
	if rs == nil {
		log.Error("noteError initNoteAddRailGun item(%+v)", item)
		return railgun.MsgPolicyIgnore
	}
	s.treatNoteAuditMsg(context.TODO(), rs)
	return railgun.MsgPolicyNormal
}

func (s *Service) initNoteAddRailGun(cfg *railgun.DatabusV1Config) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(nil, s.noteAddRailGunUnpack, s.noteAddRailGunDo)
	g := railgun.NewRailGun("NoteAddNotify", nil, inputer, processor)
	s.noteAddRailGun = g
	g.Start()
}

func (s *Service) noteAddRailGunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	log.Warn("initNoteAddRailGun msg=%s", msg.Payload())
	m := &note.NtNotify{}
	if err := json.Unmarshal(msg.Payload(), &m); err != nil {
		log.Error("noteError initNoteAddRailGun msg(%s) err(%+v)", msg.Payload(), err)
		return nil, err
	}
	rs := m.Content
	if rs == nil {
		log.Error("noteError initNoteAddRailGun msg(%s) empty", msg.Payload())
		return nil, nil
	}
	return &railgun.SingleUnpackMsg{
		Group: time.Now().Unix(),
		Item:  rs,
	}, nil
}

func (s *Service) noteAddRailGunDo(_ context.Context, item interface{}) railgun.MsgPolicy {
	rs := item.(*note.NtNotifyMsg)
	if rs == nil {
		log.Error("noteError initNoteAddRailGun item(%+v)", item)
		return railgun.MsgPolicyIgnore
	}
	c := context.TODO()
	if rs.NtAddMsg != nil {
		if rs.NtAddMsg.Aid > 0 { // TODO 废弃aid
			rs.NtAddMsg.Oid = rs.NtAddMsg.Aid
		}
		s.treatNoteAddNotifyMsg(c, rs.NtAddMsg)
	}
	if rs.NtDelMsg != nil {
		s.treatNoteDelNotifyMsg(c, rs.NtDelMsg)
	}
	if rs.NtPubMsg != nil {
		if err := s.treatNotePubNotifyMsg(c, rs.NtPubMsg); err != nil {
			log.Error("artError treatNotePubNotifyMsg msg(%+v) error(%v)", item, err)
			return railgun.MsgPolicyAttempts
		}
	}
	// 实际上这里都没有执行成功
	// 评论的replyMsg是：{"reply_msg":{"note_id":32344851383585792,"mid":278401889,"oid":812105989,"content":"{note:32344851383585792}我发布了一篇笔记，快来看看吧"}}
	// 正确的应该是: {"topic":"NoteNotify-T","content":{"nt_del_msg":{"note_ids":[32344961734151168],"mid":106381280},"reply_msg":null}}
	if rs.ReplyMsg != nil {
		if err := s.treatReplyMsg(c, rs.ReplyMsg); err != nil {
			log.Error("artError treatReplyMsg msg(%+v) error(%v)", item, err)
			return railgun.MsgPolicyAttempts
		}
	}
	return railgun.MsgPolicyNormal
}

func (s *Service) NoteAddRailgunHttp() func(*bm.Context) {
	return s.noteAddRailGun.BMHandler
}

func (s *Service) NoteAuditRailgunHttp() func(*bm.Context) {
	return s.noteAuditRailGun.BMHandler
}
