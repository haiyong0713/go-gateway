package note

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/common"
)

func (s *Service) recordNoteReplyFormat(ctx context.Context, noteId int64, replyComment int32) (err error) {
	taiShanKey := fmt.Sprintf(common.Note_Reply_Format_Taishan_Key, noteId)
	var value []byte
	if replyComment == common.Note_Comment_Format_Type_New {
		curInfo := &common.TaishanNoteReplyFormatInfo{
			FormatType: common.Note_Comment_Format_Type_New,
		}
		value, err = json.Marshal(curInfo)
		if err != nil {
			log.Errorc(ctx, "recordNoteReplyFormat marshal err %v and noteId %v", err, noteId)
			return err
		}
	} else {
		//不传或者传旧视为old
		curInfo := &common.TaishanNoteReplyFormatInfo{
			FormatType: common.Note_Comment_Format_Type_Old,
		}
		value, err = json.Marshal(curInfo)
		if err != nil {
			log.Errorc(ctx, "recordNoteReplyFormat marshal err %v and noteId %v", err, noteId)
			return err
		}
	}
	err = s.artDao.PutTaishan(ctx, taiShanKey, value, common.TaishanConfig.NoteReply)
	return err
}
