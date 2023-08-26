package dao

import (
	"context"

	xsql "go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"

	"github.com/jinzhu/gorm"
)

const (
	_batchAddTeamSql           = "insert into es_contest_teams (`contest_id`, `team_id`) values %s"
	_batchAddTeamWhenUpdateSql = "insert into es_contest_teams" +
		" (`contest_id`, `team_id`, `survival_rank`, `kill_number`, `score`, `rank_edit_status`) " +
		"values %s"
	_getContestIdsBySeasonId = "select id from es_contests where sid = ? and status = 0"
	contestTableName         = "es_contests"
	_ReplyTypeContest        = "27"

	// sql 常量
	_idFilter           = "id = ?"
	_validContestRecord = "is_deleted = ?"

	_distinctSidByTime = "select distinct sid as sid from es_contests where stime >= ? and stime <= ? and status = 0"
)

func (d *dao) ContestAddTransaction(ctx context.Context, contest *model.ContestModel, gameIds []int64, teamIds []int64, contestData []*model.ContestDataModel, adId int64) error {
	return d.orm.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Table(contestTableName).Model(&model.ContestModel{}).Create(contest).Error; err != nil {
			log.Errorc(ctx, "[Dao][ContestAddTransaction][Contest][Create][Error], err:%+v", err)
			return
		}
		contestId := contest.ID
		// gidMaps保存
		if err = d.contestGidMapUpdate(ctx, tx, contestId, gameIds); err != nil {
			log.Errorc(ctx, "[ContestAddTransaction][contestGidMapUpdate][Error], err:%+v", err)
			return
		}
		// es_contest_data保存
		if err = d.contestDataUpdate(ctx, tx, contestId, contestData); err != nil {
			return
		}
		// 添加赛程队伍
		if err = d.BatchAddTeams(ctx, tx, contestId, teamIds); err != nil {
			return
		}

		// register reply
		if err = d.RegisterReply(ctx, contestId, adId, _ReplyTypeContest); err != nil {
			return
		}
		// 添加人群包, 此时无需关注天马卡任何信息，比赛开始时，操作天马卡的初始化+推送即可
		if err = d.InitBGroup(ctx, contest); err != nil {
			return
		}
		// 状态更新handler
		err = d.contestStatusUpdateHandler(ctx, contest, nil)
		d.clearCacheWhenAddContest(ctx, contest)
		return
	})
}
func (d *dao) ContestUpdateTransaction(ctx context.Context, contest *model.ContestModel, gameIds []int64, teamIds []int64, contestData []*model.ContestDataModel) (err error) {
	return d.orm.Transaction(func(tx *gorm.DB) (err error) {
		oldContest := new(model.ContestModel)

		if err = tx.Table(contestTableName).Model(&model.ContestModel{}).Where(_idFilter, contest.ID).Find(&oldContest).Error; err != nil {
			return
		}
		if contest.ExternalID == 0 && oldContest.ExternalID != 0 {
			contest.ExternalID = oldContest.ExternalID
		}
		// 待定逻辑，是否使用updateColumns进行替换
		updateModel := formatUpdateContestModel(contest)
		if err = tx.Table(contestTableName).Model(&model.ContestModel{}).Save(updateModel).Error; err != nil {
			log.Errorc(ctx, "[Dao][ContestAddTransaction][Contest][Create][Error], err:%+v", err)
			return
		}
		contestId := contest.ID
		// gidMaps保存
		if err = d.contestGidMapUpdate(ctx, tx, contestId, gameIds); err != nil {
			log.Errorc(ctx, "[ContestUpdateTransaction][contestGidMapUpdate][Error], err:%+v", err)
			return
		}
		// es_contest_data保存
		if err = d.contestDataUpdate(ctx, tx, contestId, contestData); err != nil {
			return
		}
		// 添加赛程队伍
		if err = d.contestTeamsUpdate(ctx, contest, teamIds, tx); err != nil {
			return
		}
		// 状态更新handler
		err = d.contestStatusUpdateHandler(ctx, contest, oldContest)
		d.clearCacheWhenUpdateContest(ctx, contest, oldContest)
		return
	})
}

func formatUpdateContestModel(contestModel *model.ContestModel) *model.ContestUpdateModel {
	return &model.ContestUpdateModel{
		ID:            contestModel.ID,
		GameStage:     contestModel.GameStage,
		Stime:         contestModel.Stime,
		Etime:         contestModel.Etime,
		HomeID:        contestModel.HomeID,
		AwayID:        contestModel.AwayID,
		HomeScore:     contestModel.HomeScore,
		AwayScore:     contestModel.AwayScore,
		LiveRoom:      contestModel.LiveRoom,
		Aid:           contestModel.Aid,
		Collection:    contestModel.Collection,
		Dic:           contestModel.Dic,
		Status:        contestModel.Status,
		Sid:           contestModel.Sid,
		Mid:           contestModel.Mid,
		Special:       contestModel.Special,
		SuccessTeam:   contestModel.SuccessTeam,
		SpecialName:   contestModel.SpecialName,
		SpecialTips:   contestModel.SpecialTips,
		SpecialImage:  contestModel.SpecialImage,
		Playback:      contestModel.Playback,
		CollectionURL: contestModel.CollectionURL,
		LiveURL:       contestModel.LiveURL,
		DataType:      contestModel.DataType,
		MatchID:       contestModel.MatchID,
		GameStage1:    contestModel.GameStage1,
		GameStage2:    contestModel.GameStage2,
		SeriesId:      contestModel.SeriesId,
		PushSwitch:    contestModel.PushSwitch,
		ContestStatus: contestModel.ContestStatus,
	}
}

func (d *dao) contestGidMapUpdate(ctx context.Context, tx *gorm.DB, contestId int64, gameIds []int64) (err error) {
	gidMaps := make([]*model.GidMapModel, 0)
	for _, v := range gameIds {
		gidMaps = append(gidMaps, &model.GidMapModel{Type: model.OidContestType, Oid: contestId, Gid: v})
	}
	oldGidMaps := make([]*model.GidMapModel, 0)
	err = tx.Table(gidMapTableName).Model(&model.GidMapModel{}).Where("oid = ?", contestId).
		Where("type = ?", model.OidContestType).
		Where(_validContestRecord, model.GidMapRecordNotDeleted).
		Find(&oldGidMaps).Error
	if err != nil {
		log.Errorc(ctx, "[ContestGidMapAdd][GetOldGidMaps][Error], err:%+v", err)
		return
	}
	// 只有一个游戏映射时才进行校验，通过则不变动，否则先删后加
	if len(oldGidMaps) == 1 && len(gidMaps) == 1 && oldGidMaps[0].Gid == gidMaps[0].Gid {
		return
	}
	if len(oldGidMaps) > 0 {
		err = tx.Table(gidMapTableName).Model(&model.GidMapModel{}).Where("oid = ?", contestId).
			Where("type = ?", model.OidContestType).
			Where(_validContestRecord, model.GidMapRecordNotDeleted).
			Update("is_deleted", model.GidMapRecordDeleted).Error
		if err != nil {
			log.Errorc(ctx, "[ContestGidMapAdd][GidMapsDelete][Error], err:%+v", err)
			return
		}
	}
	sql, sqlParam := d.gidBatchAddSQL(gidMaps)
	if err = tx.Table(gidMapTableName).Model(&model.GidMapModel{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Errorc(ctx, "[contestGidMapUpdate][gidBatchAddSQL][Error], err:%+v", err)
		err = xecode.Errorf(xecode.RequestErr, "新增[BatchAddGidMapSQL]保存es_gid_map表失败(%+v)", err)
		return
	}
	return
}

func (d *dao) contestDataUpdate(ctx context.Context, tx *gorm.DB, contestId int64, contestDataList []*model.ContestDataModel) (err error) {
	if contestDataList == nil {
		return
	}
	err = tx.Table(contestDataTableName).Model(&model.ContestDataModel{}).Where("cid = ?", contestId).
		Where(_validContestRecord, ContestDataRecordNotDeleted).
		Update("is_deleted", ContestDataRecordDeleted).Error
	if err != nil {
		log.Errorc(ctx, "[Dao][ContestDataDeleted][Error], err:%+v", err)
		return
	}
	if len(contestDataList) > 0 {
		sql, sqlParam := d.batchAddCDataSQL(contestId, contestDataList)
		if err = tx.Table(contestDataTableName).Model(&model.ContestDataModel{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Errorc(ctx, "[Dao][batchAddCDataSQL][Error] sqlParams(%+v) error(%v)", sqlParam, err)
			err = xecode.Errorf(xecode.RequestErr, "新增[BatchAddCDataSQL]保存es_contest_data表失败(%+v)", err)
			return
		}
	}
	return
}

func (d *dao) ContestContestStatusUpdate(ctx context.Context, contestId int64, contestStatus int64) (err error) {
	return d.orm.Transaction(func(tx *gorm.DB) (err error) {
		oldContest := new(model.ContestModel)
		if err = tx.Table(contestTableName).Model(&model.ContestModel{}).Where(_idFilter, contestId).Find(&oldContest).Error; err != nil {
			log.Errorc(ctx, "[ContestContestStatusUpdate][Find][Contest][Error], contestId(%d) err:%+v", contestId, err)
			return
		}
		if err = tx.Table(contestTableName).Model(&model.ContestModel{}).Where(_idFilter, contestId).Update(map[string]interface{}{
			"contest_status": contestStatus}).Error; err != nil {
			log.Errorc(ctx, "[ContestContestStatusUpdate][Update][Status][Error],  contestId(%d) err:%+v", contestId, err)
			return
		}
		newContest := new(model.ContestModel)
		*newContest = *oldContest
		newContest.ContestStatus = contestStatus
		if err = d.contestStatusUpdateHandler(ctx, newContest, oldContest); err != nil {
			log.Errorc(ctx, "[ContestContestStatusUpdate][Update][contestStatusUpdateHandler][Error],  contestId(%d) err:%+v", contestId, err)
			return
		}
		d.clearCacheWhenUpdateContest(ctx, newContest, oldContest)
		return
	})
}

func (d *dao) GetSeasonContestIds(ctx context.Context, seasonId int64) (contestIds []int64, err error) {
	var rows *xsql.Rows
	defer func() {
		if rows != nil {
			rows.Close()
			err = rows.Err()
		}
	}()
	if rows, err = d.db.Query(ctx, _getContestIdsBySeasonId, seasonId); err != nil {
		return
	}
	contestIds = make([]int64, 0)
	for rows.Next() {
		var contest model.ContestModel
		if err = rows.Scan(&contest.ID); err != nil {
			log.Errorc(ctx, "GetSeasonContestIds rows.Scan error: %v", err)
		}
		contestIds = append(contestIds, contest.ID)
	}
	return
}

func (d *dao) GetContestsByIds(ctx context.Context, contestIds []int64, valid bool) (contestModels []*model.ContestModel, err error) {
	contestModels = make([]*model.ContestModel, 0)
	query := d.orm.Table(contestTableName).Model(&model.ContestModel{}).Where("id in (?)", contestIds)
	if valid {
		query.Where("status = ?", model.FreezeFalse)
	}
	if err = query.Find(&contestModels).Error; err != nil {
		log.Errorc(ctx, "[dao][GetContestsByIds][Error], err:%+v", err)
		return
	}
	// 封装gameId信息， 后续通过增加es_contests表gid列解决
	gidMaps, errG := d.GetGidByOIds(ctx, contestIds, model.OidContestType)
	if errG != nil {
		log.Errorc(ctx, "[dao][GetContestsByIds][Error], err:%+v", err)
	}
	contestGidMap := make(map[int64]int64)
	for _, gidMap := range gidMaps {
		contestGidMap[gidMap.Oid] = gidMap.Gid
	}
	for _, contestModel := range contestModels {
		if gid, ok := contestGidMap[contestModel.ID]; ok {
			contestModel.GameId = gid
		}
	}
	return
}

func (d *dao) GetContestById(ctx context.Context, contestId int64, valid bool) (contestModel *model.ContestModel, err error) {
	contestModel = new(model.ContestModel)
	query := d.orm.Table(contestTableName).Model(&model.ContestModel{}).Where("id = ?", contestId)
	if valid {
		query.Where("status = ?", model.FreezeFalse)
	}
	if err = query.Find(&contestModel).Error; err != nil {
		log.Errorc(ctx, "[dao][GetContestsByIds][Error], err:%+v", err)
		return
	}
	// 封装gameId信息， 后续通过增加es_contests表gid列解决
	gidMaps, errG := d.GetGidByOIds(ctx, []int64{contestId}, model.OidContestType)
	if errG != nil {
		log.Errorc(ctx, "[dao][GetContestsByIds][Error], err:%+v", err)
	}
	for _, gidMap := range gidMaps {
		if gidMap.Oid == contestId {
			contestModel.GameId = gidMap.Gid
			break
		}
	}
	return
}

func (d *dao) GetDistinctSeasonByTime(ctx context.Context, beginTime int64, endTime int64) (seasonIds []int64, err error) {
	seasonIds = make([]int64, 0)
	rows, err := d.db.Query(ctx, _distinctSidByTime, beginTime, endTime)
	if err != nil {
		log.Errorc(ctx, "[dao][GetDistinctSeasonByTime][Error], err:%+v", err)
		if err == xsql.ErrNoRows {
			return
		}
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		distinctSeason := new(struct {
			Sid int64 `json:"sid"`
		})
		if err = rows.Scan(&distinctSeason.Sid); err != nil {
			log.Errorc(ctx, "[dao][GetDistinctSeasonByTime][Scan][Error], err:%+v", err)
			return
		}
		seasonIds = append(seasonIds, distinctSeason.Sid)
	}
	return
}
