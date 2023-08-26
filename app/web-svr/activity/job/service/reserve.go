package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/log"
	"go-common/library/queue/databus"
	go_common_library_time "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/api"
	likeconst "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/app/web-svr/activity/job/model/mail"
	"go-gateway/app/web-svr/activity/job/model/match"
	"go-gateway/app/web-svr/activity/job/model/task"
	"go-gateway/app/web-svr/activity/tools/lib/function"
	"strconv"
	"strings"
	"time"
)

type notifyClockActivity struct {
	*like.ActSubject
	*like.SubjectRule
}

type notifyJobInfo struct {
	*like.ActSubject
	List []*like.ActSubjectNotify
}

var (
	mapNotifyReserveActivity = map[int64]*like.ActSubject{}
	mapNotifyClockActivity   = map[int64]*notifyClockActivity{}
	mapNotifyJobs            = map[int64]*notifyJobInfo{}
)

func (s *Service) ReserveNotifySubProc() {
	defer s.waiter.Done()
	if s.reserveNotifySub == nil {
		return
	}
	ctx := context.Background()
	for {
		msg, ok := <-s.reserveNotifySub.Messages()
		if !ok {
			log.Infoc(ctx, "ReserveNotifySubProc: reserveNotifySub consumer exit!")
			return
		}
		msg.Commit()
		obj := new(like.ProgressNotifyMessage)
		if err := json.Unmarshal(msg.Value, &obj); err != nil {
			log.Errorc(ctx, "ReserveNotifySubProc: json.Unmarshal error[%v]", err)
			continue
		}
		if obj.Mid > 0 {
			// 用户维度数据不关注
			continue
		}
		notifies, ok := mapNotifyJobs[obj.Sid]
		if !ok {
			continue
		}
		for _, notify := range notifies.List {
			if notify.NotifyTime > 0 {
				// 已经通知过了
				continue
			}
			if obj.RuleID == notify.RuleID && obj.Num >= notify.Threshold {
				// 触发通知
				s.occurReserveNotify(ctx, notifies.ActSubject, notify, obj.Num)
			}
		}
	}
}

func (s *Service) buildEmailList(notify *like.ActSubjectNotify) []*mail.Address {
	m := make(map[string]struct{})
	for _, receiver := range s.c.Reserve.Notify {
		m[receiver] = struct{}{}
	}
	for _, receiver := range strings.Split(notify.Receiver, ",") {
		m[receiver] = struct{}{}
	}
	m[notify.Author] = struct{}{}
	list := make([]*mail.Address, 0, len(m))
	for receiver := range m {
		list = append(list, &mail.Address{
			Address: fmt.Sprintf("%s@bilibili.com", receiver),
			Name:    receiver,
		})
	}
	return list
}

func (s *Service) occurReserveNotify(ctx context.Context, subject *like.ActSubject, notify *like.ActSubjectNotify, num int64) error {
	log.Infoc(ctx, "notify appended subject[%v] notify[%v]", *subject, notify)
	data := make(map[string]interface{})
	json.Unmarshal(notify.Ext, &data)
	var name interface{} = subject.Name
	if _, ok := data["page_name"]; ok {
		name = data["page_name"]
	}
	var interveningThreshold interface{} // 干预后的值
	if v, ok := data["intervening_threshold"]; ok {
		interveningThreshold = v
	}
	err := s.SendTextMail(ctx, s.buildEmailList(notify), fmt.Sprintf("【活动平台进度提醒】%v即将到达节点%s", name, notify.Title),
		fmt.Sprintf("你的活动（%v）即将到达设置的节点%v，http://activity-template.bilibili.co/editDynamic/activity/%v",
			name, interveningThreshold, data["page_id"]))
	if err == nil {
		affect, err := s.dao.NotifyMarkFinish(ctx, notify.ID)
		if affect > 0 {
			notify.NotifyTime = time.Now().Unix()
		}
		return err
	}
	return err
}

func (s *Service) LoadNotifySubjectInfoProc() {
	defer s.waiter.Done()
	var ctx = context.Background()
	s.LoadNotifySubjectInfo(ctx)
	for range time.Tick(time.Minute) {
		s.LoadNotifySubjectInfo(ctx)
	}
}

func (s *Service) loadNotifyJobs(ctx context.Context, subs []*like.ActSubject) error {
	tmpNotify := make(map[int64]*notifyJobInfo)
	for _, sub := range subs {
		notifies, err := s.dao.NotifyList(ctx, sub.ID)
		if err != nil {
			log.Errorc(ctx, "loadNotifyJobs: s.dao.NotifyList(%d) err[%v]", sub.ID, err)
			continue
		}
		if len(notifies) > 0 {
			job := new(notifyJobInfo)
			job.ActSubject = sub
			job.List = notifies
			tmpNotify[sub.ID] = job
		}
	}
	mapNotifyJobs = tmpNotify
	return nil
}

func (s *Service) LoadNotifySubjectInfo(ctx context.Context) (data map[string]interface{}, err error) {
	var subs []*like.ActSubject
	now := time.Now()
	// 拉取进行中的预约活动
	if subs, err = s.dao.SubjectList(ctx, []int64{likeconst.RESERVATION, likeconst.CLOCKIN}, now); err != nil {
		log.Error("LoadNotifySubjectInfo s.dao.SubjectList error(%+v)", err)
		return
	}
	// 过滤预约活动，提取打卡活动id
	tmpReserve := make(map[int64]*like.ActSubject)
	clockInID := make([]int64, 0, len(subs))
	clockInMap := make(map[int64]*like.ActSubject)
	for _, sub := range subs {
		if sub.Type == likeconst.RESERVATION {
			tmpReserve[sub.ID] = sub
		} else {
			clockInID = append(clockInID, sub.ID)
			clockInMap[sub.ID] = sub
		}
	}
	mapNotifyReserveActivity = tmpReserve

	// 拉取打卡活动规则信息
	tmpClockIn := make(map[int64]*notifyClockActivity)
	idNum := len(clockInID)
	for i := 0; i < idNum; i += sidBatchNum {
		var idArr []int64
		if i+sidBatchNum <= idNum {
			// 满一批次
			idArr = clockInID[i : i+sidBatchNum]
		} else {
			// 不足一批次
			idArr = clockInID[i:idNum]
		}
		// 拉取活动对应的统计规则
		var rules []*like.SubjectRule
		rules, err = s.dao.SubjectRulesBySids(ctx, idArr)
		if err != nil {
			log.Error("LoadNotifySubjectInfo s.dao.SubjectRulesBySids error(%+v) sid(%v)", err, idArr)
			continue
		}
		// 按照taskid组织活动数据
		for _, rule := range rules {
			tmpClockIn[rule.TaskID] = &notifyClockActivity{
				ActSubject:  clockInMap[rule.Sid],
				SubjectRule: rule,
			}
		}
	}
	mapNotifyClockActivity = tmpClockIn
	s.loadNotifyJobs(ctx, subs)
	return map[string]interface{}{
		"subs":                     subs,
		"mapNotifyReserveActivity": mapNotifyReserveActivity,
		"mapNotifyClockActivity":   mapNotifyClockActivity,
		"mapNotifyJobs":            mapNotifyJobs,
	}, nil
}

func (s *Service) PubClockInRuleStatUpdate(c context.Context, newData json.RawMessage, oldData json.RawMessage) error {
	obj := new(task.State)
	if err := json.Unmarshal(newData, &obj); err != nil {
		log.Errorc(c, "PubClockInRuleStatUpdate json.Unmarshal data[%s] err[%v]", newData, err)
		return err
	}
	if activity, ok := mapNotifyClockActivity[obj.TaskID]; !ok {
		log.Errorc(c, "PubClockInRuleStatUpdate ignore task_id[%d]", obj.TaskID)
		return nil
	} else {
		return s.PubReserveNotifyMessage(c, fmt.Sprint(activity.Sid), &like.ProgressNotifyMessage{
			Sid:     activity.ActSubject.ID,
			Num:     obj.Num,
			RuleID:  activity.SubjectRule.ID,
			Subject: activity.ActSubject,
		})
	}
	return nil
}

func (s *Service) PubBwsReserve(ctx context.Context, newData json.RawMessage, oldData json.RawMessage) (err error) {
	obj := new(api.ActInterReserve)
	if err = json.Unmarshal(newData, &obj); err != nil {
		log.Errorc(ctx, "PubBwsReserve json.Unmarshal data[%s] err[%v]", newData, err)
		return err
	}
	if len(obj.CtimeStr) > 0 {
		ctime, _ := time.ParseInLocation("2006-01-02 15:04:05", obj.CtimeStr, time.Local)
		obj.Ctime = go_common_library_time.Time(ctime.Unix())
	}

	req := &api.GiftStockReq{
		SID:     fmt.Sprintf("%v", api.ActInterReserveTicketType_StandardTicket2021),
		GiftID:  obj.ID,
		GiftVer: obj.Ctime.Time().Unix(),
		GiftNum: obj.StandardTicketNum,
	}
	if req.GiftNum > 0 {
		_, err = s.actGRPC.IncrStockInCache(ctx, req)
		log.Infoc(ctx, "PubBwsReserve Standard Ticket req:%v , reply:%+v", *req, err)
	}

	if obj.VipTicketNum > 0 {
		req.SID = fmt.Sprintf("%v", api.ActInterReserveTicketType_VipTicket2021)
		req.GiftNum = obj.VipTicketNum
		_, err = s.actGRPC.IncrStockInCache(ctx, req)
		log.Infoc(ctx, "PubBwsReserve vip Ticket req:%v , reply:%+v", *req, err)
	}
	return
}

func (s *Service) PubClockInUserUpdate(c context.Context, newData json.RawMessage, oldData json.RawMessage) error {
	// UserState time无法解析，所以自定义一个
	obj := new(struct {
		ID         int64 `json:"id"`
		MID        int64 `json:"mid"`
		BusinessID int64 `json:"business_id"`
		ForeignID  int64 `json:"foreign_id"`
		TaskID     int64 `json:"task_id"`
		Count      int   `json:"cnt"`
		RoundCount int   `json:"round_count"`
	})
	if err := json.Unmarshal(newData, &obj); err != nil {
		log.Errorc(c, "PubClockInUserUpdate json.Unmarshal data[%s] err[%v]", newData, err)
		return err
	}
	if activity, ok := mapNotifyClockActivity[obj.TaskID]; !ok {
		log.Errorc(c, "PubClockInUserUpdate ignore task_id[%d]", obj.TaskID)
		return nil
	} else {
		num := int64(obj.Count)
		// 按天统计case兼容
		if activity.Attribute&1 == 1 {
			num = int64(obj.RoundCount)
		}
		return s.PubReserveNotifyMessage(c, fmt.Sprintf("%d_%d", activity.Sid, obj.MID), &like.ProgressNotifyMessage{
			Sid:     activity.ActSubject.ID,
			Mid:     obj.MID,
			RuleID:  activity.SubjectRule.ID,
			Num:     num,
			Subject: activity.ActSubject,
		})
	}
	return nil
}

func (s *Service) PubReserveUserUpdate(c context.Context, newData *like.ActReserveField) error {
	obj := newData
	if sub, ok := mapNotifyReserveActivity[obj.Sid]; !ok {
		log.Errorc(c, "PubReserveUserUpdate ignore sid[%d]", obj.Sid)
		return nil
	} else {
		// 兼容未预约情况
		if obj.State != 1 {
			obj.Num = 0
		}
		return s.PubReserveNotifyMessage(c, fmt.Sprintf("%d_%d", obj.Sid, obj.Mid), &like.ProgressNotifyMessage{
			Sid:     obj.Sid,
			Mid:     obj.Mid,
			Num:     obj.Num,
			Subject: sub,
		})
	}
}

func (s *Service) PubReserveActivityStatUpdate(c context.Context, newData json.RawMessage, oldData json.RawMessage) error {
	obj := new(like.SubjectStat)
	if err := json.Unmarshal(newData, &obj); err != nil {
		log.Errorc(c, "PubReserveActivityStatUpdate json.Unmarshal data[%s] err[%v]", newData, err)
		return err
	}
	if sub, ok := mapNotifyReserveActivity[obj.Sid]; !ok {
		log.Errorc(c, "PubReserveActivityStatUpdate ignore sid[%d]", obj.Sid)
		return nil
	} else {
		return s.PubReserveNotifyMessage(c, fmt.Sprint(obj.Sid), &like.ProgressNotifyMessage{
			Sid:     obj.Sid,
			Num:     obj.Num,
			Subject: sub,
		})
	}
}

func (s *Service) PubReserveNotifyMessage(c context.Context, key string, message *like.ProgressNotifyMessage) (err error) {
	message.MTime = time.Now()
	for i := 0; i < 3; i++ {
		err = s.reserveNotifyPub.Send(c, key, message)
		if err == nil {
			break
		} else {
			log.Errorc(c, "PubReserveNotifyMessage key[%s] msg[%v] err[%v]", key, *message, err)
		}
	}
	return
}

func (s *Service) PubReserveActivityData(c context.Context, newData *like.ActReserveField) error {
	obj := newData

	// 暂时配置文件中获取数据
	var boot *databus.Databus
	var sendKey string
	bootName := s.c.NewYearReservePubConfig.ConfigName
	switch bootName {
	case "NewYearReserve":
		boot = s.newYearReservePub
		sendKey = fmt.Sprintf("%v_%v", obj.Sid, obj.Mid)
	default:
		return errors.New("no bootName in config")
	}

	switch s.c.NewYearReservePubConfig.RuleType {
	// 1 => 指定sid推流
	case like.ActivityReserveAutoPushTypeOne:
		var err error
		ruleConfig := s.c.NewYearReservePubConfig.RuleConfig
		for _, v := range ruleConfig.SIDs {
			if v == obj.Sid {
				message := &like.ActivityReservePub{
					Sid:   obj.Sid,
					Mid:   obj.Mid,
					State: obj.State,
				}
				for i := 0; i < 3; i++ {
					err = boot.Send(c, sendKey, message)
					if err == nil {
						break
					}
				}
				if err != nil {
					log.Errorc(c, "PubReserveActivityData Auto Pub %s key[%s] msg[%v] err[%v]", bootName, sendKey, *message, err)
				}
			}
		}
	}

	return nil
}

func (s *Service) CopyReserveItem2NewTable(ctx context.Context, action string, newData *like.ActReserveField) {
	var err error
	// 新增case
	if action == match.ActInsert {
		if err = s.dao.Insert2NewReserveTable(ctx, newData); err != nil {
			log.Errorc(ctx, "CopyReserveItem2NewTable s.dao.Insert2NewReserveTable Err newData(%+v) err(%+v)", newData, err)
		}
	} else {
		if err = s.dao.NewReserveTableOnDuplicate(ctx, newData); err != nil {
			log.Errorc(ctx, "CopyReserveItem2NewTable s.dao.NewReserveTableOnDuplicate Err newData(%+v) err(%+v)", newData, err)
		}
	}

	return
}

func (s *Service) ReserveAuditHandleByArcCronPubState(ctx context.Context, oldObj *like.Archive, newObj *like.Archive) (err error) {
	// 定时发布审核通过
	if newObj.State == like.StateForbidUserDelay && oldObj.State != like.StateForbidUserDelay {
		if err = s.ReserveAuditHandleByArcCronPubStateAccess(ctx, oldObj, newObj); err != nil {
			return errors.Wrap(err, "s.ReserveAuditHandleByArcCronPubStateAccess err")
		}
	}
	// 审核驳回
	if function.InInt64Slice(int64(newObj.State), []int64{like.StateForbidRecycle, like.StateForbidPolice, like.StateForbidLock}) &&
		!function.InInt64Slice(int64(oldObj.State), []int64{like.StateForbidRecycle, like.StateForbidPolice, like.StateForbidLock}) {
		if err = s.ReserveAuditHandleByArcCronPubStateReject(ctx, oldObj, newObj); err != nil {
			return errors.Wrap(err, "s.ReserveAuditHandleByArcCronPubStateReject err")
		}
	}
	// 定时发布审核中被删除
	if newObj.State == like.StateForbidUpDelete && oldObj.State != like.StateForbidUpDelete {
		if err = s.ReserveAuditHandleByArcCronDelete(ctx, oldObj, newObj); err != nil {
			return errors.Wrap(err, "s.ReserveAuditHandleByArcCronDelete err")
		}
	}

	return nil
}

func (s *Service) ReserveAuditHandleByArcCronPubStateAccess(ctx context.Context, oldObj *like.Archive, newObj *like.Archive) error {
	// 查询这个稿件 等待渠道审核结果的数据
	relation, err := s.dao.GetUpActReserveWaitingArcAudit(ctx, strconv.FormatInt(newObj.Aid, 10))
	if err != nil {
		return errors.Wrapf(err, "s.dao.GetUpActReserveWaitingArcAudit err relation(%+v)", relation)
	}
	// 数据不存在返回
	if relation == nil || relation.Sid == 0 {
		return nil
	}
	// 如果数据存在 非定时发布首次或非定时发布的稿件不做处理
	if relation.AuditChannel != like.UpActReserveAuditChannelArchive {
		return nil
	}
	// 首次审核通过 等待稿件审核
	// act_subject 表 允许被预约  等待审核渠道 变为0 审核中变为审核通过
	if err = s.dao.UpActReserveUnionChangeState(ctx, relation.Sid, like.ActSubjectStateNormal, like.UpActReservePass, like.UpActReserveAuditChannelDefault, relation.State); err != nil {
		return errors.Wrapf(err, "s.dao.UpActReserveUnionChangeState err relation(%+v)", relation)
	}
	// 如果状态修改成功 需要创建动态id
	// 首选获取稿件封面
	arc, err := s.dao.GetArchiveInfo(ctx, newObj.Aid)
	if err != nil {
		return errors.Wrapf(err, "s.dao.GetArchiveInfo err aid(%+v)", newObj.Aid)
	}
	if arc == nil || arc.Arc == nil || arc.Arc.Aid == 0 {
		return errors.Wrapf(err, "s.dao.GetArchiveInfo == nil aid(%+v)", newObj.Aid)
	}

	// 拿到稿件cover url 获取图片大小
	bfsInfo, err := s.dao.GetBFSFileInfo(ctx, arc.Arc.Cover+"@info")
	if err != nil {
		return errors.Wrapf(err, "s.dao.GetBFSFileInfo err arc(%+v) bfsInfo(%+v)", arc, bfsInfo)
	}
	if bfsInfo == nil || bfsInfo.FileSize == 0 || bfsInfo.Height == 0 || bfsInfo.Width == 0 {
		return errors.Wrapf(err, "s.dao.GetBFSFileInfo bfsInfo == nil cover(%s)", arc.Arc.Cover)
	}

	// 创建动态获取id
	dynamicID, err := s.dao.CreateDynamicData(ctx, s.dao.BuildDynamicData(ctx, relation, arc.Arc.Cover, bfsInfo))
	if err != nil {
		log.Warnc(ctx, like.UpActReserveArcCronLogPrefix+"s.dao.CreateDynamicData error(%+v)", err)
		err = nil
	}

	// 插入关联表内
	err = s.dao.TXUpdateSubjectAndRelationData(ctx, relation.Sid, relation.Oid, relation.Type, dynamicID, int64(api.UpCreateActReserveFrom_FromDynamic))
	if err != nil {
		return errors.Wrapf(err, "s.dao.CreateUpActReserveRelationBind relation(%+v) dynamic_id(%+v)", relation, dynamicID)
	}

	return nil
}

func (s *Service) ReserveAuditHandleByArcCronPubStateReject(ctx context.Context, oldObj *like.Archive, newObj *like.Archive) error {
	// 查询这个稿件 等待渠道审核结果的数据
	var sid int64
	relation, err := s.dao.GetUpActReserveWaitingArcAudit(ctx, strconv.FormatInt(newObj.Aid, 10))
	if err != nil {
		return fmt.Errorf("s.dao.GetUpActReserveWaitingArcAudit error(%+v)", err)
	}
	sid = relation.Sid
	// 如果数据存在 判断audit_status 为首次 则直接干掉预约
	if relation.AuditChannel == like.UpActReserveAuditChannelArchive {
		if err = s.dao.UpActReserveUnionChangeState(ctx, relation.Sid, like.ActSubjectStateReject, like.UpActReserveReject, like.UpActReserveAuditChannelDefault, int64(api.UpActReserveRelationState_UpReserveReject)); err != nil {
			return fmt.Errorf("s.dao.UpActReserveUnionChangeState error(%+v)", err)
		}
	}
	// 如果数据不存在或者非首次 无法证明是否为定时发布渠道
	if relation == nil || relation.Sid == 0 {
		// 需要去另外一张表里面查询 是否包含定时发布记录
		bind, err := s.dao.GetUpActReserveRelationBindInfo(ctx, strconv.FormatInt(newObj.Aid, 10), int64(api.UpActReserveRelationType_Archive), int64(api.UpCreateActReserveFrom_FromDynamic))
		if err != nil {
			return fmt.Errorf("s.dao.GetUpActReserveRelationBindInfo error(%+v)", err)
		}
		// 不包含则为非定时发布的稿件不关心
		if bind == nil || bind.Sid == 0 {
			log.Infoc(ctx, "s.dao.GetUpActReserveRelationBindInfo empty newObj(%+v)", newObj)
			return nil
		}
		sid = bind.Sid
	}

	// 如果包含则已经过审过 二审驳回需要送审
	// 根据预约sid获取预约标题信息
	bindRelation, err := s.dao.GetUpActReserveRelationInfoBySid(ctx, sid)
	if err != nil {
		return fmt.Errorf("s.dao.GetUpActReserveRelationInfoBySid error(%+v)", err)
	}
	if bindRelation == nil || bindRelation.ID == 0 {
		return fmt.Errorf("s.dao.GetUpActReserveRelationInfoBySid bindRelation nil bindRelation(%+v)", bindRelation)
	}
	subject, err := s.dao.ActSubjectFromMaster(ctx, sid)
	if err != nil {
		return fmt.Errorf("s.dao.ActSubject error(%+v) sid(%+v)", err, sid)
	}
	if subject == nil || subject.ID == 0 {
		return fmt.Errorf("s.dao.ActSubject nil sid(%+v)", sid)
	}
	level, err, reply := s.dao.FilterTitle(ctx, bindRelation.Mid, subject.Name)
	if err != nil {
		return fmt.Errorf("s.dao.FilterTitle error(%+v)", err)
	}
	if function.InInt64Slice(level, []int64{like.SensitiveLevelIntercept20, like.SensitiveLevelIntercept30, like.SensitiveLevelIntercept40}) {
		// 命中拒绝直接审核驳回
		if err = s.dao.UpActReserveUnionChangeState(ctx, relation.Sid, like.ActSubjectStateReject, like.UpActReserveReject, like.UpActReserveAuditChannelDefault, int64(api.UpActReserveRelationState_UpReserveReject)); err != nil {
			return fmt.Errorf("s.dao.UpActReserveUnionChangeState error(%+v)", err)
		}
		return nil
	}
	// 如果先审后发 送审
	if level == like.SensitiveLevelAudit {
		if err = s.dao.Go2Audit(ctx, bindRelation.Sid, bindRelation.Mid, subject.Name, reply); err != nil {
			return fmt.Errorf("s.dao.Go2Audit error(%+v)", err)
		}
		if err = s.dao.UpActReserveUnionChangeState(ctx, relation.Sid, like.ActSubjectStateAudit, like.UpActReserveAudit, like.UpActReserveAuditChannelPlatform, relation.State); err != nil {
			return fmt.Errorf("s.dao.UpActReserveUnionChangeState error(%+v)", err)
		}
		return nil
	}
	// 先发后审 也要送审
	if level == like.SensitiveLevelPass {
		if err = s.dao.Go2Audit(ctx, bindRelation.Sid, bindRelation.Mid, subject.Name, reply); err != nil {
			return fmt.Errorf("s.dao.Go2Audit error(%+v)", err)
		}
		if err = s.dao.UpActReserveUnionChangeState(ctx, relation.Sid, like.ActSubjectStateNormal, like.UpActReservePassDelayAudit, like.UpActReserveAuditChannelPlatform, relation.State); err != nil {
			return fmt.Errorf("s.dao.UpActReserveUnionChangeState error(%+v)", err)
		}
		return nil
	}
	return nil
}

func (s *Service) ReserveAuditHandleByArcCronDelete(ctx context.Context, oldObj *like.Archive, newObj *like.Archive) error {
	// 查询这个稿件 等待渠道审核结果的数据
	relation, err := s.dao.GetUpActReserveWaitingArcAudit(ctx, strconv.FormatInt(newObj.Aid, 10))
	if err != nil {
		return errors.Wrapf(err, "s.dao.GetUpActReserveWaitingArcAudit err relation(%+v)", relation)
	}
	// 数据不存在返回
	if relation == nil || relation.Sid == 0 {
		return nil
	}
	// 如果数据存在 非定时发布首次或非定时发布的稿件不做处理
	if relation.AuditChannel != like.UpActReserveAuditChannelArchive {
		return nil
	}
	// 稿件直接更新为撤销状态
	if err = s.dao.UpActReserveUnionChangeState(ctx, relation.Sid, like.ActSubjectStateCancel, like.UpActReservePass, like.UpActReserveAuditChannelDefault, int64(api.UpActReserveRelationState_UpReserveCancel)); err != nil {
		return errors.Wrapf(err, "s.dao.UpActReserveUnionChangeState err relation(%+v)", relation)
	}
	return nil
}
