package article

import (
	"context"
	"strconv"

	"go-common/library/ecode"
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"go-gateway/app/app-svr/hkt-note/job/model/article"
	"go-gateway/app/app-svr/hkt-note/job/model/note"

	frontgrpc "git.bilibili.co/bapis/bapis-go/frontend/bilinote/v1"
	"github.com/pkg/errors"
)

func (d *Dao) GetBiliNoteContent(c context.Context, content string, msg *note.NtPubMsg) (*frontgrpc.NoteReply, *article.PubFailMsg, error) {
	req := &frontgrpc.NoteReq{
		BiliJson: content,
		Mid:      msg.Mid,
	}
	reply, err := d.frontendClient.GetBiliNoteContent(c, req)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "GetBiliNoteContent req(%+v)", req)
	}
	if reply == nil {
		return nil, nil, errors.Wrapf(ecode.NothingFound, "GetBiliNoteContent req(%+v)", req)
	}
	// 传参错误/img服务错误，重试无效
	if reply.Status != "200" {
		errCode, _ := strconv.ParseInt(reply.Status, 10, 64)
		pubFailMsg := &article.PubFailMsg{
			NoteId:  msg.NoteId,
			Mid:     msg.Mid,
			Reason:  reply.Message,
			ErrCode: int(errCode),
		}
		return nil, pubFailMsg, errors.Wrapf(xecode.NoteFrontEndWrong, "GetBiliNoteContent req(%+v) reply(%+v)", req, reply)
	}
	return reply, nil, nil
}
