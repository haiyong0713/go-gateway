package dynamicV2

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	dyncomn "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyncampusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
)

func (list *DynListRes) FromAlumniDynamics(dyn *dyncampusgrpc.AlumniDynamicsReply, uid int64) {
	// infoc
	aiItem := &RcmdReply{}
	if len(dyn.AiInfo) > 0 {
		if err := json.Unmarshal([]byte(dyn.AiInfo), aiItem); err != nil {
			log.Error("%+v", err)
		}
	}
	tmp := &RcmdInfo{}
	tmp.FromRcmdInfoDynID(aiItem)
	list.RcmdInfo = tmp
	// dyns
	var logs []string
	list.Toast = dyn.Toast
	list.GuideBar = dyn.GuideBar
	list.CampusFeedUpdate = dyn.Update
	list.CampusHotTopic = &CampusHotTopicInfo{FeedHot: dyn.HotForumCard, FeedHots: dyn.HotForumCards}
	list.YellowBars = dyn.YellowBars
	if dyn.HasMore == 1 {
		list.HasMore = true
	}
	for _, item := range dyn.Dyns {
		if item == nil || item.Type == 0 {
			log.Warn("FromAlumniDynamics miss FromAlumniDynamics mid %v, item %+v", uid, item)
			continue
		}
		if item.Type == 1 && item.Origin == nil {
			log.Warn("FromAlumniDynamics miss forward origin nil mid %v, item %+v", uid, item)
			continue
		}
		logs = append(logs, fmt.Sprintf("dynid(%v) type(%v) rid(%v)", item.DynId, item.Type, item.Rid))
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		// trackid
		if aiInfo, ok := tmp.Listm[item.DynId]; ok {
			dynTmp.TrackID = aiInfo.TrackID
		}
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
	log.Warn("FromAlumniDynamics(new) origin mid(%d) list(%v)", uid, strings.Join(logs, "; "))
}

func (list *DynListRes) FromOfficialDynamics(dyn *dyncampusgrpc.OfficialDynamicsReply) {
	if dyn.HasMore == 1 {
		list.HasMore = true
	}
	list.OffsetInt = int64(dyn.Offset)
	for _, d := range dyn.DynsConfig {
		dynTmp := &Dynamic{}
		dynTmp.DynamicID = int64(d.DynamicId)
		// 只有稿件
		dynTmp.Type = DynTypeVideo
		dynTmp.Rid = int64(d.Rid)
		dynTmp.Desc = d.Reason
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
}

type BillboardItem struct {
	DynBrief *Dynamic
	Reason   string // 上榜原因
}

type CampusBillboardInfo struct {
	Meta  *dyncampusgrpc.BillboardReply // 原始信息
	Items []*BillboardItem
	Dyns  []*Dynamic
}

func (bi *CampusBillboardInfo) FromBillboardReply(meta *dyncampusgrpc.BillboardReply) {
	bi.Meta = meta
	bi.Items = make([]*BillboardItem, 0, len(meta.List))
	bi.Dyns = make([]*Dynamic, 0, len(meta.List))
	for _, oriDyn := range meta.List {
		if oriDyn == nil {
			continue
		}
		dyn := &Dynamic{}
		dyn.FromDynamic(oriDyn.Dyns)
		bi.Dyns = append(bi.Dyns, dyn)
		bi.Items = append(bi.Items, &BillboardItem{DynBrief: dyn, Reason: oriDyn.Reason})
	}
}

type CampusForumSquareInfo struct {
	CampusID    int64
	HasMore     bool
	RcmdTopics  []*dyncampusgrpc.ForumRcmdCard
	PublishLink string // 发布动态的link
}

func (fsi *CampusForumSquareInfo) FromForumSquareReply(campusid int64, meta *dyncampusgrpc.ForumSquareReply) {
	if meta == nil {
		return
	}
	fsi.CampusID = campusid
	fsi.HasMore = meta.HasMore == 1
	fsi.RcmdTopics = meta.List
	fsi.PublishLink = meta.PublishDynamicLink
}

func (fsi *CampusForumSquareInfo) ToV2DynamicItem(c AppDynamicConfig, addPlusMark bool) *api.DynamicItem {
	if fsi == nil {
		return nil
	}
	ret := &api.DynamicItem{
		CardType: api.DynamicType_topic_rcmd,
	}
	ret.Modules = make([]*api.Module, 0, len(fsi.RcmdTopics)+1)
	// 标题
	t, text, icon := c.GetResModuleTitleForCampusTopic()
	title := &api.Module{
		ModuleType: api.DynModuleType_module_title,
		ModuleItem: &api.Module_ModuleTitle{
			ModuleTitle: &api.ModuleTitle{
				Title: t,
			},
		},
	}
	if fsi.HasMore {
		title.ModuleItem.(*api.Module_ModuleTitle).ModuleTitle.RightBtn = &api.IconButton{
			Text: text, IconTail: icon, JumpUri: model.FillURI(model.GotoSchoolTopicList, strconv.FormatInt(fsi.CampusID, 10), nil),
		}
	}
	ret.Modules = append(ret.Modules, title)
	// 填入推荐话题信息
	const _maxRcmds = 3
	topicAdded := 0
	for i, t := range fsi.RcmdTopics {
		if i >= _maxRcmds {
			break
		}
		stat := t.GetHeatInfo()
		desc := model.StatString(stat.GetView(), "浏览") + "·" + model.StatString(stat.GetDiscuss(), "讨论")
		topic := &api.Module{
			ModuleType: api.DynModuleType_module_topic_brief,
			ModuleItem: &api.Module_ModuleTopicBrief{
				ModuleTopicBrief: &api.ModuleTopicBrief{
					Topic: &api.TopicItem{
						TopicId:   t.GetTopicId(),
						TopicName: t.GetTopicName(),
						Url:       t.GetTopicLink(),
						Desc_2:    desc,
					},
				},
			},
		}
		ret.Modules = append(ret.Modules, topic)
		topicAdded++
	}
	if addPlusMark && len(fsi.PublishLink) > 0 {
		if topicAdded >= _maxRcmds {
			// 先铥掉一个
			ret.Modules = ret.Modules[:len(ret.Modules)-1]
		}
		ret.Modules = append(ret.Modules, &api.Module{
			ModuleType: api.DynModuleType_module_button,
			ModuleItem: &api.Module_ModuleButton{
				ModuleButton: &api.ModuleButton{
					Btn: &api.IconButton{
						IconHead: c.GetPlusMarkIcon(),
						Text:     "参与话题讨论",
						JumpUri:  fsi.PublishLink,
					},
				},
			},
		})
	}
	return ret
}

type CampusHotTopicInfo struct {
	FeedHot  *dyncampusgrpc.HotForumCard
	FeedHots []*dyncampusgrpc.HotForumCard
	HomeHead *dyncampusgrpc.HomeHotForumCard
}

func (list *DynListRes) InsertIntoDynList(cards map[int]*api.DynamicItem, bars map[int]*api.DynamicItem, dst []*api.DynamicItem) (res []*api.DynamicItem) {
	if (len(cards) == 0 && len(bars) == 0) || dst == nil {
		return dst
	}
	keys := make([]int, 0, len(cards)+len(bars))
	for pos := range cards {
		keys = append(keys, pos)
	}
	for pos := range bars {
		keys = append(keys, pos)
	}
	// 对cards和bars中存储的key值（即position）进行排序
	sort.Ints(keys)
	keysLen, dstLen := len(keys), len(dst)
	res = make([]*api.DynamicItem, 0, keysLen+dstLen)
	res = append(res, dst...)
	// 所有热议卡的position都是相对于原feed流的，因此按插入位置从大到小处理
	for i := keysLen - 1; i >= 0; {
		pos := keys[i]
		// 若指定位置大于列表长度，则直接将其对应的热议卡放在列表最后
		if pos >= dstLen {
			// 若热议卡与小黄条的插入位置相同，则小黄条在前
			if _, ok := bars[pos]; ok {
				res = append(res, bars[pos])
				i--
			}
			if _, ok := cards[pos]; ok {
				res = append(res, cards[pos])
				i--
			}
		} else { // 否则插入到指定位置
			if _, ok := cards[pos]; ok {
				res = append(res, cards[pos])
				copy(res[pos+1:], res[pos:])
				res[pos] = cards[pos]
				i--
			}
			if _, ok := bars[pos]; ok {
				res = append(res, bars[pos])
				copy(res[pos+1:], res[pos:])
				res[pos] = bars[pos]
				i--
			}
		}
	}
	return
}

func (chti *CampusHotTopicInfo) ToV2CampusHomeRcmdTopic() (ret *api.CampusHomeRcmdTopic) {
	if chti == nil || (chti.FeedHot == nil && chti.HomeHead == nil) {
		return nil
	}
	ret = new(api.CampusHomeRcmdTopic)
	defer func() {
		if len(ret.Topic) <= 0 {
			ret = nil
		}
	}()
	if chti.FeedHot != nil {
		if len(chti.FeedHot.Title) > 0 {
			ret.Title = &api.ModuleTitle{
				Title: chti.FeedHot.Title, TitleStyle: 1,
			}
		}
		for _, m := range chti.FeedHot.List {
			topicItem := &api.TopicItem{
				TopicId:   m.GetTopicId(),
				TopicName: m.GetTopicName(),
				Url:       m.GetTopicLink(),
				Desc_2:    model.StatString(m.GetHeatInfo().GetView(), "浏览") + "·" + model.StatString(m.GetHeatInfo().GetDiscuss(), "讨论"),
			}
			if len(m.GetRcmdDesc()) > 0 {
				topicItem.Desc_2 = m.RcmdDesc
			}
			if len(m.GetButtonText()) > 0 {
				topicItem.Button = &api.IconButton{Text: m.GetButtonText(), JumpUri: m.GetJumpUrl()}
			}
			ret.Topic = append(ret.Topic, topicItem)
		}
	} else if chti.HomeHead != nil {
		if len(chti.HomeHead.Title) > 0 {
			ret.Title = &api.ModuleTitle{
				Title: chti.HomeHead.Title, TitleStyle: 1,
			}
		}
		for _, m := range chti.HomeHead.List {
			topicItem := &api.TopicItem{
				TopicId:   m.GetTopicId(),
				TopicName: m.GetTopicName(),
				Url:       m.GetTopicLink(),
				Desc_2:    model.StatString(m.GetHeatInfo().GetView(), "浏览") + "·" + model.StatString(m.GetHeatInfo().GetDiscuss(), "讨论"),
			}
			if len(m.GetRcmdDesc()) > 0 {
				topicItem.Desc_2 = m.RcmdDesc
			}
			if len(m.GetButtonText()) > 0 {
				topicItem.Button = &api.IconButton{Text: m.GetButtonText(), JumpUri: m.GetJumpUrl()}
			}
			ret.Topic = append(ret.Topic, topicItem)
		}
	}
	return
}

func toV2DynamicItem(hfc *dyncampusgrpc.HotForumCard) (ret *api.DynamicItem) {
	ret = &api.DynamicItem{
		CardType: api.DynamicType_topic_rcmd,
	}
	if len(hfc.Title) > 0 {
		ret.Modules = append(ret.Modules, &api.Module{
			ModuleType: api.DynModuleType_module_title,
			ModuleItem: &api.Module_ModuleTitle{
				ModuleTitle: &api.ModuleTitle{
					Title: hfc.Title, TitleStyle: 1,
				},
			},
		})
	}
	for _, m := range hfc.List {
		topicItem := &api.TopicItem{
			TopicId:   m.GetTopicId(),
			TopicName: m.GetTopicName(),
			Url:       m.GetTopicLink(),
			Desc_2:    model.StatString(m.GetHeatInfo().GetView(), "浏览") + "·" + model.StatString(m.GetHeatInfo().GetDiscuss(), "讨论"),
		}
		if len(m.GetRcmdDesc()) > 0 {
			topicItem.Desc_2 = m.RcmdDesc
		}
		if len(m.GetButtonText()) > 0 {
			topicItem.Button = &api.IconButton{Text: m.GetButtonText(), JumpUri: m.GetJumpUrl()}
		}
		ret.Modules = append(ret.Modules, &api.Module{
			ModuleType: api.DynModuleType_module_topic_brief,
			ModuleItem: &api.Module_ModuleTopicBrief{
				ModuleTopicBrief: &api.ModuleTopicBrief{
					Topic: topicItem,
				},
			},
		})
	}
	return
}

func (chti *CampusHotTopicInfo) ToV2DynamicItem() (rets map[int]*api.DynamicItem) {
	if chti == nil || (chti.FeedHots == nil && chti.FeedHot == nil && chti.HomeHead == nil) {
		return nil
	}
	defer func() {
		for pos, ret := range rets {
			if len(ret.Modules) <= 0 {
				delete(rets, pos)
			}
		}
	}()
	rets = make(map[int]*api.DynamicItem)
	if len(chti.FeedHots) > 0 {
		for _, hfc := range chti.FeedHots {
			rets[int(hfc.GetPosition())] = toV2DynamicItem(hfc)
		}
		return
	}
	if chti.FeedHot != nil {
		rets[int(chti.FeedHot.GetPosition())] = toV2DynamicItem(chti.FeedHot)
	} else if chti.HomeHead != nil {
		ret := &api.DynamicItem{
			CardType: api.DynamicType_topic_rcmd,
		}
		if len(chti.HomeHead.Title) > 0 {
			ret.Modules = append(ret.Modules, &api.Module{
				ModuleType: api.DynModuleType_module_title,
				ModuleItem: &api.Module_ModuleTitle{
					ModuleTitle: &api.ModuleTitle{
						Title: chti.HomeHead.Title, TitleStyle: 1,
					},
				},
			})
		}
		for _, m := range chti.HomeHead.List {
			topicItem := &api.TopicItem{
				TopicId:   m.GetTopicId(),
				TopicName: m.GetTopicName(),
				Url:       m.GetTopicLink(),
				Desc_2:    model.StatString(m.GetHeatInfo().GetView(), "浏览") + "·" + model.StatString(m.GetHeatInfo().GetDiscuss(), "讨论"),
			}
			if len(m.GetRcmdDesc()) > 0 {
				topicItem.Desc_2 = m.RcmdDesc
			}
			if len(m.GetButtonText()) > 0 {
				topicItem.Button = &api.IconButton{Text: m.GetButtonText(), JumpUri: m.GetJumpUrl()}
			}
			ret.Modules = append(ret.Modules, &api.Module{
				ModuleType: api.DynModuleType_module_topic_brief,
				ModuleItem: &api.Module_ModuleTopicBrief{
					ModuleTopicBrief: &api.ModuleTopicBrief{
						Topic: topicItem,
					},
				},
			})
		}
		rets[int(chti.HomeHead.GetPosition())] = ret
	}
	return
}

func getV2DynamicItem(bar *dyncampusgrpc.YellowBar) (ret *api.DynamicItem) {
	ret = &api.DynamicItem{
		CardType: api.DynamicType_notice,
	}
	defer func() {
		if len(ret.Modules) <= 0 {
			ret = nil
		}
	}()
	if len(bar.Title) > 0 {
		ret.Modules = append(ret.Modules, &api.Module{
			ModuleType: api.DynModuleType_module_notice,
			ModuleItem: &api.Module_ModuleNotice{
				ModuleNotice: &api.ModuleNotice{
					Identity:   bar.Identity,
					Icon:       bar.Icon,
					Title:      bar.Title,
					Url:        bar.JumpUrl,
					NoticeType: int32(bar.Type),
				},
			},
		})
	}
	return
}

func (list *DynListRes) GetYellowBarV2DynamicItems() (rets map[int]*api.DynamicItem) {
	if list == nil || list.YellowBars == nil || len(list.YellowBars) == 0 {
		return
	}
	rets = make(map[int]*api.DynamicItem)
	for _, bar := range list.YellowBars {
		if item := getV2DynamicItem(bar); item != nil {
			rets[int(bar.GetPosition())] = item
		}
	}
	return
}

type CampusForumDynamicsInfo struct {
	Dyns        []*Dynamic
	UpdateToast string
	HasMore     bool
	PageOffset  string
}

func (fdi *CampusForumDynamicsInfo) FromForumDynamicsReply(meta *dyncampusgrpc.ForumDynamicsReply) {
	if meta == nil {
		return
	}
	fdi.UpdateToast = meta.Toast
	fdi.HasMore = meta.HasMore == 1
	fdi.PageOffset = meta.Offset
	fdi.Dyns = make([]*Dynamic, 0, len(meta.Dyns))
	for _, dyn := range meta.Dyns {
		d := new(Dynamic)
		d.FromDynamic(dyn)
		fdi.Dyns = append(fdi.Dyns, d)
	}
}

var _campusFromMap = map[api.CampusReqFromType]dyncomn.CampusReqFromType{
	api.CampusReqFromType_DYNAMIC: dyncomn.CampusReqFromType_DYNAMIC,
	api.CampusReqFromType_HOME:    dyncomn.CampusReqFromType_HOME,
}

func ToCampusFromType(from api.CampusReqFromType) dyncomn.CampusReqFromType {
	return _campusFromMap[from]
}

type CampusMngDetailRes struct {
	CampusID   int64
	CampusName string
	Items      []*dyncampusgrpc.CampusMngItem
}

type CampusMngSubmitRes struct {
	Toast string
}

type CampusQuizOperateRes struct {
	List  []*dyncampusgrpc.QuestionItem
	Total int64
}

var (
	quizStatus2AuditStatus = map[dyncomn.QuestionStatus]api.CampusMngAuditStatus{
		dyncomn.QuestionStatus_QUESTION_STATUS_IN_PROCESS: api.CampusMngAuditStatus_campus_mng_audit_in_process,
		dyncomn.QuestionStatus_QUESTION_STATUS_ONLINE:     api.CampusMngAuditStatus_campus_mng_audit_none,
	}
)

func (qop *CampusQuizOperateRes) ToQuizDetailItems() (ret []*api.CampusMngQuizDetail) {
	if qop == nil {
		return nil
	}
	if qop.List != nil {
		ret = make([]*api.CampusMngQuizDetail, 0, len(qop.List))
		for i, q := range qop.List {
			ret = append(ret, &api.CampusMngQuizDetail{
				QuizId:          q.GetId(),
				Question:        q.GetTitle(),
				CorrectAnswer:   q.GetCorrectAnswer(),
				WrongAnswerList: q.GetWrongAnswer(),
				AuditStatus:     quizStatus2AuditStatus[q.GetStatus()],
				AuditMessage:    "已上线",
			})
			if q.Status == dyncomn.QuestionStatus_QUESTION_STATUS_IN_PROCESS {
				ret[i].AuditMessage = "审核中"
			}
		}
	}
	return
}
