package dynamicV2

import (
	"context"
	"fmt"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	pgcInlineGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcDynGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/dynamic"
)

const (
	_extendInfocTopic   = "topic"
	_extendInfocHot     = "hot"
	_extendInfocGame    = "game"
	_extendInfocLBS     = "lbs"
	_extendInfocBiliCut = "diversion"
	_extendInfocBBQ     = "bbq"
	_extendInfocAutoOGV = "ogv"
)

// 填充动态顶部新话题卡模块
func (s *Service) extNewTopic(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.ResNewTopic == nil {
		return nil
	}
	// 6.45 以后才下发新话题模块
	if !s.isDynNewTopicView(c, general) {
		// 如果版本不够就直接屏蔽新话题卡
		dynCtx.ResNewTopic = nil
		return nil
	}
	// 如果是转发卡的原卡有话题，原卡本身不展示话题模块
	if dynCtx.From == _handleTypeForward && dynCtx.Dyn.Forward != nil {
		if _, ok := dynCtx.ResNewTopic[dynCtx.Dyn.Forward.DynamicID]; ok {
			return nil
		}
	}
	if t, ok := dynCtx.ResNewTopic[dynCtx.Dyn.DynamicID]; ok {
		dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, &api.Module{
			ModuleType: api.DynModuleType_module_topic,
			ModuleItem: &api.Module_ModuleTopic{
				ModuleTopic: &api.ModuleTopic{
					Id:   t.TopicID,
					Name: t.TopicName,
					Url:  t.JumpURL,
				},
			},
		})
	}
	return nil
}

/**
 * 填充所有展示的tag
 * 优先级：必剪tag > 地点tag > ogv自动匹配tag > 游戏tag > 轻视频tag > 话题tag > 热门tag > OGV标签
 * 互斥条件：有 游戏tag 则不填充 话题tag, 有 话题tag 则不填充 热门tag, 有ogv自动匹配tag则不填充热门tag
 */

// 本卡有话题附加卡时的特殊处理
// 1.不展示话题小卡
// 2.当本卡为转发卡时，其原卡不展示话题附加卡和话题小卡

// nolint:gocognit
func (s *Service) ext(ctx context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if dynCtx.Dyn.IsForward() && (dynCtx.Interim.ForwardOrigFaild || dynCtx.Interim.IsPassExtend) {
		return nil
	}
	extend := &api.ModuleExtend{}
	for _, v := range dynCtx.Dyn.Tags {
		// 基础信息
		var extInfoCommon *api.ExtInfoCommon
		// nolint:exhaustive
		switch v.TagType {
		case dyncommongrpc.TagType_TAG_LBS:
			// LBS
			ok, lbs := dynCtx.Dyn.GetLBS()
			if !ok {
				continue
			}
			extInfoCommon = s.extLBS(lbs)
			extInfoCommon.IsShowLight = true
			xmetric.DynamicExt.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "LBS")
		case dyncommongrpc.TagType_TAG_GAME, dyncommongrpc.TagType_TAG_GAME_SDK:
			// 游戏话题小卡
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() || mdlv2.FeatureStatusFromCtx(ctx).NoGameAttach.IsOn(ctx) {
				continue
			}
			if v.TagDetail == nil {
				continue
			}
			extInfoCommon = &api.ExtInfoCommon{
				Icon:        s.c.Resource.Icon.ModuleExtendGameTopic,
				Type:        api.DynExtendType_dyn_ext_type_common,
				SubModule:   _extendInfocGame,
				IsShowLight: true,
			}
			// 展示行动点信息
			if v.ActionPoint == 1 {
				if game, ok := dynCtx.ResGameAct[v.Rid]; ok {
					extInfoCommon.ActionText = game.GameButton
					extInfoCommon.ActionUrl = v.TagDetail.Link
					extInfoCommon.Rid = v.Rid
				}
			}
			xmetric.DynamicExt.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "游戏话题")
		case dyncommongrpc.TagType_TAG_GAME_CARD_CONVERT:
			// 游戏附加通卡
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() || mdlv2.FeatureStatusFromCtx(ctx).NoGameAttach.IsOn(ctx) {
				continue
			}
			game, ok := dynCtx.ResGame[v.Rid]
			if !ok {
				continue
			}
			extInfoCommon = &api.ExtInfoCommon{
				Icon:        s.c.Resource.Icon.ModuleExtendGameTopic,
				Type:        api.DynExtendType_dyn_ext_type_common,
				SubModule:   _extendInfocGame,
				Title:       game.GameName,
				Uri:         game.GameLink,
				IsShowLight: true,
			}
			// 展示行动点信息
			if v.ActionPoint == 1 {
				if gameAct, ok := dynCtx.ResGameAct[v.Rid]; ok {
					extInfoCommon.ActionText = gameAct.GameButton
					extInfoCommon.ActionUrl = game.GameLink
					extInfoCommon.Rid = v.Rid
				}
			}
			xmetric.DynamicExt.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "游戏通卡")
		case dyncommongrpc.TagType_TAG_TOPIC:
			// 话题小卡
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 新话题上线后完全不下发底部话题卡
			if s.isDynNewTopicView(ctx, general) {
				continue
			}
			extInfoCommon = s.extTopic(dynCtx.ResAdditionalTopic[v.Rid], general)
			xmetric.DynamicExt.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "话题小卡")
		case dyncommongrpc.TagType_TAG_HOT:
			// 热门小卡
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			extInfoCommon = &api.ExtInfoCommon{
				Title:     s.c.Resource.Text.ModuleExtendHotTitle,
				Uri:       s.c.Resource.Others.ModuleExtendHotURI,
				Icon:      s.c.Resource.Icon.ModuleExtendHot,
				Type:      api.DynExtendType_dyn_ext_type_hot,
				SubModule: _extendInfocHot,
			}
			xmetric.DynamicExt.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "热门小卡")
		case dyncommongrpc.TagType_TAG_DIVERSION:
			// 必减小卡
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			biliCut := dynCtx.GetBiliCut(v.Rid, s.c.Resource.Text.ModuleExtendBiliCutDefaultTitle, s.c.Resource.Others.ModuleExtendBiliCutDefaultURI)
			extInfoCommon = &api.ExtInfoCommon{
				Title:     biliCut.Name,
				Uri:       biliCut.AppUrl,
				Icon:      s.c.Resource.Icon.ModuleExtendBiliCut,
				Type:      api.DynExtendType_dyn_ext_type_biliCut,
				SubModule: _extendInfocBiliCut,
			}
			// 展示行动点信息
			if v.ActionPoint == 1 {
				extInfoCommon.Title = s.c.Resource.Text.ModuleExtendDuversionTitle
				extInfoCommon.ActionText = s.c.Resource.Text.ModuleExtendDuversionText
				extInfoCommon.ActionUrl = extInfoCommon.Uri
				extInfoCommon.Rid = v.Rid
			}
			xmetric.DynamicExt.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "必减小卡")
		case dyncommongrpc.TagType_TAG_OGV:
			//  OGV小卡tag
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			ep, ok := dynCtx.ResPGC[int32(v.Rid)]
			if !ok || v.TagDetail == nil {
				continue
			}
			tagm := map[string]*pgcInlineGrpc.Tag{}
			if ep.DynamicMeta != nil {
				for _, tag := range ep.DynamicMeta.Tags {
					tagm[tag.Name] = tag
				}
			}
			tag, ok := tagm[v.TagDetail.Text]
			if !ok {
				continue
			}
			extInfoCommon = &api.ExtInfoCommon{
				Title:       tag.Name,
				Uri:         tag.Link,
				Icon:        tag.Icon,
				Type:        api.DynExtendType_dyn_ext_type_ogv,
				SubModule:   tag.Report.SubModule,
				IsShowLight: true,
			}
			xmetric.DynamicExt.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "OGV小卡")
		case dyncommongrpc.TagType_TAG_BBQ:
			// 轻视频
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			extInfoCommon = &api.ExtInfoCommon{
				Title:     s.c.Resource.Text.ModuleExtendBBQTitle,
				Uri:       s.c.Resource.Others.ModuleExtendBBQURI,
				Icon:      s.c.Resource.Icon.ModuleExtendBBQ,
				Type:      api.DynExtendType_dyn_ext_type_common,
				SubModule: _extendInfocBBQ,
			}
			xmetric.DynamicExt.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "轻视频小卡")
		case dyncommongrpc.TagType_TAG_AUTOOGV:
			// 自动挂卡
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			autoPGC, ok := dynCtx.ResAdditionalOGV[v.Rid]
			if !ok {
				continue
			}
			extInfoCommon = s.extAutoOGV(autoPGC)
			// 展示行动点信息
			if v.ActionPoint == 1 {
				extInfoCommon.ActionUrl = autoPGC.Link
				extInfoCommon.ActionText = "看正片"
				extInfoCommon.Rid = v.Rid
			}
		}
		if extInfoCommon == nil {
			xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "ext", "date_faild")
			log.Warn("module error mid(%v) dynid(%v) ext ext_type %v", general.Mid, dynCtx.Dyn.DynamicID, v.TagType)
			continue
		}
		extend.Extend = s.appendExt(v.GetTagDetail(), extend.Extend, extInfoCommon)
		xmetric.DynamicExt.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "OGV自动挂卡")
	}
	// module赋值
	if len(extend.Extend) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_extend,
		ModuleItem: &api.Module_ModuleExtend{
			ModuleExtend: extend,
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

// nolint:gocognit
func (s *Service) extFake(_ context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if dynCtx.Dyn.IsForward() && (dynCtx.Interim.ForwardOrigFaild || dynCtx.Interim.IsPassExtend) {
		return nil
	}
	extend := &api.ModuleExtend{}
	for _, v := range dynCtx.Dyn.Tags {
		// 基础信息
		var extInfoCommon *api.ExtInfoCommon
		// nolint:exhaustive
		switch v.TagType {
		case dyncommongrpc.TagType_TAG_LBS:
			// LBS
			ok, lbs := dynCtx.Dyn.GetLBS()
			if !ok {
				continue
			}
			extInfoCommon = s.extLBS(lbs)
		case dyncommongrpc.TagType_TAG_GAME:
			// 游戏话题小卡
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() || v.TagDetail == nil {
				continue
			}
			gameTopic, ok := dynCtx.Dyn.GetTopicInfo()
			if !ok {
				continue
			}
			extInfoCommon = s.extGameTopic(gameTopic, s.c.BottomConfig.TopicJumpLinks)
		case dyncommongrpc.TagType_TAG_TOPIC:
			// 话题小卡
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			extInfoCommon = s.extTopic(dynCtx.ResAdditionalTopic[v.Rid], general)
		case dyncommongrpc.TagType_TAG_HOT:
			// 热门小卡
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			extInfoCommon = &api.ExtInfoCommon{
				Title:     s.c.Resource.Text.ModuleExtendHotTitle,
				Uri:       s.c.Resource.Others.ModuleExtendHotURI,
				Icon:      s.c.Resource.Icon.ModuleExtendHot,
				Type:      api.DynExtendType_dyn_ext_type_hot,
				SubModule: _extendInfocHot,
			}
		case dyncommongrpc.TagType_TAG_DIVERSION:
			// 必减小卡
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			biliCut := dynCtx.GetBiliCut(v.Rid, s.c.Resource.Text.ModuleExtendBiliCutDefaultTitle, s.c.Resource.Others.ModuleExtendBiliCutDefaultURI)
			extInfoCommon = &api.ExtInfoCommon{
				Title:     biliCut.Name,
				Uri:       biliCut.AppUrl,
				Icon:      s.c.Resource.Icon.ModuleExtendBiliCut,
				Type:      api.DynExtendType_dyn_ext_type_biliCut,
				SubModule: _extendInfocBiliCut,
			}
		case dyncommongrpc.TagType_TAG_OGV:
			//  OGV小卡tag
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			ep, ok := dynCtx.ResPGC[int32(v.Rid)]
			if !ok {
				continue
			}
			if ep.DynamicMeta == nil {
				continue
			}
			for _, tag := range ep.DynamicMeta.Tags {
				ogvExt := s.extOGV(tag)
				if ogvExt == nil {
					continue
				}
				extend.Extend = s.appendExtFake(extend.Extend, ogvExt)
			}
			continue
		case dyncommongrpc.TagType_TAG_BBQ:
			// 轻视频
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			extInfoCommon = &api.ExtInfoCommon{
				Title:     s.c.Resource.Text.ModuleExtendBBQTitle,
				Uri:       s.c.Resource.Others.ModuleExtendBBQURI,
				Icon:      s.c.Resource.Icon.ModuleExtendBBQ,
				Type:      api.DynExtendType_dyn_ext_type_common,
				SubModule: _extendInfocBBQ,
			}
		case dyncommongrpc.TagType_TAG_AUTOOGV:
			// 自动挂卡
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			autoPGC, ok := dynCtx.ResAdditionalOGV[v.Rid]
			if !ok {
				continue
			}
			extInfoCommon = s.extAutoOGV(autoPGC)
		}
		if extInfoCommon == nil {
			continue
		}
		extend.Extend = s.appendExtFake(extend.Extend, extInfoCommon)
	}
	// module赋值
	if len(extend.Extend) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_extend,
		ModuleItem: &api.Module_ModuleExtend{
			ModuleExtend: extend,
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) extLBS(lbs *mdlv2.Lbs) *api.ExtInfoCommon {
	extInfo := &api.ExtInfoCommon{
		Title:     lbs.ShowTitle,
		Uri:       fmt.Sprintf(model.LBSURI, lbs.Poi, lbs.Type, lbs.Location.Lat, lbs.Location.Lng, lbs.Title, lbs.Address),
		Icon:      s.c.Resource.Icon.ModuleExtendLBS,
		PoiType:   int32(lbs.Type),
		Type:      api.DynExtendType_dyn_ext_type_lbs,
		SubModule: _extendInfocLBS,
	}
	return extInfo
}

func (s *Service) extAutoOGV(tag *pgcDynGrpc.FollowCardProto) *api.ExtInfoCommon {
	extInfo := &api.ExtInfoCommon{
		Title:     tag.Title,
		Uri:       tag.Link,
		Icon:      s.c.Resource.Icon.ModuleExtendAutoOGV,
		Type:      api.DynExtendType_dyn_ext_type_auto_ogv,
		SubModule: _extendInfocAutoOGV,
	}
	return extInfo
}

func (s *Service) extGameTopic(topics []*mdlv2.Topic, bottomConfig []conf.BottomItem) *api.ExtInfoCommon {
	for _, topic := range topics {
		if topic == nil {
			continue
		}
		for _, bottoms := range bottomConfig {
			for _, gameTopic := range bottoms.RelatedTopic {
				topicName := fmt.Sprintf("#%v#", topic.TopicName)
				if topicName == gameTopic {
					extInfo := &api.ExtInfoCommon{
						Title:     bottoms.Display,
						Uri:       bottoms.URL,
						Icon:      s.c.Resource.Icon.ModuleExtendGameTopic,
						Type:      api.DynExtendType_dyn_ext_type_common,
						SubModule: _extendInfocGame,
					}
					return extInfo
				}
			}
		}
	}
	return nil
}

func (s *Service) extTopic(topics []*mdlv2.Topic, general *mdlv2.GeneralParam) *api.ExtInfoCommon {
	var title, uri string
	for _, topic := range topics {
		if topic.Stat != 1 { // 未绑定
			continue
		}
		title = topic.TopicName
		uri = topic.TopicLink
		if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
			break
		}
	}
	if title != "" && uri != "" {
		extInfo := &api.ExtInfoCommon{
			Title:     title,
			Uri:       uri,
			Icon:      s.c.Resource.Icon.ModuleExtendTopic,
			Type:      api.DynExtendType_dyn_ext_type_topic,
			SubModule: _extendInfocTopic,
		}
		return extInfo
	}
	return nil
}

func (s *Service) extOGV(tag *pgcInlineGrpc.Tag) *api.ExtInfoCommon {
	if tag.Report == nil {
		return nil
	}
	return &api.ExtInfoCommon{
		Title:     tag.Name,
		Uri:       tag.Link,
		Icon:      tag.Icon,
		Type:      api.DynExtendType_dyn_ext_type_ogv,
		SubModule: tag.Report.SubModule,
	}
}

func (s *Service) appendExt(tag *dyncommongrpc.TagDeatil, exts []*api.ModuleExtendItem, extinfo *api.ExtInfoCommon) []*api.ModuleExtendItem {
	// 如果动态tag信息为不空则优先用
	if tag != nil {
		extinfo.Title = tag.Text
		extinfo.Uri = tag.Link
		extinfo.Icon = tag.Icon
		// 如果外层的行动点url不为空则表示当前已经是需要展示行动点，url和卡片url统一
		if extinfo.ActionUrl != "" {
			extinfo.ActionUrl = extinfo.Uri
		}
	}
	ext := &api.ModuleExtendItem{
		Type: api.DynExtendType_dyn_ext_type_common,
		Extend: &api.ModuleExtendItem_ExtInfoCommon{
			ExtInfoCommon: extinfo,
		},
	}
	var items = append(exts, ext)
	return items
}

func (s *Service) appendExtFake(exts []*api.ModuleExtendItem, extinfo *api.ExtInfoCommon) []*api.ModuleExtendItem {
	ext := &api.ModuleExtendItem{
		Type: api.DynExtendType_dyn_ext_type_common,
		Extend: &api.ModuleExtendItem_ExtInfoCommon{
			ExtInfoCommon: extinfo,
		},
	}
	var items = append(exts, ext)
	return items
}
