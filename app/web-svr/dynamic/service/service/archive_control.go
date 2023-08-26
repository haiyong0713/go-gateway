package service

import (
	"context"
	"encoding/json"
	"time"

	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/dynamic/service/model"

	"go-common/library/log"
	"go-common/library/railgun"
)

func (s *Service) initArchiveFlowControlRailGun(cfg *railgun.DatabusV1Config, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.archiveFlowControlRailGunUnpack, s.archiveFlowControlRailGunDo)
	g := railgun.NewRailGun("稿件禁止项变更", nil, inputer, processor)
	s.arcFlowControlRailGun = g
	g.Start()
}

func (s *Service) archiveFlowControlRailGunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	arcFlowControlMsg := new(model.ArchiveFlowControlMsg)
	if err := json.Unmarshal(msg.Payload(), &arcFlowControlMsg); err != nil {
		return nil, err
	}
	if arcFlowControlMsg.Router != "web-interface" || arcFlowControlMsg.Data == nil {
		return nil, nil
	}
	log.Warn("接收稿件禁止项变更消息成功,data:%s", msg.Payload())
	return &railgun.SingleUnpackMsg{
		Group: arcFlowControlMsg.Data.Oid,
		Item:  arcFlowControlMsg,
	}, nil
}

func (s *Service) archiveFlowControlRailGunDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	arcFlowControlMsg := item.(*model.ArchiveFlowControlMsg)
	aid := arcFlowControlMsg.Data.Oid
	var arcsReply *arcmdl.ArcReply
	if err := retry(func() (err error) {
		arcsReply, err = s.arcClient.Arc(ctx, &arcmdl.ArcRequest{Aid: aid})
		return err
	}); err != nil {
		log.Error("archiveFlowControlRailGunDo s.arc aid:%d error:%v", aid, err)
		return railgun.MsgPolicyAttempts
	}
	if arcsReply.GetArc() == nil {
		return railgun.MsgPolicyIgnore
	}
	arc := &model.ArchiveSub{
		Aid:       arcsReply.GetArc().Aid,
		PubTime:   time.Unix(int64(arcsReply.GetArc().PubDate), 0).Format("2006-01-02 15:04:05"),
		State:     int(arcsReply.GetArc().State),
		Typeid:    arcsReply.GetArc().TypeID,
		Copyright: int8(arcsReply.GetArc().Copyright),
		Attribute: arcsReply.GetArc().Attribute,
	}
	s.regionCache(ctx, arc)
	bs, _ := json.Marshal(arcFlowControlMsg)
	log.Warn("archiveFlowControlRailGunDo success aid:%d msg:%v", aid, string(bs))
	return railgun.MsgPolicyNormal
}

func retry(callback func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		if err = callback(); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
}
