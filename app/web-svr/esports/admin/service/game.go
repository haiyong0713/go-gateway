package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/admin/model"
)

var _emptyGameList = make([]*model.Game, 0)

// GameInfo .
func (s *Service) GameInfo(c context.Context, id int64) (game *model.Game, err error) {
	game = new(model.Game)
	if err = s.dao.DB.Where("id=?", id).First(&game).Error; err != nil {
		log.Error("GameInfo Error (%v)", err)
	}
	return
}

// GameList .
func (s *Service) GameList(c context.Context, pn, ps int64, title string) (list []*model.Game, count int64, err error) {
	source := s.dao.DB.Model(&model.Game{})
	if title != "" {
		source = source.Where("title like ?", "%"+title+"%")
	}
	source.Count(&count)
	if err = source.Offset((pn - 1) * ps).Order("rank DESC,id ASC").Limit(ps).Find(&list).Error; err != nil {
		log.Error("GameList Error (%v)", err)
	}
	return
}

// AddGame .
func (s *Service) AddGame(c context.Context, param *model.Game) (err error) {
	// TODO check name exist
	if err = s.dao.DB.Model(&model.Game{}).Create(param).Error; err != nil {
		log.Error("AddGame s.dao.DB.Model Create(%+v) error(%v)", param, err)
	}
	return
}

// EditGame .
func (s *Service) EditGame(c context.Context, param *model.Game) (err error) {
	preGame := new(model.Game)
	if err = s.dao.DB.Where("id=?", param.ID).First(&preGame).Error; err != nil {
		log.Error("EditGame s.dao.DB.Where id(%d) error(%d)", param.ID, err)
		return
	}
	if err = s.dao.DB.Model(&model.Game{}).Save(param).Error; err != nil {
		log.Error("EditGame s.dao.DB.Model Update(%+v) error(%v)", param, err)
	}
	return
}

// ForbidGame .
func (s *Service) ForbidGame(c context.Context, id int64, state int) (err error) {
	preGame := new(model.Game)
	if err = s.dao.DB.Where("id=?", id).First(&preGame).Error; err != nil {
		log.Error("GameForbid s.dao.DB.Where id(%d) error(%d)", id, err)
		return
	}
	if err = s.dao.DB.Model(&model.Game{}).Where("id=?", id).Update(map[string]int{"status": state}).Error; err != nil {
		log.Error("GameForbid s.dao.DB.Model error(%v)", err)
	}
	return
}

// gameList return game info map with oid key.
func (s *Service) gameList(typ int, oids []int64) (list map[int64][]*model.Game, err error) {
	var (
		gidMaps []*model.GIDMap
		gids    []int64
		games   []*model.Game
	)
	if len(oids) == 0 {
		return
	}
	if err = s.dao.DB.Model(&model.GIDMap{}).Where("is_deleted=?", _notDeleted).Where("type=?", typ).Where("oid IN(?)", oids).Find(&gidMaps).Error; err != nil {
		log.Error("gameList gidMap Error (%v)", err)
		return
	}
	if len(gidMaps) == 0 {
		return
	}
	gidMap := make(map[int64]int64, len(gidMaps))
	oidGidMap := make(map[int64][]int64, len(gidMaps))
	for _, v := range gidMaps {
		oidGidMap[v.Oid] = append(oidGidMap[v.Oid], v.Gid)
		if _, ok := gidMap[v.Gid]; ok {
			continue
		}
		gids = append(gids, v.Gid)
		gidMap[v.Gid] = v.Gid
	}
	if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN(?)", gids).Find(&games).Error; err != nil {
		log.Error("gameList games Error (%v)", err)
		return
	}
	if len(games) == 0 {
		return
	}
	gameMap := make(map[int64]*model.Game, len(games))
	for _, v := range games {
		gameMap[v.ID] = v
	}
	list = make(map[int64][]*model.Game, len(oids))
	for _, oid := range oids {
		if ids, ok := oidGidMap[oid]; ok {
			for _, id := range ids {
				if game, ok := gameMap[id]; ok {
					list[oid] = append(list[oid], game)
				}
			}
		}
	}
	return
}

// Types return data page game types.
func (s *Service) Types(c context.Context) (list map[int64]string, err error) {
	list = make(map[int64]string, len(s.c.GameTypes))
	for _, tp := range s.c.GameTypes {
		list[tp.ID] = tp.Name
	}
	return
}

// GameTeams .
func (s *Service) GameTeams(c context.Context, pn, ps int64, id int64, title string) (list []*model.Team, count int64, err error) {
	var (
		gidMaps []*model.GIDMap
		tids    []int64
	)
	if err = s.dao.DB.Model(&model.GIDMap{}).Where("is_deleted=?", _notDeleted).Where("type=?", model.TypeTeam).Where("gid=?", id).Find(&gidMaps).Error; err != nil {
		log.Error("GameTeams gidMap gid(%d) Error (%v)", id, err)
		return
	}
	if len(gidMaps) == 0 {
		list = _emptyTeamList
		return
	}
	for _, v := range gidMaps {
		tids = append(tids, v.Oid)
	}
	source := s.dao.DB.Model(&model.Team{})
	source.Order("id ASC")
	source = source.Where("is_deleted=?", _notDeleted).Where("id IN(?)", tids)
	if title != "" {
		source = source.Where("title like ?", "%"+title+"%")
	}
	source.Count(&count)
	if count == 0 {
		list = _emptyTeamList
		return
	}
	if err = source.Offset((pn - 1) * ps).Limit(ps).Find(&list).Error; err != nil {
		log.Error("GameTeams find  Error (%v)", err)
	}
	return
}

// GameSeasons .
func (s *Service) GameSeasons(c context.Context, pn, ps int64, id int64, title string) (list []*model.Season, count int64, err error) {
	var (
		gidMaps []*model.GIDMap
		sids    []int64
	)
	source := s.dao.DB.Model(&model.Season{})
	source = source.Order("id ASC")
	source = source.Where("status=?", _notDeleted)
	if id > 0 {
		if err = s.dao.DB.Model(&model.GIDMap{}).Where("is_deleted=?", _notDeleted).Where("type=?", model.TypeSeason).Where("gid=?", id).Find(&gidMaps).Error; err != nil {
			log.Error("GameSeasons gidMap gid(%d) Error (%v)", id, err)
			return
		}
		if len(gidMaps) == 0 {
			list = _emptySeasonList
			return
		}
		for _, v := range gidMaps {
			sids = append(sids, v.Oid)
		}
		source = source.Where("id IN(?)", sids)
	}
	if title != "" {
		source = source.Where("title like ?", "%"+title+"%")
	}
	source.Count(&count)
	if count == 0 {
		list = _emptySeasonList
		return
	}
	if err = source.Offset((pn - 1) * ps).Limit(ps).Find(&list).Error; err != nil {
		log.Error("GameSeasons find  Error (%v)", err)
	}
	return
}
