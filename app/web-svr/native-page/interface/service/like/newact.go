package like

import (
	"context"
	"fmt"
	"time"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"

	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
	dynmdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"
)

func (s *Service) formatNewactHeader(c context.Context, mou *natpagegrpc.NativeModule) *dynmdl.Item {
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
	header := &dynmdl.Item{
		Goto:         dynmdl.GotoNewactHeader,
		Title:        st.Name,
		Time:         fmt.Sprintf("任务时间：%s-%s", st.Stime.Time().Format("2006.01.02"), st.Etime.Time().Format("2006.01.02")),
		Image:        st.ActivityImage,
		SponsorTitle: "发起",
	}
	if accInfo != nil {
		header.UserInfo = &dynmdl.UserInfo{
			Mid:  accInfo.GetMid(),
			Name: accInfo.GetName(),
			Face: accInfo.GetFace(),
		}
	}
	if time.Now().After(st.Etime.Time()) {
		header.NewactFeatures = append(header.NewactFeatures, &dynmdl.Item{Title: "已结束", Color: &dynmdl.Color{BorderColor: "#9499A0"}})
	}
	for _, ft := range st.ActFeature {
		if ft == nil {
			continue
		}
		header.NewactFeatures = append(header.NewactFeatures, &dynmdl.Item{Title: ft.Title, Color: &dynmdl.Color{BorderColor: "#FF7F24"}})
	}
	item := &dynmdl.Item{}
	item.FromNewactHeaderModule(mou, []*dynmdl.Item{header})
	return item
}

func (s *Service) formatNewactAward(c context.Context, mou *natpagegrpc.NativeModule) *dynmdl.Item {
	if mou.Fid == 0 {
		return nil
	}
	stRly, err := s.actDao.ActSubject(c, mou.Fid)
	if err != nil || stRly == nil || stRly.Subject == nil {
		return nil
	}
	st := stRly.Subject
	award := &dynmdl.Item{
		Goto:  dynmdl.GotoNewactAward,
		Title: "活动奖励",
	}
	for _, ad := range st.ActivityAward {
		if ad == nil {
			continue
		}
		award.Item = append(award.Item, &dynmdl.Item{Title: ad.Title, Content: ad.Desc})
	}
	if len(award.Item) == 0 {
		return nil
	}
	item := &dynmdl.Item{}
	item.FromNewactAwardModule(mou, []*dynmdl.Item{award})
	return item
}

func (s *Service) formatNewactStatement(c context.Context, mou *natpagegrpc.NativeModule) *dynmdl.Item {
	if mou.Fid == 0 {
		return nil
	}
	stRly, err := s.actDao.ActSubject(c, mou.Fid)
	if err != nil || stRly == nil || stRly.Subject == nil {
		return nil
	}
	var childItem *dynmdl.Item
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
	item := &dynmdl.Item{}
	item.FromNewactStatementModule(mou, []*dynmdl.Item{childItem})
	return item
}

func newactTask(st *actgrpc.Subject) *dynmdl.Item {
	item := &dynmdl.Item{
		Goto:  dynmdl.GotoNewactStatement,
		Title: "任务玩法",
	}
	if st.AuditPlatformNew != nil && st.AuditPlatformNew.Rule != "" {
		item.Item = append(item.Item, &dynmdl.Item{Title: "必选要求*", Content: st.AuditPlatformNew.Rule})
	}
	if st.SelectAsk != "" {
		item.Item = append(item.Item, &dynmdl.Item{Title: "可选要求", Content: st.SelectAsk})
	}
	if len(item.Item) == 0 {
		return nil
	}
	return item
}

func newactRule(st *actgrpc.Subject) *dynmdl.Item {
	if st.RuleExplain == "" {
		return nil
	}
	return &dynmdl.Item{
		Goto:  dynmdl.GotoNewactStatement,
		Title: "规则说明",
		Item: []*dynmdl.Item{
			{Content: st.RuleExplain},
		},
	}
}

func newactDeclaration(st *actgrpc.Subject) *dynmdl.Item {
	if st.PlatStatement == "" {
		return nil
	}
	return &dynmdl.Item{
		Goto:  dynmdl.GotoNewactStatement,
		Title: "平台声明",
		Item: []*dynmdl.Item{
			{Content: st.PlatStatement},
		},
	}
}
