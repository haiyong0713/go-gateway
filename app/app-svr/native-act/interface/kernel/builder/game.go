package builder

import (
	"context"
	"strconv"
	"strings"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder/card"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type Game struct{}

func (bu Game) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	gaCfg, ok := cfg.(*config.Game)
	if !ok {
		logCfgAssertionError(config.Game{})
		return nil
	}
	modTmp := &api.Module{
		ModuleType:  model.ModuleTypeGame.String(),
		ModuleId:    cfg.ModuleBase().ModuleID,
		ModuleColor: bu.buildModuleColor(gaCfg),
		ModuleUkey:  cfg.ModuleBase().Ukey,
	}
	if gaCfg.ImageTitle != "" {
		modTmp.ModuleItems = append(modTmp.ModuleItems, card.NewImageTitle(gaCfg.ImageTitle).Build())
	} else if gaCfg.TextTitle != "" {
		modTmp.ModuleItems = append(modTmp.ModuleItems, card.NewTextTitle(gaCfg.TextTitle).Build())
	}
	tmp := make([]*api.ModuleItem, 0)
	for _, v := range gaCfg.IDs {
		if v == nil {
			continue
		}
		item, ok := material.GameCard[v.ID]
		if !ok || item == nil {
			continue
		}
		cd := bu.buildModuleCard(v.Remark, item)
		tmp = append(tmp, &api.ModuleItem{
			CardType:   model.CardTypeGame.String(),
			CardId:     strconv.FormatInt(v.ID, 10),
			CardDetail: &api.ModuleItem_GameCard{GameCard: cd},
		})
	}
	//没有游戏卡，不下发组件
	if len(tmp) == 0 {
		return nil
	}
	modTmp.ModuleItems = append(modTmp.ModuleItems, tmp...)
	return modTmp
}

func (bu Game) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Game) buildModuleColor(cfg *config.Game) *api.Color {
	return &api.Color{BgColor: cfg.BgColor, CardTitleFontColor: cfg.TitleColor}
}

func (bu Game) buildModuleCard(remark string, item *model.GaItem) *api.GameCard {
	if remark == "" {
		remark = item.GameSubtitle
	}
	return &api.GameCard{
		Image:    item.GameIcon,
		Title:    item.GameName,
		Uri:      item.GameLink,
		Subtitle: remark,
		Content:  strings.Join(item.GameTags, "/"),
	}
}
