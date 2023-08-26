package article

import (
	"context"

	"go-common/library/log"

	articleMdl "git.bilibili.co/bapis/bapis-go/article/model"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
)

func (d *Dao) ArticleMetas(c context.Context, ids []int64) (map[int64]*articleMdl.Meta, error) {
	rsp, err := d.articleGRPC.ArticleMetas(c, &articlegrpc.ArticleMetasReq{Ids: ids})
	if err != nil {
		log.Errorc(c, "Dao.ArticleMetas(ids: %+v) failed. error(%+v)", ids, err)
		return nil, err
	}
	return rsp.GetRes(), nil
}
