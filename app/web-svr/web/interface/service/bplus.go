package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	mdl "go-gateway/app/web-svr/web/interface/model"
	"go-gateway/pkg/idsafe/bvid"

	"go-common/library/sync/errgroup.v2"

	articlemdl "git.bilibili.co/bapis/bapis-go/article/model"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
	pgcShareGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/share"
)

// nolint: gocognit
func (s *Service) MaterialInfo(c context.Context, req *mdl.MaterialInfoReq) (*mdl.Dynamic, error) {
	var (
		articleIDs, archiveIDs []int64
		eps                    []int32
		articleTmp             = make(map[int64]struct{})
		archiveTmp             = make(map[int64]struct{})
		epm                    = make(map[int32]struct{})
	)
	for _, aid := range req.Aids {
		if aid == 0 {
			continue
		}
		archiveTmp[aid] = struct{}{}
	}
	for _, bid := range req.Bvids {
		var bvidTmp int64
		if bvidTmp, _ = bvid.BvToAv(bid); bvidTmp == 0 {
			continue
		}
		archiveTmp[bvidTmp] = struct{}{}
	}
	for aid := range archiveTmp {
		archiveIDs = append(archiveIDs, aid)
	}
	for _, v := range req.ArticleIDs {
		if v == 0 {
			continue
		}
		articleTmp[v] = struct{}{}
	}
	for id := range articleTmp {
		articleIDs = append(articleIDs, id)
	}
	for _, id := range req.EpIDs {
		epm[id] = struct{}{}
	}
	for id := range epm {
		eps = append(eps, id)
	}
	if len(archiveIDs) > 50 || len(articleIDs) > 50 || len(eps) > 50 {
		log.Errorc(c, "FeedInfo too many ids. archiveIDs(%+v), articleIDs(%+v)", archiveIDs, articleIDs)
		return nil, ecode.RequestErr
	}
	if len(archiveIDs) == 0 && len(articleIDs) == 0 && len(eps) == 0 {
		log.Errorc(c, "FeedInfo params is empty.")
		return nil, ecode.RequestErr
	}
	eg := errgroup.WithContext(c)
	var (
		arcm map[int64]*archivegrpc.Arc
		artm map[int64]*articlemdl.Meta
		em   []*pgcShareGrpc.ShareMessageResBody
	)
	if len(archiveIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			arcmTmp, err := s.arcGRPC.Arcs(ctx, &archivegrpc.ArcsRequest{Aids: archiveIDs})
			if err != nil {
				log.Error("MaterialInfo s.arcGRPC.Arcs(ids:%+v) failed. error(%v)", archiveIDs, err)
				return err
			}
			arcm = arcmTmp.GetArcs()
			return nil
		})
	}
	if len(articleIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			artRes, err := s.artGRPC.ArticleMetasMc(ctx, &articlegrpc.ArticleMetasReq{Ids: articleIDs})
			if err != nil {
				log.Errorc(c, "MaterialInfo s.artGRPC.ArticleMatasMc(ids:%+v) failed. error(%+v)", articleIDs, err)
				return err
			}
			artm = artRes.GetRes()
			return nil
		})
	}
	if len(eps) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if em, err = s.dao.ShareMessage(ctx, eps); err != nil {
				log.Error("%v", err)
				return err
			}
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	rsp := &mdl.Dynamic{}
	if arcm != nil {
		var archives []*mdl.Archive
		for _, item := range arcm {
			if item == nil {
				continue
			}
			a := &mdl.Archive{}
			a.FormArc(item)
			a.BVID = s.avToBv(a.AID)
			archives = append(archives, a)
		}
		rsp.Archive = archives
	}
	if artm != nil {
		var articles []*mdl.Article
		for _, item := range artm {
			if item == nil {
				continue
			}
			a := &mdl.Article{}
			a.FromArt(item)
			articles = append(articles, a)
		}
		rsp.Article = articles
	}
	for _, item := range em {
		if item == nil {
			continue
		}
		ep := &mdl.PGCShare{}
		ep.FromPgcShare(item)
		rsp.PGC = append(rsp.PGC, ep)
	}
	return rsp, nil
}
