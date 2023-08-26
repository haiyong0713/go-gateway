package dao

import (
	"context"

	articlemdl "git.bilibili.co/bapis/bapis-go/article/model"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
)

type articleDao struct {
	client articlegrpc.ArticleGRPCClient
}

func (d *articleDao) ArticleMetas(c context.Context, cvids []int64, from int32) (map[int64]*articlemdl.Meta, error) {
	req := &articlegrpc.ArticleMetasReq{Ids: cvids, From: from}
	rly, err := d.client.ArticleMetas(c, req)
	if err != nil {
		return nil, err
	}
	return rly.Res, nil
}
