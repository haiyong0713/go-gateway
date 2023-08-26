package note

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	artmdl "go-gateway/app/app-svr/hkt-note/job/model/article"
	"go-gateway/app/app-svr/hkt-note/job/model/note"

	frontgrpc "git.bilibili.co/bapis/bapis-go/frontend/bilinote/v1"
	"github.com/pkg/errors"
)

func (s *Service) treatNotePubNotifyMsg(c context.Context, msg *note.NtPubMsg) error {
	var (
		err  error
		eg   = errgroup.WithContext(c)
		cont *note.ContCache
		art  *artmdl.ArtDtlCache
	)
	// 获取笔记正文
	eg.Go(func(c context.Context) error {
		var e error
		if cont, e = s.dao.NoteContent(c, msg.NoteId); e != nil {
			return e
		}
		if cont.NoteId == -1 {
			return errors.Wrapf(ecode.NothingFound, "NoteContent noteId(%d)", msg.NoteId)
		}
		return nil
	})
	// 获取该客态笔记最新版本。若最新版本为锁定，把该cvid都删掉，重新请求专栏生成新的
	eg.Go(func(c context.Context) error {
		var e error
		if art, e = s.artDao.ArtDetail(c, msg.NoteId, artmdl.TpArtDetailNoteId, 0, 0, true); e != nil {
			return e
		}
		if art.PubStatus != artmdl.PubStatusLock {
			return nil
		}
		if e = s.artDao.DelArtContent(c, art.Cvid, msg.Mid); e != nil {
			return e
		}
		if e = s.artDao.DelArtDetail(c, art.Cvid, msg.Mid); e != nil {
			return e
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return err
	}
	// 过滤不合法的结构
	cont.Content = func() string {
		fltContent, e := note.FilterInvalid(cont.Content)
		if e != nil {
			log.Warn("ArtWarn FilterInvalid cont(%s) e(%+v)", cont.Content, e)
			return cont.Content
		}
		// save to db
		if e = s.dao.UpContent(c, fltContent, msg.NoteId); e != nil {
			log.Warn("ArtWarn FilterInvalid cont(%s) upload e(%+v)", fltContent, e)
			return cont.Content
		}
		return fltContent
	}()
	// 获得客态正文
	var (
		bnRes      *frontgrpc.NoteReply
		pubFailMsg *artmdl.PubFailMsg
	)
	// 发布失败且重试无效，将失败原因更新至缓存
	defer func() {
		if pubFailMsg != nil {
			log.Warn("ArtError treatNotePubNotifyMsg msg(%+v) pubFailMsg(%+v)", msg, pubFailMsg)
			if e := retry.WithAttempts(c, "pubFailMsg-retry", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
				return s.artDao.AddCacheArtDetail(c, pubFailMsg.NoteId, pubFailMsg.ToCache(), artmdl.TpArtDetailNoteId, s.artDao.ArtTmpExpire)
			}); e != nil {
				log.Error("ArtError treatNotePubNotifyMsg pubFailMsg(%+v) err(%+v)", pubFailMsg, e)
			}
		}
	}()
	if bnRes, pubFailMsg, err = s.artDao.GetBiliNoteContent(c, cont.Content, msg); err != nil {
		if pubFailMsg != nil {
			return nil
		}
		return err
	}
	// 专栏提审
	argArticle := msg.ToArgArticle(art.Cvid, bnRes.BiliHtml, bnRes.ImgUrls, s.c.ArticleCfg.CategoryNote)
	switch art.ToActionType() {
	case artmdl.ArtActionAdd:
		msg.Cvid, pubFailMsg, err = s.artDao.CreateArticle(c, argArticle, msg.NoteId)
	case artmdl.ArtActionEdit:
		msg.Cvid = art.Cvid
		pubFailMsg, err = s.artDao.EditArticle(c, argArticle, msg.NoteId)
	default:
		log.Warn("artWarn article(%+v) is auditing, skip", art)
		return nil
	}
	if err != nil {
		if pubFailMsg != nil {
			return nil
		}
		return err
	}
	// 专栏提审成功，更新db和缓存
	msg.PubVersion = art.PubVersion + 1
	if err2 := s.artDao.InsertArtDetail(c, msg); err2 != nil {
		log.Error("artError treatNotePubNotifyMsg error(%v)", err2)
		msg.Mtime = time.Now().Unix()
		jsonBody, jsonErr := json.Marshal(msg)
		if jsonErr != nil {
			log.Error("artError treatNotePubNotifyMsg msg(%+v) error(%v)", msg, jsonErr)
		} else {
			s.dao.AddCacheRetry(c, note.KeyRetryArtDetailDB, string(jsonBody), msg.Mtime)
		}
	}
	ac := msg.ToCont(cont.Tag, bnRes)
	if err3 := s.artDao.InsertArtContent(c, ac); err3 != nil {
		log.Error("artError treatNotePubNotifyMsg msg(%+v) err(%+v)", msg, err3)
		jsonBody, jsonErr := json.Marshal(ac)
		if jsonErr != nil {
			log.Error("artError treatNotePubNotifyMsg msg(%+v) error(%v)", msg, jsonErr)
		} else {
			s.dao.AddCacheRetry(c, note.KeyRetryArtContDB, string(jsonBody), time.Now().Unix())
		}
	}
	return nil
}

func (s *Service) treatNoteDelNotifyMsg(c context.Context, msg *note.NtDelMsg) {
	if msg.NoteId > 0 { // TODO del NoteId
		msg.NoteIds = []int64{msg.NoteId}
	}
	for _, noteId := range msg.NoteIds {
		if err := s.dao.DelNoteCont(c, noteId); err != nil {
			log.Warn("noteError consumeNoteNotifyMsg err(%+v)", err)
			s.dao.AddCacheRetry(c, note.KeyRetryDBDelCont, strconv.FormatInt(noteId, 10), time.Now().Unix())
		}
	}
	noteIdsStr := xstr.JoinInts(msg.NoteIds)
	if err := s.dao.DelNoteDetail(c, noteIdsStr, msg.Mid); err != nil {
		log.Warn("noteError consumeNoteNotifyMsg err(%+v)", err)
		s.dao.AddCacheRetry(c, note.KeyRetryDBDelDetail, fmt.Sprintf("%s-%d", noteIdsStr, msg.Mid), time.Now().Unix()+300)
	}
}

func (s *Service) treatNoteAddNotifyMsg(c context.Context, msg *note.NtAddMsg) {
	existNoteId, err := s.dao.NoteAid(c, msg.Mid, msg.Oid, msg.OidType)
	if err != nil {
		log.Warn("retryWarn treatNoteAddNotifyMsg msg(%+v) err(%+v)", msg, err)
	}
	if err == nil && existNoteId > 0 && existNoteId != msg.NoteId { // 一个稿件只能有一个笔记，若已存在笔记且databus的noteId不为该id，跳过任务
		log.Warn("retryWarn treatNoteAddNotifyMsg msg(%+v) already has note_id,skip", msg)
		return
	}
	func() { // update db note_detail
		err := s.dao.UpNoteDetail(c, msg)
		if err == nil {
			return
		}
		log.Error("noteError consumeNoteNotifyMsg msg(%+v) error(%v)", msg, err)
		jsonBody, jsonErr := json.Marshal(msg)
		if jsonErr != nil {
			log.Error("noteError consumeNoteNotifyMsg msg(%+v) error(%v)", msg, jsonErr)
			return
		}
		s.dao.AddCacheRetry(c, note.KeyRetryDBDetail, string(jsonBody), time.Now().Unix())
	}()
	contErr := func() error { // update cache note_content
		noteContent, err := s.dao.NoteContent(c, msg.NoteId)
		if err != nil {
			return err
		}
		if msg.Content != "" { // 若过audit,content由audit方法传递
			noteContent.Content = msg.Content
		}
		return s.dao.AddCacheNoteContent(c, msg.NoteId, noteContent)
	}()
	if contErr != nil {
		log.Error("noteError consumeNoteNotifyMsg msg(%+v) err(%+v)", msg, contErr)
		s.dao.AddCacheRetry(c, note.KeyRetryContent, strconv.FormatInt(msg.NoteId, 10), time.Now().Unix()+300)
	}
}

func (s *Service) treatNoteAuditMsg(c context.Context, msg *note.NtAddMsg) {
	var err error
	defer func() {
		if err != nil {
			log.Error("noteError consumeAuditNotifyMsg msg(%+v) err(%+v)", msg, err)
			jsonBody, jsonErr := json.Marshal(msg)
			if jsonErr != nil {
				log.Error("noteError consumeAuditNotifyMsg marshal msg(%+v) error(%v)", msg, jsonErr)
				return
			}
			s.dao.AddCacheRetry(c, note.KeyRetryAudit, string(jsonBody), time.Now().Unix()+300)
		}
	}()
	// get content from db
	var content *note.ContCache
	if content, err = s.dao.NoteContent(c, msg.NoteId); err != nil {
		return
	}
	if content.NoteId == -1 {
		log.Warn("noteWarn consumeAuditNotifyMsg msg(%+v) content not found", msg)
		return
	}
	msg.Content = content.Content
	var (
		auditStatus int
		fltSummary  = msg.Summary
	)
	auditStatus, err = func() (int, error) {
		// switch to only body
		body := note.ToBody(content.Content)
		if len(body) == 0 {
			return note.AuditPass, nil
		}
		if int64(len(body)) >= s.c.NoteCfg.FilterLimit { // 超过敏感词容量上限，暂时跳过审核
			log.Warn("noteInfo treatNoteAuditMsg msg(%+v) length too large, skip", msg)
			return note.AuditPass, nil
		}
		// request filter api
		sensitive, err1 := s.dao.FilterV3(c, body, msg.NoteId, msg.Mid)
		if err1 != nil {
			return note.AuditSkip, err1
		}
		// delete content cache, let it refresh
		if err1 = s.dao.DelKey(c, s.dao.ContentKey(msg.NoteId)); err1 != nil {
			return note.AuditSkip, err1
		}
		// replace sensitive words
		var (
			sensitiveStr = note.ToSensitiveStr(sensitive)
			fltContent   string
		)
		if fltContent, err1 = note.ReplaceSensitive(content.Content, sensitiveStr); err1 != nil {
			return note.AuditSkip, err1
		}
		fltSummary = note.ReplaceInStr(fltSummary, sensitiveStr)
		// save to db
		if err1 = s.dao.UpContent(c, fltContent, msg.NoteId); err1 != nil {
			return note.AuditSkip, err1
		}
		msg.Content = fltContent
		s.infocAudit(c, msg, sensitiveStr, fltContent)
		if len(sensitive) == 0 { // 没有敏感词，但需要格式过滤
			return note.AuditPass, nil
		}
		return note.AuditFail, nil
	}()
	if err != nil || auditStatus == note.AuditSkip {
		return
	}
	// change in db
	msg.Summary = fltSummary
	msg.AuditStatus = auditStatus
	log.Warn("noteInfo to treatNoteAddNotifyMsg msg(%+v)", msg)
	s.treatNoteAddNotifyMsg(c, msg)
}

func (s *Service) treatReplyMsg(c context.Context, msg *note.ReplyMsg) error {
	cvid, pubVer, pubStat, err := s.artDao.LatestArtByNoteId(c, msg)
	if err != nil {
		log.Error("ArtError treatReplyMsg err(%+v)", err)
		return err
	}
	if cvid == 0 || pubVer == 0 {
		log.Warn("ArtWarn treatReplyMsg msg(%+v) can't find", msg)
		return nil
	}
	// 拼接评论正文
	noteUrl := fmt.Sprintf(s.c.NoteCfg.ReplyCfg.ReplyUrl, msg.NoteId)
	// 若笔记已过审，直接发评论
	if pubStat == artmdl.PubStatusPassed {
		if replyErr := retry.WithAttempts(c, "autoReply-retry", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			webUrl := fmt.Sprintf(s.c.NoteCfg.ReplyCfg.WebUrl, cvid)
			replyCont := strings.Replace(msg.Content, noteUrl, webUrl, -1)
			return s.artDao.ReplyAdd(c, msg.Mid, msg.Oid, replyCont)
		}); replyErr != nil {
			log.Warn("artWarn treatReplyMsg reply err(%+v)", replyErr)
		}
		return nil
	}
	// 若笔记暂未过审，将正文写入db
	replyCont := strings.Replace(msg.Content, noteUrl, "%s", -1)
	if err = s.artDao.UpdateCommentInfo(c, cvid, pubVer, replyCont); err != nil {
		log.Error("ArtError treatReplyMsg err(%+v)", err)
		return err
	}
	return nil
}

func (s *Service) infocAudit(c context.Context, msg *note.NtAddMsg, sensitive []string, content string) {
	payload := infoc.NewLogStream("006630", msg.Aid, msg.Mid, msg.NoteId, strings.Join(sensitive, ","), content, msg.Title, time.Now().Format("2006-01-02 15:04:05"))
	log.Info("infocAudit %s", payload.Data)
	if err := s.infocV2Log.Info(c, payload); err != nil {
		log.Error("infocAudit req(%+v,%v) err(%v)",
			msg, sensitive, err)
	}
}
