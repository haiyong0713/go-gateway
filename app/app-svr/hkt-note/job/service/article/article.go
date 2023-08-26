package article

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/app-svr/hkt-note/common"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/job/model/article"
	"go-gateway/app/app-svr/hkt-note/job/model/note"
)

func (s *Service) consumeArticleBinlog() {
	defer s.waiter.Done()
	for {
		if s.closed {
			return
		}
		msg, ok := <-s.articleBinlogSub.Messages()
		if !ok {
			log.Warn("artWarn binlog consumeArticleBinlog quit")
			return
		}
		log.Warn("consumeArticleBinlog msg.value(%s)", string(msg.Value))
		if err := msg.Commit(); err != nil {
			log.Error("artError consumeArticleBinlog commit error(%v)", err)
			continue
		}
		m := &note.Binlog{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("artError consumeArticleBinlog msg(%s) error(%v)", msg.Value, err)
			continue
		}
		if m.Table != "articles" {
			continue
		}
		data := &article.ArtOriginalBlog{}
		if err := json.Unmarshal(msg.Value, data); err != nil {
			log.Error("artError consumeArticleBinlog msg(%s) error(%v)", msg.Value, err)
			continue
		}
		if data.New == nil {
			log.Warn("artWarn consumeArticleBinlog empty result")
			continue
		}
		if data.New.Type != article.ArtTpNote {
			continue
		}
		s.treatArticleBinlog(context.Background(), data)
	}
}

func (s *Service) treatArticleBinlog(c context.Context, msg *article.ArtOriginalBlog) {
	var (
		err                error
		artNeed, pubReason = msg.ArtNeed()
	)
	switch artNeed {
	case article.ArtDBNeedPub:
		msg.New.Reason = pubReason
		err = s.artDao.UpPubStatus(c, msg.New)
	case article.ArtDBNeedRm:
		err = func() error {
			if e := s.artDao.DelArtDetail(c, msg.New.Id, msg.New.Mid); e != nil {
				return e
			}
			return s.artDao.DelArtContent(c, msg.New.Id, msg.New.Mid)
		}()
		// 判断泰山是否存在该cvid,存在则调用评论删除接口及删除泰山相应kv
		s.treatReplyDel(c, msg)
	default:
	}
	if err != nil {
		log.Error("artError consumeNoteNotifyMsg msg(%+v) error(%v)", msg, err)
		jsonBody, jsonErr := json.Marshal(msg)
		if jsonErr != nil {
			log.Error("artError treatArticleBinlog msg(%+v) error(%v)", msg, jsonErr)
		} else {
			s.dao.AddCacheRetry(c, article.KeyRetryArtBinlog, string(jsonBody), time.Now().Unix())
		}
	}
}
func (s *Service) treatReplyDel(c context.Context, msg *article.ArtOriginalBlog) {
	var (
		rpidInfo = &common.TaishanCvidMappingRpidInfo{}
	)
	key := fmt.Sprintf(common.Cvid_Mapping_Rpid_Taishan_Key, msg.New.Id)
	record, err := s.artDao.GetTaishan(c, key, common.TaishanConfig.NoteReply)
	if err != nil {
		log.Errorc(c, "query user info from taiShan failed:(%v), key:%v", err, key)
		return
	}
	if record == nil || len(record.Columns) == 0 || len(record.Columns[0].Value) == 0 {
		log.Infoc(c, "query cvidInfo from taiShan record isEmpty:(%v), key:%v", record, key)
		return
	}
	err = json.Unmarshal(record.Columns[0].Value, rpidInfo)
	if err != nil {
		log.Errorc(c, "GetCvid from taiShan Unmarshal err %v and key %v", err, key)
		return
	}
	if rpidInfo.Status != common.Cvid_Rpid_Attached {
		log.Warnc(c, "rpidInfo Status from taishan is deleted rpidInfo %v", rpidInfo)
		return
	}
	if err = s.artDao.ReplyDel(c, msg.New.Mid, rpidInfo.Rpid); err != nil {
		log.Errorc(c, "ReplyDel err %v mid(%d) RpId(%d)", err, msg.New.Mid, rpidInfo.Rpid)
		return
	}
	if err = s.artDao.UnbindNoteReplyTaishan(c, msg.New.Id, rpidInfo.Rpid); err != nil {
		log.Errorc(c, "UnbindNoteReplyTaishan err %v cvid(%d) RpId(%d) ", err, msg.New.Id, rpidInfo.Rpid)
		return
	}

}
