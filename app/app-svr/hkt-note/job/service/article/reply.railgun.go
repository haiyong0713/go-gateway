package article

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/hkt-note/job/model/article"
	"time"
)

func (s *Service) initReplyDelRailGun(cfg *railgun.DatabusV1Config) {
	log.Warn("StartJob: initReplyDelRailGun OK-------")
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(nil, s.replyDelRailGunUnpack, s.replyDelRailGunDo)
	g := railgun.NewRailGun("ReplyDel", nil, inputer, processor)
	s.replyDelRailGun = g
	g.Start()
}

func (s *Service) replyDelRailGunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	log.Warn("initReplyDelRailGun msg=%s", msg.Payload())
	m := &article.ReplyDelMsg{}
	if err := json.Unmarshal(msg.Payload(), &m); err != nil {
		log.Error("replyError initReplyDelRailGun msg(%s) err(%+v)", msg.Payload(), err)
		return nil, err
	}
	if m == nil {
		log.Error("replyError initReplyDelRailGun msg(%s) empty", msg.Payload())
		return nil, nil
	}
	return &railgun.SingleUnpackMsg{
		Group: time.Now().Unix(),
		Item:  m,
	}, nil
}

func (s *Service) replyDelRailGunDo(_ context.Context, item interface{}) railgun.MsgPolicy {
	rs := item.(*article.ReplyDelMsg)
	if rs == nil {
		log.Error("replyError initReplyDelRailGun item(%+v)", item)
		return railgun.MsgPolicyIgnore
	}
	if err := s.treatReplyDelMsg(context.TODO(), rs); err != nil {
		return railgun.MsgPolicyAttempts
	}
	return railgun.MsgPolicyNormal
}
