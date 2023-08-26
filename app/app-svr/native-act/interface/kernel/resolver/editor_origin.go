package resolver

import (
	"context"
	"strconv"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	appshowgrpc "git.bilibili.co/bapis/bapis-go/app/show/v1"
	hmtgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	actplatv2grpc "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/web-svr/native-page/interface/api"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type EditorOrigin struct{}

func (r EditorOrigin) Resolve(c context.Context, ss *kernel.Session, natPage *api.NativePage, module *api.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	confSort := natModule.ConfUnmarshal()
	cfg := &config.EditorOrigin{
		BaseCfgManager:    config.NewBaseCfg(natModule),
		Position:          buildEditorPosition(natModule),
		DisplayMoreButton: natModule.IsAttrDisplayOp() == natpagegrpc.AttrModuleYes,
		BgColor:           natModule.BgColor,
		RdbType:           confSort.RdbType,
		IsFeed:            natModule.IsAttrLast() == natpagegrpc.AttrModuleYes,
	}
	r.setPageSize(cfg, ss, natModule)
	r.setBaseCfg(cfg, confSort, ss, natModule)
	return cfg
}

func (r EditorOrigin) setBaseCfg(cfg *config.EditorOrigin, confSort *natpagegrpc.ConfSort, ss *kernel.Session, module *api.NativeModule) {
	switch cfg.RdbType {
	case model.RDBMustsee:
		if cfg.IsFeed && model.IsFromIndex(ss.ReqFrom) {
			return
		}
		cfg.PageArcsReqID, _ = cfg.AddMaterialParam(model.MaterialPageArcsRly, &populargrpc.PageArcsReq{
			Offset:   ss.Offset,
			PageSize: cfg.PageSize,
			ArcType:  confSort.MseeType,
		})
		if confSort.Sid != 0 && confSort.Counter != "" && ss.Mid() != 0 {
			cfg.GetHisReqID, _ = cfg.AddMaterialParam(model.MaterialGetHisRly, &actplatv2grpc.GetHistoryReq{
				Activity: strconv.FormatInt(confSort.Sid, 10),
				Counter:  confSort.Counter,
				Mid:      ss.Mid(),
			})
		}
	case model.RDBWeek:
		if module.Fid <= 0 {
			return
		}
		cfg.SelSerieReqID, _ = cfg.AddMaterialParam(model.MaterialSelSerieRly, &appshowgrpc.SelectedSerieReq{
			Type:   model.SelTypeWeek,
			Number: module.Fid,
		})
		if confSort.Sid != 0 && confSort.Counter != "" && ss.Mid() != 0 {
			cfg.GetHisReqID, _ = cfg.AddMaterialParam(model.MaterialGetHisRly, &actplatv2grpc.GetHistoryReq{
				Activity: strconv.FormatInt(confSort.Sid, 10),
				Counter:  confSort.Counter,
				Mid:      ss.Mid(),
			})
		}
	case model.RDBRank:
		if module.Fid <= 0 {
			return
		}
		cfg.MixExtReqID, _ = cfg.AddMaterialParam(model.MaterialMixExtRly, &natpagegrpc.ModuleMixExtReq{
			ModuleID: module.ID,
			Ps:       cfg.PageSize,
			Offset:   0, //排行榜无分页
			MType:    natpagegrpc.MixRankIcon,
		})
		cfg.RankRstReqID, _ = cfg.AddMaterialParam(model.MaterialRankRstRly, &kernel.RankResultReq{
			Req: &activitygrpc.RankResultReq{
				RankID: module.Fid,
				Pn:     1, //排行榜无分页
				Ps:     cfg.PageSize,
			},
		})
	case model.RDBGAT:
		if module.Fid <= 0 {
			return
		}
		cfg.ChannelFeedReqID, _ = cfg.AddMaterialParam(model.MaterialChannelFeedRly, &kernel.ChannelFeedReq{
			Req: &hmtgrpc.ChannelFeedReq{
				Cid:     module.Fid,
				Mid:     ss.Mid(),
				Buvid:   ss.Buvid(),
				Context: &hmtgrpc.ChannelContext{Ip: ss.Ip()},
				Offset:  int32(ss.Offset),
				Ps:      int32(cfg.PageSize),
			},
			NeedMultiML: true,
		})
	}
}

func (r EditorOrigin) setPageSize(cfg *config.EditorOrigin, ss *kernel.Session, module *api.NativeModule) {
	var ps int64 = 10
	if model.IsFromIndex(ss.ReqFrom) {
		ps = module.Num
	}
	cfg.PageSize = ps
}
