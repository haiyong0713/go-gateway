package builder

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type Reply struct{}

func (bu Reply) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	replyCfg, ok := cfg.(*config.Reply)
	if !ok {
		logCfgAssertionError(config.Reply{})
		return nil
	}
	items := bu.buildModuleItems(replyCfg)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeReply.String(),
		ModuleId:    replyCfg.ModuleBase().ModuleID,
		ModuleItems: items,
		ModuleUkey:  replyCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Reply) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Reply) buildModuleItems(cfg *config.Reply) []*api.ModuleItem {
	item := &api.ModuleItem{
		CardType: model.CardTypeReply.String(),
		CardId:   strconv.FormatInt(cfg.ReplyID, 10),
		CardDetail: &api.ModuleItem_ReplyCard{
			ReplyCard: &api.ReplyCard{
				ReplyId: cfg.ReplyID,
				Type:    cfg.Type,
			},
		},
	}
	return []*api.ModuleItem{item}
}
