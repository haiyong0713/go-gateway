package article

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/common"
	"go-gateway/app/app-svr/hkt-note/job/model/article"
)

func (s *Service) treatReplyDelMsg(c context.Context, msg *article.ReplyDelMsg) (err error) {
	// 1.判断泰山中是否存在该评论id
	var (
		cvidInfo = &common.TaishanRpidMappingCvidInfo{}
	)
	key := fmt.Sprintf(common.Rpid_Mapping_Cvid_Taishan_Key, msg.RpID)
	record, err := s.artDao.GetTaishan(c, key, common.TaishanConfig.NoteReply)
	if err != nil {
		log.Errorc(c, "query cvidInfo from taishan faild:(%v), key:%v", err, key)
		return
	}
	if record == nil || len(record.Columns) == 0 || len(record.Columns[0].Value) == 0 {
		log.Infoc(c, "query cvidInfo from taishan record isEmpty:(%v), key:%v", record, key)
		return
	}
	err = json.Unmarshal(record.Columns[0].Value, cvidInfo)
	if err != nil {
		log.Errorc(c, "getCvidInfo from taishan Unmarshal err %v and key %v", err, key)
		return
	}
	if cvidInfo.Status != common.Cvid_Rpid_Attached {
		log.Warnc(c, "cvidInfo Status from taishan is deleted cvidInfo %v", cvidInfo)
		return
	}
	switch msg.DelFrom {
	case article.DelFromUp:
		if err = s.artDao.DelArtDetailByCvid(c, cvidInfo.Cvid); err != nil {
			return err
		}
		if err = s.artDao.DelArtContentByCvid(c, cvidInfo.Cvid); err != nil {
			return err
		}
		// 调用专栏提供接口（解专栏和笔记的关系）
		if err := s.artDao.UnbindArticleNote(c, cvidInfo.Cvid); err != nil {
			return err
		}
		// 删除泰山表中相应记录
		if err := s.artDao.UnbindNoteReplyTaishan(c, cvidInfo.Cvid, msg.RpID); err != nil {
			return err
		}
	case article.DelFromUser:
		// 删除泰山中的相应记录
		if err := s.artDao.UnbindNoteReplyTaishan(c, cvidInfo.Cvid, msg.RpID); err != nil {
			return err
		}
	default:
		log.Warn("Deleting the source is not in the processing scope msg:(%v)", msg)
		return
	}
	return
}
