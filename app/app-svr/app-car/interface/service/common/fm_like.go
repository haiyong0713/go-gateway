package common

import (
	"context"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
	thumbup2 "go-gateway/app/app-svr/app-car/interface/model/thumbup"
	"go-gateway/app/app-svr/app-car/interface/model/view"
	avecode "go-gateway/app/app-svr/archive/ecode"
	arc "go-gateway/app/app-svr/archive/service/api"
	mainEcode "go-gateway/ecode"
	"go-gateway/pkg/riskcontrol"

	thumbup "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	tecode "git.bilibili.co/bapis/bapis-go/community/service/thumbup/ecode"
	"github.com/pkg/errors"
)

const (
	_actionLike       = 1
	_actionCancelLike = 2

	_businessLike = "archive"
)

func (s *Service) FmLike(c context.Context, param *fm_v2.FmLikeParam) error {
	var (
		a   *arc.SimpleArc
		typ thumbup.Action
		err error
	)
	if a, err = s.archiveDao.SimpleArc(c, param.Oid); err != nil {
		if errors.Cause(err) == ecode.NothingFound {
			err = avecode.ArchiveNotExist
		}
		return err
	}
	if param.Action == _actionLike {
		if !a.IsNormal() {
			return avecode.ArchiveNotExist
		}
		typ = thumbup.Action_ACTION_LIKE
		// 点赞前先判断风控
		tec := &view.SilverEventCtx{
			Action:     model.SilverActionLike,
			Aid:        a.Aid,
			UpID:       a.Mid,
			Mid:        param.Mid,
			PubTime:    time.Unix(a.Pubdate, 0).Format("2006-01-02 15:04:05"),
			LikeSource: model.SilverSourceLike,
			Buvid:      param.Buvid,
			Ip:         metadata.String(c, metadata.RemoteIP),
			Platform:   param.Platform,
			Ctime:      time.Now().Format("2006-01-02 15:04:05"),
			Api:        param.Path,
			Origin:     param.AppKey,
			UserAgent:  param.UA,
			Build:      strconv.Itoa(param.Build),
			Token:      riskcontrol.ReportedLoginTokenFromCtx(c),
		}
		if s.sbDao.RuleCheck(c, tec, model.SilverSceneLike) {
			return mainEcode.SilverBulletLikeReject
		}
	} else if param.Action == _actionCancelLike {
		typ = thumbup.Action_ACTION_CANCEL_LIKE
	}
	req := &thumbup2.LikeReq{
		DeviceInfo: param.DeviceInfo,
		Mid:        param.Mid,
		Buvid:      param.Buvid,
		UpMid:      a.Mid,
		Business:   _businessLike,
		MsgId:      a.Aid,
		Action:     typ,
		WithStat:   true,
	}
	if err = s.thumbupDao.LikeWithNotLogin(c, req); err != nil {
		if ecode.EqualError(tecode.ThumbupDupLikeErr, err) {
			log.Error("%+v", err)
			err = nil
		}
	}
	return err
}
