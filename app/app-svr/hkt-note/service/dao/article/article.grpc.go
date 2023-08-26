package article

import (
	"context"

	"go-common/library/ecode"

	artgm "git.bilibili.co/bapis/bapis-go/article/model"
	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	artgrpc "git.bilibili.co/bapis/bapis-go/article/service"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"github.com/pkg/errors"
)

const _thumbupBusiness = "article"

func (d *Dao) ArticleAudits(c context.Context, cvids []int64) (map[int64]*artmdl.ArticleAudit, error) {
	req := &artgrpc.ArticleAuditsReq{
		Aids: cvids,
	}
	res, err := d.artClient.ArticleAudits(c, req)
	if err != nil {
		return nil, errors.Wrapf(err, "ArticleAudits req(%+v)", req)
	}
	if res == nil || res.Articles == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "ArticleAudits req(%+v)", req)
	}
	return res.Articles, nil
}

func (d *Dao) ArticleMetasSimple(c context.Context, cvids []int64) (map[int64]*artgm.Meta, error) {
	req := &artgrpc.ArticleMetasSimpleReq{
		Ids: cvids,
	}
	res, err := d.artClient.ArticleMetasSimple(c, req)
	if err != nil {
		return nil, errors.Wrapf(err, "ArticleMetasSimple req(%+v)", req)
	}
	if res == nil || res.Res == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "ArticleMetasSimple req(%+v)", req)
	}
	return res.Res, nil
}

func (d *Dao) HasLike(c context.Context, cvids []int64, mid int64) (map[int64]*thumbupgrpc.UserLikeState, error) {
	req := &thumbupgrpc.HasLikeReq{
		Business:   _thumbupBusiness,
		MessageIds: cvids,
		Mid:        mid,
	}
	res, err := d.thumbupClient.HasLike(c, req)
	if err != nil {
		return nil, errors.Wrapf(err, "HasLike req(%+v)", req)
	}
	if res == nil || res.States == nil {
		return make(map[int64]*thumbupgrpc.UserLikeState), nil
	}
	return res.States, nil
}
