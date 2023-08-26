package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Game struct{}

func (r Game) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	gameRes := module.Game
	//web暂时不支持游戏组件,没有游戏ids
	if ss.IsWeb() || gameRes == nil || len(gameRes.List) == 0 {
		return nil
	}
	var (
		mobiapp, device string
	)
	if ss.IsH5() {
		mobiapp = ss.RawWebdevice().MobiApp
	} else {
		mobiapp = ss.RawDevice().RawMobiApp
		device = ss.RawDevice().Device
	}
	if mobiapp != "android" && (mobiapp != "iphone" || device == "pad") { //仅支持粉版
		return nil
	}
	var (
		IDs     = make([]*config.GameID, 0)
		gameIDs []int64
	)
	//获取游戏id
	for _, v := range gameRes.List {
		if v == nil || v.MType != natpagegrpc.MixGame || v.ForeignID <= 0 {
			continue
		}
		gameIDs = append(gameIDs, v.ForeignID)
		mixReason := v.RemarkUnmarshal()
		IDs = append(IDs, &config.GameID{ID: v.ForeignID, Remark: mixReason.Desc})
	}
	cfg := &config.Game{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ImageTitle:     natModule.Meta,
		TextTitle:      natModule.Caption,
		IDs:            IDs,
		//背景色
		BgColor: natModule.BgColor,
		//卡片标题文字色
		TitleColor: natModule.TitleColor,
	}
	_, _ = cfg.BaseCfgManager.AddMaterialParam(model.MaterialGame, gameIDs)
	return cfg
}
