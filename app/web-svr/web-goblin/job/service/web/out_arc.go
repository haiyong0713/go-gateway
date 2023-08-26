package web

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/web-svr/web-goblin/job/model/web"
	"go-gateway/pkg/idsafe/bvid"
)

const _avExternalPools = "av_external_pools"

func (s *Service) initOutArcRailGun(cfg *railgun.DatabusV1Config, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.outArcUnpack, s.outArcDo)
	g := railgun.NewRailGun("稿件External日志", nil, inputer, processor)
	s.outArcRailgun = g
	g.Start()
}

func (s *Service) outArcUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	var v *web.OutArcMsg
	if err := json.Unmarshal(msg.Payload(), &v); err != nil {
		return nil, err
	}
	if v == nil || v.Table != _avExternalPools || v.Action != _update {
		return nil, nil
	}
	return &railgun.SingleUnpackMsg{
		Group: time.Now().Unix(),
		Item:  v,
	}, nil
}

func (s *Service) outArcDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	res := item.(*web.OutArcMsg)
	// 上线
	if (res.Old == nil || res.Old.Available != 1) && res.New != nil && res.New.Available == 1 {
		if err := s.dao.AddArc(context.Background(), res.New.Avid, res.New.Click, 1); err != nil {
			log.Error("outArcproc AddArc aid:%d error:%+v", res.New.Avid, err)
		}
	}
	// 下线
	if res.Old != nil && res.Old.Available == 1 && res.New != nil && res.New.Available == 2 {
		if err := s.dao.DelArc(context.Background(), res.New.Avid); err != nil {
			log.Error("outArcproc DelArc aid:%d error:%+v", res.New.Avid, err)
		}
		bvidStr, _ := bvid.AvToBv(res.New.Avid)
		if err := s.dao.DelXiaomiArc(context.Background(), 1, bvidStr); err != nil {
			log.Error("outArcproc DelXiaomiArc aid:%d error:%+v", res.New.Avid, err)
			return railgun.MsgPolicyAttempts
		}
	}
	log.Info("outArcDo success,data:%+v", res)
	time.Sleep(10 * time.Millisecond)
	return railgun.MsgPolicyNormal
}
