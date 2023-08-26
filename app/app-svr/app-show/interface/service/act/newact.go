package act

import (
	"context"
	"fmt"
	"time"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"

	actmdl "go-gateway/app/app-svr/app-show/interface/model/act"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

func (s *Service) FormatNewactHeader(c context.Context, mou *natpagegrpc.NativeModule) *actmdl.Item {
	if mou.Fid == 0 {
		return nil
	}
	stRly, err := s.actDao.ActSubject(c, mou.Fid)
	if err != nil || stRly == nil || stRly.Subject == nil {
		return nil
	}
	st := stRly.Subject
	var accInfo *accgrpc.Card
	if st.ActivityInitiator > 0 {
		if cards, err := s.accDao.Cards3GRPC(c, []int64{st.ActivityInitiator}); err == nil {
			if info, ok := cards[st.ActivityInitiator]; ok {
				accInfo = info
			}
		}
	}
	header := &actmdl.Item{
		Goto:         actmdl.GotoNewactHeader,
		Title:        st.Name,
		Time:         fmt.Sprintf("任务时间：%s-%s", st.Stime.Time().Format("2006.01.02"), st.Etime.Time().Format("2006.01.02")),
		Image:        st.ActivityImage,
		SponsorTitle: "发起",
	}
	if accInfo != nil {
		header.UserInfo = &actmdl.UserInfo{
			Mid:  accInfo.GetMid(),
			Name: accInfo.GetName(),
			Face: accInfo.GetFace(),
			Url:  fmt.Sprintf("bilibili://space/%d?defaultTab=dynamic", accInfo.GetMid()),
		}
	}
	if time.Now().After(st.Etime.Time()) {
		header.NewactFeatures = append(header.NewactFeatures, &actmdl.Item{Title: "已结束", Color: &actmdl.Color{BorderColor: "#9499A0"}})
	}
	for _, ft := range st.ActFeature {
		if ft == nil {
			continue
		}
		header.NewactFeatures = append(header.NewactFeatures, &actmdl.Item{Title: ft.Title, Color: &actmdl.Color{BorderColor: "#FF7F24"}})
	}
	item := &actmdl.Item{}
	item.FromNewactHeaderModule(mou, []*actmdl.Item{header})
	return item
}

func (s *Service) FormatNewactAward(c context.Context, mou *natpagegrpc.NativeModule) *actmdl.Item {
	if mou.Fid == 0 {
		return nil
	}
	stRly, err := s.actDao.ActSubject(c, mou.Fid)
	if err != nil || stRly == nil || stRly.Subject == nil {
		return nil
	}
	st := stRly.Subject
	award := &actmdl.Item{
		Goto:  actmdl.GotoNewactAward,
		Title: "活动奖励",
	}
	for _, ad := range st.ActivityAward {
		if ad == nil {
			continue
		}
		award.Item = append(award.Item, &actmdl.Item{Title: ad.Title, Content: ad.Desc})
	}
	if len(award.Item) == 0 {
		return nil
	}
	item := &actmdl.Item{}
	item.FromNewactAwardModule(mou, []*actmdl.Item{award})
	return item
}

func (s *Service) FormatNewactStatement(c context.Context, mou *natpagegrpc.NativeModule) *actmdl.Item {
	if mou.Fid == 0 {
		return nil
	}
	stRly, err := s.actDao.ActSubject(c, mou.Fid)
	if err != nil || stRly == nil || stRly.Subject == nil {
		return nil
	}
	var childItem *actmdl.Item
	confSort := mou.ConfUnmarshal()
	switch confSort.StatementType {
	case natpagegrpc.StatementNewactTask:
		childItem = newactTask(stRly.Subject)
	case natpagegrpc.StatementNewactRule:
		childItem = newactRule(stRly.Subject)
	case natpagegrpc.StatementNewactDeclaration:
		childItem = newactDeclaration(stRly.Subject)
	}
	if childItem == nil {
		return nil
	}
	item := &actmdl.Item{}
	item.FromNewactStatementModule(mou, []*actmdl.Item{childItem})
	return item
}

func newactTask(st *actgrpc.Subject) *actmdl.Item {
	item := &actmdl.Item{
		Goto:  actmdl.GotoNewactStatement,
		Title: "任务玩法",
	}
	if st.AuditPlatformNew != nil && st.AuditPlatformNew.Rule != "" {
		item.Item = append(item.Item, &actmdl.Item{Title: "必选要求*", Content: st.AuditPlatformNew.Rule})
	}
	if st.SelectAsk != "" {
		item.Item = append(item.Item, &actmdl.Item{Title: "可选要求", Content: st.SelectAsk})
	}
	if len(item.Item) == 0 {
		return nil
	}
	return item
}

func newactRule(st *actgrpc.Subject) *actmdl.Item {
	if st.RuleExplain == "" {
		return nil
	}
	return &actmdl.Item{
		Goto:  actmdl.GotoNewactStatement,
		Title: "规则说明",
		Item: []*actmdl.Item{
			{Content: st.RuleExplain},
		},
	}
}

func newactDeclaration(st *actgrpc.Subject) *actmdl.Item {
	if st.PlatStatement == "" {
		return nil
	}
	return &actmdl.Item{
		Goto:  actmdl.GotoNewactStatement,
		Title: "平台声明",
		Item: []*actmdl.Item{
			{Content: st.PlatStatement},
		},
	}
}
