package service

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/admin/model"

	errGroup "go-common/library/sync/errgroup.v2"
)

// AutotagInfo.
func (s *Service) AutotagInfo(c context.Context, id int64) (autotag *model.EsArchiveTag, err error) {
	var gameBase, matchBase []*model.BaseInfo
	autotag = new(model.EsArchiveTag)
	if err = s.dao.DB.Where("id=?", id).First(&autotag).Error; err != nil {
		log.Error("AutotagInfo Error (%v)", err)
		return
	}
	group := errGroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		gameBase = make([]*model.BaseInfo, 0)
		if autotag.GameIDs != "" {
			var (
				gids  []int64
				games []*model.Game
			)
			if gids, err = xstr.SplitInts(autotag.GameIDs); err != nil {
				log.Error("AutotagInfo xstr.SplitInts GameIDS(%s) Error (%v)", autotag.GameIDs, err)
				return err
			}
			if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN (?)", unique(gids)).Find(&games).Error; err != nil {
				log.Error("AutotagInfo games Error (%v)", err)
				return err
			}
			for _, game := range games {
				gameBase = append(gameBase, &model.BaseInfo{ID: game.ID, Name: game.Title})
			}
		}
		autotag.Games = gameBase
		return nil
	})
	group.Go(func(ctx context.Context) error {
		matchBase = make([]*model.BaseInfo, 0)
		if autotag.MatchIDs != "" {
			var (
				matchIDs []int64
				matchs   []*model.Match
			)
			if matchIDs, err = xstr.SplitInts(autotag.MatchIDs); err != nil {
				log.Error("AutotagInfo xstr.SplitInts matchIDs(%s) Error (%v)", autotag.MatchIDs, err)
				return err
			}
			if err = s.dao.DB.Model(&model.Match{}).Where("status=?", _statusOn).Where("id IN (?)", unique(matchIDs)).Find(&matchs).Error; err != nil {
				log.Error("AutotagInfo maths Error (%v)", err)
				return err
			}
			for _, match := range matchs {
				matchBase = append(matchBase, &model.BaseInfo{ID: match.ID, Name: match.Title})
			}
		}
		autotag.Matchs = matchBase
		return nil
	})
	err = group.Wait()
	return
}

// AutotagList.
func (s *Service) AutotagList(c context.Context, pn, ps int64, tag string) (list []*model.EsArchiveTag, count int64, err error) {
	source := s.dao.DB.Model(&model.EsArchiveTag{})
	source = source.Where("is_deleted=?", _notDeleted)
	if tag != "" {
		source = source.Where("tag=?", tag)
	}
	source.Count(&count)
	if err = source.Offset((pn - 1) * ps).Limit(ps).Find(&list).Error; err != nil {
		log.Error("AutotagList Error (%v)", err)
		return
	}
	group := errGroup.WithContext(c)
	for _, Autotag := range list {
		var gameBase, matchBase []*model.BaseInfo
		group.Go(func(ctx context.Context) error {
			gameBase = make([]*model.BaseInfo, 0)
			if Autotag.GameIDs != "" {
				var (
					gids  []int64
					games []*model.Game
				)
				if gids, err = xstr.SplitInts(Autotag.GameIDs); err != nil {
					log.Error("AutotagList xstr.SplitInts GameIDS(%s) Error (%v)", Autotag.GameIDs, err)
					return err
				}
				if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN(?)", unique(gids)).Find(&games).Error; err != nil {
					log.Error("AutotagList games Error (%v)", err)
					return err
				}
				for _, game := range games {
					gameBase = append(gameBase, &model.BaseInfo{ID: game.ID, Name: game.Title})
				}
			}
			Autotag.Games = gameBase
			return nil
		})
		group.Go(func(ctx context.Context) error {
			matchBase = make([]*model.BaseInfo, 0)
			if Autotag.MatchIDs != "" {
				var (
					matchIDs []int64
					matchs   []*model.Match
				)
				if matchIDs, err = xstr.SplitInts(Autotag.MatchIDs); err != nil {
					log.Error("AutotagList xstr.SplitInts matchIDs(%s) Error (%v)", Autotag.MatchIDs, err)
					return err
				}
				if err = s.dao.DB.Model(&model.Match{}).Where("status=?", _statusOn).Where("id IN(?)", unique(matchIDs)).Find(&matchs).Error; err != nil {
					log.Error("AutotagList maths Error (%v)", err)
					return err
				}
				for _, match := range matchs {
					matchBase = append(matchBase, &model.BaseInfo{ID: match.ID, Name: match.Title})
				}
			}
			Autotag.Matchs = matchBase
			return nil
		})
		err = group.Wait()
	}
	return
}

// AddAutotag.
func (s *Service) AddAutotag(c context.Context, param *model.EsArchiveTag) (err error) {
	preData := new(model.EsArchiveTag)
	s.dao.DB.Where("tag=?", param.Tag).First(&preData)
	if preData.ID > 0 {
		return fmt.Errorf("tag名称重复")
	}
	if err = s.dao.DB.Model(&model.EsArchiveTag{}).Create(param).Error; err != nil {
		log.Error("AddAutotag s.dao.DB.Model Create(%+v) error(%v)", param, err)
	}
	return
}

// EditAutotag.
func (s *Service) EditAutotag(c context.Context, param *model.EsArchiveTag) (err error) {
	if param.ID <= 0 {
		return fmt.Errorf("id不能为空")
	}
	preData := new(model.EsArchiveTag)
	s.dao.DB.Where("id != ?", param.ID).Where("tag = ?", param.Tag).First(&preData)
	if preData.ID > 0 {
		return fmt.Errorf("用户重复")
	}
	if err = s.dao.DB.Model(&model.EsArchiveTag{}).Save(param).Error; err != nil {
		log.Error("EditAutotag s.dao.DB.Model Update(%+v) error(%v)", param, err)
	}
	return
}

func (s *Service) DelAutotag(c context.Context, ids []int64) (err error) {
	return s.dao.DB.Where("id IN (?)", ids).Delete(&model.EsArchiveTag{}).Error
}

func (s *Service) AutotagImport(c context.Context, list []*model.TagImportParam) (err error) {
	if len(list) == 0 {
		return
	}
	sql, sqlParam := model.TagBatchAddSQL(list)
	if err = s.dao.DB.Model(&model.EsArchiveTag{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Error("AutotagImport s.dao.DB.Model Exec(%+v) error(%v)", list, err)
	}
	return
}
