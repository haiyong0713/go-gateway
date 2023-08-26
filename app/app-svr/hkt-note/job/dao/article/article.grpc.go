package article

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/job/model/article"

	artgrpc "git.bilibili.co/bapis/bapis-go/article/service"
	"github.com/pkg/errors"
)

func (d *Dao) CreateArticle(c context.Context, argArticle *artgrpc.ArgArticle, noteId int64) (int64, *article.PubFailMsg, error) {
	req := &artgrpc.CreateArticleReq{
		ArgArticle: argArticle,
		Source:     artgrpc.CreateArticleSource_NOTE,
	}
	reply, err := d.articleClient.CreateArticle(c, req)
	if err != nil {
		log.Error("ArtError CreateArticle argArticle(%+v) noteId(%d) err(%+v)", argArticle, noteId, err)
		if ecode.EqualError(ecode.Deadline, err) || ecode.EqualError(ecode.ServerErr, err) {
			return 0, nil, errors.Wrapf(err, "CreateArticle req(%+v)", req)
		}
		pubFailMsg := &article.PubFailMsg{
			NoteId:  noteId,
			Mid:     argArticle.Mid,
			Reason:  ecode.Cause(err).Message(),
			ErrCode: ecode.Cause(err).Code(),
		}
		return 0, pubFailMsg, errors.Wrapf(err, "CreateArticle req(%+v)", req)
	}
	if reply == nil || reply.Aid == 0 {
		pubFailMsg := &article.PubFailMsg{
			NoteId:  noteId,
			Mid:     argArticle.Mid,
			Reason:  "获取专栏id失败",
			ErrCode: int(ecode.NothingFound),
		}
		return 0, pubFailMsg, errors.Wrapf(ecode.NothingFound, "CreateArticle req(%+v)", req)
	}
	return reply.Aid, nil, nil
}

func (d *Dao) EditArticle(c context.Context, argArticle *artgrpc.ArgArticle, noteId int64) (*article.PubFailMsg, error) {
	req := &artgrpc.EditArticleReq{
		ArgArticle: argArticle,
		Source:     artgrpc.CreateArticleSource_NOTE,
	}
	if _, err := d.articleClient.EditArticle(c, req); err != nil {
		log.Error("ArtError EditArticle argArticle(%+v) noteId(%d) err(%+v)", argArticle, noteId, err)
		if err == ecode.Deadline || err == ecode.ServerErr {
			return nil, errors.Wrapf(err, "EditArticle req(%+v)", req)
		}
		pubFailMsg := &article.PubFailMsg{
			NoteId:  noteId,
			Mid:     argArticle.Mid,
			Reason:  ecode.Cause(err).Message(),
			ErrCode: ecode.Cause(err).Code(),
		}
		return pubFailMsg, errors.Wrapf(err, "EditArticle req(%+v)", req)
	}
	return nil, nil
}

func (d *Dao) UnbindArticleNote(c context.Context, cvid int64) error {
	req := &artgrpc.UnbindArticleNoteReq{
		ArticleId: cvid,
	}
	if _, err := d.articleClient.UnbindArticleNote(c, req); err != nil {
		log.Errorc(c, "ArtError UnbindArticleNote cvid(%d) err(%v)", cvid, err)
		return err
	}
	return nil
}
