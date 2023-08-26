package dynHandler

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/log"
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	relationmdl "go-gateway/app/app-svr/app-dynamic/interface/model/relation"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

const (
	// 用户模块及三点
	_threePointShareText               = "分享"
	_threePointReportText              = "举报"
	_threePointFavText                 = "收藏合集"
	_threePointCancelFavText           = "取消收藏合集"
	_threePointAutoPlayOpenV1Text      = "关闭WiFi/免流环境下自动播放"
	_threePointAutoPlayCloseV1Text     = "开启WiFi/免流环境下自动播放"
	_threePointAutoPlayOpenIPADV1Text  = "关闭WiFi下自动播放"
	_threePointAutoPlayCloseIPADV1Text = "开启WiFi下自动播放"
	_threePointAutoPlayOpenV2Text      = "开启自动播放"
	_threePointAutoPlayCloseV2Text     = "关闭自动播放"
	_threePointAutoPlayOnlyText        = "仅WiFi/免流下自动播放"
	_threePointShareIcon               = "https://i0.hdslb.com/bfs/feed-admin/ee5902a63bbe4a0d78646d11036b062ea60573f6.png"
	_threePointReportIcon              = "https://i0.hdslb.com/bfs/feed-admin/d2a0449e705dcdeac1d2ac1e9da7e05d06b73dee.png"
	_threePointFavIcon                 = "https://i0.hdslb.com/bfs/feed-admin/b2fbc0f488957045c3c84d43d246485b2fbd1dd5.png"
	_moduleAuthorDefaultFaceIcon       = "https://i0.hdslb.com/bfs/feed-admin/c4cf44cc63cbe7e9642482a600db915500fd4d2f.png"
	_weightIcon                        = "https://i0.hdslb.com/bfs/feed-admin/2565c84eaf2f20853444eaa8ff810c62281b71ea.png"
	_threePointAutoPlayOpenIcon        = "https://i0.hdslb.com/bfs/feed-admin/bbcfe9c0d2b0d2482ac8a4fd8ed7bfa04ccdb27d.png"
	_threePointAutoPlayCloseIcon       = "https://i0.hdslb.com/bfs/feed-admin/661555f2e93c240a3a84abe8535b3c4c7bb23534.png"
)

func (schema *CardSchema) authorUser(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) *dynamicapi.ModuleAuthor {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Dyn.UID == 0 {
		return nil
	}
	userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID)
	if !ok {
		log.Warn("module miss mid(%v) dynid(%v) author uid(%d)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID)
		return nil
	}
	ptimeLabelText := schema.publishLabel(dynSchemaCtx, general)
	author := &dynamicapi.ModuleAuthor{
		Mid: userInfo.Mid,
		Author: &dynamicapi.UserInfo{
			Mid:  userInfo.Mid,
			Name: userInfo.Name,
			Face: userInfo.Face,
			Official: &dynamicapi.OfficialVerify{ // 认证
				Type: userInfo.Official.Type,
				Desc: userInfo.Official.Desc,
			},
			Vip: &dynamicapi.VipInfo{ // 会员
				Type:    userInfo.Vip.Type,
				Status:  userInfo.Vip.Status,
				DueDate: userInfo.Vip.DueDate,
				Label: &dynamicapi.VipLabel{
					Path:       userInfo.Vip.Label.Path,
					Text:       userInfo.Vip.Label.Text,
					LabelTheme: userInfo.Vip.Label.LabelTheme,
				},
				ThemeType:       userInfo.Vip.ThemeType,
				AvatarSubscript: userInfo.Vip.AvatarSubscript,
				NicknameColor:   userInfo.Vip.NicknameColor,
			},
			Pendant: &dynamicapi.UserPendant{ // 头像挂件
				Pid:    int64(userInfo.Pendant.Pid),
				Name:   userInfo.Pendant.Name,
				Image:  userInfo.Pendant.Image,
				Expire: userInfo.Pendant.Expire,
			},
			Nameplate: &dynamicapi.Nameplate{ // 勋章
				Nid:        int64(userInfo.Nameplate.Nid),
				Name:       userInfo.Nameplate.Name,
				Image:      userInfo.Nameplate.Image,
				ImageSmall: userInfo.Nameplate.ImageSmall,
				Level:      userInfo.Nameplate.Level,
				Condition:  userInfo.Nameplate.Condition,
			},
			Uri:            topiccardmodel.FillURI(topiccardmodel.GotoSpaceDyn, strconv.FormatInt(userInfo.Mid, 10), nil),
			FaceNftNew:     userInfo.FaceNftNew,
			Level:          userInfo.Level,
			IsSeniorMember: userInfo.IsSeniorMember,
		},
		PtimeLabelText: ptimeLabelText,
		Uri:            dynCtx.Interim.PromoURI,                                              // 帮推
		Relation:       relationmdl.RelationChange(dynCtx.Dyn.UID, dynCtx.ResRelationUltima), // 关注组件
		ShowLevel:      schema.showLevel(),
	}
	// 置顶
	if val, ok := dynSchemaCtx.ItemFrom[dynSchemaCtx.DynCtx.Dyn.DynamicID]; ok && val == topicsvc.ItemFrom_TopShow.String() {
		author.IsTop = true
	}
	return author
}

func (schema *CardSchema) authorShell(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID)
	if !ok {
		log.Warn("module error mid(%v) dynid(%v) authorShell uid(%d)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID)
		return nil
	}
	var titles []*dynamicapi.ModuleAuthorForwardTitle
	titles = append(titles, &dynamicapi.ModuleAuthorForwardTitle{
		Text: fmt.Sprintf("@%s", userInfo.Name),
		Url:  topiccardmodel.FillURI(topiccardmodel.GotoSpaceDyn, strconv.FormatInt(userInfo.Mid, 10), nil),
	})
	userMdl := &dynamicapi.Module_ModuleAuthorForward{
		ModuleAuthorForward: &dynamicapi.ModuleAuthorForward{
			Title:          titles,
			Uid:            userInfo.Mid,
			PtimeLabelText: schema.publishLabel(dynSchemaCtx, general),
			FaceUrl:        userInfo.Face,
			ShowFollow:     schema.showFollow(dynCtx, general),
			Relation:       relationmdl.RelationChange(dynCtx.Dyn.UID, dynCtx.ResRelationUltima),
		},
	}
	userMdl.ModuleAuthorForward.TpList = schema.threePoint(dynSchemaCtx, general)
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_author_forward,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (schema *CardSchema) authorPGC(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	var userMdl *dynamicapi.Module_ModuleAuthor
	pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid))
	if !ok {
		userMdl = &dynamicapi.Module_ModuleAuthor{
			ModuleAuthor: &dynamicapi.ModuleAuthor{
				Author: &dynamicapi.UserInfo{
					Mid:  dynCtx.Dyn.UID,
					Face: _moduleAuthorDefaultFaceIcon,
				},
			},
		}
		schema.finishModulePGCAuthor(userMdl, dynSchemaCtx, general)
		return nil
	}
	userMdl = &dynamicapi.Module_ModuleAuthor{
		ModuleAuthor: &dynamicapi.ModuleAuthor{
			Author: &dynamicapi.UserInfo{Mid: dynCtx.Dyn.UID, Uri: pgc.Url},
			Uri:    pgc.Url,
		},
	}
	if pgc.Season != nil {
		userMdl.ModuleAuthor.Author.Name = pgc.Season.Title
		userMdl.ModuleAuthor.Author.Face = pgc.Season.Cover
	}
	schema.finishModulePGCAuthor(userMdl, dynSchemaCtx, general)
	return nil
}

func (schema *CardSchema) finishModulePGCAuthor(userMdl *dynamicapi.Module_ModuleAuthor, dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) {
	if userMdl == nil {
		return
	}
	dynCtx := dynSchemaCtx.DynCtx
	userMdl.ModuleAuthor.PtimeLabelText = schema.publishLabel(dynSchemaCtx, general)
	userMdl.ModuleAuthor.Weight = schema.weight(dynCtx, general)
	// 如果提权没有则展示三点
	if userMdl.ModuleAuthor.Weight == nil || len(userMdl.ModuleAuthor.Weight.Items) == 0 {
		userMdl.ModuleAuthor.TpList = schema.threePoint(dynSchemaCtx, general)
	}
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_author,
		ModuleItem: userMdl,
	}
	dynCtx.Modules = append(dynCtx.Modules, module)
}

func (schema *CardSchema) authorShellPGC(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid))
	if !ok {
		dynCtx.Interim.IsPassCard = true
		log.Error("authorShellPGC dynCtx.GetResPGC is nil general=%+v, Rid=%d", general, dynCtx.Dyn.Rid)
		return nil
	}
	var title string
	if pgc.Season != nil {
		title = pgc.Season.Title
	}
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_author_forward,
		ModuleItem: &dynamicapi.Module_ModuleAuthorForward{
			ModuleAuthorForward: &dynamicapi.ModuleAuthorForward{
				Uid: dynCtx.Dyn.UID,
				Title: []*dynamicapi.ModuleAuthorForwardTitle{
					{
						Text: title, // PGC内容不@
						Url:  pgc.Url,
					},
				},
				PtimeLabelText: schema.publishLabel(dynSchemaCtx, general),
				ShowFollow:     schema.showFollow(dynCtx, general),
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (schema *CardSchema) showFollow(dynCtx *dynmdlV2.DynamicContext, general *topiccardmodel.GeneralParam) bool {
	if dynCtx.Dyn.IsPGC() || dynCtx.Dyn.IsUGCSeason() {
		return false
	}
	if general.Mid == dynCtx.Dyn.UID {
		return false
	}
	return true
}

func (schema *CardSchema) showLevel() bool {
	return true
}

// 发布人模块
func (schema *CardSchema) author(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID)
	if !ok {
		log.Warn("module error mid(%v) dynid(%v) author uid(%d)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID)
	}
	var (
		userMdl        *dynamicapi.Module_ModuleAuthor
		ptimeLabelText = schema.publishLabel(dynSchemaCtx, general)
		isWeight       bool
	)
	if userInfo == nil { // 兜底 默认返回灰头像、空白文案
		userMdl = &dynamicapi.Module_ModuleAuthor{
			ModuleAuthor: &dynamicapi.ModuleAuthor{
				Mid:            dynCtx.Dyn.UID,
				PtimeLabelText: ptimeLabelText,
				Author: &dynamicapi.UserInfo{
					Mid:  dynCtx.Dyn.UID,
					Face: _moduleAuthorDefaultFaceIcon,
					Uri:  topiccardmodel.FillURI(topiccardmodel.GotoSpaceDyn, strconv.FormatInt(dynCtx.Dyn.UID, 10), nil),
				},
				Uri: dynCtx.Interim.PromoURI, // 帮推
			},
		}
		goto END
	}
	userMdl = &dynamicapi.Module_ModuleAuthor{
		ModuleAuthor: schema.authorUser(dynSchemaCtx, general),
	}
	if userInfo.Pendant.ImageEnhance != "" { // 动效图优先
		userMdl.ModuleAuthor.Author.Pendant.Image = userInfo.Pendant.ImageEnhance
	}
	// 直播状态
	if !dynCtx.Interim.HiddenAuthorLive {
		if userLive, ok := dynCtx.GetResUserLive(dynCtx.Dyn.UID); ok && userLive.Status != nil && userLive.Status.Password == "" {
			userMdl.ModuleAuthor.Author.Live = &dynamicapi.LiveInfo{
				IsLiving:  int32(userLive.Status.LiveStatus),
				LiveState: dynamicapi.LiveState(userLive.Status.LiveStatus),
				Uri:       topiccardmodel.FillURI(topiccardmodel.GotoLive, strconv.FormatInt(userLive.RoomId, 10), nil),
			}
			if userLivePlayURL, ok := dynCtx.GetResUserLivePlayURL(dynCtx.Dyn.UID); ok {
				userMdl.ModuleAuthor.Author.Live.Uri = userLivePlayURL.Link
			}
		}
	}
	// 提权样式
	if !dynCtx.Dyn.IsSubscription() {
		userMdl.ModuleAuthor.Weight = schema.weight(dynCtx, general)
		if userMdl.ModuleAuthor.Weight != nil && len(userMdl.ModuleAuthor.Weight.Items) > 0 {
			isWeight = true
		}
	}
	if canShowDecorateCard(dynCtx, userMdl, isWeight) {
		// 装扮卡
		decoInfo, ok := dynCtx.ResMyDecorate[userInfo.Mid]
		if ok {
			// 小于6位数 前置补0
			numberStr := strconv.Itoa(decoInfo.Fan.Number)
			// nolint:gomnd
			if decoInfo.Fan.Number < 100000 {
				numberStr = fmt.Sprintf("%06d", decoInfo.Fan.Number)
			}
			userMdl.ModuleAuthor.DecorateCard = &dynamicapi.DecorateCard{
				Id:      decoInfo.ID,
				CardUrl: decoInfo.CardURL,
				JumpUrl: decoInfo.JumpURL,
				Fan: &dynamicapi.DecoCardFan{
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
END:
	// 非旧订阅卡出三点
	if !isWeight {
		// 如果提权没有则展示三点
		userMdl.ModuleAuthor.TpList = schema.threePoint(dynSchemaCtx, general)
	}
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_author,
		ModuleItem: userMdl,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func canShowDecorateCard(dynCtx *dynmdlV2.DynamicContext, userMdl *dynamicapi.Module_ModuleAuthor, isWeight bool) bool {
	if dynCtx.ResMyDecorate == nil {
		return false
	}
	if userMdl.ModuleAuthor.IsTop {
		// 置顶不显示装扮
		return false
	}
	// 动态显示装扮判断老逻辑
	return !dynCtx.Dyn.IsSubscription() && !dynCtx.Dyn.IsSubscriptionNew() && (userMdl.ModuleAuthor.Author.Live == nil || userMdl.ModuleAuthor.Author.Live.LiveState != dynamicapi.LiveState_live_live) && !isWeight
}

func (schema *CardSchema) publishLabel(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) string {
	// 前半部分
	var labels []string
	if label1 := topiccardmodel.ConstructPubTime(general.LocalTime, dynSchemaCtx.DynCtx.Dyn.Timestamp); label1 != "" {
		labels = append(labels, label1)
	}
	// 后半部分
	if label2 := schema.publishSuffix(dynSchemaCtx); label2 != "" {
		labels = append(labels, label2)
	}
	return strings.Join(labels, " · ")
}

func (schema *CardSchema) publishSuffix(dynSchemaCtx *topiccardmodel.DynSchemaCtx) string {
	if v, ok := dynSchemaCtx.HiddenAttached[dynSchemaCtx.DynCtx.Dyn.DynamicID]; ok && v {
		return topiccardmodel.TopicHiddenAttatchedText
	}
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Dyn.IsAv() {
		return publishAvSuffix(dynSchemaCtx)
	}
	if dynCtx.Dyn.IsPGC() || dynCtx.Dyn.IsBatch() {
		return "更新了"
	}
	if dynCtx.Dyn.IsCourUp() {
		return "发布了课程"
	}
	if dynCtx.Dyn.IsCourse() {
		return "更新了课程"
	}
	if dynCtx.Dyn.IsArticle() {
		return "投稿了文章"
	}
	return ""
}

func publishAvSuffix(dynSchemaCtx *topiccardmodel.DynSchemaCtx) string {
	dynCtx := dynSchemaCtx.DynCtx
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
	case dynmdlV2.VideoStypeDynamic, dynmdlV2.VideoStypeDynamicStory:
		if dynmdlV2.GetArchiveSType(dynCtx.Dyn.SType) == dynamicapi.VideoType_video_type_story {
			return "发布了动态视频"
		}
		return "发布了动态"
	case dynmdlV2.VideoStypePlayback:
		return "投稿了直播回放"
	}
	return "投稿了视频"
}

func (schema *CardSchema) threePoint(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) []*dynamicapi.ThreePointItem {
	dynCtx := dynSchemaCtx.DynCtx
	var ext []*dynamicapi.ThreePointItem
	// 分享
	if isShare := schema.threePointShare(dynCtx, general); isShare {
		ext = append(ext, tpShare())
	}
	// 收藏
	if isFav, favID := schema.threePointFav(dynCtx, general); isFav {
		ext = append(ext, tpFav(favID))
	}
	// 与话题无关
	if pd.WithContext(dynSchemaCtx.Ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().Or().IsPlatAndroidB().Or().IsPlatAndroidI().And().Build(">", int64(6540000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsMobiAppIPhone().Or().IsMobiAppIPhoneI().Or().IsPlatIPhoneB().And().Build(">", int64(65400000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build(">", int64(33200000))
	}).MustFinish() {
		ext = append(ext, tpTopicIrrelevant(dynSchemaCtx))
	}
	// 举报
	if isReport, titles, reportMid := schema.threePointReport(dynCtx, general); isReport {
		ext = append(ext, tpReport(dynSchemaCtx.TopicId, dynCtx.Dyn.DynamicID, reportMid, titles))
	}
	// 自动播放
	if isAutoPlay, openText, closeText := schema.threePointAutoPlay(dynCtx, general); isAutoPlay {
		ext = append(ext, TpAutoPlay(dynCtx, openText, closeText))
	}
	return ext
}

func (schema *CardSchema) threePointAutoPlay(dynCtx *dynmdlV2.DynamicContext, general *topiccardmodel.GeneralParam) (isAutoPlay bool, open, close string) {
	if dynCtx.Dyn.IsAv() {
		isAutoPlay = true
		open = _threePointAutoPlayOpenV1Text
		close = _threePointAutoPlayCloseV1Text
		if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
			open = _threePointAutoPlayOpenIPADV1Text
			close = _threePointAutoPlayCloseIPADV1Text
		}
	}
	// 转发卡原卡片为UGC、PGC、直播大卡时，添加自动播放按钮
	if dynCtx.Dyn.IsForward() {
		if dynCtx.Interim.DynTypeKernel == dynmdlV2.DynTypeVideo {
			isAutoPlay = true
			open = _threePointAutoPlayOpenV1Text
			close = _threePointAutoPlayCloseV1Text
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				open = _threePointAutoPlayOpenIPADV1Text
				close = _threePointAutoPlayCloseIPADV1Text
			}
		}
		if dynCtx.Interim.DynTypeKernel == dynmdlV2.DynTypePGCBangumi ||
			dynCtx.Interim.DynTypeKernel == dynmdlV2.DynTypePGCMovie ||
			dynCtx.Interim.DynTypeKernel == dynmdlV2.DynTypePGCTv ||
			dynCtx.Interim.DynTypeKernel == dynmdlV2.DynTypePGCGuoChuang ||
			dynCtx.Interim.DynTypeKernel == dynmdlV2.DynTypePGCDocumentary ||
			dynCtx.Interim.DynTypeKernel == dynmdlV2.DynTypeBangumi {
			isAutoPlay = true
			open = _threePointAutoPlayOpenV1Text
			close = _threePointAutoPlayCloseV1Text
			if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
				open = _threePointAutoPlayOpenIPADV1Text
				close = _threePointAutoPlayCloseIPADV1Text
			}
		}
		if dynCtx.Interim.DynTypeKernel == dynmdlV2.DynTypeLiveRcmd {
			isAutoPlay = true
			open = _threePointAutoPlayOpenV1Text
			close = _threePointAutoPlayCloseV1Text
		}
		if dynCtx.Interim.DynTypeKernel == dynmdlV2.DynTypeUGCSeason {
			isAutoPlay = true
			open = _threePointAutoPlayOpenV1Text
			close = _threePointAutoPlayCloseV1Text
		}
	}
	return
}

func TpAutoPlay(_ *dynmdlV2.DynamicContext, open, close string) *dynamicapi.ThreePointItem {
	item := &dynamicapi.ThreePointItem{
		Type: dynamicapi.ThreePointType_auto_play,
		Item: &dynamicapi.ThreePointItem_AutoPlayer{
			AutoPlayer: &dynamicapi.ThreePointAutoPlay{
				OpenIcon:  _threePointAutoPlayOpenIcon,
				OpenText:  open,
				CloseIcon: _threePointAutoPlayCloseIcon,
				CloseText: close,
				// v2
				OpenIconV2:  _threePointAutoPlayOpenIcon,
				OpenTextV2:  _threePointAutoPlayOpenV2Text,
				CloseIconV2: _threePointAutoPlayCloseIcon,
				CloseTextV2: _threePointAutoPlayCloseV2Text,
				OnlyIcon:    _threePointAutoPlayOpenIcon,
				OnlyText:    _threePointAutoPlayOnlyText,
			},
		},
	}
	return item
}

func (schema *CardSchema) threePointShare(dynCtx *dynmdlV2.DynamicContext, _ *topiccardmodel.GeneralParam) bool {
	if !dynCtx.Dyn.IsCheeseBatch() && !dynCtx.Dyn.IsSubscriptionNew() {
		return true
	}
	return false
}

func tpShare() *dynamicapi.ThreePointItem {
	item := &dynamicapi.ThreePointItem{
		Type: dynamicapi.ThreePointType_share,
		Item: &dynamicapi.ThreePointItem_Share{
			Share: &dynamicapi.ThreePointShare{
				Icon:  _threePointShareIcon,
				Title: _threePointShareText,
			},
		},
	}
	return item
}

func (schema *CardSchema) threePointReport(dynCtx *dynmdlV2.DynamicContext, general *topiccardmodel.GeneralParam) (isReport bool, titles []string, reportMid int64) {
	if dynCtx.Dyn.IsPGC() {
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
	default:
		reportText = "动态"
	}
	if dynCtx.Interim.Desc != "" {
		reportText = dynCtx.Interim.Desc
	}
	titles = append(titles, reportText)
	return
}

func tpTopicIrrelevant(dynSchemaCtx *topiccardmodel.DynSchemaCtx) *dynamicapi.ThreePointItem {
	return &dynamicapi.ThreePointItem{
		Type: dynamicapi.ThreePointType_topic_irrelevant,
		Item: &dynamicapi.ThreePointItem_TopicIrrelevant{TopicIrrelevant: &dynamicapi.ThreePointTopicIrrelevant{
			Icon:    "https://i0.hdslb.com/bfs/feed-admin/86e989ccd0d454106a19e3222729e4342da68971.png",
			Title:   "与话题无关",
			Toast:   "感谢你的反馈",
			TopicId: dynSchemaCtx.TopicId,
			ResId:   dynSchemaCtx.DynCtx.Dyn.DynamicID,
			ResType: 0,
			Reason:  "与话题无关",
		}},
	}
}

func tpReport(topicId, dynid, uid int64, titles []string) *dynamicapi.ThreePointItem {
	title := url.QueryEscape(strings.Join(titles, ":"))
	if strings.IndexByte(title, '+') > -1 {
		title = strings.Replace(title, "+", "%20", -1)
	}
	item := &dynamicapi.ThreePointItem{
		Type: dynamicapi.ThreePointType_report,
		Item: &dynamicapi.ThreePointItem_Default{
			Default: &dynamicapi.ThreePointDefault{
				Icon:  _threePointReportIcon,
				Title: _threePointReportText,
				Uri:   fmt.Sprintf("bilibili://following/new_topic/report_card?uid=%d&topic_id=%d&title=%s&res_id=%d&res_type=0", uid, topicId, title, dynid),
			},
		},
	}
	return item
}

func (schema *CardSchema) threePointFav(dynCtx *dynmdlV2.DynamicContext, _ *topiccardmodel.GeneralParam) (bool, int64) {
	if dynCtx.Dyn.IsUGCSeason() {
		return true, dynCtx.Dyn.UID
	}
	return false, 0
}

func tpFav(favID int64) *dynamicapi.ThreePointItem {
	item := &dynamicapi.ThreePointItem{
		Type: dynamicapi.ThreePointType_favorite,
		Item: &dynamicapi.ThreePointItem_Favorite{
			Favorite: &dynamicapi.ThreePointFavorite{
				Icon:        _threePointFavIcon,
				Title:       _threePointFavText,
				Id:          favID,
				IsFavourite: true,
				CancelIcon:  _threePointFavIcon,
				CancelTitle: _threePointCancelFavText,
			},
		},
	}
	return item
}

func (schema *CardSchema) weight(dynCtx *dynmdlV2.DynamicContext, _ *topiccardmodel.GeneralParam) *dynamicapi.Weight {
	if (dynCtx.Dyn.PassThrough == nil || (dynCtx.Dyn.Property.RcmdType != dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_TOP_WEIGHT && dynCtx.Dyn.Property.RcmdType != dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_HISTORY_WEIGHT)) || dynCtx.Dyn.PassThrough.FeedBack == nil {
		return nil
	}
	res := &dynamicapi.Weight{
		Title: dynCtx.Dyn.PassThrough.FeedBack.IconTitle,
		Icon:  _weightIcon,
	}
	if dynCtx.Dyn.PassThrough.FeedBack.JumpButtonText != "" {
		res.Items = append(res.Items, &dynamicapi.WeightItem{
			Type: dynamicapi.WeightType_weight_jump,
			Item: &dynamicapi.WeightItem_Button{
				Button: &dynamicapi.WeightButton{
					JumpUrl: dynCtx.Dyn.PassThrough.FeedBack.JumpUrl,
					Title:   dynCtx.Dyn.PassThrough.FeedBack.JumpButtonText,
				},
			},
		})
	}
	if dynCtx.Dyn.PassThrough.FeedBack.FeedBackButtonText != "" {
		res.Items = append(res.Items, &dynamicapi.WeightItem{
			Type: dynamicapi.WeightType_weight_dislike,
			Item: &dynamicapi.WeightItem_Dislike{
				Dislike: &dynamicapi.WeightDislike{
					FeedBackType: dynCtx.Dyn.PassThrough.FeedBack.FeedBackBizType,
					Title:        dynCtx.Dyn.PassThrough.FeedBack.FeedBackButtonText,
				},
			},
		})
	}
	return res
}
