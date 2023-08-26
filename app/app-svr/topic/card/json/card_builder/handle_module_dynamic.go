package cardbuilder

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/log"
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	bcgmdl "go-gateway/app/app-svr/app-dynamic/interface/model/bcg"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	"go-gateway/pkg/idsafe/bvid"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
)

func handleModuleDynamic(metaCtx jsonwebcard.MetaContext, cardType jsonwebcard.CardType, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.ModuleDynamic {
	res := &jsonwebcard.ModuleDynamic{
		Desc:       handleDynamicDesc(dynCtx),
		Major:      handleDynamicMajor(cardType, dynCtx),
		Additional: handleDynamicAdditional(metaCtx, dynCtx),
		Topic:      nil,
	}
	return res
}

func handleDynamicAdditional(metaCtx jsonwebcard.MetaContext, dynCtx *dynmdlV2.DynamicContext) jsonwebcard.DynAdditional {
	if dynCtx.Dyn.IsForward() && (dynCtx.Interim.ForwardOrigFaild || dynCtx.Interim.IsPassAddition) {
		return nil
	}
	for _, v := range dynCtx.Dyn.AttachCardInfos {
		switch v.CardType {
		case dyncommongrpc.AttachCardType_ATTACH_CARD_VOTE:
			if res, ok := dynCtx.ResVote[v.Rid]; ok {
				return additionalVote(res)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:
			if res, ok := dynCtx.ResUpActRelationInfo[v.Rid]; ok {
				if reserveCard, ok := additionalReserve(dynCtx, res, res.Upmid == metaCtx.Mid); ok {
					return reserveCard
				}
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_GOODS:
			resGoods, ok := dynCtx.ResGood[dynCtx.Dyn.DynamicID]
			if !ok {
				continue
			}
			if res, ok := resGoods[bcgmdl.GoodsLocTypeCard]; ok {
				if goodsCard, ok := additionalGoods(res, dynCtx); ok {
					return goodsCard
				}
			}
		default:
			log.Warn("module error dynid(%v) additional unknown type %+v", dynCtx.Dyn.DynamicID, v.CardType)
			continue
		}
	}
	return nil
}

func additionalGoods(res map[string]*bcgmdl.GoodsItem, dynCtx *dynmdlV2.DynamicContext) (jsonwebcard.DynAdditional, bool) {
	goods := &jsonwebcard.Goods{
		HeadText: "UP主的推荐",
		HeadIcon: "https://i0.hdslb.com/bfs/feed-admin/3ac25959e29285fa56c378844a978841661adf78.png",
	}
	for _, id := range strings.Split(dynCtx.Dyn.Extend.OpenGoods.ItemsId, ",") {
		goodsItem, ok := res[id]
		if !ok {
			continue
		}
		tmp := &jsonwebcard.GoodsItem{
			Cover:    goodsItem.Img,
			Name:     goodsItem.Name,
			Brief:    goodsItem.Brief,
			Price:    goodsItem.PriceStr,
			JumpUrl:  goodsItem.JumpLink,
			JumpDesc: goodsItem.JumpLinkDesc,
			Id:       goodsItem.ItemsID,
		}
		goods.Items = append(goods.Items, tmp)
	}
	if len(goods.Items) > 0 {
		additional := jsonwebcard.AdditionalGoods{
			AdditionalType: jsonwebcard.AdditionalTypeGoods,
			Goods:          goods,
		}
		return additional, true
	}
	return nil, false
}

func additionalReserve(dynCtx *dynmdlV2.DynamicContext, relationInfo *activitygrpc.UpActReserveRelationInfo, isOwner bool) (jsonwebcard.DynAdditional, bool) {
	const (
		_liveStart = 1
		_liveAv    = 2
	)
	res := jsonwebcard.AdditionalReserve{
		AdditionalType: jsonwebcard.AdditionalTypeReserve,
		Reserve: &jsonwebcard.Reserve{
			Title:         constructReserveTitle(relationInfo, isOwner),
			Desc1:         constructReserveDesc1(relationInfo),
			Desc2:         constructReserveDesc2(relationInfo, isOwner),
			Desc3:         constructReserveDesc3(relationInfo),
			ReserveButton: &jsonwebcard.ReserveButton{},
			Rid:           relationInfo.Sid,
			ReserveTotal:  relationInfo.Total,
			State:         int64(relationInfo.State),
			Stype:         int64(relationInfo.Type),
			UpMid:         relationInfo.Upmid,
		},
	}

	// 主人态
	if isOwner {
		res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
			Type:    int64(dynamicapi.AddButtonType_bt_button),
			Status:  int64(dynamicapi.AdditionalButtonStatus_uncheck),
			UnCheck: upButtonCheck("", topiccardmodel.UpbuttonCancel, dynamicapi.DisableState_highlight),
			Check:   upButtonCheck("UP主已撤销预约", topiccardmodel.UpbuttonCancelOk, dynamicapi.DisableState_gary), // 置灰不可点击
		}
		if relationInfo.LotteryType == activitygrpc.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron {
			res.Reserve.ReserveButton.UnCheck = upButtonCheck("", topiccardmodel.UpbuttonCancelLotteryCron, dynamicapi.DisableState_highlight)
		}
	} else { // 客人态
		if relationInfo.IsFollow == 1 {
			res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
				Type:    int64(dynamicapi.AddButtonType_bt_button),
				Status:  int64(dynamicapi.AdditionalButtonStatus_check),
				Check:   upButtonCheck("", topiccardmodel.UpbuttonReservationOk, dynamicapi.DisableState_highlight),
				UnCheck: upButtonCheck("", topiccardmodel.UpbuttonReservation, dynamicapi.DisableState_highlight),
			}
		} else {
			res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
				Type:    int64(dynamicapi.AddButtonType_bt_button),
				Status:  int64(dynamicapi.AdditionalButtonStatus_uncheck),
				UnCheck: upButtonCheck("", topiccardmodel.UpbuttonReservation, dynamicapi.DisableState_highlight),
				Check:   upButtonCheck("", topiccardmodel.UpbuttonReservationOk, dynamicapi.DisableState_highlight),
			}
		}
		if relationInfo.LotteryType == activitygrpc.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron {
			res.Reserve.ReserveButton.UnCheck.Toast = "预约成功，已参与抽奖"
		}
	}
	// nolint:exhaustive
	switch relationInfo.State {
	case activitygrpc.UpActReserveRelationState_UpReserveRelatedWaitCallBack, activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackCancel, activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackDone:
		// nolint:exhaustive
		switch relationInfo.Type {
		case activitygrpc.UpActReserveRelationType_Archive: // 稿件
			aid, _ := strconv.ParseInt(relationInfo.Oid, 10, 64)
			ap, ok := dynCtx.GetArchive(aid)
			if !ok {
				if relationInfo.IsFollow == 1 && !isOwner {
					res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
						Type:   int64(dynamicapi.AddButtonType_bt_button),
						Status: int64(dynamicapi.AdditionalButtonStatus_check),
						Check:  upButtonCheck("不在预约时间", topiccardmodel.UpbuttonReservationOk, dynamicapi.DisableState_gary),
					}
				} else {
					res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
						Type:   int64(dynamicapi.AddButtonType_bt_button),
						Status: int64(dynamicapi.AdditionalButtonStatus_check),
						Check:  upButtonCheck("不在预约时间", topiccardmodel.UpbuttonReservation, dynamicapi.DisableState_gary),
					}
				}
			} else {
				var arc = ap.Arc
				arcBvid, _ := bvid.AvToBv(arc.Aid)
				res.Reserve.Desc2.Text = model.UpStatString(int64(arc.Stat.View), "观看")
				res.Reserve.JumpUrl = topiccardmodel.FillURI(topiccardmodel.GotoWebAv, arcBvid, topiccardmodel.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, true))
				res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
					Type:      int64(dynamicapi.AddButtonType_bt_jump),
					Status:    int64(dynamicapi.AdditionalButtonStatus_uncheck),
					JumpStyle: upButtonCheck("", topiccardmodel.UpbuttonWatch, dynamicapi.DisableState_highlight),
				}
			}
		case activitygrpc.UpActReserveRelationType_Live: // 直播
			var isok bool
			if info, ok := dynCtx.ResLiveSessionInfo[relationInfo.Oid]; ok {
				liveInfo, ok := info.SessionInfoPerLive[relationInfo.Oid]
				if ok {
					isok = true
					switch liveInfo.Status {
					case _liveStart:
						res.Reserve.JumpUrl = model.FillURI(model.GotoLive, strconv.FormatInt(info.RoomId, 10), nil)
						if liveUrl, ok := info.JumpUrl["dt_booking_dt"]; ok {
							res.Reserve.JumpUrl = liveUrl
						}
						res.Reserve.Desc2.Text = makeReserveLiveStartDesc2(liveInfo)
						res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
							Type:      int64(dynamicapi.AddButtonType_bt_jump),
							Status:    int64(dynamicapi.AdditionalButtonStatus_uncheck),
							JumpUrl:   res.Reserve.JumpUrl,
							JumpStyle: upButtonCheck("", topiccardmodel.UpbuttonWatch, dynamicapi.DisableState_highlight),
						}
						res.Reserve.Desc1 = &jsonwebcard.Desc1{
							Text:  "直播中",
							Style: int64(dynamicapi.HighlightTextStyle_style_highlight),
						}
					case _liveAv:
						aid, _ := bvid.BvToAv(liveInfo.Bvid)
						res.Reserve.JumpUrl = topiccardmodel.FillURI(topiccardmodel.GotoWebAv, liveInfo.Bvid, nil)
						res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
							Type:      int64(dynamicapi.AddButtonType_bt_jump),
							Status:    int64(dynamicapi.AdditionalButtonStatus_uncheck),
							JumpUrl:   res.Reserve.JumpUrl,
							JumpStyle: upButtonCheck("", topiccardmodel.UpbuttonReplay, dynamicapi.DisableState_highlight),
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
				res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
					Type:   int64(dynamicapi.AddButtonType_bt_button),
					Status: int64(dynamicapi.AdditionalButtonStatus_check),
					Check:  upButtonCheck("直播已结束", topiccardmodel.UpbuttonEnd, dynamicapi.DisableState_gary),
				}
			}
		case activitygrpc.UpActReserveRelationType_ESports:
			// 赛事核销
			if relationInfo.State == activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackDone {
				res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
					Type:   int64(dynamicapi.AddButtonType_bt_button),
					Status: int64(dynamicapi.AdditionalButtonStatus_check),
					Check:  upButtonCheck("已结束", topiccardmodel.UpbuttonEnd, dynamicapi.DisableState_gary),
				}
			}
		}
	case activitygrpc.UpActReserveRelationState_UpReserveCancelExpired:
		// nolint:exhaustive
		switch relationInfo.Type {
		case activitygrpc.UpActReserveRelationType_Live: // 直播
			// 客人 否则 主人
			if !isOwner {
				if relationInfo.IsFollow == 1 {
					roomid, _ := strconv.ParseInt(relationInfo.Oid, 10, 64)
					res.Reserve.JumpUrl = model.FillURI(model.GotoLive, strconv.FormatInt(roomid, 10), nil) // 开播未开播都返回url
					res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
						Type:   int64(dynamicapi.AddButtonType_bt_button),
						Status: int64(dynamicapi.AdditionalButtonStatus_check),
						Check:  upButtonCheck("预约已过期", topiccardmodel.UpbuttonReservationOk, dynamicapi.DisableState_gary),
					}
				} else {
					res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
						Type:   int64(dynamicapi.AddButtonType_bt_button),
						Status: int64(dynamicapi.AdditionalButtonStatus_check),
						Check:  upButtonCheck("预约已过期", topiccardmodel.UpbuttonReservation, dynamicapi.DisableState_gary),
					}
				}
			} else {
				res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
					Type:   int64(dynamicapi.AddButtonType_bt_button),
					Status: int64(dynamicapi.AdditionalButtonStatus_check),
					Check:  upButtonCheck("预约已过期", topiccardmodel.UpbuttonCancelOk, dynamicapi.DisableState_gary),
				}
			}
		}
	case activitygrpc.UpActReserveRelationState_UpReserveCancel:
		res.Reserve.ReserveButton = &jsonwebcard.ReserveButton{
			Type:   int64(dynamicapi.AddButtonType_bt_button),
			Status: int64(dynamicapi.AdditionalButtonStatus_check),
			Check:  upButtonCheck("UP主已撤销预约", topiccardmodel.UpbuttonCancelOk, dynamicapi.DisableState_gary),
		}
	case activitygrpc.UpActReserveRelationState_UpReserveReject:
		// 预约卡被删除
		return jsonwebcard.NewAdditionalReserveNull(), true
	}
	return res, true
}

func makeReserveLiveStartDesc2(liveInfo *livexroomgate.SessionInfoPerLive) string {
	if show := liveInfo.WatchedShow; show != nil && show.TextLarge != "" {
		return show.TextLarge
	}
	return model.UpStatString(liveInfo.PopularityCount, "人气")
}

func constructReserveDesc3(relationInfo *activitygrpc.UpActReserveRelationInfo) *jsonwebcard.Desc3 {
	// 抽奖 高亮icon跳链文案
	if relationInfo.PrizeInfo != nil {
		return &jsonwebcard.Desc3{
			Text:    relationInfo.PrizeInfo.Text,
			JumpUrl: relationInfo.PrizeInfo.JumpUrl,
			Style:   int64(dynamicapi.HighlightTextStyle_style_highlight),
			IconUrl: "https://i0.hdslb.com/bfs/feed-admin/0c62d6a31f560c3942a787ab5220b048458ab397.png",
		}
	}
	return nil
}

func constructReserveDesc2(relationInfo *activitygrpc.UpActReserveRelationInfo, isOwner bool) *jsonwebcard.Desc2 {
	if !isOwner && relationInfo.Total < relationInfo.ReserveTotalShowLimit {
		return &jsonwebcard.Desc2{
			Visible: false,
		}
	}
	return &jsonwebcard.Desc2{
		Text:    topiccardmodel.StatString(relationInfo.Total, "人预约", "0人预约"),
		Visible: true,
	}
}

func constructReserveDesc1(up *activitygrpc.UpActReserveRelationInfo) *jsonwebcard.Desc1 {
	switch up.Type {
	case activitygrpc.UpActReserveRelationType_Archive: // 稿件
		if up.LivePlanStartTime.Time().Unix() > 0 {
			return &jsonwebcard.Desc1{
				Text: fmt.Sprintf("预计%s发布", model.UpPubDataString(up.LivePlanStartTime.Time())),
			}
		}
	case activitygrpc.UpActReserveRelationType_Live: // 直播
		if up.LivePlanStartTime.Time().Unix() > 0 {
			return &jsonwebcard.Desc1{
				Text: fmt.Sprintf("%s直播", model.UpPubDataString(up.LivePlanStartTime.Time())),
			}
		}
	case activitygrpc.UpActReserveRelationType_ESports: // 赛事
		if up.Desc != "" {
			return &jsonwebcard.Desc1{
				Text: up.Desc,
			}
		}
	default:
	}
	return nil
}

func upButtonCheck(toast string, state int, disable dynamicapi.DisableState) *jsonwebcard.ReserveCheck {
	check := &jsonwebcard.ReserveCheck{
		Disable: int64(disable),
		Toast:   toast,
	}
	switch state {
	case topiccardmodel.UpbuttonReservation:
		check.IconUrl = "https://i0.hdslb.com/bfs/archive/f5b7dae25cce338e339a655ac0e4a7d20d66145c.png"
		check.Text = "预约"
	case topiccardmodel.UpbuttonReservationOk:
		check.Text = "已预约"
	case topiccardmodel.UpbuttonCancel:
		check.Text = "撤销"
	case topiccardmodel.UpbuttonCancelLotteryCron:
		check.Text = "撤销"
	case topiccardmodel.UpbuttonCancelOk:
		check.Text = "已撤销"
	case topiccardmodel.UpbuttonWatch:
		check.Text = "去观看"
	case topiccardmodel.UpbuttonReplay:
		check.Text = "看回放"
	case topiccardmodel.UpbuttonEnd:
		check.Text = "已结束"
	default:
		return nil
	}
	return check
}

func constructReserveTitle(relationInfo *activitygrpc.UpActReserveRelationInfo, isOwner bool) string {
	if isOwner && relationInfo.UpActVisible == activitygrpc.UpActVisible_OnlyUpVisible {
		return "[审核中]" + relationInfo.Title
	}
	return relationInfo.Title
}

func additionalVote(res *dyncommongrpc.VoteInfo) jsonwebcard.DynAdditional {
	return jsonwebcard.AdditionalVote{
		AdditionalType: jsonwebcard.AdditionalTypeVote,
		Vote: &jsonwebcard.Vote{
			VoteId:       res.VoteId,
			Title:        res.Title,
			ChoiceCnt:    res.ChoiceCnt,
			DefaultShare: res.DefaultShare,
			Desc:         res.Title, // web显示投票的标题字段取desc
			EndTime:      res.EndTime,
			JoinNum:      res.JoinNum,
			Status:       res.Status,
			Type:         res.Type,
			Uid:          res.VotePublisher,
		},
	}
}

func handleDynamicMajor(cardType jsonwebcard.CardType, dynCtx *dynmdlV2.DynamicContext) jsonwebcard.DynMajor {
	switch cardType {
	case jsonwebcard.CardDynamicTypeAv:
		return handleDynArchiveMajor(dynCtx)
	case jsonwebcard.CardDynamicTypeDraw:
		return handleDynDrawMajor(dynCtx)
	case jsonwebcard.CardDynamicTypeForward, jsonwebcard.CardDynamicTypeWord:
		return nil
	case jsonwebcard.CardDynamicTypeArticle:
		return handleDynArticleMajor(dynCtx)
	case jsonwebcard.CardDynamicTypeCommon:
		return handleDynCommonMajor(dynCtx)
	case jsonwebcard.CardDynamicTypePGC:
		return handleDynPGCMajor(dynCtx)
	default:
		log.Warn("unexpected handleDynamicMajor type=%+v", cardType)
		return nil
	}
}

func handleDynPGCMajor(dynCtx *dynmdlV2.DynamicContext) jsonwebcard.DynMajor {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid))
	if !ok {
		return nil
	}
	return jsonwebcard.MajorPGC{
		MajorType: jsonwebcard.MajorTypePGC,
		PGC: &jsonwebcard.PGC{
			Type:     int64(dynamicapi.MediaType_MediaTypePGC),
			SubType:  int64(dynCtx.Dyn.GetPGCSubType()),
			SeasonId: int64(pgc.Season.SeasonId),
			EpId:     int64(pgc.EpisodeId),
			Title:    pgc.CardShowTitle,
			Cover:    pgc.Cover,
			Badge:    constructPGCMajorBadge(pgc.Stat.FollowDesc, pgc.Season.TypeName),
			JumpUrl:  pgc.Url,
			VideoStat: jsonwebcard.VideoStat{
				Danmaku: topiccardmodel.StatString(pgc.Stat.Danmaku, "", ""),
				Play:    topiccardmodel.StatString(pgc.Stat.Play, "", ""),
			},
		},
	}
}

func constructPGCMajorBadge(followDesc, typeName string) jsonwebcard.Badge {
	if followDesc != "" {
		return jsonwebcard.Badge{
			Text:    followDesc,
			BgColor: "#FFFFFF",
			Color:   "#7F000000",
		}
	}
	if typeName != "" {
		return jsonwebcard.Badge{
			Text:    typeName,
			BgColor: "#FFFFFF",
			Color:   "#FB7299",
		}
	}
	return jsonwebcard.Badge{}
}

func handleDynCommonMajor(dynCtx *dynmdlV2.DynamicContext) jsonwebcard.DynMajor {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	common, ok := dynCtx.GetResCommon(dynCtx.Dyn.Rid)
	if !ok {
		return nil
	}
	return jsonwebcard.MajorCommon{
		MajorType: jsonwebcard.MajorTypeCommon,
		Common: &jsonwebcard.Common{
			ID:       common.RID,
			JumpUrl:  common.Sketch.TagURL,
			Cover:    common.Sketch.CoverURL,
			Title:    common.Sketch.Title,
			Desc:     common.Sketch.DescText,
			Label:    common.Sketch.Text,
			Biz:      int64(common.Sketch.BizType),
			SketchId: common.Sketch.SketchID,
			Badge:    constructCommonMajorBadge(common.Sketch.Tags),
			Style:    int64(constructCommonMajorStyle(dynCtx)),
		},
	}
}

func constructCommonMajorStyle(dynCtx *dynmdlV2.DynamicContext) dynamicapi.MdlDynCommonType {
	switch {
	case dynCtx.Dyn.IsCommonSquare():
		return dynamicapi.MdlDynCommonType_mdl_dyn_common_square
	case dynCtx.Dyn.IsCommonVertical():
		return dynamicapi.MdlDynCommonType_mdl_dyn_common_vertica
	default:
		return dynamicapi.MdlDynCommonType_mdl_dyn_common_none
	}
}

func constructCommonMajorBadge(in json.RawMessage) jsonwebcard.Badge {
	var tags []*dynmdlV2.DynamicCommonCardTags
	if err := json.Unmarshal(in, &tags); err != nil {
		return jsonwebcard.Badge{}
	}
	for _, tag := range tags {
		if tag == nil || tag.Name == "" {
			continue
		}
		if !strings.Contains(tag.Color, "#") {
			tag.Color = fmt.Sprintf("#%s", tag.Color)
		}
		return jsonwebcard.Badge{
			Text:    tag.Name,
			BgColor: tag.Color,
			Color:   "#FFFFFF",
		}
	}
	return jsonwebcard.Badge{}
}

func handleDynArticleMajor(dynCtx *dynmdlV2.DynamicContext) jsonwebcard.DynMajor {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	article, ok := dynCtx.GetResArticle(dynCtx.Dyn.Rid)
	if !ok {
		return nil
	}
	return jsonwebcard.MajorArticle{
		MajorType: jsonwebcard.MajorTypeArticle,
		Article: &jsonwebcard.Article{
			ID:      article.ActID,
			Title:   article.Title,
			Desc:    article.Summary,
			JumpUrl: fmt.Sprintf("//www.bilibili.com/read/cv%d", article.ID),
			Covers:  article.ImageURLs,
			Label:   topiccardmodel.StatString(article.Stats.View, "阅读", ""),
		},
	}
}

func handleDynDrawMajor(dynCtx *dynmdlV2.DynamicContext) jsonwebcard.DynMajor {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	draw, ok := dynCtx.GetResDraw(dynCtx.Dyn.Rid)
	if !ok {
		return nil
	}
	var drawItems []*jsonwebcard.DrawItem
	for _, pic := range draw.Item.Pictures {
		i := &jsonwebcard.DrawItem{
			Src:    pic.ImgSrc,
			Width:  pic.ImgWidth,
			Height: pic.ImgHeight,
		}
		for _, picTag := range pic.ImgTags {
			switch picTag.Type {
			case dynmdlV2.DrawTagTypeCommon:
				i.DrawItemTag = append(i.DrawItemTag, &jsonwebcard.DrawItemTag{
					DrawTagType: jsonwebcard.DrawTagTypeCommon,
					JumpUrl:     picTag.Url,
					Text:        picTag.Text,
					X:           picTag.X,
					Y:           picTag.Y,
					Orientation: picTag.Orientation,
					Mid:         picTag.Mid,
				})
			case dynmdlV2.DrawTagTypeGoods:
				i.DrawItemTag = append(i.DrawItemTag, &jsonwebcard.DrawItemTag{
					DrawTagType: jsonwebcard.DrawTagTypeGoods,
					JumpUrl:     picTag.Url,
					Text:        picTag.Text,
					X:           picTag.X,
					Y:           picTag.Y,
					Orientation: picTag.Orientation,
					Source:      picTag.Source,
					Tid:         picTag.Tid,
					Mid:         picTag.Mid,
					SchemaUrl:   picTag.SchemaURL,
				})
			case dynmdlV2.DrawTagTypeUser:
				i.DrawItemTag = append(i.DrawItemTag, &jsonwebcard.DrawItemTag{
					DrawTagType: jsonwebcard.DrawTagTypeUser,
					JumpUrl:     topiccardmodel.FillURI(topiccardmodel.GotoWebSpace, strconv.FormatInt(picTag.Mid, 10), nil),
					Text:        picTag.Text,
					X:           picTag.X,
					Y:           picTag.Y,
					Orientation: picTag.Orientation,
					Mid:         picTag.Mid,
				})
			case dynmdlV2.DrawTagTypeTopic:
				var topicURL string
				topicInfos, _ := dynCtx.Dyn.GetTopicInfo()
				for _, topic := range topicInfos {
					if topic != nil && topic.TopicName == picTag.Text {
						topicURL = topic.TopicLink
						break
					}
				}
				i.DrawItemTag = append(i.DrawItemTag, &jsonwebcard.DrawItemTag{
					DrawTagType: jsonwebcard.DrawTagTypeTopic,
					JumpUrl:     topicURL,
					Text:        picTag.Text,
					X:           picTag.X,
					Y:           picTag.Y,
					Orientation: picTag.Orientation,
					Tid:         picTag.Tid,
					Mid:         picTag.Mid,
				})
			case dynmdlV2.DrawTagTypeLBS:
				var (
					lbs *dynmdlV2.DrawTagLBS
					uri string
				)
				if err := json.Unmarshal([]byte(picTag.Poi), &lbs); err != nil {
					continue
				}
				if lbs != nil && lbs.PoiInfo != nil && lbs.PoiInfo.Location != nil {
					uri = fmt.Sprintf(topiccardmodel.LBSURI, lbs.PoiInfo.Poi, lbs.PoiInfo.Type, lbs.PoiInfo.Location.Lat, lbs.PoiInfo.Location.Lng, url.QueryEscape(lbs.PoiInfo.Title), url.QueryEscape(lbs.PoiInfo.Address))
				}
				i.DrawItemTag = append(i.DrawItemTag, &jsonwebcard.DrawItemTag{
					DrawTagType: jsonwebcard.DrawTagTypeLbs,
					JumpUrl:     uri,
					Text:        picTag.Text,
					X:           picTag.X,
					Y:           picTag.Y,
					Orientation: picTag.Orientation,
					Poi:         picTag.Poi,
				})
			default:
			}
		}
		drawItems = append(drawItems, i)
	}
	return jsonwebcard.MajorDraw{
		MajorType: jsonwebcard.MajorTypeDraw,
		Draw: &jsonwebcard.Draw{
			Id:    dynCtx.Dyn.Rid,
			Items: drawItems,
		},
	}
}

func handleDynArchiveMajor(dynCtx *dynmdlV2.DynamicContext) jsonwebcard.DynMajor {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid)
	if !ok {
		return nil
	}
	var archive = ap.Arc
	res := jsonwebcard.MajorArchive{
		MajorType: jsonwebcard.MajorTypeArchive,
		Archive: &jsonwebcard.Archive{
			MediaType: jsonwebcard.MediaTypeUgc,
			Aid:       archive.Aid,
			Cover:     archive.Pic,
			VideoStat: jsonwebcard.VideoStat{
				Danmaku: topiccardmodel.StatString(int64(archive.Stat.Danmaku), "", ""),
				Play:    topiccardmodel.StatString(int64(archive.Stat.View), "", ""),
			},
			DurationText: topiccardmodel.VideoDuration(archive.Duration),
			Title:        archive.Title,
			Desc:         archive.Desc,
			Badge:        constructArchiveMajorBadge(dynCtx.Dyn.SType, archive.Rights),
		},
	}
	if res.Archive == nil {
		return nil
	}
	res.Archive.Bvid, _ = bvid.AvToBv(archive.Aid)
	res.Archive.JumpUrl = topiccardmodel.FillURI(topiccardmodel.GotoWebAv, res.Archive.Bvid, topiccardmodel.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, true))
	// pgc特殊处理 jumpurl
	if archive.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && archive.RedirectURL != "" {
		res.Archive.JumpUrl = archive.RedirectURL
	}
	if archive.AttrVal(arcgrpc.AttrBitUGCPay) == arcgrpc.AttrYes {
		res.Archive.DisablePreview = true
	}
	return res
}

func constructArchiveMajorBadge(sType int64, rights arcgrpc.Rights) jsonwebcard.Badge {
	switch {
	case rights.IsCooperation == 1:
		// 合作角标
		return jsonwebcard.Badge{
			Text:    "合作",
			Color:   "#FFFFFF",
			BgColor: "#FB7299",
		}
	case rights.UGCPay == 1:
		// 付费角标
		return jsonwebcard.Badge{
			Text:    "付费",
			Color:   "#FFFFFF",
			BgColor: "#FAAB4B",
		}
	case sType == dynmdlV2.VideoStypePlayback:
		// 直播回放
		return jsonwebcard.Badge{
			Text:    "直播回放",
			Color:   "#FFFFFF",
			BgColor: "#FB7299",
		}
	default:
		return jsonwebcard.Badge{
			Text:    "投稿视频",
			Color:   "#FFFFFF",
			BgColor: "#FB7299",
		}
	}
}

func handleDynamicDesc(dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.DynDesc {
	if dynCtx.Interim == nil || dynCtx.Interim.Desc == "" {
		return nil
	}
	return &jsonwebcard.DynDesc{
		Text:         dynCtx.Interim.Desc,
		RichTextNode: descProc(dynCtx, dynCtx.Interim.Desc),
	}
}
