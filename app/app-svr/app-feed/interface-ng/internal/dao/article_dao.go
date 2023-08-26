package dao

import (
	"context"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
)

type articleDao struct {
	article articlegrpc.ArticleGRPCClient
}

func (d *articleDao) Articles(ctx context.Context, aids []int64) (map[int64]*article.Meta, error) {
	req := &articlegrpc.ArticleMetasReq{Ids: aids, From: 1} // grpc 接口ArticleMetas from = 1 代表天马请求封面将返回窄图
	res, err := d.article.ArticleMetas(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.Res, nil
}
