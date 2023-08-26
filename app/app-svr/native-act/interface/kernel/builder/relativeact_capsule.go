package builder

import (
	"context"
	"fmt"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"

	appcardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type RelativeactCapsule struct{}

func (bu RelativeactCapsule) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	rcCfg, ok := cfg.(*config.RelativeactCapsule)
	if !ok {
		logCfgAssertionError(config.RelativeactCapsule{})
		return nil
	}
	channels := bu.channels(rcCfg, material, kernel.NewMaterialLoader(c, dep, ss))
	items := bu.buildModuleItems(rcCfg, material, channels, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeRelactCapsule.String(),
		ModuleId:    rcCfg.ModuleBase().ModuleID,
		ModuleColor: &api.Color{BgColor: rcCfg.BgColor},
		ModuleItems: items,
		ModuleUkey:  rcCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu RelativeactCapsule) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu RelativeactCapsule) channels(cfg *config.RelativeactCapsule, material *kernel.Material, ml *kernel.MaterialLoader) map[int64]*channelgrpc.Channel {
	var channelIDs []int64
	for _, pid := range cfg.PageIDs {
		if _, ok := material.NativePageCards[pid]; ok {
			continue
		}
		if page, ok := material.NativeAllPages[pid]; ok && page.ForeignID > 0 {
			channelIDs = append(channelIDs, page.ForeignID)
		}
	}
	var channels map[int64]*channelgrpc.Channel
	if len(channelIDs) > 0 {
		if _, err := ml.AddItem(model.MaterialChannel, channelIDs); err == nil {
			channels = ml.Load(nil).Channels
		}
	}
	return channels
}

func (bu RelativeactCapsule) buildModuleItems(cfg *config.RelativeactCapsule, material *kernel.Material, channels map[int64]*channelgrpc.Channel, ss *kernel.Session) []*api.ModuleItem {
	cd := &api.RelativeactCapsuleCard{
		Title: cfg.TextTitle,
		Items: make([]*api.RelativeactCapsuleItem, 0, len(cfg.PageIDs)),
	}
	for _, pid := range cfg.PageIDs {
		cardItem := func() *api.RelativeactCapsuleItem {
			// NativePageCards返回上线活动，跳转优先级为：配置跳转链接 > 活动聚合页 > 单个活动页
			if naCard, ok := material.NativePageCards[pid]; ok {
				return &api.RelativeactCapsuleItem{PageId: naCard.Id, Title: naCard.Title, Uri: naCard.SkipURL}
			}
			// NativeAllPages返回剩余的活动（下线/NativePageCards失败），跳转优先级为 新频道页 > 旧频道普通话题页
			page, ok := material.NativeAllPages[pid]
			if !ok {
				return nil
			}
			channel, ok := channels[page.ForeignID]
			if !ok {
				return nil
			}
			var skipURL string
			switch channel.GetCType() {
			case model.ChannelOld:
				skipURL = appcardmdl.FillURI(appcardmdl.GotoTag, ss.RawDevice().Plat(), int(ss.RawDevice().Build),
					fmt.Sprintf("%d?type=topic", page.ForeignID), nil)
			case model.ChannelNew:
				skipURL = appcardmdl.FillURI(appcardmdl.GotoChannel, ss.RawDevice().Plat(), int(ss.RawDevice().Build),
					fmt.Sprintf("%d?tab=topic", page.ForeignID), nil)
			default:
				return nil
			}
			return &api.RelativeactCapsuleItem{PageId: page.ID, Title: page.Title, Uri: skipURL}
		}()
		if cardItem == nil {
			continue
		}
		cd.Items = append(cd.Items, cardItem)
	}
	if len(cd.Items) == 0 {
		return nil
	}
	moduleItem := &api.ModuleItem{
		CardType:   model.CardTypeRelactCapsule.String(),
		CardDetail: &api.ModuleItem_RelativeactCapsuleCard{RelativeactCapsuleCard: cd},
	}
	return []*api.ModuleItem{moduleItem}
}
