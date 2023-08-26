package dynamicV2

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	bcgmdl "go-gateway/app/app-svr/app-dynamic/interface/model/bcg"
	cheesemdl "go-gateway/app/app-svr/app-dynamic/interface/model/cheese"
	comicmdl "go-gateway/app/app-svr/app-dynamic/interface/model/comic"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	gamemdl "go-gateway/app/app-svr/app-dynamic/interface/model/game"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	"go-gateway/app/app-svr/app-dynamic/interface/model/shopping"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	"go-gateway/pkg/idsafe/bvid"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyntopicextgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic-ext"
	garbmdl "git.bilibili.co/bapis/bapis-go/garb/model"
	dramaseasongrpc "git.bilibili.co/bapis/bapis-go/maoer/drama/dramaseason"
	natpagegrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	esportGrpc "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	pgcDynGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/dynamic"
)

/**
 * 填充所有展示的附加卡，目前只填充一张卡
 * 优先级：通用附加卡（帮推) > 商品卡(视频页不出) > ugc卡 > 通用附加卡 (普通活动 > 话题活动 > ogv > 游戏 > 赛事 > 漫画 > 装扮 > 课程）> 投票附加卡(视频页不出，可以和ugc同时下发)
 */

// nolint:deadcode,varcheck
const (
	_voteWait   = 0 // 待审
	_voteOK     = 1 // 正常
	_voteDel    = 2 // 删除
	_voteRefuse = 3 // 未过审
	_voteDead   = 4 // 失效

	_voteTypeWord = 0 // 文字类型
	_voteTypePic  = 1 // 图片类型

	garbStateReserve = 1
	garbStateWait    = 2
	garbStateSell    = 3
	garbStateSellOut = 4
	garbStateOff     = 5

	goodsTypeTaoBao   = 1
	goodsTypeBilibili = 2
	goodsTypeJD       = 3 // 京东

	// UP预约状态
	_upNotStart = 0 // 未开始
	_upStart    = 1 // 预约中
	_upOnline   = 2 // 已上线
	_upDelete   = 3 // 删除
	_upCancel   = 4 // 取消
	_upEnd      = 5 // 结束
	_upAudit    = 6 // 先审后发
	_upExpired  = 7 // 已过期
	// 按钮状态
	_upbuttonReservation       = 0 // 预约
	_upbuttonReservationOk     = 1 // 已预约
	_upbuttonCancel            = 2 // 取消预约
	_upbuttonCancelOk          = 3 // 已取消
	_upbuttonWatch             = 4 // 去观看
	_upbuttonReplay            = 5 // 回放
	_upbuttonEnd               = 6 // 已结束
	_upbuttonCancelLotteryCron = 7 // 取消预约抽奖
)

// nolint:gocognit
func (s *Service) additional(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if dynCtx.Dyn.IsForward() && (dynCtx.Interim.ForwardOrigFaild || dynCtx.Interim.IsPassAddition) {
		return nil
	}
	for _, v := range dynCtx.Dyn.AttachCardInfos {
		var (
			common *api.ModuleAdditional
		)
		switch v.CardType {
		case dyncommongrpc.AttachCardType_ATTACH_CARD_GOODS:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 商品
			if resGood, ok := dynCtx.ResGood[dynCtx.Dyn.DynamicID]; ok {
				if res, ok := resGood[bcgmdl.GoodsLocTypeCard]; ok {
					common = s.additionalGood(res, dynCtx, general)
					xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "商品卡")
				}
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_MAN_TIAN_XING:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 满天星
			if res, ok := dynCtx.ResManTianXinm[v.Rid]; ok {
				common = s.additionalManTianXin(res, dynCtx, general)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_MEMBER_GOODS:
			// 会员购
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 商品
			if res, ok := dynCtx.ShoppingItems[v.Rid]; ok {
				common = s.additionalShopping(res, dynCtx, general)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_VOTE:
			// 投票
			if res, ok := dynCtx.ResVote[v.Rid]; ok {
				common = s.additionalVote(res, dynCtx)
				xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "投票")
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_UGC:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			if additionUGC, ok := dynCtx.GetArchive(v.Rid); ok {
				common = s.additionalUGC(additionUGC, dynCtx, general)
				xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "UGC卡")
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_ACTIVITY:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 帮推
			if topicID, ok := dynCtx.ResAttachedPromo[dynCtx.Dyn.DynamicID]; ok {
				if res, ok := dynCtx.ResActivity[topicID]; ok {
					common = s.additionalAttachedPromo(res, dynCtx)
					xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "帮推卡")
				}
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_OFFICIAL_ACTIVITY:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 普通活动
			if additionAct, ok := dynCtx.ResActivityRelation[v.Rid]; ok {
				if natPage, ok := dynCtx.ResNativePage[additionAct.NativeID]; ok {
					common = s.additionalNatPage(dynCtx, additionAct, natPage, general)
					xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "普通活动卡")
				}
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_UP_ACTIVITY:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// UP发布的活动
			natPage, ok := dynCtx.NativeAllPageCards[v.Rid]
			if !ok {
				continue
			}
			common = s.additionUpActivity(dynCtx, natPage, general)
			xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "UP主发起活动")
		case dyncommongrpc.AttachCardType_ATTACH_CARD_TOPIC, dyncommongrpc.AttachCardType_ATTACH_CARD_UP_TOPIC:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 话题活动
			if additionTopic, ok := dynCtx.ResAdditionalTopic[v.Rid]; ok {
				common = s.additionalTopic(additionTopic, dynCtx, dynCtx.ResTopicAdditiveCard, general)
				xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "话题活动卡")
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_OGV, dyncommongrpc.AttachCardType_ATTACH_CARD_AUTO_OGV:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// OGV
			if additionOGV, ok := dynCtx.GetResAdditionalOGV(v.Rid); ok {
				common = s.additionalOGV(additionOGV, dynCtx, general)
				xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "OGV卡")
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_MATCH:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 赛事
			if res, ok := dynCtx.ResMatch[v.Rid]; ok {
				common = s.additionalMatch(res, dynCtx)
				xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "赛事卡")
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_MANGA:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 漫画
			if res, ok := dynCtx.ResManga[v.Rid]; ok {
				common = s.additionalManga(res, dynCtx, general)
				xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "漫画卡")
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_PUGV:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 课程
			if res, ok := dynCtx.ResPUgv[v.Rid]; ok {
				common = s.additionalPugv(res, dynCtx, general)
				xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "课程卡")
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_GAME:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() || mdlv2.FeatureStatusFromCtx(c).NoGameAttach.IsOn(c) {
				continue
			}
			// 游戏
			if res, ok := dynCtx.ResGame[v.Rid]; ok {
				common = s.additionalGame(res, dynCtx)
				xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "游戏卡")
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_DECORATION:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 装扮
			if res, ok := dynCtx.ResDecorate[v.Rid]; ok {
				common = s.additionalDecorate(res, dynCtx)
				xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "装扮卡")
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:
			if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynReservePad, &feature.OriginResutl{
				BuildLimit: (general.IsPad() && general.GetBuild() < s.c.BuildLimit.DynReservePadIOSPad) ||
					(general.IsPadHD() && general.GetBuild() < s.c.BuildLimit.DynReservePadHD) ||
					(general.IsAndroidHD() && general.GetBuild() <= s.c.BuildLimit.DynReservePadAndroid)}) {
				continue
			}
			if res, ok := dynCtx.ResUpActRelationInfo[v.Rid]; ok {
				var toast string
				common, ok, toast = s.additionalUP(c, res, dynCtx.ResUpActReserveDove[v.Rid], dynCtx, general)
				// 预约卡删除了，展示失效卡
				if !ok {
					module := s.additionalNull(toast)
					dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
					continue
				}
				xmetric.DynamicAdditional.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "预约卡")
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_UP_MAOER:
			// 猫儿
			if general.IsPad() || general.IsPadHD() || general.IsAndroidHD() {
				return nil
			}
			if res, ok := dynCtx.ResFeedCardDramaInfo[v.Rid]; ok {
				common = s.additionalFeedCardDrama(res, general, dynCtx)
			}
		default:
			xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "additional", "unkown_type")
			log.Warn("module error mid(%v) dynid(%v) additional unkown_type %v", general.Mid, dynCtx.Dyn.DynamicID, v.CardType)
			continue
		}
		if common == nil {
			xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "additional", "date_faild")
			log.Warn("module error mid(%v) dynid(%v) additional addition_type %v", general.Mid, dynCtx.Dyn.DynamicID, v.CardType)
			continue
		}
		common.Rid = v.Rid
		common.NeedWriteCalender = v.NeedWriteCalender
		module := &api.Module{
			ModuleType: api.DynModuleType_module_additional,
			ModuleItem: &api.Module_ModuleAdditional{
				ModuleAdditional: common,
			},
		}
		dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	}
	return nil
}

// nolint:gocognit
func (s *Service) additionalFake(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if dynCtx.Dyn.IsForward() && (dynCtx.Interim.ForwardOrigFaild || dynCtx.Interim.IsPassAddition) {
		return nil
	}
	for _, v := range dynCtx.Dyn.AttachCardInfos {
		var (
			common *api.ModuleAdditional
		)
		switch v.CardType {
		case dyncommongrpc.AttachCardType_ATTACH_CARD_GOODS:
			// 商品
			if resGood, ok := dynCtx.ResGood[dynCtx.Dyn.DynamicID]; ok {
				if res, ok := resGood[bcgmdl.GoodsLocTypeCard]; ok {
					common = s.additionalGood(res, dynCtx, general)
				}
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_VOTE:
			// 投票
			if res, ok := dynCtx.ResVote[v.Rid]; ok {
				common = s.additionalVote(res, dynCtx)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_UGC:
			if additionUGC, ok := dynCtx.GetArchive(v.Rid); ok {
				common = s.additionalUGC(additionUGC, dynCtx, general)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_ACTIVITY:
			// 帮推
			if topicID, ok := dynCtx.ResAttachedPromo[dynCtx.Dyn.DynamicID]; ok {
				if res, ok := dynCtx.ResActivity[topicID]; ok {
					common = s.additionalAttachedPromo(res, dynCtx)
				}
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_OFFICIAL_ACTIVITY:
			// 普通活动
			if additionAct, ok := dynCtx.ResActivityRelation[v.Rid]; ok {
				if natPage, ok := dynCtx.ResNativePage[additionAct.NativeID]; ok {
					common = s.additionalNatPage(dynCtx, additionAct, natPage, general)
				}
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_TOPIC:
			// 话题活动
			if additionTopic, ok := dynCtx.ResAdditionalTopic[v.Rid]; ok {
				common = s.additionalTopic(additionTopic, dynCtx, dynCtx.ResTopicAdditiveCard, general)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_OGV, dyncommongrpc.AttachCardType_ATTACH_CARD_AUTO_OGV:
			// OGV
			if additionOGV, ok := dynCtx.GetResAdditionalOGV(v.Rid); ok {
				common = s.additionalOGV(additionOGV, dynCtx, general)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_MATCH:
			// 赛事
			if res, ok := dynCtx.ResMatch[v.Rid]; ok {
				common = s.additionalMatch(res, dynCtx)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_MANGA:
			// 漫画
			if res, ok := dynCtx.ResManga[v.Rid]; ok {
				common = s.additionalManga(res, dynCtx, general)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_PUGV:
			// 课程
			if res, ok := dynCtx.ResPUgv[v.Rid]; ok {
				common = s.additionalPugv(res, dynCtx, general)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_GAME:
			// 游戏
			if res, ok := dynCtx.ResGame[v.Rid]; ok {
				common = s.additionalGame(res, dynCtx)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_DECORATION:
			// 装扮
			if res, ok := dynCtx.ResDecorate[v.Rid]; ok {
				common = s.additionalDecorate(res, dynCtx)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:
			if res, ok := dynCtx.ResUpActRelationInfo[v.Rid]; ok {
				common, ok = s.additionalUPFake(c, res, dynCtx, general)
				// 预约卡删除了，展示失效卡
				if !ok {
					module := s.additionalNull("原预约信息已删除")
					dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
					continue
				}
			}
		default:
			continue
		}
		if common == nil {
			continue
		}
		common.Rid = v.Rid
		module := &api.Module{
			ModuleType: api.DynModuleType_module_additional,
			ModuleItem: &api.Module_ModuleAdditional{
				ModuleAdditional: common,
			},
		}
		dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	}
	return nil
}

func (s *Service) additionalUGC(ap *archivegrpc.ArcPlayer, _ *mdlv2.DynamicContext, general *mdlv2.GeneralParam) *api.ModuleAdditional {
	var arc = ap.Arc
	ugc := &api.AdditionUgc{
		HeadText:   s.c.Resource.Text.ModuleAdditionalUgcHeadText,
		Title:      arc.Title,
		Cover:      arc.Pic,
		DescText_1: "", // 客户端需要支持，当前版本服务端不下发
		DescText_2: fmt.Sprintf("%s观看 %s弹幕", s.numTransfer(int(arc.Stat.View)), s.numTransfer(int(arc.Stat.Danmaku))),
		LineFeed:   true,
		Duration:   s.videoDuration(arc.Duration),
		CardType:   "ugc",
	}
	if general.IsIPhonePick() && general.GetBuild() >= 62000200 || general.IsAndroidPick() && general.GetBuild() >= 6200000 {
		// 6.20之后版本可以不下发HeadText，老版本端上会展示一块空白
		ugc.HeadText = s.c.Resource.Text.ModuleAdditionalUgcHeadTextV2
	}
	ugc.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(arc.Aid, 10), model.AvPlayHandlerGRPCV2(ap, arc.FirstCid, true))
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_ugc,
		Item: &api.ModuleAdditional_Ugc{
			Ugc: ugc,
		},
	}
	return additional
}

func (s *Service) additionalGood(res map[string]*bcgmdl.GoodsItem, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) *api.ModuleAdditional {
	good := &api.AdditionGoods{
		CardType:   "good",
		Icon:       s.c.Resource.Icon.ModuleAdditionalGoods,
		Uri:        dynCtx.DynamicItem.Extend.CardUrl,
		AdMarkIcon: s.c.Resource.Icon.AdditionalAdMarkIcon,
	}
	for _, id := range strings.Split(dynCtx.Dyn.Extend.OpenGoods.ItemsId, ",") {
		goodsItem, ok := res[id]
		if !ok {
			continue
		}
		if (general.IsIPhonePick() && general.GetBuild() < 66000000 || general.IsAndroidPick() && general.GetBuild() < 6600000) && (goodsItem.SourceType != 1 && goodsItem.SourceType != 2) {
			continue
		}
		tmp := &api.GoodsItem{
			Cover:             goodsItem.Img,
			SchemaPackageName: goodsItem.SchemaPackageName,
			SourceType:        int32(goodsItem.SourceType),
			JumpUrl:           goodsItem.JumpLink,
			JumpDesc:          goodsItem.JumpLinkDesc,
			Title:             goodsItem.Name,
			Brief:             goodsItem.Brief,
			Price:             goodsItem.PriceStr,
			ItemId:            goodsItem.ItemsID,
			SchemaUrl:         goodsItem.SchemaURL,
			OpenWhiteList:     goodsItem.OpenWhiteList,
			UserWebV2:         goodsItem.UserAdWebV2,
			AdMark:            goodsItem.AdMark,
			AppName:           goodsItem.AppName,
			JumpType:          api.GoodsJumpType_goods_schema,
		}
		if goodsItem.OuterApp == 0 {
			tmp.JumpType = api.GoodsJumpType_goods_url
		}
		good.RcmdDesc = goodsItem.AdMark
		good.GoodsItems = append(good.GoodsItems, tmp)
	}
	if len(good.GoodsItems) > 0 {
		good.SourceType = good.GoodsItems[0].SourceType
		good.JumpType = good.GoodsItems[0].JumpType
		good.AppName = good.GoodsItems[0].AppName
		additional := &api.ModuleAdditional{
			Type: api.AdditionalType_additional_type_goods,
			Item: &api.ModuleAdditional_Goods{
				Goods: good,
			},
		}
		return additional
	}
	return nil
}

func (s *Service) additionalManTianXin(res *shopping.CardInfo, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) *api.ModuleAdditional {
	good := &api.AdditionGoods{
		CardType:   "good",
		Icon:       s.c.Resource.Icon.ModuleAdditionalGoods,
		Uri:        dynCtx.DynamicItem.Extend.CardUrl,
		AdMarkIcon: s.c.Resource.Icon.AdditionalAdMarkIcon,
		RcmdDesc:   "UP主的推荐",
	}
	tmp := &api.GoodsItem{
		Cover:      res.Img,
		SourceType: 2,
		JumpUrl:    res.JumpLink,
		JumpDesc:   res.JumpLinkDesc,
		Title:      res.Name,
		Price:      res.PriceStr,
		ItemId:     res.ItemsId,
		AdMark:     res.AdMark,
		JumpType:   api.GoodsJumpType_goods_url,
	}
	good.GoodsItems = append(good.GoodsItems, tmp)
	if len(good.GoodsItems) > 0 {
		good.SourceType = tmp.SourceType
		good.JumpType = tmp.JumpType
		good.AppName = tmp.AppName
		additional := &api.ModuleAdditional{
			Type: api.AdditionalType_additional_type_goods,
			Item: &api.ModuleAdditional_Goods{
				Goods: good,
			},
		}
		return additional
	}
	return nil
}

func (s *Service) additionalAttachedPromo(act *natpagegrpc.NativePage, dynCtx *mdlv2.DynamicContext) *api.ModuleAdditional {
	common := &api.AdditionCommon{
		HeadText:   s.c.Resource.Text.ModuleAdditionalAttachedPromoHeadText,
		Title:      act.ShareCaption,
		ImageUrl:   act.ShareImage,
		DescText_1: act.ShareTitle,
		Url:        act.SkipURL,
		Style:      api.ImageStyle_add_style_vertical,
		Type:       strconv.FormatInt(dynCtx.Dyn.Type, 10),
		CardType:   "activity",
	}
	if common.Title == "" {
		common.Title = act.Title
	}
	if common.Url == "" {
		common.Url = model.FillURI(model.GotoActivity, strconv.FormatInt(act.ID, 10), nil)
	}
	common.Button = &api.AdditionalButton{
		Type:    api.AddButtonType_bt_jump,
		JumpUrl: model.FillURI(model.GotoActivity, strconv.FormatInt(act.ID, 10), nil),
		JumpStyle: &api.AdditionalButtonStyle{
			Text: s.c.Resource.Text.ModuleAdditionalAttachedPromoButton,
		},
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_common,
		Item: &api.ModuleAdditional_Common{
			Common: common,
		},
	}
	return additional
}

func (s *Service) additionalNatPage(dynCtx *mdlv2.DynamicContext, act *activitygrpc.ActRelationInfoReply, natPage *natpagegrpc.NativePageCard, general *mdlv2.GeneralParam) *api.ModuleAdditional {
	const (
		_actNoStart = 0
		_actStart   = 1
		_actOff     = 2
		_reserve    = 1
	)
	common := &api.AdditionCommon{
		HeadText:   s.c.Resource.Text.ModuleAdditionalNatPageHeadText,
		Title:      natPage.Title,
		ImageUrl:   natPage.ShareImage,
		DescText_1: natPage.ShareTitle,
		Style:      api.ImageStyle_add_style_square,
		Type:       strconv.FormatInt(dynCtx.Dyn.Type, 10),
		CardType:   "official_activity",
	}
	if s.isEmptyHeadTextCapable(general) {
		common.HeadText = ""
	}
	if natPage.ShareCaption != "" {
		common.Title = natPage.ShareCaption
	}
	common.Url = natPage.SkipURL // web端使用 natPage.PcURL
	// 聚合预约状态
	var (
		buttonType    = api.AddButtonType_bt_jump
		buttonText    = s.c.Resource.Text.ModuleAdditionalNatPageButtonDefault
		buttonJumpURL = natPage.SkipURL // web端使用 natPage.PcURL
		buttonState   int
	)
	if act.ReserveID == 0 || act.ReserveItem == nil { // 无预约
	} else if act.ReserveItem.ActStatus == _actNoStart { // 有活动但未开始
		common.DescText_2 = s.c.Resource.Text.ModuleAdditionalNatPageNotStart
	} else if act.ReserveItem.ActStatus == _actStart { // 活动进行中
		common.DescText_2 = fmt.Sprintf("已有%v", model.StatString(act.ReserveItem.Total, "人预约"))
		buttonType = api.AddButtonType_bt_button
		buttonState = 1
		if act.ReserveItem.State == _reserve { // 已预约
			buttonState = 2
		}
	} else if act.ReserveItem.ActStatus == _actOff { // 活动结束
		common.DescText_2 = s.c.Resource.Text.ModuleAdditionalNatPageOver
	} else {
		log.Warn("addition miss offic_activity dynid(%v) type(%v) rid(%v) uid(%d) act(%v) unknown reserve type %v", dynCtx.Dyn.DynamicID, dynCtx.Dyn.Type, dynCtx.Dyn.Rid, dynCtx.Dyn.UID, act.NativeID, act.ReserveItem.ActStatus)
	}
	common.Button = &api.AdditionalButton{
		Type:    buttonType,
		JumpUrl: buttonJumpURL,
		JumpStyle: &api.AdditionalButtonStyle{
			Text: buttonText,
		},
		Uncheck: &api.AdditionalButtonStyle{
			Text: s.c.Resource.Text.ModuleAdditionalNatPageUncheck,
		},
		Check: &api.AdditionalButtonStyle{
			Text: s.c.Resource.Text.ModuleAdditionalNatPageCheck,
		},
		Status: api.AdditionalButtonStatus(buttonState),
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_common,
		Item: &api.ModuleAdditional_Common{
			Common: common,
		},
	}
	return additional
}

func (s *Service) additionalTopic(res []*mdlv2.Topic, dynCtx *mdlv2.DynamicContext, topicAdditiveCard map[int64]*dyntopicextgrpc.TopicAdditiveCard, general *mdlv2.GeneralParam) *api.ModuleAdditional {
	for _, topic := range res {
		if topic.Stat != 1 { // 未绑定 conitnue
			continue
		}
		headText := s.c.Resource.Text.ModuleAdditionalTopicHeadText
		if s.isEmptyHeadTextCapable(general) {
			headText = ""
		}
		descText2 := ""
		var jumpUrl string
		buttonText := s.c.Resource.Text.ModuleAdditionalTopicButtonText
		if additiveCard, ok := topicAdditiveCard[topic.TopicID]; ok {
			headText = additiveCard.AdditionalNotes
			descText2 = fmt.Sprintf("浏览%s·讨论%s", s.numTransfer(int(additiveCard.ViewCount)), s.numTransfer(int(additiveCard.DiscussCount)))
			jumpUrl = additiveCard.ButtonLink
			buttonText = additiveCard.ButtonTxt
		}
		if jumpUrl == "" {
			jumpUrl = topic.TopicLink
		}
		topicTitle := topic.ShareCaption
		if topicTitle == "" {
			topicTitle = topic.TopicName
		}
		common := &api.AdditionCommon{
			HeadText:   headText,
			Url:        topic.TopicLink,
			Title:      topicTitle,
			ImageUrl:   topic.ShareImage,
			DescText_1: topic.ShareTitle,
			DescText_2: descText2,
			HeadIcon:   "",
			Style:      api.ImageStyle_add_style_square,
			Type:       strconv.FormatInt(dynCtx.Dyn.Type, 10),
			CardType:   "topic",
		}
		common.Button = &api.AdditionalButton{
			Type:    api.AddButtonType_bt_jump,
			JumpUrl: jumpUrl,
			JumpStyle: &api.AdditionalButtonStyle{
				Text: buttonText,
			},
		}
		additional := &api.ModuleAdditional{
			Type: api.AdditionalType_additional_type_common,
			Rid:  topic.TopicID,
			Item: &api.ModuleAdditional_Common{
				Common: common,
			},
		}
		return additional
	}
	return nil
}

func (s *Service) additionUpActivity(dynCtx *mdlv2.DynamicContext, natPage *natpagegrpc.NativePageCard, general *mdlv2.GeneralParam) *api.ModuleAdditional {
	const (
		// 活动上线
		_online = 1
	)
	common := &api.AdditionCommon{
		HeadText:   s.c.Resource.Text.ModuleAdditionalNatPageHeadText,
		Title:      fmt.Sprintf("#%s#", natPage.Title),
		DescText_1: natPage.ShareTitle,
		ImageUrl:   natPage.ShareImage,
		Style:      api.ImageStyle_add_style_square,
		Type:       strconv.FormatInt(dynCtx.Dyn.Type, 10),
		CardType:   "up_activity",
	}
	if s.isEmptyHeadTextCapable(general) {
		common.HeadText = ""
	}
	if natPage.ShareCaption != "" {
		common.Title = fmt.Sprintf("#%s#", natPage.ShareCaption)
	}
	if userInfo, ok := dynCtx.GetUser(natPage.RelatedUid); ok {
		// 如果描述中有xxx发起，直接整个描述不下发
		if strings.Contains(natPage.ShareTitle, userInfo.Name+"发起") {
			common.DescText_1 = ""
		}
		common.DescText_2 = "发起人：" + userInfo.Name
	}
	common.Url = natPage.SkipURL // web端使用 natPage.PcURL
	// 聚合预约状态
	if natPage.State == _online {
		common.Button = &api.AdditionalButton{
			Type:    api.AddButtonType_bt_jump,
			JumpUrl: natPage.SkipURL,
			JumpStyle: &api.AdditionalButtonStyle{
				Text:    s.c.Resource.Text.AdditionUpActivityOnline,
				BgStyle: api.AddButtonBgStyle_fill,
				Disable: api.DisableState_highlight,
			},
		}
	} else {
		common.Button = &api.AdditionalButton{
			Type:   api.AddButtonType_bt_button,
			Status: api.AdditionalButtonStatus_check,
			Check: &api.AdditionalButtonStyle{
				Text:    s.c.Resource.Text.AdditionUpActivityOffline,
				BgStyle: api.AddButtonBgStyle_gray,
				Disable: api.DisableState_gary,
			},
		}
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_common,
		Item: &api.ModuleAdditional_Common{
			Common: common,
		},
	}
	return additional
}

func (s *Service) additionalOGV(res *pgcDynGrpc.FollowCardProto, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) *api.ModuleAdditional {
	common := &api.AdditionCommon{
		HeadText:   s.c.Resource.Text.ModuleAdditionalOGVHeadText,
		Title:      res.Title,
		ImageUrl:   res.Cover,
		DescText_1: res.BadgeReleaseShow,
		DescText_2: res.FollowDesc,
		Url:        res.Link,
		HeadIcon:   "",
		Style:      api.ImageStyle_add_style_vertical,
		Type:       strconv.FormatInt(dynCtx.Dyn.Type, 10),
		CardType:   "ogv",
	}
	if s.isEmptyHeadTextCapable(general) {
		common.HeadText = ""
	}
	if res.FollowInfo != nil {
		common.Button = &api.AdditionalButton{
			Type:      api.AddButtonType_bt_button,
			JumpStyle: nil,
			Uncheck: &api.AdditionalButtonStyle{
				Icon: res.FollowInfo.UnfollowIcon,
				Text: res.FollowInfo.UnfollowText,
			},
			Check: &api.AdditionalButtonStyle{
				Icon: res.FollowInfo.FollowIcon,
				Text: res.FollowInfo.FollowText,
			},
			Status: api.AdditionalButtonStatus(1),
		}
		// 服务端返回 0未追剧 1已追剧
		// 端上需要转为 1未追剧 2已追剧
		if res.FollowInfo.IsFollow == 1 {
			common.Button.Status = api.AdditionalButtonStatus(2)
		}
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_common,
		Item: &api.ModuleAdditional_Common{
			Common: common,
		},
	}
	return additional
}

func (s *Service) additionalMatch(res *esportGrpc.ContestDetail, dynCtx *mdlv2.DynamicContext) *api.ModuleAdditional {
	const (
		_buttonUnCheck = 1
		_buttonCheck   = 2
	)
	moba := &api.AdditionEsportMoba{
		Title:    res.GameStage1,
		Uri:      res.JumpURL,
		SubTitle: res.GameStage2,
		Type:     strconv.FormatInt(dynCtx.Dyn.Type, 10),
		CardType: "match",
	}
	if res.Season != nil {
		moba.HeadText = fmt.Sprintf("%v - %s", s.c.Resource.Text.ModuleAdditionalMatchHeadText, res.Season.Title)
	}
	// 队伍信息
	if res.HomeTeam != nil {
		moba.MatchTeam = append(moba.MatchTeam, &api.MatchTeam{
			Id:         res.HomeTeam.ID,
			Name:       res.HomeTeam.Title,
			Cover:      res.HomeTeam.LogoFull,
			Color:      s.c.Resource.Others.ModuleAdditionalMatchTeam.TextColor,
			NightColor: s.c.Resource.Others.ModuleAdditionalMatchTeam.TextColorNight,
		})
	}
	if res.AwayTeam != nil {
		moba.MatchTeam = append(moba.MatchTeam, &api.MatchTeam{
			Id:         res.AwayTeam.ID,
			Name:       res.AwayTeam.Title,
			Cover:      res.AwayTeam.LogoFull,
			Color:      s.c.Resource.Others.ModuleAdditionalMatchTeam.TextColor,
			NightColor: s.c.Resource.Others.ModuleAdditionalMatchTeam.TextColorNight,
		})
	}
	// 中断比赛状态文案和比分信息
	moba.AdditionEsportMobaStatus = &api.AdditionEsportMobaStatus{}
	var (
		statusDes []*api.AdditionEsportMobaStatusDesc
	)
	switch res.ContestStatus {
	case esportGrpc.ContestStatusEnum_Waiting: // 未开始
		// 比赛状态文案
		moba.AdditionEsportMobaStatus.Status = 1
		moba.AdditionEsportMobaStatus.Title = model.FormMatchTime(res.Stime)
		moba.AdditionEsportMobaStatus.Color = s.c.Resource.Others.ModuleAdditionalMatchState.TextColor
		moba.AdditionEsportMobaStatus.NightColor = s.c.Resource.Others.ModuleAdditionalMatchState.TextColorNight
		// 比赛信息
		statusDes = append(statusDes, &api.AdditionEsportMobaStatusDesc{
			Title:      s.c.Resource.Others.ModuleAdditionalMatchVS.Text,
			Color:      s.c.Resource.Others.ModuleAdditionalMatchVS.TextColor,
			NightColor: s.c.Resource.Others.ModuleAdditionalMatchVS.TextColorNight,
		})
		moba.Button = &api.AdditionalButton{
			Type: api.AddButtonType_bt_button,
		}
		moba.Button.Uncheck = &api.AdditionalButtonStyle{
			Text: s.c.Resource.Text.ModuleAdditionalMatchButtonUncheck,
		}
		moba.Button.Check = &api.AdditionalButtonStyle{
			Text: s.c.Resource.Text.ModuleAdditionalMatchButtonCheck,
		}
		moba.Button.Status = api.AdditionalButtonStatus(_buttonUnCheck)
		if res.IsSubscribed == esportGrpc.SubscribedStatusEnum_CanSubSubed { // 已订阅
			moba.Button.Status = api.AdditionalButtonStatus(_buttonCheck)
		}
	default:
		var buttonText, jumpURL string
		// nolint:exhaustive
		switch res.ContestStatus {
		case esportGrpc.ContestStatusEnum_Ing: // 进行中
			moba.AdditionEsportMobaStatus.Status = 2
			moba.AdditionEsportMobaStatus.Title = s.c.Resource.Others.ModuleAdditionalMatching.Text
			moba.AdditionEsportMobaStatus.Color = s.c.Resource.Others.ModuleAdditionalMatching.TextColor
			moba.AdditionEsportMobaStatus.NightColor = s.c.Resource.Others.ModuleAdditionalMatching.TextColorNight
			if res.LiveRoom != 0 {
				// 按钮
				buttonText = s.c.Resource.Text.ModuleAdditionalMatchStartedButtonLiveing
				jumpURL = model.FillURI(model.GotoLive, strconv.FormatInt(res.LiveRoom, 10), nil)
			}
		case esportGrpc.ContestStatusEnum_Over: // 已结束
			moba.AdditionEsportMobaStatus.Status = 3
			moba.AdditionEsportMobaStatus.Title = s.c.Resource.Others.ModuleAdditionalMatchOver.Text
			moba.AdditionEsportMobaStatus.Color = s.c.Resource.Others.ModuleAdditionalMatchOver.TextColor
			moba.AdditionEsportMobaStatus.NightColor = s.c.Resource.Others.ModuleAdditionalMatchOver.TextColorNight
			if res.GetPlayback() != "" {
				// 按钮
				buttonText = s.c.Resource.Text.ModuleAdditionalMatchStartedButtonPlayback
				jumpURL = res.GetPlayback()
			}
		}
		if buttonText != "" && jumpURL != "" {
			moba.Button = &api.AdditionalButton{
				Type:    api.AddButtonType_bt_jump,
				JumpUrl: jumpURL,
				JumpStyle: &api.AdditionalButtonStyle{
					Text: buttonText,
				},
			}
		}
		// 比赛信息
		// 主队默认深色
		homeScoreColor := s.c.Resource.Others.ModuleAdditionalMatchDard.TextColor
		homeScoreNightColor := s.c.Resource.Others.ModuleAdditionalMatchDard.TextColorNight
		// 客队默认深色
		awayScoreColor := s.c.Resource.Others.ModuleAdditionalMatchDard.TextColor
		awayScoreNightColor := s.c.Resource.Others.ModuleAdditionalMatchDard.TextColorNight
		// 比分低的转浅色
		if res.AwayScore < res.HomeScore {
			if res.AwayScore > res.HomeScore {
				awayScoreColor = s.c.Resource.Others.ModuleAdditionalMatchLight.TextColor
				awayScoreNightColor = s.c.Resource.Others.ModuleAdditionalMatchLight.TextColorNight
			}
		}
		if res.AwayScore > res.HomeScore {
			homeScoreColor = s.c.Resource.Others.ModuleAdditionalMatchLight.TextColor
			homeScoreNightColor = s.c.Resource.Others.ModuleAdditionalMatchLight.TextColorNight
		}
		statusDes = append(statusDes, &api.AdditionEsportMobaStatusDesc{
			Title:      strconv.FormatInt(res.HomeScore, 10),
			Color:      homeScoreColor,
			NightColor: homeScoreNightColor,
		})
		statusDes = append(statusDes, &api.AdditionEsportMobaStatusDesc{
			Title:      s.c.Resource.Others.ModuleAdditionalMatchMiddle.Text,
			Color:      s.c.Resource.Others.ModuleAdditionalMatchMiddle.TextColor,
			NightColor: s.c.Resource.Others.ModuleAdditionalMatchMiddle.TextColorNight,
		})
		statusDes = append(statusDes, &api.AdditionEsportMobaStatusDesc{
			Title:      strconv.FormatInt(res.AwayScore, 10),
			Color:      awayScoreColor,
			NightColor: awayScoreNightColor,
		})
	}
	moba.AdditionEsportMobaStatus.AdditionEsportMobaStatusDesc = statusDes
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_esport,
		Item: &api.ModuleAdditional_Esport{
			Esport: &api.AdditionEsport{
				Type:  strconv.FormatInt(dynCtx.Dyn.Type, 10),
				Style: api.EspaceStyle_moba,
				Item: &api.AdditionEsport_AdditionEsportMoba{
					AdditionEsportMoba: moba,
				},
				CardType: "esport",
			},
		},
	}
	return additional
}

func (s *Service) additionalGame(game *gamemdl.Game, dynCtx *mdlv2.DynamicContext) *api.ModuleAdditional {
	// 第二行文案
	var descText1 string
	// 第三行文案
	var descText2 string
	if game.GameTags != nil {
		descText1 = strings.Join(game.GameTags, "/")
	}
	if game.GameSubtitle != "" {
		descText2 = game.GameSubtitle
	}
	url := model.FillURI(model.GotoURL, game.GameLink, model.DynamicIDHandler(dynCtx.Dyn.DynamicID))
	common := &api.AdditionCommon{
		HeadText:   s.c.Resource.Text.ModuleAdditionalGameHeadText,
		Title:      game.GameName,
		ImageUrl:   game.GameIcon,
		DescText_1: descText1,
		DescText_2: descText2,
		Url:        url,
		HeadIcon:   "",
		Style:      api.ImageStyle_add_style_square,
		Type:       strconv.FormatInt(dynCtx.Dyn.Type, 10),
		CardType:   "game",
	}
	common.Button = &api.AdditionalButton{
		Type: api.AddButtonType_bt_jump,
		JumpStyle: &api.AdditionalButtonStyle{
			Text: game.GameButton,
		},
		JumpUrl: url,
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_common,
		Item: &api.ModuleAdditional_Common{
			Common: common,
		},
	}
	return additional
}

func (s *Service) additionalManga(manga *comicmdl.Comic, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) *api.ModuleAdditional {
	var descText1, descText2 string
	switch manga.IsFinish {
	case 1:
		descText1 = fmt.Sprintf("【完结】共%v话", manga.Total)
	case 0:
		descText1 = fmt.Sprintf("更新至%v", manga.LastShortTitle)
	default:
		descText1 = "未开刊"
	}
	var descText2Tmp []string
	for _, style := range manga.Styles {
		if style != nil && style.Name != "" {
			descText2Tmp = append(descText2Tmp, style.Name)
		}
	}
	if len(descText2Tmp) > 0 {
		descText2 = strings.Join(descText2Tmp, ",")
	}
	common := &api.AdditionCommon{
		HeadText:   s.c.Resource.Text.ModuleAdditionalMangaHeadText,
		Title:      manga.Title,
		ImageUrl:   manga.VerticalCover,
		DescText_1: descText1,
		DescText_2: descText2,
		Url:        model.FillURI(model.GotoURL, manga.URL, model.SuffixHandler("from=sub_card")),
		HeadIcon:   "",
		Style:      api.ImageStyle_add_style_vertical,
		Type:       strconv.FormatInt(dynCtx.Dyn.Type, 10),
		CardType:   "manga",
	}
	if s.isEmptyHeadTextCapable(general) {
		common.HeadText = ""
	}
	// 追加参数
	common.Button = &api.AdditionalButton{
		Type: api.AddButtonType_bt_button,
		Uncheck: &api.AdditionalButtonStyle{
			Icon: s.c.Resource.Icon.ModuleAdditionalManga,
			Text: s.c.Resource.Text.ModuleAdditionalMangaButtonUncheck,
		},
		Check: &api.AdditionalButtonStyle{
			Text: s.c.Resource.Text.ModuleAdditionalMangaButtonCheck,
		},
		Status: api.AdditionalButtonStatus(1),
	}
	// 服务端返回 0 未追；1 已追
	// 端上需要转为 1未追剧 2已追剧
	if manga.FavStatus == 1 {
		common.Button.Status = api.AdditionalButtonStatus(2)
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_common,
		Item: &api.ModuleAdditional_Common{
			Common: common,
		},
	}
	return additional
}

func (s *Service) additionalDecorate(garb *garbmdl.DynamicGarbInfo, _ *mdlv2.DynamicContext) *api.ModuleAdditional {
	common := &api.AdditionCommon{
		HeadText:   s.c.Resource.Text.ModuleAdditionalDecorateHeadText,
		Title:      garb.Title,
		ImageUrl:   garb.Cover,
		DescText_1: garb.Extension1,
		DescText_2: garb.Extension2,
		Url:        garb.JumpUrl,
		CardType:   "decoration",
		Style:      api.ImageStyle_add_style_square,
	}
	var (
		buttonText  string
		buttonType  api.AddButtonType
		buttonState int
	)
	switch garb.Status {
	case garbStateSell:
		buttonType = api.AddButtonType_bt_jump
		buttonText = s.c.Resource.Text.ModuleAdditionalDecorateSell
	case garbStateReserve:
		buttonType = api.AddButtonType_bt_button
		buttonState = 1
		if garb.IsReserve {
			buttonState = 2
		}
	default:
		buttonType = api.AddButtonType_bt_jump
		buttonText = s.c.Resource.Text.ModuleAdditionalDecorateDefault
	}
	common.Button = &api.AdditionalButton{
		Type:    buttonType,
		JumpUrl: garb.JumpUrl,
		JumpStyle: &api.AdditionalButtonStyle{
			Text: buttonText,
		},
		Uncheck: &api.AdditionalButtonStyle{
			Text: s.c.Resource.Text.ModuleAdditionalDecorateButtonUncheck,
		},
		Check: &api.AdditionalButtonStyle{
			Text: s.c.Resource.Text.ModuleAdditionalDecorateButtonCheck,
		},
		Status: api.AdditionalButtonStatus(buttonState),
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_common,
		Item: &api.ModuleAdditional_Common{
			Common: common,
		},
	}
	return additional
}

func (s *Service) additionalPugv(res *cheesemdl.Cheese, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) *api.ModuleAdditional {
	common := &api.AdditionCommon{
		HeadText:   s.c.Resource.Text.ModuleAdditionalPUGVHeadText,
		Title:      res.Title,
		ImageUrl:   res.Cover,
		DescText_1: res.SubTitle,
		DescText_2: res.CardInfo,
		Style:      api.ImageStyle_add_style_vertical,
		Type:       strconv.FormatInt(dynCtx.Dyn.Type, 10),
		CardType:   "pugv",
	}
	if s.isEmptyHeadTextCapable(general) {
		common.HeadText = ""
	}
	if res.Button != nil {
		// url里面增加动态ID
		url := res.Button.JumpURL
		common.Url = url
		common.Button = &api.AdditionalButton{
			Type:    api.AddButtonType(res.Button.Type),
			JumpUrl: url,
		}
		if res.Button.JumpStyle != nil {
			common.Button.JumpStyle = &api.AdditionalButtonStyle{
				Icon: res.Button.JumpStyle.Icon,
				Text: res.Button.JumpStyle.Text,
			}
		}
		if res.Button.UnCheck != nil {
			common.Button.JumpStyle = &api.AdditionalButtonStyle{
				Icon: res.Button.UnCheck.Icon,
				Text: res.Button.UnCheck.Text,
			}
		}
		if res.Button.Check != nil {
			common.Button.JumpStyle = &api.AdditionalButtonStyle{
				Icon: res.Button.Check.Icon,
				Text: res.Button.Check.Text,
			}
		}
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_common,
		Item: &api.ModuleAdditional_Common{
			Common: common,
		},
	}
	return additional
}

// nolint:gocognit
func (s *Service) additionalVote(vote *dyncommongrpc.VoteInfo, dynCtx *mdlv2.DynamicContext) *api.ModuleAdditional {
	var (
		voteTmp      *api.AdditionVote2
		labels       []string
		now          = time.Now().Unix()
		voteTotalCnt int32
	)
	if vote.Status == _voteDel || vote.Status == _voteRefuse {
		voteTmp = &api.AdditionVote2{
			AdditionVoteType: api.AdditionVoteType_addition_vote_type_none,
			Tips:             s.c.Resource.Text.ModuleAdditionalVoteTips,
		}
		goto END
	}
	voteTmp = &api.AdditionVote2{
		VoteId:             vote.GetVoteId(),
		Title:              vote.GetTitle(),
		Deadline:           vote.GetEndTime(),
		OpenText:           s.c.Resource.Text.ModuleAdditionalVoteOpen,
		CloseText:          s.c.Resource.Text.ModuleAdditionalVoteClose,
		VotedText:          s.c.Resource.Text.ModuleAdditionalVoteVoted,
		BizType:            vote.GetBizType(),
		Total:              vote.GetJoinNum(),
		CardType:           "vote",
		Uri:                fmt.Sprintf(model.VoteURI, vote.VoteId, dynCtx.Dyn.DynamicID),
		ChoiceCnt:          vote.GetChoiceCnt(),
		DefauleSelectShare: true,
	}
	// 过期判断
	if now >= vote.EndTime {
		voteTmp.State = api.AdditionVoteState_addition_vote_state_close
	} else {
		voteTmp.State = api.AdditionVoteState_addition_vote_state_open
	}
	// 文案组装
	if vote.GetJoinNum() == 0 {
		labels = append(labels, "0人参与") // 不显示 -人投票 单独兼容
	} else {
		labels = append(labels, model.StatString(vote.GetJoinNum(), "人投票"))
	}
	voteTmp.Label = strings.Join(labels, "·")
	for _, option := range vote.Options {
		if option == nil {
			continue
		}
		voteTotalCnt += option.GetCnt()
	}
	// 详情部分 文字、图片、隐藏选项
	if (vote.Status == _voteWait || vote.Status == _voteOK || vote.Status == _voteDead) && dynCtx.Dyn.IsWord() && vote.OptionsCnt <= 4 {
		switch vote.Type {
		case _voteTypeWord:
			voteTmp.AdditionVoteType = api.AdditionVoteType_addition_vote_type_word
			var (
				items      []*api.AdditionVoteWordItem
				maxOptionm = make(map[int32][]int)
				maxCnt     int32
			)
			for _, option := range vote.Options {
				if option == nil {
					continue
				}
				// 选项基础信息
				item := &api.AdditionVoteWordItem{
					OptIdx: option.GetOptIdx(),
					Title:  option.GetOptDesc(),
					Total:  option.GetCnt(),
				}
				// 选项票数百分比
				if voteTotalCnt != 0 {
					var err error
					item.Persent, err = strconv.ParseFloat(fmt.Sprintf("%.2f", float32(option.GetCnt())/float32(voteTotalCnt)), 64)
					if err != nil {
						log.Error("additionalVote persent option %+v, err %v", option, err)
					}
				}
				// 投票状态
				for _, idx := range vote.MyVotes {
					if idx == item.GetOptIdx() {
						item.IsVote = true     // 是否投过当前选项
						voteTmp.IsVoted = true // 是否参与过投票
					}
				}
				items = append(items, item)
				// 记录票数对应的选项idx
				maxOptionm[item.Total] = append(maxOptionm[item.Total], len(items)-1)
				// 各选项中最大票数
				if item.Total > maxCnt {
					maxCnt = item.Total
				}
			}
			if len(items) == 0 {
				log.Warn("additional miss vote dynid(%v) items len 0", dynCtx.Dyn.DynamicID)
				return nil
			}
			// 回填标记票数最多的选项
			for _, idx := range maxOptionm[maxCnt] {
				items[idx].IsMaxOption = true
			}
			voteTmp.Item = &api.AdditionVote2_AdditionVoteWord{
				AdditionVoteWord: &api.AdditionVoteWord{
					Item: items,
				},
			}
		case _voteTypePic:
			voteTmp.AdditionVoteType = api.AdditionVoteType_addition_vote_type_pic
			var (
				items      []*api.AdditionVotePicItem
				maxOptionm = make(map[int32][]int)
				maxCnt     int32
			)
			for _, option := range vote.Options {
				if option == nil {
					continue
				}
				// 选项基础信息
				item := &api.AdditionVotePicItem{
					OptIdx: option.GetOptIdx(),
					Cover:  option.GetImgUrl(),
					Total:  option.GetCnt(),
					Title:  option.GetOptDesc(),
				}
				// 选项票数百分比
				if voteTotalCnt != 0 {
					var err error
					item.Persent, err = strconv.ParseFloat(fmt.Sprintf("%.2f", float32(option.GetCnt())/float32(voteTotalCnt)), 64)
					if err != nil {
						log.Error("additionalVote persent option %+v, err %v", option, err)
					}
				}
				// 投票状态
				for _, idx := range vote.MyVotes {
					if idx == item.GetOptIdx() {
						item.IsVote = true     // 是否投过当前选项
						voteTmp.IsVoted = true // 是否参与过投票
					}
				}
				items = append(items, item)
				// 记录票数对应的选项idx
				maxOptionm[item.Total] = append(maxOptionm[item.Total], len(items)-1)
				// 各选项中最大票数
				if item.Total > maxCnt {
					maxCnt = item.Total
				}
			}
			if len(items) == 0 {
				log.Warn("additional miss vote dynid(%v) items len 0", dynCtx.Dyn.DynamicID)
				return nil
			}
			// 回填标记票数最多的选项
			for _, idx := range maxOptionm[maxCnt] {
				items[idx].IsMaxOption = true
			}
			voteTmp.Item = &api.AdditionVote2_AdditionVotePic{
				AdditionVotePic: &api.AdditionVotePic{
					Item: items,
				},
			}
		}
	} else {
		voteTmp.AdditionVoteType = api.AdditionVoteType_addition_vote_type_default
		var items []string
		for _, option := range vote.Options {
			if option == nil {
				continue
			}
			items = append(items, option.GetImgUrl())
		}
		voteTmp.Item = &api.AdditionVote2_AdditionVoteDefaule{
			AdditionVoteDefaule: &api.AdditionVoteDefaule{
				Cover: items,
			},
		}
	}
END:
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_vote,
		Item: &api.ModuleAdditional_Vote2{
			Vote2: voteTmp,
		},
	}
	dynCtx.Interim.VoteID = voteTmp.VoteId // 转发内外层传递
	return additional
}

// nolint:gocognit
func (s *Service) votedResult(vote *dyncommongrpc.VoteInfo, req *api.DynVoteReq) *api.ModuleAdditional {
	var (
		voteTmp      *api.AdditionVote2
		labels       []string
		now          = time.Now().Unix()
		voteTotalCnt int32
	)
	if vote.Status == _voteDel || vote.Status == _voteRefuse {
		voteTmp = &api.AdditionVote2{
			AdditionVoteType: api.AdditionVoteType_addition_vote_type_none,
			Tips:             s.c.Resource.Text.ModuleAdditionalVoteTips,
		}
		return nil
	}
	voteTmp = &api.AdditionVote2{
		VoteId:    vote.GetVoteId(),
		Title:     vote.GetTitle(),
		Deadline:  vote.GetEndTime(),
		OpenText:  s.c.Resource.Text.ModuleAdditionalVoteOpen,
		CloseText: s.c.Resource.Text.ModuleAdditionalVoteClose,
		VotedText: s.c.Resource.Text.ModuleAdditionalVoteVoted,
		BizType:   vote.GetBizType(),
		Total:     vote.GetJoinNum(),
		CardType:  "vote",
		Uri:       fmt.Sprintf(model.VoteURI, vote.VoteId, req.DynamicId),
		ChoiceCnt: vote.GetChoiceCnt(),
	}
	// 过期判断
	if now >= vote.EndTime {
		voteTmp.State = api.AdditionVoteState_addition_vote_state_close
	} else {
		voteTmp.State = api.AdditionVoteState_addition_vote_state_open
	}
	// 文案组装
	if vote.GetJoinNum() == 0 {
		labels = append(labels, "0人投票") // 不显示 -人投票 单独兼容
	} else {
		labels = append(labels, model.StatString(vote.GetJoinNum(), "人投票"))
	}
	voteTmp.Label = strings.Join(labels, "·")
	for _, option := range vote.Options {
		if option == nil {
			continue
		}
		voteTotalCnt += option.GetCnt()
	}
	// 详情部分 文字、图片、隐藏选项
	switch vote.Type {
	case _voteTypeWord:
		voteTmp.AdditionVoteType = api.AdditionVoteType_addition_vote_type_word
		var (
			items      []*api.AdditionVoteWordItem
			maxOptionm = make(map[int32][]int)
			maxCnt     int32
		)
		for _, option := range vote.Options {
			if option == nil {
				continue
			}
			// 选项基础信息
			item := &api.AdditionVoteWordItem{
				OptIdx: option.GetOptIdx(),
				Title:  option.GetOptDesc(),
				Total:  option.GetCnt(),
			}
			// 选项票数百分比
			if voteTotalCnt != 0 {
				var err error
				item.Persent, err = strconv.ParseFloat(fmt.Sprintf("%.2f", float32(option.GetCnt())/float32(voteTotalCnt)), 64)
				if err != nil {
					log.Error("additionalVote persent option %+v, err %v", option, err)
				}
			}
			// 投票状态
			for _, idx := range vote.MyVotes {
				if idx == item.GetOptIdx() {
					item.IsVote = true     // 是否投过当前选项
					voteTmp.IsVoted = true // 是否参与过投票
				}
			}
			items = append(items, item)
			// 记录票数对应的选项idx
			maxOptionm[item.Total] = append(maxOptionm[item.Total], len(items)-1)
			// 各选项中最大票数
			if item.Total > maxCnt {
				maxCnt = item.Total
			}
		}
		if len(items) == 0 {
			log.Warn("votedResult miss vote dynid(%v) voteID(%v) items len 0", req.DynamicId, req.VoteId)
			return nil
		}
		// 回填标记票数最多的选项
		for _, idx := range maxOptionm[maxCnt] {
			items[idx].IsMaxOption = true
		}
		voteTmp.Item = &api.AdditionVote2_AdditionVoteWord{
			AdditionVoteWord: &api.AdditionVoteWord{
				Item: items,
			},
		}
	case _voteTypePic:
		voteTmp.AdditionVoteType = api.AdditionVoteType_addition_vote_type_pic
		var (
			items      []*api.AdditionVotePicItem
			maxOptionm = make(map[int32][]int)
			maxCnt     int32
		)
		for _, option := range vote.Options {
			if option == nil {
				continue
			}
			// 选项基础信息
			item := &api.AdditionVotePicItem{
				OptIdx: option.GetOptIdx(),
				Cover:  option.GetImgUrl(),
				Total:  option.GetCnt(),
				Title:  option.GetOptDesc(),
			}
			// 选项票数百分比
			if voteTotalCnt != 0 {
				var err error
				item.Persent, err = strconv.ParseFloat(fmt.Sprintf("%.2f", float32(option.GetCnt())/float32(voteTotalCnt)), 64)
				if err != nil {
					log.Error("additionalVote persent option %+v, err %v", option, err)
				}
			}
			// 投票状态
			for _, idx := range vote.MyVotes {
				if idx == item.GetOptIdx() {
					item.IsVote = true     // 是否投过当前选项
					voteTmp.IsVoted = true // 是否参与过投票
				}
			}
			items = append(items, item)
			// 记录票数对应的选项idx
			maxOptionm[item.Total] = append(maxOptionm[item.Total], len(items)-1)
			// 各选项中最大票数
			if item.Total > maxCnt {
				maxCnt = item.Total
			}
		}
		if len(items) == 0 {
			log.Warn("votedResult miss vote dynid(%v) voteID(%v) items len 0", req.DynamicId, req.VoteId)
			return nil
		}
		voteTmp.Item = &api.AdditionVote2_AdditionVotePic{
			AdditionVotePic: &api.AdditionVotePic{
				Item: items,
			},
		}
	default:
		log.Error("unkonw vote type %v, req %v", vote.Type, req)
		return nil
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_vote,
		Item: &api.ModuleAdditional_Vote2{
			Vote2: voteTmp,
		},
	}
	return additional
}

// additionalUPInfo up主预约卡，info
// nolint:gocognit
func (s *Service) additionalUPInfo(c context.Context, up *activitygrpc.UpActReserveRelationInfo, reserveDove *activitygrpc.ReserveDoveActRelationInfo, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) (*api.AdditionUP, bool, string) {
	const (
		_limitToast  = "请在手机打开最新版本app查看"
		_noCardToast = "原预约信息已删除"
	)
	// mid > int32老版本抛弃当前卡片
	if s.checkMidMaxInt32(c, up.Upmid, general) {
		return nil, false, _limitToast
	}
	const (
		_liveStart = 1
		_liveAv    = 2
	)
	common := &api.AdditionUP{
		Title:        up.Title,
		DescText_2:   model.UpStatString(up.Total, "人预约"),
		CardType:     "reserve",
		ReserveTotal: up.Total,
		Rid:          up.Sid,
		LotteryType:  api.ReserveRelationLotteryType_reserve_relation_lottery_type_default,
		UpMid:        up.Upmid,
		DynamicId:    up.DynamicId,
		ShowText_2:   true,
	}
	if userInfo, ok := dynCtx.GetUser(up.Upmid); ok {
		common.UserInfo = &api.AdditionUserInfo{
			Name: userInfo.Name,
			Face: userInfo.Face,
		}
	}
	// 主态展示 预约人数
	if up.Upmid != general.Mid {
		if up.Total < up.ReserveTotalShowLimit {
			common.ShowText_2 = false
		}
	}
	// 定时抽奖
	if up.LotteryType == activitygrpc.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron {
		common.LotteryType = api.ReserveRelationLotteryType_reserve_relation_lottery_type_cron
	}
	// 高亮icon跳链文案
	if up.PrizeInfo != nil {
		common.DescText_3 = &api.HighlightText{
			Text:      up.PrizeInfo.Text,
			JumpUrl:   up.PrizeInfo.JumpUrl,
			TextStyle: api.HighlightTextStyle_style_highlight,
			Icon:      s.c.Resource.Icon.AdditionalAdditionalCron,
		}
	}
	// 鸽子蛋皮肤
	if reserveDove != nil && reserveDove.Skin != nil {
		common.ActSkin = &api.AdditionalActSkin{
			Svga:      reserveDove.Skin.Svga,
			LastImage: reserveDove.Skin.LastImg,
			PlayTimes: reserveDove.Skin.PlayTimes,
		}
	}
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynReserveDesc, &feature.OriginResutl{
		BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynReserveDescIOS) ||
			(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynReserveDescAndroid) ||
			(general.IsPad() && general.GetBuild() >= s.c.BuildLimit.DynReservePadIOSPad) ||
			(general.IsPadHD() && general.GetBuild() >= s.c.BuildLimit.DynReservePadHD) ||
			(general.IsAndroidHD() && general.GetBuild() > s.c.BuildLimit.DynReservePadAndroid)}) {
		// nolint:exhaustive
		switch up.Type {
		case activitygrpc.UpActReserveRelationType_Archive: // 稿件
			if up.LivePlanStartTime.Time().Unix() > 0 {
				common.DescText_1 = &api.HighlightText{
					Text: fmt.Sprintf("预计%s发布", model.UpPubDataString(up.LivePlanStartTime.Time())),
				}
			}
		case activitygrpc.UpActReserveRelationType_Live: // 直播
			if up.LivePlanStartTime.Time().Unix() > 0 {
				common.DescText_1 = &api.HighlightText{
					Text: fmt.Sprintf("%s直播", model.UpPubDataString(up.LivePlanStartTime.Time())),
				}
			}
			if up.Ext != "" {
				tmp := &activitygrpc.UpActReserveRelationInfoExtend{}
				// 大航海
				if err := json.Unmarshal([]byte(up.Ext), &tmp); err == nil && tmp.SubType == int64(activitygrpc.UpActReserveRelationSubType_Voyage) {
					common.BadgeText = "大航海专属"
				}
			}
		case activitygrpc.UpActReserveRelationType_ESports: // 赛事
			if up.Desc != "" {
				common.DescText_1 = &api.HighlightText{
					Text: up.Desc,
				}
			}
		case activitygrpc.UpActReserveRelationType_Course: // 课堂预约
			if up.LivePlanStartTime.Time().Unix() > 0 {
				common.DescText_1 = &api.HighlightText{
					Text: fmt.Sprintf("%s 开售", model.UpPubDataString(up.LivePlanStartTime.Time())),
				}
			}
		}
	} else {
		// nolint:exhaustive
		switch up.Type {
		case activitygrpc.UpActReserveRelationType_Archive: // 稿件
			common.DescText_1 = &api.HighlightText{
				Text: "视频预约",
			}
		case activitygrpc.UpActReserveRelationType_Live: // 直播
			common.DescText_1 = &api.HighlightText{
				Text: model.UpPubDataString(up.LivePlanStartTime.Time()) + " 进行直播",
			}
		}
	}
	// 首映
	if up.Type == activitygrpc.UpActReserveRelationType_Premiere && (general.IsIPhonePick() && general.GetBuild() < s.c.BuildLimit.DynPropertyIOS || general.IsAndroidPick() && general.GetBuild() < s.c.BuildLimit.DynPropertyAndroid || general.IsPad() || general.IsPadHD() || general.IsAndroidHD()) {
		return nil, false, _limitToast
	}
	// nolint:exhaustive
	switch up.Type {
	case activitygrpc.UpActReserveRelationType_Premiere: // 首映
		common.DescText_1 = &api.HighlightText{
			Text: fmt.Sprintf("%s首映", model.UpPubDataString(up.LivePlanStartTime.Time())),
		}
		common.IsPremiere = true
	}
	// 主人态
	if up.Upmid == general.Mid {
		common.Button = &api.AdditionalButton{
			Type:    api.AddButtonType_bt_button,
			Status:  api.AdditionalButtonStatus_uncheck,
			Uncheck: upButtonCheck("", _upbuttonCancel, api.AddButtonBgStyle_stroke, api.DisableState_highlight),
			Check:   upButtonCheck("UP主已撤销预约", _upbuttonCancelOk, api.AddButtonBgStyle_fill, api.DisableState_gary), // 置灰不可点击
		}
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.LotteryTypeCron, &feature.OriginResutl{
			MobiApp: general.GetMobiApp(),
			Device:  general.GetDevice(),
			Build:   general.GetBuild(),
			BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.LotteryTypeCronIOS) ||
				(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.LotteryTypeCronAndroid) ||
				(general.IsPad() && general.GetBuild() >= s.c.BuildLimit.DynReservePadIOSPad) ||
				(general.IsPadHD() && general.GetBuild() >= s.c.BuildLimit.DynReservePadHD) ||
				(general.IsAndroidHD() && general.GetBuild() > s.c.BuildLimit.DynReservePadAndroid)}) {
			if up.LotteryType == activitygrpc.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron {
				common.Button.Uncheck = upButtonCheck("", _upbuttonCancelLotteryCron, api.AddButtonBgStyle_stroke, api.DisableState_highlight)
			}
		}
		common.Button.Uncheck.Share, common.BusinessId, common.DynType = s.additionalButtonShare(c, general, dynCtx, up)
	} else if up.Upmid != general.Mid { // 客人态
		if up.IsFollow == 1 {
			common.Button = &api.AdditionalButton{
				Type:    api.AddButtonType_bt_button,
				Status:  api.AdditionalButtonStatus_check,
				Check:   upButtonCheck("", _upbuttonReservationOk, api.AddButtonBgStyle_fill, api.DisableState_highlight),
				Uncheck: upButtonCheck("", _upbuttonReservation, api.AddButtonBgStyle_fill, api.DisableState_highlight),
			}
		} else {
			common.Button = &api.AdditionalButton{
				Type:    api.AddButtonType_bt_button,
				Status:  api.AdditionalButtonStatus_uncheck,
				Uncheck: upButtonCheck("", _upbuttonReservation, api.AddButtonBgStyle_fill, api.DisableState_highlight),
				Check:   upButtonCheck("", _upbuttonReservationOk, api.AddButtonBgStyle_fill, api.DisableState_highlight),
			}
		}
		if up.LotteryType == activitygrpc.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron {
			common.Button.Uncheck.Toast = "预约成功，已参与抽奖"
		}
		common.Button.Check.Share, common.BusinessId, common.DynType = s.additionalButtonShare(c, general, dynCtx, up)
	}
	// 仅主态可见
	if up.UpActVisible == activitygrpc.UpActVisible_OnlyUpVisible {
		if up.Upmid != general.Mid {
			return nil, true, ""
		}
		common.Title = "[审核中]" + common.Title
	}
	// nolint:exhaustive
	switch upActState(up.State) {
	case _upAudit:
		// 当前是预约先审后发且当前用户不是主态，不下发预约卡
		if upActState(up.State) == _upAudit && up.Upmid != general.Mid {
			return nil, true, ""
		}
		common.Title = "[审核中]" + common.Title
	case _upOnline:
		// nolint:exhaustive
		switch up.Type {
		case activitygrpc.UpActReserveRelationType_Archive, activitygrpc.UpActReserveRelationType_Premiere: // 稿件、首映
			aid, _ := strconv.ParseInt(up.Oid, 10, 64)
			ap, ok := dynCtx.GetArchive(aid)
			common.Url = model.FillURI(model.GotoAv, strconv.FormatInt(aid, 10), func(uri string) string {
				if ok && ap.GetArc().GetPremiere().GetState() == archivegrpc.PremiereState_premiere_in {
					return s.inArchivePremiereArg()(uri)
				}
				return uri
			}) // 不要秒开，且稿件不存在还是返回url，进入详情页展示稿件不存在
			if !ok {
				if up.IsFollow == 1 && up.Upmid != general.Mid {
					common.Button = &api.AdditionalButton{
						Type:   api.AddButtonType_bt_button,
						Status: api.AdditionalButtonStatus_check,
						Check:  upButtonCheck("不在预约时间", _upbuttonReservationOk, api.AddButtonBgStyle_fill, api.DisableState_gary),
					}
				} else {
					common.Button = &api.AdditionalButton{
						Type:   api.AddButtonType_bt_button,
						Status: api.AdditionalButtonStatus_check,
						Check:  upButtonCheck("不在预约时间", _upbuttonReservation, api.AddButtonBgStyle_fill, api.DisableState_gary),
					}
				}
			} else {
				arc := ap.Arc
				common.DescText_2 = model.UpStatString(int64(arc.Stat.View), "观看")
				common.Button = &api.AdditionalButton{
					Type:      api.AddButtonType_bt_jump,
					Status:    api.AdditionalButtonStatus_uncheck,
					JumpUrl:   model.FillURI(model.GotoAv, strconv.FormatInt(arc.Aid, 10), model.AvPlayHandlerGRPCV2(ap, arc.FirstCid, true)),
					JumpStyle: upButtonCheck("", _upbuttonWatch, api.AddButtonBgStyle_fill, api.DisableState_highlight),
				}
				// 首映
				if up.Type == activitygrpc.UpActReserveRelationType_Premiere {
					countInfo, ok := dynCtx.ResPlayUrlCount[ap.Arc.Aid]
					if arc.Premiere != nil {
						switch arc.Premiere.State {
						case archivegrpc.PremiereState_premiere_in:
							common.DescText_1 = &api.HighlightText{
								Text:      "首映中",
								TextStyle: api.HighlightTextStyle_style_highlight,
							}
							// 首映中添加参数
							common.Button.JumpUrl = s.inArchivePremiereArg()(common.Button.JumpUrl)
							if ok {
								common.DescText_2 = model.UpStatString(countInfo.Count[_play_online_total], "人在线")
							}
						case archivegrpc.PremiereState_premiere_after:
							if ok {
								common.DescText_2 = model.UpStatString(countInfo.Count[_play_online_total], "观看")
							}
						}
					}
				}
			}
		case activitygrpc.UpActReserveRelationType_Live: // 直播
			var isok bool
			if info, ok := dynCtx.ResLiveSessionInfo[up.Oid]; ok {
				live, ok := info.SessionInfoPerLive[up.Oid]
				if ok {
					isok = true
					switch live.Status {
					case _liveStart:
						common.Url = model.FillURI(model.GotoLive, strconv.FormatInt(info.RoomId, 10), nil)
						if liveUrl, ok := info.JumpUrl["dt_booking_dt"]; ok {
							common.Url = liveUrl
						}
						common.DescText_2 = model.UpStatString(live.PopularityCount, "人气")
						if show := live.WatchedShow; show != nil && show.TextLarge != "" {
							common.DescText_2 = show.TextLarge
						}
						common.Button = &api.AdditionalButton{
							Type:      api.AddButtonType_bt_jump,
							Status:    api.AdditionalButtonStatus_uncheck,
							JumpUrl:   common.Url,
							JumpStyle: upButtonCheck("", _upbuttonWatch, api.AddButtonBgStyle_fill, api.DisableState_highlight),
						}
						common.DescText_1 = &api.HighlightText{
							Text:      "直播中",
							TextStyle: api.HighlightTextStyle_style_highlight,
						}
					case _liveAv:
						aid, _ := bvid.BvToAv(live.Bvid)
						common.Url = model.FillURI(model.GotoAv, strconv.FormatInt(aid, 10), nil)
						common.Button = &api.AdditionalButton{
							Type:      api.AddButtonType_bt_jump,
							Status:    api.AdditionalButtonStatus_uncheck,
							JumpUrl:   common.Url,
							JumpStyle: upButtonCheck("", _upbuttonReplay, api.AddButtonBgStyle_fill, api.DisableState_highlight),
						}
						_, ok := dynCtx.ResArcs[aid]
						if !ok {
							isok = false
						}
					default:
						isok = false
					}
				}
			}
			if !isok {
				// 已结束（不可回放）
				common.Button = &api.AdditionalButton{
					Type:   api.AddButtonType_bt_button,
					Status: api.AdditionalButtonStatus_check,
					Check:  upButtonCheck("直播已结束", _upbuttonEnd, api.AddButtonBgStyle_fill, api.DisableState_gary),
				}
			}
		case activitygrpc.UpActReserveRelationType_ESports:
			// 赛事核销
			if up.State == activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackDone {
				common.Button = &api.AdditionalButton{
					Type:   api.AddButtonType_bt_button,
					Status: api.AdditionalButtonStatus_check,
					Check:  upButtonCheck("已结束", _upbuttonEnd, api.AddButtonBgStyle_fill, api.DisableState_gary),
				}
			}
		case activitygrpc.UpActReserveRelationType_Course:
			// 预约课程核销 已上线不可被预约 按钮变成去观看
			common.ShowText_2 = false
			common.DescText_1 = &api.HighlightText{
				Text: model.StatString(up.OidView, "人看过"),
			}
			// 上面那块会给
			if up.OidView == 0 {
				common.DescText_1.Text = "0人看过"
			}
			common.Button = &api.AdditionalButton{
				Type:      api.AddButtonType_bt_jump,
				Status:    api.AdditionalButtonStatus_uncheck,
				JumpUrl:   up.BaseJumpUrl,
				JumpStyle: upButtonCheck("", _upbuttonWatch, api.AddButtonBgStyle_fill, api.DisableState_highlight),
			}
		}
	case _upExpired:
		// nolint:exhaustive
		switch up.Type {
		case activitygrpc.UpActReserveRelationType_Live: // 直播
			// 客人 否则 主人
			if up.Upmid != general.Mid {
				if up.IsFollow == 1 {
					roomid, _ := strconv.ParseInt(up.Oid, 10, 64)
					common.Url = model.FillURI(model.GotoLive, strconv.FormatInt(roomid, 10), nil) // 开播未开播都返回url
					common.Button = &api.AdditionalButton{
						Type:   api.AddButtonType_bt_button,
						Status: api.AdditionalButtonStatus_check,
						Check:  upButtonCheck("预约已过期", _upbuttonReservationOk, api.AddButtonBgStyle_fill, api.DisableState_gary),
					}
				} else {
					common.Button = &api.AdditionalButton{
						Type:   api.AddButtonType_bt_button,
						Status: api.AdditionalButtonStatus_check,
						Check:  upButtonCheck("预约已过期", _upbuttonReservation, api.AddButtonBgStyle_fill, api.DisableState_gary),
					}
				}
			} else {
				common.Button = &api.AdditionalButton{
					Type:   api.AddButtonType_bt_button,
					Status: api.AdditionalButtonStatus_check,
					Check:  upButtonCheck("预约已过期", _upbuttonCancelOk, api.AddButtonBgStyle_fill, api.DisableState_gary),
				}
			}
		}
	case _upCancel:
		common.Button = &api.AdditionalButton{
			Type:   api.AddButtonType_bt_button,
			Status: api.AdditionalButtonStatus_check,
			Check:  upButtonCheck("UP主已撤销预约", _upbuttonCancelOk, api.AddButtonBgStyle_fill, api.DisableState_gary),
		}
		// 撤销不展示抽奖信息
		common.LotteryType = api.ReserveRelationLotteryType_reserve_relation_lottery_type_default
		common.DescText_3 = nil
	case _upDelete:
		// 预约卡被删除
		return nil, false, _noCardToast
	}
	if common.Button != nil {
		common.Button.ClickType = api.AdditionalButtonClickType_click_up
	}
	return common, true, ""
}

// additionalUP up主预约卡
func (s *Service) additionalUP(c context.Context, up *activitygrpc.UpActReserveRelationInfo, reserveDove *activitygrpc.ReserveDoveActRelationInfo, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) (*api.ModuleAdditional, bool, string) {
	common, ok, toast := s.additionalUPInfo(c, up, reserveDove, dynCtx, general)
	if common == nil {
		return nil, ok, toast
	}
	if common.Button != nil {
		common.Button.ClickType = api.AdditionalButtonClickType_click_up
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_up_reservation,
		Item: &api.ModuleAdditional_Up{
			Up: common,
		},
	}
	return additional, true, ""
}

// additionalUP up主预约卡，假卡
// nolint:gocognit
func (s *Service) additionalUPFake(c context.Context, up *activitygrpc.UpActReserveRelationInfo, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) (*api.ModuleAdditional, bool) {
	const (
		_liveStart = 1
		_liveAv    = 2
	)
	common := &api.AdditionUP{
		Title:        up.Title,
		DescText_2:   model.UpStatString(up.Total, "人预约"),
		CardType:     "reserve",
		ReserveTotal: up.Total,
		Rid:          up.Sid,
		LotteryType:  api.ReserveRelationLotteryType_reserve_relation_lottery_type_default,
		UpMid:        up.Upmid,
		DynamicId:    up.DynamicId,
		ShowText_2:   true,
	}
	if userInfo, ok := dynCtx.GetUser(up.Upmid); ok {
		common.UserInfo = &api.AdditionUserInfo{
			Name: userInfo.Name,
			Face: userInfo.Face,
		}
	}
	// 主态展示 预约人数
	if up.Upmid != general.Mid {
		if up.Total < up.ReserveTotalShowLimit {
			common.ShowText_2 = false
		}
	}
	// 定时抽奖
	if up.LotteryType == activitygrpc.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron {
		common.LotteryType = api.ReserveRelationLotteryType_reserve_relation_lottery_type_cron
	}
	// 高亮icon跳链文案
	if up.PrizeInfo != nil {
		common.DescText_3 = &api.HighlightText{
			Text:      up.PrizeInfo.Text,
			JumpUrl:   up.PrizeInfo.JumpUrl,
			TextStyle: api.HighlightTextStyle_style_highlight,
			Icon:      s.c.Resource.Icon.AdditionalAdditionalCron,
		}
	}
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynReserveDesc, &feature.OriginResutl{
		BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynReserveDescIOS) ||
			(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynReserveDescAndroid) ||
			(general.IsPad() && general.GetBuild() >= s.c.BuildLimit.DynReservePadIOSPad) ||
			(general.IsPadHD() && general.GetBuild() >= s.c.BuildLimit.DynReservePadHD) ||
			(general.IsAndroidHD() && general.GetBuild() > s.c.BuildLimit.DynReservePadAndroid)}) {
		// nolint:exhaustive
		switch up.Type {
		case activitygrpc.UpActReserveRelationType_Archive: // 稿件
			if up.LivePlanStartTime.Time().Unix() > 0 {
				common.DescText_1 = &api.HighlightText{
					Text: fmt.Sprintf("预计%s发布", model.UpPubDataString(up.LivePlanStartTime.Time())),
				}
			}
		case activitygrpc.UpActReserveRelationType_Live: // 直播
			if up.LivePlanStartTime.Time().Unix() > 0 {
				common.DescText_1 = &api.HighlightText{
					Text: fmt.Sprintf("%s直播", model.UpPubDataString(up.LivePlanStartTime.Time())),
				}
			}
		case activitygrpc.UpActReserveRelationType_ESports: // 赛事
			if up.Desc != "" {
				common.DescText_1 = &api.HighlightText{
					Text: up.Desc,
				}
			}
		}
	} else {
		// nolint:exhaustive
		switch up.Type {
		case activitygrpc.UpActReserveRelationType_Archive: // 稿件
			common.DescText_1 = &api.HighlightText{
				Text: "视频预约",
			}
		case activitygrpc.UpActReserveRelationType_Live: // 直播
			common.DescText_1 = &api.HighlightText{
				Text: model.UpPubDataString(up.LivePlanStartTime.Time()) + " 进行直播",
			}
		}
	}
	// 主人态
	if up.Upmid == general.Mid {
		common.Button = &api.AdditionalButton{
			Type:    api.AddButtonType_bt_button,
			Status:  api.AdditionalButtonStatus_uncheck,
			Uncheck: upButtonCheck("", _upbuttonCancel, api.AddButtonBgStyle_stroke, api.DisableState_highlight),
			Check:   upButtonCheck("UP主已撤销预约", _upbuttonCancelOk, api.AddButtonBgStyle_fill, api.DisableState_gary), // 置灰不可点击
		}
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.LotteryTypeCron, &feature.OriginResutl{
			MobiApp: general.GetMobiApp(),
			Device:  general.GetDevice(),
			Build:   general.GetBuild(),
			BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.LotteryTypeCronIOS) ||
				(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.LotteryTypeCronAndroid) ||
				(general.IsPad() && general.GetBuild() >= s.c.BuildLimit.DynReservePadIOSPad) ||
				(general.IsPadHD() && general.GetBuild() >= s.c.BuildLimit.DynReservePadHD) ||
				(general.IsAndroidHD() && general.GetBuild() > s.c.BuildLimit.DynReservePadAndroid)}) {
			if up.LotteryType == activitygrpc.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron {
				common.Button.Uncheck = upButtonCheck("", _upbuttonCancelLotteryCron, api.AddButtonBgStyle_stroke, api.DisableState_highlight)
			}
		}
		common.Button.Uncheck.Share, common.BusinessId, common.DynType = s.additionalButtonShare(c, general, dynCtx, up)
	} else if up.Upmid != general.Mid { // 客人态
		common.Button = &api.AdditionalButton{
			Type:    api.AddButtonType_bt_button,
			Status:  api.AdditionalButtonStatus_uncheck,
			Uncheck: upButtonCheck("", _upbuttonReservation, api.AddButtonBgStyle_fill, api.DisableState_gary),
		}
	}
	// 仅主态可见
	if up.UpActVisible == activitygrpc.UpActVisible_OnlyUpVisible {
		if up.Upmid != general.Mid {
			return nil, true
		}
		common.Title = "[审核中]" + common.Title
	}
	// nolint:exhaustive
	switch upActState(up.State) {
	case _upAudit:
		// 当前是预约先审后发且当前用户不是主态，不下发预约卡
		if upActState(up.State) == _upAudit && up.Upmid != general.Mid {
			return nil, true
		}
		common.Title = "[审核中]" + common.Title
	case _upOnline:
		// nolint:exhaustive
		switch up.Type {
		case activitygrpc.UpActReserveRelationType_Archive: // 稿件
			aid, _ := strconv.ParseInt(up.Oid, 10, 64)
			ap, ok := dynCtx.GetArchive(aid)
			common.Url = model.FillURI(model.GotoAv, strconv.FormatInt(aid, 10), nil) // 不要秒开，且稿件不存在还是返回url，进入详情页展示稿件不存在
			if !ok {
				if up.IsFollow == 1 && up.Upmid != general.Mid {
					common.Button = &api.AdditionalButton{
						Type:   api.AddButtonType_bt_button,
						Status: api.AdditionalButtonStatus_check,
						Check:  upButtonCheck("不在预约时间", _upbuttonReservationOk, api.AddButtonBgStyle_fill, api.DisableState_gary),
					}
				} else {
					common.Button = &api.AdditionalButton{
						Type:   api.AddButtonType_bt_button,
						Status: api.AdditionalButtonStatus_check,
						Check:  upButtonCheck("不在预约时间", _upbuttonReservation, api.AddButtonBgStyle_fill, api.DisableState_gary),
					}
				}
			} else {
				var arc = ap.Arc
				common.DescText_2 = model.UpStatString(int64(arc.Stat.View), "观看")
				common.Button = &api.AdditionalButton{
					Type:      api.AddButtonType_bt_jump,
					Status:    api.AdditionalButtonStatus_uncheck,
					JumpUrl:   model.FillURI(model.GotoAv, strconv.FormatInt(arc.Aid, 10), model.AvPlayHandlerGRPCV2(ap, arc.FirstCid, true)),
					JumpStyle: upButtonCheck("", _upbuttonWatch, api.AddButtonBgStyle_fill, api.DisableState_highlight),
				}
			}
		case activitygrpc.UpActReserveRelationType_Live: // 直播
			var isok bool
			if info, ok := dynCtx.ResLiveSessionInfo[up.Oid]; ok {
				live, ok := info.SessionInfoPerLive[up.Oid]
				if ok {
					isok = true
					switch live.Status {
					case _liveStart:
						common.Url = model.FillURI(model.GotoLive, strconv.FormatInt(info.RoomId, 10), nil)
						if liveUrl, ok := info.JumpUrl["dt_booking_dt"]; ok {
							common.Url = liveUrl
						}
						common.DescText_2 = model.UpStatString(live.PopularityCount, "人气")
						if show := live.WatchedShow; show != nil && show.TextLarge != "" {
							common.DescText_2 = show.TextLarge
						}
						common.Button = &api.AdditionalButton{
							Type:      api.AddButtonType_bt_jump,
							Status:    api.AdditionalButtonStatus_uncheck,
							JumpUrl:   common.Url,
							JumpStyle: upButtonCheck("", _upbuttonWatch, api.AddButtonBgStyle_fill, api.DisableState_highlight),
						}
						common.DescText_1 = &api.HighlightText{
							Text:      "直播中",
							TextStyle: api.HighlightTextStyle_style_highlight,
						}
					case _liveAv:
						aid, _ := bvid.BvToAv(live.Bvid)
						common.Url = model.FillURI(model.GotoAv, strconv.FormatInt(aid, 10), nil)
						common.Button = &api.AdditionalButton{
							Type:      api.AddButtonType_bt_jump,
							Status:    api.AdditionalButtonStatus_uncheck,
							JumpUrl:   common.Url,
							JumpStyle: upButtonCheck("", _upbuttonReplay, api.AddButtonBgStyle_fill, api.DisableState_highlight),
						}
						_, ok := dynCtx.ResArcs[aid]
						if !ok {
							isok = false
						}
					default:
						isok = false
					}
				}
			}
			if !isok {
				// 已结束（不可回放）
				common.Button = &api.AdditionalButton{
					Type:   api.AddButtonType_bt_button,
					Status: api.AdditionalButtonStatus_check,
					Check:  upButtonCheck("直播已结束", _upbuttonEnd, api.AddButtonBgStyle_fill, api.DisableState_gary),
				}
			}
		case activitygrpc.UpActReserveRelationType_ESports:
			// 赛事核销
			if up.State == activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackDone {
				common.Button = &api.AdditionalButton{
					Type:   api.AddButtonType_bt_button,
					Status: api.AdditionalButtonStatus_check,
					Check:  upButtonCheck("", _upbuttonEnd, api.AddButtonBgStyle_fill, api.DisableState_gary),
				}
			}
		}
	case _upExpired:
		// nolint:exhaustive
		switch up.Type {
		case activitygrpc.UpActReserveRelationType_Live: // 直播
			// 客人 否则 主人
			if up.Upmid != general.Mid {
				if up.IsFollow == 1 {
					roomid, _ := strconv.ParseInt(up.Oid, 10, 64)
					common.Url = model.FillURI(model.GotoLive, strconv.FormatInt(roomid, 10), nil) // 开播未开播都返回url
					common.Button = &api.AdditionalButton{
						Type:   api.AddButtonType_bt_button,
						Status: api.AdditionalButtonStatus_check,
						Check:  upButtonCheck("预约已过期", _upbuttonReservationOk, api.AddButtonBgStyle_fill, api.DisableState_gary),
					}
				} else {
					common.Button = &api.AdditionalButton{
						Type:   api.AddButtonType_bt_button,
						Status: api.AdditionalButtonStatus_check,
						Check:  upButtonCheck("预约已过期", _upbuttonReservation, api.AddButtonBgStyle_fill, api.DisableState_gary),
					}
				}
			} else {
				common.Button = &api.AdditionalButton{
					Type:   api.AddButtonType_bt_button,
					Status: api.AdditionalButtonStatus_check,
					Check:  upButtonCheck("预约已过期", _upbuttonCancelOk, api.AddButtonBgStyle_fill, api.DisableState_gary),
				}
			}
		}
	case _upCancel:
		common.Button = &api.AdditionalButton{
			Type:   api.AddButtonType_bt_button,
			Status: api.AdditionalButtonStatus_check,
			Check:  upButtonCheck("UP主已撤销预约", _upbuttonCancelOk, api.AddButtonBgStyle_fill, api.DisableState_gary),
		}
		// 撤销不展示抽奖信息
		common.LotteryType = api.ReserveRelationLotteryType_reserve_relation_lottery_type_default
		common.DescText_3 = nil
	case _upDelete:
		// 预约卡被删除
		return nil, false
	}
	if common.Button != nil {
		common.Button.ClickType = api.AdditionalButtonClickType_click_up
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_up_reservation,
		Item: &api.ModuleAdditional_Up{
			Up: common,
		},
	}
	return additional, true
}

// UP预约状态判断
func upActState(state activitygrpc.UpActReserveRelationState) int {
	// nolint:exhaustive
	switch state {
	case activitygrpc.UpActReserveRelationState_UpReserveAudit:
		return _upAudit
	case activitygrpc.UpActReserveRelationState_UpReserveRelated, activitygrpc.UpActReserveRelationState_UpReserveRelatedAudit, activitygrpc.UpActReserveRelationState_UpReserveRelatedOnline:
		return _upStart
	case activitygrpc.UpActReserveRelationState_UpReserveRelatedWaitCallBack, activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackCancel, activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackDone:
		return _upOnline
	case activitygrpc.UpActReserveRelationState_UpReserveReject:
		return _upDelete
	case activitygrpc.UpActReserveRelationState_UpReserveCancel:
		return _upCancel
	case activitygrpc.UpActReserveRelationState_UpReserveCancelExpired:
		return _upExpired
	}
	return _upNotStart
}

// 附加卡被删除
func (s *Service) additionalNull(text string) *api.Module {
	return &api.Module{
		ModuleType: api.DynModuleType_module_item_null,
		ModuleItem: &api.Module_ModuleItemNull{
			ModuleItemNull: &api.ModuleItemNull{
				Icon: s.c.Resource.Icon.ModuleDynamicItemNull,
				Text: text,
			},
		},
	}
}

func upButtonCheck(toast string, state int, bgStyle api.AddButtonBgStyle, disable api.DisableState) *api.AdditionalButtonStyle {
	check := &api.AdditionalButtonStyle{
		BgStyle: bgStyle,
		Disable: disable,
		Toast:   toast,
	}
	switch state {
	case _upbuttonReservation: // 预约
		check.Icon = "http://i0.hdslb.com/bfs/archive/f5b7dae25cce338e339a655ac0e4a7d20d66145c.png"
		check.Text = "预约"
	case _upbuttonReservationOk: // 已预约
		check.Text = "已预约"
	case _upbuttonCancel: // 取消预约
		check.Text = "撤销"
		check.Interactive = &api.AdditionalButtonInteractive{
			Popups:  "撤销预约后，将提醒已预约用户",
			Confirm: "撤销预约",
			Cancel:  "取消",
		}
	case _upbuttonCancelLotteryCron:
		check.Text = "撤销"
		check.Interactive = &api.AdditionalButtonInteractive{
			Popups:  "撤销预约将提醒已预约用户",
			Desc:    "撤销本次预约后会关闭抽奖",
			Confirm: "撤销预约",
			Cancel:  "取消",
		}
	case _upbuttonCancelOk: // 已取消
		check.Text = "已撤销"
	case _upbuttonWatch: // 去观看
		check.Text = "去观看"
	case _upbuttonReplay: // 回放
		check.Text = "看回放"
	case _upbuttonEnd: // 已结束
		check.Text = "已结束"
	default:
		return nil
	}
	return check
}

func (s *Service) additionalButtonShare(c context.Context, general *mdlv2.GeneralParam, dynCtx *mdlv2.DynamicContext, up *activitygrpc.UpActReserveRelationInfo) (share *api.AdditionalButtonShare, businessID string, dynType int64) {
	if general.IsPad() || general.IsPadHD() || general.IsAndroidHD() || general.IsOverseas() {
		return nil, "", 0
	}
	// 首映
	if up.Type == activitygrpc.UpActReserveRelationType_Premiere {
		return nil, "", 0
	}
	dynID, _ := strconv.ParseInt(up.DynamicId, 10, 64)
	dynInfo, dynOk := dynCtx.ResDynSimpleInfos[dynID]
	if dynOk && dynInfo.Visible {
		businessID = strconv.FormatInt(dynInfo.Rid, 10)
		dynType = dynInfo.Type
	}
	if dynCtx.From == _handleTypeReservePersonal {
		// 老版本不下发预约分享
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynReservePersonalShare, &feature.OriginResutl{
			BuildLimit: (general.IsIPhonePick() && general.GetBuild() < s.c.BuildLimit.DynReservePersonalShareIOS) ||
				(general.IsAndroidPick() && general.GetBuild() <= s.c.BuildLimit.DynReservePersonalShareAndroid)}) {
			return nil, "", 0
		}
		if !dynOk || !dynInfo.Visible {
			return nil, "", 0
		}
	}
	return &api.AdditionalButtonShare{
		Icon: s.c.Resource.Icon.AdditionalButtonShareIcon,
		Text: s.c.Resource.Text.AdditionalButtonShareText,
		Show: api.AdditionalShareShowType_st_show,
	}, businessID, dynType
}

func (s *Service) additionalFeedCardDrama(res *dramaseasongrpc.FeedCardDramaInfo, _ *mdlv2.GeneralParam, dynCtx *mdlv2.DynamicContext) *api.ModuleAdditional {
	common := &api.AdditionCommon{
		HeadText:   s.c.Resource.Text.AdditionalFeedCardDramaHeadText,
		Title:      res.Title,
		ImageUrl:   res.Cover,
		DescText_1: res.LastOrdStr,
		DescText_2: res.Tag,
		Url:        res.AppJumpUrl,
		Style:      api.ImageStyle_add_style_square,
		Type:       strconv.FormatInt(dynCtx.Dyn.Type, 10),
		CardType:   "maoer_drama",
	}
	common.Button = &api.AdditionalButton{
		Type:    api.AddButtonType_bt_jump,
		JumpUrl: res.AppJumpUrl,
		JumpStyle: &api.AdditionalButtonStyle{
			Text: "去收听",
		},
	}
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_common,
		Item: &api.ModuleAdditional_Common{
			Common: common,
		},
	}
	return additional
}

func (s *Service) additionalShopping(res *shopping.Item, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) *api.ModuleAdditional {
	good := &api.AdditionGoods{
		CardType:   "good",
		Icon:       s.c.Resource.Icon.ModuleAdditionalGoods,
		Uri:        dynCtx.DynamicItem.Extend.CardUrl,
		AdMarkIcon: s.c.Resource.Icon.AdditionalAdMarkIcon,
		RcmdDesc:   "UP主的推荐",
	}
	tmp := &api.GoodsItem{
		Cover:      res.Img,
		SourceType: 2,
		JumpUrl:    res.URL,
		JumpDesc:   "去看看",
		Title:      res.Name,
		Price:      fmt.Sprintf("¥%s", res.Price),
		ItemId:     res.ID,
		JumpType:   api.GoodsJumpType_goods_url,
	}
	good.GoodsItems = append(good.GoodsItems, tmp)
	good.SourceType = tmp.SourceType
	good.JumpType = tmp.JumpType
	good.AppName = tmp.AppName
	additional := &api.ModuleAdditional{
		Type: api.AdditionalType_additional_type_goods,
		Item: &api.ModuleAdditional_Goods{
			Goods: good,
		},
	}
	return additional
}
