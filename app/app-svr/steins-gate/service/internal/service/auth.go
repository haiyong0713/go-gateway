package service

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

func (s *Service) arcUpAuth(c context.Context, aid, mid int64) (view *model.VideoUpView, err error) {
	if view, err = s.arcDao.VideoUpView(c, aid); err != nil {
		log.Error("NodeInfoPreview s.dao.VideoUpView aid(%d) error(%v)", aid, err)
		return
	}
	if err = checkInteractiveArchive(mid, view.Archive.Mid, view.Archive.Attribute); err != nil { // user must be the author
		return
	}
	return
}

func (s *Service) videoUpAuth(c context.Context, aid, cid, mid int64) (err error) {
	var view *model.VideoUpView
	if view, err = s.arcUpAuth(c, aid, mid); err != nil {
		return
	}
	err = ecode.GraphCidNotDispatched // 稿件中找不到该cid或者找到后该cid状态不为分发完成则返回错误
	for _, v := range view.Videos {
		if v == nil {
			log.Error("VideoUpView Video Aid %d Cid %d Nil", aid, cid)
			continue
		}
		if v.Cid == cid && v.XcodeState == _videoDispatchFinish {
			err = nil
			return
		}
	}
	return

}
