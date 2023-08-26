package like

import (
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/dao/like"
	mdlLike "go-gateway/app/web-svr/activity/interface/model/like"

	"go-gateway/app/web-svr/activity/ecode"

	"context"
)

const (
	pn          = 1
	ps          = 3
	maxShow     = 1000000
	showListLen = 3
)

// FestivalProcess get festival process
func (s *Service) FestivalProcess(c context.Context, mid int64) (*mdlLike.FestivalProcessReply, error) {
	var (
		subject  *mdlLike.SubjectItem
		likeList []*mdlLike.List
	)
	res := mdlLike.FestivalProcessReply{}
	imgList := make([]mdlLike.ImageList, showListLen)
	sid := s.c.SongFestival.Sid
	subject, err := s.dao.ActSubject(c, sid)
	if err != nil {
		res.ImageList = imgList
		return &res, ecode.ActivityHasOffLine
	}
	if subject.ID == 0 {
		res.ImageList = imgList
		return &res, ecode.ActivityHasOffLine
	}
	likeList, err = s.orderByCtime(c, sid, pn, ps, subject, 0)
	if err != nil {
		log.Error("s.orderByCtime(%d) error(%v)", sid, err)
		res.ImageList = imgList
		return &res, ecode.ActivityHasOffLine
	}
	err = s.getContent(c, likeList, subject, subject.Type, mid, like.ActOrderCtime)
	if err != nil {
		res.ImageList = imgList
		return &res, err
	}
	for i, v := range likeList {
		if v.Like > maxShow {
			object, ok := v.Object.(map[string]interface{})
			if ok {
				if cont, contOk := object["cont"]; contOk {
					likeContent := cont.(*mdlLike.LikeContent)
					imgList[showListLen-i-1].Img = likeContent.Image
					continue
				}
			}
			log.Error("v.Object get image error(%v)", object)
		}
	}
	res.ImageList = imgList
	return &res, nil
}
