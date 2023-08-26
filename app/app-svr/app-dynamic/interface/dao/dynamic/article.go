package dynamic

import (
	"context"

	articleMdl "git.bilibili.co/bapis/bapis-go/article/model"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
	"go-common/library/log"
)

func (d *Dao) ArticleMetasMc(c context.Context, ids []int64) (map[int64]*articleMdl.Meta, error) {
	rsp, err := d.articleGRPC.ArticleMetasMc(c, &articlegrpc.ArticleMetasReq{Ids: ids})
	if err != nil {
		log.Errorc(c, "Dao.ArticleMetasMc(ids: %+v) failed. error(%+v)", ids, err)
		return nil, err
	}
	return rsp.GetRes(), nil
}
