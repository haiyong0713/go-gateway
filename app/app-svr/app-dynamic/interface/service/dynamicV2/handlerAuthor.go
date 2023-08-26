package dynamicV2

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	relationmdl "go-gateway/app/app-svr/app-dynamic/interface/model/relation"
	submdl "go-gateway/app/app-svr/app-dynamic/interface/model/subscription"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
)

func (s *Service) authorUser(c context.Context, uid int64, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) *api.ModuleAuthor {
	if uid == 0 {
		return nil
	}
	userInfo, ok := dynCtx.GetUser(uid)
	if !ok {
		log.Warn("module miss mid(%v) dynid(%v) author uid(%d)", general.Mid, dynCtx.Dyn.DynamicID, uid)
		return nil
	}
	ptimeLabelText := s.publishLabel(c, dynCtx, general)
	switch dynCtx.From {
	case _handleTypeFake:
		// 假卡发布文案
		ptimeLabelText = s.c.Resource.Text.ModuleAuthorPublishLabelDefault
		switch {
		case dynCtx.Dyn.IsAv():
			ptimeLabelText = s.c.Resource.Text.ModuleAuthorPublishLabelArchive
		}
	}
	author := &api.ModuleAuthor{
		Mid: userInfo.Mid,
		Author: &api.UserInfo{
			Mid:  userInfo.Mid,
			Name: userInfo.Name,
			Face: userInfo.Face,
			Official: &api.OfficialVerify{ // 认证
				Type: int32(userInfo.Official.Type),
				Desc: userInfo.Official.Desc,
			},
			Vip: &api.VipInfo{ // 会员
				Type:    userInfo.Vip.Type,
				Status:  userInfo.Vip.Status,
				DueDate: userInfo.Vip.DueDate,
				Label: &api.VipLabel{
					Path:       userInfo.Vip.Label.Path,
					Text:       userInfo.Vip.Label.Text,
					LabelTheme: userInfo.Vip.Label.LabelTheme,
				},
				ThemeType:       userInfo.Vip.ThemeType,
				AvatarSubscript: userInfo.Vip.AvatarSubscript,
				NicknameColor:   userInfo.Vip.NicknameColor,
			},
			Pendant: &api.UserPendant{ // 头像挂件
				Pid:    int64(userInfo.Pendant.Pid),
				Name:   userInfo.Pendant.Name,
				Image:  userInfo.Pendant.Image,
				Expire: int64(userInfo.Pendant.Expire),
			},
			Nameplate: &api.Nameplate{ // 勋章
				Nid:        int64(userInfo.Nameplate.Nid),
				Name:       userInfo.Nameplate.Name,
				Image:      userInfo.Nameplate.Image,
				ImageSmall: userInfo.Nameplate.ImageSmall,
				Level:      userInfo.Nameplate.Level,
				Condition:  userInfo.Nameplate.Condition,
			},
			Uri:        model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(userInfo.Mid, 10), nil),
			FaceNft:    userInfo.FaceNft,
			FaceNftNew: userInfo.FaceNftNew,
		},
		PtimeLabelText: ptimeLabelText,
		Uri:            dynCtx.Interim.PromoURI,                                   // 帮推
		Relation:       relationmdl.RelationChange(uid, dynCtx.ResRelationUltima), // 关注组件
	}
	// 空间页不下发跳转空间页url
	if dynCtx.From == _handleTypeSpace && general.DynFrom == _dynFromSpace {
		author.Author.Uri = ""
	}
	// IP属地展示
	if s.appFeatureGate.UserIPDisplay().Enabled(c) && general.Mid > 0 && s.isPubLocationCapable(c, general) &&
		time.Unix(dynCtx.Dyn.Timestamp, 0).After(s.c.Ctrl.IPDisplayAfter) &&
		dynCtx.From == _handleTypeView {
		dynid, idok := dynCtx.GetDynamicID()
		manageip, ipok := dynCtx.GetManagerIpDisplay(dynid)
		// 如果管理平台已经指定了IP显示信息，则直接使用此信息；否则走常规流程
		if idok && ipok {
			if len(manageip) > 0 {
				author.PtimeLocationText = "IP属地：" + manageip
			}
		} else if loc := dynCtx.GetPublishAddr(); len(loc) > 0 {
			author.PtimeLocationText = "IP属地：" + loc
		}
		// 如果仅自见 则抹掉
		if s.appFeatureGate.UserIPDisplay().SelfVisibleOnly() && general.Mid != dynCtx.Dyn.UID {
			author.PtimeLocationText = ""
		}
	}
	// 详情页展示已编辑标识 直接拼接在IP属地文案前面
	if s.isPubLocationCapable(c, general) && dynCtx.Dyn.Property.GetEdited() && s.c.Ctrl.ShowDynEdit {
		if len(author.PtimeLabelText) > 0 {
			author.PtimeLocationText = "已编辑 " + author.PtimeLocationText
		} else {
			author.PtimeLocationText = "已编辑"
		}
	}

	return author
}

// 发布人模块
// nolint:gocognit
func (s *Service) author(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID)
	if !ok {
		xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "author", "date_faild")
		log.Warn("module error mid(%v) dynid(%v) author uid(%d)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID)
	}
	var (
		userMdl        *api.Module_ModuleAuthor
		ptimeLabelText = s.publishLabel(c, dynCtx, general)
		isWeight       bool
	)
	switch dynCtx.From {
	case _handleTypeFake:
		// 假卡发布文案
		ptimeLabelText = s.c.Resource.Text.ModuleAuthorPublishLabelDefault
		switch {
		case dynCtx.Dyn.IsAv():
			ptimeLabelText = s.c.Resource.Text.ModuleAuthorPublishLabelArchive
		}
	}
	if userInfo == nil { // 兜底 默认返回灰头像、空白文案
		userMdl = &api.Module_ModuleAuthor{
			ModuleAuthor: &api.ModuleAuthor{
				Mid:            dynCtx.Dyn.UID,
				PtimeLabelText: ptimeLabelText,
				Author: &api.UserInfo{
					Mid:  dynCtx.Dyn.UID,
					Face: s.c.Resource.Icon.ModuleAuthorDefaultFace,
					Uri:  model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(dynCtx.Dyn.UID, 10), nil),
				},
				Uri: dynCtx.Interim.PromoURI, // 帮推
			},
		}
		goto END
	}
	userMdl = &api.Module_ModuleAuthor{
		ModuleAuthor: s.authorUser(c, dynCtx.Dyn.UID, dynCtx, general),
	}
	// 课程预约召回卡未关注则展示关注按钮
	if dynCtx.From == _handleTypeSearch || dynCtx.From == _handleTypeUnLogin ||
		dynCtx.Dyn.Property.GetRcmdType() == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_CHEESE {
		// 未关注、被关注
		if (userMdl.ModuleAuthor.Relation.Status == 1 || userMdl.ModuleAuthor.Relation.Status == 3) && general.Mid != dynCtx.Dyn.UID {
			userMdl.ModuleAuthor.ShowFollow = true // 是否展示关注
		}
	}
	if userInfo.Pendant.ImageEnhance != "" { // 动效图优先
		userMdl.ModuleAuthor.Author.Pendant.Image = userInfo.Pendant.ImageEnhance
	}
	// 直播状态
	// 20201222 加密直播间视为关播
	if dynCtx.From != _handleTypeAllPersonal && dynCtx.From != _handleTypeVideoPersonal && !dynCtx.Interim.HiddenAuthorLive {
		if userLive, ok := dynCtx.GetResUserLive(dynCtx.Dyn.UID); ok && userLive.Status != nil && userLive.Status.Password == "" {
			userMdl.ModuleAuthor.Author.Live = &api.LiveInfo{
				IsLiving:  int32(userLive.Status.LiveStatus),
				LiveState: api.LiveState(userLive.Status.LiveStatus),
				Uri:       model.FillURI(model.GotoLive, strconv.FormatInt(userLive.RoomId, 10), nil),
			}
			if userLivePlayURL, ok := dynCtx.GetResUserLivePlayURL(dynCtx.Dyn.UID); ok {
				userMdl.ModuleAuthor.Author.Live.Uri = userLivePlayURL.Link
			}
		}
	}
	// 提权样式
	if !dynCtx.Dyn.IsSubscription() {
		userMdl.ModuleAuthor.Weight = s.weight(c, dynCtx, general)
		if userMdl.ModuleAuthor.Weight != nil && len(userMdl.ModuleAuthor.Weight.Items) > 0 {
			isWeight = true
		}
	}
	// 空间置顶卡
	if (dynCtx.From == _handleTypeSpace || dynCtx.From == _handleTypeServerDetail) && dynCtx.Dyn.Property != nil && dynCtx.Dyn.Property.IsSpaceTop {
		userMdl.ModuleAuthor.IsTop = true
		// 如果是置顶卡，装扮不展示
		goto END
	}
	// 详情页未关注不展示装扮
	if dynCtx.From == _handleTypeView && userMdl.ModuleAuthor.Relation.Status == 1 {
		goto END
	}
	// 订阅卡不展示装扮 测试wanglulu反馈
	// 直播中不展示装扮
	// 新订阅卡不展示装扮
	if !dynCtx.Dyn.IsSubscription() && !dynCtx.Dyn.IsSubscriptionNew() && (userMdl.ModuleAuthor.Author.Live == nil || userMdl.ModuleAuthor.Author.Live.LiveState != api.LiveState_live_live) && !isWeight {
		if dynCtx.ResMyDecorate != nil {
			decoInfo, ok := dynCtx.ResMyDecorate[userInfo.Mid]
			if ok {
				// 小于6位数 前置补0
				numberStr := strconv.Itoa(decoInfo.Fan.Number)
				// nolint:gomnd
				if decoInfo.Fan.Number < 100000 {
					numberStr = fmt.Sprintf("%06d", decoInfo.Fan.Number)
				}
				userMdl.ModuleAuthor.DecorateCard = &api.DecorateCard{
					Id:      decoInfo.ID,
					CardUrl: decoInfo.CardURL,
					JumpUrl: decoInfo.JumpURL,
					Fan: &api.DecoCardFan{
						IsFan:     int32(decoInfo.Fan.IsFan),
						Number:    int32(decoInfo.Fan.Number),
						NumberStr: numberStr,
						Color:     decoInfo.Fan.Color,
					},
				}
				if decoInfo.ImageEnhance != "" {
					userMdl.ModuleAuthor.DecorateCard.CardUrl = decoInfo.ImageEnhance
				}
			}
		}
	}
END:
	// 旧订阅卡出预约按钮 不出三点
	if dynCtx.Dyn.IsSubscription() {
		if sub, ok := dynCtx.GetResSub(dynCtx.Dyn.Rid); ok && sub.MenuText != "" {
			userMdl.ModuleAuthor.BadgeButton = &api.ModuleAuthorBadgeButton{
				Title: sub.MenuText,
				Id:    dynCtx.Dyn.Rid,
			}
		}
	}
	// 非旧订阅卡出三点
	if !dynCtx.Dyn.IsSubscription() && !isWeight {
		// 如果提权没有则展示三点
		userMdl.ModuleAuthor.TpList = s.threePoint(c, dynCtx, general)
	}
	// 盘古NFT信息
	if nftInfo, ok := dynCtx.ResNFTBatchInfo[dynCtx.Dyn.UID]; ok {
		if nftRegion, ok := dynCtx.ResNFTRegionInfo[nftInfo.NftId]; ok {
			userMdl.ModuleAuthor.Author.NftInfo = &api.NFTInfo{
				RegionType:       api.NFTRegionType(nftRegion.Type),
				RegionIcon:       nftRegion.Icon,
				RegionShowStatus: api.NFTShowStatus(nftRegion.ShowStatus),
			}
		}
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) authorPGC(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid))
	var userMdl *api.Module_ModuleAuthor
	if !ok {
		userMdl = &api.Module_ModuleAuthor{
			ModuleAuthor: &api.ModuleAuthor{
				Author: &api.UserInfo{
					Mid:  dynCtx.Dyn.UID,
					Face: s.c.Resource.Icon.ModuleAuthorDefaultFace,
				},
			},
		}
		goto END
	}
	userMdl = &api.Module_ModuleAuthor{
		ModuleAuthor: &api.ModuleAuthor{
			Author: &api.UserInfo{Mid: dynCtx.Dyn.UID, Uri: pgc.Url},
			Uri:    pgc.Url,
		},
	}
	if pgc.Season != nil {
		userMdl.ModuleAuthor.Author.Name = pgc.Season.Title
		userMdl.ModuleAuthor.Author.Face = pgc.Season.Cover
	} else {
		log.Warn("authorPGC pgc.Season nil mid %v, dynid %v, rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
	}
END:
	userMdl.ModuleAuthor.PtimeLabelText = s.publishLabel(c, dynCtx, general)
	userMdl.ModuleAuthor.Weight = s.weight(c, dynCtx, general)
	// 如果提权没有则展示三点
	if userMdl.ModuleAuthor.Weight == nil || len(userMdl.ModuleAuthor.Weight.Items) == 0 {
		userMdl.ModuleAuthor.TpList = s.threePoint(c, dynCtx, general)
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author,
		ModuleItem: userMdl,
	}
	dynCtx.Modules = append(dynCtx.Modules, module)
	return nil
}

func (s *Service) authorCheeseBatch(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	batch, ok := dynCtx.GetResCheeseBatch(dynCtx.Dyn.Rid)
	var userMdl *api.Module_ModuleAuthor
	if !ok {
		userMdl = &api.Module_ModuleAuthor{
			ModuleAuthor: &api.ModuleAuthor{
				Author: &api.UserInfo{
					Mid:  dynCtx.Dyn.UID,
					Face: s.c.Resource.Icon.ModuleAuthorDefaultFace,
				},
			},
		}
		goto END
	}
	userMdl = &api.Module_ModuleAuthor{
		ModuleAuthor: &api.ModuleAuthor{
			Author: &api.UserInfo{
				Mid:  batch.UpID,
				Name: batch.UpInfo.Name,
				Face: batch.UpInfo.Avatar,
				Official: &api.OfficialVerify{
					Type: int32(batch.UserProfile.Card.OfficialVerify.Type),
					Desc: batch.UserProfile.Card.OfficialVerify.Desc,
				},
				Vip: &api.VipInfo{
					Type:    int32(batch.UserProfile.Vip.VipType),
					Status:  int32(batch.UserProfile.Vip.VipStatus),
					DueDate: batch.UserProfile.Vip.VipDueDate,
					Label: &api.VipLabel{
						Path:       batch.UserProfile.Vip.Label.Path,
						Text:       batch.UserProfile.Vip.Label.Text,
						LabelTheme: batch.UserProfile.Vip.Label.LabelTheme,
					},
					ThemeType:       int32(batch.UserProfile.Vip.ThemeType),
					AvatarSubscript: batch.UserProfile.Vip.AvatarSubscript,
					NicknameColor:   batch.UserProfile.Vip.NicknameColor,
				},
				Uri: model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(batch.UpID, 10), nil),
				Pendant: &api.UserPendant{
					Pid:    batch.UserProfile.Pendant.Pid,
					Name:   batch.UserProfile.Pendant.Name,
					Image:  batch.UserProfile.Pendant.Image,
					Expire: batch.UserProfile.Pendant.Expire,
				},
			},
			Mid:      batch.UpID,
			Relation: relationmdl.RelationChange(batch.UpID, dynCtx.ResRelationUltima), // 关注组件
		},
	}
END:
	userMdl.ModuleAuthor.PtimeLabelText = s.publishLabel(c, dynCtx, general)
	userMdl.ModuleAuthor.Weight = s.weight(c, dynCtx, general)
	// 如果提权没有则展示三点
	if userMdl.ModuleAuthor.Weight == nil || len(userMdl.ModuleAuthor.Weight.Items) == 0 {
		userMdl.ModuleAuthor.TpList = s.threePoint(c, dynCtx, general)
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) authorUGCSeason(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	ugcSeason, _ := dynCtx.GetResUGCSeason(dynCtx.Dyn.UID)
	ap, _ := dynCtx.GetArchive(dynCtx.Dyn.Rid)
	var archive = ap.Arc
	userInfo, ok := dynCtx.GetUser(archive.Author.Mid)
	var userMdl *api.Module_ModuleAuthor
	if !ok {
		userMdl = &api.Module_ModuleAuthor{
			ModuleAuthor: &api.ModuleAuthor{
				Author: &api.UserInfo{
					Mid:  dynCtx.Dyn.UID,
					Face: s.c.Resource.Icon.ModuleAuthorDefaultFace,
				},
			},
		}
		goto END
	}
	userMdl = &api.Module_ModuleAuthor{
		ModuleAuthor: &api.ModuleAuthor{
			Author: &api.UserInfo{
				Mid:  dynCtx.Dyn.UID,
				Name: ugcSeason.Title,
				Face: ugcSeason.Cover,
				Official: &api.OfficialVerify{
					Type: int32(userInfo.Official.Type),
					Desc: userInfo.Official.Desc,
				},
			},
		},
	}
END:
	userMdl.ModuleAuthor.PtimeLabelText = s.publishLabel(c, dynCtx, general)
	userMdl.ModuleAuthor.Weight = s.weight(c, dynCtx, general)
	// 如果提权没有则展示三点
	if userMdl.ModuleAuthor.Weight == nil || len(userMdl.ModuleAuthor.Weight.Items) == 0 {
		userMdl.ModuleAuthor.TpList = s.threePoint(c, dynCtx, general)
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author,
		ModuleItem: userMdl,
	}
	dynCtx.Modules = append(dynCtx.Modules, module)
	return nil
}

func (s *Service) authorNewTopicSet(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	tps := dynCtx.GetResNewTopicSet()
	if tps == nil {
		// 主动丢弃卡片
		dynCtx.Interim.IsPassCard = true
		return nil
	}
	userMdl := &api.Module_ModuleAuthor{
		ModuleAuthor: &api.ModuleAuthor{
			Mid:            tps.SetInfo.GetBasicInfo().GetSetId(),
			PtimeLabelText: s.publishLabel(c, dynCtx, general),
			Author: &api.UserInfo{
				Mid:  tps.SetInfo.GetBasicInfo().GetSetId(),
				Name: tps.SetInfo.GetBasicInfo().GetSetName(),
				Face: tps.SetInfo.GetIconUrl(),
				Uri:  tps.SetInfo.GetBasicInfo().GetJumpUrl(),
			},
			Uri: tps.SetInfo.GetBasicInfo().GetJumpUrl(),
			// 只有一个取消订阅的按钮
			TpList: []*api.ThreePointItem{
				{
					Type: api.ThreePointType_topic_set_cancel,
					Item: &api.ThreePointItem_Default{
						Default: &api.ThreePointDefault{
							Icon:  s.c.Resource.Icon.ThreePointCampusDel, // 复用通用的删除Icon
							Title: "取消订阅",
							Id:    strconv.FormatInt(tps.SetInfo.GetBasicInfo().GetSetId(), 10),
						},
					},
				},
			},
		},
	}
	dynCtx.Modules = append(dynCtx.Modules, &api.Module{
		ModuleType: api.DynModuleType_module_author,
		ModuleItem: userMdl,
	})
	return nil
}

func (s *Service) publishLabel(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) string {
	// 订阅卡特殊逻辑
	if dynCtx.Dyn.IsSubscription() {
		if sub, ok := dynCtx.GetResSub(dynCtx.Dyn.Rid); ok {
			return sub.Tips
		}
		return ""
	}
	// 起飞广告
	if dynCtx.Dyn.IsAD() &&
		dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.AdContentType == _adContentFly && dynCtx.Dyn.PassThrough.AdAvid > 0 {
		return ""
	}

	// 前半部分
	var labels []string
	switch {
	case dynCtx.Dyn.IsLiveRcmd():
	case dynCtx.Dyn.IsSubscriptionNew():
		subNew, _ := dynCtx.GetResSubNew(dynCtx.Dyn.Rid)
		if subNew.Type == submdl.TunnelTypeLive {
			if dynCtx.Dyn.Property == nil ||
				dynCtx.Dyn.Property.RcmdType != dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_ESPORTS_RESERVE {
				labels = append(labels, s.c.Resource.Text.ModuleAuthorPublishLabelSubscriptionNewLive)
			}
		}
	default:
		if dynCtx.From == _handleTypeView {
			if label1 := s.publishDataDetail(general.LocalTime, dynCtx.Dyn.Timestamp); label1 != "" {
				labels = append(labels, label1)
			}
		} else if label1 := s.publishData(general.LocalTime, dynCtx.Dyn.Timestamp); label1 != "" {
			labels = append(labels, label1)
		}
	}
	// 后半部分
	if label2 := s.publishSuffix(c, general, dynCtx); label2 != "" {
		labels = append(labels, label2)
	}
	// 首映
	if premiereText := s.publishPremiere(c, general, dynCtx); premiereText != "" {
		labels = []string{premiereText}
	}
	return strings.Join(labels, " · ")
}

func (s *Service) publishData(localTimeZone int32, timestamp int64) string {
	// 计算时区差值(默认服务端固定东八区)
	// 与客户端约定：东一至东十二区分别1到12; 0时区0; 西一至西十一分别-1到-11
	dd, _ := time.ParseDuration(fmt.Sprintf("%dh", localTimeZone-8))
	t := time.Unix(timestamp, 0)
	// 同步平移时间
	now := time.Now().Add(dd)
	sub := now.Sub(t.Add(dd))
	// 文案格式化
	if sub < time.Minute {
		return "刚刚"
	}
	if sub < time.Hour {
		return fmt.Sprintf("%v分钟前", math.Floor(sub.Minutes()))
	}
	if sub < 24*time.Hour {
		return fmt.Sprintf("%v小时前", math.Floor(sub.Hours()))
	}
	if now.Year() == t.Add(dd).Year() {
		if now.YearDay()-t.Add(dd).YearDay() == 1 {
			return "昨天"
		}
		return t.Add(dd).Format("01-02")
	}
	return t.Add(dd).Format("2006-01-02")
}

func (s *Service) publishDataDetail(localTimeZone int32, timestamp int64) string {
	// 计算时区差值(默认服务端固定东八区)
	// 与客户端约定：东一至东十二区分别1到12; 0时区0; 西一至西十一分别-1到-11
	dd, _ := time.ParseDuration(fmt.Sprintf("%dh", localTimeZone-8))
	t := time.Unix(timestamp, 0)
	return t.Add(dd).Format("2006-01-02 15:04")
}

// nolint:gocognit
func (s *Service) publishSuffix(c context.Context, general *mdlv2.GeneralParam, dynCtx *mdlv2.DynamicContext) string {
	if dynCtx.From == _handleTypeView {
		return model.UpStatString(dynCtx.Dyn.ViewNum, "浏览")
	}
	if dynCtx.Dyn.IsAv() {
		if dynCtx.Dyn.Property != nil && dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_ARCHIVE {
			return "预约的视频"
		}
		if dynCtx.Dyn.Property != nil && dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE_PLAY_BACK {
			return "预约的直播"
		}
		if ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid); ok {
			var archive = ap.Arc
			if archive.Rights.IsCooperation == 1 {
				return "与他人联合创作"
			}
		}
		switch dynCtx.Dyn.SType {
		case mdlv2.VideoStypeDynamic, mdlv2.VideoStypeDynamicStory:
			// 新版本跳转到story页面里面
			if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynStory, &feature.OriginResutl{
				BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynStoryIOS) ||
					(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynStoryAndroid)}) {
				return "发布了动态视频"
			}
			return "发布了动态"
		case mdlv2.VideoStypePlayback:
			return "投稿了直播回放"
		}
		return "投稿了视频"
	}
	if dynCtx.Dyn.IsPGC() || dynCtx.Dyn.IsBatch() {
		return "更新了"
	}
	if dynCtx.Dyn.IsCourUp() {
		if dynCtx.Dyn.Property.GetRcmdType() == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_CHEESE {
			return "预约的课程"
		}
		return "发布了课程"
	}
	if dynCtx.Dyn.IsCourse() {
		return "更新了课程"
	}
	if dynCtx.Dyn.IsArticle() {
		return "投稿了文章"
	}
	if dynCtx.Dyn.IsMusic() {
		return "投稿了音频"
	}
	if dynCtx.Dyn.IsLiveRcmd() {
		if dynCtx.Dyn.Property != nil && (dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE || dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE_HISTORY) {
			return "预约的直播"
		}
		return "直播了"
	}
	if dynCtx.Dyn.IsAD() &&
		dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.AdContentType == _adContentFly && dynCtx.Dyn.PassThrough.AdAvid > 0 {
		return "投稿了视频"
	}
	if dynCtx.Dyn.IsUGCSeason() {
		ugcSeason, _ := dynCtx.GetResUGCSeason(dynCtx.Dyn.UID)
		if userInfo, ok := dynCtx.GetUser(ugcSeason.Mid); ok {
			return fmt.Sprintf("%v更新了合集", userInfo.Name)
		}
	}
	if dynCtx.Dyn.IsSubscriptionNew() {
		// 赛事召回卡
		if dynCtx.Dyn.Property != nil &&
			dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_ESPORTS_RESERVE {
			return s.c.Resource.Text.ModuleAuthorPublishLabelSubscriptionNewESports
		}
		return s.c.Resource.Text.ModuleAuthorPublishLabelSubscriptionNewSuffix
	}
	// 新话题-话题集订阅卡
	if dynCtx.Dyn.IsNewTopicSet() {
		return "来自我的话题集订阅"
	}
	return ""
}

// 首映状态
func (s *Service) publishPremiere(_ context.Context, general *mdlv2.GeneralParam, dynCtx *mdlv2.DynamicContext) string {
	if general.IsIPhonePick() && general.GetBuild() < s.c.BuildLimit.DynPropertyIOS || general.IsAndroidPick() && general.GetBuild() < s.c.BuildLimit.DynPropertyAndroid || general.IsPad() || general.IsPadHD() || general.IsAndroidHD() {
		return ""
	}
	// 首映召回卡 首映开始
	if dynCtx.Dyn.Property != nil && dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_PREMIERE_RESERVE {
		ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid)
		if !ok {
			return ""
		}
		var archive = ap.Arc
		if archive.Premiere == nil {
			return ""
		}
		switch archive.Premiere.State {
		case arcgrpc.PremiereState_premiere_in:
			return "首映开始了"
		default:
			return ""
		}
	}
	// 首映中
	if s.isPremiere(dynCtx, general) == arcgrpc.PremiereState_premiere_in {
		return "首映开始了"
	}
	return ""
}

// nolint:gocognit
func (s *Service) threePoint(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) []*api.ThreePointItem {
	var (
		ext        []*api.ThreePointItem
		isFakeCard bool
	)
	switch dynCtx.From {
	case _handleTypeFake:
		isFakeCard = true
	case _handleTypeView, _handleTypeRepost, _handleTypeUnLogin:
		// 详情页、转发列表、点赞列表，不需要三点，未登录页
		return nil
	case _handleTypeLight:
		// 轻浏览只要举报
		if isReport, titles, reportMid := s.threePointReport(dynCtx, general); isReport {
			ext = append(ext, s.tpReport(c, dynCtx.Dyn.DynamicID, reportMid, titles))
		}
		return ext
	case _handleTypeSpace, _handleTypeSpaceSearchDetail:
		// 空间增加置顶
		if s.threePointTop(dynCtx, general) {
			ext = append(ext, s.tpTop(dynCtx, s.c.Resource.Icon.ThreePointTopIcon, s.c.Resource.Icon.ThreePointTopCannlIcon, s.c.Resource.Text.ThreePointTopText, s.c.Resource.Text.ThreePointTopCannlText))
		}
	}
	// 下发校园三点 - 移除该内容
	if s.threePointCampus(c, dynCtx, general, false) {
		ext = append(ext, s.tpCampusDel(c, dynCtx))
	}
	// 校友反馈
	if s.threePointCampus(c, dynCtx, general, true) {
		ext = append(ext, s.tpCampusFeedback(c, dynCtx))
	}
	if !isFakeCard {
		// 校园一起不展示不感兴趣
		if dynCtx.From != _handleTypeSchool && dynCtx.From != _handleTypeSpace &&
			dynCtx.Dyn.Property.RcmdType != dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE && dynCtx.Dyn.Property.RcmdType != dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE_HISTORY {
			// 不敢兴趣
			if isDislike := s.threePointDislike(dynCtx, general); isDislike {
				ext = append(ext, s.tpDislike(c))
			}
		}
		// 稍后在看
		if iswait, waitID := s.threePointWait(dynCtx, general); iswait {
			ext = append(ext, s.tpWait(s.c.Resource.Text.ThreePointWaitAddition, s.c.Resource.Text.ThreePointWaitNotAddition, s.c.Resource.Icon.ThreePointWait, waitID))
		}
		if !general.CloseAutoPlay {
			// 自动播放
			isAutoPlay, openText, closeText := s.threePointAutoPlay(dynCtx, general)
			if isAutoPlay {
				ext = append(ext, s.tpAutoPlay(c, openText, closeText))
			}
			// 首映自动播放
			if general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynPropertyIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynPropertyAndroid {
				if isPremiereAutoPlay, openText, closeText := s.threePointPremiereAutoPlay(dynCtx, general); isPremiereAutoPlay && !isAutoPlay {
					ext = append(ext, s.tpAutoPlay(c, openText, closeText))
				}
			}
		}
	}
	// 背景
	if isDecorate, backgroundURL := s.threePointDecorate(dynCtx, general); isDecorate {
		ext = append(ext, s.tpBackground(c, backgroundURL))
	}
	// 分享
	if isShare := s.threePointShare(dynCtx, general); isShare {
		ext = append(ext, s.tpShare(c))
	}
	// 取消追漫
	if dynCtx.Dyn.IsBatch() {
		ext = append(ext, s.tpBatchCancel(c, dynCtx))
	}
	if !isFakeCard {
		// 学校页不展示取消关注
		if dynCtx.From != _handleTypeSchool && dynCtx.From != _handleTypeSchoolTopicFeed &&
			// 空间页不展示
			dynCtx.From != _handleTypeSpace &&
			// LBS和老话题页面也不展示
			dynCtx.From != _handleTypeLBS && dynCtx.From != _handleTypeLegacyTopic {
			// 关注/取消关注
			if isFollow := s.threePointFollow(dynCtx, general); isFollow {
				ext = append(ext, s.tpAttention(c, dynCtx))
			}
		}
		// 收藏
		if isFav, favID := s.threePointFav(dynCtx, general); isFav {
			ext = append(ext, s.tpFav(c, favID))
		}
		if dynCtx.From != _handleTypeVideoPersonal && dynCtx.From != _handleTypeAllPersonal {
			// 不再显示
			if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynThreeHide, &feature.OriginResutl{
				BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynThreeHideIOS) ||
					(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynThreeHideAndroid)}) {
				if dynCtx.Dyn.Property != nil && (dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE || dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE_HISTORY || dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_ARCHIVE || dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE_PLAY_BACK) {
					ext = append(ext, s.threePointHide(dynCtx))
				}
			}
		}
		// 举报
		if isReport, titles, reportMid := s.threePointReport(dynCtx, general); isReport {
			ext = append(ext, s.tpReport(c, dynCtx.Dyn.DynamicID, reportMid, titles))
		}
	}
	// 删除
	if isDel := s.threePointDel(dynCtx, general); isDel {
		ext = append(ext, s.tpDelete(c, dynCtx))
	}
	return ext
}

// 广告不感兴趣
func (s *Service) threePointAd(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) []*api.ThreePointItem {
	var (
		ext []*api.ThreePointItem
	)
	// 稍后在看
	if iswait, waitID := s.threePointWait(dynCtx, general); iswait {
		ext = append(ext, s.tpWait(s.c.Resource.Text.ThreePointWaitAddition, s.c.Resource.Text.ThreePointWaitNotAddition, s.c.Resource.Icon.ThreePointWait, waitID))
	}
	if !general.CloseAutoPlay {
		// 自动播放
		if isAutoPlay, openText, closeText := s.threePointAutoPlay(dynCtx, general); isAutoPlay {
			ext = append(ext, s.tpAutoPlay(c, openText, closeText))
		}
	}
	// 分享
	if isShare := s.threePointShare(dynCtx, general); isShare {
		ext = append(ext, s.tpShare(c))
	}
	// 关注
	if isFollow := s.threePointFollow(dynCtx, general); isFollow {
		ext = append(ext, s.tpAttention(c, dynCtx))
	}
	// 不敢兴趣
	if isDislike := s.threePointAdDislike(dynCtx, general); isDislike {
		ext = append(ext, s.tpDislike(c))
	}
	return ext
}

func (s *Service) threePointDislike(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) bool {
	if dynCtx.Dyn.IsLiveRcmd() {
		return true
	}
	if dynCtx.Dyn.IsSubscriptionNew() {
		return true
	}
	return false
}

func (s *Service) threePointAdDislike(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) bool {
	if dynCtx.Dyn.IsAD() &&
		dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.AdContentType == _adContentFly && dynCtx.Dyn.PassThrough.AdAvid > 0 {
		return true
	}
	return false
}

func (s *Service) tpDislike(_ context.Context) *api.ThreePointItem {
	item := &api.ThreePointItem{
		Type: api.ThreePointType_dislike,
		Item: &api.ThreePointItem_Dislike{
			Dislike: &api.ThreePointDislike{
				Icon:  s.c.Resource.Icon.ThreePointDislike,
				Title: s.c.Resource.Text.ThreePointDislike,
			},
		},
	}
	return item
}

func (s *Service) threePointWait(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) (isWait bool, waitID int64) {
	if dynCtx.Dyn.IsAv() {
		ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid)
		if ok && ap.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrNo {
			isWait = true
			var archive = ap.Arc
			waitID = archive.Aid
		}
	}
	if dynCtx.Dyn.IsForward() {
		if dynCtx.Interim.DynTypeKernel == mdlv2.DynTypeVideo || dynCtx.Interim.DynTypeKernel == mdlv2.DynTypeUGCSeason {
			isWait = true
			waitID = dynCtx.Interim.KernelRID
		}
	}
	if dynCtx.Dyn.IsUGCSeason() {
		isWait = true
		waitID = dynCtx.Dyn.Rid
	}
	// 广告起飞卡
	if dynCtx.Dyn.IsAD() &&
		dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.AdContentType == _adContentFly && dynCtx.Dyn.PassThrough.AdAvid > 0 {
		ap, ok := dynCtx.GetArchive(dynCtx.Dyn.PassThrough.AdAvid)
		if ok && ap.Arc.IsNormal() {
			isWait = true
			var archive = ap.Arc
			waitID = archive.Aid
		}
	}
	return
}

func (s *Service) threePointTop(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) bool {
	return dynCtx.From == _handleTypeSpace && general.Mid == dynCtx.Dyn.UID
}

func (s *Service) tpTop(dynCtx *mdlv2.DynamicContext, topIcon, cannelIcon, topText, cannelText string) *api.ThreePointItem {
	topItem := &api.ThreePointTop{
		Icon:  topIcon,
		Title: topText,
		Type:  api.TopType_top_none,
	}
	// 当前已经置顶
	if dynCtx.Dyn.Property != nil && dynCtx.Dyn.Property.IsSpaceTop {
		topItem = &api.ThreePointTop{
			Icon:  cannelIcon,
			Title: cannelText,
			Type:  api.TopType_top_cancel,
		}
	}
	item := &api.ThreePointItem{
		Type: api.ThreePointType_top,
		Item: &api.ThreePointItem_Top{
			Top: topItem,
		},
	}
	return item
}

func (s *Service) tpWait(addition, notAddition, icon string, aid int64) *api.ThreePointItem {
	item := &api.ThreePointItem{
		Type: api.ThreePointType_wait,
		Item: &api.ThreePointItem_Wait{
			Wait: &api.ThreePointWait{
				AdditionIcon:   icon,
				AdditionText:   addition,
				NoAdditionIcon: icon,
				NoAdditionText: notAddition,
				Id:             aid,
			},
		},
	}
	return item
}

func (s *Service) threePointAutoPlay(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) (isAutoPlay bool, open, close string) {
	if dynCtx.Dyn.IsAv() || dynCtx.Dyn.IsUGCSeason() {
		isAutoPlay = true
		open = s.c.Resource.Text.ThreePointAutoPlayOpenV1
		close = s.c.Resource.Text.ThreePointAutoPlayCloseV1
		if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
			open = s.c.Resource.Text.ThreePointAutoPlayOpenIPADV1
			close = s.c.Resource.Text.ThreePointAutoPlayCloseIPADV1
		}
	}
	if dynCtx.Dyn.IsPGC() {
		isAutoPlay = true
		open = s.c.Resource.Text.ThreePointAutoPlayOpenV1
		close = s.c.Resource.Text.ThreePointAutoPlayCloseV1
		if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
			open = s.c.Resource.Text.ThreePointAutoPlayOpenIPADV1
			close = s.c.Resource.Text.ThreePointAutoPlayCloseIPADV1
		}
	}
	if dynCtx.Dyn.IsLiveRcmd() {
		isAutoPlay = true
		open = s.c.Resource.Text.ThreePointAutoPlayOpenV1
		close = s.c.Resource.Text.ThreePointAutoPlayCloseV1
	}
	// 转发卡原卡片为UGC、PGC、直播大卡时，添加自动播放按钮
	if dynCtx.Dyn.IsForward() {
		if dynCtx.Interim.DynTypeKernel == mdlv2.DynTypeVideo {
			isAutoPlay = true
			open = s.c.Resource.Text.ThreePointAutoPlayOpenV1
			close = s.c.Resource.Text.ThreePointAutoPlayCloseV1
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				open = s.c.Resource.Text.ThreePointAutoPlayOpenIPADV1
				close = s.c.Resource.Text.ThreePointAutoPlayCloseIPADV1
			}
		}
		if dynCtx.Interim.DynTypeKernel == mdlv2.DynTypePGCBangumi ||
			dynCtx.Interim.DynTypeKernel == mdlv2.DynTypePGCMovie ||
			dynCtx.Interim.DynTypeKernel == mdlv2.DynTypePGCTv ||
			dynCtx.Interim.DynTypeKernel == mdlv2.DynTypePGCGuoChuang ||
			dynCtx.Interim.DynTypeKernel == mdlv2.DynTypePGCDocumentary ||
			dynCtx.Interim.DynTypeKernel == mdlv2.DynTypeBangumi {
			isAutoPlay = true
			open = s.c.Resource.Text.ThreePointAutoPlayOpenV1
			close = s.c.Resource.Text.ThreePointAutoPlayCloseV1
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				open = s.c.Resource.Text.ThreePointAutoPlayOpenIPADV1
				close = s.c.Resource.Text.ThreePointAutoPlayCloseIPADV1
			}
		}
		if dynCtx.Interim.DynTypeKernel == mdlv2.DynTypeLiveRcmd {
			isAutoPlay = true
			open = s.c.Resource.Text.ThreePointAutoPlayOpenV1
			close = s.c.Resource.Text.ThreePointAutoPlayCloseV1
		}
		if dynCtx.Interim.DynTypeKernel == mdlv2.DynTypeUGCSeason {
			isAutoPlay = true
			open = s.c.Resource.Text.ThreePointAutoPlayOpenV1
			close = s.c.Resource.Text.ThreePointAutoPlayCloseV1
		}
	}
	if dynCtx.Dyn.IsUGCSeason() {
		isAutoPlay = true
		open = s.c.Resource.Text.ThreePointAutoPlayOpenV1
		close = s.c.Resource.Text.ThreePointAutoPlayCloseV1
	}
	if dynCtx.Dyn.IsSubscriptionNew() {
		subNew, _ := dynCtx.GetResSubNew(dynCtx.Dyn.Rid)
		if subNew.Type == submdl.TunnelTypeLive {
			isAutoPlay = true
			open = s.c.Resource.Text.ThreePointAutoPlayOpenV1
			close = s.c.Resource.Text.ThreePointAutoPlayCloseV1
		}
	}
	// 广告起飞卡
	if dynCtx.Dyn.IsAD() &&
		dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.AdContentType == _adContentFly && dynCtx.Dyn.PassThrough.AdAvid > 0 {
		isAutoPlay = true
		open = s.c.Resource.Text.ThreePointAutoPlayOpenV1
		close = s.c.Resource.Text.ThreePointAutoPlayCloseV1
	}
	return
}

func (s *Service) threePointPremiereAutoPlay(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) (isAutoPlay bool, open, close string) {
	tmpDyn := &mdlv2.Dynamic{}
	if dynCtx.Dyn.Origin != nil {
		*tmpDyn = *dynCtx.Dyn.Origin
	} else {
		*tmpDyn = *dynCtx.Dyn
	}
	for _, v := range tmpDyn.AttachCardInfos {
		// nolint:exhaustive
		switch v.CardType {
		case dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:
			up, ok := dynCtx.ResUpActRelationInfo[v.Rid]
			if !ok {
				continue
			}
			// 非首映的直接抛弃
			if up.Type != activitygrpc.UpActReserveRelationType_Premiere {
				continue
			}
			aid, _ := strconv.ParseInt(up.Oid, 10, 64)
			ap, ok := dynCtx.GetArchive(aid)
			if !ok {
				continue
			}
			var archive = ap.Arc
			// 不是首映前
			if archive.Premiere != nil && archive.Premiere.State != arcgrpc.PremiereState_premiere_before {
				isAutoPlay = true
				open = s.c.Resource.Text.ThreePointAutoPlayOpenV1
				close = s.c.Resource.Text.ThreePointAutoPlayCloseV1
				if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
					open = s.c.Resource.Text.ThreePointAutoPlayOpenIPADV1
					close = s.c.Resource.Text.ThreePointAutoPlayCloseIPADV1
				}
			}
		}
	}
	return
}

func (s *Service) tpAutoPlay(_ context.Context, open, close string) *api.ThreePointItem {
	item := &api.ThreePointItem{
		Type: api.ThreePointType_auto_play,
		Item: &api.ThreePointItem_AutoPlayer{
			AutoPlayer: &api.ThreePointAutoPlay{
				OpenIcon:  s.c.Resource.Icon.ThreePointAutoPlayClose,
				OpenText:  open,
				CloseIcon: s.c.Resource.Icon.ThreePointAutoPlayOpen,
				CloseText: close,
				// v2
				OpenIconV2:  s.c.Resource.Icon.ThreePointAutoPlayClose,
				OpenTextV2:  s.c.Resource.Text.ThreePointAutoPlayOpenV2,
				CloseIconV2: s.c.Resource.Icon.ThreePointAutoPlayOpen,
				CloseTextV2: s.c.Resource.Text.ThreePointAutoPlayCloseV2,
				OnlyIcon:    s.c.Resource.Icon.ThreePointAutoPlayClose,
				OnlyText:    s.c.Resource.Text.ThreePointAutoPlayOnly,
			},
		},
	}
	return item
}

func (s *Service) threePointDecorate(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) (isDecorate bool, backgroundURL string) {
	// 空间直播不下发 "使用此装扮"
	if dynCtx.From == _handleTypeSpace && general.DynFrom == _dynFromLive {
		return false, ""
	}
	// 新订阅卡不出 "使用此装扮"
	if dynCtx.Dyn.IsSubscriptionNew() {
		return false, ""
	}
	if dynCtx.ResMyDecorate != nil {
		if decoInfo, ok := dynCtx.ResMyDecorate[dynCtx.Dyn.UID]; ok && decoInfo != nil {
			isDecorate = true
			backgroundURL = decoInfo.JumpURL
		}
	}
	return
}

func (s *Service) tpBackground(_ context.Context, uri string) *api.ThreePointItem {
	item := &api.ThreePointItem{
		Type: api.ThreePointType_background,
		Item: &api.ThreePointItem_Default{
			Default: &api.ThreePointDefault{
				Icon:  s.c.Resource.Icon.ThreePointBackground,
				Title: s.c.Resource.Text.ThreePointBackground,
				Uri:   uri,
			},
		},
	}
	return item
}

func (s *Service) threePointShare(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) bool {
	if !dynCtx.Dyn.IsCheeseBatch() && !dynCtx.Dyn.IsSubscriptionNew() {
		return true
	}
	return false
}

func (s *Service) tpShare(_ context.Context) *api.ThreePointItem {
	item := &api.ThreePointItem{
		Type: api.ThreePointType_share,
		Item: &api.ThreePointItem_Share{
			Share: &api.ThreePointShare{
				Icon:  s.c.Resource.Icon.ThreePointShare,
				Title: s.c.Resource.Text.ThreePointShare,
			},
		},
	}
	return item
}

func (s *Service) threePointFollow(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) (isFollow bool) {
	if dynCtx.Dyn.UID == general.Mid {
		return false
	}
	switch {
	case dynCtx.Dyn.IsPGC():
		return false
	case dynCtx.Dyn.IsCourse():
		return false
	case dynCtx.Dyn.IsUGCSeason():
		return false
	case dynCtx.Dyn.IsSubscriptionNew():
		return false
	case dynCtx.Dyn.IsBatch():
		return false
	case dynCtx.Dyn.IsAD():
		if dynCtx.Dyn.PassThrough == nil || dynCtx.Dyn.PassThrough.AdContentType != _adContentFly || dynCtx.Dyn.PassThrough.AdAvid == 0 ||
			dynCtx.Dyn.PassThrough.AdverMid == general.Mid {
			return false
		}
		if _, ok := dynCtx.GetUser(dynCtx.Dyn.PassThrough.AdverMid); !ok {
			return false
		}
	default:
		if _, ok := dynCtx.GetUser(dynCtx.Dyn.UID); !ok {
			return false
		}
	}
	return true
}

func (s *Service) tpAttention(_ context.Context, dynCtx *mdlv2.DynamicContext) *api.ThreePointItem {
	var rel int32
	if dynCtx.ResRelation != nil {
		rel = dynCtx.ResRelation[dynCtx.Dyn.UID]
	}
	item := &api.ThreePointItem{
		Type: api.ThreePointType_attention,
		Item: &api.ThreePointItem_Attention{
			Attention: &api.ThreePointAttention{
				AttentionIcon:    s.c.Resource.Icon.ThreePointFollow,
				AttentionText:    s.c.Resource.Text.ThreePointFollow,
				NotAttentionIcon: s.c.Resource.Icon.ThreePointFollowCancel,
				NotAttentionText: s.c.Resource.Text.ThreePointFollowCancel,
				Status:           api.ThreePointAttentionStatus(rel),
			},
		},
	}
	return item
}

func (s *Service) threePointReport(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) (isReport bool, titles []string, reportMid int64) {
	if dynCtx.Dyn.IsPGC() || dynCtx.Dyn.IsBatch() || dynCtx.Dyn.IsCourUp() {
		return
	}
	if dynCtx.Dyn.UID == general.Mid {
		return
	}
	if dynCtx.Interim.UName != "" {
		titles = append(titles, dynCtx.Interim.UName)
	}
	var reportText string
	isReport = true
	reportMid = dynCtx.Dyn.UID
	switch {
	case dynCtx.Dyn.IsAv():
		reportText = "视频"
	case dynCtx.Dyn.IsDraw():
		reportText = "图片"
	case dynCtx.Dyn.IsArticle():
		reportText = "专栏"
	case dynCtx.Dyn.IsMusic():
		reportText = "音乐"
	case dynCtx.Dyn.IsLive():
		reportText = "直播分享"
	case dynCtx.Dyn.IsUGCSeason():
		ap, _ := dynCtx.GetArchive(dynCtx.Dyn.Rid)
		var archive = ap.Arc
		reportMid = archive.Author.Mid
		reportText = "合集更新了"
	default:
		reportText = "动态"
	}
	if dynCtx.Interim.Desc != "" {
		reportText = dynCtx.Interim.Desc
	}
	titles = append(titles, reportText)
	return
}

func (s *Service) tpReport(_ context.Context, dynid, uid int64, titles []string) *api.ThreePointItem {
	title := model.QueryEscape(strings.Join(titles, ":"))
	item := &api.ThreePointItem{
		Type: api.ThreePointType_report,
		Item: &api.ThreePointItem_Default{
			Default: &api.ThreePointDefault{
				Icon:  s.c.Resource.Icon.ThreePointReport,
				Title: s.c.Resource.Text.ThreePointReport,
				Uri:   fmt.Sprintf("bilibili://following/report?dynamicId=%v&uid=%v&title=%v", dynid, uid, title),
			},
		},
	}
	return item
}

// 判断是否要下发校园三点
func (s *Service) threePointCampus(ctx context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam, isFeedback bool) (isCampusTP bool) {
	if (dynCtx.From != _handleTypeSchool && dynCtx.From != _handleTypeSchoolTopicFeed) ||
		dynCtx.ResUserProfileStat == nil || dynCtx.ResUserProfileStat[general.Mid] == nil {
		return false
	}
	if currentProfile := dynCtx.ResUserProfileStat[general.Mid]; isFeedback && currentProfile.GetSchool().GetSchoolId() != dynCtx.CampusID {
		return false
	}
	if feature.GetBuildLimit(ctx, s.c.Feature.FeatureBuildLimit.DynSchoolThreePoint, &feature.OriginResutl{
		BuildLimit: general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, s.c.BuildLimit.DynSchoolThreePointAndroid, s.c.BuildLimit.DynSchoolThreePointIOS),
	}) {
		return true
	}
	return false
}

// 校园 - 校友反馈三点入口
func (s *Service) tpCampusFeedback(_ context.Context, dynCtx *mdlv2.DynamicContext) *api.ThreePointItem {
	upMid := dynCtx.Dyn.UID
	upName := "校友UP"
	if up, ok := dynCtx.GetUser(upMid); ok {
		upName = up.Name
	}
	const _maxFeedbackTextLen = 20
	// 保证截断不会出现问题
	desc := []rune(dynCtx.Interim.Desc)
	finDesc := ""
	if len(desc) == 0 {
		switch {
		case dynCtx.Dyn.IsAv():
			finDesc = "视频"
		case dynCtx.Dyn.IsDraw():
			finDesc = "图片"
		case dynCtx.Dyn.IsArticle():
			finDesc = "专栏"
		case dynCtx.Dyn.IsMusic():
			finDesc = "音乐"
		case dynCtx.Dyn.IsLive():
			finDesc = "直播分享"
		case dynCtx.Dyn.IsUGCSeason():
			finDesc = "合集更新了"
		default:
			finDesc = "动态"
		}
	} else if len(desc) > _maxFeedbackTextLen {
		finDesc = string(append(desc[:_maxFeedbackTextLen], []rune("...")...))
	} else {
		finDesc = string(desc)
	}
	dynamicID := dynCtx.Dyn.DynamicID
	// 默认校友圈反馈
	from := 0
	if dynCtx.From == _handleTypeSchoolTopicFeed {
		// 标注来自话题反馈 详情见CampusFeedBack接口pb定义
		from = 2
	}
	params := fmt.Sprintf("report_text=%s&dynamic_id=%d&up_id=%d&campus_id=%d&from=%d",
		model.QueryEscape(upName+":"+finDesc), dynamicID, upMid, dynCtx.CampusID, from)

	return &api.ThreePointItem{
		Type: api.ThreePointType_report,
		Item: &api.ThreePointItem_Default{
			Default: &api.ThreePointDefault{
				Icon:  s.c.Resource.Icon.ThreePointCampusFeedback,
				Title: s.c.Resource.Text.ThreePointCampusFeedback,
				Uri:   "bilibili://campus/alumnae_feedback?" + params,
			},
		},
	}
}

// 校园 - 移除该内容
func (s *Service) tpCampusDel(_ context.Context, dynCtx *mdlv2.DynamicContext) *api.ThreePointItem {
	from := 0
	if dynCtx.From == _handleTypeSchoolTopicFeed {
		from = 2
	}
	return &api.ThreePointItem{
		Type: api.ThreePointType_campus_delete,
		Item: &api.ThreePointItem_Default{
			Default: &api.ThreePointDefault{
				Icon:  s.c.Resource.Icon.ThreePointCampusDel,
				Title: s.c.Resource.Text.ThreePointCampusDelText,
				Toast: &api.ThreePointDefaultToast{
					Desc: s.c.Resource.Text.ThreePointCampusDelToast,
				},
				Id: fmt.Sprintf("dynamic_id=%d&up_id=%d&from=%d", dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID, from),
			},
		},
	}
}

func (s *Service) threePointDel(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) bool {
	if general.Mid != dynCtx.Dyn.UID {
		return false
	}
	if dynCtx.Dyn.IsAv() || dynCtx.Dyn.IsCheeseBatch() || dynCtx.Dyn.IsArticle() || dynCtx.Dyn.IsMusic() || dynCtx.Dyn.IsUGCSeason() || dynCtx.Dyn.IsSubscriptionNew() || dynCtx.Dyn.IsBatch() || dynCtx.Dyn.IsCourUp() {
		return false
	}
	if s.isPremiere(dynCtx, general) != arcgrpc.PremiereState_premiere_none {
		return false
	}
	return true
}

func (*Service) threePointEdit(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) *api.ThreePointItem {
	if dynCtx.Dyn.Property.GetEditable() {
		dynEdit := &api.ThreePointDynEdit{
			DynId: dynCtx.Dyn.DynamicID,
		}
		//uri := fmt.Sprintf("bilibili://following/publish?dynamicId=%d", dynCtx.Dyn.DynamicID)
		if dynCtx.Dyn.IsForward() && dynCtx.Dyn.Origin != nil {
			//uri = fmt.Sprintf("bilibili://following/publish/share?key_repost=true&dynamicId=%d&repostSrc=%d", dynCtx.Dyn.DynamicID, dynCtx.Dyn.Origin.DynamicID)
			dynEdit.OriginId = dynCtx.Dyn.Origin.DynamicID
			dynEdit.IsOriginDeleted = !dynCtx.Dyn.Origin.Visible
		}
		return &api.ThreePointItem{
			Type: api.ThreePointType_dynamic_edit,
			Item: &api.ThreePointItem_DynEdit{
				// 目前这块跳转不走路由，客户端写死 所以相关资源也不用下发
				DynEdit: dynEdit,
			},
		}
	}
	return nil
}

// 首映状态
func (s *Service) isPremiere(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) arcgrpc.PremiereState {
	// 首映
	for _, v := range dynCtx.Dyn.AttachCardInfos {
		// nolint:exhaustive
		switch v.CardType {
		case dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:
			up, ok := dynCtx.ResUpActRelationInfo[v.Rid]
			if !ok {
				continue
			}
			// 非首映的直接抛弃
			if up.Type != activitygrpc.UpActReserveRelationType_Premiere {
				continue
			}
			aid, _ := strconv.ParseInt(up.Oid, 10, 64)
			ap, ok := dynCtx.GetArchive(aid)
			if !ok {
				continue
			}
			var archive = ap.Arc
			if archive.Premiere == nil {
				continue
			}
			return archive.Premiere.State
		}
	}
	return arcgrpc.PremiereState_premiere_none
}

func (s *Service) tpDelete(_ context.Context, dynCtx *mdlv2.DynamicContext) *api.ThreePointItem {
	item := &api.ThreePointItem{
		Type: api.ThreePointType_delete,
		Item: &api.ThreePointItem_Default{
			Default: &api.ThreePointDefault{
				Icon:  s.c.Resource.Icon.ThreePointDeleted,
				Title: s.c.Resource.Text.ThreePointDeleted,
				Toast: s.tpDeleteReserveToast(dynCtx),
			},
		},
	}
	return item
}

func (s *Service) threePointFav(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) (bool, int64) {
	if dynCtx.Dyn.IsUGCSeason() {
		return true, dynCtx.Dyn.UID
	}
	return false, 0
}

func (s *Service) tpFav(_ context.Context, favID int64) *api.ThreePointItem {
	item := &api.ThreePointItem{
		Type: api.ThreePointType_favorite,
		Item: &api.ThreePointItem_Favorite{
			Favorite: &api.ThreePointFavorite{
				Icon:        s.c.Resource.Icon.ThreePointFav,
				Title:       s.c.Resource.Text.ThreePointFav,
				Id:          favID,
				IsFavourite: true,
				CancelIcon:  s.c.Resource.Icon.ThreePointFav,
				CancelTitle: s.c.Resource.Text.ThreePointCancelFav,
			},
		},
	}
	return item
}

func (s *Service) tpBatchCancel(_ context.Context, dynCtx *mdlv2.DynamicContext) *api.ThreePointItem {
	item := &api.ThreePointItem{
		Type: api.ThreePointType_batch_cancel,
		Item: &api.ThreePointItem_Default{
			Default: &api.ThreePointDefault{
				Id:    strconv.FormatInt(dynCtx.Dyn.UID, 10),
				Icon:  s.c.Resource.Icon.ThreePointBatchCancel,
				Title: s.c.Resource.Text.ThreePointBatchCancel,
				Toast: &api.ThreePointDefaultToast{
					Desc: s.c.Resource.Text.ThreePointBatchCancelDesc,
				},
			},
		},
	}
	return item
}

func (s *Service) authorShell(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID)
	if !ok {
		xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "forward_author", "date_faild")
		log.Warn("module error mid(%v) dynid(%v) authorShell uid(%d)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID)
		return nil
	}
	var titles []*api.ModuleAuthorForwardTitle
	titles = append(titles, &api.ModuleAuthorForwardTitle{
		Text: fmt.Sprintf("@%s", userInfo.Name),
		Url:  model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(userInfo.Mid, 10), nil),
	})
	userMdl := &api.Module_ModuleAuthorForward{
		ModuleAuthorForward: &api.ModuleAuthorForward{
			Title:          titles,
			Uid:            userInfo.Mid,
			PtimeLabelText: s.publishLabel(c, dynCtx, general),
			FaceUrl:        userInfo.Face,
			ShowFollow:     s.showFollow(dynCtx, general),
			Relation:       relationmdl.RelationChange(dynCtx.Dyn.UID, dynCtx.ResRelationUltima),
		},
	}
	userMdl.ModuleAuthorForward.TpList = s.threePoint(c, dynCtx, general)
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author_forward,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) authorShellPGC(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	pgc, _ := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid))
	var (
		titles []*api.ModuleAuthorForwardTitle
		title  string
	)
	if pgc.Season != nil {
		title = pgc.Season.Title
	} else {
		log.Warn("GetResPGC season nil mid %v, dynid %v, rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
	}
	titles = append(titles, &api.ModuleAuthorForwardTitle{
		Text: title, // PGC内容不@
		Url:  pgc.Url,
	})
	userMdl := &api.Module_ModuleAuthorForward{
		ModuleAuthorForward: &api.ModuleAuthorForward{
			Uid:            dynCtx.Dyn.UID,
			Title:          titles,
			PtimeLabelText: s.publishLabel(c, dynCtx, general),
			ShowFollow:     s.showFollow(dynCtx, general),
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author_forward,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) authorShellBatch(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	batch, ok := dynCtx.ResBatch[dynCtx.Dyn.Rid]
	var (
		titles []*api.ModuleAuthorForwardTitle
	)
	if ok {
		titles = append(titles, &api.ModuleAuthorForwardTitle{
			Text: batch.Name,
			Url:  batch.JumpURL,
		})
	}
	userMdl := &api.Module_ModuleAuthorForward{
		ModuleAuthorForward: &api.ModuleAuthorForward{
			Uid:            dynCtx.Dyn.UID,
			Title:          titles,
			PtimeLabelText: s.publishLabel(c, dynCtx, general),
			ShowFollow:     s.showFollow(dynCtx, general),
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author_forward,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) authorShellCheeseBatch(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	batch, _ := dynCtx.GetResCheeseBatch(dynCtx.Dyn.Rid)
	var titles []*api.ModuleAuthorForwardTitle
	titles = append(titles, &api.ModuleAuthorForwardTitle{
		Text: fmt.Sprintf("@%v", batch.UpInfo.Name),
		Url:  batch.URL,
	})
	userMdl := &api.Module_ModuleAuthorForward{
		ModuleAuthorForward: &api.ModuleAuthorForward{
			Title:          titles,
			Uid:            batch.UpID,
			PtimeLabelText: s.publishLabel(c, dynCtx, general),
			ShowFollow:     s.showFollow(dynCtx, general),
			Relation:       relationmdl.RelationChange(batch.UpID, dynCtx.ResRelationUltima), // 关注组件
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author_forward,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) authorShellCheeseSeason(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	season, _ := dynCtx.GetResCheeseSeason(dynCtx.Dyn.Rid)
	var titles []*api.ModuleAuthorForwardTitle
	titles = append(titles, &api.ModuleAuthorForwardTitle{
		Text: fmt.Sprintf("@%v", season.UpInfo.Name),
		Url:  season.URL,
	})
	userMdl := &api.Module_ModuleAuthorForward{
		ModuleAuthorForward: &api.ModuleAuthorForward{
			Title:          titles,
			Uid:            season.UpID,
			PtimeLabelText: s.publishLabel(c, dynCtx, general),
			ShowFollow:     s.showFollow(dynCtx, general),
			Relation:       relationmdl.RelationChange(season.UpID, dynCtx.ResRelationUltima), // 关注组件
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author_forward,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) authorShellCourUp(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	season, _ := dynCtx.GetResCheeseSeason(dynCtx.Dyn.Rid)
	var titles []*api.ModuleAuthorForwardTitle
	titles = append(titles, &api.ModuleAuthorForwardTitle{
		Text: fmt.Sprintf("@%v", season.UpInfo.Name),
		Url:  season.URL,
	})
	userMdl := &api.Module_ModuleAuthorForward{
		ModuleAuthorForward: &api.ModuleAuthorForward{
			Title:          titles,
			Uid:            season.UpID,
			PtimeLabelText: s.publishLabel(c, dynCtx, general),
			ShowFollow:     s.showFollow(dynCtx, general),
			Relation:       relationmdl.RelationChange(season.UpID, dynCtx.ResRelationUltima), // 关注组件
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author_forward,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) authorCourUp(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	season, ok := dynCtx.GetResCheeseSeason(dynCtx.Dyn.Rid)
	var userMdl *api.Module_ModuleAuthor
	if !ok {
		userMdl = &api.Module_ModuleAuthor{
			ModuleAuthor: &api.ModuleAuthor{
				Author: &api.UserInfo{
					Mid:  dynCtx.Dyn.UID,
					Face: s.c.Resource.Icon.ModuleAuthorDefaultFace,
				},
			},
		}
		goto END
	}
	userMdl = &api.Module_ModuleAuthor{
		ModuleAuthor: &api.ModuleAuthor{
			Author: &api.UserInfo{
				Mid:  season.UpID,
				Name: season.UpInfo.Name,
				Face: season.UpInfo.Avatar,
				Official: &api.OfficialVerify{
					Type: int32(season.UserProfile.Card.OfficialVerify.Type),
					Desc: season.UserProfile.Card.OfficialVerify.Desc,
				},
				Vip: &api.VipInfo{
					Type:    int32(season.UserProfile.Vip.VipType),
					Status:  int32(season.UserProfile.Vip.VipStatus),
					DueDate: season.UserProfile.Vip.VipDueDate,
					Label: &api.VipLabel{
						Path:       season.UserProfile.Vip.Label.Path,
						Text:       season.UserProfile.Vip.Label.Text,
						LabelTheme: season.UserProfile.Vip.Label.LabelTheme,
					},
					ThemeType:       int32(season.UserProfile.Vip.ThemeType),
					AvatarSubscript: season.UserProfile.Vip.AvatarSubscript,
					NicknameColor:   season.UserProfile.Vip.NicknameColor,
				},
				Uri: model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(season.UpID, 10), nil),
				Pendant: &api.UserPendant{
					Pid:    season.UserProfile.Pendant.Pid,
					Name:   season.UserProfile.Pendant.Name,
					Image:  season.UserProfile.Pendant.Image,
					Expire: season.UserProfile.Pendant.Expire,
				},
			},
			Mid:      season.UpID,
			Relation: relationmdl.RelationChange(season.UpID, dynCtx.ResRelationUltima), // 关注组件
		},
	}
END:
	userMdl.ModuleAuthor.PtimeLabelText = s.publishLabel(c, dynCtx, general)
	userMdl.ModuleAuthor.Weight = s.weight(c, dynCtx, general)
	// 如果提权没有则展示三点
	if userMdl.ModuleAuthor.Weight == nil || len(userMdl.ModuleAuthor.Weight.Items) == 0 {
		userMdl.ModuleAuthor.TpList = s.threePoint(c, dynCtx, general)
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) authorShellUGCSeason(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	ugcSeason, _ := dynCtx.GetResUGCSeason(dynCtx.Dyn.UID)
	ap, _ := dynCtx.GetArchive(dynCtx.Dyn.Rid)
	var (
		archive = ap.Arc
		titles  []*api.ModuleAuthorForwardTitle
	)
	titles = append(titles, &api.ModuleAuthorForwardTitle{
		Text: ugcSeason.Title,
		Url:  model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, archive.FirstCid, true)),
	})
	userMdl := &api.Module_ModuleAuthorForward{
		ModuleAuthorForward: &api.ModuleAuthorForward{
			Title:          titles,
			PtimeLabelText: s.publishLabel(c, dynCtx, general),
			ShowFollow:     s.showFollow(dynCtx, general),
			Relation:       relationmdl.RelationChange(archive.Author.Mid, dynCtx.ResRelationUltima),
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author_forward,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) showFollow(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) bool {
	if dynCtx.Dyn.IsPGC() || dynCtx.Dyn.IsUGCSeason() || dynCtx.Dyn.IsBatch() {
		return false
	}
	if general.Mid == dynCtx.Dyn.UID {
		return false
	}
	return true
}

func (s *Service) authorInfo(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID)
	if !ok {
		log.Warn("module miss mid(%v) dynid(%v) author uid(%d)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID)
		return nil
	}
	ptimeLabelText := s.publishLabel(c, dynCtx, general)
	userMdl := &api.Module_ModuleAuthor{
		ModuleAuthor: &api.ModuleAuthor{
			Mid: userInfo.Mid,
			Author: &api.UserInfo{
				Mid:  userInfo.Mid,
				Name: userInfo.Name,
				Face: userInfo.Face,
				Official: &api.OfficialVerify{ // 认证
					Type: int32(userInfo.Official.Type),
					Desc: userInfo.Official.Desc,
				},
				Vip: &api.VipInfo{ // 会员
					Type:    userInfo.Vip.Type,
					Status:  userInfo.Vip.Status,
					DueDate: userInfo.Vip.DueDate,
					Label: &api.VipLabel{
						Path: userInfo.Vip.Label.Path,
					},
					ThemeType: userInfo.Vip.ThemeType,
				},
				Pendant: &api.UserPendant{ // 头像挂件
					Pid:    int64(userInfo.Pendant.Pid),
					Name:   userInfo.Pendant.Name,
					Image:  userInfo.Pendant.Image,
					Expire: int64(userInfo.Pendant.Expire),
				},
				Nameplate: &api.Nameplate{ // 勋章
					Nid:        int64(userInfo.Nameplate.Nid),
					Name:       userInfo.Nameplate.Name,
					Image:      userInfo.Nameplate.Image,
					ImageSmall: userInfo.Nameplate.ImageSmall,
					Level:      userInfo.Nameplate.Level,
					Condition:  userInfo.Nameplate.Condition,
				},
				Uri:   model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(userInfo.Mid, 10), nil),
				Level: userInfo.Level,
			},
			Uri:            dynCtx.Interim.PromoURI, // 帮推
			PtimeLabelText: ptimeLabelText,
		},
	}
	if userInfo.Pendant.ImageEnhance != "" { // 动效图优先
		userMdl.ModuleAuthor.Author.Pendant.Image = userInfo.Pendant.ImageEnhance
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

// 发布人模块
func (s *Service) authorView(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID)
	if !ok {
		log.Warn("module miss mid(%v) dynid(%v) author uid(%d)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID)
	}
	var (
		userMdl        *api.Module_ModuleAuthor
		ptimeLabelText = s.publishLabel(c, dynCtx, general)
	)
	if userInfo == nil { // 兜底 默认返回灰头像、空白文案
		userMdl = &api.Module_ModuleAuthor{
			ModuleAuthor: &api.ModuleAuthor{
				Mid:            dynCtx.Dyn.UID,
				PtimeLabelText: ptimeLabelText,
				Author: &api.UserInfo{
					Mid:  dynCtx.Dyn.UID,
					Face: s.c.Resource.Icon.ModuleAuthorDefaultFace,
					Uri:  model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(dynCtx.Dyn.UID, 10), nil),
				},
				Uri: dynCtx.Interim.PromoURI, // 帮推
			},
		}
		goto END
	}
	userMdl = &api.Module_ModuleAuthor{
		ModuleAuthor: s.authorUser(c, dynCtx.Dyn.UID, dynCtx, general),
	}
	// 未关注、被关注
	if (userMdl.ModuleAuthor.Relation.Status == 1 || userMdl.ModuleAuthor.Relation.Status == 3) && general.Mid != dynCtx.Dyn.UID {
		userMdl.ModuleAuthor.ShowFollow = true // 是否展示关注
	}
	if userInfo.Pendant.ImageEnhance != "" { // 动效图优先
		userMdl.ModuleAuthor.Author.Pendant.Image = userInfo.Pendant.ImageEnhance
	}
	// 直播状态
	// 20201222 加密直播间视为关播
	if dynCtx.From != _handleTypeAllPersonal && dynCtx.From != _handleTypeVideoPersonal && !dynCtx.Interim.HiddenAuthorLive {
		if userLive, ok := dynCtx.GetResUserLive(dynCtx.Dyn.UID); ok && userLive.Status != nil && userLive.Status.Password == "" {
			userMdl.ModuleAuthor.Author.Live = &api.LiveInfo{
				IsLiving:  int32(userLive.Status.LiveStatus),
				LiveState: api.LiveState(userLive.Status.LiveStatus),
				Uri:       model.FillURI(model.GotoLive, strconv.FormatInt(userLive.RoomId, 10), nil),
			}
			if userLivePlayURL, ok := dynCtx.GetResUserLivePlayURL(dynCtx.Dyn.UID); ok {
				userMdl.ModuleAuthor.Author.Live.Uri = userLivePlayURL.Link
			}
		}
	}
	// 订阅卡不展示装扮 测试wanglulu反馈
	// 直播中不展示装扮
	// 新订阅卡不展示装扮
	if !dynCtx.Dyn.IsSubscription() && !dynCtx.Dyn.IsSubscriptionNew() && (userMdl.ModuleAuthor.Author.Live == nil || userMdl.ModuleAuthor.Author.Live.LiveState != api.LiveState_live_live) {
		if dynCtx.ResMyDecorate != nil {
			decoInfo, ok := dynCtx.ResMyDecorate[userInfo.Mid]
			if ok {
				// 小于6位数 前置补0
				numberStr := strconv.Itoa(decoInfo.Fan.Number)
				// nolint:gomnd
				if decoInfo.Fan.Number < 100000 {
					numberStr = fmt.Sprintf("%06d", decoInfo.Fan.Number)
				}
				userMdl.ModuleAuthor.DecorateCard = &api.DecorateCard{
					Id:      decoInfo.ID,
					CardUrl: decoInfo.CardURL,
					JumpUrl: decoInfo.JumpURL,
					Fan: &api.DecoCardFan{
						IsFan:     int32(decoInfo.Fan.IsFan),
						Number:    int32(decoInfo.Fan.Number),
						NumberStr: numberStr,
						Color:     decoInfo.Fan.Color,
					},
				}
				if decoInfo.ImageEnhance != "" {
					userMdl.ModuleAuthor.DecorateCard.CardUrl = decoInfo.ImageEnhance
				}
			}
		}
	}
END:
	// 盘古NFT信息
	if nftInfo, ok := dynCtx.ResNFTBatchInfo[dynCtx.Dyn.UID]; ok {
		if nftRegion, ok := dynCtx.ResNFTRegionInfo[nftInfo.NftId]; ok {
			userMdl.ModuleAuthor.Author.NftInfo = &api.NFTInfo{
				RegionType:       api.NFTRegionType(nftRegion.Type),
				RegionIcon:       nftRegion.Icon,
				RegionShowStatus: api.NFTShowStatus(nftRegion.ShowStatus),
			}
		}
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) weight(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) *api.Weight {
	if (dynCtx.Dyn.PassThrough == nil || (dynCtx.Dyn.Property.RcmdType != dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_TOP_WEIGHT && dynCtx.Dyn.Property.RcmdType != dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_HISTORY_WEIGHT)) || dynCtx.Dyn.PassThrough.FeedBack == nil {
		return nil
	}
	res := &api.Weight{
		Title: dynCtx.Dyn.PassThrough.FeedBack.IconTitle,
		Icon:  s.c.Resource.WeightIcon,
	}
	if dynCtx.Dyn.PassThrough.FeedBack.JumpButtonText != "" {
		res.Items = append(res.Items, &api.WeightItem{
			Type: api.WeightType_weight_jump,
			Item: &api.WeightItem_Button{
				Button: &api.WeightButton{
					JumpUrl: dynCtx.Dyn.PassThrough.FeedBack.JumpUrl,
					Title:   dynCtx.Dyn.PassThrough.FeedBack.JumpButtonText,
				},
			},
		})
	}
	if dynCtx.Dyn.PassThrough.FeedBack.FeedBackButtonText != "" {
		res.Items = append(res.Items, &api.WeightItem{
			Type: api.WeightType_weight_dislike,
			Item: &api.WeightItem_Dislike{
				Dislike: &api.WeightDislike{
					FeedBackType: dynCtx.Dyn.PassThrough.FeedBack.FeedBackBizType,
					Title:        dynCtx.Dyn.PassThrough.FeedBack.FeedBackButtonText,
				},
			},
		})
	}
	return res
}

func (s *Service) tpDeleteReserveToast(dynCtx *mdlv2.DynamicContext) *api.ThreePointDefaultToast {
	for _, v := range dynCtx.Dyn.AttachCardInfos {
		if v.CardType != dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE {
			continue
		}
		up, ok := dynCtx.ResUpActRelationInfo[v.Rid]
		if !ok {
			continue
		}
		if upActState(up.State) != _upStart {
			continue
		}
		return &api.ThreePointDefaultToast{
			Desc: s.c.Resource.Text.ThreePointDelReserveTitle + "\n" + s.c.Resource.Text.ThreePointDelReserveDesc,
		}
	}
	return nil
}

func (s *Service) threePointHide(dynCtx *mdlv2.DynamicContext) *api.ThreePointItem {
	const (
		_blookLive       = "LIVE_RESERVE_RECALL"
		_blookAv         = "ARCHIVE_RESERVE_RECALL"
		_blookAvPlatBack = "LIVE_PLAY_BACK_RECALL"
	)
	hide := &api.ThreePointHide{
		Icon:  s.c.Resource.Icon.ThreePointHideIcon,
		Title: s.c.Resource.Text.ThreePointHideText,
		Interactive: &api.ThreePointHideInteractive{
			Title:   "确认隐藏预约的内容吗？",
			Confirm: "确认隐藏",
			Cancel:  "我再想想",
			Toast:   "已隐藏",
		},
		BlookFid: dynCtx.Dyn.Rid,
	}
	if dynCtx.Dyn.IsAv() {
		hide.BlookType = _blookAv
		if dynCtx.Dyn.SType == mdlv2.VideoStypePlayback {
			hide.BlookType = _blookAvPlatBack
		}
	}
	if dynCtx.Dyn.IsLiveRcmd() {
		hide.BlookType = _blookLive
	}
	item := &api.ThreePointItem{
		Type: api.ThreePointType_hide,
		Item: &api.ThreePointItem_Hide{
			Hide: hide,
		},
	}

	return item
}

func (s *Service) authorBatch(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	batch, ok := dynCtx.ResBatch[dynCtx.Dyn.Rid]
	var userMdl *api.Module_ModuleAuthor
	if !ok {
		userMdl = &api.Module_ModuleAuthor{
			ModuleAuthor: &api.ModuleAuthor{
				Author: &api.UserInfo{
					Mid:  dynCtx.Dyn.UID,
					Face: s.c.Resource.Icon.ModuleAuthorDefaultFace,
				},
			},
		}
		goto END
	}
	userMdl = &api.Module_ModuleAuthor{
		ModuleAuthor: &api.ModuleAuthor{
			Author: &api.UserInfo{
				Mid:  dynCtx.Dyn.UID,
				Uri:  batch.JumpURL,
				Name: batch.Name,
				Face: batch.Face,
			},
			Uri: batch.JumpURL,
		},
	}
END:
	userMdl.ModuleAuthor.PtimeLabelText = s.publishLabel(c, dynCtx, general)
	userMdl.ModuleAuthor.Weight = s.weight(c, dynCtx, general)
	// 如果提权没有则展示三点
	if userMdl.ModuleAuthor.Weight == nil || len(userMdl.ModuleAuthor.Weight.Items) == 0 {
		userMdl.ModuleAuthor.TpList = s.threePoint(c, dynCtx, general)
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_author,
		ModuleItem: userMdl,
	}
	dynCtx.Modules = append(dynCtx.Modules, module)
	return nil
}
