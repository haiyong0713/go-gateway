package dynamic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-dynamic/interface/api"
	mdl "go-gateway/app/app-svr/app-dynamic/interface/model"
	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"

	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

const (
	_spaceURI        = "https://space.bilibili.com/%d/dynamic"
	_voteURI         = "https://t.bilibili.com/vote/h5/index/#/result?vote_id=%v&dynamic_id=%v"
	_lbsURI          = "bilibili://following/dynamic_location?poi=%v&poi_type=%v&lat=%v&lng=%v&title=%v&address=%v"
	_topicURI        = "bilibili://pegasus/channel/0/?name=%s&topic_from=card"
	_lottURI         = "https://t.bilibili.com/lottery/h5/index/#/result?business_id=%d&business_type=%d&card=%s"
	_currText        = "等%d个视频"
	_emojiRex        = `[[][^\[\]]+[]]`
	_foldTextUnite   = "展开%d条相关动态"
	_foldTextPublish = "展开%d条相关动态"
	_foldTextLimit   = "%d 条动态被折叠"
)

// baseInfo 基础信息
func (s *Service) baseInfo(_ context.Context, dynCtx *dynmdl.DynContext) error {
	if dynCtx.DynInfo.IsAv() {
		dynCtx.DynamicItem.CardType = "av"
	}
	if dynCtx.DynInfo.IsPGC() {
		dynCtx.DynamicItem.CardType = "pgc"
	}
	if dynCtx.DynInfo.IsCurr() {
		dynCtx.DynamicItem.CardType = "courses"
	}
	dynCtx.DynamicItem.DynIdStr = strconv.FormatInt(dynCtx.DynInfo.DynamicID, 10)
	return nil
}

// author 发布人信息
func (s *Service) author(c context.Context, dynCtx *dynmdl.DynContext) error {
	var (
		module *api.Module
		err    error
	)
	if dynCtx.DynInfo.IsAv() {
		module, err = s.authorByMid(c, dynCtx)
		if err != nil {
			log.Errorc(c, "authorByMid failed(dynamic_id:%v). error(%v)", dynCtx.DynInfo.DynamicID, err)
			return nil
		}
	}
	if dynCtx.DynInfo.IsPGC() {
		module, err = s.authorByPGC(c, dynCtx)
		if err != nil {
			log.Errorc(c, "authorByPGC failed(dynamic_id:%v). error(%v)", dynCtx.DynInfo.DynamicID, err)
			return nil
		}
	}
	if dynCtx.DynInfo.IsCurrSeason() {
		module, err = s.authorBySeason(c, dynCtx)
		if err != nil {
			log.Errorc(c, "authorBySeason failed(dynamic_id:%v). error(%v)", dynCtx.DynInfo.DynamicID, err)
			return nil
		}
	}
	if dynCtx.DynInfo.IsCurrBatch() {
		module, err = s.authorByBatch(c, dynCtx)
		if err != nil {
			log.Errorc(c, "authorByBatch failed(dynamic_id:%v). error(%v)", dynCtx.DynInfo.DynamicID, err)
			return nil
		}
	}
	if module != nil {
		dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	}
	return nil
}

// dispute 争议小黄条
func (s *Service) dispute(_ context.Context, dynCtx *dynmdl.DynContext) error {
	if dynCtx.DynInfo.Extend.Dispute == nil {
		return nil
	}
	dynDisp := dynCtx.DynInfo.Extend.Dispute
	disp := &api.Module_ModuleDispute{
		ModuleDispute: &api.ModuleDispute{
			Title: dynDisp.Content,
			Desc:  dynDisp.Desc,
			Uri:   dynDisp.Url,
		},
	}
	module := &api.Module{
		ModuleType: "dispute",
		ModuleItem: disp,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

// description 正文内容
func (s *Service) description(c context.Context, dynCtx *dynmdl.DynContext) error {
	var desc string
	if dynCtx.DynInfo.IsAv() {
		desc = s.getDescArc(c, dynCtx)
	}
	if dynCtx.DynInfo.IsPGC() || dynCtx.DynInfo.IsCurr() {
		return nil
	}
	if desc == "" {
		return nil
	}
	// 客户端所带的高亮信息，@、抽奖、投票、商品
	descArr := s.descCtrl(desc, dynCtx)
	// emoji表情
	descArr, em := s.descEmoji(descArr, dynCtx)
	dynCtx.Emoji = em
	// 话题信息
	descArr = s.descTopic(descArr, dynCtx)
	moduleDesc := &api.Module_ModuleDesc{
		ModuleDesc: &api.ModuleDesc{
			Desc: descArr,
		},
	}
	module := &api.Module{
		ModuleType: "desc",
		ModuleItem: moduleDesc,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

// dynCard 卡片内容
func (s *Service) dynCard(c context.Context, dynCtx *dynmdl.DynContext) error {
	var (
		module *api.Module
		err    error
	)
	if dynCtx.DynInfo.IsAv() {
		module, err = s.dynCardArc(c, dynCtx)
		if err != nil {
			log.Errorc(c, "dynCardArc() failed(dynamic_id:%v, rid:%v). error(%v)", dynCtx.DynInfo.DynamicID, dynCtx.DynInfo.Rid, err)
			return err
		}
	}
	if dynCtx.DynInfo.IsPGC() {
		module, err = s.dynCardPGC(c, dynCtx)
		if err != nil {
			log.Errorc(c, "dynCardPGC() failed(dynamic_id:%v, rid:%v). error(%v)", dynCtx.DynInfo.DynamicID, dynCtx.DynInfo.Rid, err)
			return err
		}
	}
	if dynCtx.DynInfo.IsCurrBatch() {
		module, err = s.dynCardBatch(c, dynCtx)
		if err != nil {
			log.Errorc(c, "dynCardBatch() failed(dynamic_id:%v, rid:%v). error(%v)", dynCtx.DynInfo.DynamicID, dynCtx.DynInfo.Rid, err)
			return err
		}
	}
	if dynCtx.DynInfo.IsCurrSeason() {
		module, err = s.dynCardSeason(c, dynCtx)
		if err != nil {
			log.Errorc(c, "dynCardSeason() failed(dynamic_id:%v, rid:%v). error(%v)", dynCtx.DynInfo.DynamicID, dynCtx.DynInfo.Rid, err)
			return err
		}
	}
	if module == nil {
		log.Errorc(c, "dynCard() failed(dynamic_id:%v, rid:%v). module is nil", dynCtx.DynInfo.DynamicID, dynCtx.DynInfo.Rid)
		return ecode.NothingFound
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

// state 计数信息
func (s *Service) state(c context.Context, dynCtx *dynmdl.DynContext) error {
	var (
		module *api.Module
		err    error
	)
	if dynCtx.DynInfo.IsAv() {
		module, err = s.stateArc(c, dynCtx)
		if err != nil {
			log.Errorc(c, "stateArc() failed(dynamic_id:%v, rid:%v). error(%v)", dynCtx.DynInfo.DynamicID, dynCtx.DynInfo.Rid, err)
			return err
		}
	}
	if dynCtx.DynInfo.IsPGC() {
		module, err = s.statePGC(c, dynCtx)
		if err != nil {
			log.Errorc(c, "statePGC() failed(dynamic_id:%v, rid:%v). error(%v)", dynCtx.DynInfo.DynamicID, dynCtx.DynInfo.Rid, err)
			return err
		}
	}
	if dynCtx.DynInfo.IsCurrBatch() {
		module, err = s.stateBatch(c, dynCtx)
		if err != nil {
			log.Errorc(c, "stateBatch() failed(dynamic_id:%v, rid:%v). error(%v)", dynCtx.DynInfo.DynamicID, dynCtx.DynInfo.Rid, err)
			return err
		}
	}
	if dynCtx.DynInfo.IsCurrSeason() {
		module, err = s.stateSeason(c, dynCtx)
		if err != nil {
			log.Errorc(c, "stateSeason() failed(dynamic_id:%v, rid:%v). error(%v)", dynCtx.DynInfo.DynamicID, dynCtx.DynInfo.Rid, err)
			return err
		}
	}
	if module == nil {
		log.Errorc(c, "state() failed(dynamic_id:%v, rid:%v). module is nil", dynCtx.DynInfo.DynamicID, dynCtx.DynInfo.Rid)
		return ecode.NothingFound
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

// extend 拓展信息
func (s *Service) extend(c context.Context, dynCtx *dynmdl.DynContext) error {
	extend := &api.ModuleExtend{}
	// 游戏小卡
	if dynCtx.ResBottom != nil && dynCtx.ResBottom[dynCtx.DynInfo.DynamicID] != nil {
		extend.Extend = append(extend.Extend, s.extendBottomCli(dynCtx))
	} else {
		bottoms, err := s.extendBottomCfg(dynCtx)
		if err == nil {
			extend.Extend = append(extend.Extend, bottoms...)
		}
	}
	// lbs
	if dynCtx.DynInfo.Extend.Lbs != nil {
		extend.Extend = append(extend.Extend, s.extendLBS(c, dynCtx))
	}
	// 话题小卡（活动话题）
	hotFlag := true
	if dynCtx.ResTopic != nil && dynCtx.ResTopic[dynCtx.DynInfo.DynamicID] != nil {
		topicExt, err := s.extendTopic(c, dynCtx)
		if err == nil {
			extend.Extend = append(extend.Extend, topicExt)
			hotFlag = false
		}
	}
	// 热门视频（当有话题小卡时则不出）
	if hotFlag && dynCtx.DynInfo.Type == dynmdl.DynTypeVideo && s.resRcmd != nil {
		if _, ok := s.resRcmd[dynCtx.DynInfo.Rid]; ok {
			hotExt := s.extendHot(c, dynCtx)
			extend.Extend = append(extend.Extend, hotExt)
		}
	}
	if len(extend.Extend) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType: "extend",
		ModuleItem: &api.Module_ModuleExtend{
			ModuleExtend: extend,
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

// likeUser 点赞用户
func (s *Service) likeUser(_ context.Context, dynCtx *dynmdl.DynContext) error {
	if dynCtx.DynInfo.Display.LikeUsers == nil || len(dynCtx.DynInfo.Display.LikeUsers) == 0 ||
		dynCtx.ResUid == nil || dynCtx.ResUid.Cards == nil {
		return nil
	}
	var likes []*api.LikeUser
	for _, uid := range dynCtx.DynInfo.Display.LikeUsers {
		res, ok := dynCtx.ResUid.Cards[uid]
		if !ok {
			continue
		}
		likeTmp := &api.LikeUser{
			Uid:   uid,
			Uname: res.Name,
			Uri:   fmt.Sprintf(_spaceURI, uid),
		}
		likes = append(likes, likeTmp)
	}
	like := &api.Module_ModuleLikeUser{
		ModuleLikeUser: &api.ModuleLikeUser{
			LikeUsers:   likes,
			DisplayText: s.c.Resource.LikeDisplay,
		},
	}
	module := &api.Module{
		ModuleType: "likeUser",
		ModuleItem: like,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) authorByMid(_ context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResUid == nil {
		return nil, ecode.NothingFound
	}
	userInfo, ok := dynCtx.ResUid.Cards[dynCtx.DynInfo.UID]
	if !ok {
		return nil, ecode.NothingFound
	}
	official := &api.OfficialVerify{
		Type: int32(userInfo.Official.Type),
		Desc: userInfo.Official.Desc,
	}
	pendant := &api.UserPendant{
		Pid:    int64(userInfo.Pendant.Pid),
		Name:   userInfo.Pendant.Name,
		Image:  userInfo.Pendant.Image,
		Expire: int64(userInfo.Pendant.Expire),
	}
	nameplate := &api.Nameplate{
		Nid:        int64(userInfo.Nameplate.Nid),
		Name:       userInfo.Nameplate.Name,
		Image:      userInfo.Nameplate.Image,
		ImageSmall: userInfo.Nameplate.ImageSmall,
		Level:      userInfo.Nameplate.Level,
		Condition:  userInfo.Nameplate.Condition,
	}
	vip := &api.VipInfo{
		Type:    userInfo.Vip.Type,
		Status:  userInfo.Vip.Status,
		DueDate: userInfo.Vip.DueDate,
		Label: &api.VipLabel{
			Path: userInfo.Vip.Label.Path,
		},
		ThemeType: userInfo.Vip.ThemeType,
	}
	timeLabel := s.timeLabel(dynCtx.DynInfo.Timestamp) + " · " + s.timeLabelType(dynCtx)
	author := &api.UserInfo{
		Mid:       userInfo.Mid,
		Name:      userInfo.Name,
		Face:      userInfo.Face,
		Official:  official,
		Vip:       vip,
		Pendant:   pendant,
		Nameplate: nameplate,
		Uri:       fmt.Sprintf(_spaceURI, userInfo.Mid),
	}
	dynCtx.Interim.Face = userInfo.Face
	dynCtx.Interim.UName = userInfo.Name
	userMdl := &api.Module_ModuleAuthor{
		ModuleAuthor: &api.ModuleAuthor{
			Id:             userInfo.Mid,
			Author:         author,
			PtimeLabelText: timeLabel,
		},
	}
	if dynCtx.ResDecorate != nil {
		decoInfo, ok := dynCtx.ResDecorate[userInfo.Mid]
		if ok {
			decorate := &api.DecorateCard{
				Id:      decoInfo.ID,
				CardUrl: decoInfo.CardURL,
				JumpUrl: decoInfo.JumpURL,
				Fan: &api.DecoCardFan{
					IsFan:  int32(decoInfo.Fan.IsFan),
					Number: int32(decoInfo.Fan.Number),
					Color:  decoInfo.Fan.Color,
				},
			}
			userMdl.ModuleAuthor.DecorateCard = decorate
		}
	}
	module := &api.Module{
		ModuleType: "author",
		ModuleItem: userMdl,
	}
	return module, nil
}

func (s *Service) authorByPGC(_ context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResPGC == nil {
		return nil, ecode.NothingFound
	}
	res, ok := dynCtx.ResPGC[dynCtx.DynInfo.Rid]
	if !ok {
		return nil, ecode.NothingFound
	}
	author := &api.UserInfo{}
	if res.Season != nil {
		author.Name = res.Season.Title
		author.Face = res.Season.Cover
	} else {
		logStr, _ := json.Marshal(res)
		log.Error("authorByPGC season is nil. res: %v", string(logStr))
	}
	timeLabel := s.timeLabel(dynCtx.DynInfo.Timestamp) + " · " + s.timeLabelType(dynCtx)
	userMdl := &api.Module_ModuleAuthor{
		ModuleAuthor: &api.ModuleAuthor{
			Author:         author,
			PtimeLabelText: timeLabel,
		},
	}
	module := &api.Module{
		ModuleType: "author",
		ModuleItem: userMdl,
	}
	return module, nil
}

func (s *Service) authorBySeason(_ context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResPGCSeason == nil {
		return nil, ecode.NothingFound
	}
	res, ok := dynCtx.ResPGCSeason[dynCtx.DynInfo.Rid]
	if !ok {
		return nil, ecode.NothingFound
	}
	author := &api.UserInfo{
		Name: res.UpInfo.Name,
		Face: res.UpInfo.Avatar,
		Uri:  fmt.Sprintf(_spaceURI, res.UpID),
	}
	timeLabel := s.timeLabel(dynCtx.DynInfo.Timestamp) + " · " + s.timeLabelType(dynCtx)
	userMdl := &api.Module_ModuleAuthor{
		ModuleAuthor: &api.ModuleAuthor{
			Id:             res.UpID,
			Author:         author,
			PtimeLabelText: timeLabel,
		},
	}
	module := &api.Module{
		ModuleType: "author",
		ModuleItem: userMdl,
	}
	return module, nil
}

func (s *Service) authorByBatch(_ context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResPGCBatch == nil {
		return nil, ecode.NothingFound
	}
	res, ok := dynCtx.ResPGCBatch[dynCtx.DynInfo.Rid]
	if !ok {
		return nil, ecode.NothingFound
	}
	official := &api.OfficialVerify{
		Desc: res.UserProfile.Card.OfficialVerify.Desc,
		Type: int32(res.UserProfile.Card.OfficialVerify.Type),
	}
	pendant := &api.UserPendant{
		Pid:    res.UserProfile.Pendant.Pid,
		Name:   res.UserProfile.Pendant.Name,
		Image:  res.UserProfile.Pendant.Image,
		Expire: res.UserProfile.Pendant.Expire,
	}
	vip := &api.VipInfo{
		Type:    int32(res.UserProfile.Vip.VipType),
		Status:  int32(res.UserProfile.Vip.VipStatus),
		DueDate: res.UserProfile.Vip.VipDueDate,
		Label: &api.VipLabel{
			Path: res.UserProfile.Vip.Label.Path,
		},
	}
	author := &api.UserInfo{
		Name:     res.UpInfo.Name,
		Face:     res.UpInfo.Avatar,
		Uri:      fmt.Sprintf(_spaceURI, res.UpID),
		Official: official,
		Pendant:  pendant,
		Vip:      vip,
	}
	timeLabel := s.timeLabel(dynCtx.DynInfo.Timestamp) + " · " + s.timeLabelType(dynCtx)
	userMdl := &api.Module_ModuleAuthor{
		ModuleAuthor: &api.ModuleAuthor{
			Id:             res.UpID,
			Author:         author,
			PtimeLabelText: timeLabel,
		},
	}
	module := &api.Module{
		ModuleType: "author",
		ModuleItem: userMdl,
	}
	return module, nil
}

func (s *Service) timeLabel(timestamp int64) string {
	return cardmdl.PubDataString(time.Unix(timestamp, 0))
}

func (s *Service) timeLabelType(dynCtx *dynmdl.DynContext) string {
	if dynCtx.DynInfo.IsAv() {
		if dynCtx.ResArcs != nil && dynCtx.ResArcs[dynCtx.DynInfo.Rid] != nil {
			res, ok := dynCtx.ResArcs[dynCtx.DynInfo.Rid]
			if ok && res.Arc != nil && res.Arc.Rights.IsCooperation == 1 {
				return "与他人联合创作"
			}
		}
		return "投稿了视频"
	}
	if dynCtx.DynInfo.IsPGC() {
		return "更新了"
	}
	if dynCtx.DynInfo.IsCurr() {
		return "更新了视频"
	}
	return ""
}

func (s *Service) getDescArc(_ context.Context, dynCtx *dynmdl.DynContext) string {
	if dynCtx.ResArcs == nil {
		return ""
	}
	res, ok := dynCtx.ResArcs[dynCtx.DynInfo.Rid]
	if !ok || res.Arc == nil {
		return ""
	}
	return res.Arc.Dynamic
}

func (s *Service) descCtrl(desc string, dynCtx *dynmdl.DynContext) []*api.Description {
	if dynCtx.DynInfo.Extend.Ctrl == nil || len(dynCtx.DynInfo.Extend.Ctrl) == 0 {
		rsp := &api.Description{
			Text: desc,
			Type: "text",
		}
		return []*api.Description{rsp}
	}
	// ctrl 排序，根据location位置升序排列
	ctrls := dynmdl.CtrlSort{}
	for _, ct := range dynCtx.DynInfo.Extend.Ctrl {
		if ct == nil {
			continue
		}
		ctrls = append(ctrls, ct)
	}
	sort.Sort(ctrls)
	locPre := 0
	descR := []rune(desc)
	var rsp []*api.Description
	for _, ct := range ctrls {
		// *拆前置部分
		lengthEnd := ct.Location
		// 判断是否越界
		if len(descR) < lengthEnd {
			lengthEnd = len(descR)
		}
		if locPre > lengthEnd {
			log.Warn("descCtrl waring 1 desc %v : locPre %v, ct.Location %v, lengthEnd %v", desc, locPre, ct.Location, lengthEnd)
			return rsp
		}
		ru := descR[locPre:lengthEnd]
		if len(ru) != 0 {
			tmp := &api.Description{
				Text: string(ru),
				Type: "text",
			}
			rsp = append(rsp, tmp)
		}
		length := 0
		switch ct.Type {
		case dynmdl.CtrlTypeAite:
			length = ct.Length
		default:
			length, _ = strconv.Atoi(ct.Data)
		}
		lengthEnd2 := lengthEnd + length
		if lengthEnd > lengthEnd2 || len(descR) < lengthEnd || len(descR) < lengthEnd2 {
			continue
		}
		ru = descR[lengthEnd:lengthEnd2]
		if len(ru) != 0 {
			tmp := &api.Description{
				Text: string(ru),
				Type: ct.TranType(),
			}
			switch tmp.Type {
			case dynmdl.DescTypeLottery:
				lottParams := &dynmdl.LottURIParam{
					Uid:        dynCtx.DynInfo.UID,
					Face:       dynCtx.Interim.Face,
					Name:       dynCtx.Interim.UName,
					CreateTime: dynCtx.DynInfo.Timestamp,
					Content:    desc,
				}
				cardInfo, err := json.Marshal(lottParams)
				if err != nil {
					log.Error("lotteryParams json.Marshal() failed. error(%+v)", err)
					break
				}
				tmp.Uri = fmt.Sprintf(_lottURI, dynCtx.DynInfo.Rid, 8, string(cardInfo))
			case dynmdl.DescTypeVote:
				tmp.Uri = fmt.Sprintf(_voteURI, dynCtx.DynInfo.Extend.Vote.VoteID, dynCtx.DynInfo.DynamicID)
				if dynCtx.DynInfo.Extend.Vote.VoteID == 0 {
					log.Error("voteID is empty. mid:(%v), desc(%v), dynID:(%v), Vote(%+v)", dynCtx.Mid, desc, dynCtx.DynIdStr, dynCtx.DynInfo.Extend.Vote)
				}
			}
			rsp = append(rsp, tmp)
		}
		locPre = lengthEnd + length
	}
	if len(descR) < locPre {
		return rsp
	}
	ru := descR[locPre:]
	if len(ru) != 0 {
		tmp := &api.Description{
			Text: string(ru),
			Type: "text",
		}
		rsp = append(rsp, tmp)
	}
	return rsp
}

func (s *Service) descTopicProc(desc string, topics []*dynmdl.FromContent, index int) []*api.Description {
	if len(topics)-1 < index {
		tmp := &api.Description{
			Text: desc,
			Type: "text",
		}
		return []*api.Description{tmp}
	}
	if topics[index] == nil || topics[index].TopicName == "" {
		return s.descTopicProc(desc, topics, index+1)
	}
	r, err := regexp.Compile("#" + topics[index].TopicName + "#")
	if err != nil {
		log.Error("descProcTopic regexp.Compile(%v) failed. error(%+v)", "#"+topics[index].TopicName+"#", err)
		if (len(topics) - 1) == index {
			tmp := &api.Description{
				Text: desc,
				Type: "text",
			}
			return []*api.Description{tmp}
		}
		return s.descTopicProc(desc, topics, index+1)
	}
	fIndex := r.FindStringIndex(desc)
	var res []*api.Description
	if len(fIndex) == 0 {
		tmp := &api.Description{
			Text: desc,
			Type: "text",
		}
		res = append(res, tmp)
		return res
	}
	pre := desc[:fIndex[0]]
	top := desc[fIndex[0]:fIndex[1]]
	aft := desc[fIndex[1]:]
	if pre != "" {
		if (len(topics) - 1) == index {
			tmp := &api.Description{
				Text: pre,
				Type: "text",
			}
			res = append(res, tmp)
		} else {
			descArr := s.descTopicProc(pre, topics, index+1)
			res = append(res, descArr...)
		}
	}
	tmp := &api.Description{
		Text: top,
		Type: "topic",
		Uri:  topics[index].TopicLink,
	}
	if topics[index].TopicLink != "" {
		tmp.Uri = topics[index].TopicLink
	} else {
		tmp.Uri = fmt.Sprintf(_topicURI, url.QueryEscape(topics[index].TopicName))
	}
	res = append(res, tmp)
	if aft != "" {
		if (len(topics) - 1) == index {
			tmp := &api.Description{
				Text: aft,
				Type: "text",
			}
			res = append(res, tmp)
		} else {
			descArr := s.descTopicProc(aft, topics, index+1)
			res = append(res, descArr...)
		}
	}
	return res
}

func (s *Service) descEmoji(descArr []*api.Description, dynCtx *dynmdl.DynContext) ([]*api.Description, map[string]struct{}) {
	em := make(map[string]struct{})
	var rsp []*api.Description
	for _, desc := range descArr {
		if desc.Type != dynmdl.DescTypeText {
			rsp = append(rsp, desc)
			continue
		}
		tmp := s.descEmojiProc(desc.Text, dynCtx.DynInfo.Extend.EmojiType, em)
		rsp = append(rsp, tmp...)
	}
	return rsp, em
}

func (s *Service) descEmojiProc(desc string, emojiType int, em map[string]struct{}) []*api.Description {
	r := regexp.MustCompile(_emojiRex)
	fIndex := r.FindStringIndex(desc)
	var res []*api.Description
	if len(fIndex) == 0 {
		tmp := &api.Description{
			Text: desc,
			Type: "text",
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
			Type: "text",
		}
		res = append(res, tmp)
	}
	tmp := &api.Description{
		Text: top,
		Type: "emoji",
		Uri:  "",
	}
	em[top] = struct{}{}
	if emojiType == 0 {
		tmp.EmojiType = "old"
	} else {
		tmp.EmojiType = "new"
	}
	res = append(res, tmp)
	if aft != "" {
		tmp := s.descEmojiProc(aft, emojiType, em)
		res = append(res, tmp...)
	}
	return res
}

func (s *Service) descTopic(descArr []*api.Description, dynCtx *dynmdl.DynContext) []*api.Description {
	if dynCtx.ResTopic == nil || dynCtx.ResTopic[dynCtx.DynInfo.DynamicID] == nil {
		return descArr
	}
	topics := dynCtx.ResTopic[dynCtx.DynInfo.DynamicID]
	if topics.FromContent == nil || len(topics.FromContent) == 0 {
		return descArr
	}
	var rsp []*api.Description
	for _, v := range descArr {
		if v.Type != dynmdl.DescTypeText {
			rsp = append(rsp, v)
			continue
		}
		tmp := s.descTopicProc(v.Text, topics.FromContent, 0)
		rsp = append(rsp, tmp...)
	}
	return rsp
}

func (s *Service) dynCardArc(_ context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResArcs == nil {
		return nil, ecode.NothingFound
	}
	arc, ok := dynCtx.ResArcs[dynCtx.DynInfo.Rid]
	if !ok {
		return nil, ecode.NothingFound
	}
	res := arc.Arc
	card := &api.CardUGC{
		Title:           res.Title,
		Cover:           res.Pic,
		CoverLeftText_1: s.videoDuration(res.Duration),
		CoverLeftText_2: fmt.Sprintf("%s观看", s.numTransfer(int(res.Stat.View))),
		CoverLeftText_3: fmt.Sprintf("%s弹幕", s.numTransfer(int(res.Stat.Danmaku))),
		Avid:            res.Aid,
		Cid:             res.FirstCid,
		MediaType:       api.MediaType_MediaTypeUGC,
		Dimension: &api.Dimension{
			Height: res.Dimension.Height,
			Width:  res.Dimension.Width,
			Rotate: res.Dimension.Rotate,
		},
	}
	card.Uri = mdl.FillURI(mdl.GotoAv, strconv.FormatInt(res.Aid, 10), mdl.AvPlayHandlerGRPCV2(arc, res.FirstCid, true))
	if res.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && res.RedirectURL != "" {
		card.Uri = res.RedirectURL
	}
	if res.Rights.IsCooperation == 1 {
		card.Badge = append(card.Badge, dynmdl.CooperationBadge)
	}
	if res.Rights.UGCPay == 1 {
		card.Badge = append(card.Badge, dynmdl.PayBadge)
	}
	card.CanPlay = res.Rights.Autoplay
	dynamic := &api.Module_ModuleDynamic{
		ModuleDynamic: &api.ModuleDynamic{
			CardType: "ugc",
			Card: &api.ModuleDynamic_CardUgc{
				CardUgc: card,
			},
		},
	}
	module := &api.Module{
		ModuleType: "dynamic",
		ModuleItem: dynamic,
	}
	return module, nil
}

func (s *Service) dynCardPGC(_ context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResPGC == nil {
		return nil, ecode.NothingFound
	}
	res, ok := dynCtx.ResPGC[dynCtx.DynInfo.Rid]
	if !ok {
		return nil, ecode.NothingFound
	}
	card := &api.CardPGC{
		Title:           res.NewDesc,
		Cover:           res.Cover,
		Uri:             res.URL,
		CoverLeftText_1: s.videoDuration(res.Duration),
		CoverLeftText_2: fmt.Sprintf("%s观看", s.numTransfer(res.Stat.Play)),
		CoverLeftText_3: fmt.Sprintf("%s弹幕", s.numTransfer(res.Stat.Danmaku)),
		Cid:             res.Cid,
		Epid:            res.EpisodeID,
		Aid:             res.Aid,
		MediaType:       api.MediaType_MediaTypeUGC,
		IsPreview:       int32(res.IsPreview),
		Dimension: &api.Dimension{
			Height: res.Dimension.Height,
			Width:  res.Dimension.Width,
			Rotate: res.Dimension.Rotate,
		},
	}
	if res.Season != nil {
		season := &api.PGCSeason{
			IsFinish: int32(res.Season.IsFinish),
			Title:    res.Season.Title,
			Type:     int32(res.Season.Type),
		}
		card.Season = season
		card.SeasonId = res.Season.SeasonID
	}
	var canPlay int32
	if res.PlayerInfo != nil {
		canPlay = 1
	}
	card.CanPlay = canPlay
	card.SubType = dynCtx.DynInfo.GetPGCSubType()
	dynamic := &api.Module_ModuleDynamic{
		ModuleDynamic: &api.ModuleDynamic{
			CardType: "pgc",
			Card: &api.ModuleDynamic_CardPgc{
				CardPgc: card,
			},
		},
	}
	module := &api.Module{
		ModuleType: "dynamic",
		ModuleItem: dynamic,
	}
	return module, nil
}

func (s *Service) dynCardBatch(_ context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResPGCBatch == nil {
		return nil, ecode.NothingFound
	}
	res, ok := dynCtx.ResPGCBatch[dynCtx.DynInfo.Rid]
	if !ok {
		return nil, ecode.NothingFound
	}
	card := &api.CardCurrBatch{
		Title:  res.NewEp.Title,
		Cover:  res.NewEp.Cover,
		Uri:    res.URL,
		Text_1: res.Title,
	}
	badge := &api.VideoBadge{
		Text:           res.Badge.Text,
		TextColor:      res.Badge.TextColor,
		TextColorNight: res.Badge.TextDarkColor,
		BgColor:        res.Badge.BgColor,
		BgColorNight:   res.Badge.BgDarkColor,
	}
	card.Badge = badge
	if res.UpdateCount > 1 {
		card.Text_2 = fmt.Sprintf(_currText, res.UpdateCount)
	}
	dynamic := &api.Module_ModuleDynamic{
		ModuleDynamic: &api.ModuleDynamic{
			CardType: "currBatch",
			Card: &api.ModuleDynamic_CardCurrBatch{
				CardCurrBatch: card,
			},
		},
	}
	module := &api.Module{
		ModuleType: "dynamic",
		ModuleItem: dynamic,
	}
	return module, nil
}

func (s *Service) dynCardSeason(_ context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResPGCSeason == nil {
		return nil, ecode.NothingFound
	}
	res, ok := dynCtx.ResPGCSeason[dynCtx.DynInfo.Rid]
	if !ok {
		return nil, ecode.NothingFound
	}
	card := &api.CardCurrSeason{
		Title:  res.Title,
		Cover:  res.Cover,
		Desc:   res.Subtitle,
		Uri:    res.URL,
		Text_1: res.Title,
	}
	badge := &api.VideoBadge{
		Text:           res.Badge.Text,
		TextColor:      res.Badge.TextColor,
		TextColorNight: res.Badge.TextDarkColor,
		BgColor:        res.Badge.BgColor,
		BgColorNight:   res.Badge.BgDarkColor,
	}
	card.Badge = badge
	dynamic := &api.Module_ModuleDynamic{
		ModuleDynamic: &api.ModuleDynamic{
			CardType: "currSeason",
			Card: &api.ModuleDynamic_CardCurrSeason{
				CardCurrSeason: card,
			},
		},
	}
	module := &api.Module{
		ModuleType: "dynamic",
		ModuleItem: dynamic,
	}
	return module, nil
}

// nolint:gomnd
func (s *Service) numTransfer(num int) string {
	if num < 10000 {
		return strconv.Itoa(num)
	}
	integer := num / 10000
	decimals := num % 10000
	decimals = decimals / 1000
	return fmt.Sprintf("%d.%d万", integer, decimals)
}

func (s *Service) videoDuration(du int64) string {
	hour := du / dynmdl.PerHour
	du = du % dynmdl.PerHour
	minute := du / dynmdl.PerMinute
	second := du % dynmdl.PerMinute
	if hour != 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
	}
	return fmt.Sprintf("%02d:%02d", minute, second)
}

func (s *Service) stateArc(c context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResArcs == nil {
		return nil, ecode.NothingFound
	}
	arc, ok := dynCtx.ResArcs[dynCtx.DynInfo.Rid]
	if !ok {
		return nil, ecode.NothingFound
	}
	res := arc.Arc
	state := &api.Module_ModuleState{
		ModuleState: &api.ModuleState{
			Repost: int32(dynCtx.DynInfo.Repost),
			Reply:  res.Stat.Reply,
			Like:   res.Stat.Like,
		},
	}
	s.likeIcon(c, dynCtx, state)
	s.attribute(c, dynCtx, state)
	var thum int32
	if dynCtx.ResThumStats != nil && dynCtx.ResThumStats.Business != nil && dynCtx.ResThumStats.Business[dynmdl.BusTypeVideo] != nil {
		for rid, v := range dynCtx.ResThumStats.Business[dynmdl.BusTypeVideo].Records {
			if rid == dynCtx.DynInfo.Rid {
				if v.LikeState == thumgrpc.State_STATE_LIKE {
					thum = 1
				}
				break
			}
		}
	}
	state.ModuleState.LikeInfo.IsLike = thum
	module := &api.Module{
		ModuleItem: state,
		ModuleType: "state",
	}
	return module, nil
}

func (s *Service) statePGC(c context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResPGC == nil {
		return nil, ecode.NothingFound
	}
	res, ok := dynCtx.ResPGC[dynCtx.DynInfo.Rid]
	if !ok {
		return nil, ecode.NothingFound
	}
	state := &api.Module_ModuleState{
		ModuleState: &api.ModuleState{
			Repost: int32(dynCtx.DynInfo.Repost),
			Reply:  int32(res.Stat.Reply),
		},
	}
	s.likeIcon(c, dynCtx, state)
	s.attribute(c, dynCtx, state)
	var thum, thumNum int32
	if dynCtx.ResThumStats != nil && dynCtx.ResThumStats.Business != nil && dynCtx.ResThumStats.Business[dynmdl.BusTypePGC] != nil {
		for rid, v := range dynCtx.ResThumStats.Business[dynmdl.BusTypePGC].Records {
			if rid == dynCtx.DynInfo.Rid {
				if v.LikeState == 1 {
					thum = 1
				}
				thumNum = int32(v.LikeNumber)
				break
			}
		}
	}
	state.ModuleState.LikeInfo.IsLike = thum
	state.ModuleState.Like = thumNum
	module := &api.Module{
		ModuleItem: state,
		ModuleType: "state",
	}
	return module, nil
}

func (s *Service) stateBatch(c context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResPGCBatch == nil {
		return nil, ecode.NothingFound
	}
	res, ok := dynCtx.ResPGCBatch[dynCtx.DynInfo.Rid]
	if !ok {
		return nil, ecode.NothingFound
	}
	state := &api.Module_ModuleState{
		ModuleState: &api.ModuleState{
			Repost: int32(dynCtx.DynInfo.Repost),
			Reply:  int32(res.NewEp.Reply),
		},
	}
	s.likeIcon(c, dynCtx, state)
	s.attribute(c, dynCtx, state)
	var thum, thumNum int32
	if dynCtx.ResThumStats != nil && dynCtx.ResThumStats.Business != nil && dynCtx.ResThumStats.Business[dynmdl.BusTypeCheese] != nil {
		for rid, v := range dynCtx.ResThumStats.Business[dynmdl.BusTypeCheese].Records {
			if rid == dynCtx.DynInfo.Rid {
				if v.LikeState == 1 {
					thum = 1
				}
				thumNum = int32(v.LikeNumber)
				break
			}
		}
	}
	state.ModuleState.LikeInfo.IsLike = thum
	state.ModuleState.Like = thumNum
	module := &api.Module{
		ModuleItem: state,
		ModuleType: "state",
	}
	return module, nil
}

func (s *Service) stateSeason(c context.Context, dynCtx *dynmdl.DynContext) (*api.Module, error) {
	if dynCtx.ResPGCSeason == nil {
		return nil, ecode.NothingFound
	}
	state := &api.Module_ModuleState{
		ModuleState: &api.ModuleState{
			Repost: int32(dynCtx.DynInfo.Repost),
		},
	}
	s.likeIcon(c, dynCtx, state)
	s.attribute(c, dynCtx, state)
	module := &api.Module{
		ModuleItem: state,
		ModuleType: "state",
	}
	return module, nil
}

func (s *Service) likeIcon(_ context.Context, dynCtx *dynmdl.DynContext, module *api.Module_ModuleState) {
	icon := &api.LikeAnimation{}
	if dynCtx.ResLikeIcon != nil && dynCtx.ResLikeIcon[dynCtx.DynInfo.DynamicID] != nil {
		res := dynCtx.ResLikeIcon[dynCtx.DynInfo.DynamicID]
		icon.Begin = res.StartURL
		icon.End = res.EndURL
		icon.Proc = res.ActionURL
		icon.LikeIconId = res.NewIconID
	}
	module.ModuleState.LikeInfo = &api.LikeInfo{
		Animation: icon,
	}
}

func (s *Service) attribute(_ context.Context, dynCtx *dynmdl.DynContext, module *api.Module_ModuleState) {
	if dynCtx.DynInfo.ACL.CommentBan == 1 {
		module.ModuleState.NoComment = true
	}
	if dynCtx.DynInfo.ACL.RepostBan == 1 {
		module.ModuleState.NoForward = true
	}
}

func (s *Service) extendLBS(_ context.Context, dynCtx *dynmdl.DynContext) *api.Extend {
	lbsInfo := dynCtx.DynInfo.Extend.Lbs
	lbs := &api.Extend_ExtInfoLbs{
		ExtInfoLbs: &api.ExtInfoLBS{
			Title:   lbsInfo.Title,
			PoiType: int32(lbsInfo.Type),
			Uri:     fmt.Sprintf(_lbsURI, lbsInfo.Poi, lbsInfo.Type, lbsInfo.Location.Lat, lbsInfo.Location.Lng, lbsInfo.Title, lbsInfo.Address),
			Icon:    s.c.Resource.LbsIcon,
		},
	}
	return &api.Extend{
		Type:   "lbs",
		Extend: lbs,
	}
}

func (s *Service) extendTopic(_ context.Context, dynCtx *dynmdl.DynContext) (*api.Extend, error) {
	topicInfo := dynCtx.ResTopic[dynCtx.DynInfo.DynamicID]
	var topic *api.Extend_ExtInfoTopic
	if topicInfo.TopicActivity != nil {
		topic = &api.Extend_ExtInfoTopic{
			ExtInfoTopic: &api.ExtInfoTopic{
				Title: topicInfo.TopicActivity.TopicName,
				Uri:   topicInfo.TopicActivity.TopicLink,
				Icon:  s.c.Resource.TopicIcon,
			},
		}
	}
	if topic != nil {
		extend := &api.Extend{
			Type:   "topic",
			Extend: topic,
		}
		return extend, nil
	}
	return nil, ecode.NothingFound
}

func (s *Service) extendHot(_ context.Context, _ *dynmdl.DynContext) *api.Extend {
	hot := &api.Extend_ExtInfoHot{
		ExtInfoHot: &api.ExtInfoHot{
			Title: "热门",
			Icon:  s.c.Resource.HotIcon,
			Uri:   s.c.Resource.HotURI,
		},
	}
	extend := &api.Extend{
		Type:   "hot",
		Extend: hot,
	}
	return extend
}

func (s *Service) procEmoji(_ context.Context, dynCtx *dynmdl.DynContext) {
	if dynCtx.ResEmoji != nil {
		return
	}
	res := dynCtx.ResEmoji
	for _, module := range dynCtx.Modules {
		if module.ModuleType == "desc" {
			descs, ok := module.ModuleItem.(*api.Module_ModuleDesc)
			if !ok || descs.ModuleDesc == nil {
				continue
			}
			for _, desc := range descs.ModuleDesc.Desc {
				if desc.Type == "emoji" {
					emojiInfo, ok := res[desc.Text]
					if !ok {
						continue
					}
					desc.Uri = emojiInfo.URL
				}
			}
		}
	}
}

func (s *Service) extendBottomCli(dynCtx *dynmdl.DynContext) *api.Extend {
	bot := dynCtx.ResBottom[dynCtx.DynInfo.DynamicID]
	bottom := &api.Extend_ExtInfoGame{
		ExtInfoGame: &api.ExtInfoGame{
			Title: bot.BottomInfo.Content,
			Uri:   bot.BottomInfo.JumpURL,
			Icon:  s.c.Resource.GameIcon,
		},
	}
	module := &api.Extend{
		Type:   "game",
		Extend: bottom,
	}
	return module
}

func (s *Service) extendBottomCfg(dynCtx *dynmdl.DynContext) ([]*api.Extend, error) {
	if s.bottomMap == nil || dynCtx.ResTopic == nil {
		return nil, ecode.NothingFound
	}
	res, ok := dynCtx.ResTopic[dynCtx.DynInfo.DynamicID]
	if !ok || res.FromContent == nil {
		return nil, ecode.NothingFound
	}

	var moduleMap = make(map[string]*api.Extend_ExtInfoGame)
	for _, topic := range res.FromContent {
		btmRes, ok := s.bottomMap["#"+topic.TopicName+"#"]
		if !ok {
			continue
		}
		bottom := &api.Extend_ExtInfoGame{
			ExtInfoGame: &api.ExtInfoGame{
				Title: btmRes.Display,
				Uri:   btmRes.URL,
				Icon:  s.c.Resource.GameIcon,
			},
		}
		moduleMap[topic.TopicName] = bottom
	}
	var rsp []*api.Extend
	for _, bottom := range moduleMap {
		extend := &api.Extend{
			Type:   "game",
			Extend: bottom,
		}
		rsp = append(rsp, extend)
	}
	return rsp, nil
}

func (s *Service) foldUnite(_ context.Context, list *dynmdl.FoldList, ignore map[string]*dynmdl.FoldItem) (map[string]*dynmdl.FoldItem, error) {
	var (
		group    = make(map[int64][]*dynmdl.FoldItem)
		foldMap  = make(map[string]dynmdl.FoldMapItem)
		copyList []*dynmdl.FoldItem
	)
	// Step 1. 得到按 rid 分组的map，并且copy一个临时 list
	for _, item := range list.List {
		if item.IsAv() && ignore[item.DynIdStr] == nil {
			group[item.Rid] = append(group[item.Rid], item)
		}
		copyList = append(copyList, item)
	}
	// Step 2. 分组后，同一个 rid 如果存在多条时，需要折叠。此步骤得到折叠map
	for _, items := range group {
		num := 1
		orig := new(dynmdl.FoldItem)
		if len(items) > 1 {
			baseID := ""
			// nolint:gomnd
			for _, v := range items {
				fold := dynmdl.FoldMapItem{
					DynItem: v,
				}
				if num == 1 {
					fold.FoldType = dynmdl.FoldMapTypeShow
					foldMap[v.DynIdStr] = fold
					baseID = fold.DynItem.DynIdStr
					num++
					continue
				}
				if num == 2 {
					orig = v
					fold.FoldType = dynmdl.FoldMapTypeFirst
					fold.InsertBase = baseID
				}
				fold.OrigDyn = orig
				foldMap[v.DynIdStr] = fold
			}
		}
	}
	// Step 3. 清空原list，根据折叠map重新赋值，并重构ignore
	list.List = []*dynmdl.FoldItem{}
	for _, item := range copyList {
		res, ok := foldMap[item.DynIdStr]
		if !ok {
			list.List = append(list.List, item)
			continue
		}
		ignore[item.DynIdStr] = item
		if res.FoldType == dynmdl.FoldMapTypeShow {
			item.DynamicItem.HasFold = dynmdl.HasFold
			list.List = append(list.List, item)
			continue
		}
		if res.FoldType == dynmdl.FoldMapTypeFirst {
			list.List = s.foldUniteInsert(list.List, item, res.InsertBase)
			item.FoldList = append(item.FoldList, item.DynIdStr)
			item.FoldType = api.FoldType_FoldTypeUnite
			continue
		}
		res.OrigDyn.FoldList = append(res.OrigDyn.FoldList, item.DynIdStr)
	}
	return ignore, nil
}

func (s *Service) foldUniteInsert(list []*dynmdl.FoldItem, dynItem *dynmdl.FoldItem, baseID string) []*dynmdl.FoldItem {
	var rsp []*dynmdl.FoldItem
	for _, item := range list {
		rsp = append(rsp, item)
		if item.DynIdStr == baseID {
			rsp = append(rsp, dynItem)
		}
	}
	return rsp
}

// nolint:gocognit
func (s *Service) foldPublish(_ context.Context, list *dynmdl.FoldList, ignore map[string]*dynmdl.FoldItem) (map[string]*dynmdl.FoldItem, error) {
	var (
		foldMap    = make(map[string]dynmdl.FoldMapItem)
		copyList   []*dynmdl.FoldItem
		valueTime  int64
		preItem    = &dynmdl.FoldItem{}
		first      = true
		tmpFoldArr []*dynmdl.FoldItem
	)
	// 循环判断相邻动态是否是同一人发布的同一类型，并 copy 一个临时 list
	for _, item := range list.List {
		copyList = append(copyList, item)
		if !item.CanFoldFrequent() {
			goto FOREND
		}
		if s.c.FoldPublishList != nil && s.c.FoldPublishList.White != nil {
			for _, v := range s.c.FoldPublishList.White {
				if v == item.Uid {
					goto FOREND
				}
			}
		}
		if preItem.Type == item.Type && preItem.Uid == item.Uid && ignore[item.DynIdStr] == nil {
			sub := valueTime - item.Timestamp
			if sub < 0 {
				sub = -sub
			}
			if sub < item.GetLimitTime(s.c) {
				if first {
					tmpFoldArr = append(tmpFoldArr, preItem)
					first = false
				}
				tmpFoldArr = append(tmpFoldArr, item)
				valueTime = item.Timestamp
				continue
			}
		}
	FOREND:
		// 当临时数组长度大于等于3时，记录至折叠map中
		// nolint:gomnd
		if len(tmpFoldArr) >= 3 {
			orig := new(dynmdl.FoldItem)
			for k, v := range tmpFoldArr {
				fold := dynmdl.FoldMapItem{
					DynItem: v,
				}
				if k == 0 {
					fold.FoldType = dynmdl.FoldMapTypeShow
				}
				if k == 1 {
					fold.FoldType = dynmdl.FoldMapTypeFirst
					orig = v
				}
				fold.OrigDyn = orig
				foldMap[v.DynIdStr] = fold
			}
		}
		preItem = item
		first = true
		valueTime = item.Timestamp
		tmpFoldArr = []*dynmdl.FoldItem{}
	}
	// Step2. 清空list，根据折叠map重新赋值并重构ignore
	list.List = []*dynmdl.FoldItem{}
	for _, item := range copyList {
		res, ok := foldMap[item.DynIdStr]
		if !ok {
			list.List = append(list.List, item)
			continue
		}
		ignore[item.DynIdStr] = item
		if res.FoldType == dynmdl.FoldMapTypeShow {
			item.DynamicItem.HasFold = dynmdl.HasFold
			list.List = append(list.List, item)
			continue
		}
		if res.FoldType == dynmdl.FoldMapTypeFirst {
			list.List = append(list.List, item)
			item.FoldList = append(item.FoldList, item.DynIdStr)
			item.FoldType = api.FoldType_FoldTypePublish
			continue
		}
		res.OrigDyn.FoldList = append(res.OrigDyn.FoldList, item.DynIdStr)
	}
	return ignore, nil
}

func (s *Service) foldLimit(_ context.Context, list *dynmdl.FoldList, ignore map[string]*dynmdl.FoldItem) (map[string]*dynmdl.FoldItem, error) {
	var (
		foldMap       = make(map[string]dynmdl.FoldMapItem)
		copyList      []*dynmdl.FoldItem
		continuousNum = 0
		orig          = &dynmdl.FoldItem{}
	)
	// Step1. 循环判断动态是否被受限折叠，如果设置了就保存进折叠 map 中。copy 一个临时 list
	for _, item := range list.List {
		copyList = append(copyList, item)
		if item.Acl.FoldLimit != 1 {
			continuousNum = 0
			orig = &dynmdl.FoldItem{}
			continue
		}
		fold := dynmdl.FoldMapItem{
			DynItem: item,
		}
		if continuousNum == 0 {
			fold.FoldType = dynmdl.FoldMapTypeFirst
			orig = item
		}
		fold.OrigDyn = orig
		foldMap[item.DynIdStr] = fold
		continuousNum++
	}
	// Step2. 重构list、ignore
	list.List = []*dynmdl.FoldItem{}
	for _, item := range copyList {
		res, ok := foldMap[item.DynIdStr]
		if !ok {
			list.List = append(list.List, item)
			continue
		}
		if res.FoldType == dynmdl.FoldMapTypeFirst {
			list.List = append(list.List, item)
			item.FoldType = api.FoldType_FoldTypeLimit
		}
		res.OrigDyn.FoldList = append(res.OrigDyn.FoldList, item.DynIdStr)
	}
	return ignore, nil
}

func (s *Service) foldFinish(list *dynmdl.FoldList) []*api.DynamicItem {
	var rsp []*api.DynamicItem
	for _, item := range list.List {
		if len(item.FoldList) == 0 {
			rsp = append(rsp, item.DynamicItem)
			continue
		}
		fold := &api.Module_ModuleFold{
			ModuleFold: &api.ModuleFold{
				FoldTypeV2: item.FoldType,
				Text:       s.getFoldText(item),
				FoldIds:    strings.Join(item.FoldList, ","),
			},
		}
		module := &api.Module{
			ModuleType: "fold",
			ModuleItem: fold,
		}
		item.DynamicItem.CardType = dynmdl.FOLD
		item.DynamicItem.Modules = []*api.Module{module}
		rsp = append(rsp, item.DynamicItem)
	}
	return rsp
}

func (s *Service) getFoldText(item *dynmdl.FoldItem) string {
	if item.FoldType == api.FoldType_FoldTypePublish {
		return fmt.Sprintf(_foldTextPublish, len(item.FoldList))
	}
	if item.FoldType == api.FoldType_FoldTypeUnite {
		return fmt.Sprintf(_foldTextUnite, len(item.FoldList))
	}
	if item.FoldType == api.FoldType_FoldTypeLimit {
		return fmt.Sprintf(_foldTextLimit, len(item.FoldList))
	}
	return ""
}

func (s *Service) FromSVideoAuthor(_ context.Context, svm *dynmdl.SVideoMaterial, _ *dynmdl.Header, req *api.SVideoReq) error {
	if svm == nil || svm.Arc == nil || svm.Arc.Arc == nil {
		return ecode.NothingFound
	}
	pubDesc := PubDataString(svm.Arc.Arc.PubDate.Time())
	if !dynmdl.IsPopularSv(req) {
		pubDesc += " · 发布了动态"
	}
	a := &api.SVideoModuleAuthor{
		Mid:         svm.Arc.Arc.Author.Mid,
		Face:        svm.Arc.Arc.Author.Face,
		Name:        svm.Arc.Arc.Author.Name,
		Uri:         mdl.FillURI(mdl.GotoSpaceDyn, strconv.FormatInt(svm.Arc.Arc.Author.Mid, 10), nil),
		IsAttention: svm.IsAtten,
		PubDesc:     pubDesc,
	}
	module := &api.SVideoModule{
		ModuleType: mdl.ModuleTypeAuthor,
		ModuleItem: &api.SVideoModule_ModuleAuthor{ModuleAuthor: a},
	}
	svm.SVideoItem.Modules = append(svm.SVideoItem.Modules, module)
	return nil
}

// PubDataString is.
func PubDataString(t time.Time) (s string) {
	if t.IsZero() {
		return
	}
	now := time.Now()
	sub := now.Sub(t)
	if sub < time.Minute {
		s = "刚刚"
		return
	}
	if sub < time.Hour {
		s = strconv.FormatFloat(sub.Minutes(), 'f', 0, 64) + "分钟前"
		return
	}
	if sub < 24*time.Hour {
		s = strconv.FormatFloat(sub.Hours(), 'f', 0, 64) + "小时前"
		return
	}
	if now.Year() == t.Year() {
		if now.YearDay()-t.YearDay() == 1 {
			s = "昨天"
			return
		}
		s = t.Format("01-02")
		return
	}
	s = t.Format("2006-01-02")
	return
}

func (s *Service) FromSVideoPlayer(_ context.Context, svm *dynmdl.SVideoMaterial, header *dynmdl.Header, _ *api.SVideoReq) error {
	if svm == nil || svm.Arc == nil || svm.Arc.Arc == nil {
		return ecode.NothingFound
	}
	p := &api.SVideoModulePlayer{
		Title:    svm.Arc.Arc.Title,
		Cover:    svm.Arc.Arc.Pic,
		Uri:      mdl.FillURI(mdl.GotoAv, strconv.FormatInt(svm.Arc.Arc.Aid, 10), cardmdl.ArcPlayHandler(svm.Arc.Arc, cardmdl.ArcPlayURL(svm.Arc, 0), "", nil, header.Build, header.MobiApp, true)),
		Aid:      svm.Arc.Arc.Aid,
		Cid:      svm.Arc.Arc.FirstCid,
		Duration: svm.Arc.Arc.Duration,
		Dimension: &api.Dimension{
			Height: svm.Arc.Arc.Dimension.Height,
			Width:  svm.Arc.Arc.Dimension.Width,
			Rotate: svm.Arc.Arc.Dimension.Rotate,
		},
	}
	module := &api.SVideoModule{
		ModuleType: mdl.ModuleTypePlayer,
		ModuleItem: &api.SVideoModule_ModulePlayer{ModulePlayer: p},
	}
	svm.SVideoItem.Modules = append(svm.SVideoItem.Modules, module)
	return nil
}

func (s *Service) FromSVideoDesc(_ context.Context, svm *dynmdl.SVideoMaterial, header *dynmdl.Header, req *api.SVideoReq) error {
	if svm == nil || svm.Arc == nil || svm.Arc.Arc == nil {
		return ecode.NothingFound
	}
	d := &api.SVideoModuleDesc{
		Uri: mdl.FillURI(mdl.GotoAv, strconv.FormatInt(svm.Arc.Arc.Aid, 10), cardmdl.ArcPlayHandler(svm.Arc.Arc, cardmdl.ArcPlayURL(svm.Arc, 0), "", nil, header.Build, header.MobiApp, true)),
	}
	if !dynmdl.IsPopularSv(req) {
		d.Text = svm.Arc.Arc.Desc
	}
	module := &api.SVideoModule{
		ModuleType: mdl.ModuleTypeDesc,
		ModuleItem: &api.SVideoModule_ModuleDesc{ModuleDesc: d},
	}
	svm.SVideoItem.Modules = append(svm.SVideoItem.Modules, module)
	return nil
}

func (s *Service) FromSVideoStat(_ context.Context, svm *dynmdl.SVideoMaterial, _ *dynmdl.Header, _ *api.SVideoReq) error {
	if svm == nil || svm.Arc == nil || svm.Arc.Arc == nil {
		return ecode.NothingFound
	}
	var err error
	shareInfo := &api.ShareInfo{
		Aid:   svm.Arc.Arc.Aid,
		Title: svm.Arc.Arc.Title,
		Cover: svm.Arc.Arc.Pic,
		Mid:   svm.Arc.Arc.Author.Mid,
		Name:  svm.Arc.Arc.Author.Name,
	}
	if shareInfo.Bvid, err = bvid.AvToBv(svm.Arc.Arc.Aid); err != nil {
		log.Error("avtobv aid:%d err(%v)", svm.Arc.Arc.Aid, err)
	}
	// nolint:gomnd
	if svm.Arc.Arc.Stat.View > 100000 {
		tmp := strconv.FormatFloat(float64(svm.Arc.Arc.Stat.View)/10000, 'f', 1, 64)
		shareInfo.Subtitle = "已观看" + strings.TrimSuffix(tmp, ".0") + "万次"
	}
	var si []*api.SVideoStatInfo
	si = append(si, &api.SVideoStatInfo{Icon: mdl.IconShare, Num: int64(svm.Arc.Arc.Stat.Share)})
	si = append(si, &api.SVideoStatInfo{
		Icon: mdl.IconReply,
		Num:  int64(svm.Arc.Arc.Stat.Reply),
		Uri:  mdl.FillReplyURL(mdl.FillURI(mdl.GotoAv, strconv.FormatInt(svm.Arc.Arc.Aid, 10), mdl.AvPlayHandlerGRPC(svm.Arc.Arc, nil, nil)), "comment_on=1"),
	})
	si = append(si, &api.SVideoStatInfo{Icon: mdl.IconLike, Num: int64(svm.Arc.Arc.Stat.Like), Selected: svm.IsLike})
	st := &api.SVideoModuleStat{ShareInfo: shareInfo, StatInfo: si}
	module := &api.SVideoModule{
		ModuleType: mdl.ModuleTypeStat,
		ModuleItem: &api.SVideoModule_ModuleStat{ModuleStat: st},
	}
	svm.SVideoItem.Modules = append(svm.SVideoItem.Modules, module)
	return nil
}
