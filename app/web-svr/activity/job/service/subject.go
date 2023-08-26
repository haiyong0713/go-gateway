package service

import (
	"context"
	"encoding/json"
	"fmt"
	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	liveapi "git.bilibili.co/bapis/bapis-go/live/xroom"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/component"
	fmdl "go-gateway/app/web-svr/activity/job/model/fit"
	"go-gateway/app/web-svr/activity/job/tool"
	"go-gateway/app/web-svr/activity/tools/lib/function"
	"go-main/pkg/idsafe/bvid"
	"strconv"
	"strings"
	"time"

	garb "git.bilibili.co/bapis/bapis-go/garb/service"
	bGroup "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	tunnel "git.bilibili.co/bapis/bapis-go/platform/service/tunnel/v2"

	"go-common/library/log"
	"go-common/library/xstr"
	l "go-gateway/app/web-svr/activity/job/model/like"

	tunnelCommon "git.bilibili.co/bapis/bapis-go/platform/common/tunnel"
	tunnelEcode "git.bilibili.co/bapis/bapis-go/platform/common/tunnel/ecode"
	videoup "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

const (
	_subTypeClockIn  = 22
	FLAGRESERVEPUSH  = uint(26)
	notPlayUrl       = 1
	notifyTypeCreate = 1
	notifyTypeReset  = 2
)

func tunnelFlagKey(sid int64) string {
	return fmt.Sprintf("tunnel_flag_%d", sid)
}

func tunnelCountKey(sid int64) string {
	return fmt.Sprintf("tunnel_count_%d", sid)
}

func (s *Service) subjectRuleStatproc() {
	defer s.waiter.Done()
	var (
		err error
	)
	if s.subRuleStatSub == nil {
		return
	}
	for {
		msg, ok := <-s.subRuleStatSub.Messages()
		if !ok {
			log.Info("subjectRuleStatproc exit!")
			return
		}
		msg.Commit()
		m := &l.SubRuleStat{}
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("subjectRuleStatproc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		if m.Mid == 0 || m.Raw.Rule == 0 {
			log.Warn("subjectRuleStatproc data(%+v) not allowed", m)
			continue
		}
		rule, err := s.dao.SubjectRule(context.Background(), m.Raw.Rule)
		if err != nil {
			log.Error("subjectRuleStatproc SubjectRule id:%d error(%v)", m.Raw.Rule, err)
			continue
		}
		if rule == nil || rule.State != 1 || rule.TaskID == 0 {
			log.Warn("subjectRuleStatproc rule:%+v not allowed", rule)
			continue
		}
		if m.StateChange == 2 { // 增加活动资格
			if err = s.dao.DoTask(context.Background(), rule.TaskID, m.Mid); err != nil {
				log.Error("subjectRuleStatproc DoTask task:%d mid:%d error(%v)", rule.TaskID, m.Mid, err)
				continue
			}
		} else if m.StateChange == 1 { // 去掉活动资格，只扣除计数
			if err = s.dao.UpTaskState(context.Background(), rule.Sid, rule.TaskID, m.Mid); err != nil {
				log.Error("subjectRuleStatproc UpTaskState sid:%d task:%d mid:%d error(%v)", rule.Sid, rule.TaskID, m.Mid, err)
				continue
			}
		}
		log.Info("subjectRuleStatproc success key:%s partition:%d offset:%d value:%s ", msg.Key, msg.Partition, msg.Offset, string(msg.Value))
	}
}

// upSubject update act_subject cache .
func (s *Service) upSubject(c context.Context, upMsg json.RawMessage, preMsg json.RawMessage) (err error) {
	var (
		subObj = new(l.ActSubject)
	)
	if err = json.Unmarshal(upMsg, subObj); err != nil {
		log.Error("upSubject json.Unmarshal(%s) error(%+v)", upMsg, err)
		return
	}
	if err = s.dao.SubjectUp(c, subObj.ID); err != nil {
		log.Error(" s.dao.SubjectUp(%d) error(%+v)", subObj.ID, err)
		return
	}
	if subObj.Type == _subTypeClockIn {
		if err = s.dao.UpArcEventRule(c, subObj.ID); err != nil {
			log.Error("s.dao.UpArcEventRule(%d) error(%+v)", subObj.ID, err)
			return
		}
	}
	if len(preMsg) > 0 {
		func() {
			preSubObj := new(l.ActSubject)
			if err = json.Unmarshal(upMsg, &preSubObj); err != nil {
				log.Error("upSubject json.Unmarshal(%s) error(%+v)", upMsg, err)
				return
			}
			if preSubObj.Flag != subObj.Flag {
				s.upSubjectLikeStickTop(c, preSubObj, subObj)
			}
		}()
	}
	// up arc list
	log.Info("upSubject success  s.dao.SubjectUp(%d)", subObj.ID)
	return
}

func (s *Service) upSubjectRule(c context.Context, upMsg json.RawMessage) (err error) {
	var (
		subObj = new(l.SubjectRule)
	)
	if err = json.Unmarshal(upMsg, subObj); err != nil {
		log.Error("upSubjectRule json.Unmarshal(%s) error(%+v)", upMsg, err)
		return
	}
	s.SyncFullData2Counter(c, subObj)
	if err = s.dao.DelCacheSubjectRulesBySid(c, subObj.Sid); err != nil {
		log.Error("upSubjectRule s.dao.DelCacheSubjectRulesBySid(%d) error(%+v)", subObj.Sid, err)
		return
	}
	if err = s.retryUpArcEventRule(c, subObj.Sid); err != nil {
		log.Error("upSubjectRule s.retryUpArcEventRule(%d) error(%+v)", subObj.Sid, err)
		return
	}
	log.Info("upSubjectRule success obj(%+v)", subObj)
	return
}

func (s *Service) retryUpArcEventRule(c context.Context, sid int64) (err error) {
	for i := 0; i < _retryTimes; i++ {
		if err = s.dao.UpArcEventRule(c, sid); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) loadAwardSubjects() {
	now := time.Now()
	awards, err := s.dao.AwardSubjectList(context.Background(), now)
	if err != nil {
		log.Error("loadAwardSubjects %v", err)
		return
	}
	tmp := make(map[int64]*l.AwardSubject)
	for _, v := range awards {
		tmp[v.Sid] = v
		sids, err := xstr.SplitInts(v.OtherSids)
		if err != nil {
			log.Error("loadAwardSubjects xstr.SplitInts %s error(%v)", v.OtherSids, err)
			continue
		}
		for _, sid := range sids {
			tmp[sid] = v
		}
	}
	s.awardSubjectTask = tmp
}

func isReservePush(flag int64) bool {
	return ((flag >> FLAGRESERVEPUSH) & int64(1)) == 1
}

func (s *Service) pushEditTunnel(ctx context.Context, newMsg, oldMsg json.RawMessage) {
	var (
		err        error
		newReserve = new(l.Reserve)
		oldReserve = new(l.Reserve)
	)
	if err = json.Unmarshal(newMsg, newReserve); err != nil {
		log.Error("pushEditTunnel json.Unmarshal(%s) error(%+v)", newMsg, err)
		return
	}
	if err = json.Unmarshal(oldMsg, oldReserve); err != nil {
		log.Error("pushEditTunnel json.Unmarshal(%s) error(%+v)", oldMsg, err)
		return
	}
	// 订阅判断不同时，更改推送
	if oldReserve.State != newReserve.State {
		s.pushTunnel(ctx, newReserve)
	}
}

func (s *Service) pushEditTunnelGroup(ctx context.Context, newMsg, oldMsg json.RawMessage) {
	var (
		err        error
		newReserve = new(l.Reserve)
		oldReserve = new(l.Reserve)
	)
	if err = json.Unmarshal(newMsg, newReserve); err != nil {
		log.Error("pushEditTunnel json.Unmarshal(%s) error(%+v)", newMsg, err)
		return
	}
	if err = json.Unmarshal(oldMsg, oldReserve); err != nil {
		log.Error("pushEditTunnel json.Unmarshal(%s) error(%+v)", oldMsg, err)
		return
	}
	// 订阅判断不同时，更改推送
	if oldReserve.State != newReserve.State {
		s.dao.AsyncSendGroupDatabus(ctx, newReserve, time.Now().Unix())
	}
}

func (s *Service) pushTunnel(ctx context.Context, data *l.Reserve) {
	var (
		flag, count int64
		err         error
	)
	if flag, err = s.subjectFlag(ctx, data.Sid); err != nil {
		log.Error("pushTunnel s.subjectFlag sid(%d) data(%+v) error(%v)", data.Sid, data, err)
		return
	}
	// 判断是否开启推送
	if !isReservePush(flag) {
		return
	}
	if count, err = s.templateCount(ctx, data.Sid); err != nil {
		log.Error("pushTunnel s.dao.TunnelTemplateCnt sid(%d) data(%+v) error(%v)", data.Sid, data, err)
		return
	}
	log.Info("pushTunnel Reserve mid(%d) data(%+v) template count(%d) ", data.Mid, data, count)
	// 判断是否添加模板
	if count == 0 {
		return
	}
	s.dao.AsyncSendTunnelDatabus(ctx, data)
}

func (s *Service) subjectFlag(ctx context.Context, sid int64) (res int64, err error) {
	flagKey := tunnelFlagKey(sid)
	if res, err = s.dao.CacheIntValue(ctx, flagKey); err != nil {
		log.Errorc(ctx, "s.dao.CacheIntValue sid(%+d) error(%+v)", sid, err)
		err = nil
	}
	if res > 0 {
		return
	}
	if res, err = s.dao.SubjectFlag(ctx, sid); err != nil {
		log.Error("s.dao.SubjectFlag sid(%d) error(%v)", sid, err)
		return
	}
	if res > 0 {
		s.dao.AddCacheIntValue(ctx, flagKey, res)
	}
	return
}

func (s *Service) templateCount(ctx context.Context, sid int64) (res int64, err error) {
	flagKey := tunnelCountKey(sid)
	if res, err = s.dao.CacheIntValue(ctx, flagKey); err != nil {
		log.Errorc(ctx, "s.dao.CacheIntValue sid(%+d) error(%+v)", sid, err)
		err = nil
	}
	if res > 0 {
		return
	}
	if res, err = s.dao.TunnelTemplateCnt(ctx, sid); err != nil {
		log.Error("s.dao.TunnelTemplateCnt sid(%d) error(%v)", sid, err)
		return
	}
	if res > 0 {
		s.dao.AddCacheIntValue(ctx, flagKey, res)
	}
	return
}

func (s *Service) CallInternalSyncActSubjectInfoDB2Cache() {
	ctx := context.Background()
	req := &api.InternalSyncActSubjectInfoDB2CacheReq{
		From: "JOB",
	}
	if _, err := s.actGRPC.InternalSyncActSubjectInfoDB2Cache(ctx, req); err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]CallInternalSyncActSubjectInfoDB2Cache Err(%v)", err)
		return
	}
	return
}

func (s *Service) CallInternalSyncActSubjectReserveIDsInfoDB2Cache() {
	ctx := context.Background()
	req := &api.InternalSyncActSubjectReserveIDsInfoDB2CacheReq{
		From: "JOB",
	}
	if _, err := s.actGRPC.InternalSyncActSubjectReserveIDsInfoDB2Cache(ctx, req); err != nil {
		log.Errorc(ctx, "[HOT-DATA-FAIL]CallInternalSyncActSubjectReserveIDsInfoDB2Cache Err(%v)", err)
		return
	}
	return
}

// 同步预约数据到渠道中台通知包
func (s *Service) CreateUpActNotify2Platform(ctx context.Context, newObj *l.ActSubject) (err error) {
	if err = s.Regist2PlatformGroupRepeatable(ctx, newObj.ID); err != nil {
		err = errors.Wrapf(err, "Regist2PlatformGroupRepeatable err")
		return
	}
	if err = s.Regist2PlatformTunnelRepeatable(ctx, newObj.ID); err != nil {
		err = errors.Wrapf(err, "Regist2PlatformTunnelRepeatable err")
		return
	}
	return
}

func (s *Service) UpActReserveRelationChanged(ctx context.Context, oldObj *l.UpActReserveRelation, newObj *l.UpActReserveRelation) {
	preData := l.UpAct41{
		Mid:  newObj.Mid,
		Sid:  newObj.Sid,
		Time: time.Now().UnixNano() / 1000,
	}

	// 4月2日不推数据
	if time.Now().Unix() > 1617292800 {
		return
	}

	// -200 草稿太数据不care
	if newObj.State == int64(api.UpActReserveRelationState_UpReserveEdit) {
		return
	}

	// 展示挂件
	if tool.InInt64Slice(newObj.State, []int64{int64(api.UpActReserveRelationState_UpReserveRelated), int64(api.UpActReserveRelationState_UpReserveRelatedOnline)}) {
		preData.State = 1
	} else {
		// 隐藏挂件
		preData.State = -1
	}
	data, _ := json.Marshal(preData)
	var err error
	for i := 0; i < 3; i++ {
		if err = component.UpActReserveRelationPub.Send(ctx, strconv.FormatInt(newObj.Mid, 10), data); err == nil {
			return
		}
	}
	// 接口降级
	req := &garb.AppointActivityLogReq{
		Mid:         preData.Mid,
		Sid:         preData.Sid,
		TimeVersion: preData.Time,
		State:       preData.State,
	}
	for i := 0; i < 3; i++ {
		if _, err = client.GarbClient.AppointActivityLog(ctx, req); err == nil {
			return
		}
	}
	log.Errorc(ctx, "[UpActReserveRelationChanged]rpc failed err(%v) req(%+v)", err, req)
	return
}

func (s *Service) UpActReserveChanged(ctx context.Context, newObj *l.SubjectStat) {
	var (
		err   error
		total int64
		res   *api.GetActReserveTotalReply
	)

	// 4月2日不推数据
	if time.Now().Unix() > 1617292800 {
		return
	}

	// rpc获取预约数量
	for i := 0; i < 3; i++ {
		if res, err = client.ActivityClient.GetActReserveTotal(ctx, &api.GetActReserveTotalReq{
			Sid: newObj.Sid,
		}); err == nil {
			total = res.Total
			break
		}
	}
	if err != nil {
		log.Errorc(ctx, "[UpActReserveChanged]get rpc data failed err(%v) sid(%v)", err, newObj.Sid)
		return
	}

	// send databus数据
	preData := l.UpActReserve41{
		Sid:   newObj.Sid,
		Time:  time.Now().UnixNano() / 1000,
		Total: total,
	}
	data, _ := json.Marshal(preData)
	for i := 0; i < 3; i++ {
		if err = component.UpActReservePub.Send(ctx, strconv.FormatInt(newObj.Sid, 10), data); err == nil {
			return
		}
	}
	// 接口降级
	req := &garb.AppointActivityTotalLogReq{
		Sid:         preData.Sid,
		Total:       total,
		TimeVersion: preData.Time,
	}
	for i := 0; i < 3; i++ {
		if _, err = client.GarbClient.AppointActivityTotalLog(ctx, req); err == nil {
			return
		}
	}
	log.Errorc(ctx, "[UpActReserveChanged]get rpc data failed err(%v) req(%+v)", err, req)
	return
}

// 人群包创建 https://info.bilibili.co/pages/viewpage.action?pageId=184996626#id-%E4%BA%BA%E7%BE%A4%E5%8C%85service%E6%9C%8D%E5%8A%A1%E6%8E%A5%E5%8F%A3%E6%96%87%E6%A1%A3-%E4%BA%BA%E7%BE%A4%E5%8C%85%E5%88%9B%E5%BB%BA
func (s *Service) Regist2PlatformGroupRepeatable(ctx context.Context, sid int64) (err error) {
	req := &bGroup.AddBGroupReq{
		Type:       3,
		Name:       strconv.FormatInt(sid, 10),
		AppName:    "pink",
		Business:   "activity",
		Creator:    "activity",
		Definition: "{\"oid\":" + strconv.FormatInt(sid, 10) + "}",
		Dimension:  1,
	}

	err = retry.WithAttempts(ctx, "", l.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = client.BGroupClient.AddBGroup(ctx, req)
		log.Infoc(ctx, "client.BGroupClient.AddBGroup req(%+v)", req)
		if ecode.Cause(err).Code() == l.ErrCodeBGroupExist { //人群包已经存在
			err = nil
		}
		return err
	})

	if err != nil {
		return errors.Wrapf(err, "client.BGroupClient.AddBGroup err")
	}

	return
}

// 东风平台事件注册 https://info.bilibili.co/pages/viewpage.action?pageId=184989531#id-%E4%B8%9C%E9%A3%8E%E7%BB%9F%E4%B8%80%E5%8D%8F%E8%AE%AE%E4%B8%9A%E5%8A%A1%E5%AF%B9%E6%8E%A5-%E6%B3%A8%E5%86%8C%E4%BA%8B%E4%BB%B6%E6%B3%A8%E5%86%8C%E4%BA%8B%E4%BB%B6%E6%8E%A5%E5%8F%A3
func (s *Service) Regist2PlatformTunnelRepeatable(ctx context.Context, sid int64) (err error) {
	req := &tunnel.AddEventReq{
		BizId:    l.PlatformActivityBizID,
		UniqueId: sid,
		Title:    "up主预约活动" + strconv.FormatInt(sid, 10),
	}

	err = retry.WithAttempts(ctx, "", l.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = client.TunnelClient.AddEvent(ctx, req)
		log.Infoc(ctx, "Regist2PlatformTunnelRepeatable client.TunnelClient.AddEvent req(%+v)", req)
		if ecode.Cause(err).Code() == fmdl.TunnelV2EventAlready { // 事件已注册不用返回错误
			err = nil
		}
		return err
	})

	if err != nil {
		return errors.Wrapf(err, "client.TunnelClient.AddEvent err")
	}

	return
}

// 初始化参数
type info struct {
	aid               int64                      // 稿件
	bid               string                     // 稿件bvid
	link              string                     // 跳转地址
	linkText          string                     // 跳转链接文案
	triggerType       string                     // 触发方式
	startTime         string                     // 开始投递时间
	endTime           string                     // 结束投递时间
	duration          string                     // 投递时长
	moduleTitle       string                     // 模块标题
	description       string                     // 备注
	livePlanStartTime string                     // 直播预计开始时间
	upNickName        string                     // up主昵称
	moduleContents    []string                   // 模块内容
	notifyCode        string                     // 通知类型
	senderID          int64                      // 私信发送账号id
	cardUniqueId      int64                      // 卡片类型
	cardResource      *tunnelCommon.CardResource // 卡片资源
}

func (s *Service) RegisterPushVerifyCard(ctx context.Context, uniqueId int64, title string) (err error) {
	req := &tunnel.AddEventReq{
		BizId:    l.PlatformActivityBizID,
		UniqueId: uniqueId,
		Title:    title,
	}
	err = retry.WithAttempts(ctx, "", l.UpActReserverelationRetry, netutil.DefaultBackoffConfig, func(c context.Context) (err error) {
		_, err = client.TunnelClient.AddEvent(ctx, req)
		log.Infoc(ctx, "RegisterPushVerifyCard client.TunnelClient.AddEvent req(%v+v)", req)
		return
	})
	if ecode.EqualError(tunnelEcode.EventAlreadyRegistered, err) {
		err = nil
		return
	}
	if err != nil {
		err = errors.Wrapf(err, "client.TunnelClient.AddEvent err")
	}
	return
}

// 创建私信卡片
func (s *Service) CreatePlatformNotifyRepeatable(ctx context.Context, newObj *l.UpActReserveRelation, notifyType int) (err error) {
	info := new(info)

	info.upNickName, err = s.upNickName(ctx, newObj.Mid)
	if err != nil {
		return errors.Wrapf(err, "s.upNickName err")
	}

	// 填写稿件或直播建卡参数
	if newObj.Type == int64(api.UpActReserveRelationType_Archive) {
		err = s.fillArcNotifyCardInfo(ctx, newObj, info, notifyType)
	} else if newObj.Type == int64(api.UpActReserveRelationType_Live) {
		err = s.fillLiveNotifyCardInfo(ctx, newObj, info, notifyType)
	}
	if err != nil {
		return errors.Wrapf(err, "fill info err, reserve type: %v", newObj.Type)
	}

	bGroupInfo := &tunnelCommon.BGroupInfo{
		Name:     strconv.FormatInt(newObj.Sid, 10),
		Business: "activity",
	}
	bGI, _ := json.Marshal(bGroupInfo)

	req := &tunnel.UpsertCardMsgTemplateReq{
		BizId:        l.PlatformActivityBizID,
		UniqueId:     newObj.Sid,
		CardUniqueId: info.cardUniqueId,
		TriggerType:  info.triggerType,
		StartTime:    info.startTime,
		EndTime:      info.endTime,
		Duration:     info.duration,
		TargetUserGroup: &tunnelCommon.TargetUserGroup{
			UserType: 3,
			UserInfo: string(bGI),
		},
		CardContent: &tunnelCommon.MsgTemplateCardContent{
			NotifyCode:     info.notifyCode,
			SenderUid:      info.senderID,
			Resource:       info.cardResource,
			ModuleContents: info.moduleContents,
		},
		Description: info.description,
	}

	if info.link != "" && req.CardContent != nil {
		req.CardContent.JumpUriConfigs = []*tunnelCommon.MsgUriPlatform{
			{
				AllUri: info.link,
				Text:   info.linkText,
			},
		}
	}

	err = retry.WithAttempts(ctx, "CreatePlatformNotifyRepeatable", l.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		_, err = client.TunnelClient.UpsertCardMsgTemplate(ctx, req)
		log.Infoc(ctx, "client.TunnelClient.UpsertCardMsgTemplate req(%+v)", req)
		if ecode.Cause(err).Code() == l.ErrCodeTunnelCardState { // 卡片状态错误
			err = nil
		}
		return err
	})
	if err != nil {
		err = errors.Wrap(err, "client.TunnelClient.UpsertCardMsgTemplate err")
	}
	return
}

// 创建or更新 动态卡
func (s *Service) CreatePlatformCardRepeatable(ctx context.Context, newObj *l.UpActReserveRelation) (err error) {
	// 获取活动基本信息
	subject, err := s.dao.ActSubjectFromMaster(ctx, newObj.Sid)
	if err != nil {
		return errors.Wrapf(err, "s.dao.ActSubjectFromMaster err")
	}
	if subject == nil || subject.ID == 0 {
		return errors.Errorf("s.dao.ActSubjectFromMaster nil subject(%+v)", subject)
	}

	// 初始化参数
	var (
		aid         int64  // 稿件
		cover       string // 封面
		bid         string // 稿件bvid
		link        string // 跳转地址
		triggerType string // 触发方式
		duration    string // 投递时长
		description string // 描述
	)
	cardResource := &tunnelCommon.CardResource{}

	// 稿件
	if newObj.Type == int64(api.UpActReserveRelationType_Archive) {
		// 获取稿件封面信息
		aid, err = strconv.ParseInt(newObj.Oid, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "CreatePlatformCardRepeatable StrConv.ParseInt err")
		}
		var arcReply *videoup.ArchiveSimpleReply
		err = retry.WithAttempts(ctx, "CreatePlatformCardRepeatable", l.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			arcReply, err = client.VideoClient.ArchiveSimple(ctx, &videoup.ArchiveSimpleReq{
				Aid: aid,
			})
			log.Infoc(ctx, "client.VideoClient.ArchiveSimple aid(%+v) reply(%+v)", aid, arcReply)
			if ecode.EqualError(tunnelEcode.CardNotExists, err) {
				err = nil
			}
			return err
		})
		if err != nil {
			return errors.Wrapf(err, "CreatePlatformCardRepeatable s.arcs failed")
		}
		if arcReply == nil || arcReply.Arc == nil || arcReply.Arc.Aid == 0 {
			return errors.Errorf("CreatePlatformCardRepeatable s.arcs nil reply(%+v)", arcReply)
		}

		bid, err = bvid.AvToBv(aid)
		if err != nil {
			return errors.Wrapf(err, "CreatePlatformCardRepeatable bvid.AvToBv(aid) err")
		}

		cover = arcReply.Arc.Cover
		link = "https://www.bilibili.com/video/" + bid
		triggerType = "archive"
		duration = "72h"
		cardResource = &tunnelCommon.CardResource{
			Type: "ugc",
			Oid:  newObj.Oid,
		}
		description = "动态稿件卡"
	} else if newObj.Type == int64(api.UpActReserveRelationType_Live) {
		triggerType = "live"
		cardResource = &tunnelCommon.CardResource{
			Type: "live_session",
			Oid:  newObj.Oid,
		}
		duration = "72h"
		description = "动态直播卡"
	}

	bGroupInfo := &tunnelCommon.BGroupInfo{
		Name:     strconv.FormatInt(newObj.Sid, 10),
		Business: "activity",
	}
	bGI, _ := json.Marshal(bGroupInfo)

	req := &tunnel.UpsertCardDynamicBasicReq{
		BizId:        l.PlatformActivityBizID,
		UniqueId:     newObj.Sid,
		CardUniqueId: newObj.Sid,
		TriggerType:  triggerType,
		Duration:     duration,
		TargetUserGroup: &tunnelCommon.TargetUserGroup{
			UserType: 3,
			UserInfo: string(bGI),
		},
		CardContent: &tunnelCommon.DynamicBasicCardContent{
			Title:     subject.Name,
			Text:      subject.Name,
			Icon:      cover,
			Link:      link,
			Resource:  cardResource,
			SenderUid: newObj.Mid,
		},
		Description: description,
	}
	err = retry.WithAttempts(ctx, "CreatePlatformCardRepeatable", l.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		_, err = client.TunnelClient.UpsertCardDynamicBasic(ctx, req)
		log.Infoc(ctx, "client.TunnelClient.UpsertCardDynamicBasic req(%+v)", req)
		if ecode.Cause(err).Code() == l.ErrCodeTunnelCardState { // 卡片状态错误
			err = nil
		}
		return err
	})
	if err != nil {
		err = errors.Wrap(err, "client.TunnelClient.UpsertCardDynamicBasic err")
		return
	}

	return
}

// 删除私信卡片
func (s *Service) DeletePlatformNotifyRepeatable(ctx context.Context, oldObj *l.UpActReserveRelation, newObj *l.UpActReserveRelation) (err error) {
	req := &tunnel.OperateCardReq{
		BizId:        l.PlatformActivityBizID,
		UniqueId:     newObj.Sid,
		CardUniqueId: int64(tunnelCommon.MessageNotifyCard),
	}
	err = retry.WithAttempts(ctx, "DeletePlatformNotifyRepeatable", l.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		_, err = client.TunnelClient.DeleteCard(ctx, req)
		log.Infoc(ctx, "client.TunnelClient.DeleteCard req(%+v)", req)
		if ecode.EqualError(tunnelEcode.CardNotExists, err) {
			err = nil
		}
		return err
	})
	if err != nil {
		err = errors.Wrap(err, "client.TunnelClient.DeleteCard err")
		return
	}

	return
}

// 删除动态卡片
func (s *Service) DeletePlatformCardRepeatable(ctx context.Context, oldObj *l.UpActReserveRelation, newObj *l.UpActReserveRelation) (err error) {
	req := &tunnel.OperateCardReq{
		BizId:        l.PlatformActivityBizID,
		UniqueId:     newObj.Sid,
		CardUniqueId: newObj.Sid,
	}
	err = retry.WithAttempts(ctx, "DeletePlatformCardRepeatable", l.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		_, err = client.TunnelClient.DeleteCard(ctx, req)
		log.Infoc(ctx, "client.TunnelClient.DeleteCard req(%+v)", req)
		if ecode.EqualError(tunnelEcode.CardNotExists, err) {
			err = nil
		}
		return err
	})
	if err != nil {
		err = errors.Wrap(err, "client.TunnelClient.DeleteCard err")
		return
	}
	return
}

//// 撤销预约创建卡片
//func (s *Service) CreatePlatformResetNotifyRepeatable(ctx context.Context, oldObj *l.UpActReserveRelation, newObj *l.UpActReserveRelation) (err error) {
//	// 初始化参数
//	var (
//		triggerType string // 触发方式
//		upNickName  string // up主昵称
//	)
//
//	// 判断数据类型
//	if !tool.InInt64Slice(newObj.Type, []int64{int64(api.UpActReserveRelationType_Archive), int64(api.UpActReserveRelationType_Live)}) {
//		err = errors.Wrap(err, "illegal type err")
//		return
//	}
//
//	upNickName, err = s.upNickName(ctx, newObj.Mid)
//	if err != nil {
//		err = errors.Wrap(err, "s.upNickName err")
//		return
//	}
//
//	if newObj.Type == int64(api.UpActReserveRelationType_Archive) {
//		triggerType = "archive"
//	}
//	if newObj.Type == int64(api.UpActReserveRelationType_Archive) {
//		triggerType = "time"
//	}
//
//	bGroupInfo := &tunnelCommon.BGroupInfo{
//		Name:     strconv.FormatInt(newObj.Sid, 10),
//		Business: "activity",
//	}
//	bGI, _ := json.Marshal(bGroupInfo)
//
//	sTime := function.Now()
//	eTime := sTime + 60*60*72
//	req := &tunnel.UpsertCardMsgTemplateReq{
//		BizId:        l.PlatformActivityBizID,
//		UniqueId:     newObj.Sid,
//		CardUniqueId: l.NotifyMessageTypeResetReserve,
//		TriggerType:  triggerType,
//		StartTime:    time.Unix(sTime, 0).Format("2006-01-02 15:04:05"),
//		EndTime:      time.Unix(eTime, 0).Format("2006-01-02 15:04:05"),
//		TargetUserGroup: &tunnelCommon.TargetUserGroup{
//			UserType: 3,
//			UserInfo: string(bGI),
//		},
//		CardContent: &tunnelCommon.MsgTemplateCardContent{
//			NotifyCode: s.c.UpActReserveNotify.ResetTmpID,
//			Params:     []string{upNickName},
//			SenderUid:  s.c.UpActReserveNotify.ResetSenderUID,
//		},
//		Description: "私信卡-up主撤销预约",
//	}
//
//	err = retry.WithAttempts(ctx, "CreatePlatformResetNotifyRepeatable", l.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
//		_, err = client.TunnelClient.UpsertCardMsgTemplate(ctx, req)
//		log.Infoc(ctx, "client.TunnelClient.UpsertCardMsgTemplate req(%+v)", req)
//		return err
//	})
//	if err != nil {
//		return errors.Wrapf(err, "client.TunnelClient.UpsertCardMsgTemplate err")
//	}
//
//	return
//}

func (s *Service) UpActReserveRelationTableMonitor(ctx context.Context, action string, old *l.UpActReserveRelation, new *l.UpActReserveRelation) {
	var err error
	preData := l.UpActReserveRelationMonitor{
		Action:      action,
		Old:         old,
		New:         new,
		TimeVersion: time.Now().UnixNano() / 1000,
	}
	data, _ := json.Marshal(preData)
	partitionKey := strconv.FormatInt(new.Mid, 10) + strconv.FormatInt(new.Sid, 10)
	for i := 0; i < 3; i++ {
		if err = component.UpActReserveRelationTableMonitor.Send(ctx, partitionKey, data); err == nil {
			return
		}
	}
	log.Errorc(ctx, "[UpActReserveRelationTableMonitor]err(%v) pre(%+v)", err, preData)
	return
}

func (s *Service) DeleteDynamicRelatedByLiveReserve(ctx context.Context, oldObj, newObj *l.UpActReserveRelation) (err error) {
	log.Infoc(ctx, "DeleteDynamicRelatedByLiveReserve oldObj(%+v), newObj(%+v)", oldObj, newObj)
	newState := api.UpActReserveRelationState(newObj.State)
	if newState == api.UpActReserveRelationState_UpReserveCancel || newState == api.UpActReserveRelationState_UpReserveReject {
		if newObj.From == int64(api.UpCreateActReserveFrom_FromBiliApp) || newObj.From == int64(api.UpCreateActReserveFrom_FromBiliLive) ||
			newObj.From == int64(api.UpCreateActReserveFrom_FROMPCBILILIVE) || newObj.From == int64(api.UpCreateActReserveFrom_FROMBILIWEB) {
			if newObj.DynamicID == "" {
				err = errors.Wrapf(err, "s.DeleteDynamicRelatedByLiveReserve err, dynamicID is nil:(%+v)", newObj)
				return
			}
			if err = s.dao.DeleteDynamicRelatedByLiveReserve(ctx, newObj.Mid, newObj.DynamicID); err != nil {
				err = errors.Wrap(err, "s.dao.DeleteDynamicRelatedByLiveReserve err")
				return
			}
		}
	}
	return
}

// 新版状态流转
func (s *Service) BindCard2PlatformRepeatableNew(ctx context.Context, oldObj *l.UpActReserveRelation, newObj *l.UpActReserveRelation) (err error) {
	log.Infoc(ctx, "BindCard2PlatformRepeatableNew oldObj(%+v) newObj(%+v)", oldObj, newObj)
	state, typ, audit := api.UpActReserveRelationState(newObj.State), api.UpActReserveRelationType(newObj.Type), newObj.Audit
	// 审核通过
	if tool.InInt64Slice(audit, []int64{l.UpActReservePassDelayAudit, l.UpActReservePass}) {
		if typ == api.UpActReserveRelationType_Archive {
			switch state {
			case api.UpActReserveRelationState_UpReserveReject: // -300 预约审核不通过删除（删除所有卡片）
				if err = s.UpActReserveDeleteGifts(ctx, oldObj, newObj); err != nil {
					return errors.Wrap(err, "UpActReserveDeleteGifts err")
				}
			case api.UpActReserveRelationState_UpReserveCancel: // -100	预约被UP主取消（删除所有卡片 创建私信撤销卡）
				if err = s.UpActReserveDeleteGifts(ctx, oldObj, newObj); err != nil {
					return errors.Wrap(err, "UpActReserveDeleteGifts err")
				}
				//if err = s.CreatePlatformNotifyRepeatable(ctx, newObj, notifyTypeReset); err != nil {
				//	return errors.Wrapf(err, "CreatePlatformNotifyRepeatable err")
				//}
			case api.UpActReserveRelationState_UpReserveRelated: // 100	等待关联稿件或直播
				if err = s.UpActReserveDeleteGifts(ctx, oldObj, newObj); err != nil {
					return errors.Wrap(err, "UpActReserveDeleteGifts err")
				}
			case api.UpActReserveRelationState_UpReserveRelatedOnline: // 120	审核完毕等待核销（绑定成功等待核销）
				if err = s.UpActReserveCreateGifts(ctx, newObj); err != nil {
					return errors.Wrap(err, "UpActReserveCreateGifts err")
				}
			}

		} else if typ == api.UpActReserveRelationType_Live {
			switch state {
			case api.UpActReserveRelationState_UpReserveReject: // -300 预约审核不通过删除
				if err = s.UpActReserveDeleteGifts(ctx, oldObj, newObj); err != nil {
					return errors.Wrap(err, "UpActReserveDeleteGifts err")
				}
			case api.UpActReserveRelationState_UpReserveCancel: // -100	预约被UP主取消
				if err = s.CreatePlatformNotifyRepeatable(ctx, newObj, notifyTypeReset); err != nil {
					return errors.Wrapf(err, "CreatePlatformNotifyRepeatable err")
				}
			case api.UpActReserveRelationState_UpReserveRelatedWaitCallBack: // 130	 开始核销
				if err = s.UpActReserveCreateGifts(ctx, newObj); err != nil {
					return errors.Wrap(err, "CreatePlatformCardRepeatable err")
				}
			}
		}
	}

	if tool.InInt64Slice(audit, []int64{l.UpActReserveAudit, l.UpActReserveReject}) {
		if err = s.UpActReserveDeleteGifts(ctx, oldObj, newObj); err != nil {
			return errors.Wrap(err, "UpActReserveDeleteGifts err")
		}
	}

	return
}

// 删除大礼包 包括[动态（直播+稿件卡）+ 私信通知卡 + 私信撤销通知卡]
func (s *Service) UpActReserveDeleteGifts(ctx context.Context, oldObj *l.UpActReserveRelation, newObj *l.UpActReserveRelation) (err error) {
	if err = s.DeletePlatformCardRepeatable(ctx, oldObj, newObj); err != nil {
		return errors.Wrap(err, "DeletePlatformCardRepeatable err")
	}
	if err = s.DeletePlatformNotifyRepeatable(ctx, oldObj, newObj); err != nil {
		return errors.Wrap(err, "DeletePlatformNotifyRepeatable err")
	}
	if err = s.DeletePlatformResetNotifyRepeatable(ctx, oldObj, newObj); err != nil {
		return errors.Wrap(err, "DeletePlatformResetNotifyRepeatable err")
	}
	return
}

// 创建大礼包 包括[动态（直播+稿件卡）+ 私信通知卡]
func (s *Service) UpActReserveCreateGifts(ctx context.Context, newObj *l.UpActReserveRelation) (err error) {
	if err = s.CreatePlatformCardRepeatable(ctx, newObj); err != nil {
		return errors.Wrapf(err, "CreatePlatformCardRepeatable err")
	}
	if err = s.CreatePlatformNotifyRepeatable(ctx, newObj, notifyTypeCreate); err != nil {
		return errors.Wrapf(err, "CreatePlatformNotifyRepeatable err")
	}
	return
}

// 删除私信撤销卡片
func (s *Service) DeletePlatformResetNotifyRepeatable(ctx context.Context, oldObj *l.UpActReserveRelation, newObj *l.UpActReserveRelation) (err error) {
	req := &tunnel.OperateCardReq{
		BizId:        l.PlatformActivityBizID,
		UniqueId:     newObj.Sid,
		CardUniqueId: int64(l.NotifyMessageTypeResetReserve),
	}
	err = retry.WithAttempts(ctx, "DeletePlatformResetNotifyRepeatable", l.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		_, err = client.TunnelClient.DeleteCard(ctx, req)
		log.Infoc(ctx, "client.TunnelClient.DeleteCard req(%+v)", req)
		if ecode.EqualError(tunnelEcode.CardNotExists, err) {
			err = nil
		}
		return err
	})
	if err != nil {
		err = errors.Wrap(err, "DeletePlatformResetNotifyRepeatable err")
		return
	}

	return
}

func (s *Service) fillArcNotifyCardInfo(ctx context.Context, newObj *l.UpActReserveRelation, info *info, notifyType int) (err error) {
	if notifyType == notifyTypeCreate {
		// 获取稿件封面信息
		info.aid, err = strconv.ParseInt(newObj.Oid, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "strconv.ParseInt err")
		}
		info.bid, err = bvid.AvToBv(info.aid)
		if err != nil {
			return errors.Wrapf(err, "bvid.AvToBv(%+v) err", info.aid)
		}
		info.link = "https://www.bilibili.com/video/" + info.bid
		info.linkText = "观看视频"
		info.moduleTitle, err = s.arcTitle(ctx, info.aid)
		if err != nil {
			return errors.Wrapf(err, "s.arcTitle err")
		}
		info.notifyCode = s.c.UpActReserveNotify.ArcTemplateID
		info.senderID = newObj.Mid
		info.cardUniqueId = l.NotifyMessageTypeReserve
		info.triggerType = "archive"
		info.duration = "72h"
		info.cardResource = &tunnelCommon.CardResource{
			Type: "ugc",
			Oid:  newObj.Oid,
		}
		info.description = "私信卡-up主稿件对外开放"
	} else if notifyType == notifyTypeReset {
		subject, err := s.dao.ActSubject(ctx, newObj.Sid)
		if err != nil {
			return errors.Wrapf(err, "s.dao.ActSubject err")
		}
		if subject == nil || subject.ID == 0 {
			return errors.Errorf("s.dao.ActSubject nil subject(%+v)", subject)
		}
		info.moduleTitle = strings.TrimPrefix(subject.Name, "直播预约：")
		info.notifyCode = s.c.UpActReserveNotify.ResetTmpID
		info.senderID = s.c.UpActReserveNotify.ArcResetSenderUID
		info.cardUniqueId = l.NotifyMessageTypeResetReserve
		info.triggerType = "time"
		info.startTime = time.Unix(function.Now()+30, 0).Format("2006-01-02 15:04:05")
		info.endTime = time.Unix(function.Now()+30+60*60*72, 0).Format("2006-01-02 15:04:05")
		info.description = "稿件预约被up主撤销"
	}
	info.moduleContents = []string{info.moduleTitle, info.upNickName}
	return
}

func (s *Service) fillLiveNotifyCardInfo(ctx context.Context, newObj *l.UpActReserveRelation, info *info, notifyType int) (err error) {
	info.link, err = s.liveRoomJumpUrl(ctx, "", newObj.Mid)
	if err != nil {
		return errors.Wrapf(err, "s.liveRoomJumpUrl err")
	}
	info.duration = "72h"
	subject, err := s.dao.ActSubject(ctx, newObj.Sid)
	if err != nil {
		return errors.Wrapf(err, "s.dao.ActSubject err")
	}
	if subject == nil || subject.ID == 0 {
		return errors.Errorf("s.dao.ActSubject nil subject(%+v)", subject)
	}
	info.moduleTitle = strings.TrimPrefix(subject.Name, "直播预约：")
	if notifyType == notifyTypeCreate {
		info.linkText = "观看直播"
		info.description = "私信卡-up主开始直播"
		info.notifyCode = s.c.UpActReserveNotify.LiveTemplateID
		info.senderID = newObj.Mid
		info.triggerType = "live"
		info.livePlanStartTime = time.Unix(function.Now(), 0).Format("2006-01-02 15:04")
		info.cardUniqueId = l.NotifyMessageTypeReserve
		info.cardResource = &tunnelCommon.CardResource{
			Type: "live_session",
			Oid:  newObj.Oid,
		}
	} else if notifyType == notifyTypeReset {
		info.link = ""
		info.description = "直播被up主撤销"
		info.notifyCode = s.c.UpActReserveNotify.LiveResetTempID
		info.senderID = s.c.UpActReserveNotify.LiveResetSenderUID
		tmp := newObj.LivePlanStartTime
		if len(tmp) >= 3 {
			tmp = tmp[:len(tmp)-3]
		}
		info.livePlanStartTime = tmp
		info.cardUniqueId = l.NotifyMessageTypeResetReserve
		info.triggerType = "time"
		info.startTime = time.Unix(function.Now()+30, 0).Format("2006-01-02 15:04:05")
		info.endTime = time.Unix(function.Now()+30+60*60*72, 0).Format("2006-01-02 15:04:05")
	}
	info.moduleContents = []string{info.moduleTitle, info.livePlanStartTime, info.upNickName}
	return
}

// arcTitle
func (s *Service) arcTitle(ctx context.Context, aid int64) (title string, err error) {
	var arcReply *videoup.ArchiveSimpleReply
	err = retry.WithAttempts(ctx, "CreatePlatformNotifyRepeatable", l.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		arcReply, err = client.VideoClient.ArchiveSimple(ctx, &videoup.ArchiveSimpleReq{
			Aid: aid,
		})
		log.Infoc(ctx, "client.VideoClient.ArchiveSimple aid(%+v) reply(%+v)", aid, arcReply)
		if ecode.EqualError(tunnelEcode.CardNotExists, err) {
			err = nil
		}
		return err
	})
	if err != nil {
		err = errors.Wrapf(err, "client.VideoClient.ArchiveSimple err")
		return
	}
	if arcReply == nil || arcReply.Arc == nil || arcReply.Arc.Aid == 0 {
		err = errors.Errorf("client.VideoClient.ArchiveSimple reply nil")
		return
	}
	title = arcReply.Arc.Title
	return
}

// liveRoomJumpUrl
func (s *Service) liveRoomJumpUrl(ctx context.Context, path string, mid int64) (url string, err error) {
	var liveRes = &liveapi.EntryRoomInfoResp{}
	data := &liveapi.EntryRoomInfoReq{
		Uids:       []int64{mid},
		NotPlayurl: notPlayUrl,
		EntryFrom:  []string{"NONE"},
		ReqBiz:     path,
	}
	liveRes, err = client.LiveClient.EntryRoomInfo(ctx, data)
	if err != nil {
		err = errors.Wrapf(err, " client.LiveClient.EntryRoomInfo data(%+v) err(%v)", data, err)
		return
	}
	if liveRes == nil || liveRes.List == nil {
		err = errors.Wrapf(err, "client.LiveClient.EntryRoomInfo, got nil res, data:(%v)", data)
		return
	}
	url = liveRes.List[mid].JumpUrl["NONE"]
	return
}

func (s *Service) upNickName(ctx context.Context, mid int64) (upNickName string, err error) {
	// 获取账号up主昵称
	var reply *accapi.InfoReply
	err = retry.WithAttempts(ctx, "CreatePlatformNotifyRepeatable", l.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		reply, err = s.accClient.Info3(ctx, &accapi.MidReq{Mid: mid})
		log.Infoc(ctx, "s.accClient.Info3 mid(%+v) reply(%+v)", mid, reply)
		if err == nil && reply != nil && reply.Info != nil {
			upNickName = reply.Info.Name
		}
		return err
	})

	if err != nil {
		err = errors.Wrapf(err, "s.accClient.Info3 err mid(%+v)", mid)
		return
	}
	if reply == nil {
		err = errors.Wrapf(err, "s.accClient.Info3 reply nil mid(%+v)", mid)
	}
	return
}

// 创建抽奖私信卡片
func (s *Service) CreateUpActReserveRelationLotteryCard(ctx context.Context, relation *api.UpActReserveInfo) (err error) {
	// 发私信
	if !(relation.LotteryType == api.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron && relation.Type == api.UpActReserveRelationType_Live) {
		return
	}

	// 获取抽奖信息
	text, url, err := s.dao.GetDynamicLotteryInfo(ctx, strconv.FormatInt(relation.Sid, 10), int64(relation.Type))
	if err != nil {
		err = errors.Wrapf(err, "s.dao.GetDynamicLotteryInfo err")
		return
	}

	// 标题处理
	title := strings.Replace(relation.Title, "直播预约：", "", 1)

	req := &tunnel.UpsertCardMsgTemplateReq{
		BizId:        l.PlatformActivityBizID,
		UniqueId:     relation.Sid,
		CardUniqueId: l.NotifyMessageTypeLotteryReserve,
		TriggerType:  "time",
		StartTime:    time.Unix(function.Now(), 0).Format("2006-01-02 15:04:05"),
		EndTime:      time.Unix(function.Now()+86400*30*6, 0).Format("2006-01-02 15:04:05"), // 结束时间暂定是开始时间延后6个月
		CardContent: &tunnelCommon.MsgTemplateCardContent{
			NotifyCode:     s.c.UpActReserveNotify.LotteryReserveTmpID,
			ModuleContents: []string{title, relation.LivePlanStartTime.Time().Format("2006-01-02 15:04"), text, url},
			SenderUid:      relation.Upmid,
			JumpUriConfig: &tunnelCommon.MsgUriPlatform{
				AllUri: url,
				Text:   "查看详情",
			},
		},
		Description: fmt.Sprintf("私信卡-抽奖预约通知 预约ID:%d UPMID:%d", relation.Sid, relation.Upmid),
	}

	if err = retry.WithAttempts(ctx, "CreateUpActReserveRelationLotteryCard", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = client.TunnelClient.UpsertCardMsgTemplate(ctx, req)
		log.Infoc(ctx, "CreateUpActReserveRelationLotteryCard client.TunnelClient.UpsertCardMsgTemplate req(%+v)", req)
		return err
	}); err != nil {
		err = errors.Wrapf(err, "client.TunnelClient.UpsertCardMsgTemplate err")
		return
	}

	return
}

func (s *Service) CreateUpActReserveRelationLotteryNotify(ctx context.Context, newObj *l.Reserve) (err error) {
	if err = s.dao.SendLotteryNotify2Tunnel(ctx, &l.LotteryReserveNotify{
		BizID:        l.PlatformActivityBizID,
		UniqueID:     newObj.Sid,
		CardUniqueID: l.NotifyMessageTypeLotteryReserve,
		Mids:         []int64{newObj.Mid},
		State:        1,
		Timestamp:    function.Now(),
	}); err != nil {
		err = errors.Wrap(err, "s.dao.SendLotteryNotify2Tunnel err")
		return
	}
	return
}
