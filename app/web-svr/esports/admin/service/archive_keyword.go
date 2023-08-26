package service

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/admin/model"

	errGroup "go-common/library/sync/errgroup.v2"
)

// KeywordInfo.
func (s *Service) KeywordInfo(c context.Context, id int64) (keyword *model.EsArchiveKeyword, err error) {
	var gameBase, matchBase []*model.BaseInfo
	keyword = new(model.EsArchiveKeyword)
	if err = s.dao.DB.Where("id=?", id).First(&keyword).Error; err != nil {
		log.Error("KeywordInfo Error (%v)", err)
		return
	}
	group := errGroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		gameBase = make([]*model.BaseInfo, 0)
		if keyword.GameIDs != "" {
			var (
				gids  []int64
				games []*model.Game
			)
			if gids, err = xstr.SplitInts(keyword.GameIDs); err != nil {
				log.Error("KeywordInfo xstr.SplitInts GameIDS(%s) Error (%v)", keyword.GameIDs, err)
				return err
			}
			if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN(?)", unique(gids)).Find(&games).Error; err != nil {
				log.Error("KeywordInfo games Error (%v)", err)
				return err
			}
			for _, game := range games {
				gameBase = append(gameBase, &model.BaseInfo{ID: game.ID, Name: game.Title})
			}
		}
		keyword.Games = gameBase
		return nil
	})
	group.Go(func(ctx context.Context) error {
		matchBase = make([]*model.BaseInfo, 0)
		if keyword.MatchIDs != "" {
			var (
				matchIDs []int64
				matchs   []*model.Match
			)
			if matchIDs, err = xstr.SplitInts(keyword.MatchIDs); err != nil {
				log.Error("KeywordInfo xstr.SplitInts matchIDs(%s) Error (%v)", keyword.MatchIDs, err)
				return err
			}
			if err = s.dao.DB.Model(&model.Match{}).Where("status=?", _statusOn).Where("id IN(?)", unique(matchIDs)).Find(&matchs).Error; err != nil {
				log.Error("KeywordInfo maths Error (%v)", err)
				return err
			}
			for _, match := range matchs {
				matchBase = append(matchBase, &model.BaseInfo{ID: match.ID, Name: match.Title})
			}
		}
		keyword.Matchs = matchBase
		return nil
	})
	err = group.Wait()
	return
}

// KeywordList.
func (s *Service) KeywordList(c context.Context, pn, ps int64, keyword string) (list []*model.EsArchiveKeyword, count int64, err error) {
	source := s.dao.DB.Model(&model.EsArchiveKeyword{})
	source = source.Where("is_deleted=?", _notDeleted)
	if keyword != "" {
		source = source.Where("keyword=?", keyword)
	}
	source.Count(&count)
	if err = source.Offset((pn - 1) * ps).Limit(ps).Find(&list).Error; err != nil {
		log.Error("KeywordList Error (%v)", err)
		return
	}
	group := errGroup.WithContext(c)
	for _, Keyword := range list {
		var gameBase, matchBase []*model.BaseInfo
		group.Go(func(ctx context.Context) error {
			gameBase = make([]*model.BaseInfo, 0)
			if Keyword.GameIDs != "" {
				var (
					gids  []int64
					games []*model.Game
				)
				if gids, err = xstr.SplitInts(Keyword.GameIDs); err != nil {
					log.Error("KeywordList xstr.SplitInts GameIDS(%s) Error (%v)", Keyword.GameIDs, err)
					return err
				}
				if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN(?)", unique(gids)).Find(&games).Error; err != nil {
					log.Error("KeywordList games Error (%v)", err)
					return err
				}
				for _, game := range games {
					gameBase = append(gameBase, &model.BaseInfo{ID: game.ID, Name: game.Title})
				}
			}
			Keyword.Games = gameBase
			return nil
		})
		group.Go(func(ctx context.Context) error {
			matchBase = make([]*model.BaseInfo, 0)
			if Keyword.MatchIDs != "" {
				var (
					matchIDs []int64
					matchs   []*model.Match
				)
				if matchIDs, err = xstr.SplitInts(Keyword.MatchIDs); err != nil {
					log.Error("KeywordList xstr.SplitInts matchIDs(%s) Error (%v)", Keyword.MatchIDs, err)
					return err
				}
				if err = s.dao.DB.Model(&model.Match{}).Where("status=?", _statusOn).Where("id IN(?)", unique(matchIDs)).Find(&matchs).Error; err != nil {
					log.Error("KeywordList maths Error (%v)", err)
					return err
				}
				for _, match := range matchs {
					matchBase = append(matchBase, &model.BaseInfo{ID: match.ID, Name: match.Title})
				}
			}
			Keyword.Matchs = matchBase
			return nil
		})
		err = group.Wait()
	}
	return
}

// AddKeyword.
func (s *Service) AddKeyword(c context.Context, param *model.EsArchiveKeyword) (err error) {
	preData := new(model.EsArchiveKeyword)
	s.dao.DB.Where("Keyword=?", param.Keyword).First(&preData)
	if preData.ID > 0 {
		return fmt.Errorf("keyword重复")
	}
	if err = s.dao.DB.Model(&model.EsArchiveKeyword{}).Create(param).Error; err != nil {
		log.Error("AddKeyword s.dao.DB.Model Create(%+v) error(%v)", param, err)
	}
	return
}

// EditKeyword.
func (s *Service) EditKeyword(c context.Context, param *model.EsArchiveKeyword) (err error) {
	if param.ID <= 0 {
		return fmt.Errorf("id不能为空")
	}
	preData := new(model.EsArchiveKeyword)
	s.dao.DB.Where("id != ?", param.ID).Where("Keyword = ?", param.Keyword).First(&preData)
	if preData.ID > 0 {
		return fmt.Errorf("keyword重复")
	}
	if err = s.dao.DB.Model(&model.EsArchiveKeyword{}).Save(param).Error; err != nil {
		log.Error("EditKeyword s.dao.DB.Model Update(%+v) error(%v)", param, err)
	}
	return
}

func (s *Service) DelKeyword(c context.Context, ids []int64) (err error) {
	return s.dao.DB.Where("id IN (?)", ids).Delete(&model.EsArchiveKeyword{}).Error
}

func (s *Service) KeywordImport(c context.Context, list []*model.KeywordImportParam) (err error) {
	if len(list) == 0 {
		return
	}
	sql, sqlParam := model.KeywordBatchAddSQL(list)
	if err = s.dao.DB.Model(&model.EsArchiveKeyword{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Error("KeywordImport s.dao.DB.Model Exec(%+v) error(%v)", list, err)
	}
	return
}
