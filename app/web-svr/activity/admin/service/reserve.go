package service

import (
	"context"
	"encoding/json"
	"fmt"
	actplatadminapi "git.bilibili.co/bapis/bapis-go/platform/admin/act-plat"
	actapi "go-gateway/app/web-svr/activity/interface/api"
	"strconv"

	acccli "git.bilibili.co/bapis/bapis-go/account/service"
	ecodecommon "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model"
	"go-gateway/app/web-svr/activity/admin/model/reserve"
	"go-gateway/app/web-svr/activity/ecode"
	"strings"

	api "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"github.com/jinzhu/gorm"

	"go-common/library/sync/errgroup.v2"
)

// IsFollowReply ...
type IsFollowReply struct {
	IsFollow bool `json:"is_follow"`
}

// Following 是否预约
func (s *Service) Following(c context.Context, sid int64, mid int64) (reply *IsFollowReply, err error) {
	res, err := s.actClient.ReserveFollowing(c, &actapi.ReserveFollowingReq{
		Sid: sid,
		Mid: mid,
	})
	reply = &IsFollowReply{}
	if err != nil {
		log.Errorc(c, "s.actClient.ReserveFollowing err(%v)", err)
		return
	}
	reply.IsFollow = res.IsFollow
	return reply, nil
}

func (s *Service) ReserveList(c context.Context, arg *reserve.ParamList) (rly *reserve.SearchReply, err error) {
	var (
		list    *reserve.ListReply
		mids    []int64
		account *acccli.InfosReply
		stat    *model.SubjectStat
	)
	// 获取列表数据
	if list, err = s.dao.SearchReserve(c, arg.Sid, arg.Mid, arg.Pn, arg.Ps); err != nil {
		log.Errorc(c, "s.dao.SearchReserve(%d,%d) error(%v)", arg.Sid, arg.Mid, err)
		return
	}
	rly = &reserve.SearchReply{Ps: arg.Ps, Pn: arg.Pn}
	if list == nil || list.Count == 0 {
		return
	}
	rly.Count = list.Count
	for _, val := range list.List {
		if val.Mid == 0 {
			continue
		}
		mids = append(mids, val.Mid)
	}
	// 获取用户信息
	eg := errgroup.WithContext(c)
	if len(mids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if account, e = s.accClient.Infos3(ctx, &acccli.MidsReq{Mids: mids}); e != nil {
				log.Errorc(c, " s.accClient.Infos3(%v) error(%v)", mids, e)
				e = nil
			}
			return
		})
	}
	eg.Go(func(ctx context.Context) (e error) {
		if stat, e = s.dao.SubStat(ctx, arg.Sid); e != nil {
			log.Errorc(c, "s.dao.SubStat(%d) error(%v)", arg.Sid, e)
			e = nil
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if stat != nil {
		rly.ReserveNum = stat.Num
	}
	if account == nil {
		return
	}
	for _, val := range list.List {
		tep := &reserve.Item{ActReserve: val}
		if acc, ok := account.Infos[val.Mid]; ok {
			tep.Account = &reserve.AccInfo{Mid: val.Mid, Name: acc.Name, Face: acc.Face}
		}
		tep.TotalScore = val.Score + val.AdjustScore
		rly.List = append(rly.List, tep)
	}
	return
}

// AddReserve .
func (s *Service) AddReserve(c context.Context, arg *reserve.ParamAddReserve) (err error) {
	if err = s.dao.AddReserve(c, arg.Sid, arg.Mid, arg.Num); err != nil {
		log.Errorc(c, "s.dao.AddReserve(%d,%d) error(%v)", arg.Sid, arg.Mid)
	}
	return
}

func (s *Service) ImportReserve(c context.Context, arg *reserve.ParamImportReserve, mids []int64) (err error) {
	for _, mid := range mids {
		if err = s.dao.AddReserve(c, arg.Sid, mid, arg.Num); err != nil {
			log.Errorc(c, "s.dao.AddReserve(%d,%d) error(%v)", arg.Sid, mid)
			return
		} else if arg.Sid == s.c.Bws.ReserveSid {
			if err = s.dao.BwsReserveGift(c, mid); err != nil {
				log.Errorc(c, "s.dao.BwsReserveGift(%d) error(%v)", mid)
			}
		}
	}
	return
}

// ReserveScoreUpdate ...
func (s *Service) ReserveScoreUpdate(c context.Context, arg *reserve.ParamReserveScoreUpdate) (err error) {
	actSubject := new(model.ActSubject)
	if err = s.DB.Where("id = ?", arg.Sid).Last(actSubject).Error; err != nil {
		log.Errorc(c, "s.DB.Where(id = ?, %d).Last(%v) error(%v)", arg.Sid, actSubject, err)
		return
	}
	if actSubject.Type != model.USERACTIONSTAT {
		err = ecodecommon.Error(ecodecommon.RequestErr, "subject 类型不对")
		return
	}
	var reserve *reserve.ActReserve
	reserve, err = s.dao.GetReserve(c, arg.Sid, arg.Mid)
	if err != nil {
		return err
	}
	if reserve == nil {
		return ecode.ActivityReserveFirst
	}
	reserve.AdjustScore = arg.Score
	s.dao.UpdateReserve(c, reserve)
	return nil
}

func (s *Service) buildEmailList(notify *reserve.ActSubjectNotify) []string {
	m := make(map[string]struct{})
	for _, receiver := range s.c.Reserve.Notify {
		m[receiver] = struct{}{}
	}
	for _, receiver := range strings.Split(notify.Receiver, ",") {
		m[receiver] = struct{}{}
	}
	m[notify.Author] = struct{}{}
	list := make([]string, 0, len(m))
	for receiver := range m {
		if receiver == "" {
			continue
		}
		list = append(list, fmt.Sprintf("%s@bilibili.com", receiver))
	}
	return list
}

func (s *Service) buildFullNotifyReq(c context.Context, subject *model.ActSubject, notify *reserve.ActSubjectNotify) (*api.FullNotifierReq, error) {
	rule := &model.SubjectRule{}
	if err := s.dao.DB.Model(&model.SubjectRule{}).Where("id=?", notify.RuleID).First(&rule).Error; err != nil {
		log.Errorc(c, "s.dao.DB id:%d error(%v)", notify.RuleID, err)
		return nil, err
	}
	// todo 本应使用模板化方式，暂时无模板化需求，先偷懒下。后续以模板化方式支持
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
	content, _ := json.Marshal(map[string]interface{}{
		"value":       notify.Threshold,
		"notify_dest": "mail",
		"notify_topic": map[string]interface{}{
			"receivers": s.buildEmailList(notify),
			"subject":   fmt.Sprintf("【活动平台进度提醒】%v即将到达节点%s", name, notify.Title),
			"body": fmt.Sprintf("你的活动（%v）即将到达设置的节点%v，http://activity-template.bilibili.co/editDynamic/activity/%v",
				name, interveningThreshold, data["page_id"]),
		},
	})
	return &api.FullNotifierReq{
		Name:     fmt.Sprintf("act_notify_%d", notify.ID),
		Activity: fmt.Sprint(subject.ID),
		Counter:  "SUM_" + rule.RuleName,
		Type:     1,
		Content:  string(content),
	}, nil
}

func (s *Service) ReserveNotifyUpdate(c context.Context, sid int64, notifies []*reserve.ActSubjectNotify) (err error) {
	actSubject := new(model.ActSubject)
	if err = s.DB.Where("id = ?", sid).Last(actSubject).Error; err != nil {
		log.Errorc(c, "s.DB.Where(id = ?, %d).Last(%v) error(%v)", sid, actSubject, err)
		return
	}
	if actSubject.Type != model.USERACTIONSTAT && actSubject.Type != model.RESERVATION && actSubject.Type != model.CLOCKIN {
		err = ecodecommon.Error(ecodecommon.RequestErr, "subject 类型不对")
		return
	}
	if actSubject.Type != model.RESERVATION {
		for _, notify := range notifies {
			if notify.RuleID == 0 {
				err = ecodecommon.Error(ecodecommon.RequestErr, "rule_id 缺失")
				return
			}
		}
	}
	for _, notify := range notifies {
		// 检查是否存在
		obj := new(reserve.ActSubjectNotify)
		if err = s.DB.Where("notify_id = ? AND sid = ?", notify.NotifyID, sid).Last(&obj).Error; err != nil && err != gorm.ErrRecordNotFound {
			log.Errorc(c, "s.DB.Where(notify_id = ? AND sid = ?, %s).Last() error(%v)", notify.NotifyID, sid, err)
			return
		}
		if obj.ID > 0 {
			var notifyTime = obj.NotifyTime
			if notify.State == reserve.ActSubjectNotifyStateNormal && obj.NotifyTime > 0 && obj.Threshold != notify.Threshold {
				// 通知阈值修改，并且已通知过，重置通知标识
				notifyTime = 0
			}
			if err = s.DB.Model(&reserve.ActSubjectNotify{}).Where("id = ?", obj.ID).Update(map[string]interface{}{
				"rule_id":     notify.RuleID,
				"notify_type": notify.NotifyType,
				"title":       notify.Title,
				"receiver":    notify.Receiver,
				"state":       notify.State,
				"threshold":   notify.Threshold,
				"notify_time": notifyTime,
				"template_id": notify.TemplateID,
				"ext":         notify.Ext,
			}).Error; err != nil {
				log.Errorc(c, "s.DB.Where(id = ?, %s).Update(%v) error(%v)", obj.ID, notify, err)
				return
			}
			if actSubject.Type == model.USERACTIONSTAT {
				notify.ID = obj.ID
				req, _ := s.buildFullNotifyReq(c, actSubject, notify)
				if notify.State == reserve.ActSubjectNotifyStateNormal {
					_, err = s.actPlatClient.UpdateNotifier(c, req)
				} else {
					_, err = s.actPlatClient.CloseNotifier(c, &api.SimpleNotifierReq{
						Name:     req.Name,
						Activity: req.Activity,
						Counter:  req.Counter,
					})
				}
				if err != nil {
					log.Errorc(c, "s.actPlatClient.UpdateNotifier/CloseNotifier(%v) error(%v)", *req, err)
					return
				}
			}
		} else {
			if err = s.DB.Save(&notify).Error; err != nil {
				log.Errorc(c, "s.DB.Save(%v) error(%v)", notify, err)
				return
			}
			if actSubject.Type == model.USERACTIONSTAT {
				req, _ := s.buildFullNotifyReq(c, actSubject, notify)
				_, err = s.actPlatClient.AddNotifier(c, req)
				if err != nil {
					log.Errorc(c, "s.actPlatClient.AddNotifier(%v) error(%v)", *req, err)
					return
				}
			}
		}
	}
	return nil
}

func (s *Service) ReserveNotifyDelete(c context.Context, sid int64, author string, notifyID []string) (err error) {
	actSubject := new(model.ActSubject)
	if err = s.DB.Where("id = ?", sid).Last(actSubject).Error; err != nil {
		log.Errorc(c, "s.DB.Where(id = ?, %d).Last(%v) error(%v)", sid, actSubject, err)
		return
	}
	if actSubject.Type != model.USERACTIONSTAT && actSubject.Type != model.RESERVATION && actSubject.Type != model.CLOCKIN {
		err = ecodecommon.Error(ecodecommon.RequestErr, "subject 类型不对")
		return
	}
	if err = s.DB.Model(&reserve.ActSubjectNotify{}).Where("sid = ? AND notify_id IN (?)", sid, notifyID).
		Updates(map[string]interface{}{"state": reserve.ActSubjectNotifyStateDelete}).Error; err != nil {
		log.Errorc(c, "s.DB.Where(sid = ? AND notify_id IN (?), %d, %v).Updates() error(%v)", sid, notifyID, err)
	}
	if err == nil && actSubject.Type == model.USERACTIONSTAT {
		for _, id := range notifyID {
			notify := new(reserve.ActSubjectNotify)
			if err = s.DB.Where("notify_id = ? AND sid = ?", id, sid).Last(&notify).Error; err != nil && err != gorm.ErrRecordNotFound {
				log.Errorc(c, "s.DB.Where(notify_id = ? AND sid = ?, %s).Last() error(%v)", id, sid, err)
				return
			}
			rule := &model.SubjectRule{}
			if err := s.dao.DB.Model(&model.SubjectRule{}).Where("id=?", notify.RuleID).First(&rule).Error; err != nil {
				log.Errorc(c, "s.dao.DB id:%d error(%v)", notify.RuleID, err)
				return err
			}
			s.actPlatClient.CloseNotifier(c, &api.SimpleNotifierReq{
				Name:     fmt.Sprintf("act_notify_%d", notify.ID),
				Activity: fmt.Sprint(actSubject.ID),
				Counter:  "SUM_" + rule.RuleName,
			})
		}
	}
	return
}

func (s *Service) ReserveCounterGroupList(c context.Context, sid int64) (rly *reserve.GroupListReply, err error) {
	obj := make([]*reserve.GroupItem, 0, 10)
	if err = s.DB.Where("sid = ? AND state = ?", sid, reserve.CounterGroupStateNormal).Find(&obj).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, "s.DB.Where(sid = ?, %s).Find() error(%v)", sid, err)
		return
	}
	rly = &reserve.GroupListReply{
		Count: int64(len(obj)),
		Pn:    1,
		Ps:    len(obj),
		List:  obj,
	}
	return
}

func (s *Service) ReserveCounterGroupUpdate(c context.Context, obj *reserve.ParamCounterGroupUpdate) (res interface{}, err error) {
	actSubject := new(model.ActSubject)
	if err = s.DB.Where("id = ?", obj.Sid).Last(actSubject).Error; err != nil {
		log.Errorc(c, "s.DB.Where(id = ?, %d).Last(%v) error(%v)", obj.Sid, actSubject, err)
		return
	}
	if actSubject.Type != model.USERACTIONSTAT {
		err = ecodecommon.Error(ecodecommon.RequestErr, "subject 类型不对")
		return
	}
	pointReq := &actplatadminapi.PointsUnlockGroup{
		Notifiers: []*actplatadminapi.PointsUnlockGroupNotify{},
	}
	if err = json.Unmarshal([]byte(obj.CounterInfo), &pointReq); err != nil {
		log.Errorc(c, "json.Unmarshal(%s, &pointReq) error(%v)", obj.CounterInfo, err)
		return
	}
	extend := &reserve.Ext{}
	if err = json.Unmarshal([]byte(obj.Ext), &extend); err != nil {
		log.Errorc(c, "json.Unmarshal(%s, &extend) error(%v)", obj.Ext, err)
		return
	}

	return nil, s.DB.Transaction(func(tx *gorm.DB) (err error) {
		if obj.ID == 0 {
			// 先添加节点组
			obj.State = reserve.CounterGroupStateNormal
			if err = tx.Save(&obj.GroupItem).Error; err != nil {
				log.Errorc(c, "s.DB.Save(%v) error(%v)", obj.GroupItem, err)
				return
			}
		} else {
			group := reserve.GroupItem{}
			if err = tx.Where("id = ?", obj.ID).Last(&group).Error; err != nil {
				log.Errorc(c, "s.DB.Where(id = ?, %d).Last(%v) error(%v)", obj.ID, group, err)
				return
			}
			if group.Sid == 0 {
				err = ecodecommon.Error(ecodecommon.RequestErr, "group 不存在")
				return
			}
			if err = tx.Model(&reserve.GroupItem{}).Where("id = ?", obj.ID).Update(map[string]interface{}{
				"group_name":   obj.GroupName,
				"dim1":         obj.Dim1,
				"dim2":         obj.Dim2,
				"threshold":    obj.Threshold,
				"counter_info": obj.CounterInfo,
				"author":       obj.Author,
				"ext":          obj.Ext,
			}).Error; err != nil {
				log.Errorc(c, "s.DB.Where(id = ?, %s).Update(%v) error(%v)", obj.ID, obj.GroupItem, err)
				return
			}
			obj.Sid = group.Sid
		}
		for _, n := range obj.Nodes {
			n.Sid = obj.Sid
			n.GroupID = obj.ID
			if n.ID > 0 {
				// 检查是否存在
				nn := reserve.NodeItem{}
				if err = tx.Where("id = ? AND group_id = ?", n.ID, n.GroupID).Last(&nn).Error; err != nil {
					log.Errorc(c, "s.DB.Where(id = ? AND group_id = ?, %d).Last(%v) error(%v)", n.ID, n.GroupID, nn, err)
					return
				}
				if nn.Sid == 0 {
					err = ecodecommon.Error(ecodecommon.RequestErr, "node 不存在")
					return
				}
				if err = tx.Model(&reserve.NodeItem{}).Where("id = ? AND group_id = ?", n.ID, n.GroupID).Update(map[string]interface{}{
					"node_name": n.NodeName,
					"node_val":  n.NodeVal,
					"state":     n.State,
				}).Error; err != nil {
					log.Errorc(c, "s.DB.Where(id = ? AND group_id = ?, %d,%d).Update(%v) error(%v)", n.ID, n.GroupID, *n, err)
					return
				}
			} else {
				if err = tx.Save(n).Error; err != nil {
					log.Errorc(c, "s.DB.Save(%v) error(%v)", *n, err)
					return
				}
			}
			var needNotify bool
			//if obj.Dim1 == reserve.CounterGroupDim1Personal || obj.Threshold > 0 {
			//	needNotify = true
			//}
			if obj.Threshold > 0 {
				needNotify = true
			}
			if n.State == reserve.CounterNodeStateNormal && n.NodeVal > 0 && needNotify {
				var notifyType actplatadminapi.PointsUnlockGroupNotifyChannelType
				var value int64
				var channelConf = &actplatadminapi.PointsUnlockGroupNotifyChannelConfig{}
				notifyType = reserve.NotifyTypeEmail
				// databus通知notify
				//if obj.Dim1 == reserve.CounterGroupDim1Personal {
				//	notifyType = reserve.NotifyTypeDatabus
				//	value = n.NodeVal
				//}
				// 邮件通知
				if obj.Threshold > 0 {
					value = n.NodeVal * obj.Threshold / 100
					channelConf = &actplatadminapi.PointsUnlockGroupNotifyChannelConfig{
						Email: &actplatadminapi.PointsUnlockGroupNotifyEmailConfig{
							Receivers: s.buildEmailList(&reserve.ActSubjectNotify{
								Receiver: obj.Author,
							}),
							Subject: fmt.Sprintf("【活动平台进度提醒】%s即将到达节点%s", actSubject.Name, n.NodeName),
							Msg: fmt.Sprintf("你的活动（%s）即将到达设置的节点%s，http://activity-template.bilibili.co/source/reserve?keyword=%d",
								actSubject.Name, n.NodeName, actSubject.ID),
						},
					}
				}
				notifer := &actplatadminapi.PointsUnlockGroupNotify{
					Name:        fmt.Sprintf("%d_%d_%d", obj.Sid, obj.ID, n.ID),
					Type:        notifyType,
					Value:       value,
					ChannelConf: channelConf,
				}
				pointReq.Notifiers = append(pointReq.Notifiers, notifer)
			}
		}

		// 下游通知
		if extend.DownStream.Switch {
			var typ actplatadminapi.PointsUnlockGroupNotifyChannelType
			switch extend.DownStream.Type {
			case reserve.NotifyActionTypePointsUnlockGroupNotifyChannelTypeDatabus:
				typ = actplatadminapi.PointsUnlockGroupNotifyChannelTypeDatabus
			case reserve.NotifyActionTypePointsUnlockGroupNotifyTypeTotalDiffAndKeyMid:
				typ = actplatadminapi.PointsUnlockGroupNotifyTypeTotalDiffAndKeyMid
			case reserve.NotifyActionTypePointsUnlockGroupNotifyTypeEachLargerThan:
				typ = actplatadminapi.PointsUnlockGroupNotifyTypeEachLargerThan
			default:
				err = fmt.Errorf("extend.DownStream.Type extend(%+v)", extend)
				log.Errorc(c, err.Error())
				return
			}
			var val int64
			if typ == actplatadminapi.PointsUnlockGroupNotifyChannelTypeDatabus {
				// value 存在多个值
				vals := strings.Split(extend.DownStream.Value, ",")
				for _, v := range vals {
					val, _ = strconv.ParseInt(v, 10, 64)
					pointReq.Notifiers = append(pointReq.Notifiers, &actplatadminapi.PointsUnlockGroupNotify{
						Name:  fmt.Sprintf("%d_%d", obj.ID, val),
						Type:  typ,
						Value: val,
					})
				}
			} else {
				val, _ = strconv.ParseInt(extend.DownStream.Value, 10, 64)
				pointReq.Notifiers = append(pointReq.Notifiers, &actplatadminapi.PointsUnlockGroupNotify{
					Name:  fmt.Sprintf("%d_%d", obj.ID, val),
					Type:  typ,
					Value: val,
				})
			}
		}

		// 拼装数据通知任务系统
		pointReq.Activity = fmt.Sprint(obj.Sid)
		pointReq.GroupName = fmt.Sprint(obj.ID)
		pointReq.Dim01 = actplatadminapi.Dim01(obj.Dim1)
		pointReq.Dim02 = actplatadminapi.Dim02(obj.Dim2)
		_, err = s.platAdminClient.PointsUnlockUpdateGroup(c, pointReq)
		if err != nil {
			b, _ := json.Marshal(pointReq)
			log.Errorc(c, "s.platAdminClient.PointsUnlockUpdateGroup(c, %s) err[%v]", b, err)
		}
		return
	})
}

func (s *Service) ReserveCounterNodeList(c context.Context, groupID int64) (rly *reserve.NodeListReply, err error) {
	obj := make([]*reserve.NodeItem, 0, 10)
	if err = s.DB.Where("group_id = ? AND state = ?", groupID, reserve.CounterNodeStateNormal).Order("id asc").Find(&obj).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, "s.DB.Where(group_id = ?, %s).Find() error(%v)", groupID, err)
		return
	}
	rly = &reserve.NodeListReply{
		Count: int64(len(obj)),
		Pn:    1,
		Ps:    len(obj),
		List:  obj,
	}
	return
}
