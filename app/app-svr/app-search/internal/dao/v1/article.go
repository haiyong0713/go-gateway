package v1

import (
	"context"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	artclient "git.bilibili.co/bapis/bapis-go/article/service"

	"github.com/pkg/errors"
)

func (d *dao) Articles(c context.Context, aids []int64) (arts map[int64]*article.Meta, err error) {
	arg := &artclient.ArticleMetasReq{Ids: aids}
	res, err := d.artClient.ArticleMetas(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%+v", arg)
		return
	}
	arts = res.Res
	return
}
