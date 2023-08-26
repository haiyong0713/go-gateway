package note

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/hkt-note/job/model/article"
	"go-gateway/app/app-svr/hkt-note/job/model/note"
)

func (s *Service) retryArtBinlog() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryArtDtlBinlog)
		if err != nil {
			log.Error("retryError retryArtBinlog err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryArtDtlBinlog, res)
		idsArr := note.ToIds(res)
		if len(idsArr) != 2 { // nolint:gomnd
			log.Error("retryError retryArtBinlog key(%s) invalid", res)
			continue
		}
		cvid, noteId := idsArr[0], idsArr[1]
		err = func() error {
			// 主人态笔记详情(noteId维度最后一次发布版本)
			artHost, e := s.artDao.ArtDetail(c, noteId, article.TpArtDetailNoteId, 0, 0, false)
			if e != nil {
				return e
			}
			// 客态笔记详情(cvid维度最后一次过审版本)
			artPub, e := s.artDao.ArtDetail(c, cvid, article.TpArtDetailCvid, article.PubStatusPassed, 0, true)
			if e != nil {
				return e
			}
			var (
				publicStat = artPub.ToPublicStatus(artHost)
				msg        = artHost.ToDetailDB(noteId)
			)
			// artHost.cvid与重试的cvid不一致(原cvid被锁定需要申请新cvid时),因此原cvid需带入方法
			s.treatArtDetailBinlog(c, msg, publicStat, false, cvid)
			return nil
		}()
		if err != nil {
			log.Error("retryError retryArtBinlog res(%s) err(%+v)", res, err)
			s.dao.AddCacheRetry(c, note.KeyRetryArtDtlBinlog, res, time.Now().Unix()+60)
			continue
		}
	}
}

func (s *Service) retryArtContDB() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryArtContDB)
		if err != nil {
			log.Error("retryError retryArtContDB err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryArtContDB, res)
		msg := &article.ArtContCache{}
		if err = json.Unmarshal([]byte(res), msg); err != nil {
			log.Error("retryError retryArtContDB key(%s) err(%+v)", res, err)
			continue
		}
		err = func() error {
			cont, e := s.artDao.LatestArtCont(c, msg.Cvid)
			if e != nil {
				return e
			}
			if cont.Deleted == 1 {
				log.Warn("retryInfo retryArtContDB db(%+v) is deleted,skip with msg(%+v)", cont, msg)
				return nil
			}
			if cont.PubVersion > msg.PubVersion || (cont.PubVersion == msg.PubVersion && cont.Mtime > msg.Mtime) {
				log.Warn("retryInfo retryArtContDB db(%+v) is newer than retry(%+v),skip", cont, msg)
				return nil
			}
			return s.artDao.InsertArtContent(c, msg)
		}()
		if err != nil {
			log.Error("retryWarn retryArtContDB msg(%+v) err(%+v)", msg, err)
			s.dao.AddCacheRetry(c, note.KeyRetryArtContDB, res, time.Now().Unix()+300)
			continue
		}
	}
}

func (s *Service) retryArtDetailDB() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryArtDetailDB)
		if err != nil {
			log.Error("retryError retryArtDetailDB err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryArtDetailDB, res)
		msg := &note.NtPubMsg{}
		if err = json.Unmarshal([]byte(res), msg); err != nil {
			log.Error("retryError retryArtDetailDB key(%s) err(%+v)", res, err)
			continue
		}
		err = func() error {
			art, e := s.artDao.ArtDetail(c, msg.Cvid, article.TpArtDetailCvid, 0, 0, false)
			if e != nil {
				return e
			}
			if art.Deleted == 1 {
				log.Warn("retryInfo retryArtDetailDB db(%+v) is deleted,skip with msg(%+v)", art, msg)
				return nil
			}
			if art.PubVersion > msg.PubVersion || (art.PubVersion == msg.PubVersion && art.Mtime > xtime.Time(msg.Mtime)) {
				log.Warn("retryInfo retryArtDetailDB db(%+v) is newer than retry(%+v),skip", art, msg)
				return nil
			}
			return s.artDao.InsertArtDetail(c, msg)
		}()
		if err != nil {
			log.Error("retryWarn retryArtDetailDB msg(%+v) err(%+v)", msg, err)
			s.dao.AddCacheRetry(c, note.KeyRetryArtDetailDB, res, time.Now().Unix())
			continue
		}
	}
}

func (s *Service) retryAid() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryAid)
		if err != nil {
			log.Error("retryError retryAid err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryAid, res)
		idsArr := note.ToIds(res)
		if len(idsArr) != 3 { // nolint:gomnd
			log.Error("retryError retryAid key(%s) invalid", res)
			continue
		}
		mid, aid, oidType := idsArr[0], idsArr[1], idsArr[2]
		err = func() error {
			noteId, err := s.dao.NoteAid(c, mid, aid, int(oidType))
			if err != nil {
				return err
			}
			if err := s.dao.AddCacheNoteAid(c, mid, aid, noteId, int(oidType)); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			log.Warn("retryWarn retryAid mid(%d) aid(%d) err(%+v)", mid, aid, err)
			s.dao.AddCacheRetry(c, note.KeyRetryAid, fmt.Sprintf("%d-%d", mid, aid), time.Now().Unix()+300)
			continue
		}
	}
}

func (s *Service) retryUser() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryUser)
		if err != nil {
			log.Error("retryError retryUser err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryUser, res)
		mid, _ := strconv.ParseInt(res, 10, 64)
		if mid == 0 {
			log.Error("retryError retryUser key(%s) invalid", res)
			continue
		}
		if err := s.updateUser(c, mid); err != nil {
			log.Error("retryWarn retryUser mid(%d) err(%+v)", mid, err)
			s.dao.AddCacheRetry(c, note.KeyRetryUser, strconv.FormatInt(mid, 10), time.Now().Unix()+300)
			continue
		}
	}
}

func (s *Service) retryNoteListRem() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryListRem)
		if err != nil {
			log.Error("retryError retryNoteListRem err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryListRem, res)
		idsArr := note.ToIds(res)
		if len(idsArr) != 3 { // nolint:gomnd
			log.Error("retryError retryNoteListRem key(%s) invalid", res)
			continue
		}
		noteId, mid, aid := idsArr[0], idsArr[1], idsArr[2]
		if err = s.dao.RemCacheNoteList(c, mid, fmt.Sprintf("%d-%d", noteId, aid)); err != nil {
			log.Error("retryWarn retryNoteListRem noteId(%d) mid(%d) aid(%d) err(%+v)", noteId, mid, aid, err)
			s.dao.AddCacheRetry(c, note.KeyRetryListRem, fmt.Sprintf("%d-%d-%d", noteId, mid, aid), time.Now().Unix()+300)
			continue
		}
	}
}

func (s *Service) retryNoteList() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryList)
		if err != nil {
			log.Error("retryError retryNoteList err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryList, res)
		idsArr := note.ToIds(res)
		if len(idsArr) != 3 { // nolint:gomnd
			log.Error("retryError retryNoteList key(%s) invalid", res)
			continue
		}
		noteId, mid, aid := idsArr[0], idsArr[1], idsArr[2]
		err = func() error {
			dtlCache, err := s.dao.NoteDetail(c, noteId, mid)
			if err != nil {
				return err
			}
			if dtlCache.NoteId == -1 {
				log.Info("retryInfo retryNoteList note(%d) mid(%d) deleted", noteId, mid)
				return nil
			}
			if err = s.dao.AddCacheNoteList(c, mid, fmt.Sprintf("%d-%d", noteId, aid), dtlCache.Mtime.Time().Unix()); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			log.Error("retryWarn retryNoteList noteId(%d) mid(%d) aid(%d) err(%+v)", noteId, mid, aid, err)
			s.dao.AddCacheRetry(c, note.KeyRetryList, fmt.Sprintf("%d-%d-%d", noteId, mid, aid), time.Now().Unix()+300)
			continue
		}
	}
}

func (s *Service) retryDetail() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryDetail)
		if err != nil {
			log.Error("retryError retryDetail err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryDetail, res)
		idsArr := note.ToIds(res)
		if len(idsArr) != 2 { //nolint:gomnd
			log.Error("retryError retryDetail key(%s) invalid", res)
			continue
		}
		noteId, mid := idsArr[0], idsArr[1]
		err = func() error {
			dtlCache, err := s.dao.NoteDetail(c, noteId, mid)
			if err != nil {
				return err
			}
			if err = s.dao.AddCacheNoteDetail(c, noteId, dtlCache); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			log.Warn("retryWarn retryDetail noteId(%d) mid(%d) err(%+v)", noteId, mid, err)
			s.dao.AddCacheRetry(c, note.KeyRetryDetail, fmt.Sprintf("%d-%d", noteId, mid), time.Now().Unix()+300)
			continue
		}
	}
}

func (s *Service) retryDetailDB() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryDBDetail)
		if err != nil {
			log.Error("retryError retryDetail err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryDBDetail, res)
		rs := &note.NtAddMsg{}
		if err := json.Unmarshal([]byte(res), rs); err != nil {
			log.Warn("retryWarn retryDetailDB msg(%s) ignore, err(%+v)", res, err)
			continue
		}
		var existNoteId int64
		if existNoteId, err = s.dao.NoteAid(c, rs.Mid, rs.Oid, rs.OidType); err != nil { // 一个稿件只能有一个笔记，若已存在笔记且重试的noteId不为该id，不更新db
			s.dao.AddCacheRetry(c, note.KeyRetryDBDetail, res, time.Now().Unix()+300)
			log.Warn("retryWarn retryDetailDB msg(%s) db error,wait until next time", res)
			continue
		}
		if existNoteId != rs.NoteId {
			log.Warn("retryWarn retryDetailDB noteId not valid,existId(%d) retryId(%d)", existNoteId, rs.NoteId)
			continue
		}
		err = func() error {
			dtl, err := s.dao.NoteDetail(c, rs.NoteId, rs.Mid)
			if err != nil {
				return err
			}
			if dtl.Mtime > rs.Mtime { // 说明db已经被新数据更新过了
				log.Warn("retryInfo retryDetailDB dbdata(%+v) newer than retrydata(%+v),skip", dtl, res)
				return nil
			}
			if err = s.dao.UpNoteDetail(c, rs); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			log.Error("retryWarn retryDetailDB msg(%s) err(%+v)", res, err)
			s.dao.AddCacheRetry(c, note.KeyRetryDBDetail, res, time.Now().Unix()+300)
			continue
		}
	}
}

func (s *Service) retryContDBDel() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryDBDelCont)
		if err != nil {
			log.Error("retryError retryContDBDel err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryDBDelCont, res)
		noteId, _ := strconv.ParseInt(res, 10, 64)
		if noteId == 0 {
			log.Error("retryError retryContentDB invalid noteId(%d),skip", noteId)
			continue
		}
		if err := s.dao.DelNoteCont(c, noteId); err != nil {
			log.Error("retryWarn retryContentDB msg(%s) err(%+v)", res, err)
			s.dao.AddCacheRetry(c, note.KeyRetryDBDelCont, res, time.Now().Unix()+300)
			continue
		}
	}
}

func (s *Service) retryDetailDBDel() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryDBDelDetail)
		if err != nil {
			log.Error("retryError retryDetail err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryDBDelDetail, res)
		noteIdStr, mid := note.ToStrIds(res)
		if noteIdStr == "" || mid == 0 {
			log.Error("retryError retryDetail key(%s) invalid,skip", res)
			continue
		}
		if err := s.dao.DelNoteDetail(c, noteIdStr, mid); err != nil {
			log.Warn("retryWarn retryDetailDBDel msg(%s) err(%+v)", res, err)
			s.dao.AddCacheRetry(c, note.KeyRetryDBDelDetail, res, time.Now().Unix()+300)
			continue
		}
	}
}

func (s *Service) retryContent() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryContent)
		if err != nil {
			log.Error("retryError retryContent err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryContent, res)
		noteId, _ := strconv.ParseInt(res, 10, 64)
		if noteId == 0 {
			log.Error("retryError retryContent key(%s) invalid", res)
			continue
		}
		err = func() error {
			contCache, err := s.dao.NoteContent(c, noteId)
			if err != nil {
				return err
			}
			if err := s.dao.AddCacheNoteContent(c, noteId, contCache); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			log.Error("retryWarn retryContent noteId(%d) err(%+v)", noteId, err)
			s.dao.AddCacheRetry(c, note.KeyRetryContent, strconv.FormatInt(noteId, 10), time.Now().Unix()+300)
			continue
		}
	}
}

func (s *Service) retryDelCache() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryDel)
		if err != nil {
			log.Error("retryError retryDelCache err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryDel, res)
		err = s.dao.DelKey(c, res)
		if err != nil {
			log.Error("retryWarn retryDelCache key(%s) err(%+v)", res, err)
			s.dao.AddCacheRetry(c, note.KeyRetryDel, res, time.Now().Unix()+300)
			continue
		}
	}
}

func (s *Service) retryAudit() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, note.KeyRetryAudit)
		if err != nil {
			log.Error("retryError retryAudit err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, note.KeyRetryAudit, res)
		rs := &note.NtAddMsg{}
		if err := json.Unmarshal([]byte(res), rs); err != nil {
			log.Warn("retryWarn retryAudit msg(%s) ignore, err(%+v)", res, err)
			continue
		}
		s.treatNoteAuditMsg(c, rs)
	}
}
