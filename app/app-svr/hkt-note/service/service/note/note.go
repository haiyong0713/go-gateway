package note

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/article"
	"go-gateway/app/app-svr/hkt-note/service/model/note"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
	cssngrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
)

func (s *Service) NoteListInArc(c context.Context, req *api.NoteListInArcReq) (*api.NoteListInArcReply, error) {
	res, err := s.dao.NoteAid(c, req)
	if err != nil {
		log.Error("NoteError NoteListInArc err(%+v)", err)
		return nil, err
	}
	return &api.NoteListInArcReply{NoteIds: res}, nil
}

func (s *Service) NoteCount(c context.Context, req *api.NoteCountReq) (*api.NoteCountReply, error) {
	noteUser, err := s.dao.NoteUser(c, req.Mid)
	if err != nil {
		log.Error("NoteError NoteCount err(%+v)", err)
		return nil, err
	}
	if noteUser.NoteCount == 0 {
		return &api.NoteCountReply{}, nil
	}
	// 获取所有note_id，计算各种类个数
	listKeys, err := s.dao.NoteList(c, req.Mid, 0, -1, noteUser.NoteCount)
	if err != nil {
		log.Error("NoteError NoteCount err(%+v)", err)
		return nil, err
	}
	// 拆解zset value
	_, noteIds := note.ToNtKeys(req.Mid, listKeys)
	// 批量获取笔记详情
	cacheDetails, err := s.dao.NoteDetails(c, noteIds, req.Mid)
	if err != nil {
		log.Error("noteError NoteList mid(%d) noteIds(%v) err(%+v)", req.Mid, noteIds, err)
		return nil, err
	}
	aids, sids := note.ToVideoIds(cacheDetails)
	res := &api.NoteCountReply{
		NoteCount:   noteUser.NoteCount,
		FromArchive: int64(len(aids)),
		FromCheese:  int64(len(sids)),
	}
	return res, nil
}

func (s *Service) NoteSize(c context.Context, req *api.NoteSizeReq) (*api.NoteSizeReply, error) {
	eg := errgroup.WithContext(c)
	var (
		userSize int64
		noteSize int64
	)
	if req.NoteId != 0 {
		eg.Go(func(c context.Context) error {
			cacheDetail, err := s.dao.NoteDetail(c, req.NoteId, req.Mid)
			if err != nil {
				return err
			}
			noteSize = cacheDetail.NoteSize
			return nil
		})
	}
	eg.Go(func(c context.Context) error {
		cacheUser, err := s.dao.NoteUser(c, req.Mid)
		if err != nil {
			return err
		}
		userSize = cacheUser.NoteSize
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("NoteError NoteSize err(%+v)", err)
		return nil, err
	}
	return &api.NoteSizeReply{
		TotalSize:  userSize,
		SingleSize: noteSize,
	}, nil
}

func (s *Service) NoteInfo(c context.Context, req *api.NoteInfoReq) (*api.NoteInfoReply, error) {
	eg := errgroup.WithContext(c)
	var (
		cacheDetail  *note.DtlCache
		cacheContent *note.ContCache
		artDetail    *article.ArtDtlCache
	)
	eg.Go(func(c context.Context) (err error) {
		if cacheDetail, err = s.dao.NoteDetail(c, req.NoteId, req.Mid); err != nil {
			return err
		}
		if cacheDetail.NoteId == -1 {
			return xecode.NoteDetailNotFound
		}
		if cacheDetail.Mid != req.Mid {
			return xecode.NoteUserUnfit
		}
		return nil
	})
	eg.Go(func(c context.Context) (err error) {
		if cacheContent, err = s.dao.NoteContent(c, req.NoteId); err != nil {
			return err
		}
		if cacheContent.NoteId == -1 {
			return xecode.NoteContentNotFound
		}
		return nil
	})
	eg.Go(func(c context.Context) error {
		var e error
		artDetail, e = s.artDao.ArtDetail(c, req.NoteId, article.TpArtDetailNoteId)
		return e
	})
	if err := eg.Wait(); err != nil {
		log.Error("noteError NoteInfo req(%+v) err(%+v)", req, err)
		return nil, err
	}
	var (
		eg2         = errgroup.WithContext(c)
		tags        []*api.NoteTag
		cidCnt      int64
		auditReason string
	)
	eg2.Go(func(c context.Context) error {
		var e error
		tags, cidCnt, e = s.dao.ToTags(c, cacheDetail.Oid, cacheDetail.NoteId, cacheContent.Tag, cacheDetail.OidType)
		return e
	})
	if artDetail != nil && artDetail.Cvid > 0 {
		eg2.Go(func(c context.Context) error {
			audit, e := s.artDao.ArticleAudits(c, []int64{artDetail.Cvid})
			if e != nil {
				log.Warn("artWarn NoteInfo req(%+v) err(%+v)", req, e)
				return nil
			}
			if _, ok := audit[artDetail.Cvid]; !ok {
				log.Warn("artWarn NoteInfo req(%+v) ArticleAudits cvid(%d) nil", req, artDetail.Cvid)
				return nil
			}
			auditReason = audit[artDetail.Cvid].Reason
			return nil
		})
	}
	if err := eg2.Wait(); err != nil {
		log.Error("noteError NoteInfo req(%+v) err(%+v)", req, err)
		return nil, err
	}
	pubStatus, pubVersion, pubReason := artDetail.ToPubInfo(auditReason)
	res := &api.NoteInfoReply{
		Title:       cacheDetail.Title,
		Summary:     cacheDetail.Summary,
		Content:     cacheContent.Content,
		CidCount:    cidCnt,
		AuditStatus: int64(cacheDetail.AuditStatus),
		Tags:        tags,
		Oid:         cacheDetail.Oid,
		PubStatus:   pubStatus,
		PubReason:   pubReason,
		PubVersion:  pubVersion,
	}
	return res, nil
}

func (s *Service) NoteList(c context.Context, req *api.NoteListReq) (*api.NoteListReply, error) {
	// 用户笔记总数
	noteUser, err := s.dao.NoteUser(c, req.Mid)
	if err != nil {
		log.Error("noteError NoteList req(%+v) err(%+v)", req, err)
		return nil, err
	}
	page := &api.Page{
		Total: noteUser.NoteCount,
		Size_: req.Ps,
		Num:   req.Pn,
	}
	if page.Total == 0 {
		return &api.NoteListReply{Page: page}, nil
	}
	min, max := note.ToPage(req.Pn, req.Ps)
	// ps, pn超过total上限
	if min >= page.Total {
		return &api.NoteListReply{Page: page}, nil
	}
	// 当前页笔记ids
	listKeys, err := s.dao.NoteList(c, req.Mid, min, max, noteUser.NoteCount)
	if err != nil {
		log.Error("noteError NoteList req(%+v) err(%+v)", req, err)
		return nil, err
	}
	if len(listKeys) == 0 {
		return &api.NoteListReply{Page: page}, nil
	}
	// 拆解zset value
	ntList, noteIds := note.ToNtKeys(req.Mid, listKeys)
	// 批量获取笔记详情
	cacheDetails, err := s.dao.NoteDetails(c, noteIds, req.Mid)
	if err != nil {
		log.Error("noteError NoteList mid(%d) noteIds(%v) err(%+v)", req.Mid, noteIds, err)
		return nil, err
	}
	// 获取笔记所属稿件详情
	var (
		eg2        = errgroup.WithContext(c)
		aids, sids = note.ToVideoIds(cacheDetails)
		arcs       map[int64]*arcapi.Arc
		arcsForbid map[int64]bool
		ssns       map[int32]*cssngrpc.SeasonCard
		artDetails map[int64]*article.ArtDtlCache
		cvids      []int64
	)
	if len(aids) > 0 {
		eg2.Go(func(c context.Context) (err error) {
			arcs, err = s.dao.Arcs(c, aids)
			return err
		})
		eg2.Go(func(c context.Context) error {
			fbdRes, e := s.ArcsForbid(c, &api.ArcsForbidReq{Aids: aids})
			if e != nil {
				return e
			}
			if fbdRes == nil || fbdRes.Items == nil {
				arcsForbid = make(map[int64]bool)
				return nil
			}
			arcsForbid = fbdRes.Items
			return nil
		})
	}
	if len(sids) > 0 {
		eg2.Go(func(c context.Context) (err error) {
			ssns, err = s.dao.CheeseSeasons(c, sids)
			return err
		})
	}
	eg2.Go(func(c context.Context) (err error) {
		if artDetails, err = s.artDao.ArtDetails(c, noteIds, article.TpArtDetailNoteId); err != nil {
			return err
		}
		for _, art := range artDetails {
			if art.Cvid > 0 {
				cvids = append(cvids, art.Cvid)
			}
		}
		return nil
	})
	if err = eg2.Wait(); err != nil {
		log.Error("noteError NoteList req(%+v) err(%+v)", req, err)
		return nil, err
	}
	res := note.DealNoteListItem(page, ntList, cacheDetails, arcs, ssns, artDetails, s.c.NoteCfg.WebUrlFromSpace, arcsForbid)
	return res, nil
}

func (s *Service) SimpleNotes(c context.Context, req *api.SimpleNotesReq) (*api.SimpleNotesReply, error) {
	var (
		eg      = errgroup.WithContext(c)
		noteDtl map[int64]*note.DtlCache
		artDtl  map[int64]*article.ArtDtlCache
	)
	eg.Go(func(c context.Context) error {
		var err error
		noteDtl, err = s.dao.NoteDetails(c, req.NoteIds, req.Mid)
		return err
	})
	if req.Tp == api.SimpleNoteType_PUBLISH {
		eg.Go(func(c context.Context) error {
			var err error
			artDtl, err = s.artDao.ArtDetails(c, req.NoteIds, article.TpArtDetailNoteId)
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("NoteError SimpleNotes err(%+v)", err)
		return nil, err
	}
	res := make(map[int64]*api.SimpleNoteCard)
	for _, d := range noteDtl {
		sn := d.ToSimpleCard(artDtl[d.NoteId])
		if sn != nil {
			res[d.NoteId] = sn
		}
	}
	return &api.SimpleNotesReply{Items: res}, nil
}
