package article

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/hkt-note/job/model/article"
)

func (s *Service) retryArticleBinlog() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		time.Sleep(time.Duration(s.c.NoteCfg.RetryFre))
		c := context.TODO()
		res, err := s.dao.CacheRetry(c, article.KeyRetryArtBinlog)
		if err != nil {
			log.Error("retryError retryArticleBinlog err(%+v)", err)
			continue
		}
		if res == "" {
			continue
		}
		s.dao.RemCacheRetry(c, article.KeyRetryArtBinlog, res)
		msg := &article.ArtOriginalBlog{}
		if err = json.Unmarshal([]byte(res), msg); err != nil {
			log.Error("retryError retryArticleBinlog key(%s) err(%+v)", res, err)
			continue
		}
		var art *article.ArtDtlCache
		art, err = s.artDao.ArtDetail(c, msg.New.Id, article.TpArtDetailCvid, 0, 0, false)
		if err != nil {
			log.Error("retryWarn retryArticleBinlog msg(%+v) err(%+v)", msg, err)
			s.dao.AddCacheRetry(c, article.KeyRetryArtBinlog, res, time.Now().Unix())
			continue
		}
		if art.Deleted == 1 {
			log.Warn("retryInfo retryArticleBinlog db(%+v) is deleted(%+v),skip with msg", art, msg.New)
			continue
		}
		mtime, _ := time.ParseInLocation("2006-01-02 15:04:05", msg.New.Mtime, time.Local)
		if art.Mtime > xtime.Time(mtime.Unix()) {
			log.Warn("retryInfo retryArticleBinlog db(%+v) is newer than retry(%+v),skip", art, msg.New)
			continue
		}
		s.treatArticleBinlog(c, msg)
	}
}
