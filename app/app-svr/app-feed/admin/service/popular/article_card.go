package popular

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/article"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	_addArticleCard  = "add"
	_editArticleCard = "up"
)

func (s *Service) ArticleCardSave(ctx context.Context, param *show.ArticleCard, uname string, uid int64) (err error) {
	obj := map[string]interface{}{
		"article_id": param.ArticleID,
	}
	if param.ID == 0 {
		if param.ID, err = s.showDao.ArticleCardAdd(ctx, &show.ArticleCardAD{
			ArticleID: param.ArticleID,
			Cover:     param.Cover,
			CreateBy:  param.CreateBy,
		}); err != nil {
			log.Error("[ArticleCardSave] s.ArticleCardSave() error(%v)", err)
			return
		}
		if err = util.AddLogs(common.LogArticleCard, uname, uid, param.ID, _addArticleCard, obj); err != nil {
			log.Error("[ArticleCardSave] AddLogs error(%v)", err)
		}
	}
	if err = s.showDao.ArticleCardUpdate(ctx, &show.ArticleCardUP{
		ID:        param.ID,
		ArticleID: param.ArticleID,
		Cover:     param.Cover,
	}); err != nil {
		log.Error("[ArticleCardSave] s.ArticleCardSave() error(%v)", err)
		return
	}
	if err = util.AddLogs(common.LogArticleCard, uname, uid, param.ID, _editArticleCard, obj); err != nil {
		log.Error("[ArticleCardEdit] AddLogs error(%v)", err)
	}
	return
}

func (s *Service) ArticleCardList(ctx context.Context, id int64, state int, createBy string, pn, ps int) (res *show.ArticleCardRes, err error) {
	if res, err = s.showDao.ArticleCardList(ctx, id, state, createBy, pn, ps); err != nil {
		log.Error("[ArticleCardList] s.ArticleCardList() error(%v)", err)
		return
	}
	var (
		artID []int64
		artm  map[int64]*article.ArticleInfo
	)
	for _, item := range res.Items {
		artID = append(artID, item.ArticleID)
	}
	if len(artID) != 0 {
		var err error
		if artm, err = s.artDao.ArticlesInfo(ctx, artID); err != nil {
			log.Error("[ArticlesInfo] s.ArticlesInfo() error(%v)", err)
		}
	}
	for _, item := range res.Items {
		item.MtimeStr = item.Mtime.Format(_timeFormat)
		if art, ok := artm[item.ArticleID]; ok {
			item.ArticleTitle = art.Title
		}
	}
	return
}

func (s *Service) ArticleCardOperate(ctx context.Context, id int64, state int, uname string, uid int64) (err error) {
	if err = s.showDao.ArticleCardOperate(ctx, id, state); err != nil {
		log.Error("[ArticleCardOperate] s.ArticleCardOperate() id(%d) error(%v)", id, err)
		return
	}
	obj := map[string]interface{}{
		"state": state,
	}
	if err = util.AddLogs(common.LogArticleCard, uname, uid, id, _editArticleCard, obj); err != nil {
		log.Error("[ArticleCardOperate] AddLogs error(%v)", err)
	}
	return
}
