package cardbuilder

import (
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
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	"go-gateway/pkg/idsafe/bvid"
)

func descProc(dynCtx *dynmdlV2.DynamicContext, desc string) []*jsonwebcard.RichTextNode {
	if desc == "" {
		return []*jsonwebcard.RichTextNode{}
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
	descArr = descTopic(descArr, dynCtx)
	// 搜索词
	if dynCtx.SearchWordRed {
		descArr = descSearchWord(descArr, dynCtx)
	}
	return descArr
}

// nolint:gocognit
func descCtrl(desc string, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	if dynCtx.Dyn.Extend == nil || len(dynCtx.Dyn.Extend.Ctrl) == 0 {
		return []*jsonwebcard.RichTextNode{
			{
				Text:         desc,
				DescItemType: jsonwebcard.RichTextNodeTypeText,
			},
		}
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
	var rsp []*jsonwebcard.RichTextNode
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
			rsp = append(rsp, &jsonwebcard.RichTextNode{
				Text:         string(utf16.Decode(ru)),
				DescItemType: jsonwebcard.RichTextNodeTypeText,
			})
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
			tmp := &jsonwebcard.RichTextNode{
				Text:         string(utf16.Decode(ru)),
				DescItemType: transferDescItemType(ct),
			}
			switch tmp.DescItemType {
			case jsonwebcard.RichTextNodeTypeAt:
				tmp.JumpUrl = topiccardmodel.FillURI(topiccardmodel.GotoWebSpace, ct.Data, nil)
				tmp.Rid = ct.Data
			case jsonwebcard.RichTextNodeTypeLottery:
				if dynCtx.Dyn.Extend != nil && dynCtx.Dyn.Extend.Lott != nil {
					lott := dynCtx.Dyn.Extend.Lott
					tmp.JumpUrl = fmt.Sprintf(topiccardmodel.LottURI, dynCtx.Dyn.Rid, dynCtx.Dyn.Type, lott.LotteryID, dynCtx.Dyn.DynamicID)
					tmp.Rid = strconv.FormatInt(lott.LotteryID, 10)
					if lott.LotteryID == 0 {
						log.Error("dynamic(%v) LotteryID is empty mid(%v), dynid(%v), desc(%v), Lott(%+v)", dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID, desc, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Extend.Lott)
					}
				}
			case jsonwebcard.RichTextNodeTypeVote:
				if dynCtx.Dyn.Extend != nil && dynCtx.Dyn.Extend.Vote != nil {
					vote := dynCtx.Dyn.Extend.Vote
					tmp.JumpUrl = fmt.Sprintf(topiccardmodel.VoteURI, vote.VoteID, dynCtx.Dyn.DynamicID)
					tmp.Rid = strconv.FormatInt(vote.VoteID, 10)
					if vote.VoteID == 0 {
						log.Error("dynamic(%v) VoteID is empty mid(%v), dynid(%v), desc(%v), Vote(%+v)", dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID, desc, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Extend.Vote)
					}
				}
			case jsonwebcard.RichTextNodeTypeGoods:
				if dynCtx.ResGood != nil && dynCtx.ResGood[dynCtx.Dyn.DynamicID] != nil {
					if goods := dynCtx.ResGood[dynCtx.Dyn.DynamicID][topiccardmodel.GoodsLocTypeCard]; goods != nil {
						var goodID string
						if ids := strings.Split(ct.TypeID, "_"); len(ids) > 1 {
							goodID = ids[1]
						}
						if goodsItem, ok := goods[goodID]; ok {
							tmp.JumpUrl = goodsItem.JumpLink
							tmp.Rid = strconv.FormatInt(goodsItem.ItemsID, 10)
							tmp.RichTextNodeGood = &jsonwebcard.RichTextNodeGood{
								Type:    int64(goodsItem.SourceType),
								JumpUrl: goodsItem.JumpLink,
								Text:    goodsItem.Name,
								IconUrl: goodsItem.IconURL,
							}
							switch goodsItem.SourceType {
							case topiccardmodel.GoodsTypeTaoBao:
								tmp.IconName = "ic_prefix_tb.png"
							}
							tmp.IconUrl = goodsItem.IconURL
						}
					}
				}
			default:
			}
			rsp = append(rsp, tmp)
			locStart = lengthEnd
		}
	}
	ru := descR[locStart:]
	if len(ru) != 0 {
		tmp := &jsonwebcard.RichTextNode{
			Text:         string(utf16.Decode(ru)),
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		}
		rsp = append(rsp, tmp)
	}
	return rsp
}

func transferDescItemType(ct *dynmdlV2.Ctrl) jsonwebcard.RichTextNodeType {
	switch ct.Type {
	case dynmdlV2.CtrlTypeAite:
		return jsonwebcard.RichTextNodeTypeAt
	case dynmdlV2.CtrlTypeLottery:
		return jsonwebcard.RichTextNodeTypeLottery
	case dynmdlV2.CtrlTypeVote:
		return jsonwebcard.RichTextNodeTypeVote
	case dynmdlV2.CtrlTypeGoods:
		return jsonwebcard.RichTextNodeTypeGoods
	default:
		return jsonwebcard.RichTextNodeTypeText
	}
}

func descMail(descArr []*jsonwebcard.RichTextNode, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	var rsp []*jsonwebcard.RichTextNode
	for _, desc := range descArr {
		if desc.DescItemType != jsonwebcard.RichTextNodeTypeText {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descMailProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

// nolint:unparam
func descMailProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	fIndex := regexp.MustCompile(topiccardmodel.MailRex).FindStringIndex(desc)
	var res []*jsonwebcard.RichTextNode
	if len(fIndex) == 0 {
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         desc,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         pre,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
	}
	res = append(res, &jsonwebcard.RichTextNode{
		Text:         top,
		DescItemType: jsonwebcard.RichTextNodeTypeText,
		JumpUrl:      top,
	})
	if aft != "" {
		tmp := descMailProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func descWeb(descArr []*jsonwebcard.RichTextNode, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	var rsp []*jsonwebcard.RichTextNode
	for _, desc := range descArr {
		if desc.DescItemType != jsonwebcard.RichTextNodeTypeText {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descWebProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descWebProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	fIndex := regexp.MustCompile(topiccardmodel.WebRex).FindStringIndex(desc)
	var res []*jsonwebcard.RichTextNode
	if len(fIndex) == 0 {
		tmp := &jsonwebcard.RichTextNode{
			Text:         desc,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
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
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         pre,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
	}
	res = append(res, &jsonwebcard.RichTextNode{
		Text:         "网页链接",
		DescItemType: jsonwebcard.RichTextNodeTypeWeb,
		JumpUrl:      top,
	})
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

func descAV(descArr []*jsonwebcard.RichTextNode, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	var rsp []*jsonwebcard.RichTextNode
	for _, desc := range descArr {
		if desc.DescItemType != jsonwebcard.RichTextNodeTypeText {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descAVProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descAVProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	fIndex := regexp.MustCompile(topiccardmodel.AvRex).FindStringIndex(desc)
	var res []*jsonwebcard.RichTextNode
	if len(fIndex) == 0 {
		tmp := &jsonwebcard.RichTextNode{
			Text:         desc,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		tmp := &jsonwebcard.RichTextNode{
			Text:         pre,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		}
		res = append(res, tmp)
	}
	aid := strings.Replace(top, "av", "", -1)
	res = append(res, &jsonwebcard.RichTextNode{
		Text:         top,
		Rid:          aid,
		DescItemType: jsonwebcard.RichTextNodeTypeAv,
		JumpUrl:      topiccardmodel.FillURI(topiccardmodel.GotoWebAv, aid, nil),
	})
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

func descBV(descArr []*jsonwebcard.RichTextNode, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	var rsp []*jsonwebcard.RichTextNode
	for _, desc := range descArr {
		if desc.DescItemType != jsonwebcard.RichTextNodeTypeText {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descBVProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descBVProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	fIndex := regexp.MustCompile(topiccardmodel.BvRex).FindStringIndex(desc)
	var res []*jsonwebcard.RichTextNode
	if len(fIndex) == 0 {
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         desc,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         pre,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
	}
	res = append(res, &jsonwebcard.RichTextNode{
		Text:         top,
		Rid:          top,
		DescItemType: jsonwebcard.RichTextNodeTypeBv,
		JumpUrl:      topiccardmodel.FillURI(topiccardmodel.GotoWebAv, top, nil),
	})
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

func descCV(descArr []*jsonwebcard.RichTextNode, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	var rsp []*jsonwebcard.RichTextNode
	for _, desc := range descArr {
		if desc.DescItemType != jsonwebcard.RichTextNodeTypeText {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descCVProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descCVProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	fIndex := regexp.MustCompile(topiccardmodel.CvRex).FindStringIndex(desc)
	var res []*jsonwebcard.RichTextNode
	if len(fIndex) == 0 {
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         desc,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         pre,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
	}
	articleID := strings.Replace(top, "cv", "", -1)
	res = append(res, &jsonwebcard.RichTextNode{
		Text:         top,
		Rid:          articleID,
		DescItemType: jsonwebcard.RichTextNodeTypeCv,
		JumpUrl:      topiccardmodel.FillURI(topiccardmodel.GotoArticle, articleID, nil),
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

func descVC(descArr []*jsonwebcard.RichTextNode, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	var rsp []*jsonwebcard.RichTextNode
	for _, desc := range descArr {
		if desc.DescItemType != jsonwebcard.RichTextNodeTypeText {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descVCProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descVCProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	fIndex := regexp.MustCompile(topiccardmodel.VcRex).FindStringIndex(desc)
	var res []*jsonwebcard.RichTextNode
	if len(fIndex) == 0 {
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         desc,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         pre,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
	}
	clipID := strings.Replace(top, "vc", "", -1)
	res = append(res, &jsonwebcard.RichTextNode{
		Text:         top,
		Rid:          clipID,
		DescItemType: jsonwebcard.RichTextNodeTypeVc,
		JumpUrl:      topiccardmodel.FillURI(topiccardmodel.GotoClip, clipID, nil),
	})
	if aft != "" {
		tmp := descVCProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func descEmoji(descArr []*jsonwebcard.RichTextNode, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	var rsp []*jsonwebcard.RichTextNode
	for _, desc := range descArr {
		if desc.DescItemType != jsonwebcard.RichTextNodeTypeText {
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

func descEmojiProc(desc string, emojiType int, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	fIndex := regexp.MustCompile(topiccardmodel.EmojiRex).FindStringIndex(desc)
	var res []*jsonwebcard.RichTextNode
	if len(fIndex) == 0 {
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         desc,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         pre,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
	}
	if dynCtx.Emoji == nil {
		dynCtx.Emoji = make(map[string]struct{})
	}
	dynCtx.Emoji[top] = struct{}{}
	tmp := &jsonwebcard.RichTextNode{
		Text:              top,
		DescItemType:      jsonwebcard.RichTextNodeTypeEmoji,
		RichTextNodeEmoji: &jsonwebcard.RichTextNodeEmoji{Text: top},
	}
	tmp.RichTextNodeEmoji.Type = int64(emojiType)
	res = append(res, tmp)
	if aft != "" {
		tmp := descEmojiProc(aft, emojiType, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func descTopic(descArr []*jsonwebcard.RichTextNode, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	var rsp []*jsonwebcard.RichTextNode
	for _, v := range descArr {
		if v.DescItemType != jsonwebcard.RichTextNodeTypeText {
			rsp = append(rsp, v)
			continue
		}
		tmp := descTopicProc(v.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descTopicProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	fIndex := regexp.MustCompile(topiccardmodel.TopicRex).FindStringIndex(desc)
	var res []*jsonwebcard.RichTextNode
	if len(fIndex) == 0 {
		tmp := &jsonwebcard.RichTextNode{
			Text:         desc,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		}
		res = append(res, tmp)
		return res
	}
	pre, top, aft := desc[:fIndex[0]], desc[fIndex[0]:fIndex[1]], desc[fIndex[1]:]
	if pre != "" {
		res = append(res, &jsonwebcard.RichTextNode{
			Text:         pre,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
		})
	}
	// 搜索词跳转依然带双#号
	res = append(res, &jsonwebcard.RichTextNode{
		Text:         top,
		DescItemType: jsonwebcard.RichTextNodeTypeTopic,
		JumpUrl:      fmt.Sprintf("https://search.bilibili.com/all?keyword=%s", url.QueryEscape(top)),
	})
	if aft != "" {
		tmp := descTopicProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func descSearchWord(descArr []*jsonwebcard.RichTextNode, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	var rsp []*jsonwebcard.RichTextNode
	for _, desc := range descArr {
		if desc.DescItemType != jsonwebcard.RichTextNodeTypeText {
			rsp = append(rsp, desc)
			continue
		}
		tmp := descSearchWordProc(desc.Text, dynCtx)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func descSearchWordProc(desc string, dynCtx *dynmdlV2.DynamicContext) []*jsonwebcard.RichTextNode {
	index := -1
	wordLen := 0
	for _, searchWord := range dynCtx.SearchWords {
		index = strings.Index(desc, searchWord)
		if index != -1 {
			wordLen = len(searchWord)
			break
		}
	}
	var res []*jsonwebcard.RichTextNode
	if index == -1 {
		tmp := &jsonwebcard.RichTextNode{
			Text:         desc,
			DescItemType: jsonwebcard.RichTextNodeTypeText,
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
	tmp := &jsonwebcard.RichTextNode{
		Text:         top,
		DescItemType: jsonwebcard.RichTextNodeTypeSearchWord,
	}
	res = append(res, tmp)
	if aft != "" {
		tmp := descSearchWordProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}

func formDescItem(descItem *jsonwebcard.RichTextNode, dynCtx *dynmdlV2.DynamicContext) {
	descItem.OrigText = descItem.Text
	if descItem.DescItemType == jsonwebcard.RichTextNodeTypeEmoji {
		emoji, ok := dynCtx.ResEmoji[descItem.Text]
		if !ok {
			descItem.DescItemType = jsonwebcard.RichTextNodeTypeText
			return
		}
		descItem.RichTextNodeEmoji.IconUrl = emoji.URL
		descItem.RichTextNodeEmoji.Size = int64(emoji.Meta.Size)
	}
	if descItem.DescItemType == jsonwebcard.RichTextNodeTypeAv {
		if aid, _ := strconv.ParseInt(descItem.Rid, 10, 64); aid != 0 {
			if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
				var archive = ap.Arc
				descItem.OrigText = descItem.Text
				descItem.Text = titleLimit(archive.Title)
				descItem.IconName = "common_video_icon"
			}
		}
	}
	if descItem.DescItemType == jsonwebcard.RichTextNodeTypeBv {
		if aid, _ := bvid.BvToAv(descItem.Rid); aid != 0 {
			if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
				var archive = ap.Arc
				descItem.OrigText = descItem.Text
				descItem.Text = titleLimit(archive.Title)
				descItem.IconName = "common_video_icon"
			}
		}
	}
	if descItem.DescItemType == jsonwebcard.RichTextNodeTypeCv {
		if cvid, _ := strconv.ParseInt(descItem.Rid, 10, 64); cvid != 0 {
			if article, ok := dynCtx.ResBackfillArticle[cvid]; ok {
				descItem.OrigText = descItem.Text
				descItem.Text = titleLimit(article.Title)
				descItem.IconName = "common_article_icon"
			}
		}
	}
	if descItem.DescItemType == jsonwebcard.RichTextNodeTypeWeb {
		if descURL, ok := dynCtx.BackfillDescURL[descItem.JumpUrl]; ok && descURL != nil {
			if descURL.Type == dynamicapi.DescType_desc_type_av {
				if aid, _ := strconv.ParseInt(descURL.Rid, 10, 64); aid != 0 {
					if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
						var archive = ap.Arc
						descItem.OrigText = descItem.Text
						descItem.Text = titleLimit(archive.Title)
						descItem.IconName = "common_video_icon"
					}
				}
			}
			if descURL.Type == dynamicapi.DescType_desc_type_bv {
				if aid, _ := bvid.BvToAv(descURL.Rid); aid != 0 {
					if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
						var archive = ap.Arc
						descItem.OrigText = descItem.Text
						descItem.Text = titleLimit(archive.Title)
						descItem.IconName = "common_video_icon"
					}
				}
			}
			if descURL.Type == dynamicapi.DescType_desc_type_ogv_season {
				if ssid, _ := strconv.ParseInt(descURL.Rid, 10, 64); ssid != 0 {
					if season, ok := dynCtx.ResBackfillSeason[int32(ssid)]; ok {
						descItem.OrigText = descItem.Text
						descItem.Text = titleLimit(season.Title)
						descItem.IconName = "common_video_icon"
					}
				}
			}
			if descURL.Type == dynamicapi.DescType_desc_type_ogv_ep {
				if epid, _ := strconv.ParseInt(descURL.Rid, 10, 64); epid != 0 {
					if episode, ok := dynCtx.ResBackfillEpisode[int32(epid)]; ok && episode.Season != nil {
						descItem.OrigText = descItem.Text
						descItem.Text = titleLimit(episode.Season.Title)
						descItem.IconName = "common_video_icon"
					}
				}
			}
			if descURL.Type == dynamicapi.DescType_desc_type_cv {
				if cvid, _ := strconv.ParseInt(descURL.Rid, 10, 64); cvid != 0 {
					if article, ok := dynCtx.ResBackfillArticle[cvid]; ok {
						descItem.OrigText = descItem.Text
						descItem.Text = titleLimit(article.Title)
						descItem.IconName = "common_article_icon"
					}
				}
			}
		}
	}
	// 兼容逻辑 客户端暂不支持mail类型
	if descItem.DescItemType == jsonwebcard.RichTextNodeTypeMail {
		descItem.DescItemType = jsonwebcard.RichTextNodeTypeText
	}
}

// nolint:gomnd
func titleLimit(title string) string {
	if tmp := []rune(title); len(tmp) > 20 {
		return fmt.Sprintf("%v...", string(tmp[:20]))
	}
	return title
}
