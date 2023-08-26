package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup"
	"go-common/library/xstr"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/esports/admin/bvav"
	"go-gateway/app/web-svr/esports/admin/model"
	"go-gateway/app/web-svr/esports/ecode"

	liveRoom "git.bilibili.co/bapis/bapis-go/live/xroom"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	_arcsSize     = 50
	_ReplyTypeAct = "25"
)

var (
	_emptyModules    = make([]*model.Module, 0)
	_emptMatchMod    = make([]*model.MatchModule, 0)
	_emptyActive     = make([]*model.MatchModule, 0)
	_emptyActiveLive = make([]*model.Activelive, 0)
)

func liveRoomInfo(ctx context.Context, roomIDs []int64) (infos map[int64]*liveRoom.Infos, err error) {
	infos = make(map[int64]*liveRoom.Infos, 0)

	req := new(liveRoom.RoomIDsReq)
	{
		req.RoomIds = roomIDs
		req.Attrs = []string{"show", "status"}
	}

	res, fetchErr := liveRoomClient.GetMultiple(ctx, req)
	if fetchErr != nil {
		err = fetchErr

		return
	}

	if res == nil {
		err = errors.New(fmt.Sprintf("live room(%v) info is empty", roomIDs))

		return
	}

	if len(res.List) > 0 {
		infos = res.List
	} else {
		err = errors.New(fmt.Sprintf("live room(%v) info is not matched", roomIDs))
	}

	return
}

func isLiveRoomValid(ctx context.Context, roomIDs []int64) (isValid bool) {
	if len(roomIDs) == 0 {
		return
	}

	for _, v := range roomIDs {
		if v == 0 {
			return
		}
	}

	if d, err := liveRoomInfo(ctx, roomIDs); err == nil && len(d) == len(roomIDs) {
		isValid = true
	}

	return
}

// AddAct .
func (s *Service) AddAct(c context.Context, param *model.ParamMA) (arcs map[string][]int64, err error) {
	var ms []*model.Module
	if param.Modules != "" {
		if err = json.Unmarshal([]byte(param.Modules), &ms); err != nil {
			err = ecode.EsportsActModErr
			return
		}
		for _, v := range ms {
			if v.Oids, err = bvav.ToAvsStr(v.Oids); err != nil {
				return
			}
		}
		if arcs, err = s.checkArc(c, ms); err != nil {
			return
		}
	}
	var (
		actLives []*model.Activelive
	)

	if param.ActiveLive != "" {
		if err = json.Unmarshal([]byte(param.ActiveLive), &actLives); err != nil {
			return
		}
	}

	liveIDs := make([]int64, 0)
	{
		for _, v := range actLives {
			liveIDs = append(liveIDs, v.LiveId)
		}
	}

	if !isLiveRoomValid(c, liveIDs) {
		err = ecode.EsportsMatchLiveInvalid

		return
	}

	tx := s.dao.DB.Begin()
	if err = tx.Model(&model.MatchActive{}).Create(&param.MatchActive).Error; err != nil {
		log.Error("AddAct MatchActive tx.Model Create(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	maID := param.ID
	if len(ms) > 0 {
		sql, sqlParam := model.BatchAddModuleSQL(maID, ms)
		if err = tx.Model(&model.Module{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("AddAct Module tx.Model Create(%+v) error(%v)", sqlParam, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(actLives) > 0 {
		sql, sqlParam := model.BatchAddActLiveSQL(param.ID, actLives)
		if err = tx.Model(&model.Module{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("AddAct Module tx.Model BatchAddActLiveSQL(%+v) error(%v)", actLives, err)
			err = tx.Rollback().Error
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		return
	}
	// register reply
	if err = s.dao.RegReply(c, maID, param.Adid, _ReplyTypeAct); err != nil {
		err = nil
	}
	return
}

// EditAct .
func (s *Service) EditAct(c context.Context, param *model.ParamMA) (arcs map[string][]int64, err error) {
	var (
		tmpMs, upM, addM, ms []*model.Module
		mapMID               map[int64]int64
		pMID                 map[int64]*model.Module
		delM                 []int64
	)
	if param.Modules != "" {
		if err = json.Unmarshal([]byte(param.Modules), &ms); err != nil {
			err = ecode.EsportsActModErr
			return
		}
		for _, v := range ms {
			if v.Oids, err = bvav.ToAvsStr(v.Oids); err != nil {
				return
			}
		}
		if arcs, err = s.checkArc(c, ms); err != nil {
			return
		}
	}
	var (
		actLives []*model.Activelive
	)
	if param.ActiveLive != "" {
		if err = json.Unmarshal([]byte(param.ActiveLive), &actLives); err != nil {
			return
		}
	}

	liveIDs := make([]int64, 0)
	{
		for _, v := range actLives {
			liveIDs = append(liveIDs, v.LiveId)
		}
	}

	if !isLiveRoomValid(c, liveIDs) {
		err = ecode.EsportsMatchLiveInvalid

		return
	}

	// check module
	if err = s.dao.DB.Model(&model.Module{}).Where("ma_id=?", param.ID).Where("status=?", _notDeleted).Find(&tmpMs).Error; err != nil {
		log.Error("EditAct s.dao.DB.Model Find (%+v) error(%v)", param.ID, err)
		return
	}
	mapMID = make(map[int64]int64, len(tmpMs))
	for _, m := range tmpMs {
		mapMID[m.ID] = m.MaID
	}
	pMID = make(map[int64]*model.Module, len(ms))
	for _, m := range ms {
		if _, ok := mapMID[m.ID]; m.ID > 0 && !ok {
			err = ecode.EsportsActModNot
			return
		}
		pMID[m.ID] = m
		if m.ID == 0 {
			addM = append(addM, m)
		}
	}
	for _, m := range tmpMs {
		if mod, ok := pMID[m.ID]; ok {
			upM = append(upM, mod)
		} else {
			delM = append(delM, m.ID)
		}
	}
	// save
	tx := s.dao.DB.Begin()
	upFields := map[string]interface{}{"sid": param.Sid, "mid": param.Mid, "background": param.Background,
		"back_color": param.BackColor, "color_step": param.ColorStep, "live_id": param.LiveID, "intr": param.Intr,
		"focus": param.Focus, "url": param.URL, "status": param.Status,
		"h5_background": param.H5Background, "h5_back_color": param.H5BackColor,
		"h5_focus": param.H5Focus, "h5_url": param.H5URL, "intr_logo": param.IntrLogo, "intr_title": param.IntrTitle,
		"intr_text": param.IntrText, "is_live": param.IsLive, "sids": param.Sids}
	if err = tx.Model(&model.MatchActive{}).Where("id = ?", param.ID).Update(upFields).Error; err != nil {
		log.Error("EditAct MatchActive tx.Model Create(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	if len(upM) > 0 {
		moduleSql, moduleParam := model.BatchEditModuleSQL(upM)
		if err = tx.Model(&model.Module{}).Exec(moduleSql, moduleParam...).Error; err != nil {
			log.Error("EditAct Module tx.Model Exec(%+v) error(%v)", upM, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(delM) > 0 {
		if err = tx.Model(&model.Module{}).Where("id IN (?)", delM).Updates(map[string]interface{}{"status": _deleted}).Error; err != nil {
			log.Error("EditAct Module tx.Model Updates(%+v) error(%v)", delM, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(addM) > 0 {
		sql, sqlParam := model.BatchAddModuleSQL(param.ID, addM)
		if err = tx.Model(&model.Module{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("EditAct Module tx.Model Create(%+v) error(%v)", addM, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(actLives) > 0 {
		var (
			mapOldALData, mapNewALData     map[int64]*model.Activelive
			oldALData, upALData, addALData []*model.Activelive
			delALData                      []int64
		)
		// check active live
		if err = s.dao.DB.Model(&model.Activelive{}).Where("ma_id=?", param.ID).Where("is_deleted=?", _notDeleted).Find(&oldALData).Error; err != nil {
			log.Error("EditAct s.dao.DB.Model Find (%+v) error(%v)", param.ID, err)
			return
		}
		mapOldALData = make(map[int64]*model.Activelive, len(oldALData))
		for _, v := range oldALData {
			mapOldALData[v.ID] = v
		}
		//新数据在老数据中 更新老数据。新的数据不在老数据 添加新数据
		for _, alData := range actLives {
			if _, ok := mapOldALData[alData.ID]; ok {
				upALData = append(upALData, alData)
			} else {
				addALData = append(addALData, alData)
			}
		}
		mapNewALData = make(map[int64]*model.Activelive, len(oldALData))
		for _, v := range actLives {
			mapNewALData[v.ID] = v
		}
		//老数据在新中 上面已经处理。老数据不在新数据中 删除老数据
		for _, alData := range oldALData {
			if _, ok := mapNewALData[alData.ID]; !ok {
				delALData = append(delALData, alData.ID)
			}
		}
		if len(upALData) > 0 {
			sql, sqlParam := model.BatchEditActLiveSQL(upALData)
			if err = tx.Model(&model.Activelive{}).Exec(sql, sqlParam...).Error; err != nil {
				log.Error("EditAct s.dao.DB.Model tx.Model Exec(%+v) error(%v)", upALData, err)
				err = tx.Rollback().Error
				return
			}
		}
		if len(delALData) > 0 {
			if err = tx.Model(&model.Activelive{}).Where("id IN (?)", delALData).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
				log.Error("EditAct s.dao.DB.Model Updates(%+v) error(%v)", delALData, err)
				err = tx.Rollback().Error
				return
			}
		}
		if len(addALData) > 0 {
			upSql, sqlParam := model.BatchAddActLiveSQL(param.ID, addALData)
			if err = tx.Model(&model.Activelive{}).Exec(upSql, sqlParam...).Error; err != nil {
				log.Error("EditContest s.dao.DB.Model Create(%+v) error(%v)", addALData, err)
				err = tx.Rollback().Error
				return
			}
		}
	} else {
		if err = tx.Model(&model.Activelive{}).Where("ma_id IN (?)", param.ID).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
			log.Error("EditAct s.dao.DB.Model Updates(%+v) error(%v)", param.ID, err)
			err = tx.Rollback().Error
			return
		}
	}
	err = tx.Commit().Error
	return
}

func (s *Service) checkArc(c context.Context, ms []*model.Module) (rsAids map[string][]int64, err error) {
	var (
		name       string
		aids       []int64
		allAids    []int64
		tmpMap     map[int64]struct{}
		repeatAids []int64
		wrongAids  []int64
		isWrong    bool
	)
	rsAids = make(map[string][]int64, 2)
	for _, m := range ms {
		// check name only .
		if m.Name != "" && name == m.Name {
			err = ecode.EsportsModNameErr
			return
		}
		name = m.Name
		if aids, err = xstr.SplitInts(m.Oids); err != nil {
			err = xecode.RequestErr
			return
		}
		tmpMap = make(map[int64]struct{})
		for _, aid := range aids {
			if _, ok := tmpMap[aid]; ok {
				repeatAids = append(repeatAids, aid)
				continue
			}
			tmpMap[aid] = struct{}{}
		}
		allAids = append(allAids, aids...)
	}
	// check aids .
	if wrongAids, err = s.wrongArc(c, allAids); err != nil {
		err = ecode.EsportsArcServerErr
		return
	}
	if len(repeatAids) > 0 {
		rsAids["repeat"] = repeatAids
		isWrong = true
	}
	if len(wrongAids) > 0 {
		rsAids["wrong"] = wrongAids
		isWrong = true
	}
	if isWrong {
		err = ecode.EsportsModArcErr
	}
	return
}

func (s *Service) wrongArc(c context.Context, aids []int64) (list []int64, err error) {
	var (
		arcErr    error
		arcNormal map[int64]struct{}
		mutex     = sync.Mutex{}
	)
	group, errCtx := errgroup.WithContext(c)
	aidsLen := len(aids)
	arcNormal = make(map[int64]struct{}, aidsLen)
	for i := 0; i < aidsLen; i += _arcsSize {
		var partAids []int64
		if i+_arcsSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_arcsSize]
		}
		group.Go(func() (err error) {
			var tmpRes *arcmdl.ArcsReply
			if tmpRes, arcErr = s.arcClient.Arcs(errCtx, &arcmdl.ArcsRequest{Aids: partAids}); arcErr != nil {
				log.Error("wrongArc s.arcClient.Arcs(%v) error %v", partAids, err)
				return arcErr
			}
			if tmpRes != nil {
				for _, arc := range tmpRes.Arcs {
					if arc != nil && arc.IsNormal() {
						mutex.Lock()
						arcNormal[arc.Aid] = struct{}{}
						mutex.Unlock()
					}
				}
			}
			return nil
		})
	}
	if err = group.Wait(); err != nil {
		return
	}
	for _, aid := range aids {
		if _, ok := arcNormal[aid]; !ok {
			list = append(list, aid)
		}
	}
	return
}

// ForbidAct .
func (s *Service) ForbidAct(c context.Context, id int64, state int) (err error) {
	if err = s.dao.DB.Model(&model.MatchActive{}).Where("id=?", id).Updates(map[string]interface{}{"status": state}).Error; err != nil {
		log.Error("ForbidAct MatchActive s.dao.DB.Model Updates(%d) error(%v)", id, err)
	}
	return
}

// ListAct .
func (s *Service) ListAct(c context.Context, mid, pn, ps int64) (rs []*model.MatchModule, count int64, err error) {
	var (
		mas                                          []*model.MatchActive
		maIDs, matchIDs, seasonIDs                   []int64
		mapMs                                        map[int64][]*model.Module
		mapMatch                                     map[int64]*model.Match
		mapSeaon                                     map[int64]*model.Season
		matchTitle, matchSub, seasonTitle, seasonSub string
		mapActiveLive                                map[int64][]*model.Activelive
	)
	maDB := s.dao.DB.Model(&model.MatchActive{})
	if mid > 0 {
		maDB = maDB.Where("mid=?", mid)
	}
	maDB.Count(&count)
	if count == 0 {
		rs = _emptyActive
	}
	if err = maDB.Offset((pn - 1) * ps).Order("id ASC").Limit(ps).Find(&mas).Error; err != nil {
		log.Error("ListAct MatchActive s.dao.DB.Model Find error(%v)", err)
		return
	}
	if len(mas) == 0 {
		rs = _emptMatchMod
		return
	}
	for _, ma := range mas {
		maIDs = append(maIDs, ma.ID)
		matchIDs = append(matchIDs, ma.Mid)
		//sids不为空 以sids为准否则以sid为准
		if ma.Sids != "" {
			sids := strings.Split(ma.Sids, ",")
			if len(sids) > 0 {
				var i int64
				for _, v := range sids {
					if v == "" {
						continue
					}
					if i, err = strconv.ParseInt(v, 10, 64); err != nil {
						return
					}
					seasonIDs = append(seasonIDs, i)
				}
			}
		} else {
			seasonIDs = append(seasonIDs, ma.Sid)
		}
	}
	if ids := unique(matchIDs); len(ids) > 0 {
		var matchs []*model.Match
		if err = s.dao.DB.Model(&model.Match{}).Where("id IN (?)", ids).Find(&matchs).Error; err != nil {
			log.Error("ListAct match Error (%v)", err)
			return
		}
		mapMatch = make(map[int64]*model.Match, len(matchs))
		for _, v := range matchs {
			mapMatch[v.ID] = v
		}
	}
	if ids := unique(seasonIDs); len(ids) > 0 {
		var seasons []*model.Season
		if err = s.dao.DB.Model(&model.Match{}).Where("id IN (?)", ids).Find(&seasons).Error; err != nil {
			log.Error("ListAct season Error (%v)", err)
			return
		}
		mapSeaon = make(map[int64]*model.Season, len(seasonIDs))
		for _, v := range seasons {
			mapSeaon[v.ID] = v
		}
	}
	if mapMs, err = s.modules(maIDs, count); err != nil {
		log.Error("ListAct s.modules maIDs(%+v) faild(%+v)", maIDs, err)
		return
	}
	if mapActiveLive, err = s.ActiveLive(maIDs); err != nil {
		log.Error("ListAct s.ActiveLive maIDs(%+v) faild(%+v)", maIDs, err)
		return
	}
	for _, ma := range mas {
		if match, ok := mapMatch[ma.Mid]; ok {
			matchTitle = match.Title
			matchSub = match.SubTitle
		} else {
			matchTitle = ""
			matchSub = ""
		}
		if ma.Sids != "" {
			sids := strings.Split(ma.Sids, ",")
			var i int64
			seasonTitle = ""
			seasonSub = ""
			for _, v := range sids {
				if i, err = strconv.ParseInt(v, 10, 64); err != nil {
					return
				}
				if season, ok := mapSeaon[i]; ok {
					if seasonTitle == "" {
						seasonTitle = season.Title
						seasonSub = season.SubTitle
					} else {
						seasonTitle = seasonTitle + "," + season.Title
						seasonSub = seasonSub + "," + season.SubTitle
					}
				} else {
					seasonTitle = ""
					seasonSub = ""
				}
			}
		} else {
			if season, ok := mapSeaon[ma.Sid]; ok {
				seasonTitle = season.Title
				seasonSub = season.SubTitle
			} else {
				seasonTitle = ""
				seasonSub = ""
			}
		}
		if rsMs, ok := mapMs[ma.ID]; ok {
			tmpMs := rsMs
			rs = append(rs, &model.MatchModule{MatchActive: ma, Modules: tmpMs, MatchTitle: matchTitle, MatchSubTitle: matchSub, SeasonTitle: seasonTitle, SeasonSubTitle: seasonSub})
		} else {
			rs = append(rs, &model.MatchModule{MatchActive: ma, Modules: _emptyModules, MatchTitle: matchTitle, MatchSubTitle: matchSub, SeasonTitle: seasonTitle, SeasonSubTitle: seasonSub})
		}
	}
	for k, v := range rs {
		if al, ok := mapActiveLive[v.ID]; ok {
			rs[k].ActiveLive = al
		} else {
			rs[k].ActiveLive = _emptyActiveLive
		}
		for _, module := range v.Modules {
			if module.Bvids, err = bvav.EditToBvsStr(module.Oids); err != nil {
				log.Error("ListAct EditToBvsStr Values(%v) error(%v)", module.Oids, err)
				err = nil
				continue
			}
		}
	}
	return
}

func (s *Service) ActiveLive(maIDs []int64) (rs map[int64][]*model.Activelive, err error) {
	var (
		data []*model.Activelive
	)
	rs = make(map[int64][]*model.Activelive)
	if err = s.dao.DB.Model(&model.Activelive{}).Where("ma_id in(?)", maIDs).Where("is_deleted=?", _notDeleted).Find(&data).Order("ma_id ASC").Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return rs, nil
		}
		err = fmt.Errorf("ActiveLive Find error(%v)", err)
		return
	}
	for _, v := range data {
		rs[v.MaId] = append(rs[v.MaId], v)
	}
	return
}

func (s *Service) modules(maIDs []int64, count int64) (rs map[int64][]*model.Module, err error) {
	var ms []*model.Module
	if err = s.dao.DB.Model(&model.Module{}).Where("ma_id in(?)", maIDs).Where("status=?", _notDeleted).Find(&ms).Order("ma_id ASC").Error; err != nil {
		err = errors.Wrap(err, "modules map Model Find")
		return
	}
	rs = make(map[int64][]*model.Module, count)
	for _, m := range ms {
		tmpM := m
		rs[tmpM.MaID] = append(rs[tmpM.MaID], tmpM)
	}
	return
}
