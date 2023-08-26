package service

import (
	"context"
	"encoding/json"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/up-archive/job/internal/model"
	"go-gateway/app/app-svr/up-archive/service/api"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

const _archiveTable = "archive"

func (s *Service) initArchiveRailGun(cfg *railgun.DatabusV1Config, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.archiveRailGunUnpack, s.archiveRailGunDo)
	g := railgun.NewRailGun("稿件状态变更", nil, inputer, processor)
	s.archiveRailGun = g
	g.Start()
}

func (s *Service) archiveRailGunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	log.Warn("接收稿件状态变更消息成功,data:%s", msg.Payload())
	canalMsg := new(model.CanalMsg)
	if err := json.Unmarshal(msg.Payload(), &canalMsg); err != nil {
		return nil, err
	}
	switch canalMsg.Table {
	case _archiveTable:
		arcCanalMsg := new(model.ArchiveCanalMsg)
		if err := json.Unmarshal(msg.Payload(), &arcCanalMsg); err != nil {
			return nil, err
		}
		return &railgun.SingleUnpackMsg{
			Group: arcCanalMsg.New.Aid,
			Item:  arcCanalMsg,
		}, nil
	default:
		return nil, nil
	}
}

// nolint:gocognit
func (s *Service) archiveRailGunDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	arcCanalMsg := item.(*model.ArchiveCanalMsg)
	newArc := arcCanalMsg.New
	if newArc == nil {
		return railgun.MsgPolicyIgnore
	}
	fItem, err := s.dao.ContentFlowControlInfo(ctx, newArc.Aid)
	if err != nil {
		log.Error("日志告警 archiveRailGunDo ContentFlowControlInfo aid:%d error:%+v", newArc.Aid, err)
		return railgun.MsgPolicyAttempts
	}
	arcFn := func(arcSub *model.ArchiveSub) (*model.UpArc, error) {
		if arcSub == nil {
			return nil, nil
		}
		arc := new(model.UpArc)
		if err := arc.FromArchiveSub(arcSub); err != nil {
			return nil, err
		}
		return arc, nil
	}
	newArchive, err := arcFn(newArc)
	if err != nil {
		log.Error("日志告警 archiveRailGunDo FromArchiveSub newArc:%+v error:%+v", newArc, err)
		return railgun.MsgPolicyIgnore
	}
	oldArc := arcCanalMsg.Old
	oldArchive, err := arcFn(oldArc)
	if err != nil {
		log.Error("日志告警 archiveRailGunDo FromArchiveSub oldArc:%+v error:%+v", oldArc, err)
		return railgun.MsgPolicyIgnore
	}
	var needAddArc, needDelArc, needDelStory *model.UpArc
	switch arcCanalMsg.Action {
	case model.ActionInsert:
		func() {
			if newArc.IsNormal() && newArc.IsAllowed(fItem) {
				needAddArc = newArchive
				return
			}
			needDelArc = newArchive
		}()
	case model.ActionUpdate:
		if oldArc == nil {
			return railgun.MsgPolicyIgnore
		}
		func() {
			// 稿件状态转为正常，进行新增
			if !oldArc.IsNormal() && newArc.IsNormal() && newArc.IsAllowed(fItem) {
				needAddArc = newArchive
				return
			}
			// 稿件状态转为异常，进行删除
			if oldArc.IsNormal() && (!newArc.IsNormal() || !newArc.IsAllowed(fItem)) {
				needDelArc = oldArchive
				return
			}
			// 稿件状态前后都为异常，就啥也不干
			if !oldArc.IsNormal() && (!newArc.IsNormal() || !newArc.IsAllowed(fItem)) {
				log.Warn("前后一致的异常稿件,mid:%v,aid:%v,old:%+v,new:%+v", newArc.Mid, newArc.Aid, oldArc, newArc)
				return
			}
			// 稿件非状态变更，但是mid变了
			if oldArc.Mid != 0 && oldArc.Mid != newArc.Mid { // mid change
				// needDelArc 包含 needDelStory
				needDelArc = oldArchive
				return
			}
			// 稿件非状态变更，检查story状态变更
			if oldArc.IsStory() && !newArc.IsStory() {
				needDelStory = oldArchive
				return
			}
			// 异常case，archive-job的异常重试，会导致insert变为update，正常稿件，考虑无脑needAdd
			// needAdd 包含 needAddStory
			log.Warn("前后一致的正常稿件,mid:%v,aid:%v,old:%+v,new:%+v", newArc.Mid, newArc.Aid, oldArc, newArc)
			needAddArc = newArchive
		}()
	}
	if needAddArc == nil && needDelArc == nil && needDelStory == nil {
		return railgun.MsgPolicyNormal
	}
	s.updateArchiveList(ctx, needAddArc, needDelArc, needDelStory, fItem)
	return railgun.MsgPolicyNormal
}

func (s *Service) updateArchiveList(ctx context.Context, needAddArc, needDelArc, needDelStory *model.UpArc, fItem []*cfcgrpc.ForbiddenItem) {
	var needAddStory bool
	if needAddArc != nil {
		if needAddArc.IsStory() {
			needAddStory = true
		}
		func() {
			if err := retry(func() error {
				without := []api.Without{api.Without_none, api.Without_staff}
				if !needAddArc.IsUpNoSpace(fItem) {
					without = append(without, api.Without_no_space)
				}
				return s.appendCacheArcPassed(ctx, needAddArc.Mid, needAddArc, without...)
			}); err != nil {
				log.Error("日志告警 updateArchiveList needAddArc arc:%+v error:%v", needAddArc, err)
				return
			}
			log.Warn("updateArchiveList success needAddArc mid:%d arc:%+v", needAddArc.Mid, needAddArc)
		}()
	}
	if needAddStory {
		func() {
			if err := retry(func() error {
				return s.dao.AppendCacheArcStoryPassed(ctx, needAddArc.Mid, []*model.UpArc{needAddArc})
			}); err != nil {
				log.Error("日志告警 updateArchiveList needAddStory arc:%+v error:%v", needAddArc, err)
				return
			}
			log.Warn("updateArchiveList success needAddStory mid:%d arc:%+v", needAddArc.Mid, needAddArc)
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
				log.Error("日志告警 updateArchiveList needAddArc RawStaffMids aid:%d error:%v", needAddArc.Aid, err)
				return
			}
			if len(staffMids) == 0 {
				return
			}
			//联合投稿副投稿人忽略空间防刷:up_no_space
			without := []api.Without{api.Without_none, api.Without_no_space}
			for _, staffMid := range staffMids {
				if err := retry(func() error {
					return s.appendCacheArcPassed(ctx, staffMid, needAddArc, without...)
				}); err != nil {
					log.Error("日志告警 updateArchiveList needAddArc staffMid:%d arc:%+v error:%v", staffMid, needAddArc, err)
				}
			}
			log.Warn("updateArchiveList success needAddArc staffMids:%v arc:%+v", staffMids, needAddArc)
			if needAddStory {
				for _, staffMid := range staffMids {
					if err := retry(func() error {
						return s.dao.AppendCacheArcStoryPassed(ctx, staffMid, []*model.UpArc{needAddArc})
					}); err != nil {
						log.Error("日志告警 updateArchiveList needAddStory staffMid:%d arc:%+v error:%v", staffMid, needAddArc, err)
					}
				}
				log.Warn("updateArchiveList success needAddStory staffMids:%v arc:%+v", staffMids, needAddArc)
			}
		}()
	}
	if needDelArc != nil {
		func() {
			if err := retry(func() error {
				return s.delCacheArcPassed(ctx, needDelArc.Mid, &model.UpArc{Aid: needDelArc.Aid}, api.Without_none, api.Without_staff, api.Without_no_space)
			}); err != nil {
				log.Error("日志告警 updateArchiveList needDelArc mid:%d aid:%d error:%v", needDelArc.Mid, needDelArc.Aid, err)
				return
			}
			log.Warn("updateArchiveList success needDelArc mid:%d aid:%d", needDelArc.Mid, needDelArc.Aid)
		}()
		if needDelArc.IsStory() {
			func() {
				if err := retry(func() error {
					return s.dao.DelCacheArcStoryPassed(ctx, needDelArc.Mid, []*model.UpArc{{Aid: needDelArc.Aid}})
				}); err != nil {
					log.Error("日志告警 updateArchiveList needDelArcStory mid:%d aid:%d error:%v", needDelArc.Mid, needDelArc.Aid, err)
					return
				}
				log.Warn("updateArchiveList success needDelArcStory mid:%d aid:%d", needDelArc.Mid, needDelArc.Aid)
			}()
		}
	}
	if needDelStory != nil {
		func() {
			if err := retry(func() error {
				return s.dao.DelCacheArcStoryPassed(ctx, needDelStory.Mid, []*model.UpArc{{Aid: needDelStory.Aid}})
			}); err != nil {
				log.Error("日志告警 updateArchiveList needDelStory mid:%d aid:%d error:%v", needDelStory.Mid, needDelStory.Aid, err)
				return
			}
			log.Warn("updateArchiveList success needDelStory mid:%d aid:%d", needDelStory.Mid, needDelStory.Aid)
		}()
	}
	if needDelArc != nil || needDelStory != nil {
		var delAid int64
		if needDelArc != nil {
			delAid = needDelArc.Aid
		}
		if delAid <= 0 && needDelStory != nil {
			delAid = needDelStory.Aid
		}
		func() {
			var staffMids []int64
			if err := retry(func() error {
				var err error
				staffMids, err = s.dao.RawStaffMids(ctx, delAid)
				return err
			}); err != nil {
				log.Error("日志告警 updateArchiveList RawStaffMids aid:%d error:%v", delAid, err)
				return
			}
			if len(staffMids) == 0 {
				return
			}
			if needDelArc != nil {
				for _, staffMid := range staffMids {
					if err := retry(func() error {
						return s.delCacheArcPassed(ctx, staffMid, &model.UpArc{Aid: needDelArc.Aid}, api.Without_none, api.Without_staff, api.Without_no_space)
					}); err != nil {
						log.Error("日志告警 updateArchiveList DelCacheArcPassed staffMid:%d aid:%d error:%v", staffMid, needDelArc.Aid, err)
					}
				}
				log.Warn("updateArchiveList success DelCacheArcPassed staffMid:%v aid:%d", staffMids, delAid)
			}
			if needDelStory != nil {
				for _, staffMid := range staffMids {
					if err := retry(func() error {
						return s.dao.DelCacheArcStoryPassed(ctx, staffMid, []*model.UpArc{{Aid: needDelStory.Aid}})
					}); err != nil {
						log.Error("日志告警 updateArchiveList DelCacheArcStoryPassed staffMid:%d aid:%d error:%v", staffMid, needDelStory.Aid, err)
					}
				}
				log.Warn("updateArchiveList success DelCacheArcStoryPassed staffMid:%v aid:%d", staffMids, delAid)
			}
		}()
	}
}

func (s *Service) appendCacheArcPassed(ctx context.Context, mid int64, arc *model.UpArc, without ...api.Without) error {
	withoutm := make(map[api.Without]struct{}, len(without))
	for _, val := range without {
		switch val {
		case api.Without_none:
			withoutm[api.Without_none] = struct{}{}
		case api.Without_staff:
			withoutm[api.Without_staff] = struct{}{}
		case api.Without_no_space:
			withoutm[api.Without_no_space] = struct{}{}
		}
	}
	g := errgroup.WithContext(ctx)
	for val := range withoutm {
		without := val
		g.Go(func(ctx context.Context) error {
			if err := retry(func() error {
				return s.dao.AppendCacheArcPassed(ctx, mid, []*model.UpArc{arc}, without)
			}); err != nil {
				log.Error("日志告警 appendCacheArcPassed AppendCacheArcPassed mid:%d aid:%d without:%s error:%v", mid, arc.Aid, without.String(), err)
			}
			log.Warn("appendCacheArcPassed success AppendCacheArcPassed mid:%d aid:%d without:%s", mid, arc.Aid, without.String())
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}

func (s *Service) delCacheArcPassed(ctx context.Context, mid int64, arc *model.UpArc, without ...api.Without) error {
	withoutm := make(map[api.Without]struct{}, len(without))
	for _, val := range without {
		switch val {
		case api.Without_none:
			withoutm[api.Without_none] = struct{}{}
		case api.Without_staff:
			withoutm[api.Without_staff] = struct{}{}
		case api.Without_no_space:
			withoutm[api.Without_no_space] = struct{}{}
		}
	}
	g := errgroup.WithContext(ctx)
	for val := range withoutm {
		without := val
		g.Go(func(ctx context.Context) error {
			if err := retry(func() error {
				return s.dao.DelCacheArcPassed(ctx, mid, []*model.UpArc{arc}, without)
			}); err != nil {
				log.Error("日志告警 delCacheArcPassed DelCacheArcPassed mid:%d aid:%d without:%s error:%v", mid, arc.Aid, without.String(), err)
			}
			log.Warn("delCacheArcPassed success DelCacheArcPassed mid:%d aid:%d without:%s", mid, arc.Aid, without.String())
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}
