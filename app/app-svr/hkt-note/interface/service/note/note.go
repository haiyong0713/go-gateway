package note

import (
	"context"
	"math"
	"strconv"

	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"go-gateway/app/app-svr/hkt-note/interface/model/article"
	"go-gateway/app/app-svr/hkt-note/interface/model/note"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/pkg/errors"
	common "go-gateway/app/app-svr/hkt-note/common"
)

func (s *Service) NoteAdd(c context.Context, req *note.NoteAddReq, mid int64) (*note.NoteAddRes, error) {
	var (
		err      error
		isCreate = req.NoteId == 0
		isPub    = req.Publish == note.NeedPublish
	)
	defer func() {
		if err != nil {
			log.Error("NoteError NoteAdd req(%+v) mid(%d) err(%+v)", req, mid, err)
			if e := s.dao.DelKey(c, s.dao.AidKey(req, mid)); e != nil { // 防止写入了错误的noteId
				log.Warn("NoteWarn DelKey req(%+v) mid(%d) err(%+v)", req, mid, e)
			}
		}
	}()
	// 判断该noteId是否合法
	var (
		arc      *note.ArcCore
		noteSize int64
	)
	// 若发布，判断发布参数是否正确
	if isPub {
		if err = s.isNotePubValid(c, req, mid, isCreate); err != nil {
			return nil, err
		}
	}
	if arc, noteSize, err = s.isNoteAddValid(c, req, mid, isCreate, isPub); err != nil {
		return nil, err
	}
	// 生成note_id
	if req.NoteId == 0 {
		if req.NoteId, err = s.dao.SeqId(c); err != nil {
			return nil, err
		}
	}
	// update db note_content
	if err = s.dao.UpContent(c, req.ToNtContent(mid)); err != nil {
		return nil, err
	}
	// if create,update cache note_aid
	if isCreate {
		canSetnx, e := s.dao.NoteAidSetNX(c, mid, req)
		if e != nil {
			log.Warn("NoteWarn NoteAdd err(%+v)", e)
		}
		if !canSetnx && e == nil { // 说明在短时间内有多个create操作，拦截
			return nil, errors.Wrapf(xecode.NoteInArcAlreadyExisted, "NoteAidSetNX can't setnx,req(%+v)", req)
		}
	}
	ntAddMsg := req.ToNtNotifyMsg(mid, noteSize)
	if req.NeedAudit == note.NeedAudit {
		// 过敏感词审核，更新数据
		err = s.sendNoteAuditNotify(c, ntAddMsg)
	} else {
		// 不过敏感词，更新数据
		err = s.sendNoteNotify(c, &note.NtNotifyMsg{NtAddMsg: ntAddMsg}, req.NoteId)
	}
	if err != nil {
		return nil, err
	}
	eg := errgroup.WithContext(c)
	// 广播笔记变更消息
	if req.Hash != "" {
		eg.Go(func(c context.Context) error {
			if e := s.dao.BroadcastSync(c, req.NoteId, req.Hash); e != nil {
				log.Error("noteWarn NoteAdd err(%+v)", e)
			}
			return nil
		})
	}
	// 新增笔记时数据上报
	if isCreate {
		eg.Go(func(c context.Context) error {
			s.infocNote(c, arc, mid, note.ActionAdd, note.ToInfocPlat(req.Device), req.NoteId)
			return nil
		})
	}
	// 笔记发布时，发送databus
	if isPub {
		eg.Go(func(c context.Context) error {
			if e := s.sendNoteNotify(c, &note.NtNotifyMsg{NtPubMsg: req.ToNtPubMsg(mid, arc)}, req.NoteId); e != nil {
				return e
			}
			// 更新审核状态为待审核，锁住笔记编辑
			if e := s.artDao.AddCacheArtDetail(c, req.NoteId, article.TpArtDetailNoteId, &article.ArtDtlCache{PubStatus: article.PubStatusWaiting}); e != nil {
				log.Warn("noteWarn AddCacheArtDetail err(%+v)", e)
			}
			s.infocNote(c, arc, mid, note.ActionPub, note.ToInfocPlat(req.Device), req.NoteId)
			// 记录笔记的评论区样式
			_ = s.recordNoteReplyFormat(c, req.NoteId, req.CommentFormat)
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return nil, err
	}
	return &note.NoteAddRes{NoteId: strconv.FormatInt(req.NoteId, 10)}, nil
}

func (s *Service) isNotePubValid(c context.Context, req *note.NoteAddReq, mid int64, isCreate bool) error {
	if req.OidType != note.OidTypeUgc {
		log.Warnc(c, "isNotePubValid invalid because oidType req %v", req)
		return errors.Wrapf(xecode.ArtPublishInvalid, "isNotePubValid req(%+v)", req)
	}
	if req.CommentFormat != 0 &&
		req.CommentFormat != common.Note_Comment_Format_Type_Old &&
		req.CommentFormat != common.Note_Comment_Format_Type_New {
		log.Warnc(c, "isNotePubValid invalid because commentFormat req %v", req)
		return errors.Wrapf(xecode.ArtPublishInvalid, "isNotePubValid req(%+v)", req)
	}
	if isCreate {
		return nil
	}
	// 发布已保存过的note_id,判断是否在审核中
	spNotes, err := s.dao.SimpleNotes(c, []int64{req.NoteId}, mid, notegrpc.SimpleNoteType_PUBLISH)
	if err != nil {
		return errors.Wrapf(err, "isNotePubValid req(%+v)", req)
	}
	sp, ok := spNotes[req.NoteId]
	if !ok {
		return errors.Wrapf(xecode.ArtPublishInvalid, "isNotePubValid req(%+v) can't find pub_status", req)
	}
	if sp.PubStatus == article.PubStatusPending {
		log.Warnc(c, "isNotePubValid invalid because pubStatus pending req %v", req)
		return errors.Wrapf(xecode.ArtPublishInvalid, "isNotePubValid req(%+v) pubStatus(%+v) invalid", req, sp)
	}
	return nil
}

func (s *Service) isNoteAddValid(c context.Context, req *note.NoteAddReq, mid int64, isCreate, isPub bool) (*note.ArcCore, int64, error) {
	// 正文字数是否超过上限
	// 前端传入字数
	if len(req.Summary) > s.c.NoteCfg.MaxSummarySize {
		return nil, 0, errors.Wrapf(xecode.NoteOverSizeLimit, "content size(%d)", req.ContLen)
	}
	if req.ContLen > s.c.NoteCfg.MaxContSize {
		return nil, 0, errors.Wrapf(xecode.NoteOverSizeLimit, "content size(%d)", req.ContLen)
	}
	// 后端计算字数
	req.ContLen = note.ToContentLen(req.Content)
	if req.ContLen > s.c.NoteCfg.MaxContSize {
		return nil, 0, errors.Wrapf(xecode.NoteOverSizeLimit, "content size(%d)", req.ContLen)
	}
	var (
		existNoteId int64
		oldSize     *notegrpc.NoteSizeReply
		arcCore     *note.ArcCore
		eg          = errgroup.WithContext(c)
	)
	// 新增笔记需判断稿件合法性，稿件被删除后仍可修改，因此修改态不判断arc合法性
	// 发布笔记需获取稿件封面
	if isCreate || isPub {
		eg.Go(func(c context.Context) error {
			arcCore = s.dao.ToArcCore(c, req.Oid, req.OidType)
			if isCreate && arcCore.Status == note.ArcStatusWrong {
				return errors.Wrapf(xecode.NoteOidInvalid, "toArcCore req (%+v) arc(%+v) invalid", req, arcCore)
			}
			return nil
		})
	}
	eg.Go(func(c context.Context) error { // 当前用户在当前稿件的笔记id
		listInArc, err := s.dao.NoteListArc(c, req.Oid, mid, req.OidType)
		if err != nil {
			return err
		}
		if len(listInArc.NoteIds) > 0 {
			existNoteId = listInArc.NoteIds[0]
		}
		return nil
	})
	eg.Go(func(c context.Context) error { // 获取当前用户总使用量和笔记
		var err error
		oldSize, err = s.dao.NoteSize(c, req.NoteId, mid)
		return err
	})
	if req.NoteId > 0 { // 若编辑，但该笔记详情未找到，报错
		eg.Go(func(c context.Context) error {
			if noteInfo, err := s.dao.NoteInfo(c, req.NoteId, mid); err != nil || noteInfo == nil {
				return xecode.NoteNotFound
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, 0, err
	}
	if req.NoteId == 0 && existNoteId != 0 { // 若新增，且同稿件同用户已存在笔记，报错
		return nil, 0, errors.Wrapf(xecode.NoteInArcAlreadyExisted, "req.NoteId(%d) existNoteId(%d)", req.NoteId, existNoteId)
	}
	if req.NoteId != 0 {
		if existNoteId != req.NoteId { // 若编辑，但noteId不为同稿件同用户已存在笔记，报错
			return nil, 0, errors.Wrapf(xecode.NoteInArcAlreadyExisted, "req.NoteId(%d) existNoteId(%d)", req.NoteId, existNoteId)
		}
	}
	noteSize := int64(math.Ceil(float64(len(req.Content)) / float64(1024)))       // kb
	if oldSize.TotalSize > s.c.NoteCfg.MaxSize && noteSize > oldSize.SingleSize { // 超过用户总容量，报错
		return nil, 0, xecode.NoteOverTotalSizeLimit
	}
	return arcCore, noteSize, nil
}

func (s *Service) NoteInfo(c context.Context, req *note.NoteInfoReq, mid int64) (*note.NoteInfoRes, error) {
	var (
		eg       = errgroup.WithContext(c)
		noteInfo *notegrpc.NoteInfoReply
		arcCore  *note.ArcCore
		isForbid bool
	)
	eg.Go(func(c context.Context) error { // 获取noteInfo
		var e error
		noteInfo, e = s.dao.NoteInfo(c, req.NoteId, mid)
		return e
	})
	eg.Go(func(c context.Context) error { // 获取稿件详情用于上报
		arcCore = s.dao.ToArcCore(c, req.Oid, req.OidType)
		return nil
	})
	if req.OidType == note.OidTypeUgc {
		eg.Go(func(c context.Context) error {
			fbdRes, e := s.dao.ArcsForbid(c, []int64{req.Oid})
			if e != nil {
				log.Warn("noteWarn NoteInfo err(%+v)", e)
				return nil
			}
			isForbid = fbdRes[req.Oid]
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("noteError err(%+v)", err)
		return nil, err
	}
	res := &note.NoteInfoRes{}
	res.From(noteInfo, arcCore, isForbid)
	s.infocNote(c, arcCore, mid, note.ActionView, note.ToInfocPlat(req.Device), req.NoteId)
	if noteInfo.AuditStatus != note.AuditPass { // 命中敏感词 上报
		s.infocNote(c, arcCore, mid, note.ActionAuditFail, note.ToInfocPlat(req.Device), req.NoteId)
	}
	return res, nil
}

func (s *Service) NoteListArc(c context.Context, oid int64, mid int64, oidType int) (*note.NoteListInArcReply, error) {
	grpcRes, err := s.dao.NoteListArc(c, oid, mid, oidType)
	if err != nil {
		log.Error("NoteError NoteListArc err(%+v)", err)
		return nil, err
	}
	out := &note.NoteListInArcReply{}
	for _, v := range grpcRes.NoteIds {
		out.NoteIds = append(out.NoteIds, strconv.FormatInt(v, 10))
	}
	log.Warn("noteInfo NoteListArc oid(%d) mid(%d) res(%+v)", oid, mid, grpcRes)
	return out, nil
}

func (s *Service) NoteDel(c context.Context, req *note.NoteDelReq, mid int64) error {
	spNotes, err := s.dao.SimpleNotes(c, req.NoteIds, mid, notegrpc.SimpleNoteType_DEFAULT)
	if err != nil {
		log.Warn("noteWarn NoteDel err(%+v)", err)
		return err
	}
	var (
		validIds = make([]int64, 0, len(spNotes))
		invalIds = make([]int64, 0, len(spNotes))
	)
	for _, id := range req.NoteIds {
		if _, ok := spNotes[id]; ok {
			validIds = append(validIds, id)
		} else {
			invalIds = append(invalIds, id)
		}
	}
	if len(invalIds) > 0 {
		log.Warn("noteInfo noteDel req(%+v) valid(%v) invalid(%v)", req, validIds, invalIds)
	}
	if len(validIds) == 0 {
		return xecode.NoteNotFound
	}
	if err = s.sendNoteNotify(c, note.ToDelNotifyMsg(mid, validIds), mid); err != nil {
		log.Warn("NoteWarn NoteDel err(%+v)", err)
		return err
	}
	eg := errgroup.WithContext(c)
	if len(spNotes) > 10 { // nolint:gomnd
		eg.GOMAXPROCS(10)
	}
	// 提前删除zset中的该笔记，防止job异步更新zset太慢影响下一次请求
	eg.Go(func(c context.Context) error {
		if err := s.dao.RemCacheNoteList(c, spNotes); err != nil {
			log.Warn("NoteWarn NoteDel err(%+v)", err)
		}
		return nil
	})
	for _, n := range spNotes {
		curN := n
		eg.Go(func(c context.Context) error {
			s.infocNote(c, &note.ArcCore{Oid: curN.Oid}, mid, note.ActionDel, note.ToInfocPlat(req.Device), curN.NoteId)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Warn("NoteWarn NoteDel err(%+v)", err)
	}
	return nil
}

func (s *Service) NoteList(c context.Context, mid, pn, ps int64) (*notegrpc.NoteListReply, error) {
	grpcRes, err := s.dao.NoteList(c, mid, pn, ps, 0, 0, 0, notegrpc.NoteListType_USER_ALL)
	if err != nil {
		log.Error("NoteError NoteList err(%+v)", err)
		return nil, err
	}
	return grpcRes, nil
}

func (s *Service) IsGray(c context.Context, mid int64) (*note.UserGray, error) {
	return &note.UserGray{IsGray: s.displayNote(mid)}, nil
}

func (s *Service) displayNote(mid int64) bool {
	if mid == 0 {
		return false
	}
	for _, w := range s.c.Gray.WhiteList {
		if mid == w {
			return true
		}
	}
	return int(mid%100) < s.c.Gray.NoteWebGray
}

func (s *Service) infocNote(c context.Context, arc *note.ArcCore, mid int64, action int, plat, noteId int64) {
	data := note.ToNtInfoc(arc, mid, action, plat, noteId)
	payload := infoc.NewLogStream("005752", data.Mid, data.Aid, data.Title, data.UpMid, data.UpName, data.TypeId, data.TypeName, data.Ctime, data.Action, data.Plat, data.NoteId)
	log.Info("infocNote %s", payload.Data)
	if err := s.infocV2Log.Info(c, payload); err != nil {
		log.Error("infocNote req(%+v) err(%v)",
			data, err)
	}
}

func (s *Service) Links(c context.Context) (*note.Links, error) {
	return &note.Links{CheeseQALink: s.c.NoteCfg.CheeseQALink}, nil
}

func (s *Service) NoteCount(c context.Context, mid int64) (*notegrpc.NoteCountReply, error) {
	return s.dao.NoteCount(c, mid)
}

func (s *Service) IsForbid(c context.Context, aid int64) (*note.IsForbidReply, error) {
	arcsForbid, err := s.dao.ArcsForbid(c, []int64{aid})
	if err != nil {
		log.Error("NoteError IsForbid aid(%d) err(%v)", aid, err)
		return nil, err
	}
	return &note.IsForbidReply{ForbidNoteEntrance: arcsForbid[aid]}, nil
}
