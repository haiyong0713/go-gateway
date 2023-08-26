package service

import (
	"context"
	"fmt"
	"sync"

	accwarden "git.bilibili.co/bapis/bapis-go/account/service"
	acpAPI "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/admin/model"

	errGroup "go-common/library/sync/errgroup.v2"
)

// WhiteInfo.
func (s *Service) WhiteInfo(c context.Context, id int64) (white *model.EsArchiveWhite, err error) {
	var gameBase, matchBase []*model.BaseInfo
	white = new(model.EsArchiveWhite)
	if err = s.dao.DB.Where("id=?", id).First(&white).Error; err != nil {
		log.Error("WhiteInfo Error (%v)", err)
		return
	}
	var (
		infoReply *accwarden.InfoReply
		ip        = metadata.String(c, metadata.RemoteIP)
	)
	if infoReply, err = s.accClient.Info3(c, &accwarden.MidReq{Mid: white.Mid, RealIp: ip}); err != nil || infoReply == nil {
		log.Error("WhiteInfo 账号Infos:grpc错误 s.accClient.Info mid(%+v) error(%v)", white.Mid, err)
		return
	}
	white.Uname = infoReply.Info.Name
	group := errGroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		gameBase = make([]*model.BaseInfo, 0)
		if white.GameIDs != "" {
			var (
				gids  []int64
				games []*model.Game
			)
			if gids, err = xstr.SplitInts(white.GameIDs); err != nil {
				log.Error("WhiteInfo xstr.SplitInts GameIDS(%s) Error (%v)", white.GameIDs, err)
				return err
			}
			if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN(?)", unique(gids)).Find(&games).Error; err != nil {
				log.Error("WhiteInfo games Error (%v)", err)
				return err
			}
			for _, game := range games {
				gameBase = append(gameBase, &model.BaseInfo{ID: game.ID, Name: game.Title})
			}
		}
		white.Games = gameBase
		return nil
	})
	group.Go(func(ctx context.Context) error {
		matchBase = make([]*model.BaseInfo, 0)
		if white.MatchIDs != "" {
			var (
				matchIDs []int64
				matchs   []*model.Match
			)
			if matchIDs, err = xstr.SplitInts(white.MatchIDs); err != nil {
				log.Error("WhiteInfo xstr.SplitInts matchIDs(%s) Error (%v)", white.MatchIDs, err)
				return err
			}
			if err = s.dao.DB.Model(&model.Match{}).Where("status=?", _statusOn).Where("id IN(?)", unique(matchIDs)).Find(&matchs).Error; err != nil {
				log.Error("WhiteInfo maths Error (%v)", err)
				return err
			}
			for _, match := range matchs {
				matchBase = append(matchBase, &model.BaseInfo{ID: match.ID, Name: match.Title})
			}
		}
		white.Matchs = matchBase
		return nil
	})
	err = group.Wait()
	return
}

// WhiteList.
func (s *Service) WhiteList(c context.Context, pn, ps int64, mid string) (list []*model.EsArchiveWhite, count int64, err error) {
	source := s.dao.DB.Model(&model.EsArchiveWhite{})
	source = source.Where("is_deleted=?", _notDeleted)
	if mid != "" {
		source = source.Where("mid=?", mid)
	}
	source.Count(&count)
	if err = source.Offset((pn - 1) * ps).Limit(ps).Find(&list).Error; err != nil {
		log.Error("WhiteList Error (%v)", err)
		return
	}
	var (
		mids       []int64
		infosReply *accwarden.InfosReply
		ip         = metadata.String(c, metadata.RemoteIP)
	)
	if len(list) == 0 {
		return
	}
	for _, v := range list {
		mids = append(mids, v.Mid)
	}
	if infosReply, err = s.accClient.Infos3(c, &accwarden.MidsReq{Mids: unique(mids), RealIp: ip}); err != nil || infosReply == nil {
		log.Error("WhiteList 账号Infos3:grpc错误 s.accClient.Infos3 mids(%+v) error(%v)", mids, err)
		return
	}
	group := errGroup.WithContext(c)
	for _, white := range list {
		if user, ok := infosReply.Infos[white.Mid]; ok {
			white.Uname = user.Name
		}
		var gameBase, matchBase []*model.BaseInfo
		group.Go(func(ctx context.Context) error {
			gameBase = make([]*model.BaseInfo, 0)
			if white.GameIDs != "" {
				var (
					gids  []int64
					games []*model.Game
				)
				if gids, err = xstr.SplitInts(white.GameIDs); err != nil {
					log.Error("WhiteList xstr.SplitInts GameIDS(%s) Error (%v)", white.GameIDs, err)
					return err
				}
				if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN(?)", unique(gids)).Find(&games).Error; err != nil {
					log.Error("WhiteList games Error (%v)", err)
					return err
				}
				for _, game := range games {
					gameBase = append(gameBase, &model.BaseInfo{ID: game.ID, Name: game.Title})
				}
			}
			white.Games = gameBase
			return nil
		})
		group.Go(func(ctx context.Context) error {
			matchBase = make([]*model.BaseInfo, 0)
			if white.MatchIDs != "" {
				var (
					matchIDs []int64
					matchs   []*model.Match
				)
				if matchIDs, err = xstr.SplitInts(white.MatchIDs); err != nil {
					log.Error("WhiteInfo xstr.SplitInts matchIDs(%s) Error (%v)", white.MatchIDs, err)
					return err
				}
				if err = s.dao.DB.Model(&model.Match{}).Where("status=?", _statusOn).Where("id IN(?)", unique(matchIDs)).Find(&matchs).Error; err != nil {
					log.Error("WhiteInfo maths Error (%v)", err)
					return err
				}
				for _, match := range matchs {
					matchBase = append(matchBase, &model.BaseInfo{ID: match.ID, Name: match.Title})
				}
			}
			white.Matchs = matchBase
			return nil
		})
		err = group.Wait()
	}
	return
}

// AddWhite.
func (s *Service) AddWhite(c context.Context, param *model.EsArchiveWhite) (err error) {
	var block bool
	if block, err = s.ACPInfo(c, param.Mid); err != nil {
		log.Error("WhiteImport s.ACPInfo mid(%+v) error(%v)", param.Mid, err)
		return
	}
	if block {
		return fmt.Errorf("用户被封禁")
	}
	preData := new(model.EsArchiveWhite)
	s.dao.DB.Where("mid=?", param.Mid).First(&preData)
	if preData.ID > 0 {
		return fmt.Errorf("用户重复")
	}
	if err = s.dao.DB.Model(&model.EsArchiveWhite{}).Create(param).Error; err != nil {
		log.Error("AddWhite s.dao.DB.Model Create(%+v) error(%v)", param, err)
	}
	return
}

// EditWhite.
func (s *Service) EditWhite(c context.Context, param *model.EsArchiveWhite) (err error) {
	if param.ID <= 0 {
		return fmt.Errorf("id不能为空")
	}
	var block bool
	if block, err = s.ACPInfo(c, param.Mid); err != nil {
		log.Error("WhiteImport s.ACPInfo mid(%+v) error(%v)", param.Mid, err)
		return
	}
	if block {
		return fmt.Errorf("用户被封禁")
	}
	preData := new(model.EsArchiveWhite)
	s.dao.DB.Where("id != ?", param.ID).Where("mid = ?", param.Mid).First(&preData)
	if preData.ID > 0 {
		return fmt.Errorf("用户重复")
	}
	if err = s.dao.DB.Model(&model.EsArchiveWhite{}).Save(param).Error; err != nil {
		log.Error("EditWhite s.dao.DB.Model Update(%+v) error(%v)", param, err)
	}
	return
}

func (s *Service) DelWhite(c context.Context, ids []int64) (err error) {
	return s.dao.DB.Where("id IN (?)", ids).Delete(&model.EsArchiveWhite{}).Error
}

func (s *Service) WhiteImport(c context.Context, listParam []*model.WhiteImportParam) (err error) {
	var (
		mids []int64
		list []*model.WhiteImportParam
	)
	if len(listParam) == 0 {
		return
	}
	for _, v := range listParam {
		mids = append(mids, v.Mid)
	}
	if list, err = s.ACPInfos(c, listParam); err != nil {
		log.Error("WhiteImport s.ACPInfos mids(%+v) error(%v)", mids, err)
		return
	}
	sql, sqlParam := model.WhiteBatchAddSQL(list)
	if err = s.dao.DB.Model(&model.EsArchiveWhite{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Error("WhiteImport s.dao.DB.Model Exec(%+v) error(%v)", mids, err)
	}
	return
}

// ACPInfos is block infos from account control plane service.
func (s *Service) ACPInfos(c context.Context, importParams []*model.WhiteImportParam) ([]*model.WhiteImportParam, error) {
	var result []*model.WhiteImportParam
	wg := errGroup.WithCancel(c)
	var mutex sync.Mutex
	for _, importP := range importParams {
		mid := importP.Mid
		tmpImportT := importP
		wg.Go(func(ctx context.Context) error {
			reply, err := s.ACPClient.HasControlRole(c, &acpAPI.HasControlRoleReq{Mid: mid, ControlRole: []string{model.MainBlockRole, model.AllBlockRole}})
			if err != nil {
				return err
			}
			blockInfo := &model.BlockInfo{}
			blockInfo.FromControlRoleToBlockInfo(reply)
			mutex.Lock()
			if blockInfo.BlockStatus == model.BlockStatusFalse {
				result = append(result, tmpImportT)
			}
			mutex.Unlock()
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// ACPInfo is block info from account control plane service.
func (s *Service) ACPInfo(c context.Context, mid int64) (bool, error) {
	var result bool
	reply, err := s.ACPClient.HasControlRole(c, &acpAPI.HasControlRoleReq{Mid: mid, ControlRole: []string{model.MainBlockRole, model.AllBlockRole}})
	if err != nil {
		return result, err
	}
	blockInfo := &model.BlockInfo{}
	blockInfo.FromControlRoleToBlockInfo(reply)
	result = blockInfo.BlockStatus != model.BlockStatusFalse
	return result, nil
}
