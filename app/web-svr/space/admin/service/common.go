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
	_clearMc         = `1_26_1`
	_clearTitle      = `空间公告清除通知`
	_clearFmt        = `您好，你的%s已被清除，清除理由是%s`
	_clearTopArcFmt  = `AV%d的置顶理由`
	_clearMpFmt      = `AV%d的代表作理由`
	_clearChNameFmt  = `%s频道的标题`
	_clearChIntroFmt = `%s频道的简介`
)

func (s *Service) ClearMsg(c context.Context, typ, reason int, mid, id, uid int64, uname string) (err error) {
	if (typ == model.ClearTypeMp || typ == model.ClearTypeChName || typ == model.ClearTypeChIntro) && id <= 0 {
		err = ecode.RequestErr
		return
	}
	reasonStr, ok := model.ClearMsgReasons[reason]
	if !ok {
		err = ecode.RequestErr
		return
	}
	action := ""
	var (
		oldTopArc  *model.TopArc
		oldMp      *model.Masterpiece
		oldChannel *model.Channel
		notifyMsg  string
		oldMsg     string
	)
	switch typ {
	case model.ClearTypeArc:
		action = "clear_toparc"
		oldTopArc, err = s.TopArcClear(c, mid)
	case model.ClearTypeMp:
		action = "clear_masterpiece"
		oldMp, err = s.MasterpieceClear(c, mid, id)
	case model.ClearTypeChName:
		action = "clear_channel_name"
		oldChannel, err = s.ChannelClear(c, mid, id, _channelName)
	case model.ClearTypeChIntro:
		action = "clear_channel_intro"
		oldChannel, err = s.ChannelClear(c, mid, id, _channelIntro)
	default:
		err = ecode.RequestErr
	}
	if err != nil {
		return
	}
	switch typ {
	case model.ClearTypeArc:
		notifyMsg = fmt.Sprintf(_clearFmt, fmt.Sprintf(_clearTopArcFmt, oldTopArc.Aid), reasonStr)
		oldMsg = oldTopArc.RecommendReason
	case model.ClearTypeMp:
		notifyMsg = fmt.Sprintf(_clearFmt, fmt.Sprintf(_clearMpFmt, oldMp.Aid), reasonStr)
		oldMsg = oldMp.RecommendReason
	case model.ClearTypeChName:
		notifyMsg = fmt.Sprintf(_clearFmt, fmt.Sprintf(_clearChNameFmt, oldChannel.Name), reasonStr)
		oldMsg = oldChannel.Name
	case model.ClearTypeChIntro:
		notifyMsg = fmt.Sprintf(_clearFmt, fmt.Sprintf(_clearChIntroFmt, oldChannel.Name), reasonStr)
		oldMsg = oldChannel.Intro
	default:
	}
	// add admin log
	managerInfo := &actionlog.ManagerInfo{
		Uname:    uname,
		UID:      uid,
		Business: model.NoticeLogID,
		Type:     model.ManagerLogType[typ],
		Oid:      mid,
		Action:   action,
		Ctime:    time.Now(),
		Index:    []interface{}{id},
		Content: map[string]interface{}{
			"reason": reasonStr,
			"old":    oldMsg,
		},
	}
	if err = actionlog.Manager(managerInfo); err != nil {
		log.Error("ClearMsg report.Manager(%+v) error(%v)", managerInfo, err)
		err = nil
	}
	// send notify
	s.cache.Do(c, func(c context.Context) {
		//nolint:errcheck
		s.dao.ClearCache(c, mid, id, typ)
		//nolint:errcheck
		s.dao.SendSystemMessage(c, []int64{mid}, _clearMc, _clearTitle, notifyMsg)
	})
	return
}

func (s *Service) AddLog(name string, uid int64, oid int64, action string, obj interface{}) {
	//nolint:errcheck
	actionlog.Manager(&actionlog.ManagerInfo{
		Uname:    name,
		UID:      uid,
		Business: model.NoticeLogID,
		Type:     model.LogExamine,
		Oid:      oid,
		Action:   action,
		Ctime:    time.Now(),
		Content: map[string]interface{}{
			"json": obj,
		},
	})
}
