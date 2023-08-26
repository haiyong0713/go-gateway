package web

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/web-svr/web-goblin/job/model/web"
)

func (s *Service) initArcNotifyRailGun(cfg *railgun.DatabusV1Config, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.arcNotifyUnpack, s.arcNotifyDo)
	g := railgun.NewRailGun("稿件通知日志", nil, inputer, processor)
	s.arcNotifyRailgun = g
	g.Start()
}

func (s *Service) arcNotifyUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	var v *web.ArcMsg
	if err := json.Unmarshal(msg.Payload(), &v); err != nil {
		return nil, err
	}
	if v == nil || v.Table != _archive {
		return nil, nil
	}
	return &railgun.SingleUnpackMsg{
		Group: time.Now().Unix(),
		Item:  v,
	}, nil
}

func (s *Service) arcNotifyDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	res := item.(*web.ArcMsg)
	if err := s.UgcIncrement(ctx, res); err != nil {
		return railgun.MsgPolicyAttempts
	}
	log.Info("arcNotifyDo success,data:%+v", res)
	time.Sleep(10 * time.Millisecond)
	return railgun.MsgPolicyNormal
}
