package dao

import (
	"context"

	"go-common/library/ecode"

	articlemdl "git.bilibili.co/bapis/bapis-go/article/model"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"

	"github.com/pkg/errors"
)

func (d *Dao) Articles(c context.Context, aids []int64) (map[int64]*articlemdl.Meta, error) {
	arg := &articlegrpc.ArticleMetasReq{Ids: aids, From: 10}
	res, err := d.artClient.ArticleMetas(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", arg)
	}
	if res.GetRes() == nil {
		return nil, ecode.NothingFound
	}
	return res.GetRes(), nil
}
