package teenagers

import (
	"context"
	"strconv"
	"time"

	bgroupgrpc "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	accountmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/account"
	model "go-gateway/app/app-svr/app-interface/interface-legacy/model/anti_addiction"
	pushmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/push"
)

var (
	adultSetup    = &model.SleepRemindSetup{Switch: false, Stime: "23:00", Etime: "07:00"}
	underAgeSetup = &model.SleepRemindSetup{Switch: true, Stime: "23:00", Etime: "07:00"}
)

func (s *Service) AntiAddictionRule(ctx context.Context, mid int64) (*model.RuleRly, error) {
	if s.c.AntiAddictionRule == nil || !s.c.AntiAddictionRule.Switch {
		return &model.RuleRly{}, nil
	}
	cfg := s.c.AntiAddictionRule
	inRly, _ := s.bgroupDao.MemberIn(ctx, &bgroupgrpc.MemberInReq{
		Member: strconv.FormatInt(mid, 10),
		Groups: []*bgroupgrpc.MemberInReq_MemberInReqSingle{
			{Business: cfg.BgroupBiz, Name: cfg.BgroupName},
		},
		Dimension: bgroupgrpc.Mid,
	})
	if in, ok := inRly[cfg.BgroupName]; !ok || !in {
		return &model.RuleRly{}, nil
	}
	return &model.RuleRly{
		Rules: buildDefaultRule(cfg),
	}, nil
}

func (s *Service) AggregationStatus(ctx context.Context, req *model.AggregationStatusReq, mid int64) (*model.AggregationStatusRly, error) {
	bizTypes := make(map[string]struct{})
	for _, typ := range req.BizTypes {
		bizTypes[typ] = struct{}{}
	}
	isNeeded := func(typ string) bool {
		if len(bizTypes) == 0 {
			return true
		}
		_, ok := bizTypes[typ]
		return ok
	}
	rly := &model.AggregationStatusRly{}
	eg := errgroup.WithContext(ctx)
	if isNeeded(model.BizAntiAddiction) && mid > 0 {
		eg.Go(func(ctx context.Context) error {
			rly.AntiAddiction, _ = s.AntiAddictionRule(ctx, mid)
			return nil
		})
	}
	if isNeeded(model.BizSleepRemind) {
		eg.Go(func(ctx context.Context) error {
			rly.SleepRemind = s.sleepRemind(ctx, mid)
			return nil
		})
	}
	if isNeeded(model.BizTimelock) && mid > 0 {
		eg.Go(func(ctx context.Context) error {
			rly.FamilyTimelock = s.timelock(ctx, mid)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("Fail to handle AggregationStatus, mid=%+v error=%+v", mid, err)
	}
	return rly, nil
}

func (s *Service) sleepRemind(ctx context.Context, mid int64) (st *model.SleepRemindSetup) {
	defer func() {
		if st == nil {
			return
		}
		st.Push = &pushmdl.Message{
			Title:     "夜深了，睡个好觉吧~",
			Summary:   "点击设置睡眠时间",
			Position:  1,                                  //顶部
			Duration:  3,                                  //3s
			Expire:    time.Now().AddDate(1, 0, 0).Unix(), //1年后
			MsgSource: pushmdl.MsgSourceSleepRemind,
			HideArrow: false,
			Link:      "bilibili://user_center/setting/sleep_remind",
		}
	}()
	if !s.graySleepRemind(ctx, mid) {
		return nil
	}
	if mid == 0 {
		return underAgeSetup
	}
	sr, err := s.userModelDao.SleepRemind(ctx, mid)
	if err != nil {
		return nil
	}
	if sr != nil {
		return &model.SleepRemindSetup{Switch: sr.Switch == 1, Stime: sr.Stime, Etime: sr.Etime}
	}
	ageCheck, err := s.accountDao.RealnameTeenAgeCheck(ctx, mid, metadata.String(ctx, metadata.RemoteIP))
	if err == nil && ageCheck != nil && ageCheck.Realname == accountmdl.RealnameVerified && !ageCheck.After18 {
		return underAgeSetup
	}
	return adultSetup
}

func (s *Service) SetSleepRemind(ctx context.Context, req *model.SetSleepRemindReq, mid int64) error {
	oldVal, err := s.userModelDao.SleepRemind(ctx, mid)
	if err != nil {
		return err
	}
	newVal := &model.SleepRemind{
		Mid:    mid,
		Switch: 0,
		Stime:  req.Stime,
		Etime:  req.Etime,
	}
	if req.Switch {
		newVal.Switch = 1
	}
	if oldVal != nil {
		newVal.ID = oldVal.ID
		return s.userModelDao.UpdateSleepRemind(ctx, newVal)
	}
	return s.userModelDao.AddSleepRemind(ctx, newVal)
}

func buildDefaultRule(cfg *conf.AntiAddictionRule) []*model.Rule {
	rule := &model.Rule{
		Id:        cfg.Id,
		Version:   cfg.Version,
		Frequency: cfg.Frequency,
		Conditions: []*model.Condition{
			{
				Type: model.CondSeries,
				Series: &model.ConditionSeries{
					MaxDuration: int64(time.Duration(cfg.SeriesDuration) / time.Second),
					Interval:    int64(time.Duration(cfg.SeriesInterval) / time.Second),
				},
			},
		},
		Control: &model.Control{
			Type: model.CtrPush,
			Push: &model.ControlPush{
				Message: &pushmdl.Message{
					Title:     cfg.PushTitle,
					Summary:   cfg.PushSubtitle,
					Position:  1,                                  //顶部
					Duration:  3,                                  //3s
					Expire:    time.Now().AddDate(1, 0, 0).Unix(), //1年后
					MsgSource: pushmdl.MsgSourceUnderage,
					HideArrow: true,
				},
			},
		},
	}
	return []*model.Rule{rule}
}

func (s *Service) graySleepRemind(ctx context.Context, mid int64) bool {
	if s.c.SleepRemind == nil || !s.c.SleepRemind.Switch {
		// 关闭睡眠提醒
		return false
	}
	if s.c.SleepRemind.FullSwitch {
		// 全量开启
		return true
	}
	if mid == 0 {
		// 非全量,灰度只针对登陆用户
		return false
	}
	for _, v := range s.c.SleepRemind.Whitelist {
		if v == mid {
			return true
		}
	}
	inRly, _ := s.bgroupDao.MemberIn(ctx, &bgroupgrpc.MemberInReq{
		Member: strconv.FormatInt(mid, 10),
		Groups: []*bgroupgrpc.MemberInReq_MemberInReqSingle{
			{Business: s.c.SleepRemind.BgroupBiz, Name: s.c.SleepRemind.BgroupName},
		},
		Dimension: bgroupgrpc.Mid,
	})
	if in, ok := inRly[s.c.SleepRemind.BgroupName]; !ok || !in {
		return false
	}
	return true
}
