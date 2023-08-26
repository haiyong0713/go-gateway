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

	"github.com/pkg/errors"
)

const _archiveStaffTable = "archive_staff"

func (s *Service) initStaffRailGun(cfg *railgun.DatabusV1Config, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.staffRailGunUnpack, s.staffRailGunDo)
	g := railgun.NewRailGun("稿件联合创作变更", nil, inputer, processor)
	s.staffRailGun = g
	g.Start()
}

func (s *Service) staffRailGunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	log.Warn("接收稿件联合创作变更消息成功,data:%s", msg.Payload())
	canalMsg := new(model.CanalMsg)
	if err := json.Unmarshal(msg.Payload(), &canalMsg); err != nil {
		return nil, err
	}
	switch canalMsg.Table {
	case _archiveStaffTable:
		arcCanalMsg := new(model.ArchiveStaffCanalMsg)
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

func (s *Service) staffRailGunDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	staffCanalMsg := item.(*model.ArchiveStaffCanalMsg)
	if staffCanalMsg.New == nil {
		return railgun.MsgPolicyIgnore
	}
	var needInsert bool
	switch staffCanalMsg.Action {
	case model.ActionInsert:
		needInsert = true
	case model.ActionUpdate:
		if staffCanalMsg.Old == nil {
			return railgun.MsgPolicyIgnore
		}
		if staffCanalMsg.New.State == staffCanalMsg.Old.State { // no state change
			return railgun.MsgPolicyIgnore
		}
		if staffCanalMsg.New.State == model.StaffStateNormal {
			needInsert = true
		}
	}
	if needInsert {
		func() {
			arc, fItem, err := s.arc(ctx, staffCanalMsg.New.Aid)
			if err != nil {
				log.Error("日志告警 staffRailGunDo s.arc aid:%d error:%v", staffCanalMsg.New.Aid, err)
				return
			}
			if arc == nil {
				return
			}
			if !arc.IsAllowed(fItem) {
				return
			}
			arc.RandScoreNumber()
			if err := retry(func() error {
				without := []api.Without{api.Without_none}
				if arc.Mid != staffCanalMsg.New.Mid { //联合投稿稿件的副投稿人忽略空间防刷
					without = append(without, api.Without_no_space)
				} else if !arc.IsUpNoSpace(fItem) {
					without = append(without, api.Without_no_space)
				}
				return s.appendCacheArcPassed(ctx, staffCanalMsg.New.StaffMid, arc, without...)
			}); err != nil {
				log.Error("日志告警 staffRailGunDo AppendCacheArcPassed staffMid:%d arc:%+v error:%v", staffCanalMsg.New.StaffMid, arc, err)
			}
			if arc.IsStory() {
				if err := retry(func() error {
					return s.dao.AppendCacheArcStoryPassed(ctx, staffCanalMsg.New.StaffMid, []*model.UpArc{arc})
				}); err != nil {
					log.Error("日志告警 staffRailGunDo AppendCacheArcStoryPassed staffMid:%d arc:%+v error:%v", staffCanalMsg.New.StaffMid, arc, err)
				}
			}
			log.Warn("staffRailGunDo success AppendCacheArcPassed mid:%d arc:%+v isStory:%v", staffCanalMsg.New.StaffMid, arc, arc.IsStory())
		}()
		return railgun.MsgPolicyNormal
	}
	var fItem []*cfcgrpc.ForbiddenItem
	if err := retry(func() error {
		var err error
		fItem, err = s.dao.ContentFlowControlInfo(ctx, staffCanalMsg.New.Aid)
		return err
	}); err != nil {
		log.Error("日志告警 staffRailGunDo ContentFlowControlInfo aid:%d error:%v", staffCanalMsg.New.Aid, err)
		return railgun.MsgPolicyFailure
	}
	if err := retry(func() error {
		return s.delCacheArcPassed(ctx, staffCanalMsg.New.StaffMid, &model.UpArc{Aid: staffCanalMsg.New.Aid}, api.Without_none, api.Without_no_space)
	}); err != nil {
		log.Error("日志告警 staffRailGunDo s.delCacheArcPassed mid:%d aid:%d error:%v", staffCanalMsg.New.StaffMid, staffCanalMsg.New.Aid, err)
	}
	if err := retry(func() error {
		return s.dao.DelCacheArcStoryPassed(ctx, staffCanalMsg.New.StaffMid, []*model.UpArc{{Aid: staffCanalMsg.New.Aid}})
	}); err != nil {
		log.Error("日志告警 staffRailGunDo DelCacheArcStoryPassed mid:%d aid:%d error:%v", staffCanalMsg.New.StaffMid, staffCanalMsg.New.Aid, err)
	}
	log.Warn("staffRailGunDo success DelCacheArcPassed mid:%d aid:%d", staffCanalMsg.New.StaffMid, staffCanalMsg.New.Aid)
	return railgun.MsgPolicyNormal
}

func (s *Service) arc(ctx context.Context, aid int64) (*model.UpArc, []*cfcgrpc.ForbiddenItem, error) {
	var (
		arc   *model.UpArc
		fItem []*cfcgrpc.ForbiddenItem
	)
	group := errgroup.WithContext(ctx)
	group.Go(func(ctx context.Context) error {
		if err := retry(func() (err error) {
			arc, err = s.dao.RawArc(ctx, aid)
			return err
		}); err != nil {
			return errors.Wrapf(err, "RawArc aid:%v", aid)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if err := retry(func() (err error) {
			fItem, err = s.dao.ContentFlowControlInfo(ctx, aid)
			return err
		}); err != nil {
			return errors.Wrapf(err, "ContentFlowControlInfo aid:%v", aid)
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		return nil, nil, err
	}
	return arc, fItem, nil
}
