package account

import (
	"context"
	"fmt"
	"hash/crc32"
	"strconv"
	"sync"
	"time"

	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	ott "go-gateway/app/app-svr/app-interface/api-dependence/ott-service"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	acm "go-gateway/app/app-svr/app-interface/interface-legacy/model/account"
	bl "go-gateway/app/app-svr/app-interface/interface-legacy/model/bili_link"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"
	relApi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	answerApi "git.bilibili.co/bapis/bapis-go/community/interface/answer"
	locmdl "git.bilibili.co/bapis/bapis-go/community/service/location"
	newmentapi "git.bilibili.co/bapis/bapis-go/newmont/service/v1"
	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	resApi "git.bilibili.co/bapis/bapis-go/resource/service/v1"
	vipclient "git.bilibili.co/bapis/bapis-go/vip/service"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

const (
	_initSidebarKey   = "sidebar_%d_%d_%s"
	_defaultFace      = "https://static.hdslb.com/images/member/noface.gif"
	_rankMember       = 10000
	_answered         = 3 // 答题完成
	_teenagersmodeURI = "bilibili://user_center/teenagersmode"
	_lessonsmodeURI   = "bilibili://user_center/lessonsmode"
	//离线缓存
	_downloadURI                = "bilibili://user_center/download"
	_vipPosition                = 3
	_creativeDefaultVal         = 1
	_mineBubble                 = 9
	_mineOperate                = 3 //创作/直播运营位
	_mineCommonOp               = 7 //通用运营位
	_ipadSectionNew             = 1 //ipad【我的】页新样式
	_ipadSectionOld             = 0 //ipad【我的】页老样式
	_androidRecommendGameCenter = 389
)

var (
	_showWhiteList = map[string]struct{}{

		"bilibili://uper/user_center/archive_list":                     {},
		"bilibili://uper/homevc":                                       {},
		"https://member.bilibili.com/studio/up-allowance-h5#/":         {},
		"https://member.bilibili.com/studio/up-allowance-h5#":          {},
		"https://www.bilibili.com/blackboard/x/activity-tougao-h5/all": {},
		"bilibili://user_center/myfollows":                             {},
		"https://member.bilibili.com/college?navhide=1&from=my":        {},
		"bilibili://user_center/feedback":                              {},
		_teenagersmodeURI:                                              {},
		_lessonsmodeURI:                                                {},
		// 安卓创作中心
		"bilibili://main/drawer/upper-upload": {},
		"bilibili://main/drawer/upper-academy?redirect=https%3a%2f%2fmember.bilibili.com%2fcollege%3ffrom%3dmy%26navhide%3d1": {},
		"bilibili://main/drawer/upper": {},
		"bilibili://main/drawer/upper-hot?redirect=https%3a%2f%2fwww.bilibili.com%2fblackboard%2fx%2factivity-tougao-h5%2fall": {},
		// 安卓个人中心
		"bilibili://main/drawer/main-page": {},
		// 安卓我的服务
		"bilibili://mall/order/list?msource=mine_v5.29&from=mine_v5.29": {},
		"https://www.bilibili.com/h5/faq":                               {},
		"https://www.bilibili.com/h5/customer-service":                  {},
		"https://www.bilibili.com/h5/teenagers/home?navhide=1":          {},
	}
	_showLessonsWhiteList = map[string]struct{}{
		"bilibili://user_center/download":     {},
		"bilibili://user_center/history":      {},
		"bilibili://user_center/favourite":    {},
		"bilibili://user_center/watch_later":  {},
		"bilibili://user_center/free_traffic": {},
		"bilibili://main/drawer/history":      {},
		"bilibili://main/drawer/offline":      {},
		"bilibili://main/drawer/favorites":    {},
		"bilibili://main/drawer/watch-later":  {},
		"bilibili://main/drawer/freedata":     {},
	}
	_showIPadWhiteList = map[string]struct{}{
		"bilibili://user_center/feedback": {},
		_teenagersmodeURI:                 {},
		_lessonsmodeURI:                   {},
		// ipad个人中心-设置
		"bilibili://user_center/setting":                       {},
		"https://www.bilibili.com/h5/customer-service":         {},
		"https://www.bilibili.com/h5/teenagers/home?navhide=1": {},
	}
	_showLessonsIpadWhiteList = map[string]struct{}{
		"bilibili://user_center/history":     {},
		"bilibili://user_center/download":    {},
		"bilibili://user_center/favourite":   {},
		"bilibili://user_center/watch_later": {},
	}
)

func (s *Service) whiteList(uri string, isIpad bool, teenagersMode, lessonsMode int) bool {
	if teenagersMode == 0 && lessonsMode == 0 {
		return true
	}
	if isIpad {
		if _, ok := _showIPadWhiteList[uri]; ok {
			return true
		}
		if _, ok := _showLessonsIpadWhiteList[uri]; ok && lessonsMode != 0 {
			return true
		}
		return false
	}
	if _, ok := _showWhiteList[uri]; ok {
		return true
	}
	if _, ok := _showLessonsWhiteList[uri]; ok && lessonsMode != 0 {
		return true
	}
	return false
}

func translateTCMine(dst *space.Mine) {
	for _, sec := range dst.SectionsV2 {
		i18n.TranslateAsTCV2(&sec.Title)
		if sec.UpTitle != "" {
			i18n.TranslateAsTCV2(&sec.Title)
		}
		if sec.BeUpTitle != "" {
			i18n.TranslateAsTCV2(&sec.BeUpTitle)
		}
		if sec.Button != nil {
			i18n.TranslateAsTCV2(&sec.Button.Text)
		}
		for _, sSec := range sec.Items {
			i18n.TranslateAsTCV2(&sSec.Title)
		}
	}
	for _, sec := range dst.Sections {
		i18n.TranslateAsTCV2(&sec.Title)
		for _, sSec := range sec.Items {
			i18n.TranslateAsTCV2(&sSec.Title)
		}
	}
	for _, sec := range dst.IpadSections {
		i18n.TranslateAsTCV2(&sec.Title)
	}
	for _, sec := range dst.IpadUpperSections {
		i18n.TranslateAsTCV2(&sec.Title)
	}
	if dst.Vip.Label.Text != "" {
		i18n.TranslateAsTCV2(&dst.Vip.Label.Text)
	}
	if dst.MallHome != nil {
		i18n.TranslateAsTCV2(&dst.MallHome.Title)
	}
	if dst.Answer != nil {
		i18n.TranslateAsTCV2(&dst.Answer.Text)
	}
	if dst.VIPSection != nil {
		i18n.TranslateAsTCV2(&dst.VIPSection.Title)
	}
	if dst.VIPSectionRight != nil {
		i18n.TranslateAsTCV2(&dst.VIPSectionRight.Title, &dst.VIPSectionRight.Tip)
	}
	if dst.VIPSectionV2 != nil {
		i18n.TranslateAsTCV2(&dst.VIPSectionV2.Title, &dst.VIPSectionV2.Subtitle, &dst.VIPSectionV2.Desc)
	}
}

// Mine mine center for iphone/android
func (s *Service) Mine(c context.Context, mid int64, platform, lang, filtered, channel, mobiApp string, build, teenagersMode, lessonsMode int, plat int8, device string, slocale, clocale, buvid string, biliLinkNew int64) (mine *space.Mine, err error) {
	var (
		whiteMap, rdMap map[int64]bool
		liveCenter      []*space.SectionItem
		ctCenter        *space.SectionV2
		mineSec         []*newmentapi.Section
		vipTip          *vipclient.TipsVipReply
		auths           map[int64]*locmdl.ZoneLimitAuth
		ctbTip          *space.ContributeTip
		hasGameCenter   bool
	)
	if mine, whiteMap, rdMap, ctCenter, mineSec, vipTip, auths, err = s.userInfo(c, mid, platform, channel, lang, mobiApp, plat, build, teenagersMode, lessonsMode, device, buvid, biliLinkNew, i18n.PreferTraditionalChinese(c, slocale, clocale)); err != nil {
		log.Error("日志告警 我的页请求账号接口错误: %+v", err)
		return
	}
	mine.LiveTip, ctbTip, liveCenter, hasGameCenter, mine.Bubbles = s.PreComposeSection(mineSec, mid)
	if i18n.PreferTraditionalChinese(c, slocale, clocale) {
		defer translateTCMine(mine)
	}
	if s.newMine(plat, build) {
		mine.SectionsV2 = s.newSecs(mineSec, build, teenagersMode, lessonsMode, filtered == "1", plat, liveCenter, ctCenter, mine.ShowCreative, mine.FirstLiveTime, ctbTip, hasGameCenter)
	} else if platform == "ios" {
		mine.Sections = s.sections(c, whiteMap, rdMap, mid, build, teenagersMode, lessonsMode, filtered == "1", plat, lang, liveCenter, ctCenter, auths)
	} else if platform == "android" {
		mine.Sections = s.androidSections(c, whiteMap, rdMap, mid, build, teenagersMode, lessonsMode, plat, lang, channel, liveCenter, ctCenter, auths)
	}
	// 开关控制，默认不下发皮肤装扮入口&&620及以上版本才下发（国际版不下发）
	if s.c.Switch.SkinOpen && feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.MineSkin, &feature.OriginResutl{
		MobiApp:    mobiApp,
		Device:     device,
		Build:      int64(build),
		BuildLimit: (mobiApp == "iphone" && build > s.c.BuildLimit.SkinOpenIOSBuild) || (mobiApp == "android" && build >= s.c.BuildLimit.SkinOpenAndroidBuild),
	}) {
		mine.MallHome = &space.MallHome{URI: "https://www.bilibili.com/h5/mall/skin/setting?navhide=1&from=mine", Title: "个性装扮", Icon: s.c.Custom.MallIcon}
	}
	// vip tips
	if vipTip == nil || len(vipTip.List) == 0 {
		log.Error("error: %+v, vipTips: %v", err, vipTip)
		return mine, nil
	}
	var (
		vipTipDetail *vipclient.TipsVipDetail
		url          = "https://big.bilibili.com/mobile/home"
	)
	for _, v := range vipTip.List {
		if v != nil && v.Position == _vipPosition {
			vipTipDetail = v
			break
		}
	}
	if vipTipDetail == nil {
		return mine, nil
	}
	if vipTipDetail.Link != "" {
		url = vipTipDetail.Link
	}
	mine.VIPSection = &space.VIPSection{
		Title: vipTipDetail.Tip,
		URL:   url,
		STime: vipTipDetail.StartTime,
		ETime: vipTipDetail.EndTime,
	}
	// 67版本之后使用双入口，左侧为VIPSectionV2,右侧为VIPSectionRight
	mine.VIPSectionRight = &space.VIPSectionRight{
		ID:    vipTipDetail.Id,
		Title: vipTipDetail.RightTitle,
		Link:  vipTipDetail.RightLink,
		Tip:   vipTipDetail.RightTip,
		Img:   vipTipDetail.Img,
	}
	// 58版本开始vip模块内容均由vip服务下发
	if s.newMine(plat, build) {
		mine.VIPSectionV2 = &space.VIPSectionV2{
			ID:    vipTipDetail.Id,
			Title: vipTipDetail.Title,
			URL:   url,
			Desc:  vipTipDetail.Tip,
			// 该字段5.42-6.6版本服务端下发
			Subtitle: vipTipDetail.SubTitle,
		}
	} else {
		// 58版本之前保留使用SubTitle字段
		mine.FromVIPSectionV2(vipTipDetail, url)
	}

	return mine, nil
}

func (s *Service) PreComposeSection(mineSec []*newmentapi.Section, mid int64) (*space.LiveTip, *space.ContributeTip, []*space.SectionItem, bool, []*space.Bubble) {
	var (
		liveTip       *space.LiveTip
		contributeTip *space.ContributeTip
		liveCenter    []*space.SectionItem
		hasGameCenter bool
		bubbles       []*space.Bubble
	)
	for _, ms := range mineSec {
		if ms == nil || len(ms.Items) == 0 {
			continue
		}
		if ms.Style == _mineBubble {
			bubbles = buildMineBubbles(ms.Items, mid)
			continue
		}
		if ms.Style == _mineOperate {
			liveTip, contributeTip = buildMineTips(ms.OpStyleType, ms.Items[0])
		}
		if _, ok := model.LiveModules[ms.Id]; ok {
			liveCenter = buildLiveCenter(ms.Items)
		}
		//如果有游戏中心强入口，则推荐服务不出游戏中心
		if _, ok := s.c.GameModuleID[strconv.FormatInt(ms.Id, 10)]; ok {
			hasGameCenter = true
		}
	}
	return liveTip, contributeTip, liveCenter, hasGameCenter, bubbles
}

func buildMineBubbles(items []*newmentapi.SectionItem, mid int64) []*space.Bubble {
	bubbleNameToType := map[string]space.BubbleType{
		"校园气泡": space.SchoolBubble,
	}
	var bubbles []*space.Bubble
	for _, v := range items {
		if v.NeedLogin == 1 && mid == 0 {
			continue
		}
		bubble := &space.Bubble{
			ID:   v.Id,
			Icon: v.Icon,
			Type: bubbleNameToType[v.Title],
		}
		bubbles = append(bubbles, bubble)
	}
	return bubbles
}

func buildMineTips(style int32, item *newmentapi.SectionItem) (*space.LiveTip, *space.ContributeTip) {
	if item == nil {
		return nil, nil
	}
	if style == 0 { //运营位配置
		liveTip := &space.LiveTip{
			Text:       item.OpTitle,
			Icon:       item.OpTitleIcon,
			ButtonText: item.OpLinkText,
			ButtonIcon: item.OpLinkIcon,
			Url:        item.Uri,
			Mod:        int64(item.OpLinkType),
			UrlText:    item.OpLinkText,
			Id:         item.Id,
		}
		return liveTip, nil
	}
	if style == 1 { //投稿强化卡
		contributeTip := &space.ContributeTip{
			BeUpTitle:  item.OpTitle,
			TipTitle:   item.OpSubTitle,
			TipIcon:    item.OpTitleIcon,
			ButtonText: item.OpLinkText,
			ButtonUrl:  item.Uri,
		}
		return nil, contributeTip
	}
	return nil, nil
}

func buildLiveCenter(msi []*newmentapi.SectionItem) []*space.SectionItem {
	var liveCenter []*space.SectionItem
	for _, item := range msi {
		if item == nil {
			continue
		}
		liveItem := &space.SectionItem{
			ID:           item.Id,
			Title:        item.Title,
			URI:          item.Uri,
			Icon:         item.Icon,
			NeedLogin:    int8(item.NeedLogin),
			RedDot:       int8(item.RedDot),
			GlobalRedDot: int8(item.GlobalRedDot),
			Display:      item.Display,
			RedDotForNew: item.RedDotForNew,
		}
		if item.MngIcon != nil {
			liveItem.MngRes = &space.MngRes{
				Icon:   item.MngIcon.Icon,
				IconID: item.MngIcon.Id,
			}
		}
		liveCenter = append(liveCenter, liveItem)
	}
	return liveCenter
}

// MineIpad mine center for ipad
func (s *Service) MineIpad(c context.Context, mid int64, mobiApp, platform, lang, filtered, channel, slocale, clocale string, build, teenagersMode, lessonsMode int, plat int8, buvid string) (mine *space.Mine, err error) {
	var (
		whiteMap, rdMap map[int64]bool
		auths           map[int64]*locmdl.ZoneLimitAuth
		mineSections    []*newmentapi.Section
	)
	mine = new(space.Mine)
	if i18n.PreferTraditionalChinese(c, slocale, clocale) {
		defer func() {
			translateTCMine(mine)
		}()
	}
	if mine, whiteMap, rdMap, _, mineSections, _, auths, err = s.userInfo(c, mid, platform, channel, lang, mobiApp, plat, build, teenagersMode, lessonsMode, "", buvid, 0, i18n.PreferTraditionalChinese(c, slocale, clocale)); err != nil {
		return
	}
	mine.IpadSectionStyle = s.chooseSectionStyle(mid, plat, build)
	switch mine.IpadSectionStyle {
	case _ipadSectionOld:
		mine.IpadSections, mine.IpadUpperSections = s.ipadSections(c, whiteMap, rdMap, mid, build, teenagersMode, lessonsMode, filtered == "1", plat, mobiApp, lang, auths)
	case _ipadSectionNew:
		mine.IpadSections, mine.IpadUpperSections, mine.IpadRecommendSections, mine.IpadMoreSections = s.ipadNewSections(mineSections, teenagersMode, lessonsMode, filtered == "1")
	default:
		log.Error("unknown IpadSectionStyle")
	}
	return
}

func (s *Service) chooseSectionStyle(mid int64, plat int8, build int) int64 {
	if !s.newMine(plat, build) {
		return _ipadSectionOld
	}
	sectionStyle := func() int64 {
		//ipad已全量
		if plat != model.PlatAndroidHD {
			return _ipadSectionNew
		}
		if s.c.Custom.AndroidPadSectionExp == 100 { //全量之后未登录用户也使用新样式
			return _ipadSectionNew
		}
		if mid == 0 { //没有全量,未登录用户不做实验
			return _ipadSectionOld
		}
		midStr := strconv.FormatInt(mid, 10)
		if _, ok := s.c.Custom.IpadNewSectionMid[midStr]; ok || int64(crc32.ChecksumIEEE([]byte(midStr+"ipadSectionExp"))%100) < s.c.Custom.AndroidPadSectionExp {
			return _ipadSectionNew
		}
		return _ipadSectionOld
	}()
	return sectionStyle
}

func (s *Service) ipadNewSections(mineSections []*newmentapi.Section, teenagersMode, lessonsMode int, isFiltered bool) (ipadSections, ipadUpperSections, ipadRecommendSections, ipadMoreSections []*space.SectionItem) {
	for _, mineSection := range mineSections {
		// 判断审核态下是否展示
		if isFiltered && mineSection.AuditShow == 0 {
			continue
		}
		switch mineSection.Title {
		case "创作中心":
			if teenagersMode != 0 || lessonsMode != 0 {
				continue
			}
			ipadUpperSections = s.handleSectionItem(mineSection.Items, teenagersMode, lessonsMode, isFiltered)
		case "视频":
			if teenagersMode != 0 {
				continue
			}
			ipadSections = s.handleSectionItem(mineSection.Items, teenagersMode, lessonsMode, isFiltered)
		case "推荐服务":
			ipadRecommendSections = s.handleSectionItem(mineSection.Items, teenagersMode, lessonsMode, isFiltered)
		case "更多服务":
			ipadMoreSections = s.handleSectionItem(mineSection.Items, teenagersMode, lessonsMode, isFiltered)
		default:
			continue
		}
	}
	return
}

func (s *Service) handleSectionItem(in []*newmentapi.SectionItem, teenagersMode, lessonsMode int, isFiltered bool) []*space.SectionItem {
	var out []*space.SectionItem
	for _, v := range in {
		if !s.whiteList(v.Uri, true, teenagersMode, lessonsMode) {
			continue
		}
		//审核态下屏蔽离线缓存
		if isFiltered && v.Uri == _downloadURI {
			continue
		}
		// 青少年模式下不展示课堂模式入口
		if teenagersMode != 0 && v.Uri == _lessonsmodeURI {
			continue
		}
		temp := &space.SectionItem{
			ID:           v.Id,
			Title:        v.Title,
			URI:          v.Uri,
			Icon:         v.Icon,
			NeedLogin:    int8(v.NeedLogin),
			RedDot:       int8(v.RedDot),
			RedDotForNew: v.RedDotForNew,
			GlobalRedDot: int8(v.GlobalRedDot),
			Display:      v.Display,
			MngRes: &space.MngRes{
				Icon:   v.MngIcon.GetIcon(),
				IconID: v.MngIcon.GetId(),
			},
		}
		out = append(out, temp)
	}
	return out
}

func wrapAccVipInfo(ctx context.Context, in accmdl.VipInfo, useHantVipImage bool) space.VipInfo {
	out := space.VipInfo{
		Type:       in.Type,
		Status:     in.Status,
		DueDate:    in.DueDate,
		VipPayType: in.VipPayType,
		ThemeType:  in.ThemeType,
		ThemeType_: in.ThemeType,
		Label: space.VipLabel{
			Path:        in.Label.Path,
			Text:        in.Label.Text,
			LabelTheme:  in.Label.LabelTheme,
			TextColor:   in.Label.TextColor,
			BgStyle:     in.Label.BgStyle,
			BgColor:     in.Label.BgColor,
			BorderColor: in.Label.BorderColor,
		},
		AvatarSubscript:    in.AvatarSubscript,
		NicknameColor:      in.NicknameColor,
		Role:               in.Role,
		AvatarSubscriptUrl: in.AvatarSubscriptUrl,
	}
	func() {
		if !in.Label.UseImgLabel {
			return
		}
		if useHantVipImage {
			if in.Label.ImgLabelUriHant == "" { //没有动态图降级使用静态图
				out.Label.Image = in.Label.ImgLabelUriHantStatic
				return
			}
			out.Label.Image = in.Label.ImgLabelUriHant
			return
		}
		if in.Label.ImgLabelUriHans == "" { //没有动态图降级使用静态图
			out.Label.Image = in.Label.ImgLabelUriHansStatic
			return
		}
		out.Label.Image = in.Label.ImgLabelUriHans
	}()
	return out
}

//nolint:gocognit
func (s *Service) userInfo(c context.Context, mid int64, platform, channel, lang, mobiApp string, plat int8, build, teenagersMode, lessonsMode int, device, buvid string, biliLinkNew int64, useHantVipImage bool) (mine *space.Mine, whiteMap, rdMap map[int64]bool, ctCenter *space.SectionV2, mineSections []*newmentapi.Section, vipTip *vipclient.TipsVipReply, auths map[int64]*locmdl.ZoneLimitAuth, err error) {
	mine = new(space.Mine)
	mine.Official.Type = -1
	eg := errgroup.WithContext(c)
	if mid > 0 {
		whiteMap = make(map[int64]bool)
		rdMap = make(map[int64]bool)
		var ps *accmdl.ProfileStatReply
		if ps, err = s.accDao.Profile3(c, mid); err != nil || ps.Profile == nil {
			log.Error("s.accDao.UserInfo(%d) error(%v) or ps.Profile is nil", mid, err)
			return
		}
		mine.Silence = ps.Profile.Silence
		mine.Mid = ps.Profile.Mid
		mine.Name = ps.Profile.Name
		mine.Face = ps.Profile.Face
		mine.FaceNftNew = ps.Profile.FaceNftNew
		mine.InRegAudit = ps.Profile.InRegAudit
		if ps.Profile.Face == "" {
			mine.Face = _defaultFace
		}
		mine.Coin = ps.Coins
		if ps.Profile.Pendant.Image != "" {
			mine.Pendant = &space.Pendant{Image: ps.Profile.Pendant.Image, ImageEnhance: ps.Profile.Pendant.ImageEnhance}
		}
		switch ps.Profile.Sex {
		case "男":
			mine.Sex = 1
		case "女":
			mine.Sex = 2
		default:
			mine.Sex = 0
		}
		mine.Rank = ps.Profile.Rank
		mine.Level = ps.Profile.Level
		mine.Vip = wrapAccVipInfo(c, ps.Profile.Vip, useHantVipImage)
		if ps.Profile.Vip.Status == model.VipStatusNormal { // 1-正常
			mine.VipType = ps.Profile.Vip.Type
		} else if ps.Profile.Vip.Status == model.VipStatusExpire && ps.Profile.Vip.DueDate > 0 { // 0-过期 (2-冻结 3-封禁不展示）
			mine.Vip.Label.Path = model.VipLabelExpire
		}
		if ps.Profile.Official.Role != 0 {
			mine.Official.Desc = ps.Profile.Official.Title
		}
		mine.Official.Type = int8(ps.Profile.Official.Type)
		func() {
			const topLevel = 6
			if ps.Profile.Level < topLevel {
				return
			}
			if int64(time.Since(ps.LevelInfo.LevelUp.Time()).Hours()) > s.c.Custom.TopLevelExTime {
				return
			}
			mine.Achievement.TopLevelFlash = &space.TopLevelFlash{Icon: s.c.AchievementConf.TopLevelIcon}
		}()
		// 获取用户头像和昵称是否修改过，未修改过展示引导
		eg.Go(func(ctx context.Context) (err error) {
			upPrompting, err := s.accDao.Prompting(ctx, mid)
			if err != nil {
				log.Error("s.accDao.Prompting mid(%d) err(%+v)", mid, err)
				return nil
			}
			if upPrompting != nil {
				mine.ShowNameGuide = !upPrompting.NameUpdated
				mine.ShowFaceGuide = !upPrompting.FaceUpdated
			}
			if !mine.ShowFaceGuide {
				return nil
			}
			// 如果头像未修改过，检验是否存在nft修改头像入口
			ownerReply, err := s.galleryDao.IsNFTFaceOwner(ctx, &gallerygrpc.MidReq{Mid: mid, RealIp: metadata.RemoteIP})
			if err != nil {
				log.Error("s.galleryDao.AccountHasNFT mid(%d) err(%+v)", mid, err)
				return nil
			}
			if ownerReply.Status == gallerygrpc.OwnerStatus_ISOWNER {
				mine.ShowNftFaceGuide = true
			}
			return nil
		})
		if ps.Profile.Silence == 1 {
			eg.Go(func(ctx context.Context) (err error) {
				if mine.EndTime, err = s.accDao.BlockTime(ctx, mid); err != nil {
					log.Error("s.accDao.BlockTime mid(%d) err(%+v)", mid, err)
					err = nil
				}
				return
			})
		}
		eg.Go(func(ctx context.Context) (err error) {
			if ps.Profile.Rank < _rankMember {
				var answer *answerApi.AnswerStatus
				if answer, err = s.asDao.AnswerStatus(ctx, mid, mobiApp, model.AnswerSourceMyinfo); err != nil {
					log.Error("s.asDao.AnswerStatus mid(%d) err(%+v)", mid, err)
					err = nil
				}
				if answer != nil && answer.Status != _answered {
					mine.AnswerStatus = answer.Status
					mine.Answer = &space.Answer{
						Text:     answer.Text,
						URL:      answer.Url,
						Progress: answer.Progress,
					}
				}
			}
			if mine.Answer != nil { //没有答题则出看板娘
				return
			}
			reply, err := s.accDao.CharacterUsageStatus(ctx, mid, int64(build), platform, mobiApp, device, buvid)
			if err != nil {
				log.Error("s.accDao.CharacterUsageStatus error(%+v)", err)
				return nil
			}
			if reply == nil {
				return nil
			}
			mine.Billboard = &space.BillBoard{
				Switch:           reply.Switch,
				Guided:           reply.Guided,
				CharacterUrl:     reply.CharacterUrl,
				BackgroundId:     reply.BackgroundId,
				FullscreenSwitch: reply.FullscreenSwitch,
			}
			return
		})
		// following and follower
		eg.Go(func(ctx context.Context) (err error) {
			var stat *relApi.StatReply
			if stat, err = s.relDao.Stat(ctx, mid); err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			if stat == nil {
				return
			}
			mine.Following = stat.Following
			mine.Follower = stat.Follower
			return
		})
		// new followers
		// iOS 538客户端渲染有bug，会导致闪退
		//6.22以后ios可出气泡动效
		if platform != "ios" || (plat == model.PlatIPhone && build > s.c.BuildLimit.NewFansIOSBuild) ||
			(plat == model.PlatIPhoneB && build > s.c.BuildLimit.NewFansIOSBBuild) {
			eg.Go(func(ctx context.Context) (err error) {
				if mine.NewFollowers, err = s.relsh1Dao.FollowersUnreadCount(ctx, mid); err != nil {
					log.Error("%+v", err)
					err = nil
					return
				}
				mine.NewFollowersRTime = s.c.Custom.NewFollowersRTime
				return
			})
		}
		// dynamic count
		eg.Go(func(ctx context.Context) (err error) {
			var count int64
			if count, err = s.bplusDao.DynamicCount(ctx, mid); err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			mine.Dynamic = count
			return
		})
		// bcoin
		eg.Go(func(ctx context.Context) (err error) {
			var bp float64
			if bp, err = s.payDao.UserWalletInfo(ctx, mid, platform); err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			mine.BCoin = bp
			return
		})
		// creative
		eg.Go(func(ctx context.Context) (err error) {
			var isUp, show int
			func() {
				// 兜底方案为默认展示投稿按钮和up主样式的创作中心
				mine.ShowVideoup = _creativeDefaultVal
				mine.ShowCreative = _creativeDefaultVal
				if isUp, show, err = s.voDao.Creative(ctx, mid); err != nil {
					log.Error("%+v", err)
					err = nil
					return
				}
				// 该字段用于控制是否展示创作中心标题栏（判断逻辑由创作中心控制 例如是否黑名单
				mine.ShowVideoup = show
				// 该字段用于控制是否展示创作中心内容栏（判断逻辑由创作中心控制 例如是否up主
				// 6.16之后走平台判断
				mine.ShowCreative = isUp
			}()
			func() {
				if !s.newCreativeControl(plat, build) { // 6.16之后走平台判断是否是up主
					return
				}
				mine.ShowCreative = _creativeDefaultVal
				if isUp, err = s.resDao.IsUp(ctx, mid); err != nil {
					log.Error("%+v", err)
					err = nil
					return
				}
				mine.ShowCreative = isUp
			}()
			return
		})
		var mutex sync.Mutex
		// white
		for _, v := range s.white[plat] {
			tmpID := v.ID
			tmpURL := v.URL
			eg.Go(func(ctx context.Context) (err error) {
				ok, err := s.accDao.UserCheck(ctx, mid, tmpURL)
				if err != nil {
					log.Error("s.accDao.UserCheck error(%+v)", err)
					err = nil
					return
				}
				if ok {
					mutex.Lock()
					whiteMap[tmpID] = true
					mutex.Unlock()
				}
				return
			})
		}
		// redDot
		for _, v := range s.redDot[plat] {
			tmpID := v.ID
			tmpURL := v.URL
			eg.Go(func(ctx context.Context) (err error) {
				ok, err := s.accDao.RedDot(ctx, mid, tmpURL)
				if err != nil {
					log.Error("s.accDao.RedDot error(%+v)", err)
					err = nil
					return
				}
				if ok {
					mutex.Lock()
					rdMap[tmpID] = true
					mutex.Unlock()
				}
				return
			})
		}
		// 直播中心模块
		if s.newLive(plat, build, teenagersMode, lessonsMode) {
			eg.Go(func(ctx context.Context) (err error) {
				var firstLiveTime int64
				if firstLiveTime, err = s.liveDao.LiveCenter(ctx, mid, build, platform); err != nil {
					log.Error("s.liveDao.LiveCenter err(%+v)", err)
					err = nil
					return
				}
				mine.FirstLiveTime = firstLiveTime
				return
			})
		}
		// 创作中心大模块
		if s.newCreative(plat, build) {
			if platform == "android" {
				eg.Go(func(ctx context.Context) (err error) {
					if ctCenter, err = s.voDao.AndroidCreative(ctx, mid, build); err != nil {
						log.Error("s.voDao.AndroidCreative mid(%d) err(%+v)", mid, err)
					}
					return nil
				})
			}
			if platform == "ios" {
				eg.Go(func(ctx context.Context) (err error) {
					if ctCenter, err = s.voDao.IOSCreative(ctx, mid, build); err != nil {
						log.Error("s.voDao.IOSCreative mid(%d) err(%+v)", mid, err)
					}
					return nil
				})
			}
		}
		eg.Go(func(ctx context.Context) error {
			reply, err := s.asDao.SeniorGate(ctx, mid, int64(build), mobiApp, device)
			if err != nil {
				log.Error("s.asDao.SeniorGate error(%+v), mid(%d)", err, mid)
				return nil
			}
			mine.SeniorGate = &space.SeniorGate{
				Identity:   int64(reply.Member),
				Text:       reply.Text,
				Url:        reply.Url,
				Mode:       reply.Mode,
				MemberText: reply.MemberText,
			}
			if s.c.Custom.BirthdaySwitchOn && isBirthday(time.Unix(int64(ps.Profile.Birthday), 0), time.Now()) && reply.Member == answerApi.SeniorGateResp_Senior {
				mine.SeniorGate.BirthdayConf = &space.BirthdayConf{
					Icon:       s.c.SeniorGateBirthday.Icon,
					Url:        s.c.SeniorGateBirthday.Url,
					BubbleText: s.c.SeniorGateBirthday.BubbleText,
				}
			}
			func() {
				if reply.Member != answerApi.SeniorGateResp_Senior {
					return
				}
				if int64(time.Since(time.Unix(reply.PassTime, 0)).Hours()) > s.c.Custom.SeniorGateExTime {
					return
				}
				mine.Achievement.SeniorGateFlash = &space.SeniorGateFlash{Icon: s.c.AchievementConf.SeniorGateIcon}
			}()
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			nftReply, err := s.accDao.NFTBatchInfo(ctx, &memberAPI.NFTBatchInfoReq{
				Mids:   []int64{mid},
				Status: "inUsing",
				Source: "face",
			})
			if err != nil {
				log.Error("s.accDao.NFTBatchInfo error(%+v), mid(%d)", err, mid)
				return nil
			}
			nftInfo, ok := nftReply.GetNftInfos()[strconv.FormatInt(mid, 10)]
			if !ok {
				return nil
			}
			region, err := s.galleryDao.GetNFTRegion(ctx, nftInfo.NftId)
			if err != nil {
				if ecode.Cause(err) != ecode.NothingFound {
					log.Error("s.panguDao.GetNFTRegion error(%+v), mid(%d), nftId(%s)", err, mid, nftInfo.NftId)
				}
				return nil
			}
			mine.NFT = &space.NFT{
				RegionType: int64(region.Type),
				NFTIcon: space.NFTIcon{
					ShowStatus: int64(region.ShowStatus),
					Url:        region.Icon,
				},
			}
			return nil
		})
	}
	if zoneLimitIDs, ok := s.authZoneLimitIDs[plat]; ok && len(zoneLimitIDs) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			req := &locmdl.ZoneLimitPoliciesReq{
				UserIp:       metadata.String(ctx, metadata.RemoteIP),
				DefaultAuths: zoneLimitIDs,
			}
			reply, err := s.loc.ZoneLimitPolicies(ctx, req)
			if err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			auths = reply.Auths
			return nil
		})
	}

	// when account info error,return,ingore else error
	// 大会员入口模块
	if platform != "" && teenagersMode == 0 && lessonsMode == 0 {
		eg.Go(func(ctx context.Context) (err error) {
			vipTip, err = s.vipDao.VipTips(c, mid, platform, mobiApp, device, build, buvid)
			if err != nil {
				log.Error("s.vipDao.VipTips err(+%v)", err)
			}
			return nil
		})
	}
	var fansPeak int64
	eg.Go(func(ctx context.Context) error {
		fansPeak, err = s.relDao.PeakStats(ctx, mid)
		if err != nil {
			log.Error("s.relDao.PeakStats err(%+v)", err)
		}
		return nil
	})
	// 必连入口
	eg.Go(func(ctx context.Context) error {
		req := &ott.BiliLinkEntryReq{
			Build:       int64(build),
			Mid:         mid,
			Platform:    platform,
			MobiApp:     mobiApp,
			Channel:     channel,
			BiliLinkNew: biliLinkNew,
		}
		reply, err := s.ottclient.BiliLinkEntry(ctx, req)
		if err != nil {
			log.Error("s.ottclient.BiliLinkEntry error(%+v), mid(%d)", err, mid)
			return nil
		}
		mine.EnableBiliLink = reply.Show
		mine.BiliLinkBubble = reply.BiliLinkBubble
		return nil
	})
	if (plat == model.PlatIpadHD || plat == model.PlatIPad || plat == model.PlatAndroidHD) && mid > 0 {
		mine.SectionUpdateTime = new(space.SectionUpdateTime)
		eg.Go(func(ctx context.Context) error {
			//稍后再看
			var err error
			if mine.SectionUpdateTime.ToView, err = s.toViewDao.LastToViewTime(ctx, mid); err != nil {
				log.Error("%+v", err)
				return nil
			}
			return nil
		})

		eg.Go(func(ctx context.Context) error {
			//收藏
			var err error
			if mine.SectionUpdateTime.Favorite, err = s.favDao.LastFavTime(ctx, mid); err != nil {
				log.Error("%+v", err)
				return nil
			}
			return nil
		})

		eg.Go(func(ctx context.Context) error {
			lastFollowingTime, err := s.relDao.FetchLastFollowingTime(ctx, mid)
			if err != nil {
				log.Error("s.relDao.FetchLastFollowingTime error(%+v) mid(%d)", err, mid)
				return nil
			}
			mine.SectionUpdateTime.Following = lastFollowingTime
			return nil
		})

		eg.Go(func(ctx context.Context) error {
			hisRes, err := s.hisDao.History(ctx, mid, 1, 1)
			if err != nil {
				log.Error("s.hisDao.History error(%+v) mid(%d)", err, mid)
				return nil
			}
			if len(hisRes) == 0 {
				return nil
			}
			mine.SectionUpdateTime.History = hisRes[0].Unix
			return nil
		})
	}
	if pd.WithContext(c).IsPlatAndroid().MustFinish() {
		eg.Go(func(ctx context.Context) error {
			gameTipsReply, err := s.gameDao.FetchGameTip(ctx, mid, int64(build), int64(plat), buvid)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			gameTipsReply = s.gameTipsDmpFilter(ctx, gameTipsReply, mid)
			mine.GameTip = gameTipsReply
			return nil
		})
	}
	if egErr := eg.Wait(); egErr != nil {
		log.Error("%+v", egErr)
	}
	if s.newMine(plat, build) {
		mgrPlat := s.convertInterfacePlatToMgr(plat)
		mineSections, err = s.resDao.MineSectionsNewmont(c, mid, int32(mgrPlat), int32(build), channel, lang, mine.ShowCreative, mine.FirstLiveTime, fansPeak, buvid)
		if err != nil {
			log.Error("s.resDao.MineSections err(%+v)", err)
			err = nil
		}
	}
	return
}

func (s *Service) gameTipsDmpFilter(ctx context.Context, in []*space.GameTip, mid int64) []*space.GameTip {
	var (
		gameTips = space.GameTips(in)
		dmpIDs   = getDmpIds(gameTips)
		out      []*space.GameTip
	)
	if len(dmpIDs) == 0 {
		return in
	}
	gameDmps, err := s.gameDao.FetchGameDmps(ctx, mid, dmpIDs)
	if err != nil { //如果游戏人群包接口出错则过滤掉所有定向人群包的运营位
		gameTips.Iter(func(g *space.GameTip) {
			if g.IsDirected == 1 {
				out = append(out, g)
			}
		})
		log.Error("s.gameDao.FetchGameDmps error(%+v) mid(%d)", err, mid)
		return out
	}
	//取出游戏人群包的结果做校验
	dmpsResMap := make(map[int64]int64, 0)
	for _, v := range gameDmps {
		if v == nil {
			continue
		}
		dmpsResMap[v.ID] = v.Val
	}
	gameTips.Iter(func(g *space.GameTip) {
		if g.IsDirected == 1 { //没有人群包限制
			out = append(out, g)
			return
		}
		if dmpsResMap[g.DmpId] == 1 { //有人群包限制但是命中了人群包
			out = append(out, g)
		}
	})
	return out
}

func getDmpIds(gameTips space.GameTips) []int64 {
	var dmpIds []int64
	gameTips.Iter(func(g *space.GameTip) {
		if g.IsDirected == 1 {
			return
		}
		dmpIds = append(dmpIds, g.DmpId)
	})
	return dmpIds
}

func (s *Service) buildDefaultCreative(plat int8) (items []*space.SectionItem, upTitle, beUpTitle string) {
	upTitle = s.c.Creative.DefaultUpTitle
	beUpTitle = s.c.Creative.DefaultBeUpTitle
	if plat == model.PlatIPhone || plat == model.PlatIPhoneB {
		for _, v := range s.c.IOSPreIcons {
			item := &space.SectionItem{
				ID:           v.ID,
				Title:        v.Title,
				URI:          v.URL,
				Icon:         v.Icon,
				NeedLogin:    v.NeedLogin,
				RedDot:       v.RedDot,
				GlobalRedDot: v.GlobalRedDot,
				Display:      v.Display,
			}
			items = append(items, item)
		}
		return
	}
	if plat == model.PlatAndroid || plat == model.PlatAndroidB {
		for _, v := range s.c.AndPreIcons {
			item := &space.SectionItem{
				ID:           v.ID,
				Title:        v.Title,
				URI:          v.URL,
				Icon:         v.Icon,
				NeedLogin:    v.NeedLogin,
				RedDot:       v.RedDot,
				GlobalRedDot: v.GlobalRedDot,
				Display:      v.Display,
			}
			items = append(items, item)
		}
		return
	}
	return
}

//nolint:gocognit
func (s *Service) newSecs(mineSec []*newmentapi.Section, build, teenagersMode, lessonsMode int, filtered bool, plat int8, liveCenter []*space.SectionItem, ctCenter *space.SectionV2, isup int, firstLiveTime int64, ctip *space.ContributeTip, hasGameCenter bool) (sections []*space.SectionV2) {
	for _, ms := range mineSec {
		if ms == nil || ms.Style == _mineOperate || ms.Style == _mineBubble { //创作/直播运营位单独处理
			continue
		}
		// 判断审核态下是否展示
		if filtered && ms.AuditShow == 0 {
			continue
		}
		var (
			items                                 []*space.SectionItem
			mngIconMap                            = make(map[int64]*space.MngRes)
			tmpMngInfo                            *space.MngInfo
			upTitle, beUpTitle, tipIcon, tipTitle string
			secType                               int32
		)
		for _, si := range ms.Items {
			if ok := s.whiteList(si.Uri, false, teenagersMode, lessonsMode); !ok {
				continue
			}
			//有游戏中心强入口时，不下发推荐服务中的游戏中心入口（只有安卓有游戏中心入口，需要过滤）
			if hasGameCenter && si.Id == _androidRecommendGameCenter {
				continue
			}
			// 青少年模式下不展示课堂模式入口
			if teenagersMode != 0 && si.Uri == _lessonsmodeURI {
				continue
			}
			// 审核态不展示离线缓存
			if si.Uri == _downloadURI && filtered {
				continue
			}
			tmpItem := &space.SectionItem{
				ID:           si.Id,
				Title:        si.Title,
				Icon:         si.Icon,
				NeedLogin:    int8(si.NeedLogin),
				URI:          si.Uri,
				RedDot:       int8(si.RedDot),
				GlobalRedDot: int8(si.GlobalRedDot),
				RedDotForNew: si.RedDotForNew,
				CommonOpItem: &space.CommonOpItem{
					LinkType:           int64(si.OpLinkType),
					TitleIcon:          si.OpTitleIcon,
					Title:              si.OpTitle,
					SubTitle:           si.OpSubTitle,
					LinkIcon:           si.OpLinkIcon,
					BackgroundColor:    si.OpBackgroundColor,
					TitleColor:         si.OpTitleColor,
					LinkContainerColor: si.OpLinkContainerColor,
					Text:               si.OpLinkText,
				},
			}
			if si.MngIcon != nil {
				tmpItem.MngRes = &space.MngRes{
					Icon:   si.MngIcon.Icon,
					IconID: si.MngIcon.Id,
				}
			}
			mngIconMap[si.Id] = tmpItem.MngRes
			items = append(items, tmpItem)
		}
		// 如果新版创作中心，则信任业务方接口，老版本走后台配置
		// 6.16之后文案由网关下发
		if _, ok := model.CreativeModules[ms.Id]; ok {
			secType = 1
			items, upTitle, beUpTitle = s.buildDefaultCreative(plat)
			if ctCenter != nil {
				items = ctCenter.Items
				upTitle = ctCenter.UpTitle
				beUpTitle = ctCenter.BeUpTitle
				tipIcon = ctCenter.TipIcon
				tipTitle = ctCenter.TipTitle
			}
			if s.newCreativeControl(plat, build) {
				// 同下,防止新老版本up主判断不一致
				upTitle = s.c.Creative.UpTitle
				beUpTitle = ""
				tipIcon = ""
				tipTitle = ""
				if ctip != nil { // 后台有配置下发，说明非up主
					beUpTitle = ctip.BeUpTitle
					tipIcon = ctip.TipIcon
					tipTitle = ctip.TipTitle
					ms.ButtonName = ctip.ButtonText
					if ctip.ButtonUrl != "" {
						ms.ButtonUrl = ctip.ButtonUrl
					}
					// 防止老版本判断是up主但是新版本判断不是up主还会下发老版本中的upTitle字段
					upTitle = ""
				}
			}
			//6.20之后，如果是up主而且开播过出新样式:直播中心和创作中心融合
			if s.mixCreativeControl(plat, build, isup, firstLiveTime) {
				for _, item := range liveCenter {
					item.Display = 1 //创作中心模块下，客户端会判断display来展示item
				}
				items = append(items, liveCenter...)
			}
		}
		// 如果是直播中心，则信任业务方接口，老版本无直播
		if _, ok := model.LiveModules[ms.Id]; ok {
			if !s.newLive(plat, build, teenagersMode, lessonsMode) {
				continue
			}
			//6.20出新样式之后就不展示直播中心
			if s.mixCreativeControl(plat, build, isup, firstLiveTime) {
				continue
			}
			items = liveCenter
		}
		//游戏中心secType=2
		if _, ok := s.c.GameModuleID[strconv.FormatInt(ms.Id, 10)]; ok {
			secType = 2
		}
		//通用运营位secType=3
		if ms.Style == _mineCommonOp {
			secType = 3
		}
		if len(items) == 0 && secType != 1 { // 保证创作中心投稿按钮正常下发
			continue
		}
		for k, i := range items {
			mi, ok := mngIconMap[i.ID]
			if !ok {
				continue
			}
			items[k].MngRes = mi
		}
		if ms.IsMng == 1 {
			tmpMngInfo = &space.MngInfo{
				TitleColor:      ms.TitleColor,
				Subtitle:        ms.Subtitle,
				SubtitleURL:     ms.SubtitleUrl,
				SubtitleColor:   ms.SubtitleColor,
				Background:      ms.Background,
				BackgroundColor: ms.BackgroundColor,
			}
		}
		tmpSec := &space.SectionV2{
			Title: ms.Title,
			Items: items,
			Style: ms.Style,
			Button: &space.Button{
				Text:  ms.ButtonName,
				URL:   ms.ButtonUrl,
				Icon:  ms.ButtonIcon,
				Style: ms.ButtonStyle,
			},
			MngInfo:   tmpMngInfo,
			UpTitle:   upTitle,
			BeUpTitle: beUpTitle,
			Type:      secType,
			TipTitle:  tipTitle,
			TipIcon:   tipIcon,
		}
		sections = append(sections, tmpSec)
	}
	return
}

//nolint:gocognit
func (s *Service) sections(c context.Context, whiteMap, rdMap map[int64]bool, mid int64, build, teenagersMode, lessonsMode int, filtered bool, plat int8, lang string, liveCenter []*space.SectionItem, ctCenter *space.SectionV2, auths map[int64]*locmdl.ZoneLimitAuth) (sections []*space.Section) {
	menus := model.IPhoneNormalMenu[model.PlatIPhone]
	if m, ok := model.IPhoneNormalMenu[plat]; ok {
		menus = m
	}
	if fm, ok := model.IPhoneFilterMenu[plat]; ok && filtered {
		menus = fm
	}
	var (
		sids []int64
	)
	for _, module := range menus {
		var items []*space.SectionItem
		if _, ok := model.CreativeModules[int64(module)]; ok && s.newCreative(plat, build) {
			if ctCenter != nil {
				items = ctCenter.Items
			}
		} else if _, ok := model.LiveModules[int64(module)]; ok && s.newLive(plat, build, teenagersMode, lessonsMode) {
			items = liveCenter
		} else {
			key := fmt.Sprintf(_initSidebarKey, plat, module, lang)
			ss, ok := s.sectionCache[key]
			if !ok {
				continue
			}
			for _, si := range ss {
				ignore := false
				if !si.CheckLimit(build) {
					continue
				}
				if ok := s.whiteList(si.Item.Param, false, teenagersMode, lessonsMode); !ok {
					continue
				}
				// 地区限制
				// areaInt, _ := strconv.ParseInt(si.Item.Area, 10, 64)
				// if areaInt > 0 {
				// 	if auth, ok := auths[areaInt]; ok && auth.Play != int64(locmdl.Status_Forbidden) {
				// 		continue
				// 	}
				// }
				if si.Item.AreaPolicy > 0 {
					if auth, ok := auths[si.Item.AreaPolicy]; ok && auth.Play == locmdl.Status_Forbidden {
						continue
					}
				}
				// 青少年模式下不展示课堂模式入口
				if teenagersMode != 0 && si.Item.Param == _lessonsmodeURI {
					continue
				}
				if si.Item.Name == "离线缓存" && filtered {
					continue
				}
				if si.Item.Name == "直播中心" && mid == 0 { // 针对直播中心特殊处理白名单逻辑 未登录用户都展示
					ignore = true
				}
				// 有白名单的接口，未登录默认展示
				// TODO 上线后跟产品确认直播中心配置好后可以把上面的临时判断删除
				if si.Item.WhiteURLShow == 1 && mid == 0 {
					ignore = true
				}
				if !ignore && si.Item.WhiteURL != "" && !whiteMap[si.Item.ID] {
					continue
				}
				tmpItem := &space.SectionItem{
					ID:        si.Item.ID,
					Title:     si.Item.Name,
					Icon:      si.Item.Logo,
					NeedLogin: si.Item.NeedLogin,
					URI:       si.Item.Param,
				}
				if si.Item.Red != "" && rdMap[si.Item.ID] {
					tmpItem.RedDot = 1
				}
				items = append(items, tmpItem)
			}
		}
		if len(items) == 0 {
			continue
		}
		for _, v := range items {
			if v.ID == 0 {
				continue
			}
			sids = append(sids, v.ID)
		}
		sections = append(sections, &space.Section{
			Tp:    model.IPhoneMenuTp[module],
			Title: model.IPhoneMenu[module],
			Items: items,
		})
	}
	if len(sections) > 0 {
		sections = s.FromSections(c, sids, mid, plat, build, "", sections, false)
	}
	return
}

//nolint:gocognit
func (s *Service) ipadSections(_ context.Context, whiteMap, rdMap map[int64]bool, _ int64, build, teenagersMode, lessonsMode int, filtered bool, plat int8, mobiApp, lang string, auths map[int64]*locmdl.ZoneLimitAuth) (ipadSections, ipadUpperSections []*space.SectionItem) {
	menus := model.IPadNormalMenu[plat]
	if filtered {
		menus = model.IPadFilterMenu[plat]
	}
	for _, module := range menus {
		key := fmt.Sprintf(_initSidebarKey, plat, module, lang)
		ss, ok := s.sectionCache[key]
		if !ok {
			continue
		}
		teenMap := make(map[string]struct{})
		lessonMap := make(map[string]struct{})
		for _, si := range ss {
			if !si.CheckLimit(build) {
				continue
			}
			// iPad HD 入口展示
			// 老逻辑 > 12200 展示青少年入口 需要登录
			// 新逻辑 > 12200 & < 12220    展示青少年入口 需要登录
			//       > 12210              展示青少年入口 不需要登录
			// 后台粉版ipad 可能会配置多条青少年记录，故用 teenMap 做去重拦截
			if mobiApp == "ipad" && si.Item.Param == _teenagersmodeURI {
				if _, ok := teenMap[_teenagersmodeURI]; ok || build <= 12200 {
					continue
				}
				//nolint:gomnd
				if build > 12210 {
					si.Item.NeedLogin = 0
				}
				teenMap[_teenagersmodeURI] = struct{}{}
			}
			// iPad HD 课堂模式入口展示
			// build <= 12360 不展示
			if mobiApp == "ipad" && si.Item.Param == _lessonsmodeURI {
				if _, ok := lessonMap[_lessonsmodeURI]; ok || build <= 12360 {
					continue
				}
				lessonMap[_lessonsmodeURI] = struct{}{}
			}
			if si.Item.Name == "离线缓存" && filtered {
				continue
			}
			if ok := s.whiteList(si.Item.Param, true, teenagersMode, lessonsMode); !ok {
				continue
			}
			// 地区限制
			// areaInt, _ := strconv.ParseInt(si.Item.Area, 10, 64)
			// if auth, ok := auths[areaInt]; ok && auth.Play != int64(locmdl.Status_Forbidden) {
			// 	continue
			// }
			if si.Item.AreaPolicy > 0 {
				if auth, ok := auths[si.Item.AreaPolicy]; ok && auth.Play == locmdl.Status_Forbidden {
					continue
				}
			}
			// 青少年模式下不展示课堂模式入口
			if teenagersMode != 0 && si.Item.Param == _lessonsmodeURI {
				continue
			}
			if si.Item.WhiteURL != "" && !whiteMap[si.Item.ID] {
				continue
			}
			tmpItem := &space.SectionItem{
				ID:        si.Item.ID,
				Title:     si.Item.Name,
				Icon:      si.Item.Logo,
				NeedLogin: si.Item.NeedLogin,
				URI:       si.Item.Param,
			}
			if si.Item.Red != "" && rdMap[si.Item.ID] {
				tmpItem.RedDot = 1
			}
			if _, ok := model.CreativeModules[int64(module)]; ok {
				ipadUpperSections = append(ipadUpperSections, tmpItem)
			} else {
				ipadSections = append(ipadSections, tmpItem)
			}
		}
	}
	return
}

//nolint:gocognit
func (s *Service) androidSections(c context.Context, whiteMap, rdMap map[int64]bool, mid int64, build, teenagersMode, lessonsMode int, plat int8, lang, channel string, liveCenter []*space.SectionItem, ctCenter *space.SectionV2, auths map[int64]*locmdl.ZoneLimitAuth) (sections []*space.Section) {
	menus := model.AndroidMenu[model.PlatAndroid]
	if m, ok := model.AndroidMenu[plat]; ok {
		menus = m
	}
	var (
		sids []int64
	)
	for _, module := range menus {
		var items []*space.SectionItem
		if _, ok := model.CreativeModules[int64(module)]; ok && s.newCreative(plat, build) {
			if ctCenter != nil {
				items = ctCenter.Items
			}
		} else if _, ok := model.LiveModules[int64(module)]; ok && s.newLive(plat, build, teenagersMode, lessonsMode) {
			items = liveCenter
		} else {
			key := fmt.Sprintf(_initSidebarKey, plat, module, lang)
			ss, ok := s.sectionCache[key]
			if !ok {
				continue
			}
			for _, si := range ss {
				ignore := false
				if !si.CheckLimit(build) {
					continue
				}
				if ok := s.whiteList(si.Item.Param, false, teenagersMode, lessonsMode); !ok {
					continue
				}
				// 地区限制
				// areaInt, _ := strconv.ParseInt(si.Item.Area, 10, 64)
				// if auth, ok := auths[areaInt]; ok && auth.Play != int64(locmdl.Status_Forbidden) {
				// 	continue
				// }
				if si.Item.AreaPolicy > 0 {
					if auth, ok := auths[si.Item.AreaPolicy]; ok && auth.Play == locmdl.Status_Forbidden {
						continue
					}
				}
				// 青少年模式下不展示课堂模式入口
				if teenagersMode != 0 && si.Item.Param == _lessonsmodeURI {
					continue
				}
				if si.Item.Name == "直播中心" && mid == 0 { // 针对直播中心特殊处理白名单逻辑 未登录用户都展示
					ignore = true
				}
				if !ignore && si.Item.WhiteURL != "" && !whiteMap[si.Item.ID] {
					continue
				}
				tmpItem := &space.SectionItem{
					ID:           si.Item.ID,
					Title:        si.Item.Name,
					Icon:         si.Item.Logo,
					NeedLogin:    si.Item.NeedLogin,
					URI:          si.Item.Param,
					GlobalRedDot: si.Item.GlobalRed,
				}
				if si.Item.Red != "" && rdMap[si.Item.ID] {
					tmpItem.RedDot = 1
				}
				items = append(items, tmpItem)
			}
		}
		if len(items) == 0 {
			continue
		}
		for _, v := range items {
			if v.ID == 0 {
				continue
			}
			sids = append(sids, v.ID)
		}
		sections = append(sections, &space.Section{Items: items})
	}
	if len(sections) > 0 {
		sections = s.FromSections(c, sids, mid, plat, build, channel, sections, true)
	}
	return
}

// Myinfo simple myinfo
func (s *Service) Myinfo(c context.Context, mid int64, mobiApp string) (myinfo *space.Myinfo, err error) {
	var pf *accmdl.ProfileStatReply
	myinfo = new(space.Myinfo)
	if pf, err = s.accDao.Profile3(c, mid); err != nil || pf.Profile == nil {
		log.Error("s.accDao.Profile3 err(%+v) or pf.Profile is nil", err)
		return
	}
	myinfo.Coins = pf.Coins
	myinfo.Sign = pf.Profile.Sign
	myinfo.IsTourist = pf.Profile.IsTourist
	myinfo.PinPrompting = pf.Profile.PinPrompting
	myinfo.InRegAudit = pf.Profile.InRegAudit
	switch pf.Profile.Sex {
	case "男":
		myinfo.Sex = 1
	case "女":
		myinfo.Sex = 2
	default:
		myinfo.Sex = 0
	}
	myinfo.Mid = mid
	myinfo.Birthday = pf.Profile.Birthday.Time().Format("2006-01-02")
	myinfo.Name = pf.Profile.Name
	myinfo.Face = pf.Profile.Face
	myinfo.FaceNftNew = pf.Profile.FaceNftNew
	myinfo.Rank = pf.Profile.Rank
	myinfo.Level = pf.Profile.Level
	myinfo.Vip = pf.Profile.Vip
	if pf.Profile.Vip.Status == model.VipStatusExpire && pf.Profile.Vip.DueDate > 0 { // 0-过期
		myinfo.Vip.Label.Path = model.VipLabelExpire
	}
	myinfo.Silence = pf.Profile.Silence
	myinfo.EmailStatus = pf.Profile.EmailStatus
	myinfo.TelStatus = pf.Profile.TelStatus
	myinfo.Official = pf.Profile.Official
	//nolint:gomnd
	if myinfo.Official.Role == 7 {
		myinfo.Official.Role = 1
	}
	myinfo.Identification = pf.Profile.Identification
	if pf.Profile.Pendant.Image != "" {
		myinfo.Pendant = &space.Pendant{Image: pf.Profile.Pendant.Image, ImageEnhance: pf.Profile.Pendant.ImageEnhance}
	}
	eg := errgroup.WithContext(c)
	if pf.Profile.Silence == 1 {
		eg.Go(func(ctx context.Context) (e error) {
			if myinfo.EndTime, err = s.accDao.BlockTime(ctx, mid); err != nil {
				log.Error("s.accDao.BlockTime err(%+v)", err)
				err = nil
			}
			return
		})
	}
	if pf.Profile.Rank < _rankMember {
		eg.Go(func(ctx context.Context) (err error) {
			var answer *answerApi.AnswerStatus
			if answer, err = s.asDao.AnswerStatus(ctx, mid, mobiApp, model.AnswerSourceMyinfo); err != nil {
				log.Error("s.asDao.AnswerStatus err(%+v)", err)
				err = nil
				return
			}
			if answer != nil && answer.Status != _answered {
				myinfo.AnswerStatus = answer.Status
			}
			return
		})
	}
	invite := &space.Invite{Display: false, InviteRemind: 0}
	eg.Go(func(ctx context.Context) error {
		if inviteCount, e := s.usersuitDao.InviteCountStat(ctx, mid); e == nil && inviteCount != nil {
			if inviteCount.CurrentLimit > 0 {
				invite.InviteRemind = inviteCount.CurrentRemain
				invite.Display = true
			}
		}
		myinfo.Invite = invite
		return nil
	})
	// 是否有nft头像
	eg.Go(func(ctx context.Context) error {
		ownerReply, err := s.galleryDao.IsNFTFaceOwner(ctx, &gallerygrpc.MidReq{Mid: mid, RealIp: metadata.RemoteIP})
		if err != nil {
			log.Error("s.galleryDao.AccountHasNFT mid(%d) err(%+v)", mid, err)
			return nil
		}
		if ownerReply.Status == gallerygrpc.OwnerStatus_ISOWNER {
			myinfo.HasFaceNft = true
		}
		return nil
	})
	//nolint:errcheck
	eg.Wait()
	return
}

//nolint:bilirailguncheck
func (s *Service) ConfigSet(c context.Context, mid int64, buvid string, adSpecial int, sensorAccess int) (err error) {
	//老版本的sensorAccess为-1
	if pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().And().Build("<", int64(67600000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build("<", int64(6760000))
	}).FinishOr(true) {
		sensorAccess = -1
	}
	data := &space.ConfigSet{
		Mid:          mid,
		Buvid:        buvid,
		AdSpecial:    adSpecial,
		SensorAccess: sensorAccess,
	}
	if err = s.configSetPub.Send(c, strconv.FormatInt(mid, 10), data); err != nil {
		log.Error("configSet(%v) error(%v)", data, err)
		return
	}
	log.Warn("configSet success %v", data)
	return
}

func (s *Service) FromSections(c context.Context, sids []int64, mid int64, plat int8, build int, channel string, sections []*space.Section, filterHidden bool) (res []*space.Section) {
	hiddenMap := make(map[int64]bool, len(sids))
	iconMap := make(map[int64]*resApi.MngIcon, len(sids))
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if iconMap, err = s.resDao.MngIcon(ctx, sids, mid, plat); err != nil {
			log.Error("s.resDao.MngIcon err(%+v)", err)
		}
		return nil
	})
	if filterHidden {
		eg.Go(func(ctx context.Context) (err error) {
			if hiddenMap, err = s.resDao.EntrancesIsHidden(ctx, sids, build, plat, channel); err != nil {
				log.Error("s.resDao.EntrancesIsHidden err(%+v)", err)
			}
			return nil
		})
	}
	//nolint:errcheck
	eg.Wait()
	for _, ts := range sections {
		var resItem []*space.SectionItem
		for _, tsi := range ts.Items {
			// 判断是否在入口屏蔽配置里
			if isHidden, ok := hiddenMap[tsi.ID]; ok && isHidden {
				continue
			}
			if ic, ok := iconMap[tsi.ID]; ok && ic != nil {
				tsi.MngRes = &space.MngRes{Icon: ic.Icon, IconID: ic.Id}
				if ic.GlobalRed == 1 {
					tsi.GlobalRedDot = 1
				}
			}
			resItem = append(resItem, tsi)
		}
		if len(resItem) == 0 {
			continue
		}
		ts.Items = resItem
		res = append(res, ts)
	}
	return res
}

func (s *Service) newCreative(plat int8, build int) bool {
	if (plat == model.PlatAndroid && build >= s.c.BuildLimit.AndMineCreative) || plat == model.PlatAndroidB ||
		(plat == model.PlatIPhone && build > s.c.BuildLimit.IOSMineCreative) || (plat == model.PlatIPhoneB && build > s.c.BuildLimit.IPhoneBMineCreative) {
		return true
	}
	return false
}

func (s *Service) newLive(plat int8, build, teenagersMode, lessonsMode int) bool {
	if teenagersMode == 0 && lessonsMode == 0 && ((plat == model.PlatIPhone && build > s.c.BuildLimit.IOSMineLive) || (plat == model.PlatAndroid && build > s.c.BuildLimit.AndMineLive)) {
		return true
	}
	return false
}

func (s *Service) newMine(plat int8, build int) bool {
	if (plat == model.PlatIPhone && build > s.c.BuildLimit.NewMineIOSBuild) || (plat == model.PlatAndroid && build >= s.c.BuildLimit.NewMineAndBuild) ||
		(plat == model.PlatIPhoneB && build >= s.c.BuildLimit.NewMineIPhoneBBuild) || (plat == model.PlatAndroidB && build >= s.c.BuildLimit.NewMineAndBBuild) ||
		(plat == model.PlatAndroidI && build > s.c.BuildLimit.NewMineAndIBuild) || (plat == model.PlatIPhoneI) ||
		(plat == model.PlatIPad && build > s.c.BuildLimit.NewMineIPad) || (plat == model.PlatIpadHD && build > s.c.BuildLimit.NewMineIPadHD) ||
		(plat == model.PlatAndroidHD && build > s.c.BuildLimit.NewMineAndroidHD) {
		return true
	}
	return false
}

func (s *Service) newCreativeControl(plat int8, build int) bool {
	if (plat == model.PlatIPhone && build > s.c.BuildLimit.NewCreativeIOSBuild) ||
		(plat == model.PlatAndroid && build > s.c.BuildLimit.NewCreativeAndBuild) ||
		(plat == model.PlatAndroidB && build > s.c.BuildLimit.NewCreativeAndBBuild) ||
		(plat == model.PlatIPhoneB && build > s.c.BuildLimit.NewCreativeIOSBBuild) {
		return true
	}
	return false
}

func (s *Service) mixCreativeControl(plat int8, build, isUp int, firstLiveTime int64) bool {
	if ((plat == model.PlatIPhone && build > s.c.BuildLimit.NewMixCreativeIOSBuild) ||
		(plat == model.PlatAndroid && build > s.c.BuildLimit.NewMixCreativeAndBuild) ||
		(plat == model.PlatAndroidB && build > s.c.BuildLimit.NewMixCreativeAndBBuild) ||
		(plat == model.PlatIPhoneB && build > s.c.BuildLimit.NewMixCreativeIOSBBuild)) && isUp == 1 && firstLiveTime > 0 {
		return true
	}
	return false
}

func (s *Service) BiliLinkReport(c context.Context, report *bl.BiliLinkReport) error {
	req := &ott.BiliLinkReportReq{
		ActType: report.ActType,
		Id:      report.Id,
		Mid:     report.Mid,
	}
	if _, err := s.ottclient.BiliLinkReport(c, req); err != nil {
		return err
	}
	return nil
}

func (s *Service) NFTSettingButton(c context.Context, req *acm.NFTSettingButtonReq) (*acm.NFTSettingButtonReply, error) {
	nftRes, err := s.galleryDao.OwnNFTStatus(c, req.Mid, req.MobiApp)
	if err != nil {
		log.Error("s.galleryDao.OwnNFTStatus error(%+v), mid(%d)", err, req.Mid)
		return nil, err
	}
	if nftRes.OwnStatus == 0 {
		return &acm.NFTSettingButtonReply{}, nil
	}
	text := nftRes.ButtonChs
	if i18n.PreferTraditionalChinese(c, req.SLocale, req.CLocale) {
		text = nftRes.ButtonCht
	}
	return &acm.NFTSettingButtonReply{
		Url:  nftRes.JumpUrl,
		Text: text,
	}, nil
}

func isBirthday(birthday time.Time, now time.Time) bool {
	if birthday.Month() != now.Month() {
		return false
	}
	if birthday.Day() != now.Day() {
		return false
	}
	return true
}
