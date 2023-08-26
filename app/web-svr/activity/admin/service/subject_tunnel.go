package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	acccli "git.bilibili.co/bapis/bapis-go/account/service"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model"
	"go-gateway/app/web-svr/activity/admin/model/stime"

	tunnelmdl "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	"github.com/jinzhu/gorm"
)

const (
	_index        = 1
	_letter       = 2
	_dynamic      = 3
	_platform     = 3
	_eventAlready = 108009
	_noAddEvent   = 108007
)

func (s *Service) HaveTemplate(ctx context.Context, tp int64) (res []*model.PushTemplate, err error) {
	var current []string
	switch tp {
	case _index:
		current = s.c.TunnelPush.Index
	case _letter:
		current = s.c.TunnelPush.Letter
	case _dynamic:
		current = s.c.TunnelPush.Dynamic
	default:
		err = xecode.NothingFound
		return
	}
	if len(current) == 0 {
		err = xecode.NothingFound
		return
	}
	for _, tpValue := range current {
		var template *model.PushTemplate
		if err = json.Unmarshal([]byte(tpValue), &template); err != nil {
			log.Errorc(ctx, "HaveTemplate type(%d) json.Unmarshal(%+v)", tp, err)
			return
		}
		res = append(res, template)
	}
	return
}

func (s *Service) checkPushTime(isAdd bool, data *model.SubjectTunnelParam) error {
	pushStart := stime.FromString(data.PushStart)
	pushEnd := stime.FromString(data.PushEnd)
	if isAdd && pushStart.Time().Unix() < time.Now().Unix() {
		return fmt.Errorf("推送开始时间必须大于当前时间")
	}
	if pushStart.Time().Unix() > pushEnd.Time().Unix() {
		return fmt.Errorf("推送开始时间不能小于结束时间")
	}
	hours := pushEnd.Time().Sub(time.Unix(pushStart.Time().Unix(), 0)).Hours()
	if int(hours/24) > 20 {
		return fmt.Errorf("结束时间和开始时间的时间差距最大为20天")
	}
	return nil
}

func (s *Service) checkSubject(ctx context.Context, sid int64) (res *model.ActSubject, err error) {
	res = new(model.ActSubject)
	if err = s.DB.Where("id = ?", sid).Last(res).Error; err != nil {
		log.Errorc(ctx, "s.DB.Where(id = ?, %d).Last(%v) error(%v)", sid, res, err)
		return nil, fmt.Errorf("act_subject表查询(%d)出错", sid)
	}
	if res == nil || res.ID == 0 {
		return nil, fmt.Errorf("数据源sid(%d)不存在", sid)
	}
	yyType := reserveType()
	if _, ok := yyType[res.Type]; !ok {
		return nil, fmt.Errorf("数据源type(%d)不是预约活动", res.Type)
	}
	return
}

func reserveType() (res map[int]string) {
	res = make(map[int]string, 3)
	res[18] = "预约活动"
	res[22] = "预约打卡活动"
	res[23] = "预约积分活动"
	return res
}

func (s *Service) checkHaveTunnel(ctx context.Context, sid int64) (err error) {
	var res []*model.ActSubjectTunnel
	if err = s.DB.Where("sid = ?", sid).Where("template_id>0").Find(&res).Error; err != nil {
		log.Errorc(ctx, "s.DB.Where(id = ?, %d).Last(%v) error(%v)", sid, res, err)
		return fmt.Errorf("act_subject_tunnel表查询(%d)出错", sid)
	}
	if len(res) == 0 {
		return fmt.Errorf("推送模板sid(%d)不存在，不能推送", sid)
	}
	return
}

func (s *Service) StartPush(ctx context.Context, sid int64) (err error) {
	var (
		actSubject *model.ActSubject
	)
	if actSubject, err = s.checkSubject(ctx, sid); err != nil {
		return
	}
	if actSubject.IsPush == 1 {
		return fmt.Errorf("一次预约数据源只支持一次推送,请刷新页面")
	}
	if err = s.checkHaveTunnel(ctx, sid); err != nil {
		return err
	}
	// 激活
	if err = s.activeEvent(ctx, actSubject); err != nil {
		if xecode.Cause(err).Code() == _noAddEvent {
			// 注册事件
			if err = s.addEvent(ctx, actSubject); err != nil {
				return
			}
			// 添加事件后，重新激活
			if err = s.activeEvent(ctx, actSubject); err != nil {
				err = fmt.Errorf("激活事件出错(%+v)", err)
				return err
			}
		}
	}
	actSubject.IsPush = 1
	for i := 0; i < 3; i++ {
		if err = s.DB.Model(&model.ActSubject{}).Where("id = ?", sid).Update(actSubject).Error; err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		log.Errorc(ctx, "s.DB.Model(&model.ActSubject{}).Where(id = ?, %d).Update(%v) error(%v)", sid, actSubject, err)
		return fmt.Errorf("act_subject表更新数据库出错")
	}
	return
}

func (s *Service) activeEvent(ctx context.Context, actSubject *model.ActSubject) (err error) {
	// 激活事件
	activeArg := &tunnelmdl.ActiveEventReq{
		BizId:     s.c.TunnelPush.TunnelBizID,
		UniqueId:  actSubject.ID,
		Platform:  _platform,
		StartTime: actSubject.PushStart.Time().Format("2006-01-02 15:04:05"),
		EndTime:   actSubject.PushEnd.Time().Format("2006-01-02 15:04:05"),
	}
	if _, err = s.tunnelClient.ActiveEvent(ctx, activeArg); err != nil {
		log.Errorc(ctx, "s.tunnelClient.ActiveEvent error(%v)", err)
		return err
	}
	log.Errorc(ctx, "activeEvent success sid(%d)", actSubject.ID)
	return
}

func (s *Service) InfoPush(ctx context.Context, sid int64) (res interface{}, err error) {
	var (
		actSubject                               *model.ActSubject
		list                                     []*model.ActSubjectTunnel
		indexTunnel, letterTunnel, dynamicTunnel *model.SubjectTunnel
	)
	if actSubject, err = s.checkSubject(ctx, sid); err != nil {
		return
	}
	if err = s.DB.Where("sid = ?", sid).Find(&list).Error; err != nil {
		log.Errorc(ctx, " db.Model(&model.SubjectTunnel{}).Find() sid(%d) error(%v)", sid, err)
		return
	}
	if len(list) == 0 {
		res = struct{}{}
		return
	}
	for _, tunnel := range list {
		var (
			title   *model.TunnelTitle
			content *model.TunnelContent
		)
		if tunnel.Title == "" {
			title = &model.TunnelTitle{}
		} else {
			json.Unmarshal([]byte(tunnel.Title), &title)
		}
		if tunnel.Content == "" {
			content = &model.TunnelContent{}
		} else {
			json.Unmarshal([]byte(tunnel.Content), &content)
		}
		switch tunnel.Type {
		case _index:
			indexTunnel = &model.SubjectTunnel{
				ID:         tunnel.ID,
				Sid:        sid,
				Type:       tunnel.Type,
				TemplateID: tunnel.TemplateID,
				Titles:     title,
				Contents:   content,
				Icon:       tunnel.Icon,
				Link:       tunnel.Link,
				SenderUid:  tunnel.SenderUid,
			}
		case _letter:
			letterTunnel = &model.SubjectTunnel{
				ID:         tunnel.ID,
				Sid:        sid,
				Type:       tunnel.Type,
				TemplateID: tunnel.TemplateID,
				Titles:     title,
				Contents:   content,
				Icon:       tunnel.Icon,
				Link:       tunnel.Link,
				SenderUid:  tunnel.SenderUid,
			}
		case _dynamic:
			dynamicTunnel = &model.SubjectTunnel{
				ID:         tunnel.ID,
				Sid:        sid,
				Type:       tunnel.Type,
				TemplateID: tunnel.TemplateID,
				Titles:     title,
				Contents:   content,
				Icon:       tunnel.Icon,
				Link:       tunnel.Link,
				SenderUid:  tunnel.SenderUid,
			}
		default:
			log.Errorc(ctx, "InfoPush Type(%d) error", tunnel.Type)
		}
	}
	res = &model.TunnelInfo{
		Sid:       sid,
		IsPush:    actSubject.IsPush,
		Index:     indexTunnel,
		Letter:    letterTunnel,
		Dynamic:   dynamicTunnel,
		PushStart: actSubject.PushStart,
		PushEnd:   actSubject.PushEnd,
	}
	return
}

func (s *Service) AddPush(ctx context.Context, data *model.SubjectTunnelParam) (err error) {
	var (
		tunnels    []*model.SubjectTunnel
		actSubject *model.ActSubject
	)
	if err = s.checkPushTime(true, data); err != nil {
		return
	}
	if actSubject, err = s.checkSubject(ctx, data.Sid); err != nil {
		return
	}
	if tunnels, err = s.getAddTunnels(ctx, data, nil); err != nil {
		return
	}
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("AddPush s.dao.DB.Begin error(%v)", err)
		return fmt.Errorf("数据库begin事务出错")
	}
	sql, sqlParam := model.TunnelBatchAddSQL(tunnels)
	if err = tx.Model(&model.SubjectTunnel{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Errorc(ctx, "AddPush SubjectTunnel s.dao.DB.Model Create(%+v) error(%v)", sqlParam, err)
		tx.Rollback()
		return fmt.Errorf("SubjectTunnel 插入数据库出错")
	}
	actSubject.PushStart = stime.FromString(data.PushStart)
	actSubject.PushEnd = stime.FromString(data.PushEnd)
	if err = tx.Model(&model.ActSubject{}).Where("id = ?", data.Sid).Update(actSubject).Error; err != nil {
		log.Errorc(ctx, "EditPush s.DB.Model(&model.ActSubject{}).Where(id = ?, %d).Update(%v) error(%v)", data.Sid, actSubject, err)
		tx.Rollback()
		return fmt.Errorf("act_subject表更新数据库出错")
	}
	// 注册事件
	if err = s.addEvent(ctx, actSubject); err != nil {
		tx.Rollback()
		return
	}
	// 更新卡片
	if err = s.upsertCard(ctx, actSubject.ID, tunnels); err != nil {
		err = fmt.Errorf("更新卡片出错(%+v)", err)
		tx.Rollback()
		return
	}
	err = tx.Commit().Error
	if err != nil {
		err = fmt.Errorf("更新数据库出错")
	}
	return
}

func (s *Service) addEvent(ctx context.Context, actSubject *model.ActSubject) (err error) {
	// 注册事件
	var eventRly *tunnelmdl.AddEventReply
	eventArg := &tunnelmdl.AddEventReq{BizId: s.c.TunnelPush.TunnelBizID, UniqueId: actSubject.ID, Title: actSubject.Name, Platform: _platform}
	if eventRly, err = s.tunnelClient.AddEvent(ctx, eventArg); err != nil {
		if xecode.Cause(err).Code() == _eventAlready { // 事件已注册不用返回错误
			err = nil
		} else {
			log.Errorc(ctx, "s.tunnelClient.AddEvent error(%v)", err)
			return fmt.Errorf("注册事件出错(%+v)", err)
		}
	}
	if eventRly != nil {
		log.Errorc(ctx, "addEvent success event_id(%+v)", eventRly.EventId)
	}
	return
}

func (s *Service) tunnelMap(ctx context.Context, sid int64) (res map[int64]*model.ActSubjectTunnel, err error) {
	var list []*model.ActSubjectTunnel
	if err = s.DB.Where("sid = ?", sid).Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, " db.Model(&model.SubjectTunnel{}).Find() sid(%d) error(%v)", sid, err)
		return
	}
	res = make(map[int64]*model.ActSubjectTunnel, 3)
	for _, tunnel := range list {
		res[tunnel.Type] = tunnel
	}
	return
}
func (s *Service) EditPush(ctx context.Context, data *model.SubjectTunnelParam) (err error) {
	var (
		tunnels    []*model.SubjectTunnel
		actSubject *model.ActSubject
		tunnelMap  map[int64]*model.ActSubjectTunnel
	)
	if err = s.checkPushTime(false, data); err != nil {
		return
	}
	if actSubject, err = s.checkSubject(ctx, data.Sid); err != nil {
		return
	}
	if tunnelMap, err = s.tunnelMap(ctx, data.Sid); err != nil {
		return
	}
	if len(tunnelMap) == 0 {
		return fmt.Errorf("没有添加模板")
	}
	if tunnels, err = s.getAddTunnels(ctx, data, tunnelMap); err != nil {
		return
	}
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("EditPush s.dao.DB.Begin error(%v)", err)
		return fmt.Errorf("数据库begin事务出错")
	}
	sql, sqlParam := model.TunnelBatchEditSQL(tunnels)
	if err = tx.Model(&model.SubjectTunnel{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Errorc(ctx, "EditPush SubjectTunnel s.dao.DB.Model Update(%+v) error(%v)", sqlParam, err)
		tx.Rollback()
		return fmt.Errorf("SubjectTunnel 更新数据库出错")
	}
	actSubject.PushStart = stime.FromString(data.PushStart)
	actSubject.PushEnd = stime.FromString(data.PushEnd)
	if err = tx.Model(&model.ActSubject{}).Where("id = ?", data.Sid).Update(actSubject).Error; err != nil {
		log.Errorc(ctx, "EditPush s.DB.Model(&model.ActSubject{}).Where(id = ?, %d).Update(%v) error(%v)", data.Sid, actSubject, err)
		tx.Rollback()
		return fmt.Errorf("act_subject表更新数据库出错")
	}
	// 更新卡片
	if err = s.upsertCard(ctx, actSubject.ID, tunnels); err != nil {
		if xecode.Cause(err).Code() == _noAddEvent {
			// 注册事件
			if err = s.addEvent(ctx, actSubject); err != nil {
				tx.Rollback()
				return
			}
			// 更新卡片
			if err = s.upsertCard(ctx, actSubject.ID, tunnels); err != nil {
				err = fmt.Errorf("更新卡片出错(%+v)", err)
				tx.Rollback()
				return
			}
		} else {
			tx.Rollback()
			return
		}
	}
	err = tx.Commit().Error
	if err != nil {
		err = fmt.Errorf("更新数据库出错")
	}
	return
}

func (s *Service) getAddTunnels(ctx context.Context, data *model.SubjectTunnelParam, tunnelMap map[int64]*model.ActSubjectTunnel) (tunnels []*model.SubjectTunnel, err error) {
	var (
		indexTunnel, letterTunnel, dynamicTunnel *model.SubjectTunnel
		dbTunnel                                 *model.ActSubjectTunnel
		ok                                       bool
		accRly                                   *acccli.CardReply
	)
	isEdit := len(tunnelMap) > 0
	// index 赋值
	if isEdit {
		if dbTunnel, ok = tunnelMap[_index]; !ok || dbTunnel == nil || dbTunnel.ID == 0 {
			return nil, fmt.Errorf("index id 参数不存在")
		}
	}
	if data.Index == "" {
		if isEdit {
			tunnels = append(tunnels, &model.SubjectTunnel{
				ID: dbTunnel.ID,
			})
		} else {
			tunnels = append(tunnels, &model.SubjectTunnel{
				Sid:  data.Sid,
				Type: _index,
			})
		}
	} else {
		if err = json.Unmarshal([]byte(data.Index), &indexTunnel); err != nil {
			return nil, fmt.Errorf("index 参数解析出错")
		}
		if isEdit {
			indexTunnel.ID = dbTunnel.ID
		}
		indexTunnel.Sid = data.Sid
		indexTunnel.Type = _index
		tunnels = append(tunnels, indexTunnel)
	}
	// letter 赋值
	if isEdit {
		if dbTunnel, ok = tunnelMap[_letter]; !ok || dbTunnel == nil || dbTunnel.ID == 0 {
			return nil, fmt.Errorf("letter id 参数不存在")
		}
	}
	if data.Letter == "" {
		if isEdit {
			tunnels = append(tunnels, &model.SubjectTunnel{
				ID: dbTunnel.ID,
			})
		} else {
			tunnels = append(tunnels, &model.SubjectTunnel{
				Sid:  data.Sid,
				Type: _letter,
			})
		}
	} else {
		if err = json.Unmarshal([]byte(data.Letter), &letterTunnel); err != nil {
			return nil, fmt.Errorf("letter 参数解析出错")
		}
		if isEdit {
			letterTunnel.ID = dbTunnel.ID
		}
		letterTunnel.Sid = data.Sid
		letterTunnel.Type = _letter
		tunnels = append(tunnels, letterTunnel)
	}
	// dynamic 赋值
	if isEdit {
		if dbTunnel, ok = tunnelMap[_dynamic]; !ok || dbTunnel == nil || dbTunnel.ID == 0 {
			return nil, fmt.Errorf("dynamic id 参数不存在")
		}
	}
	if data.Dynamic == "" {
		if isEdit {
			tunnels = append(tunnels, &model.SubjectTunnel{
				ID: dbTunnel.ID,
			})
		} else {
			tunnels = append(tunnels, &model.SubjectTunnel{
				Sid:  data.Sid,
				Type: _dynamic,
			})
		}
	} else {
		if err = json.Unmarshal([]byte(data.Dynamic), &dynamicTunnel); err != nil {
			return nil, fmt.Errorf("dynamic 参数解析出错")
		}
		if accRly, err = s.accClient.Card3(ctx, &acccli.MidReq{Mid: dynamicTunnel.SenderUid}); err != nil {
			log.Error("s.accClient.Card3(%d) error(%v)", dynamicTunnel.SenderUid, err)
			return
		}
		if accRly.Card.Official.Role < 3 {
			log.Errorc(ctx, "s.accClient.Card3 mid(%d) role(%d)", dynamicTunnel.SenderUid, accRly.Card.Official.Role)
			return nil, fmt.Errorf("请填写蓝V认证的官号")
		}
		if isEdit {
			dynamicTunnel.ID = dbTunnel.ID
		}
		dynamicTunnel.Sid = data.Sid
		dynamicTunnel.Type = _dynamic
		tunnels = append(tunnels, dynamicTunnel)
	}
	return
}

func (s *Service) upsertCard(ctx context.Context, sid int64, tunnels []*model.SubjectTunnel) (err error) {
	var (
		aiCard *tunnelmdl.AiCommonCard
		dyCard *tunnelmdl.DynamicCommonCard
		msCard *tunnelmdl.MessageCard
	)
	for _, tunnel := range tunnels {
		// 获取模板
		params := getParams(tunnel.Titles, tunnel.Contents)
		switch tunnel.Type {
		case _index:
			aiCard = &tunnelmdl.AiCommonCard{
				TemplateId: tunnel.TemplateID,
				Link:       tunnel.Link,
				Icon:       tunnel.Icon,
				Params:     params,
			}
		case _letter:
			msCard = &tunnelmdl.MessageCard{
				TemplateId: tunnel.TemplateID,
				SenderUid:  tunnel.SenderUid,
				JumpUrl:    tunnel.Link,
				Params:     params,
			}
		case _dynamic:
			dyCard = &tunnelmdl.DynamicCommonCard{
				TemplateId: tunnel.TemplateID,
				Link:       tunnel.Link,
				Icon:       tunnel.Icon,
				SenderUid:  tunnel.SenderUid,
				TagName:    s.c.TunnelPush.DynamicCardTag,
				Params:     params,
			}
		}
	}
	arg := &tunnelmdl.UpsertCardReq{
		BizId:       s.c.TunnelPush.TunnelBizID,
		UniqueId:    sid,
		Platform:    _platform,
		AiCard:      aiCard,
		MessageCard: msCard,
		DynamicCard: dyCard,
	}
	if _, err = s.tunnelClient.UpsertCard(ctx, arg); err != nil {
		log.Errorc(ctx, "s.tunnelClient.UpsertCard arg(%+v) error(%v)", arg, err)
	} else {
		log.Errorc(ctx, "UpsertCard success sid(%d)", sid)
	}
	return
}

// 固定标题与内容最多配制5个变量
func getParams(titles *model.TunnelTitle, contents *model.TunnelContent) (params map[string]string) {
	// 如果模板有新增变量，代码需要更改
	params = make(map[string]string)
	if titles != nil {
		if titles.Title != "" {
			params["title"] = titles.Title
		}
		if titles.Title1 != "" {
			params["title1"] = titles.Title1
		}
		if titles.Title2 != "" {
			params["title2"] = titles.Title2
		}
		if titles.Title3 != "" {
			params["title3"] = titles.Title3
		}
		if titles.Title4 != "" {
			params["title4"] = titles.Title4
		}
		if titles.Title5 != "" {
			params["title5"] = titles.Title5
		}
	}
	if contents != nil {
		if contents.Content != "" {
			params["content"] = contents.Content
		}
		if contents.Content1 != "" {
			params["content1"] = contents.Content1
		}
		if contents.Content2 != "" {
			params["content2"] = contents.Content2
		}
		if contents.Content3 != "" {
			params["content3"] = contents.Content3
		}
		if contents.Content4 != "" {
			params["content4"] = contents.Content4
		}
		if contents.Content5 != "" {
			params["content5"] = contents.Content5
		}
	}
	return
}
