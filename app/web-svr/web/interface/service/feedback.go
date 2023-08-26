package service

import (
	"context"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/web/ecode"
	"go-gateway/app/web-svr/web/interface/model"
)

// Feedback web player feedback.
func (s *Service) Feedback(c context.Context, feedParams *model.Feedback) (err error) {
	var location string
	if feedParams.Other != "" {
		if location, err = s.upload(c, feedParams.Other); err != nil {
			log.Error("s.upload error(%v)", err)
			err = nil
		} else {
			feedParams.Content.URL = location
		}
	}
	err = s.dao.Feedback(c, feedParams)
	return
}

func (s *Service) upload(c context.Context, Other string) (location string, err error) {
	if len(Other) > s.c.Bfs.MaxFileSize {
		err = ecode.FeedbackBodyTooLarge
		return
	}
	location, err = s.dao.Upload(c, Other, time.Now().Unix())
	return
}
