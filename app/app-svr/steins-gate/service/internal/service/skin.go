package service

import (
	"context"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

// SkinList .
func (s *Service) SkinList(c context.Context) (list []*model.Skin) {
	list = s.skinList
	if len(list) == 0 {
		list = make([]*model.Skin, 0)
	}
	return
}

// loadSkinList is
func (s *Service) loadSkinList() {
	list, err := s.dao.RawSkinList(context.Background())
	if err != nil {
		log.Error("s.dao.RawSkinList err(%+v)", err)
		return
	}
	s.skinList = list
}

// loadSkinListProc is
func (s *Service) loadSkinListproc() {
	for {
		time.Sleep(time.Duration(s.c.Interval.SkinInterval))
		s.loadSkinList()
	}
}

func (s *Service) skinInfo(graph *api.GraphInfo, node *api.GraphNode) (skin *api.Skin) {
	var skinID int64
	if node != nil {
		skinID = node.SkinId
	}
	if skinID == 0 {
		skinID = graph.SkinId
	}
	for _, v := range s.skinList {
		if v.ID == skinID {
			skin = &api.Skin{
				ChoiceImage:            v.Image,
				TitleTextColor:         v.TitleTextColor,
				TitleShadowColor:       v.TitleShadowColor,
				TitleShadowOffsetX:     v.TitleShadowOffsetX,
				TitleShadowOffsetY:     v.TitleShadowOffsetY,
				TitleShadowRadius:      v.TitleShadowRadius,
				ProgressbarColor:       v.ProgressBarColor,
				ProgressbarShadowColor: v.ProgressShadowColor,
			}
			break
		}
	}
	if skin == nil {
		skin = s.c.DefaultSkin
	}
	return

}
