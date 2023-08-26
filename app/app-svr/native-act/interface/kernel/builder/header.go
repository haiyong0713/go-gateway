package builder

import (
	"context"
	"strconv"

	appcardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

const (
	_headerHighImage = "http://i0.hdslb.com/bfs/activity-plat/static/20200616/82ac2611e49c304c91fb79cc76b9b762/eDsMwRL6Y.png"
	_headerLowImage  = "http://i0.hdslb.com/bfs/activity-plat/static/20200616/82ac2611e49c304c91fb79cc76b9b762/IGGf9Vleq.png"
)

type Header struct{}

func (bu Header) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	headerCfg, ok := cfg.(*config.Header)
	if !ok {
		logCfgAssertionError(config.Header{})
		return nil
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeHeader.String(),
		ModuleId:      cfg.ModuleBase().ModuleID,
		ModuleColor:   bu.buildModuleColor(headerCfg),
		ModuleSetting: bu.buildSetting(headerCfg, material),
		ModuleItems:   bu.buildModuleItems(headerCfg, material),
		ModuleUkey:    cfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Header) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Header) buildModuleColor(cfg *config.Header) *api.Color {
	return &api.Color{BgColor: cfg.BgColor}
}

func (bu Header) buildSetting(cfg *config.Header, material *kernel.Material) *api.Setting {
	setting := &api.Setting{
		DisplayViewNum:      cfg.DisplayViewNum,
		DisplaySubscribeBtn: cfg.DisplaySubscribeBtn,
	}
	if stat, ok := material.ActiveUsersRlys[cfg.ActiveUsersReqID]; !(ok && stat.ViewCount > 0 && stat.DiscussCount > 0) {
		setting.DisplayViewNum = false
	}
	if _, ok := material.Tags[cfg.TopicID]; !ok {
		setting.DisplaySubscribeBtn = false
	}
	return setting
}

func (bu Header) buildModuleItems(cfg *config.Header, material *kernel.Material) []*api.ModuleItem {
	cd := &api.HeaderCard{}
	if cfg.DisplayUser {
		cd.SponsorContent = cfg.SponsorContent
		cd.HighLightImage = _headerHighImage
		cd.LowLightImage = _headerLowImage
		if user, ok := material.Accounts[cfg.SponsorMid]; ok && user != nil {
			cd.Mid = user.Mid
			cd.UserName = user.Name
			cd.UserImage = user.Face
			cd.Uri = appcardmdl.FillURI(appcardmdl.GotoDynamicMid, 0, 0, strconv.FormatInt(user.Mid, 10), nil)
		}
	}
	if cfg.DisplayViewNum {
		if stat, ok := material.ActiveUsersRlys[cfg.ActiveUsersReqID]; ok {
			cd.ViewNum = stat64StringWithEmpty(stat.ViewCount)
			cd.DiscussNum = stat64StringWithEmpty(stat.DiscussCount)
		}
	}
	if cfg.DisplaySubscribeBtn {
		if tag, ok := material.Tags[cfg.TopicID]; ok {
			cd.IsSubscribed = tag.Attention == 1
		}
	}
	item := &api.ModuleItem{
		CardType:   model.CardTypeHeader.String(),
		CardId:     strconv.FormatInt(cfg.ModuleBase().ModuleID, 10),
		CardDetail: &api.ModuleItem_HeaderCard{HeaderCard: cd},
	}
	return []*api.ModuleItem{item}
}

func stat64StringWithEmpty(number int64) string {
	if number == 0 {
		return ""
	}
	return appcardmdl.Stat64String(number, "")
}
