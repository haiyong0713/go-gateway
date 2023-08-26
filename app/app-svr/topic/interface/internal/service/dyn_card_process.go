package service

import (
	"go-common/library/log"

	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	"go-gateway/app/app-svr/topic/card/proto/dyn_handler"
	api "go-gateway/app/app-svr/topic/interface/api"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"

	"github.com/pkg/errors"
)

// dynCardProcess: 动态卡片构造流程,此处与app-dynamic构造流程保持一致
func (s *Service) dynCardProcess(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam, params []*topiccardmodel.DynMetaCardListParam) ([]*api.TopicCardItem, error) {
	// Step 1. 获取dynamic_list
	dynList, err := s.dynBriefs(dynSchemaCtx, general, params)
	if err != nil {
		log.Error("s.dynBriefs params=%+v, error=%+v", params, err)
		return nil, err
	}
	// Step 2. 初始化返回值 & 获取物料信息
	dynSchemaCtx, err = s.getMaterial(dynSchemaCtx, general, dynList)
	if err != nil {
		log.Error("s.getMaterial dynList.Dynamics=%+v, error=%+v", dynList.Dynamics, err)
		return nil, err
	}
	// Step 3. 对物料信息处理，获取详情列表
	schema := &dynHandler.CardSchema{}
	dynRawList := schema.ProcListReply(dynSchemaCtx, dynList.Dynamics, general, "")
	// Step 4. 回填
	s.procBackfill(dynSchemaCtx, general, dynRawList, schema)
	// Step 5. 直接按顺序取列表
	var res []*api.TopicCardItem
	for _, item := range dynRawList.List {
		res = append(res, &api.TopicCardItem{
			Type:        api.TopicCardType_DYNAMIC,
			DynamicItem: item.Item,
		})
	}
	return res, nil
}

// dynWebCardProcess: 动态web卡片构造流程，依赖dynmdlV2.DynamicContext结构物料通过新构造器构造json卡片
func (s *Service) dynWebCardProcess(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam, params []*topiccardmodel.DynMetaCardListParam) ([]*jsonwebcard.TopicCard, error) {
	// Step 1. 获取dynamic_list
	dynList, err := s.dynBriefs(dynSchemaCtx, general, params)
	if err != nil {
		return nil, errors.Wrapf(err, "s.dynBriefs params=%+v", params)
	}
	// Step 2. 初始化返回值 & 获取物料信息
	dynSchemaCtx, err = s.getMaterial(dynSchemaCtx, general, dynList)
	if err != nil {
		return nil, errors.Wrapf(err, "s.getMaterial dynList.Dynamics=%+v", dynList.Dynamics)
	}
	// Step 3. 卡片构造抽象化，内部逻辑模块化
	var output []*jsonwebcard.TopicCard
	for _, item := range dynList.Dynamics {
		dynSchemaCtx.DynCtx.Dyn = item                                                          // 原始数据
		dynSchemaCtx.DynCtx.DynamicItem = &dynamicapi.DynamicItem{Extend: &dynamicapi.Extend{}} // 聚合结果
		dynSchemaCtx.DynCtx.Interim = &dynmdlV2.Interim{}                                       // 临时逻辑
		builder, ok := webDynCardGetBuilder(dynSchemaCtx.DynCtx.Dyn)
		if !ok {
			log.Error("Unsupported Dyn Type=%+v", dynSchemaCtx.DynCtx.Dyn.Type)
			continue
		}
		cardOutput, err := builder.Build(dynSchemaCtx, general)
		if err != nil {
			log.Error("Failed to build card output: %+v", err)
			continue
		}
		// Step 4. 回填
		s.backfillGetMaterial(dynSchemaCtx, general)
		fillCardOutput := builder.BackFill(cardOutput, dynSchemaCtx.DynCtx)
		output = append(output, fillCardOutput)
	}
	return output, nil
}

func convertToDynBriefsParams(params []*topiccardmodel.DynMetaCardListParam, general *topiccardmodel.GeneralParam) *dyngrpc.DynBriefsReq {
	var dynIds []int64
	for _, v := range params {
		dynIds = append(dynIds, v.DynId)
	}
	return &dyngrpc.DynBriefsReq{
		Uid:    general.Mid,
		DynIds: dynIds,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:    general.GetBuildStr(),
			Platform: general.GetPlatform(),
			MobiApp:  general.GetMobiApp(),
			Buvid:    general.GetBuvid(),
			Device:   general.GetDevice(),
			Ip:       general.IP,
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
	}
}
