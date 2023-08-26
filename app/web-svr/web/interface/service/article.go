package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
)

const (
	_sortNew  = 1
	_firstPn  = 1
	_samplePn = 1
	_samplePs = 1
)

var (
	_emptyArticleList = make([]*model.Meta, 0)
	_emptyAuthorList  = make([]*model.Info, 0)
	_emptyArtMetas    = make([]*artmdl.Meta, 0)
)

// ArticleList get article list.
func (s *Service) ArticleList(c context.Context, rid, mid int64, sort, pn, ps int32, aids []int64) ([]*model.Meta, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &artapi.RecommendsReq{Aids: aids, Cid: rid, Pn: pn, Ps: ps, Sort: sort, Ip: ip}
	artReply, err := s.artGRPC.Recommends(c, arg)
	if err != nil {
		log.Error("s.artGRPC.Recommends(%d,%d,%d,%d) error(%v)", rid, pn, ps, sort, err)
		return nil, err
	}
	artMetas := artReply.Res
	if len(artMetas) == 0 {
		return _emptyArticleList, nil
	}
	var likes *artapi.HadLikesByMidReply
	if mid > 0 {
		func() {
			var (
				likeAids []int64
				likeErr  error
			)
			for _, art := range artMetas {
				if art != nil && art.Meta != nil {
					likeAids = append(likeAids, art.Meta.ID)
				}
			}
			likes, likeErr = s.artGRPC.HadLikesByMid(c, &artapi.HadLikesByMidReq{Mid: mid, Aids: likeAids, Ip: ip})
			if likeErr != nil {
				log.Error("s.artGRPC.HadLikesByMid(%d,%v) error(%v)", mid, likeAids, err)
			}
		}()
	}
	var res []*model.Meta
	for _, art := range artMetas {
		if art != nil && art.Meta != nil {
			tmp := &model.Meta{Meta: art.Meta}
			if like, ok := likes.GetRes()[art.Meta.ID]; ok {
				tmp.Like = like
			}
			res = append(res, tmp)
		}
	}
	return res, nil
}

// ArticleUpList get article up list.
func (s *Service) ArticleUpList(c context.Context, mid int64) (res []*model.Info, err error) {
	if res, err = s.articleUps(c, mid); err != nil {
		err = nil
	} else if len(res) > 0 {
		if err := s.cache.Do(c, func(c context.Context) {
			if err := s.dao.SetArticleUpListCache(c, res); err != nil {
				log.Error("%+v", err)
			}
		}); err != nil {
			log.Error("%+v", err)
		}
		return
	}
	res, err = s.dao.ArticleUpListCache(c)
	if len(res) == 0 {
		res = _emptyAuthorList
	}
	return
}

// Categories get article categories list
func (s *Service) Categories(c context.Context) (res []*artmdl.Category, err error) {
	var reply *artapi.CategoriesReply
	ip := metadata.String(c, metadata.RemoteIP)
	if reply, err = s.artGRPC.Categories(c, &artapi.CategoriesReq{Ip: ip}); err != nil {
		log.Error("s.artGRPC.Categories error(%v)", err)
	} else {
		res = reply.Res
	}
	return
}

func (s *Service) articleUps(c context.Context, mid int64) (res []*model.Info, err error) {
	var (
		mids       []int64
		reply      *artapi.RecommendsReply
		cardsReply *accmdl.CardsReply
		relaReply  *accmdl.RelationsReply
		ip         = metadata.String(c, metadata.RemoteIP)
	)
	res = make([]*model.Info, 0)
	arg := &artapi.RecommendsReq{Sort: _sortNew, Pn: 1, Ps: s.c.Rule.ArtUpListGetCnt, Ip: ip}
	if reply, err = s.artGRPC.Recommends(c, arg); err != nil {
		log.Error("s.art.Recommends() error(%v)", err)
		return
	}
	if len(reply.Res) == 0 {
		log.Warn("s.art.Recommends warn len list cnt 0")
		return
	}
	listMap := make(map[int64]*artmdl.Meta, s.c.Rule.ArtUpListCnt)
	for _, v := range reply.Res {
		if len(listMap) == s.c.Rule.ArtUpListCnt {
			break
		}
		if _, ok := listMap[v.Meta.Author.Mid]; ok {
			continue
		}
		listMap[v.Meta.Author.Mid] = v.Meta
		mids = append(mids, v.Meta.Author.Mid)
	}
	if cardsReply, err = s.accGRPC.Cards3(c, &accmdl.MidsReq{Mids: mids}); err != nil {
		log.Error("s.accGRPC.Cards3(%v) error(%v)", mids, err)
		return
	}
	if mid > 0 {
		if relaReply, err = s.accGRPC.Relations3(c, &accmdl.RelationsReq{Mid: mid, Owners: mids, RealIp: ip}); err != nil {
			log.Error("s.accGRPC.Relations3(%d,%v) error(%v)", mid, mids, err)
			err = nil
		}
	}
	for _, mid := range mids {
		if card, ok := cardsReply.Cards[mid]; ok {
			info := &model.Info{ID: listMap[mid].ID, Title: listMap[mid].Title, PublishTime: listMap[mid].PublishTime}
			info.FromCard(card)
			if relaReply != nil {
				if relation, ok := relaReply.Relations[mid]; ok {
					info.Following = relation.Following
				}
			}
			res = append(res, info)
		}
	}
	return
}

// NewCount get new publish article count
func (s *Service) NewCount(c context.Context, pubTime int64) (count int64, err error) {
	var reply *artapi.NewArticleCountReply
	ip := metadata.String(c, metadata.RemoteIP)
	if reply, err = s.artGRPC.NewArticleCount(c, &artapi.NewArticleCountReq{Ptime: pubTime, Ip: ip}); err != nil {
		log.Error("s.art.NewArticleCount(%d) error(%v)", pubTime, err)
		err = nil
		return
	}
	count = reply.Res
	return
}

// UpMoreArts get up more articles
func (s *Service) UpMoreArts(c context.Context, aid int64) (res []*artmdl.Meta, err error) {
	var reply *artapi.UpMoreArtsReply
	ip := metadata.String(c, metadata.RemoteIP)
	if reply, err = s.artGRPC.UpMoreArts(c, &artapi.UpMoreArtsReq{Aid: aid, Ip: ip}); err != nil {
		log.Error("s.art.UpMoreArts(%d) error(%v)", aid, err)
		return
	}
	res = reply.Res
	if len(reply.Res) == 0 {
		res = _emptyArtMetas
	}
	return
}
