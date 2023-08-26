package service

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/archive-honor/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

// Honor is get honor by  aid
func (s *Service) Honor(c context.Context, aid int64) (res []*api.Honor, err error) {
	honors, err := s.d.HonorsByAid(c, aid)
	if err != nil {
		log.Error("s.d.HonorsByAid aid(%d) err(%v)", aid, err)
		return
	}
	for _, o := range api.TypeOrder {
		h, ok := honors[o]
		if !ok {
			continue
		}
		h.Url = s.checkUrl(c, h)
		res = append(res, h)
	}
	return
}

// HonorUpdate update honor
func (s *Service) HonorUpdate(c context.Context, aid int64, typ int32, url, desc, naUrl string) {
	rows, err := s.d.HonorUpdate(c, aid, typ, url, desc, naUrl)
	if err != nil {
		log.Error("s.d.HonorUpdate aid(%d) type(%d) url(%s) desc(%s) naUrl(%s) err(%v)", aid, typ, url, desc, naUrl, err)
		rt := &api.RetryInfo{Action: api.ActionUpdate}
		rt.Data.Aid = aid
		rt.Data.Type = typ
		rt.Data.URL = url
		rt.Data.Desc = desc
		rt.Data.NaUrl = naUrl
		s.PushToRetryList(context.Background(), rt)
		return
	}
	//CloseHot 控制热门推送
	if !s.c.Custom.CloseHot && rows == 1 && typ == api.TypeHot { //发送热门消息，每个aid入选热门仅发送一次
		s.SendMsg(c, aid, url)
	}
}

// HonorDel del honor
func (s *Service) HonorDel(c context.Context, aid int64, typ int32) {
	if err := s.d.HonorDel(c, aid, typ); err != nil {
		log.Error("s.d.HonorDel aid(%d) type(%d) err(%v)", aid, typ, err)
		rt := &api.RetryInfo{Action: api.ActionDel}
		rt.Data.Aid = aid
		rt.Data.Type = typ
		s.PushToRetryList(context.Background(), rt)
	}
}

// Honors is multi get honors by aids
func (s *Service) Honors(c context.Context, aids []int64) (res map[int64]*api.HonorReply, err error) {
	honors, err := s.d.HonorsByAids(c, aids)
	if err != nil {
		log.Error("s.d.HonorsByAids aids(%v) err(%v)", aids, err)
		return
	}
	res = make(map[int64]*api.HonorReply)
	for aid, h := range honors {
		tmpHonor := new(api.HonorReply)
		for _, o := range api.TypeOrder {
			tmph, ok := h[o]
			if !ok {
				continue
			}
			tmph.Url = s.checkUrl(c, tmph)
			tmpHonor.Honor = append(tmpHonor.Honor, tmph)
		}
		if len(tmpHonor.Honor) > 0 {
			res[aid] = tmpHonor
		}
	}
	return
}

func (s *Service) checkUrl(c context.Context, h *api.Honor) string {
	if h.Type != api.TypeWeeklySelection || h.NaUrl == "" {
		return h.Url
	}
	if feature.GetBuildLimit(c, "service.archive.honor", nil) {
		return h.NaUrl
	}
	return h.Url
}
