package dynamicV2

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	bcgmdl "go-gateway/app/app-svr/app-dynamic/interface/model/bcg"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	submdl "go-gateway/app/app-svr/app-dynamic/interface/model/subscription"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
)

const (
	_emojiRex = `[[][^\[\]]+[]]`
	_mailRex  = `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	_webRex   = `(http(s)?://)?([a-z0-9A-Z-]+\.)?(bilibili\.(com|tv|cn)|biligame\.(com|cn|net)|(bilibiliyoo|im9)\.com|biliapi\.net|b23\.tv|bili22\.cn|bili33\.cn|bili23\.cn|bili2233\.cn|(sugs\.suning\.com)|kaola\.com|bigfun\.cn|mcbbs\.net|mp\.weixin\.qq\.com|static\.cdsb\.com|bjnews\.com\.cn|720yun\.com|cctv\.com|jueze2021\.peopleapp\.com)($|/)([/.$*?~=#!%@&A-Za-z0-9_-]*)`
	_avRex    = `(AV|av|Av|aV)[0-9]+`
	_bvRex    = `(BV|bv|Bv|bV)1[1-9A-NP-Za-km-z]{9}`
	_cvRex    = `((CV|cv|Cv|cV)[0-9]+|(mobile/[0-9]+))`
	_vcRex    = `(VC|vc|Vc|vC)[0-9]+`
	_ogvssRex = `(SS|ss|Ss|sS)[0-9]+`
	_ogvepRex = `(EP|ep|Eo|eP)[0-9]+`
	_topicRex = `#[^#@\r\n]{1,32}#`

	// _shortURLRex   = `(?i)https://(b23.tv|bili22.cn|bili33.cn|bili23.cn|bili2233.cn)/[1-9A-NP-Za-km-z]{6,10}($|/|)([/.$*?~=#!%@&-A-Za-z0-9_]*)`
	_shortURLRex   = `(?i)https://(b23.tv|bili22.cn|bili33.cn|bili23.cn|bili2233.cn)/[0-9A-NP-Za-km-z]{6,10}($|/|)([/.$*?~=#!%@&-A-Za-z0-9_]*)`
	_ogvURLRex     = `(?i)((http(s)?://)?((uat-)?www.bilibili.com/bangumi/(play/|media/)|(b23.tv|bili22.cn|bili33.cn|bili23.cn|bili2233.cn)/)(ss|ep)[0-9]+)($|/|)([/.$*?~=#!%@&-A-Za-z0-9_]*)`
	_ugcURLRex     = `(?i)(http(s)?://)?(((uat-)?www.bilibili.com)|(b23.tv|bili22.cn|bili33.cn|bili23.cn|bili2233.cn))(/video)?/((av[0-9]+)|((BV)1[1-9A-NP-Za-km-z]{9}))($|/|)([/.$*?~=#!%@&-A-Za-z0-9_]*)`
	_articleURLRex = `(?i)(http(s)?://)?(uat-)?www.bilibili.com/read/((cv[0-9]+)|(native\?id=[0-9]+)|(app/[0-9]+)|(native/[0-9]+)|(mobile/[0-9]+))($|/|)([/.$*?~=#!%@&-A-Za-z0-9_]*)`
	_idRex         = `[\d]+`
)

// 对于经常使用的正则表达式保存regexp对象
// 避免每次使用时重新构建状态机
var (
	_emojiRgx = regexp.MustCompile(_emojiRex)
	_mailRgx  = regexp.MustCompile(_mailRex)
	_webRgx   = regexp.MustCompile(_webRex)
	_avRgx    = regexp.MustCompile(_avRex)
	_bvRgx    = regexp.MustCompile(_bvRex)
	_cvRgx    = regexp.MustCompile(_cvRex)
	_vcRgx    = regexp.MustCompile(_vcRex)
	_ogvssRgx = regexp.MustCompile(_ogvssRex)
	_ogvepRgx = regexp.MustCompile(_ogvepRex)
	_topicRgx = regexp.MustCompile(_topicRex)

	_shortURLRgx   = regexp.MustCompile(_shortURLRex)
	_ogvURLRgx     = regexp.MustCompile(_ogvURLRex)
	_ugcURLRgx     = regexp.MustCompile(_ugcURLRex)
	_articleURLRgx = regexp.MustCompile(_articleURLRex)
	_idRgx         = regexp.MustCompile(_idRex)
)

func (s *Service) getDesc(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) string {
	var desc string
	if dynCtx.Dyn.IsForward() {
		desc = s.descriptionForward(dynCtx, general)
	}
	if dynCtx.Dyn.IsAv() {
		desc = s.descriptionAV(dynCtx, general)
	}
	if dynCtx.Dyn.IsWord() {
		desc = s.descriptionWord(dynCtx, general)
	}
	if dynCtx.Dyn.IsDraw() {
		desc = s.descriptionDraw(dynCtx, general)
	}
	if dynCtx.Dyn.IsArticle() {
		desc = s.descriptionArticle(dynCtx, general)
	}
	if dynCtx.Dyn.IsMusic() {
		desc = s.descriptionMusic(dynCtx, general)
	}
	if dynCtx.Dyn.IsCommon() {
		common, ok := dynCtx.GetResCommon(dynCtx.Dyn.Rid)
		if !ok {
			xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "data_faild")
			log.Warn("description miss 1 mid(%v) dynid(%v) type(%v) rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Type, dynCtx.Dyn.Rid)
			return ""
		}
		if common.Vest == nil {
			xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "data_invalid")
			log.Warn("description miss 2 mid(%v) dynid(%v) type(%v) rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Type, dynCtx.Dyn.Rid)
			return ""
		}
		var ctrl []*mdlv2.Ctrl
		if common.Vest.Ctrl != "[]" {
			if err := json.Unmarshal([]byte(common.Vest.Ctrl), &ctrl); err != nil {
				xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "data_invalid")
				log.Warn("description miss 3 mid(%v) dynid(%v) type(%v) rid(%v) vest.Ctrl %v unmarshal err %v", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Type, dynCtx.Dyn.Rid, common.Vest.Ctrl, err)
			}
		}
		if dynCtx.Dyn.Extend == nil {
			dynCtx.Dyn.Extend = &mdlv2.Extend{}
		}
		dynCtx.Dyn.Extend.Ctrl = ctrl
		desc = common.Vest.Content
	}
	if dynCtx.Dyn.IsApplet() {
		desc = s.descriptionApplet(dynCtx, general)
	}
	if dynCtx.Dyn.IsSubscription() {
		desc = s.descriptionSub(dynCtx, general)
	}
	if dynCtx.Dyn.IsLiveRcmd() {
		desc = s.descriptionLiveRcmd(dynCtx, general)
	}
	if dynCtx.Dyn.IsSubscriptionNew() {
		desc = s.descriptionSubNew(dynCtx, general)
	}
	if dynCtx.Dyn.IsCourUp() && (dynCtx.From != _handleTypeForward && dynCtx.From != _handleTypeShare) {
		desc = s.descriptionCourUp(dynCtx, general)
	}
	if dynCtx.Dyn.IsNewTopicSet() {
		desc = s.descriptionNewTopicSet(dynCtx, general)
	}
	return desc
}

// 描述信息
func (s *Service) description(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	descArr := s.descProc(c, dynCtx.Interim.Desc, dynCtx, general)
	if len(descArr) == 0 {
		return nil
	}
	moduleDesc := &api.Module_ModuleDesc{
		ModuleDesc: &api.ModuleDesc{
			Desc:    descArr,
			Text:    dynCtx.Interim.Desc,
			JumpUri: dynCtx.Interim.PromoURI, // 帮推
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_desc,
		ModuleItem: moduleDesc,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	// 拓展字段内容
	if dynCtx.Dyn.IsForward() {
		var descTmp []*api.Description
		descTmp = append(descTmp, &api.Description{
			Type: api.DescType_desc_type_text,
			Text: "//",
		})
		descTmp = append(descTmp, &api.Description{
			Text: fmt.Sprintf("@%s", dynCtx.DynamicItem.Extend.OrigName),
			Type: api.DescType_desc_type_aite,
			Uri:  model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(dynCtx.DynamicItem.Extend.Uid, 10), nil),
			Rid:  strconv.FormatInt(dynCtx.DynamicItem.Extend.Uid, 10),
		})
		descTmp = append(descTmp, &api.Description{
			Type: api.DescType_desc_type_text,
			Text: ":",
		})
		dynCtx.DynamicItem.Extend.Desc = descTmp
		dynCtx.DynamicItem.Extend.Desc = append(dynCtx.DynamicItem.Extend.Desc, descArr...)
	}
	return nil
}

func (s *Service) descriptionForward(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) string {
	if dynCtx.ResWords != nil && dynCtx.ResWords[dynCtx.Dyn.Rid] != "" {
		return dynCtx.ResWords[dynCtx.Dyn.Rid]
	}
	xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "date_faild")
	return ""
}

func (s *Service) descriptionAV(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) string {
	// 分享卡原卡不展示视频信息
	if dynCtx.From == _handleTypeShare {
		return ""
	}
	if ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid); ok {
		var archive = ap.Arc
		return archive.Dynamic
	}
	xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "date_faild")
	return ""
}

func (s *Service) descriptionWord(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) string {
	if dynCtx.ResWords != nil && dynCtx.ResWords[dynCtx.Dyn.Rid] != "" {
		return dynCtx.ResWords[dynCtx.Dyn.Rid]
	}
	xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "date_faild")
	return ""
}

func (s *Service) descriptionDraw(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) string {
	if draw, ok := dynCtx.GetResDraw(dynCtx.Dyn.Rid); ok {
		return draw.Item.Description
	}
	xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "date_faild")
	return ""
}

func (s *Service) descriptionArticle(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) string {
	if article, ok := dynCtx.GetResArticle(dynCtx.Dyn.Rid); ok {
		return article.Dynamic
	}
	xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "date_faild")
	return ""
}

func (s *Service) descriptionMusic(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) string {
	if music, ok := dynCtx.GetResMusic(dynCtx.Dyn.Rid); ok {
		return music.Intro
	}
	xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "date_faild")
	return ""
}

func (s *Service) descriptionApplet(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) string {
	if applet, ok := dynCtx.GetResApple(dynCtx.Dyn.Rid); ok {
		return applet.Content
	}
	xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "date_faild")
	return ""
}

func (s *Service) descriptionSub(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) string {
	if sub, ok := dynCtx.GetResSub(dynCtx.Dyn.Rid); ok {
		return sub.Desc
	}
	xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "date_faild")
	return ""
}

func (s *Service) descriptionLiveRcmd(_ *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) string {
	return ""
}

func (s *Service) descriptionSubNew(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) string {
	if subNew, ok := dynCtx.GetResSubNew(dynCtx.Dyn.Rid); ok {
		if subNew.Type == submdl.TunnelTypeDraw {
			var subNewDraw *submdl.Subscription
			if err := json.Unmarshal([]byte(subNew.ImageInfo), &subNewDraw); err != nil || subNewDraw == nil {
				xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "desc", "date_faild")
				log.Warn("module error mid(%v) dynid(%v) descriptionSubNew %v", general.Mid, dynCtx.Dyn.DynamicID, subNew.ImageInfo)
				return ""
			}
			return subNewDraw.Desc
		}
	}
	return ""
}

func (s *Service) descriptionCourUp(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) string {
	season, ok := dynCtx.GetResCheeseSeason(dynCtx.Dyn.Rid)
	if ok {
		// 预约召回的使用预约专属文案
		if dynCtx.Dyn.Property.GetRcmdType() == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_CHEESE && len(season.DynamicReserveContent) > 0 {
			return season.DynamicReserveContent
		}
		return season.DynamicShareContent
	}
	return ""
}

// 话题集订阅更新卡文案是根据push id唯一确定的内容
func (s *Service) descriptionNewTopicSet(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) string {
	if dynCtx.Interim.IsPassCard {
		return ""
	}
	tps := dynCtx.GetResNewTopicSet()
	if tps == nil {
		dynCtx.Interim.IsPassCard = true
		return ""
	}
	return tps.TopicList.GetContent()
}

/*
	字段拆分逻辑
*/

func (s *Service) descProc(c context.Context, desc string, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) []*api.Description {
	if desc == "" {
		return []*api.Description{}
	}
	// 客户端所带的高亮信息，@、抽奖、投票、商品
	descArr := s.descCtrl(c, desc, dynCtx, general)
	// 邮箱
	descArr = s.descMail(descArr, dynCtx)
	// 网页地址
	descArr = s.descWeb(descArr, dynCtx)
	// bvid
	descArr = s.descBV(descArr, dynCtx)
	// avid
	descArr = s.descAV(descArr, dynCtx)
	// 专栏
	descArr = s.descCV(descArr, dynCtx)
	// 小视频
	descArr = s.descVC(descArr, dynCtx)
	// emoji表情
	descArr = s.descEmoji(descArr, dynCtx)
	// 话题信息
	descArr = s.descTopic(c, descArr, dynCtx, general)
	// 搜索词
	if dynCtx.SearchWordRed {
		descArr = s.descSearchWord(descArr, dynCtx)
	}
	return descArr
}

func (s *Service) descProcCommunity(_ context.Context, desc string, dynCtx *mdlv2.DynamicContext) []*api.Description {
	descArr := []*api.Description{
		{
			Text: desc,
			Type: api.DescType_desc_type_text,
		},
	}
	// emoji表情
	descArr = s.descEmoji(descArr, dynCtx)
	return descArr
}

// nolint:gocognit
func (s *Service) descCtrl(_ context.Context, desc string, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) []*api.Description {
	if dynCtx.Dyn.Extend == nil || len(dynCtx.Dyn.Extend.Ctrl) == 0 {
		rsp := &api.Description{
			Text: desc,
			Type: api.DescType_desc_type_text,
		}
		return []*api.Description{rsp}
	}
	// ctrl 排序，根据location位置升序排列
	ctrls := mdlv2.CtrlSort{}
	for _, ct := range dynCtx.Dyn.Extend.Ctrl {
		if ct == nil {
			continue
		}
		ctrls = append(ctrls, ct)
	}
	sort.Sort(ctrls)
	// 循环拆分desc
	locStart := 0
	descR := utf16.Encode([]rune(desc))
	var rsp []*api.Description
	for _, ct := range ctrls {
		// *拆前置部分
		lengthEnd := ct.Location
		// 判断是否越界
		if len(descR) < lengthEnd {
			lengthEnd = len(descR)
		}
		if locStart > lengthEnd {
			log.Warn("descCtrl waring 1 desc %v : locStart %v, ct.Location %v, lengthEnd %v", desc, locStart, ct.Location, lengthEnd)
			return rsp
		}
		ru := descR[locStart:lengthEnd]
		if len(ru) != 0 {
			tmp := &api.Description{
				Text: string(utf16.Decode(ru)),
				Type: api.DescType_desc_type_text,
			}
			rsp = append(rsp, tmp)
			locStart = lengthEnd
		}
		// 达到最大长度 中断 返回
		if len(descR) == lengthEnd {
			return rsp
		}
		// *拆核心部分
		var hightLengh int
		hightLengh, _ = strconv.Atoi(ct.Data)
		if ct.TranType() == api.DescType_desc_type_aite {
			hightLengh = ct.Length
		}
		lengthEnd = lengthEnd + hightLengh
		if len(descR) < lengthEnd {
			lengthEnd = len(descR)
		}
		if locStart > lengthEnd {
			log.Warn("descCtrl waring 2 desc %v : locStart %v, ct.Length %v, lengthEnd %v", desc, locStart, ct.Length, lengthEnd)
			continue
		}
		ru = descR[locStart:lengthEnd]
		if len(ru) != 0 {
			tmp := &api.Description{
				Text: string(utf16.Decode(ru)),
				Type: ct.TranType(),
			}
			//nolint:exhaustive
			switch tmp.Type {
			case api.DescType_desc_type_aite:
				tmp.Uri = model.FillURI(model.GotoSpaceDyn, ct.Data, nil)
				tmp.Rid = ct.Data
			case api.DescType_desc_type_lottery:
				if dynCtx.Dyn.Extend != nil && dynCtx.Dyn.Extend.Lott != nil {
					lott := dynCtx.Dyn.Extend.Lott
					tmp.Uri = fmt.Sprintf(model.LottURI, dynCtx.Dyn.Rid, dynCtx.Dyn.Type, lott.LotteryID, dynCtx.Dyn.DynamicID)
					tmp.Rid = strconv.FormatInt(lott.LotteryID, 10)
					if lott.LotteryID == 0 {
						log.Error("dynamic(%v) LotteryID is empty mid(%v), dynid(%v), desc(%v), Lott(%+v)", dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID, desc, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Extend.Lott)
					}
				}
			case api.DescType_desc_type_vote:
				if dynCtx.Dyn.Extend != nil && dynCtx.Dyn.Extend.Vote != nil {
					vote := dynCtx.Dyn.Extend.Vote
					tmp.Uri = fmt.Sprintf(model.VoteURI, vote.VoteID, dynCtx.Dyn.DynamicID)
					tmp.Rid = strconv.FormatInt(vote.VoteID, 10)
					if vote.VoteID == 0 {
						log.Error("dynamic(%v) VoteID is empty mid(%v), dynid(%v), desc(%v), Vote(%+v)", dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID, desc, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Extend.Vote)
					}
				}
			case api.DescType_desc_type_goods:
				if dynCtx.ResGood != nil && dynCtx.ResGood[dynCtx.Dyn.DynamicID] != nil {
					if goods := dynCtx.ResGood[dynCtx.Dyn.DynamicID][bcgmdl.GoodsLocTypeCard]; goods != nil {
						var goodID string
						if ids := strings.Split(ct.TypeID, "_"); len(ids) > 1 {
							goodID = ids[1]
						}
						if goodsItem, ok := goods[goodID]; ok {
							tmp.Uri = goodsItem.JumpLink
							tmp.Rid = strconv.FormatInt(goodsItem.ItemsID, 10)
							tmp.Goods = &api.ModuleDescGoods{
								SourceType:        int32(goodsItem.SourceType),
								JumpUrl:           goodsItem.JumpLink,
								ItemId:            goodsItem.ItemsID,
								SchemaUrl:         goodsItem.SchemaURL,
								OpenWhiteList:     goodsItem.OpenWhiteList,
								UserWebV2:         goodsItem.UserAdWebV2,
								AdMark:            goodsItem.AdMark,
								SchemaPackageName: goodsItem.SchemaPackageName,
								JumpType:          api.GoodsJumpType_goods_schema,
								AppName:           goodsItem.AppName,
							}
							switch goodsItem.SourceType {
							case goodsTypeTaoBao:
								tmp.IconName = "ic_prefix_tb.png"
							case goodsTypeJD:
								tmp.IconName = "ic_prefix_jd.png"
							}
							tmp.IconUrl = goodsItem.IconURL
							if (general.IsIPhonePick() && general.GetBuild() < 66000000 || general.IsAndroidPick() && general.GetBuild() < 6600000) && (goodsItem.SourceType != 1 && goodsItem.SourceType != 2) {
								continue
							}
							// SourceType 1-淘宝、2-会员购  3-京东
							if goodsItem.OuterApp == 0 {
								tmp.Goods.JumpType = api.GoodsJumpType_goods_url
							}
						}
					}
				}
				if tmp.Uri == "" {
					log.Error("dynamic(%v) goods uri is empty mid(%v), dynid(%v), desc(%v), ctrl(%+v)", dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID, desc, dynCtx.Dyn.DynamicID, ct)
				}
			}
			rsp = append(rsp, tmp)
			locStart = lengthEnd
		}
	}
	ru := descR[locStart:]
	if len(ru) != 0 {
		tmp := &api.Description{
			Text: string(utf16.Decode(ru)),
			Type: api.DescType_desc_type_text,
		}
		rsp = append(rsp, tmp)
	}
	return rsp
}

func (s *Service) descMail(descArr []*api.Description, dynCtx *mdlv2.DynamicContext) []*api.Description {
	var rsp []*api.Description
	for _, desc := range descArr {
		if desc.Type != api.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := s.descMailProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

// nolint:unparam
func (s *Service) descMailProc(desc string, dynCtx *mdlv2.DynamicContext) []*api.Description {
	fIndex := _mailRgx.FindStringIndex(desc)
	var res []*api.Description
	if len(fIndex) == 0 {
		tmp := &api.Description{
			Text: desc,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &api.Description{
			Text: pre,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	tmp := &api.Description{
		Text: top,
		Type: api.DescType_desc_type_mail,
		Uri:  top,
	}
	res = append(res, tmp)
	if aft != "" {
		tmp := s.descMailProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func (s *Service) descWeb(descArr []*api.Description, dynCtx *mdlv2.DynamicContext) []*api.Description {
	var rsp []*api.Description
	for _, desc := range descArr {
		if desc.Type != api.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := s.descWebProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func (s *Service) descWebProc(desc string, dynCtx *mdlv2.DynamicContext) []*api.Description {
	fIndex := _webRgx.FindStringIndex(desc)
	var res []*api.Description
	if len(fIndex) == 0 {
		tmp := &api.Description{
			Text: desc,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	// 去掉top末尾双斜杠问题
	if top[len(top)-2:] == "//" {
		topTmp := top[:len(top)-2]
		aft = top[len(top)-2:] + aft
		top = topTmp
	}
	if pre != "" {
		tmp := &api.Description{
			Text: pre,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	tmp := &api.Description{
		Text: "网页链接",
		Type: api.DescType_desc_type_web,
		Uri:  top,
	}
	res = append(res, tmp)
	if top != "" {
		if dynCtx.BackfillDescURL == nil {
			dynCtx.BackfillDescURL = make(map[string]*mdlv2.BackfillDescURLItem)
		}
		dynCtx.BackfillDescURL[top] = nil
	}
	if aft != "" {
		tmp := s.descWebProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func (s *Service) descAV(descArr []*api.Description, dynCtx *mdlv2.DynamicContext) []*api.Description {
	var rsp []*api.Description
	for _, desc := range descArr {
		if desc.Type != api.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := s.descAVProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func (s *Service) descAVProc(desc string, dynCtx *mdlv2.DynamicContext) []*api.Description {
	fIndex := _avRgx.FindStringIndex(desc)
	var res []*api.Description
	if len(fIndex) == 0 {
		tmp := &api.Description{
			Text: desc,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &api.Description{
			Text: pre,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	aid := strings.Replace(top, "av", "", -1)
	tmp := &api.Description{
		Text: top,
		Rid:  aid,
		Type: api.DescType_desc_type_av,
		Uri:  model.FillURI(model.GotoAv, aid, nil),
	}
	res = append(res, tmp)
	if dynCtx.BackfillAvID == nil {
		dynCtx.BackfillAvID = make(map[string]struct{})
	}
	dynCtx.BackfillAvID[aid] = struct{}{}
	if aft != "" {
		tmp := s.descAVProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func (s *Service) descBV(descArr []*api.Description, dynCtx *mdlv2.DynamicContext) []*api.Description {
	var rsp []*api.Description
	for _, desc := range descArr {
		if desc.Type != api.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := s.descBVProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func (s *Service) descBVProc(desc string, dynCtx *mdlv2.DynamicContext) []*api.Description {
	fIndex := _bvRgx.FindStringIndex(desc)
	var res []*api.Description
	if len(fIndex) == 0 {
		tmp := &api.Description{
			Text: desc,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &api.Description{
			Text: pre,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	tmp := &api.Description{
		Text: top,
		Rid:  top,
		Type: api.DescType_desc_type_bv,
		Uri:  model.FillURI(model.GotoAv, top, nil),
	}
	res = append(res, tmp)
	if dynCtx.BackfillBvID == nil {
		dynCtx.BackfillBvID = make(map[string]struct{})
	}
	dynCtx.BackfillBvID[top] = struct{}{}
	if aft != "" {
		tmp := s.descBVProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func (s *Service) descCV(descArr []*api.Description, dynCtx *mdlv2.DynamicContext) []*api.Description {
	var rsp []*api.Description
	for _, desc := range descArr {
		if desc.Type != api.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := s.descCVProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func (s *Service) descCVProc(desc string, dynCtx *mdlv2.DynamicContext) []*api.Description {
	fIndex := _cvRgx.FindStringIndex(desc)
	var res []*api.Description
	if len(fIndex) == 0 {
		tmp := &api.Description{
			Text: desc,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &api.Description{
			Text: pre,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	articleID := strings.Replace(top, "cv", "", -1)
	tmp := &api.Description{
		Text: top,
		Rid:  articleID,
		Type: api.DescType_desc_type_cv,
		Uri:  model.FillURI(model.GotoArticle, articleID, nil),
	}
	res = append(res, tmp)
	if dynCtx.BackfillCvID == nil {
		dynCtx.BackfillCvID = make(map[string]struct{})
	}
	dynCtx.BackfillCvID[articleID] = struct{}{}
	if aft != "" {
		tmp := s.descCVProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func (s *Service) descVC(descArr []*api.Description, dynCtx *mdlv2.DynamicContext) []*api.Description {
	var rsp []*api.Description
	for _, desc := range descArr {
		if desc.Type != api.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := s.descVCProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

// nolint:unparam
func (s *Service) descVCProc(desc string, dynCtx *mdlv2.DynamicContext) []*api.Description {
	fIndex := _vcRgx.FindStringIndex(desc)
	var res []*api.Description
	if len(fIndex) == 0 {
		tmp := &api.Description{
			Text: desc,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &api.Description{
			Text: pre,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	clipID := strings.Replace(top, "vc", "", -1)
	tmp := &api.Description{
		Text: top,
		Rid:  clipID,
		Type: api.DescType_desc_type_vc,
		Uri:  model.FillURI(model.GotoClip, clipID, nil),
	}
	res = append(res, tmp)
	if aft != "" {
		tmp := s.descVCProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func (s *Service) descEmoji(descArr []*api.Description, dynCtx *mdlv2.DynamicContext) []*api.Description {
	var rsp []*api.Description
	for _, desc := range descArr {
		if desc.Type != api.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		var emojiType int
		if dynCtx.Dyn.Extend != nil {
			emojiType = dynCtx.Dyn.Extend.EmojiType
		}
		tmp := s.descEmojiProc(desc.Text, emojiType, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func (s *Service) descEmojiProc(desc string, emojiType int, dynCtx *mdlv2.DynamicContext) []*api.Description {
	fIndex := _emojiRgx.FindStringIndex(desc)
	var res []*api.Description
	if len(fIndex) == 0 {
		tmp := &api.Description{
			Text: desc,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &api.Description{
			Text: pre,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	tmp := &api.Description{
		Text: top,
		Type: api.DescType_desc_type_emoji,
		Uri:  "",
	}
	if dynCtx.Emoji == nil {
		dynCtx.Emoji = make(map[string]struct{})
	}
	dynCtx.Emoji[top] = struct{}{}
	if emojiType == 0 {
		tmp.EmojiType = api.EmojiType_emoji_old
	} else {
		tmp.EmojiType = api.EmojiType_emoji_new
	}
	res = append(res, tmp)
	if aft != "" {
		tmp := s.descEmojiProc(aft, emojiType, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func (s *Service) descTopic(c context.Context, descArr []*api.Description, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) []*api.Description {
	var rsp []*api.Description
	for _, v := range descArr {
		if v.Type != api.DescType_desc_type_text {
			rsp = append(rsp, v)
			continue
		}
		tmp := s.descTopicProc(c, v.Text, dynCtx, general)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func (s *Service) descTopicProc(c context.Context, desc string, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) []*api.Description {
	fIndex := _topicRgx.FindStringIndex(desc)
	var res []*api.Description
	if len(fIndex) == 0 {
		tmp := &api.Description{
			Text: desc,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &api.Description{
			Text: pre,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	var (
		topid     int64
		uri       string
		topicName = strings.Replace(top, "#", "", -1)
	)
	if s.isDynNewTopicView(c, general) {
		// Pad下发全站搜索跳链
		if general.IsPadHD() || general.IsPad() {
			uri = fmt.Sprintf("bilibili://search/?keyword=%s", model.QueryEscape(topicName))
		} else {
			// 兜底下发粉板垂搜跳链
			// 搜索词跳转依然带双#号
			uri = fmt.Sprintf("bilibili://following/dynamic_search?query=%s", model.QueryEscape(top))
		}
		res = append(res, &api.Description{
			Text: top,
			Type: api.DescType_desc_type_topic,
			Uri:  uri,
		})
	} else {
		// 老版本下发对应老话题跳链
		// 兜底跳转旧话题详情页
		uri = fmt.Sprintf("bilibili://tag/0/?name=%v&type=topic", model.QueryEscape(topicName))
		topicInfos, _ := dynCtx.Dyn.GetTopicInfo()
		for _, topic := range topicInfos {
			if topic != nil && topic.TopicName == topicName {
				topid = topic.TopicID
				uri = topic.TopicLink
				break
			}
		}
		tmp := &api.Description{
			Text: top,
			Rid:  strconv.FormatInt(topid, 10),
			Type: api.DescType_desc_type_topic,
			Uri:  uri,
		}
		res = append(res, tmp)
	}
	if aft != "" {
		tmp := s.descTopicProc(c, aft, dynCtx, general)
		res = append(res, tmp...)
	}
	return res
}

func (s *Service) descSearchWord(descArr []*api.Description, dynCtx *mdlv2.DynamicContext) []*api.Description {
	var rsp []*api.Description
	for _, desc := range descArr {
		if desc.Type != api.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := s.descSearchWordProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func (s *Service) descSearchWordProc(desc string, dynCtx *mdlv2.DynamicContext) []*api.Description {
	index := -1
	wordLen := 0
	for _, searchWord := range dynCtx.SearchWords {
		index = strings.Index(desc, searchWord)
		if index != -1 {
			wordLen = len(searchWord)
			break
		}
	}
	var res []*api.Description
	if index == -1 {
		tmp := &api.Description{
			Text: desc,
			Type: api.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	end := index + wordLen
	pre := desc[:index]
	top := desc[index:end]
	aft := desc[end:]
	if pre != "" {
		tmp := s.descSearchWordProc(pre, dynCtx)
		res = append(res, tmp...)
	}
	tmp := &api.Description{
		Text: top,
		Type: api.DescType_desc_type_search_word,
	}
	res = append(res, tmp)
	if aft != "" {
		tmp := s.descSearchWordProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}
