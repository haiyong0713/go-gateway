package common

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/article"
)

// Article .
func (s *Service) Article(c context.Context, ids []int64) (article *article.Article, err error) {
	if article, err = s.articleDao.Article(c, ids); err != nil {
		if err == ecode.NothingFound {
			err = fmt.Errorf("id错误，没有专栏相关信息！")
			return
		}
		log.Error("common.Lives param(%q)error %v", ids, err)
		return
	}
	if article == nil {
		err = fmt.Errorf("id错误，没有专栏相关信息！")
		return
	}
	return
}
