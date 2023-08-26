package service

import (
	"context"
	"encoding/json"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/up-archive/job/internal/model"
	"go-gateway/app/app-svr/up-archive/service/api"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
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
	if arcFlowControlMsg.Router != "up-archive-service" || arcFlowControlMsg.Data == nil {
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
	flowState := arcFlowControlMsg.Data.NewFlowState
	fItem := []*cfcgrpc.ForbiddenItem{
		{Key: "no_space", Value: flowState.NoSpace},
		{Key: "up_no_space", Value: flowState.UpNoSpace},
	}
	aid := arcFlowControlMsg.Data.Oid
	var arc *model.UpArc
	if err := retry(func() (err error) {
		arc, err = s.dao.RawArc(ctx, aid)
		return err
	}); err != nil {
		log.Error("日志告警 archiveFlowControlRailGunDo s.arc aid:%d error:%v", aid, err)
		return railgun.MsgPolicyAttempts
	}
	if arc == nil {
		return railgun.MsgPolicyIgnore
	}
	arc.RandScoreNumber()
	s.updateArchiveFlowControl(ctx, arc, fItem)
	log.Warn("archiveFlowControlRailGunDo success aid:%d flowState:%+v", aid, flowState)
	return railgun.MsgPolicyNormal
}

func (s *Service) updateArchiveFlowControl(ctx context.Context, arc *model.UpArc, fItem []*cfcgrpc.ForbiddenItem) {
	var needAddArc, needDelArc, needDelUpNoSpaceArc *model.UpArc
	func() {
		if !arc.IsAllowed(fItem) { // no_space 全都删除
			needDelArc = arc
			return
		}
		if arc.IsUpNoSpace(fItem) {
			needAddArc = arc          // none 保留
			needDelUpNoSpaceArc = arc // without_up_no_space 删除
			return
		}
		needAddArc = arc
		return
	}()
	if needAddArc == nil && needDelArc == nil && needDelUpNoSpaceArc == nil {
		return
	}
	var needAddStory bool
	if needAddArc != nil {
		func() {
			if err := retry(func() error {
				without := []api.Without{api.Without_none, api.Without_staff}
				if !needAddArc.IsUpNoSpace(fItem) {
					without = append(without, api.Without_no_space)
				}
				return s.appendCacheArcPassed(ctx, needAddArc.Mid, needAddArc, without...)
			}); err != nil {
				log.Error("日志告警 updateArchiveFlowControl needAddArc arc:%+v error:%v", needAddArc, err)
			}
			if needAddArc.IsStory() {
				needAddStory = true
			}
			log.Warn("updateArchiveFlowControl success needAddArc mid:%d arc:%+v", needAddArc.Mid, needAddArc)
		}()
	}
	if needAddStory {
		func() {
			if err := retry(func() error {
				return s.dao.AppendCacheArcStoryPassed(ctx, needAddArc.Mid, []*model.UpArc{needAddArc})
			}); err != nil {
				log.Error("日志告警 updateArchiveList needAddStory arc:%+v error:%v", needAddArc, err)
			}
			log.Warn("updateArchiveFlowControl success needAddStory mid:%d arc:%+v", needAddArc.Mid, needAddArc)
		}()
	}
	// 联合投稿
	if needAddArc != nil {
		func() {
			var staffMids []int64
			if err := retry(func() error {
				var err error
				staffMids, err = s.dao.RawStaffMids(ctx, needAddArc.Aid)
				return err
			}); err != nil {
				log.Error("日志告警 updateArchiveFlowControl needAddArc RawStaffMids aid:%d error:%v", needAddArc.Aid, err)
				return
			}
			if len(staffMids) == 0 {
				return
			}
			for _, staffMid := range staffMids {
				if err := retry(func() error {
					without := []api.Without{api.Without_none, api.Without_no_space}
					return s.appendCacheArcPassed(ctx, staffMid, needAddArc, without...)
				}); err != nil {
					log.Error("日志告警 updateArchiveFlowControl needAddArc staffMid:%d arc:%+v error:%v", staffMid, needAddArc, err)
				}
			}
			log.Warn("updateArchiveFlowControl success needAddArc staffMids:%v arc:%+v", staffMids, needAddArc)
			if needAddStory {
				for _, staffMid := range staffMids {
					if err := retry(func() error {
						return s.dao.AppendCacheArcStoryPassed(ctx, staffMid, []*model.UpArc{needAddArc})
					}); err != nil {
						log.Error("日志告警 updateArchiveFlowControl needAddStory story staffMid:%d arc:%+v error:%v", staffMid, needAddArc, err)
					}
				}
				log.Warn("updateArchiveFlowControl success needAddStory staffMids:%v arc:%+v", staffMids, needAddArc)
			}
		}()
	}
	if needDelUpNoSpaceArc != nil {
		if err := retry(func() error {
			return s.delCacheArcPassed(ctx, needDelUpNoSpaceArc.Mid, &model.UpArc{Aid: needDelUpNoSpaceArc.Aid}, api.Without_no_space)
		}); err != nil {
			log.Error("日志告警 updateArchiveFlowControl needDelUpNoSpaceArc mid:%d aid:%d error:%v", needDelUpNoSpaceArc.Mid, needDelUpNoSpaceArc.Aid, err)
		}
		log.Warn("updateArchiveFlowControl success needDelUpNoSpaceArc mid:%d aid:%d", needDelUpNoSpaceArc.Mid, needDelUpNoSpaceArc.Aid)
	}
	if needDelArc != nil {
		if err := retry(func() error {
			return s.delCacheArcPassed(ctx, needDelArc.Mid, &model.UpArc{Aid: needDelArc.Aid}, api.Without_none, api.Without_staff, api.Without_no_space)
		}); err != nil {
			log.Error("日志告警 updateArchiveFlowControl needDelArc mid:%d aid:%d error:%v", needDelArc.Mid, needDelArc.Aid, err)
		}
		if needDelArc.IsStory() {
			if err := retry(func() error {
				return s.dao.DelCacheArcStoryPassed(ctx, needDelArc.Mid, []*model.UpArc{{Aid: needDelArc.Aid}})
			}); err != nil {
				log.Error("日志告警 updateArchiveFlowControl needDelArcStory mid:%d aid:%d error:%v", needDelArc.Mid, needDelArc.Aid, err)
			}
		}
		log.Warn("updateArchiveFlowControl success needDelArcStory mid:%d aid:%d", needDelArc.Mid, needDelArc.Aid)
	}
	if needDelUpNoSpaceArc != nil || needDelArc != nil {
		var delAid int64
		if needDelUpNoSpaceArc != nil {
			delAid = needDelUpNoSpaceArc.Aid
		}
		if delAid <= 0 {
			delAid = needDelArc.Aid
		}
		func() {
			var staffMids []int64
			if err := retry(func() error {
				var err error
				staffMids, err = s.dao.RawStaffMids(ctx, delAid)
				return err
			}); err != nil {
				log.Error("日志告警 updateArchiveFlowControl RawStaffMids aid:%d error:%v", delAid, err)
				return
			}
			if len(staffMids) == 0 {
				return
			}
			for _, staffMid := range staffMids {
				if needDelUpNoSpaceArc != nil && staffMid == needDelUpNoSpaceArc.Mid { //空间防刷不删除副投稿人数据
					if err := retry(func() error {
						return s.delCacheArcPassed(ctx, staffMid, &model.UpArc{Aid: delAid}, api.Without_no_space)
					}); err != nil {
						log.Error("日志告警 updateArchiveFlowControl needDelUpNoSpaceArc mid:%d aid:%d error:%v", staffMid, delAid, err)
					}
					log.Warn("updateArchiveFlowControl success needDelUpNoSpaceArc mid:%d aid:%d", staffMid, delAid)
				}
				if needDelArc != nil { //审核空间禁止需要删除副投稿人数据
					if err := retry(func() error {
						return s.delCacheArcPassed(ctx, staffMid, &model.UpArc{Aid: delAid}, api.Without_none, api.Without_no_space)
					}); err != nil {
						log.Error("日志告警 updateArchiveFlowControl DelCacheArcPassed staffMid:%d aid:%d error:%v", staffMid, delAid, err)
					}
					log.Warn("updateArchiveFlowControl success DelCacheArcPassed staffMid:%v aid:%d", staffMid, delAid)
				}
			}
		}()
	}
}
