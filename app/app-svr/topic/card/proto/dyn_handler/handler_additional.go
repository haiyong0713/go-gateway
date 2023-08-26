package dynHandler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	bcgmdl "go-gateway/app/app-svr/app-dynamic/interface/model/bcg"
	cheesemdl "go-gateway/app/app-svr/app-dynamic/interface/model/cheese"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	midint64 "go-gateway/app/app-svr/app-interface/interface-legacy/middleware/midInt64"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	"go-gateway/pkg/idsafe/bvid"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	natpagegrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
)

const (
	_voteWait   = 0 // 待审
	_voteOK     = 1 // 正常
	_voteDel    = 2 // 删除
	_voteRefuse = 3 // 未过审
	_voteDead   = 4 // 失效

	_voteTypeWord = 0 // 文字类型
	_voteTypePic  = 1 // 图片类型

	// UP预约状态
	_upNotStart = 0 // 未开始
	_upStart    = 1 // 预约中
	_upOnline   = 2 // 已上线
	_upDelete   = 3 // 删除
	_upCancel   = 4 // 取消
	_upAudit    = 6 // 先审后发
	_upExpired  = 7 // 已过期

	// 附加卡模块
	_additionUpActivityOffline       = "已下线"
	_additionUpActivityOnline        = "去看看"
	_moduleAdditionalNatPageHeadText = "推荐活动"
	_moduleAdditionalVoteVoted       = "已投票"
	_moduleAdditionalVoteClose       = "去查看"
	_moduleAdditionalVoteOpen        = "去投票"
	_moduleAdditionalVoteTips        = "投票不见了"
	_moduleAdditionalItemNull        = "https://i0.hdslb.com/bfs/feed-admin/e08c12a37975eb3da72291167ac1519a16af61e7.png"
	_moduleAdditionalIcon            = "https://i0.hdslb.com/bfs/feed-admin/0c62d6a31f560c3942a787ab5220b048458ab397.png"
	_additionalButtonShareIcon       = "https://i0.hdslb.com/bfs/feed-admin/466b93bea49193b2ca71ce8bc52fdc8db933c335.png"
)

func (schema *CardSchema) additional(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if dynCtx.Dyn.IsForward() && (dynCtx.Interim.ForwardOrigFaild || dynCtx.Interim.IsPassAddition) {
		return nil
	}
	for _, v := range dynCtx.Dyn.AttachCardInfos {
		var common *dynamicapi.ModuleAdditional
		switch v.CardType {
		case dyncommongrpc.AttachCardType_ATTACH_CARD_UGC:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// ugc附加卡
			if additionUGC, ok := dynCtx.GetArchive(v.Rid); ok {
				common = schema.additionalUGC(additionUGC)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_PUGV:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 课程
			if res, ok := dynCtx.ResPUgv[v.Rid]; ok {
				common = schema.additionalPugv(res, dynCtx)
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_VOTE:
			// 投票
			if res, ok := dynCtx.ResVote[v.Rid]; ok {
				common = schema.additionalVote(res, dynCtx)
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
			common = schema.additionUpActivity(dynCtx, natPage, general)
		case dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:
			if !dynSchemaCtx.CanBeAdditionalReserve() {
				continue
			}
			if res, ok := dynCtx.ResUpActRelationInfo[v.Rid]; ok {
				if dynSchemaCtx.IsDisableInt64MidVersion && midint64.CheckHasInt64InMids(res.Upmid) {
					// mid > int32老版本抛弃当前卡片
					continue
				}
				common, ok = schema.additionalUP(res, dynCtx, general)
				// 预约卡删除了，展示失效卡
				if !ok {
					module := schema.additionalNull("原预约信息已删除")
					dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
					continue
				}
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_GOODS:
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				continue
			}
			// 商品
			resGoods, ok := dynCtx.ResGood[dynCtx.Dyn.DynamicID]
			if ok {
				if res, ok := resGoods[bcgmdl.GoodsLocTypeCard]; ok {
					common = schema.additionalGoods(res, dynCtx, general)
				}
			}
		default:
			log.Warn("module error mid(%v) dynid(%v) additional unknown type %+v", general.Mid, dynCtx.Dyn.DynamicID, v.CardType)
			continue
		}
		if common == nil {
			log.Warn("module error mid(%v) dynid(%v) additional addition_type %+v", general.Mid, dynCtx.Dyn.DynamicID, v.CardType)
			continue
		}
		common.Rid = v.Rid
		common.NeedWriteCalender = v.NeedWriteCalender
		module := &dynamicapi.Module{
			ModuleType: dynamicapi.DynModuleType_module_additional,
			ModuleItem: &dynamicapi.Module_ModuleAdditional{
				ModuleAdditional: common,
			},
		}
		dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	}
	return nil
}

func (schema *CardSchema) additionalUGC(ap *archivegrpc.ArcPlayer) *dynamicapi.ModuleAdditional {
	var arc = ap.Arc
	ugc := &dynamicapi.AdditionUgc{
		HeadText:   "相关视频",
		Cover:      arc.Pic,
		Title:      arc.Title,
		DescText_2: fmt.Sprintf("%s观看 %s弹幕", schema.numTransfer(int(arc.Stat.View)), schema.numTransfer(int(arc.Stat.Danmaku))),
		LineFeed:   true,
		Duration:   topiccardmodel.VideoDuration(arc.Duration),
		CardType:   "ugc",
		Uri:        topiccardmodel.FillURI(topiccardmodel.GotoAv, strconv.FormatInt(arc.Aid, 10), topiccardmodel.AvPlayHandlerGRPCV2(ap, arc.FirstCid, true)),
	}
	return &dynamicapi.ModuleAdditional{
		Type: dynamicapi.AdditionalType_additional_type_ugc,
		Item: &dynamicapi.ModuleAdditional_Ugc{
			Ugc: ugc,
		},
	}
}

func (schema *CardSchema) additionalPugv(res *cheesemdl.Cheese, dynCtx *dynmdlV2.DynamicContext) *dynamicapi.ModuleAdditional {
	common := &dynamicapi.AdditionCommon{
		HeadText:   "相关付费课程",
		Title:      res.Title,
		ImageUrl:   res.Cover,
		DescText_1: res.SubTitle,
		DescText_2: res.CardInfo,
		Style:      dynamicapi.ImageStyle_add_style_vertical,
		Type:       strconv.FormatInt(dynCtx.Dyn.Type, 10),
		CardType:   "pugv",
	}
	if res.Button != nil {
		common.Url = res.Button.JumpURL
		common.Button = &dynamicapi.AdditionalButton{
			Type:    dynamicapi.AddButtonType(res.Button.Type),
			JumpUrl: res.Button.JumpURL,
		}
		if res.Button.JumpStyle != nil {
			common.Button.JumpStyle = &dynamicapi.AdditionalButtonStyle{
				Icon: res.Button.JumpStyle.Icon,
				Text: res.Button.JumpStyle.Text,
			}
		}
		if res.Button.UnCheck != nil {
			common.Button.JumpStyle = &dynamicapi.AdditionalButtonStyle{
				Icon: res.Button.UnCheck.Icon,
				Text: res.Button.UnCheck.Text,
			}
		}
		if res.Button.Check != nil {
			common.Button.JumpStyle = &dynamicapi.AdditionalButtonStyle{
				Icon: res.Button.Check.Icon,
				Text: res.Button.Check.Text,
			}
		}
	}
	return &dynamicapi.ModuleAdditional{
		Type: dynamicapi.AdditionalType_additional_type_common,
		Item: &dynamicapi.ModuleAdditional_Common{
			Common: common,
		},
	}
}

// nolint:gocognit
func (schema *CardSchema) additionalVote(vote *dyncommongrpc.VoteInfo, dynCtx *dynmdlV2.DynamicContext) *dynamicapi.ModuleAdditional {
	var (
		voteTmp      *dynamicapi.AdditionVote2
		labels       []string
		now          = time.Now().Unix()
		voteTotalCnt int32
	)
	if vote.Status == _voteDel || vote.Status == _voteRefuse {
		voteTmp = &dynamicapi.AdditionVote2{
			AdditionVoteType: dynamicapi.AdditionVoteType_addition_vote_type_none,
			Tips:             _moduleAdditionalVoteTips,
		}
		goto END
	}
	voteTmp = &dynamicapi.AdditionVote2{
		VoteId:             vote.GetVoteId(),
		Title:              vote.GetTitle(),
		Deadline:           vote.GetEndTime(),
		OpenText:           _moduleAdditionalVoteOpen,
		CloseText:          _moduleAdditionalVoteClose,
		VotedText:          _moduleAdditionalVoteVoted,
		BizType:            vote.GetBizType(),
		Total:              vote.GetJoinNum(),
		CardType:           "vote",
		Uri:                fmt.Sprintf(topiccardmodel.VoteURI, vote.VoteId, dynCtx.Dyn.DynamicID),
		ChoiceCnt:          vote.GetChoiceCnt(),
		DefauleSelectShare: true,
	}
	// 过期判断
	if now >= vote.EndTime {
		voteTmp.State = dynamicapi.AdditionVoteState_addition_vote_state_close
	} else {
		voteTmp.State = dynamicapi.AdditionVoteState_addition_vote_state_open
	}
	// 文案组装
	if vote.GetJoinNum() == 0 {
		labels = append(labels, "0人参与")
	} else {
		labels = append(labels, topiccardmodel.StatString(vote.GetJoinNum(), "人投票", ""))
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
			voteTmp.AdditionVoteType = dynamicapi.AdditionVoteType_addition_vote_type_word
			var (
				items      []*dynamicapi.AdditionVoteWordItem
				maxOptionm = make(map[int32][]int)
				maxCnt     int32
			)
			for _, option := range vote.Options {
				if option == nil {
					continue
				}
				// 选项基础信息
				item := &dynamicapi.AdditionVoteWordItem{
					OptIdx: option.GetOptIdx(),
					Title:  option.GetOptDesc(),
					Total:  option.GetCnt(),
				}
				// 选项票数百分比
				if voteTotalCnt != 0 {
					var err error
					item.Persent, err = strconv.ParseFloat(fmt.Sprintf("%.2f", float32(option.GetCnt())/float32(voteTotalCnt)), 64)
					if err != nil {
						log.Error("additionalVote percent option %+v, err %v", option, err)
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
				log.Warn("additional miss vote dynid(%v) items len=0", dynCtx.Dyn.DynamicID)
				return nil
			}
			// 回填标记票数最多的选项
			for _, idx := range maxOptionm[maxCnt] {
				items[idx].IsMaxOption = true
			}
			voteTmp.Item = &dynamicapi.AdditionVote2_AdditionVoteWord{
				AdditionVoteWord: &dynamicapi.AdditionVoteWord{
					Item: items,
				},
			}
		case _voteTypePic:
			voteTmp.AdditionVoteType = dynamicapi.AdditionVoteType_addition_vote_type_pic
			var (
				items      []*dynamicapi.AdditionVotePicItem
				maxOptionm = make(map[int32][]int)
				maxCnt     int32
			)
			for _, option := range vote.Options {
				if option == nil {
					continue
				}
				// 选项基础信息
				item := &dynamicapi.AdditionVotePicItem{
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
			voteTmp.Item = &dynamicapi.AdditionVote2_AdditionVotePic{
				AdditionVotePic: &dynamicapi.AdditionVotePic{
					Item: items,
				},
			}
		}
	} else {
		voteTmp.AdditionVoteType = dynamicapi.AdditionVoteType_addition_vote_type_default
		var items []string
		for _, option := range vote.Options {
			if option == nil {
				continue
			}
			items = append(items, option.GetImgUrl())
		}
		voteTmp.Item = &dynamicapi.AdditionVote2_AdditionVoteDefaule{
			AdditionVoteDefaule: &dynamicapi.AdditionVoteDefaule{
				Cover: items,
			},
		}
	}
END:
	additional := &dynamicapi.ModuleAdditional{
		Type: dynamicapi.AdditionalType_additional_type_vote,
		Item: &dynamicapi.ModuleAdditional_Vote2{
			Vote2: voteTmp,
		},
	}
	dynCtx.Interim.VoteID = voteTmp.VoteId // 转发内外层传递
	return additional
}

// additionalUP up主预约卡
func (schema *CardSchema) additionalUP(up *activitygrpc.UpActReserveRelationInfo, dynCtx *dynmdlV2.DynamicContext, general *topiccardmodel.GeneralParam) (*dynamicapi.ModuleAdditional, bool) {
	common, ok := schema.additionalUPInfo(up, dynCtx, general)
	if common == nil {
		return nil, ok
	}
	if common.Button != nil {
		common.Button.ClickType = dynamicapi.AdditionalButtonClickType_click_up
	}
	additional := &dynamicapi.ModuleAdditional{
		Type: dynamicapi.AdditionalType_additional_type_up_reservation,
		Item: &dynamicapi.ModuleAdditional_Up{
			Up: common,
		},
	}
	return additional, true
}

func (schema *CardSchema) additionalUPInfo(up *activitygrpc.UpActReserveRelationInfo, dynCtx *dynmdlV2.DynamicContext, general *topiccardmodel.GeneralParam) (*dynamicapi.AdditionUP, bool) {
	const (
		_liveStart      = 1
		_liveAv         = 2
		playOnlineTotal = "total"
	)
	common := &dynamicapi.AdditionUP{
		Title:        up.Title,
		DescText_2:   topiccardmodel.StatString(up.Total, "人预约", ""),
		CardType:     "reserve",
		ReserveTotal: up.Total,
		Rid:          up.Sid,
		LotteryType:  dynamicapi.ReserveRelationLotteryType_reserve_relation_lottery_type_default,
		UpMid:        up.Upmid,
		DynamicId:    up.DynamicId,
		ShowText_2:   true,
	}
	if userInfo, ok := dynCtx.GetUser(up.Upmid); ok {
		common.UserInfo = &dynamicapi.AdditionUserInfo{
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
		common.LotteryType = dynamicapi.ReserveRelationLotteryType_reserve_relation_lottery_type_cron
	}
	// 高亮icon跳链文案
	if up.PrizeInfo != nil {
		common.DescText_3 = &dynamicapi.HighlightText{
			Text:      up.PrizeInfo.Text,
			JumpUrl:   up.PrizeInfo.JumpUrl,
			TextStyle: dynamicapi.HighlightTextStyle_style_highlight,
			Icon:      _moduleAdditionalIcon,
		}
	}

	switch up.Type {
	case activitygrpc.UpActReserveRelationType_Archive: // 稿件
		if up.LivePlanStartTime.Time().Unix() > 0 {
			common.DescText_1 = &dynamicapi.HighlightText{
				Text: fmt.Sprintf("预计%s发布", model.UpPubDataString(up.LivePlanStartTime.Time())),
			}
		}
	case activitygrpc.UpActReserveRelationType_Live: // 直播
		if up.LivePlanStartTime.Time().Unix() > 0 {
			common.DescText_1 = &dynamicapi.HighlightText{
				Text: fmt.Sprintf("%s直播", model.UpPubDataString(up.LivePlanStartTime.Time())),
			}
		}
		if up.Ext != "" {
			tmp := &activitygrpc.UpActReserveRelationInfoExtend{}
			// 大航海
			if err := json.Unmarshal([]byte(up.Ext), &tmp); err == nil && tmp.SubType == 1 {
				common.BadgeText = "大航海专属"
			}
		}
	case activitygrpc.UpActReserveRelationType_ESports: // 赛事
		if up.Desc != "" {
			common.DescText_1 = &dynamicapi.HighlightText{
				Text: up.Desc,
			}
		}
	case activitygrpc.UpActReserveRelationType_Premiere: // 首映
		if general.IsIPhonePick() && general.GetBuild() >= 66700000 || general.IsAndroidPick() && general.GetBuild() >= 6670000 {
			common.DescText_1 = &dynamicapi.HighlightText{
				Text: fmt.Sprintf("%s首映", model.UpPubDataString(up.LivePlanStartTime.Time())),
			}
			common.IsPremiere = true
		}
	default:
	}

	// 主人态
	if up.Upmid == general.Mid {
		common.Button = &dynamicapi.AdditionalButton{
			Type:    dynamicapi.AddButtonType_bt_button,
			Status:  dynamicapi.AdditionalButtonStatus_uncheck,
			Uncheck: upButtonCheck("", topiccardmodel.UpbuttonCancel, dynamicapi.AddButtonBgStyle_stroke, dynamicapi.DisableState_highlight),
			Check:   upButtonCheck("UP主已撤销预约", topiccardmodel.UpbuttonCancelOk, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_gary), // 置灰不可点击
		}
		if up.LotteryType == activitygrpc.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron {
			common.Button.Uncheck = upButtonCheck("", topiccardmodel.UpbuttonCancelLotteryCron, dynamicapi.AddButtonBgStyle_stroke, dynamicapi.DisableState_highlight)
		}
		common.Button.Uncheck.Share, common.BusinessId, common.DynType = schema.additionalButtonShare(up, dynCtx, general)
	} else if up.Upmid != general.Mid { // 客人态
		if up.IsFollow == 1 {
			common.Button = &dynamicapi.AdditionalButton{
				Type:    dynamicapi.AddButtonType_bt_button,
				Status:  dynamicapi.AdditionalButtonStatus_check,
				Check:   upButtonCheck("", topiccardmodel.UpbuttonReservationOk, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_highlight),
				Uncheck: upButtonCheck("", topiccardmodel.UpbuttonReservation, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_highlight),
			}
		} else {
			common.Button = &dynamicapi.AdditionalButton{
				Type:    dynamicapi.AddButtonType_bt_button,
				Status:  dynamicapi.AdditionalButtonStatus_uncheck,
				Uncheck: upButtonCheck("", topiccardmodel.UpbuttonReservation, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_highlight),
				Check:   upButtonCheck("", topiccardmodel.UpbuttonReservationOk, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_highlight),
			}
		}
		if up.LotteryType == activitygrpc.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron {
			common.Button.Uncheck.Toast = "预约成功，已参与抽奖"
		}
		common.Button.Check.Share, common.BusinessId, common.DynType = schema.additionalButtonShare(up, dynCtx, general)
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
		case activitygrpc.UpActReserveRelationType_Archive, activitygrpc.UpActReserveRelationType_Premiere: // 稿件
			aid, _ := strconv.ParseInt(up.Oid, 10, 64)
			ap, ok := dynCtx.GetArchive(aid)
			common.Url = model.FillURI(model.GotoAv, strconv.FormatInt(aid, 10), nil) // 不要秒开，且稿件不存在还是返回url，进入详情页展示稿件不存在
			if !ok {
				if up.IsFollow == 1 && up.Upmid != general.Mid {
					common.Button = &dynamicapi.AdditionalButton{
						Type:   dynamicapi.AddButtonType_bt_button,
						Status: dynamicapi.AdditionalButtonStatus_check,
						Check:  upButtonCheck("不在预约时间", topiccardmodel.UpbuttonReservationOk, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_gary),
					}
				} else {
					common.Button = &dynamicapi.AdditionalButton{
						Type:   dynamicapi.AddButtonType_bt_button,
						Status: dynamicapi.AdditionalButtonStatus_check,
						Check:  upButtonCheck("不在预约时间", topiccardmodel.UpbuttonReservation, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_gary),
					}
				}
			} else {
				var arc = ap.Arc
				common.DescText_2 = model.UpStatString(int64(arc.Stat.View), "观看")
				common.Button = &dynamicapi.AdditionalButton{
					Type:      dynamicapi.AddButtonType_bt_jump,
					Status:    dynamicapi.AdditionalButtonStatus_uncheck,
					JumpUrl:   model.FillURI(model.GotoAv, strconv.FormatInt(arc.Aid, 10), model.AvPlayHandlerGRPCV2(ap, arc.FirstCid, true)),
					JumpStyle: upButtonCheck("", topiccardmodel.UpbuttonWatch, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_highlight),
				}
				// 首映
				if up.Type == activitygrpc.UpActReserveRelationType_Premiere {
					countInfo, ok := dynCtx.ResPlayUrlCount[ap.Arc.Aid]
					if arc.Premiere != nil {
						switch arc.Premiere.State {
						case archivegrpc.PremiereState_premiere_in:
							common.DescText_1 = &dynamicapi.HighlightText{
								Text:      "首映中",
								TextStyle: dynamicapi.HighlightTextStyle_style_highlight,
							}
							if ok {
								common.DescText_2 = model.UpStatString(countInfo.Count[playOnlineTotal], "人在线")
							}
						case archivegrpc.PremiereState_premiere_after:
							if ok {
								common.DescText_2 = model.UpStatString(countInfo.Count[playOnlineTotal], "观看")
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
						common.DescText_2 = makeReserveLiveStartDesc2(live)
						common.Button = &dynamicapi.AdditionalButton{
							Type:      dynamicapi.AddButtonType_bt_jump,
							Status:    dynamicapi.AdditionalButtonStatus_uncheck,
							JumpUrl:   common.Url,
							JumpStyle: upButtonCheck("", topiccardmodel.UpbuttonWatch, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_highlight),
						}
						common.DescText_1 = &dynamicapi.HighlightText{
							Text:      "直播中",
							TextStyle: dynamicapi.HighlightTextStyle_style_highlight,
						}
					case _liveAv:
						aid, _ := bvid.BvToAv(live.Bvid)
						common.Url = model.FillURI(model.GotoAv, strconv.FormatInt(aid, 10), nil)
						common.Button = &dynamicapi.AdditionalButton{
							Type:      dynamicapi.AddButtonType_bt_jump,
							Status:    dynamicapi.AdditionalButtonStatus_uncheck,
							JumpUrl:   common.Url,
							JumpStyle: upButtonCheck("", topiccardmodel.UpbuttonReplay, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_highlight),
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
				common.Button = &dynamicapi.AdditionalButton{
					Type:   dynamicapi.AddButtonType_bt_button,
					Status: dynamicapi.AdditionalButtonStatus_check,
					Check:  upButtonCheck("直播已结束", topiccardmodel.UpbuttonEnd, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_gary),
				}
			}
		case activitygrpc.UpActReserveRelationType_ESports:
			// 赛事核销
			if up.State == activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackDone {
				common.Button = &dynamicapi.AdditionalButton{
					Type:   dynamicapi.AddButtonType_bt_button,
					Status: dynamicapi.AdditionalButtonStatus_check,
					Check:  upButtonCheck("已结束", topiccardmodel.UpbuttonEnd, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_gary),
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
					common.Button = &dynamicapi.AdditionalButton{
						Type:   dynamicapi.AddButtonType_bt_button,
						Status: dynamicapi.AdditionalButtonStatus_check,
						Check:  upButtonCheck("预约已过期", topiccardmodel.UpbuttonReservationOk, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_gary),
					}
				} else {
					common.Button = &dynamicapi.AdditionalButton{
						Type:   dynamicapi.AddButtonType_bt_button,
						Status: dynamicapi.AdditionalButtonStatus_check,
						Check:  upButtonCheck("预约已过期", topiccardmodel.UpbuttonReservation, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_gary),
					}
				}
			} else {
				common.Button = &dynamicapi.AdditionalButton{
					Type:   dynamicapi.AddButtonType_bt_button,
					Status: dynamicapi.AdditionalButtonStatus_check,
					Check:  upButtonCheck("预约已过期", topiccardmodel.UpbuttonCancelOk, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_gary),
				}
			}
		}
	case _upCancel:
		common.Button = &dynamicapi.AdditionalButton{
			Type:   dynamicapi.AddButtonType_bt_button,
			Status: dynamicapi.AdditionalButtonStatus_check,
			Check:  upButtonCheck("UP主已撤销预约", topiccardmodel.UpbuttonCancelOk, dynamicapi.AddButtonBgStyle_fill, dynamicapi.DisableState_gary),
		}
		// 撤销不展示抽奖信息
		common.LotteryType = dynamicapi.ReserveRelationLotteryType_reserve_relation_lottery_type_default
		common.DescText_3 = nil
	case _upDelete:
		// 预约卡被删除
		return nil, false
	}
	if common.Button != nil {
		common.Button.ClickType = dynamicapi.AdditionalButtonClickType_click_up
	}
	return common, true
}

func (schema *CardSchema) additionalGoods(res map[string]*bcgmdl.GoodsItem, dynCtx *dynmdlV2.DynamicContext, general *topiccardmodel.GeneralParam) *dynamicapi.ModuleAdditional {
	goods := &dynamicapi.AdditionGoods{
		CardType:   "goods",
		Icon:       "https://i0.hdslb.com/bfs/feed-admin/3ac25959e29285fa56c378844a978841661adf78.png",
		Uri:        dynCtx.DynamicItem.Extend.CardUrl,
		AdMarkIcon: "https://i0.hdslb.com/bfs/feed-admin/3ac25959e29285fa56c378844a978841661adf78.png",
	}
	for _, id := range strings.Split(dynCtx.Dyn.Extend.OpenGoods.ItemsId, ",") {
		goodsItem, ok := res[id]
		if !ok {
			continue
		}
		if (general.IsIPhonePick() && general.GetBuild() < 66000000 || general.IsAndroidPick() && general.GetBuild() < 6600000) && (goodsItem.SourceType != 1 && goodsItem.SourceType != 2) {
			continue
		}
		tmp := &dynamicapi.GoodsItem{
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
			JumpType:          dynamicapi.GoodsJumpType_goods_schema,
		}
		if goodsItem.OuterApp == 0 {
			tmp.JumpType = dynamicapi.GoodsJumpType_goods_url
		}
		goods.RcmdDesc = goodsItem.AdMark
		goods.GoodsItems = append(goods.GoodsItems, tmp)
	}
	if len(goods.GoodsItems) > 0 {
		goods.SourceType = goods.GoodsItems[0].SourceType
		goods.JumpType = goods.GoodsItems[0].JumpType
		goods.AppName = goods.GoodsItems[0].AppName
		additional := &dynamicapi.ModuleAdditional{
			Type: dynamicapi.AdditionalType_additional_type_goods,
			Item: &dynamicapi.ModuleAdditional_Goods{
				Goods: goods,
			},
		}
		return additional
	}
	return nil
}

func makeReserveLiveStartDesc2(liveInfo *livexroomgate.SessionInfoPerLive) string {
	if show := liveInfo.WatchedShow; show != nil && show.TextLarge != "" {
		return show.TextLarge
	}
	return model.UpStatString(liveInfo.PopularityCount, "人气")
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

func (schema *CardSchema) additionalButtonShare(up *activitygrpc.UpActReserveRelationInfo, dynCtx *dynmdlV2.DynamicContext, general *topiccardmodel.GeneralParam) (share *dynamicapi.AdditionalButtonShare, businessID string, dynType int64) {
	if general.IsPad() || general.IsPadHD() || general.IsAndroidHD() || general.IsOverseas() {
		return nil, "", 0
	}
	if up.Type == activitygrpc.UpActReserveRelationType_Premiere {
		return nil, "", 0
	}
	dynID, _ := strconv.ParseInt(up.DynamicId, 10, 64)
	dynInfo, dynOk := dynCtx.ResDynSimpleInfos[dynID]
	if dynOk && dynInfo.Visible {
		businessID = strconv.FormatInt(dynInfo.Rid, 10)
		dynType = dynInfo.Type
	}
	return &dynamicapi.AdditionalButtonShare{
		Icon: _additionalButtonShareIcon,
		Text: "分享",
		Show: dynamicapi.AdditionalShareShowType_st_show,
	}, businessID, dynType
}

func (schema *CardSchema) additionUpActivity(dynCtx *dynmdlV2.DynamicContext, natPage *natpagegrpc.NativePageCard, _ *topiccardmodel.GeneralParam) *dynamicapi.ModuleAdditional {
	const (
		// 活动上线
		_online = 1
	)
	common := &dynamicapi.AdditionCommon{
		HeadText:   _moduleAdditionalNatPageHeadText,
		Title:      fmt.Sprintf("#%s#", natPage.Title),
		DescText_1: natPage.ShareTitle,
		ImageUrl:   natPage.ShareImage,
		Style:      dynamicapi.ImageStyle_add_style_square,
		Type:       strconv.FormatInt(dynCtx.Dyn.Type, 10),
		CardType:   "up_activity",
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
	common.Url = natPage.PcURL
	// 聚合预约状态
	if natPage.State == _online {
		common.Button = &dynamicapi.AdditionalButton{
			Type:    dynamicapi.AddButtonType_bt_jump,
			JumpUrl: natPage.SkipURL,
			JumpStyle: &dynamicapi.AdditionalButtonStyle{
				Text:    _additionUpActivityOnline,
				BgStyle: dynamicapi.AddButtonBgStyle_fill,
				Disable: dynamicapi.DisableState_highlight,
			},
		}
	} else {
		common.Button = &dynamicapi.AdditionalButton{
			Type:   dynamicapi.AddButtonType_bt_button,
			Status: dynamicapi.AdditionalButtonStatus_check,
			Check: &dynamicapi.AdditionalButtonStyle{
				Text:    _additionUpActivityOffline,
				BgStyle: dynamicapi.AddButtonBgStyle_gray,
				Disable: dynamicapi.DisableState_gary,
			},
		}
	}
	additional := &dynamicapi.ModuleAdditional{
		Type: dynamicapi.AdditionalType_additional_type_common,
		Item: &dynamicapi.ModuleAdditional_Common{
			Common: common,
		},
	}
	return additional
}

// 附加卡被删除
func (schema *CardSchema) additionalNull(text string) *dynamicapi.Module {
	return &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_item_null,
		ModuleItem: &dynamicapi.Module_ModuleItemNull{
			ModuleItemNull: &dynamicapi.ModuleItemNull{
				Icon: _moduleAdditionalItemNull,
				Text: text,
			},
		},
	}
}

func upButtonCheck(toast string, state int, bgStyle dynamicapi.AddButtonBgStyle, disable dynamicapi.DisableState) *dynamicapi.AdditionalButtonStyle {
	check := &dynamicapi.AdditionalButtonStyle{
		BgStyle: bgStyle,
		Disable: disable,
		Toast:   toast,
	}
	switch state {
	case topiccardmodel.UpbuttonReservation: // 预约
		check.Icon = "https://i0.hdslb.com/bfs/archive/f5b7dae25cce338e339a655ac0e4a7d20d66145c.png"
		check.Text = "预约"
	case topiccardmodel.UpbuttonReservationOk: // 已预约
		check.Text = "已预约"
	case topiccardmodel.UpbuttonCancel: // 取消预约
		check.Text = "撤销"
		check.Interactive = &dynamicapi.AdditionalButtonInteractive{
			Popups:  "撤销预约后，将提醒已预约用户",
			Confirm: "撤销预约",
			Cancel:  "取消",
		}
	case topiccardmodel.UpbuttonCancelLotteryCron:
		check.Text = "撤销"
		check.Interactive = &dynamicapi.AdditionalButtonInteractive{
			Popups:  "撤销预约将提醒已预约用户",
			Desc:    "撤销本次预约后会关闭抽奖",
			Confirm: "撤销预约",
			Cancel:  "取消",
		}
	case topiccardmodel.UpbuttonCancelOk: // 已取消
		check.Text = "已撤销"
	case topiccardmodel.UpbuttonWatch: // 去观看
		check.Text = "去观看"
	case topiccardmodel.UpbuttonReplay: // 回放
		check.Text = "看回放"
	case topiccardmodel.UpbuttonEnd: // 已结束
		check.Text = "已结束"
	default:
		return nil
	}
	return check
}
