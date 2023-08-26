package topiccardschema

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"

	"go-common/library/log"

	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
)

func getDesc(dynCtx *dynmdlV2.DynamicContext) string {
	switch {
	case dynCtx.Dyn.IsForward():
		return descriptionForward(dynCtx)
	case dynCtx.Dyn.IsAv():
		return descriptionAV(dynCtx)
	case dynCtx.Dyn.IsDraw():
		return descriptionDraw(dynCtx)
	case dynCtx.Dyn.IsWord():
		return descriptionWord(dynCtx)
	case dynCtx.Dyn.IsArticle():
		return descriptionArticle(dynCtx)
	case dynCtx.Dyn.IsCommon():
		common, ok := dynCtx.GetResCommon(dynCtx.Dyn.Rid)
		if !ok {
			return ""
		}
		if common.Vest == nil {
			return ""
		}
		var ctrl []*dynmdlV2.Ctrl
		if common.Vest.Ctrl != "[]" {
			if err := json.Unmarshal([]byte(common.Vest.Ctrl), &ctrl); err != nil {
				log.Warn("common.Vest.Ctrl Unmarshal dynCtx.Dyn=%+v, common.Vest.Ctrl=%+v, err=%+v", dynCtx.Dyn, common.Vest.Ctrl, err)
			}
		}
		dynCtx.Dyn.Extend.Ctrl = ctrl
		return common.Vest.Content
	default:
		return ""
	}
}

func descriptionArticle(dynCtx *dynmdlV2.DynamicContext) string {
	if article, ok := dynCtx.GetResArticle(dynCtx.Dyn.Rid); ok {
		return article.Dynamic
	}
	return ""
}

func descriptionWord(dynCtx *dynmdlV2.DynamicContext) string {
	if dynCtx.ResWords != nil && dynCtx.ResWords[dynCtx.Dyn.Rid] != "" {
		return dynCtx.ResWords[dynCtx.Dyn.Rid]
	}
	return ""
}

func descriptionForward(dynCtx *dynmdlV2.DynamicContext) string {
	if dynCtx.ResWords != nil && dynCtx.ResWords[dynCtx.Dyn.Rid] != "" {
		return dynCtx.ResWords[dynCtx.Dyn.Rid]
	}
	return ""
}

func descriptionAV(dynCtx *dynmdlV2.DynamicContext) string {
	if ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid); ok {
		var archive = ap.Arc
		return archive.Dynamic
	}
	return ""
}

func descriptionDraw(dynCtx *dynmdlV2.DynamicContext) string {
	if draw, ok := dynCtx.GetResDraw(dynCtx.Dyn.Rid); ok {
		return draw.Item.Description
	}
	return ""
}

/*
	字段拆分逻辑
*/

func DescProc(dynCtx *dynmdlV2.DynamicContext, desc string, general *topiccardmodel.GeneralParam) []*dynamicapi.Description {
	if general.Source != "" {
		return nil
	}
	if desc == "" {
		return []*dynamicapi.Description{}
	}
	// 客户端所带的高亮信息，@、抽奖、投票、商品
	descArr := descCtrl(desc, dynCtx)
	// 邮箱
	descArr = descMail(descArr, dynCtx)
	// 网页地址
	descArr = descWeb(descArr, dynCtx)
	// bvid
	descArr = descBV(descArr, dynCtx)
	// avid
	descArr = descAV(descArr, dynCtx)
	// 专栏
	descArr = descCV(descArr, dynCtx)
	// 小视频
	descArr = descVC(descArr, dynCtx)
	// emoji表情
	descArr = descEmoji(descArr, dynCtx)
	// 话题信息
	descArr = descTopic(descArr, dynCtx, general)
	// 搜索词
	if dynCtx.SearchWordRed {
		descArr = descSearchWord(descArr, dynCtx)
	}
	return descArr
}

func descCtrl(desc string, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	if dynCtx.Dyn.Extend == nil || len(dynCtx.Dyn.Extend.Ctrl) == 0 {
		rsp := &dynamicapi.Description{
			Text: desc,
			Type: dynamicapi.DescType_desc_type_text,
		}
		return []*dynamicapi.Description{rsp}
	}
	// ctrl 排序，根据location位置升序排列
	ctrls := dynmdlV2.CtrlSort{}
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
	var rsp []*dynamicapi.Description
	for _, ct := range ctrls {
		// *拆前置部分
		lengthEnd := ct.Location
		// 判断是否越界
		if len(descR) < lengthEnd {
			lengthEnd = len(descR)
		}
		if locStart > lengthEnd {
			log.Warn("descCtrl warning 1 desc %v : locStart %v, ct.Location %v, lengthEnd %v", desc, locStart, ct.Location, lengthEnd)
			return rsp
		}
		ru := descR[locStart:lengthEnd]
		if len(ru) != 0 {
			tmp := &dynamicapi.Description{
				Text: string(utf16.Decode(ru)),
				Type: dynamicapi.DescType_desc_type_text,
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
		if ct.TranType() == dynamicapi.DescType_desc_type_aite {
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
			tmp := &dynamicapi.Description{
				Text: string(utf16.Decode(ru)),
				Type: ct.TranType(),
			}
			switch tmp.Type {
			case dynamicapi.DescType_desc_type_aite:
				tmp.Uri = topiccardmodel.FillURI(topiccardmodel.GotoSpaceDyn, ct.Data, nil)
				tmp.Rid = ct.Data
			case dynamicapi.DescType_desc_type_lottery:
				if dynCtx.Dyn.Extend != nil && dynCtx.Dyn.Extend.Lott != nil {
					lott := dynCtx.Dyn.Extend.Lott
					tmp.Uri = fmt.Sprintf(topiccardmodel.LottURI, dynCtx.Dyn.Rid, dynCtx.Dyn.Type, lott.LotteryID, dynCtx.Dyn.DynamicID)
					tmp.Rid = strconv.FormatInt(lott.LotteryID, 10)
					if lott.LotteryID == 0 {
						log.Error("dynamic(%v) LotteryID is empty mid(%v), dynid(%v), desc(%v), Lott(%+v)", dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID, desc, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Extend.Lott)
					}
				}
			case dynamicapi.DescType_desc_type_vote:
				if dynCtx.Dyn.Extend != nil && dynCtx.Dyn.Extend.Vote != nil {
					vote := dynCtx.Dyn.Extend.Vote
					tmp.Uri = fmt.Sprintf(topiccardmodel.VoteURI, vote.VoteID, dynCtx.Dyn.DynamicID)
					tmp.Rid = strconv.FormatInt(vote.VoteID, 10)
					if vote.VoteID == 0 {
						log.Error("dynamic(%v) VoteID is empty mid(%v), dynid(%v), desc(%v), Vote(%+v)", dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID, desc, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Extend.Vote)
					}
				}
			case dynamicapi.DescType_desc_type_goods:
				if dynCtx.ResGood != nil && dynCtx.ResGood[dynCtx.Dyn.DynamicID] != nil {
					if goods := dynCtx.ResGood[dynCtx.Dyn.DynamicID][topiccardmodel.GoodsLocTypeCard]; goods != nil {
						var goodID string
						if ids := strings.Split(ct.TypeID, "_"); len(ids) > 1 {
							goodID = ids[1]
						}
						if goodsItem, ok := goods[goodID]; ok {
							tmp.Uri = goodsItem.JumpLink
							tmp.Rid = strconv.FormatInt(goodsItem.ItemsID, 10)
							tmp.Goods = &dynamicapi.ModuleDescGoods{
								SourceType:        int32(goodsItem.SourceType),
								JumpUrl:           goodsItem.JumpLink,
								ItemId:            goodsItem.ItemsID,
								SchemaUrl:         goodsItem.SchemaURL,
								OpenWhiteList:     goodsItem.OpenWhiteList,
								UserWebV2:         goodsItem.UserAdWebV2,
								AdMark:            goodsItem.AdMark,
								SchemaPackageName: goodsItem.SchemaPackageName,
							}
							switch goodsItem.SourceType {
							case topiccardmodel.GoodsTypeTaoBao:
								tmp.IconName = "ic_prefix_tb.png"
							}
							tmp.IconUrl = goodsItem.IconURL
						}
					}
				}
			}
			rsp = append(rsp, tmp)
			locStart = lengthEnd
		}
	}
	ru := descR[locStart:]
	if len(ru) != 0 {
		tmp := &dynamicapi.Description{
			Text: string(utf16.Decode(ru)),
			Type: dynamicapi.DescType_desc_type_text,
		}
		rsp = append(rsp, tmp)
	}
	return rsp
}

func descMail(descArr []*dynamicapi.Description, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	var rsp []*dynamicapi.Description
	for _, desc := range descArr {
		if desc.Type != dynamicapi.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descMailProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

// nolint:unparam
func descMailProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	r := regexp.MustCompile(topiccardmodel.MailRex)
	fIndex := r.FindStringIndex(desc)
	var res []*dynamicapi.Description
	if len(fIndex) == 0 {
		tmp := &dynamicapi.Description{
			Text: desc,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &dynamicapi.Description{
			Text: pre,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	tmp := &dynamicapi.Description{
		Text: top,
		Type: dynamicapi.DescType_desc_type_mail,
		Uri:  top,
	}
	res = append(res, tmp)
	if aft != "" {
		tmp := descMailProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func descWeb(descArr []*dynamicapi.Description, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	var rsp []*dynamicapi.Description
	for _, desc := range descArr {
		if desc.Type != dynamicapi.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descWebProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descWebProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	r := regexp.MustCompile(topiccardmodel.WebRex)
	fIndex := r.FindStringIndex(desc)
	var res []*dynamicapi.Description
	if len(fIndex) == 0 {
		tmp := &dynamicapi.Description{
			Text: desc,
			Type: dynamicapi.DescType_desc_type_text,
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
		tmp := &dynamicapi.Description{
			Text: pre,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	tmp := &dynamicapi.Description{
		Text: "网页链接",
		Type: dynamicapi.DescType_desc_type_web,
		Uri:  top,
	}
	res = append(res, tmp)
	if top != "" {
		if dynCtx.BackfillDescURL == nil {
			dynCtx.BackfillDescURL = make(map[string]*dynmdlV2.BackfillDescURLItem)
		}
		dynCtx.BackfillDescURL[top] = nil
	}
	if aft != "" {
		tmp := descWebProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func descAV(descArr []*dynamicapi.Description, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	var rsp []*dynamicapi.Description
	for _, desc := range descArr {
		if desc.Type != dynamicapi.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descAVProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descAVProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	r := regexp.MustCompile(topiccardmodel.AvRex)
	fIndex := r.FindStringIndex(desc)
	var res []*dynamicapi.Description
	if len(fIndex) == 0 {
		tmp := &dynamicapi.Description{
			Text: desc,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &dynamicapi.Description{
			Text: pre,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	aid := strings.Replace(top, "av", "", -1)
	tmp := &dynamicapi.Description{
		Text: top,
		Rid:  aid,
		Type: dynamicapi.DescType_desc_type_av,
		Uri:  topiccardmodel.FillURI(topiccardmodel.GotoAv, aid, nil),
	}
	res = append(res, tmp)
	if dynCtx.BackfillAvID == nil {
		dynCtx.BackfillAvID = make(map[string]struct{})
	}
	dynCtx.BackfillAvID[aid] = struct{}{}
	if aft != "" {
		tmp := descAVProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func descBV(descArr []*dynamicapi.Description, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	var rsp []*dynamicapi.Description
	for _, desc := range descArr {
		if desc.Type != dynamicapi.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descBVProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descBVProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	r := regexp.MustCompile(topiccardmodel.BvRex)
	fIndex := r.FindStringIndex(desc)
	var res []*dynamicapi.Description
	if len(fIndex) == 0 {
		tmp := &dynamicapi.Description{
			Text: desc,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &dynamicapi.Description{
			Text: pre,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	tmp := &dynamicapi.Description{
		Text: top,
		Rid:  top,
		Type: dynamicapi.DescType_desc_type_bv,
		Uri:  topiccardmodel.FillURI(topiccardmodel.GotoAv, top, nil),
	}
	res = append(res, tmp)
	if dynCtx.BackfillBvID == nil {
		dynCtx.BackfillBvID = make(map[string]struct{})
	}
	dynCtx.BackfillBvID[top] = struct{}{}
	if aft != "" {
		tmp := descBVProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func descCV(descArr []*dynamicapi.Description, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	var rsp []*dynamicapi.Description
	for _, desc := range descArr {
		if desc.Type != dynamicapi.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descCVProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descCVProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	fIndex := regexp.MustCompile(topiccardmodel.CvRex).FindStringIndex(desc)
	var res []*dynamicapi.Description
	if len(fIndex) == 0 {
		res = append(res, &dynamicapi.Description{
			Text: desc,
			Type: dynamicapi.DescType_desc_type_text,
		})
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		res = append(res, &dynamicapi.Description{
			Text: pre,
			Type: dynamicapi.DescType_desc_type_text,
		})
	}
	articleID := strings.Replace(top, "cv", "", -1)
	res = append(res, &dynamicapi.Description{
		Text: top,
		Rid:  articleID,
		Type: dynamicapi.DescType_desc_type_cv,
		Uri:  topiccardmodel.FillURI(topiccardmodel.GotoArticle, articleID, nil),
	})
	if dynCtx.BackfillCvID == nil {
		dynCtx.BackfillCvID = make(map[string]struct{})
	}
	dynCtx.BackfillCvID[articleID] = struct{}{}
	if aft != "" {
		tmp := descCVProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func descVC(descArr []*dynamicapi.Description, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	var rsp []*dynamicapi.Description
	for _, desc := range descArr {
		if desc.Type != dynamicapi.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descVCProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

// nolint:unparam
func descVCProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	fIndex := regexp.MustCompile(topiccardmodel.VcRex).FindStringIndex(desc)
	var res []*dynamicapi.Description
	if len(fIndex) == 0 {
		tmp := &dynamicapi.Description{
			Text: desc,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &dynamicapi.Description{
			Text: pre,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	clipID := strings.Replace(top, "vc", "", -1)
	tmp := &dynamicapi.Description{
		Text: top,
		Rid:  clipID,
		Type: dynamicapi.DescType_desc_type_vc,
		Uri:  topiccardmodel.FillURI(topiccardmodel.GotoClip, clipID, nil),
	}
	res = append(res, tmp)
	if aft != "" {
		tmp := descVCProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func descEmoji(descArr []*dynamicapi.Description, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	var rsp []*dynamicapi.Description
	for _, desc := range descArr {
		if desc.Type != dynamicapi.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		var emojiType int
		if dynCtx.Dyn.Extend != nil {
			emojiType = dynCtx.Dyn.Extend.EmojiType
		}
		tmp := descEmojiProc(desc.Text, emojiType, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descEmojiProc(desc string, emojiType int, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	fIndex := regexp.MustCompile(topiccardmodel.EmojiRex).FindStringIndex(desc)
	var res []*dynamicapi.Description
	if len(fIndex) == 0 {
		tmp := &dynamicapi.Description{
			Text: desc,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &dynamicapi.Description{
			Text: pre,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	tmp := &dynamicapi.Description{
		Text: top,
		Type: dynamicapi.DescType_desc_type_emoji,
	}
	if dynCtx.Emoji == nil {
		dynCtx.Emoji = make(map[string]struct{})
	}
	dynCtx.Emoji[top] = struct{}{}
	if emojiType == 0 {
		tmp.EmojiType = dynamicapi.EmojiType_emoji_old
	} else {
		tmp.EmojiType = dynamicapi.EmojiType_emoji_new
	}
	res = append(res, tmp)
	if aft != "" {
		tmp := descEmojiProc(aft, emojiType, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func descTopic(descArr []*dynamicapi.Description, dynCtx *dynmdlV2.DynamicContext, general *topiccardmodel.GeneralParam) []*dynamicapi.Description {
	var rsp []*dynamicapi.Description
	for _, v := range descArr {
		if v.Type != dynamicapi.DescType_desc_type_text {
			rsp = append(rsp, v)
			continue
		}
		tmp := descTopicProc(v.Text, dynCtx, general)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

// nolint:unparam
func descTopicProc(desc string, dynCtx *dynmdlV2.DynamicContext, general *topiccardmodel.GeneralParam) []*dynamicapi.Description {
	r := regexp.MustCompile(topiccardmodel.TopicRex)
	fIndex := r.FindStringIndex(desc)
	var res []*dynamicapi.Description
	if len(fIndex) == 0 {
		tmp := &dynamicapi.Description{
			Text: desc,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &dynamicapi.Description{
			Text: pre,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
	}
	res = append(res, &dynamicapi.Description{
		Text: top,
		Type: dynamicapi.DescType_desc_type_topic,
		Uri:  makeProtoCardDescSearchUri(top, general),
	})
	if aft != "" {
		tmp := descTopicProc(aft, dynCtx, general)
		res = append(res, tmp...)
	}
	return res
}

func makeProtoCardDescSearchUri(top string, general *topiccardmodel.GeneralParam) string {
	if general.IsPadHD() || general.IsPad() {
		return fmt.Sprintf("bilibili://search/?keyword=%s", url.QueryEscape(strings.Replace(top, "#", "", -1)))
	}
	return fmt.Sprintf("bilibili://following/dynamic_search?query=%s", url.QueryEscape(top))
}

func descSearchWord(descArr []*dynamicapi.Description, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	var rsp []*dynamicapi.Description
	for _, desc := range descArr {
		if desc.Type != dynamicapi.DescType_desc_type_text {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descSearchWordProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descSearchWordProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	index := -1
	wordLen := 0
	for _, searchWord := range dynCtx.SearchWords {
		index = strings.Index(desc, searchWord)
		if index != -1 {
			wordLen = len(searchWord)
			break
		}
	}
	var res []*dynamicapi.Description
	if index == -1 {
		tmp := &dynamicapi.Description{
			Text: desc,
			Type: dynamicapi.DescType_desc_type_text,
		}
		res = append(res, tmp)
		return res
	}
	end := index + wordLen
	pre := desc[:index]
	top := desc[index:end]
	aft := desc[end:]
	if pre != "" {
		tmp := descSearchWordProc(pre, dynCtx)
		res = append(res, tmp...)
	}
	tmp := &dynamicapi.Description{
		Text: top,
		Type: dynamicapi.DescType_desc_type_search_word,
	}
	res = append(res, tmp)
	if aft != "" {
		tmp := descSearchWordProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

// 控制外露评论
func DescProcCommunity(desc string, dynCtx *dynmdlV2.DynamicContext) []*dynamicapi.Description {
	descArr := []*dynamicapi.Description{
		{
			Text: desc,
			Type: dynamicapi.DescType_desc_type_text,
		},
	}
	// emoji表情
	descArr = descEmoji(descArr, dynCtx)
	return descArr
}
