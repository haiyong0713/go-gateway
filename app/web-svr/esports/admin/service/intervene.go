package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/admin/model"
)

var (
	_emptyInterveneList   = make([]*model.EsSearchCard, 0)
	_emptyInterveneDetail = make([]*model.EsSearchContest, 0)
)

const (
	_firstContest = 0
	_maxDetail    = 3
)

// InterveneInfo .
func (s *Service) InterveneInfo(c context.Context, id int64) (rs *model.SearchInfo, err error) {
	var (
		intervenes []*model.EsSearchCard
		list       []*model.SearchInfo
	)
	main := new(model.EsSearchCard)
	if err = s.dao.DB.Where("id=?", id).First(&main).Error; err != nil {
		log.Error("InterveneInfo Error (%v)", err)
		return
	}
	intervenes = append(intervenes, main)
	if list, err = s.interveneInfos(intervenes); err != nil {
		log.Error("s.InterveneInfo Error (%v)", err)
	}
	if len(list) == 0 {
		err = ecode.NothingFound
		return
	}
	rs = list[0]
	return
}

// InterveneList .
func (s *Service) InterveneList(c context.Context, pn, ps, srt int64, title string) (list []*model.SearchInfo, count int64, err error) {
	var intervenes []*model.EsSearchCard
	source := s.dao.DB.Model(&model.EsSearchCard{})
	if srt == _sortDesc {
		source = source.Order("mtime DESC")
	} else if srt == _sortASC {
		source = source.Order("mtime ASC")
	}
	if title != "" {
		if title == "_" {
			title = "\\_"
		}
		source = source.Where("query_name like ?", "%"+title+"%")
	}
	source.Count(&count)
	if err = source.Offset((pn - 1) * ps).Limit(ps).Find(&intervenes).Error; err != nil {
		log.Error("InterveneList Error (%v)", err)
		return
	}
	if len(intervenes) == 0 {
		intervenes = _emptyInterveneList
		return
	}
	if list, err = s.interveneInfos(intervenes); err != nil {
		log.Error("s.interveneInfos Error (%v)", err)
	}
	return
}

// AddIntervene .
func (s *Service) AddIntervene(c context.Context, param *model.EsSearchCard) (err error) {
	var (
		details   []*model.EsSearchContest
		paramCids []int64
	)
	if param.Stime >= param.Etime {
		return fmt.Errorf("截止时间不得低于起始时间")
	}
	preData := new(model.EsSearchCard)
	s.dao.DB.Where("query_name=?", param.QueryName).First(&preData)
	if preData.ID > 0 {
		return fmt.Errorf("query名重复")
	}
	if err = json.Unmarshal([]byte(param.Detail), &paramCids); err != nil || len(paramCids) == 0 {
		return fmt.Errorf("detail 出错")
	}
	if len(paramCids) > _maxDetail {
		return fmt.Errorf("至多添加三个赛程")
	}
	if len(paramCids) == 0 {
		return fmt.Errorf("赛程不能为空")
	}
	// check detail not repeat
	if err = s.checkRepeat(paramCids); err != nil {
		return
	}
	param.Mtime = time.Now().Format("2006-01-02 15:04:05")
	tx := s.dao.DB.Begin()
	if err = tx.Model(&model.EsSearchCard{}).Create(param).Error; err != nil {
		log.Error("AddIntervene tx.Model Create(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	for _, cid := range paramCids {
		details = append(details, &model.EsSearchContest{Mid: param.ID, Cid: cid})
	}
	sql, sqlParam := model.BatchAddDSearchSQL(param.ID, details)
	if err = tx.Model(&model.EsSearchContest{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Error("AddIntervene Module tx.Model Create(%+v) error(%v)", sqlParam, err)
		err = tx.Rollback().Error
		return
	}
	if err = tx.Commit().Error; err != nil {
		log.Error("AddIntervene tx.Commit(%+v) error(%v)", sqlParam, err)
	}
	return
}

func (s *Service) checkRepeat(cids []int64) (err error) {
	var findCount int
	source := s.dao.DB.Model(&model.Contest{}).Where("id in (?)", cids)
	source.Count(&findCount)
	if findCount == 0 {
		err = fmt.Errorf("赛程不存在")
	} else if findCount < len(cids) {
		err = fmt.Errorf("赛程重复")
	}
	return
}

// EditIntervene .
func (s *Service) EditIntervene(c context.Context, param *model.EsSearchCard) (err error) {
	var (
		details   []*model.EsSearchContest
		paramCids []int64
	)
	if param.ID <= 0 {
		return fmt.Errorf("id不存在")
	}
	if param.Stime >= param.Etime {
		return fmt.Errorf("截止时间不得低于起始时间")
	}
	//check name not repeat.
	preData := new(model.EsSearchCard)
	s.dao.DB.Where("id != ?", param.ID).Where("query_name = ?", param.QueryName).First(&preData)
	if preData.ID > 0 {
		return fmt.Errorf("query名重复")
	}
	if err = json.Unmarshal([]byte(param.Detail), &paramCids); err != nil || len(paramCids) == 0 {
		return fmt.Errorf("detail 出错")
	}
	if len(paramCids) > _maxDetail {
		return fmt.Errorf("至多添加三个赛程")
	}
	if len(paramCids) == 0 {
		return fmt.Errorf("赛程不能为空")
	}
	// check detail not repeat
	if err = s.checkRepeat(paramCids); err != nil {
		return
	}
	param.Mtime = time.Now().Format("2006-01-02 15:04:05")
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("s.dao.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&model.EsSearchCard{}).Save(param).Error; err != nil {
		log.Error("EditIntervene Update(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	if err = tx.Model(&model.EsSearchContest{}).Where("mid = ? ", param.ID).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
		log.Error("EditIntervene s.dao.DB.Model mainID(%d) error(%v)", param.ID, err)
		err = tx.Rollback().Error
		return
	}
	for _, cid := range paramCids {
		details = append(details, &model.EsSearchContest{Mid: param.ID, Cid: cid})
	}
	sql, sqlParam := model.BatchAddDSearchSQL(param.ID, details)
	if err = tx.Model(&model.EsSearchContest{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Error("EditIntervene s.dao.DB.Model Create(%+v) error(%v)", details, err)
		err = tx.Rollback().Error
		return
	}
	if err = tx.Commit().Error; err != nil {
		log.Error("EditIntervene tx.Commit cid(%d) error(%v)", param.ID, err)
	}
	return
}

// ForbidIntervene .
func (s *Service) ForbidIntervene(c context.Context, id int64, state int) (err error) {
	preMain := new(model.EsSearchCard)
	if err = s.dao.DB.Where("id=?", id).First(&preMain).Error; err != nil {
		log.Error("ForbidIntervene s.dao.DB.Where id(%d) error(%d)", id, err)
		err = ecode.RequestErr
		return
	}
	if err = s.dao.DB.Model(&model.EsSearchCard{}).Where("id=?", id).Update(map[string]int{"status": state}).Error; err != nil {
		log.Error("ForbidIntervene s.dao.DB.Model error(%v)", err)
	}
	return
}

func (s *Service) interveneInfos(intervenes []*model.EsSearchCard) (list []*model.SearchInfo, err error) {
	var (
		mainIDs    []int64
		detailMap  map[int64][]*model.EsSearchContest
		details    []*model.EsSearchContest
		cids       []int64
		contests   []*model.Contest
		rsContest  map[int64]*model.ContestCard
		seasonName string
	)
	for _, v := range intervenes {
		mainIDs = append(mainIDs, v.ID)
	}
	if err = s.dao.DB.Model(&model.EsSearchContest{}).Where(map[string]interface{}{"is_deleted": _notDeleted}).Where("mid IN (?)", mainIDs).Find(&details).Error; err != nil {
		log.Error("ContestList Find ContestData Error (%v)", err)
		return
	}
	detailMap = make(map[int64][]*model.EsSearchContest, len(details))
	for _, v := range details {
		detailMap[v.Mid] = append(detailMap[v.Mid], v)
		cids = append(cids, v.Cid)
	}
	source := s.dao.DB.Model(&model.Contest{})
	source = source.Where("id in (?)", cids)
	if err = source.Find(&contests).Error; err != nil {
		log.Error("interveneInfos Error (%v)", err)
		return
	}
	if rsContest, err = s.fmtContest(contests); err != nil {
		return
	}
	for _, v := range intervenes {
		intervene := &model.SearchInfo{EsSearchCard: v}
		if len(details) > 0 {
			if ds, ok := detailMap[v.ID]; ok {
				for index, detail := range ds {
					if contest, ok := rsContest[detail.Cid]; ok {
						if index == _firstContest {
							seasonName = contest.SeasonName
						} else {
							seasonName = ""
						}
						intervene.Detail = append(intervene.Detail, &model.EsSearchContest{
							ID:          detail.ID,
							Mid:         detail.Mid,
							Cid:         detail.Cid,
							SeasonName:  seasonName,
							ContestName: contest.HomeName + " vs " + contest.AwayName,
						})
					}
				}
			}
		} else {
			intervene.Detail = _emptyInterveneDetail
		}
		list = append(list, intervene)
	}
	return
}

func (s *Service) fmtContest(contests []*model.Contest) (rs map[int64]*model.ContestCard, err error) {
	var (
		conIDs, teamIDs, seasonIDs []int64
		teamMap                    map[int64]*model.Team
		seasonMap                  map[int64]*model.Season
		hasTeam                    bool
	)
	for _, v := range contests {
		conIDs = append(conIDs, v.ID)
		if v.HomeID > 0 {
			teamIDs = append(teamIDs, v.HomeID)
		}
		if v.AwayID > 0 {
			teamIDs = append(teamIDs, v.AwayID)
		}
		if v.SuccessTeam > 0 {
			teamIDs = append(teamIDs, v.SuccessTeam)
		}
		seasonIDs = append(seasonIDs, v.Sid)
	}
	if ids := unique(teamIDs); len(ids) > 0 {
		var teams []*model.Team
		if err = s.dao.DB.Model(&model.Team{}).Where("id IN (?)", ids).Find(&teams).Error; err != nil {
			log.Error("fmtContest team Error (%v)", err)
			return
		}
		if len(teams) > 0 {
			hasTeam = true
		}
		teamMap = make(map[int64]*model.Team, len(teams))
		for _, v := range teams {
			teamMap[v.ID] = v
		}
	}
	if ids := unique(seasonIDs); len(ids) > 0 {
		var seasons []*model.Season
		if err = s.dao.DB.Model(&model.Season{}).Where("id IN (?)", ids).Find(&seasons).Error; err != nil {
			log.Error("fmtContest season Error (%v)", err)
			return
		}
		seasonMap = make(map[int64]*model.Season, len(seasons))
		for _, v := range seasons {
			seasonMap[v.ID] = v
		}
	}
	rs = make(map[int64]*model.ContestCard, len(contests))
	for _, v := range contests {
		contest := &model.ContestCard{Contest: v}
		if hasTeam {
			if team, ok := teamMap[v.HomeID]; ok {
				contest.HomeName = team.Title
			}
			if team, ok := teamMap[v.AwayID]; ok {
				contest.AwayName = team.Title
			}
			if team, ok := teamMap[v.SuccessTeam]; ok {
				contest.SuccessName = team.Title
			}
		}
		if season, ok := seasonMap[v.Sid]; ok {
			contest.SeasonName = season.Title
		}
		rs[v.ID] = contest
	}
	return
}
