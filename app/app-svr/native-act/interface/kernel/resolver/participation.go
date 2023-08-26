package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/web-svr/native-page/interface/api"
)

type Participation struct{}

func (r Participation) Resolve(c context.Context, ss *kernel.Session, natPage *api.NativePage, module *api.Module) config.BaseCfgManager {
	if module.Participation == nil || len(module.Participation.List) == 0 {
		return nil
	}
	cfg := &config.Participation{BaseCfgManager: config.NewBaseCfg(module.NativeModule)}
	var sids []int64
	for _, v := range module.Participation.List {
		// 专栏不支持带活动信息暂时略过
		if v.IsPartVideo() && v.ForeignID > 0 {
			sids = append(sids, v.ForeignID)
		}
		item := &config.ParticipationItem{
			Type:          int64(v.MType),
			Sid:           v.ForeignID,
			ButtonContent: v.Title,
			UploadType:    int64(v.UpType),
		}
		if ext, err := model.UnmarshalParticipationExt(v.Ext); err == nil {
			item.NewTid = ext.NewTid
		}
		cfg.Items = append(cfg.Items, item)
	}
	_, _ = cfg.BaseCfgManager.AddMaterialParam(model.MaterialActSubProto, sids)
	return cfg
}
