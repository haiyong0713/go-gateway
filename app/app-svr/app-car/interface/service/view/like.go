package view

import (
	"context"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/view"
	"go-gateway/app/app-svr/archive/service/api"
	mainEcode "go-gateway/ecode"
	"go-gateway/pkg/riskcontrol"
	avecode "go-main/app/app-svr/archive/ecode"

	thumbup "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	tecode "git.bilibili.co/bapis/bapis-go/community/service/thumbup/ecode"
	"github.com/pkg/errors"
)

const (
	_businessLike = "archive"
)

func (s *Service) Like(c context.Context, mid int64, buvid, path, ua string, param *view.LikeParam) (err error) {
	var (
		a   *api.SimpleArc
		typ thumbup.Action
	)
	if a, err = s.arc.SimpleArc(c, param.Aid); err != nil {
		if errors.Cause(err) == ecode.NothingFound {
			err = avecode.ArchiveNotExist
		}
		return err
	}
	if param.Like == 0 {
		if !a.IsNormal() {
			return avecode.ArchiveNotExist
		}
		typ = thumbup.Action_ACTION_LIKE
		// 点赞前先判断风控
		tec := &view.SilverEventCtx{
			Action:     model.SilverActionLike,
			Aid:        a.Aid,
			UpID:       a.Mid,
			Mid:        mid,
			PubTime:    time.Unix(a.Pubdate, 0).Format("2006-01-02 15:04:05"),
			LikeSource: model.SilverSourceLike,
			Buvid:      buvid,
			Ip:         metadata.String(c, metadata.RemoteIP),
			Platform:   param.Platform,
			Ctime:      time.Now().Format("2006-01-02 15:04:05"),
			Api:        path,
			Origin:     param.AppKey,
			UserAgent:  ua,
			Build:      strconv.Itoa(param.Build),
			Token:      riskcontrol.ReportedLoginTokenFromCtx(c),
		}
		if s.silverDao.RuleCheck(c, tec, model.SilverSceneLike) {
			return mainEcode.SilverBulletLikeReject
		}
	} else if param.Like == 1 {
		typ = thumbup.Action_ACTION_CANCEL_LIKE
	}
	if _, err = s.thumbupDao.Like(c, mid, a.Mid, _businessLike, a.Aid, typ, true, param.MobiApp, param.Device, param.Platform); err != nil {
		if ecode.EqualError(tecode.ThumbupDupLikeErr, err) {
			log.Error("%+v", err)
			err = nil
		}
		return
	}
	return nil
}

func (s *Service) CommunityPGC(c context.Context, mid int64, buvid string, param *view.CommunityParam) (*view.Community, error) {
	pgcAv, err := s.bgm.AvInfo(c, int32(param.EpId))
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := &view.Community{}
	if mid > 0 || buvid != "" {
		likeState, err := s.thumbupDao.HasLike(c, mid, _businessLike, buvid, pgcAv.Aid)
		if err != nil {
			log.Error("%+v", err)
			return res, nil
		}
		if likeState == thumbup.State_STATE_LIKE {
			res.Like = 1
		}
	}
	return res, nil
}
