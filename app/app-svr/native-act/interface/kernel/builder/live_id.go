package builder

import (
	"context"
	"strconv"

	livexroomgrpc "git.bilibili.co/bapis/bapis-go/live/xroom-feed"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder/card"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type LiveID struct{}

func (bu LiveID) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	liveCfg, ok := cfg.(*config.LiveID)
	if !ok {
		logCfgAssertionError(config.LiveID{})
		return nil
	}
	liveItem, ok := material.LiveCard[uint64(liveCfg.ID)]
	if !ok || liveItem == nil {
		return nil
	}
	// 未开播
	if liveItem.LiveStatus != 1 && liveCfg.LiveType == 0 {
		return nil
	}
	modTmp := &api.Module{
		ModuleType:    model.ModuleTypeLive.String(),
		ModuleId:      cfg.ModuleBase().ModuleID,
		ModuleColor:   bu.buildModuleColor(liveCfg),
		ModuleUkey:    cfg.ModuleBase().Ukey,
		ModuleSetting: &api.Setting{DisplayTitle: liveCfg.DisplayTitle},
	}
	if liveCfg.ImageTitle != "" {
		modTmp.ModuleItems = append(modTmp.ModuleItems, card.NewImageTitle(liveCfg.ImageTitle).Build())
	} else if liveCfg.TextTitle != "" {
		modTmp.ModuleItems = append(modTmp.ModuleItems, card.NewTextTitle(liveCfg.TextTitle).Build())
	}
	defCover := ""
	if liveItem.LiveStatus == 1 && liveCfg.Cover != "" {
		defCover = liveCfg.Cover
	}
	cd := &api.LiveCard{
		Content: bu.buildLiveItem(liveItem, defCover),
	}
	if liveCfg.Stime < liveItem.LastEndTime {
		cd.HasLive = 1
	}
	modTmp.ModuleItems = append(modTmp.ModuleItems, &api.ModuleItem{
		CardType:   model.CardTypeLive.String(),
		CardId:     strconv.FormatInt(liveCfg.ID, 10),
		CardDetail: &api.ModuleItem_LiveCard{LiveCard: cd},
	})
	return modTmp
}

func (bu LiveID) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu LiveID) buildModuleColor(cfg *config.LiveID) *api.Color {
	return &api.Color{BgColor: cfg.BgColor, FontColor: cfg.FontColor, TextTitleFontColor: cfg.DisplayColor}
}

func (bu LiveID) buildLiveItem(liveItem *livexroomgrpc.LiveCardInfo, defCover string) *api.LiveItem {
	rly := &api.LiveItem{
		RoomId:         liveItem.RoomId,
		Uid:            liveItem.Uid,
		LiveStatus:     liveItem.LiveStatus,
		RoomType:       liveItem.RoomType,
		PlayType:       liveItem.PlayType,
		Title:          liveItem.Title,
		Cover:          liveItem.Cover,
		Online:         liveItem.Online,
		AreaId:         liveItem.AreaId,
		AreaName:       liveItem.AreaName,
		ParentAreaId:   liveItem.ParentAreaId,
		ParentAreaName: liveItem.ParentAreaName,
		LiveScreenType: liveItem.LiveScreenType,
		LastEndTime:    liveItem.LastEndTime,
		Link:           liveItem.Link,
		LiveId:         liveItem.LiveId,
	}
	if defCover != "" {
		rly.Cover = defCover
	}
	if ws := liveItem.WatchedShow; ws != nil {
		rly.WatchedShow = &api.LiveWatchedShow{
			Switch:       ws.Switch,
			Num:          ws.Num,
			TextSmall:    ws.TextSmall,
			TextLarge:    ws.TextLarge,
			Icon:         ws.Icon,
			IconLocation: ws.IconLocation,
		}
	}
	return rly
}
