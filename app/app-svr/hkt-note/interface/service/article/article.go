package article

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"go-gateway/app/app-svr/hkt-note/interface/model/article"
	"go-gateway/app/app-svr/hkt-note/interface/model/note"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	accountRelationGrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"github.com/pkg/errors"
)

func (s *Service) PublishDel(c context.Context, cvids []int64, mid int64) error {
	if err := s.artDao.DelUpArticles(c, cvids, mid); err != nil {
		log.Error("ArtError PublishDel err(%+v)", err)
		return err
	}
	// 提前删除zset中的该笔记，防止job异步更新zset太慢影响下一次请求
	if err := func() error {
		arts, e := s.artDao.SimpleArticles(c, cvids)
		if e != nil {
			return e
		}
		return s.artDao.RemCachesArtListUser(c, mid, arts)
	}(); err != nil {
		log.Warn("ArtWarn PublishDel RemList err(%+v)", err)
	}
	return nil
}

func (s *Service) PubListInUser(c context.Context, pn, ps, mid int64) (*notegrpc.NoteListReply, error) {
	grpcRes, err := s.noteDao.NoteList(c, mid, pn, ps, 0, 0, 0, notegrpc.NoteListType_USER_PUBLISHED)
	if err != nil {
		log.Error("ArtError PubListInUser err(%+v)", err)
		return nil, err
	}
	return grpcRes, nil
}

func (s *Service) PubListInArc(c context.Context, req *article.PubListInArcReq, mid int64) (*article.PubListInArcRes, error) {
	// 判断up是否开放笔记展示
	arcCore := s.noteDao.ToArcCore(c, req.Oid, req.OidType)
	if arcCore.UpMid == 0 {
		err := errors.Wrapf(xecode.NoteOidInvalid, "PubListInArc req(%+v)", req)
		log.Error("ArtError err(%+v)", err)
		return nil, err
	}
	showNote, err := s.artDao.UpSwitch(c, arcCore.UpMid)
	if err != nil {
		log.Error("ArtError err(%+v)", err)
		return nil, err
	}
	if !showNote {
		return &article.PubListInArcRes{Message: s.c.NoteCfg.Messages.UpSwitchMsg}, nil
	}
	// 获取笔记信息
	pubList, err := s.noteDao.NoteList(c, mid, req.Pn, req.Ps, req.Oid, int64(req.OidType), req.UperMid, notegrpc.NoteListType_ARCHIVE_PUBLISHED)
	if err != nil {
		log.Error("ArtError err(%+v)", err)
		return nil, err
	}
	if pubList.Page != nil && pubList.Page.Total == 0 {
		return &article.PubListInArcRes{ShowPublicNote: true, Message: s.c.NoteCfg.Messages.ListNoneMsg}, nil
	}
	var accs map[int64]*accgrpc.Card
	if len(pubList.List) > 0 {
		mids := make([]int64, 0, len(pubList.List))
		for _, l := range pubList.List {
			mids = append(mids, l.Mid)
		}
		// 获取作者信息
		if accs, err = s.artDao.AccCards(c, mids); err != nil {
			log.Error("ArtError err(%+v)", err)
			return nil, err
		}
	}
	res := article.ToPubListInArcRes(pubList, accs)
	return res, nil
}

func (s *Service) PublishNoteInfo(c context.Context, req *article.PubNoteInfoReq) (*article.PubNoteInfoRes, error) {
	info, err := s.artDao.PubNoteInfo(c, req.Cvid)
	if err != nil {
		log.Error("ArtError PublishNoteInfo err(%+v)", err)
		return nil, err
	}
	var (
		eg        = errgroup.WithContext(c)
		arcCore   *note.ArcCore
		author    *article.Author
		isForbid  bool
		statReply *accountRelationGrpc.StatReply
	)
	// 获取作者信息
	eg.Go(func(c context.Context) error {
		acc, e := s.artDao.AccCards(c, []int64{info.Mid})
		if e != nil {
			return e
		}
		if _, ok := acc[info.Mid]; !ok {
			return xecode.AuthorNotFound
		}
		author = article.FromAcc(acc[info.Mid])
		return nil
	})
	// 获取作者计数信息
	eg.Go(func(ctx context.Context) error {
		rsp, err := s.artDao.AccountRelationStats(ctx, []int64{info.Mid})
		if err != nil {
			return err
		}
		if rsp == nil {
			return nil
		}
		if value, ok := rsp.StatReplyMap[info.Mid]; ok {
			statReply = value
		}
		return nil
	})
	// 获取稿件信息
	eg.Go(func(c context.Context) error {
		arcCore = s.noteDao.ToArcCore(c, info.Oid, int(info.OidType))
		return nil
	})
	if info.OidType == note.OidTypeUgc {
		eg.Go(func(c context.Context) error {
			fbdRes, e := s.noteDao.ArcsForbid(c, []int64{info.Oid})
			if e != nil {
				log.Warn("noteWarn NoteInfo err(%+v)", e)
				return nil
			}
			isForbid = fbdRes[info.Oid]
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("ArtError PublishNoteInfo err(%+v)", err)
		return nil, err
	}
	res := article.ToPubNoteInfoRes(info, author, statReply, arcCore, isForbid)
	return res, nil
}
