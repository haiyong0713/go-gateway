package resolver

import (
	"context"
	"strings"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Dynamic struct{}

func (r Dynamic) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Dynamic{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ImageTitle:     natModule.Meta,
		TextTitle:      natModule.Caption,
		IsFeed:         natModule.IsAttrLast() == natpagegrpc.AttrModuleYes,
		TopicID:        r.topicID(natModule, natPage),
		SortBy:         r.sortBy(natModule.DySort),
		IsMaster:       r.isMaster(natPage, ss),
	}
	r.setColor(cfg, natModule)
	r.setPageSize(cfg, natModule, ss)
	r.setContentSelect(cfg, module.Dynamic)
	r.setBaseCfg(cfg, ss)
	r.setOldVersion(cfg, module, natPage)
	return cfg
}

func (r Dynamic) sortBy(dySort int32) int32 {
	// 动态的综合即时间
	if dySort == model.DynSortCompre {
		return model.DynSortTime
	}
	return dySort
}

func (r Dynamic) isMaster(natPage *natpagegrpc.NativePage, ss *kernel.Session) bool {
	return natPage.IsUpTopicAct() && natPage.RelatedUid > 0 && natPage.RelatedUid == ss.Mid()
}

func (r Dynamic) topicID(natModule *natpagegrpc.NativeModule, natPage *natpagegrpc.NativePage) int64 {
	if natModule.Fid > 0 {
		return natModule.Fid
	}
	return natPage.ForeignID
}

func (r Dynamic) setColor(cfg *config.Dynamic, natModule *natpagegrpc.NativeModule) {
	colors := natModule.ColorsUnmarshal()
	cfg.BgColor = natModule.BgColor
	cfg.FontColor = colors.DisplayColor
}

func (r Dynamic) setPageSize(cfg *config.Dynamic, natModule *natpagegrpc.NativeModule, ss *kernel.Session) {
	pageSize := int32(natModule.Num)
	if ss.ReqFrom == model.ReqFromSubPage {
		pageSize = 10
	}
	cfg.PageSize = pageSize
}

func (r Dynamic) setContentSelect(cfg *config.Dynamic, dyn *natpagegrpc.Dynamic) {
	if dyn == nil || len(dyn.SelectList) == 0 {
		return
	}
	// 是否是单选选项
	isSingle := false
	defer func() {
		if !isSingle {
			return
		}
		cfg.Contents = nil
	}()
	for _, v := range dyn.SelectList {
		// 精选
		if v.ClassType == model.DynClassChoice {
			cfg.PickID = v.ClassID
			isSingle = true
			break
		}
		// 全部
		if v.SelectType <= 0 {
			isSingle = true
			break
		}
		cfg.Contents = append(cfg.Contents, v.SelectType)
	}
}

func (r Dynamic) setBaseCfg(cfg *config.Dynamic, ss *kernel.Session) {
	if cfg.IsFeed && model.IsFromIndex(ss.ReqFrom) {
		req := &dyntopicgrpc.HasDynsReq{
			TopicId: cfg.TopicID,
			PickId:  cfg.PickID,
			SortBy:  cfg.SortBy,
			Uid:     ss.Mid(),
			Types:   cfg.Contents,
		}
		cfg.HasDynsReqID, _ = cfg.AddMaterialParam(model.MaterialHasDynsRly, req)
		return
	}
	req := &dyntopicgrpc.ListDynsReq{
		TopicId:  cfg.TopicID,
		Uid:      ss.Mid(),
		SortBy:   cfg.SortBy,
		PageSize: cfg.PageSize,
		PickId:   cfg.PickID,
		WithTop:  true,
		Offset:   ss.FeedOffset,
		Types:    cfg.Contents,
		VersionCtrl: &dyncommongrpc.MetaDataCtrl{
			Platform:     ss.RawDevice().RawPlatform,
			Build:        ss.FormatInt(ss.RawDevice().Build),
			MobiApp:      ss.RawDevice().RawMobiApp,
			Buvid:        ss.RawDevice().Buvid,
			Device:       ss.RawDevice().Device,
			FromSpmid:    ss.FromSpmid,
			TraceId:      ss.TraceId(),
			TeenagerMode: ss.TeenagerMode(),
			Version:      ss.RawDevice().VersionName,
			Network:      int32(ss.RawNetwork().Type),
			Ip:           ss.Ip(),
		},
	}
	if ss.IsColdStart {
		req.VersionCtrl.ColdStart = 1
	}
	cfg.ListDynsReqID, _ = cfg.AddMaterialParam(model.MaterialListDynsRly, req)
}

func (r Dynamic) setOldVersion(cfg *config.Dynamic, module *natpagegrpc.Module, natPage *natpagegrpc.NativePage) {
	natModule := module.NativeModule
	cfg.ModuleTitle = natModule.Title
	cfg.Sort = r.oldSort(module)
	cfg.PageTitle = natPage.Title
	cfg.PageID = natPage.ID
	if natModule.Fid > 0 {
		_, _ = cfg.AddMaterialParam(model.MaterialTag, []int64{natModule.Fid})
	}
}

func (r Dynamic) oldSort(module *natpagegrpc.Module) string {
	if module.Dynamic == nil || len(module.Dynamic.SelectList) == 0 {
		return ""
	}
	var types string
	tys := make([]string, 0, len(module.Dynamic.SelectList))
	for _, v := range module.Dynamic.SelectList {
		// 精选或者全选时，是不支持多选的
		if tempType, isSingle := v.JoinMultiDyTypes(); isSingle {
			types = tempType
			tys = []string{}
			break
		} else {
			tys = append(tys, tempType)
		}
	}
	if len(tys) > 0 {
		types = strings.Join(tys, ",")
	}
	return types
}
