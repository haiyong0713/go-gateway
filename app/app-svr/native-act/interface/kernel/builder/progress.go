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

type Progress struct{}

func (bu Progress) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	pgCfg, ok := cfg.(*config.Progress)
	if !ok {
		logCfgAssertionError(config.Progress{})
		return nil
	}
	items := bu.buildModuleItems(pgCfg, material)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeProgress.String(),
		ModuleId:      pgCfg.ModuleBase().ModuleID,
		ModuleColor:   bu.buildColor(pgCfg),
		ModuleSetting: bu.buildSetting(pgCfg),
		ModuleItems:   items,
		ModuleUkey:    pgCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Progress) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Progress) buildColor(cfg *config.Progress) *api.Color {
	return &api.Color{
		BgColor:          cfg.BgColor,
		ProgressBarColor: cfg.BarColor,
	}
}

func (bu Progress) buildSetting(cfg *config.Progress) *api.Setting {
	return &api.Setting{
		DisplayProgressNum: cfg.DisplayProgressNum,
		DisplayNodeNum:     cfg.DisplayNodeNum,
		DisplayNodeDesc:    cfg.DisplayNodeDesc,
	}
}

func (bu Progress) buildModuleItems(cfg *config.Progress, material *kernel.Material) []*api.ModuleItem {
	group, ok := material.ActProgressGroups[cfg.Sid][cfg.GroupID]
	if !ok {
		return nil
	}
	nodes := bu.buildNodes(group.Nodes)
	if len(nodes) == 0 {
		return nil
	}
	cd := &api.ProgressCard{
		Style:      bu.style(cfg.Style),
		SlotType:   bu.slotType(cfg.SlotType),
		BarType:    bu.barType(cfg.BarType),
		Num:        group.Total,
		DisplayNum: StatString(group.Total),
		Nodes:      nodes,
	}
	if cd.BarType == api.ProgressBar_PBarTexture {
		cd.TextureImage = bu.textureImage(cfg.TextureType)
	}
	moduleItem := &api.ModuleItem{
		CardType:   model.CardTypeProgress.String(),
		CardId:     strconv.FormatInt(cfg.GroupID, 10),
		CardDetail: &api.ModuleItem_ProgressCard{ProgressCard: cd},
	}
	return []*api.ModuleItem{moduleItem}
}

func (bu Progress) buildNodes(nodes []*activitygrpc.ActivityProgressNodeInfo) []*api.ProgressNode {
	items := make([]*api.ProgressNode, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		items = append(items, &api.ProgressNode{
			Name:       node.Desc,
			Num:        node.Val,
			DisplayNum: StatString(node.Val),
		})
	}
	return items
}

var _progressStyle = map[int64]api.ProgressStyle{
	model.PgStyleRound:     api.ProgressStyle_PStyleRound,
	model.PgStyleRectangle: api.ProgressStyle_PStyleRectangle,
	model.PgStyleNode:      api.ProgressStyle_PStyleNode,
}

func (bu Progress) style(in int64) api.ProgressStyle {
	if out, ok := _progressStyle[in]; ok {
		return out
	}
	return api.ProgressStyle_PStyleDefault
}

var _progressSlot = map[string]api.ProgressSlot{
	model.PgSlotOutline: api.ProgressSlot_PSlotOutline,
	model.PgSlotFill:    api.ProgressSlot_PSlotFill,
}

func (bu Progress) slotType(in string) api.ProgressSlot {
	if out, ok := _progressSlot[in]; ok {
		return out
	}
	return api.ProgressSlot_PSlotDefault
}

var _progressBar = map[string]api.ProgressBar{
	model.PgBarColor:   api.ProgressBar_PBarColor,
	model.PgBarTexture: api.ProgressBar_PBarTexture,
}

func (bu Progress) barType(in string) api.ProgressBar {
	if out, ok := _progressBar[in]; ok {
		return out
	}
	return api.ProgressBar_PBarDefault
}

var _progressTextureImage = map[int64]string{
	model.PgTexture1: "https://i0.hdslb.com/bfs/activity-plat/static/20200811/8a3e1fa14e30dc3be9c5324f604e5991/I2CHCvboo.png",
	model.PgTexture2: "https://i0.hdslb.com/bfs/activity-plat/static/20200811/8a3e1fa14e30dc3be9c5324f604e5991/bxL-soEal.png",
	model.PgTexture3: "https://i0.hdslb.com/bfs/activity-plat/static/20200811/8a3e1fa14e30dc3be9c5324f604e5991/vniA~jzxp.png",
}

func (bu Progress) textureImage(in int64) string {
	if out, ok := _progressTextureImage[in]; ok {
		return out
	}
	return ""
}
