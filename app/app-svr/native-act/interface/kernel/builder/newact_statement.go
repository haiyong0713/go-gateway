package builder

import (
	"context"
	"strconv"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type NewactStatement struct{}

func (bu NewactStatement) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	nsCfg, ok := cfg.(*config.NewactStatement)
	if !ok {
		logCfgAssertionError(config.NewactStatement{})
		return nil
	}
	items := bu.buildModuleItems(nsCfg, material)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeNewactStatement.String(),
		ModuleId:    nsCfg.ModuleBase().ModuleID,
		ModuleItems: items,
		ModuleUkey:  nsCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu NewactStatement) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu NewactStatement) buildModuleItems(cfg *config.NewactStatement, material *kernel.Material) []*api.ModuleItem {
	st, ok := material.ActSubjects[cfg.ReqID][cfg.Sid]
	if !ok {
		return nil
	}
	var cd *api.NewactStatement
	switch cfg.Type {
	case model.StatementNewactTask:
		cd = bu.buildTask(st)
	case model.StatementNewactRule:
		cd = bu.buildRule(st)
	case model.StatementNewactDeclaration:
		cd = bu.buildDeclaration(st)
	}
	if cd == nil {
		return nil
	}
	moduleItem := &api.ModuleItem{
		CardType:   model.CardTypeNewactStatement.String(),
		CardId:     strconv.FormatInt(cfg.Sid, 10),
		CardDetail: &api.ModuleItem_NewactStatementCard{NewactStatementCard: cd},
	}
	return []*api.ModuleItem{moduleItem}
}

func (bu NewactStatement) buildTask(st *activitygrpc.Subject) *api.NewactStatement {
	cd := &api.NewactStatement{
		Title: "任务玩法",
		Items: nil,
	}
	if st.AuditPlatformNew != nil && st.AuditPlatformNew.Rule != "" {
		cd.Items = append(cd.Items, &api.NewactStatementItem{Title: "必选要求*", Content: st.AuditPlatformNew.Rule})
	}
	if st.SelectAsk != "" {
		cd.Items = append(cd.Items, &api.NewactStatementItem{Title: "可选要求", Content: st.SelectAsk})
	}
	if len(cd.Items) == 0 {
		return nil
	}
	return cd
}

func (bu NewactStatement) buildRule(st *activitygrpc.Subject) *api.NewactStatement {
	if st.RuleExplain == "" {
		return nil
	}
	return &api.NewactStatement{
		Title: "规则说明",
		Items: []*api.NewactStatementItem{
			{Content: st.RuleExplain},
		},
	}
}

func (bu NewactStatement) buildDeclaration(st *activitygrpc.Subject) *api.NewactStatement {
	if st.PlatStatement == "" {
		return nil
	}
	return &api.NewactStatement{
		Title: "平台声明",
		Items: []*api.NewactStatementItem{
			{Content: st.PlatStatement},
		},
	}
}
