package article

import (
	"context"

	"go-common/library/ecode"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	artgrpc "git.bilibili.co/bapis/bapis-go/article/service"
	"github.com/pkg/errors"
)

func (d *Dao) PubNoteInfo(c context.Context, cvid int64) (*notegrpc.PublishNoteInfoReply, error) {
	req := &notegrpc.PublishNoteInfoReq{
		Cvid: cvid,
	}
	reply, err := d.noteClient.PublishNoteInfo(c, req)
	if err != nil {
		return nil, errors.Wrapf(err, "PubNoteInfo req(%+v)", req)
	}
	if reply == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "PubNoteInfo req(%+v)", req)
	}
	return reply, nil
}

func (d *Dao) DelUpArticles(c context.Context, cvids []int64, mid int64) error {
	req := &artgrpc.DelUpArticlesReq{
		Mid:  mid,
		Aids: cvids,
	}
	if _, err := d.artClient.DelUpArticles(c, req); err != nil {
		return errors.Wrapf(err, "DelUpArticles req(%+v)", req)
	}
	return nil
}

func (d *Dao) SimpleArticles(c context.Context, cvids []int64) (map[int64]*notegrpc.SimpleArticleCard, error) {
	req := &notegrpc.SimpleArticlesReq{Cvids: cvids}
	res, err := d.noteClient.SimpleArticles(c, req)
	if err != nil {
		return nil, errors.Wrapf(err, "SimpleArticles req(%+v)", req)
	}
	if res == nil || res.Items == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "SimpleArticles req(%+v)", req)
	}
	return res.Items, nil
}
