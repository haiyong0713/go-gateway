package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/space/interface/model"

	artmdlModel "git.bilibili.co/bapis/bapis-go/article/model"
	artmdl "git.bilibili.co/bapis/bapis-go/article/service"
)

var _emptyArticle = make([]*artmdlModel.Meta, 0)

// Article get articles by upMid.
func (s *Service) Article(c context.Context, mid int64, pn, ps, sort int32) (res *artmdl.UpArtMetasReply, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	if res, err = s.artClient.UpArtMetas(c, &artmdl.UpArtMetasReq{Mid: mid, Pn: pn, Ps: ps, Sort: sort, Ip: ip}); err != nil {
		log.Error("s.artClient.UpArtMetas(%d,%d,%d) error(%v)", mid, pn, ps, err)
		return
	}
	if res != nil && len(res.Articles) == 0 {
		res.Articles = _emptyArticle
	}
	return
}

// UpArtStat get up all article stat.
func (s *Service) UpArtStat(c context.Context, mid int64) (data *model.UpArtStat, err error) {
	addCache := true
	if data, err = s.dao.UpArtCache(c, mid); err != nil {
		addCache = false
	} else if data != nil {
		return
	}
	if data, err = s.dao.UpArtStat(c, mid); data != nil && addCache {
		s.cache.Do(c, func(c context.Context) {
			_ = s.dao.SetUpArtCache(c, mid, data)
		})
	}
	return
}
