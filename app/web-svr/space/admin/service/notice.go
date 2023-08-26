package service

import (
	"context"
	"fmt"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/queue/databus/actionlog"
	"go-gateway/app/web-svr/space/admin/model"
)

const (
	_noticeClearMc        = `1_26_2`
	_noticeClearForbidMc  = `1_26_3`
	_noticeClearTitle     = `空间公告清除通知`
	_noticeClearFmt       = `您好，您的空间公告，因%s的原因已被清除。请自觉遵守国家相关法律法规及《社区规则》，良好的社区氛围需要大家一起维护。`
	_noticeClearForbidFmt = `您好，您的空间公告，因%s的原因已被清除并隐藏。请自觉遵守国家相关法律法规及《社区规则》，良好的社区氛围需要大家一起维护。`
)

// Notice get notice data.
func (s *Service) Notice(c context.Context, mid int64) (data *model.Notice, err error) {
	data = &model.Notice{Mid: mid}
	if err = s.dao.DB.Table(data.TableName()).Where("mid=?", mid).First(&data).Error; err != nil {
		log.Error("Notice (mid:%d) error (%v)", mid, err)
		if err == ecode.NothingFound {
			err = nil
		}
	}
	return
}

// NoticeUp notice clear and forbid.
func (s *Service) NoticeUp(c context.Context, arg *model.NoticeUpArg) (err error) {
	var (
		action    string
		notifyMc  string
		notifyMsg string
	)
	notice := &model.Notice{Mid: arg.Mid}
	if err = s.dao.DB.Table(notice.TableName()).Where("mid=?", arg.Mid).First(&notice).Error; err != nil {
		log.Error("NoticeForbid error (mid:%d) (%v)", arg.Mid, err)
		if err != ecode.NothingFound {
			return
		}
	}
	up := make(map[string]interface{})
	switch arg.Type {
	case model.NoticeTypeClear:
		up["notice"] = ""
		action = model.NoticeClear
		notifyMc = _noticeClearMc
		notifyMsg = fmt.Sprintf(_noticeClearFmt, model.ClearMsgReasons[arg.Reason])
	case model.NoticeTypeClearAndForbid:
		up["notice"] = ""
		up["is_forbid"] = model.NoticeForbid
		action = model.NoticeClearAndForbid
		notifyMc = _noticeClearForbidMc
		notifyMsg = fmt.Sprintf(_noticeClearForbidFmt, model.ClearMsgReasons[arg.Reason])
	case model.NoticeTypeUnForbid:
		up["is_forbid"] = model.NoticeNoForbid
		action = model.NoticeUnForbid
	}
	if err != ecode.NothingFound {
		if err = s.dao.DB.Table(notice.TableName()).Where("id=?", notice.ID).Update(up).Error; err != nil {
			log.Error("NoticeForbid (mid:%d) update error (%v)", arg.Mid, err)
			return
		}
	} else {
		create := &model.Notice{Mid: arg.Mid}
		if arg.Type == model.NoticeTypeClearAndForbid {
			create.IsForbid = model.NoticeForbid
		}
		if err = s.dao.DB.Table(notice.TableName()).Create(create).Error; err != nil {
			log.Error("NoticeForbid (mid:%d) insert error (%v)", arg.Mid, err)
			return
		}
	}
	if err = actionlog.Manager(&actionlog.ManagerInfo{
		Uname:    arg.Uname,
		UID:      arg.UID,
		Business: model.NoticeLogID,
		Type:     0,
		Oid:      arg.Mid,
		Action:   action,
		Ctime:    time.Now(),
		Content: map[string]interface{}{
			"old":    notice,
			"reason": model.ClearMsgReasons[arg.Reason],
		},
	}); err != nil {
		return
	}
	if (arg.Type == model.NoticeTypeClear || arg.Type == model.NoticeTypeClearAndForbid) && arg.Reason != 0 {
		s.cache.Do(c, func(c context.Context) {
			//nolint:errcheck
			s.dao.SendSystemMessage(c, []int64{arg.Mid}, notifyMc, _noticeClearTitle, notifyMsg)
		})
	}
	return
}
