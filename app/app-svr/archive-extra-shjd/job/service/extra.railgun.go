package service

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-common/library/railgun"

	"go-gateway/app/app-svr/archive-extra-shjd/job/conf"
	"go-gateway/app/app-svr/archive-extra-shjd/job/model"
)

type Railgun struct {
	cfg  *conf.SingleRailgun
	r    *railgun.Railgun
	tp   string
	name string
}

func (s *Service) initArchiveExtraRg() {
	s.ArchiveExtraRgs = []*Railgun{
		{name: "ArchiveExtra", cfg: s.c.ArchiveExtraBizRailgun, tp: model.ArchiveExtraBinlog},
	}
	for _, r := range s.ArchiveExtraRgs {
		r.r = railgun.NewRailGun(r.name, r.cfg.Cfg,
			railgun.NewDatabusV1Inputer(r.cfg.Databus),
			railgun.NewSingleProcessor(r.cfg.Single, s.unpackGen(r.tp), s.railgunDo),
		)
		r.r.Start()
	}
}

func (s *Service) closeArchiveExtraRg() {
	for _, rg := range s.ArchiveExtraRgs {
		rg.r.Close()
	}
}

// statUnpackGen 消费
func (s *Service) unpackGen(tp string) func(msg railgun.Message) (res *railgun.SingleUnpackMsg, err error) {
	return func(msg railgun.Message) (res *railgun.SingleUnpackMsg, err error) {
		log.Info("unpackGen got message(%+v)", msg)
		switch tp {
		case model.ArchiveExtraBinlog:
			var ms = &model.ArchiveExtraBizMsg{}
			if err = json.Unmarshal(msg.Payload(), ms); err != nil {
				log.Error("archiveExtraBizUnpack json.Unmarshal(%s) error(%v)", msg.Payload(), err)
				return
			}
			if ms.Table != model.TableArchiveExtraBiz || ms.New == nil || ms.New.Aid <= 0 {
				log.Warn("archiveExtraBizUnpack unexpected ms(%+v) ", ms)
				return
			}
			return &railgun.SingleUnpackMsg{
				Group: ms.New.Aid,
				Item:  ms,
			}, nil
		default:
			log.Error("unpackGen unknown type(%s) message(%s)", tp, msg.Payload())
			return
		}
	}
}

func (s *Service) railgunDo(c context.Context, item interface{}) railgun.MsgPolicy {
	switch item.(type) {
	case *model.ArchiveExtraBizMsg:
		return s.handleArchiveExtraBizMsg(c, item)
	default:
		log.Error("railgunDo unknown item(%+v)", item)
		return railgun.MsgPolicyIgnore
	}
}

func (s *Service) handleArchiveExtraBizMsg(c context.Context, item interface{}) railgun.MsgPolicy {
	bizMsg := item.(*model.ArchiveExtraBizMsg)
	if bizMsg.New == nil {
		log.Error("handleArchiveExtraBizMsg bizMsg.New is nil, bizMsg(%+v)", bizMsg)
		return railgun.MsgPolicyFailure
	}

	switch bizMsg.Action {
	case model.BinlogInsert:
		if err := s.d.AddArchiveExtraCache(c, bizMsg.New.Aid, bizMsg.New.BizType, bizMsg.New.BizValue); err != nil {
			return railgun.MsgPolicyAttempts
		}
	case model.BinlogUpdate:
		if bizMsg.New.IsDeleted == 1 {
			if err := s.d.DelArchiveExtraCache(c, bizMsg.New.Aid, bizMsg.New.BizType); err != nil {
				return railgun.MsgPolicyAttempts
			}
		} else {
			if err := s.d.AddArchiveExtraCache(c, bizMsg.New.Aid, bizMsg.New.BizType, bizMsg.New.BizValue); err != nil {
				return railgun.MsgPolicyAttempts
			}
		}
	default:
		log.Error("handleArchiveExtraBizMsg unknown action bizMsg(%+v)", bizMsg)
		return railgun.MsgPolicyIgnore
	}

	log.Info("handleArchiveExtraBizMsg handle msg succeeded, bizMsg(%+v)", bizMsg)
	return railgun.MsgPolicyNormal
}
