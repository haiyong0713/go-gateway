package dao

import (
	"context"
	"fmt"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"strings"

	"go-gateway/app/web-svr/esports/service/internal/model"
)

const (
	// sql
	_gidMapInsertSQL = "INSERT INTO es_gid_map(`type`,`oid`,`gid`) VALUES %s"

	gidMapTableName = "es_gid_map"
)

//func (d *dao) getGidMapsByOid(ctx context.Context, tx *gorm.DB, oid int64, typeValue int) (gidMaps []*model.GidMapModel, err error) {
//	return
//}

// GidBatchAddSQL .
func (d *dao) gidBatchAddSQL(gidMap []*model.GidMapModel) (sql string, param []interface{}) {
	if len(gidMap) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range gidMap {
		rowStrings = append(rowStrings, "(?,?,?)")
		param = append(param, v.Type, v.Oid, v.Gid)
	}
	return fmt.Sprintf(_gidMapInsertSQL, strings.Join(rowStrings, ",")), param
}

func (d *dao) GetContestGameById(ctx context.Context, contestId int64) (gameModel *model.GameModel, err error) {
	gameModel = new(model.GameModel)
	gidMapModel := new(model.GidMapModel)
	if err = d.orm.Table(gidMapTableName).Where("type = ?", model.OidContestType).
		Where("oid = ?", contestId).
		Where("is_deleted = ?", model.IsDeletedFalse).Limit(1).Find(&gidMapModel).Error; err != nil {
		log.Errorc(ctx, "[Dao][GetContestGameById][Error], err:%+v", err)
		return
	}
	gameId := gidMapModel.Gid
	res, err := d.GetGamesByIds(ctx, []int64{gameId})
	if err != nil {
		log.Errorc(ctx, "[Dao][GetContestGameById][GetGamesByIds][Error], err:%+v", err)
		return
	}
	if len(res) == 0 {
		err = xecode.Errorf(xecode.RequestErr, "游戏信息不存在")
		return
	}
	gameModel = res[0]
	return
}

func (d *dao) GetGidByOIds(ctx context.Context, oids []int64, typeValue int64) (gidMapModels []*model.GidMapModel, err error) {
	gidMapModels = make([]*model.GidMapModel, 0)
	if err = d.orm.Table(gidMapTableName).Where("type = ?", typeValue).
		Where("oid in (?)", oids).
		Where("is_deleted = ?", model.IsDeletedFalse).Find(&gidMapModels).Error; err != nil {
		log.Errorc(ctx, "[Dao][GetGidByOIds][Error], err:%+v", err)
		return
	}
	return
}
