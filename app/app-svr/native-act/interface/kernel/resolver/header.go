package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Header struct{}

func (r Header) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Header{
		BaseCfgManager:      config.NewBaseCfg(natModule),
		BgColor:             natModule.BgColor,
		SponsorContent:      natModule.Title,
		DisplayUser:         natModule.IsAttrDisplayUser() == natpagegrpc.AttrModuleYes,
		DisplayH5Header:     natModule.IsAttrDisplayH5Header() == natpagegrpc.AttrModuleYes,
		DisplayViewNum:      natModule.IsAttrIsCloseViewNum() != natpagegrpc.AttrModuleYes,
		DisplaySubscribeBtn: natModule.IsAttrIsCloseSubscribeBtn() != natpagegrpc.AttrModuleYes,
		TopicID:             natPage.ForeignID,
	}
	if cfg.DisplayUser && natPage.RelatedUid != 0 {
		cfg.SponsorMid = natPage.RelatedUid
		_, _ = cfg.AddMaterialParam(model.MaterialAccount, []int64{natPage.RelatedUid})
	}
	if natPage.IsTopicAct() {
		if cfg.DisplayViewNum {
			cfg.ActiveUsersReqID, _ = cfg.AddMaterialParam(model.MaterialActiveUsersRly, &model.ActiveUsersReq{
				TopicID: natPage.ForeignID,
				NoLimit: natPage.IsAttrDisplayCounty(),
			})
		}
		if cfg.DisplaySubscribeBtn && ss.Mid() > 0 {
			_, _ = cfg.AddMaterialParam(model.MaterialTag, []int64{cfg.TopicID})
		}
	}
	return cfg
}
