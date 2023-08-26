package selected

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"

	"github.com/pkg/errors"
)

const (
	_updatePos = "UPDATE selected_resource SET position = position + 1 WHERE serie_id = ? AND status = 1 AND deleted = 0" // status = 1即通过的卡片，拒绝的卡片没有位置
	_rejectPos = "UPDATE selected_resource SET position = position - 1 WHERE serie_id = ? AND position > ? AND status = 1 AND deleted = 0"
	_sortPos   = "UPDATE selected_resource SET position = ? WHERE id = ? AND deleted = 0"
)

// PickSerie picks one serie from DB with the type and the number
func (d *Dao) PickSerie(c context.Context, req *selected.FindSerie) (res *selected.Serie, err error) {
	res = &selected.Serie{}
	db := d.DB.Model(&selected.Serie{}).Where("deleted = 0")
	if req.Number != 0 && req.Type != "" {
		db = db.Where("type = ?", req.Type).Where("number = ?", req.Number)
	} else if req.ID != 0 {
		db = db.Where("id = ?", req.ID)
	}
	err = db.First(&res).Error
	return
}

// 获取本周要发布的系列
func (d *Dao) PickPublishSerie(c context.Context, typ string) (res *selected.Serie, err error) {
	res = new(selected.Serie)
	err = d.DB.Model(res).Where("deleted = 0").
		Where("date_sub(now(), INTERVAL 1 DAY) BETWEEN stime AND etime").
		Where("type = ?", typ).Order("ctime desc").First(&res).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return
}

// PickRes picks the resource with its ID
func (d *Dao) PickRes(c context.Context, id int64) (res *selected.Resource, err error) {
	res = &selected.Resource{}
	err = d.DB.Model(&selected.Resource{}).Where("id = ?", id).Where("deleted = 0").First(&res).Error
	return
}

// 通过 sid, source 获取 resource
func (d *Dao) PickResBySidSource(c context.Context, sid int64, source int) (res []*selected.Resource, err error) {
	err = d.DB.Model(&selected.Resource{}).
		Where("serie_id = ?", sid).
		Where("deleted = 0").
		Where("status = 1").
		Where("source = ?", source).
		Find(&res).Error
	return
}

// PickSeries picks series from DB and put them into a map
func (d *Dao) PickSeries(c context.Context, sids []int64) (res map[int64]*selected.Serie, err error) {
	var results []*selected.Serie
	if err = d.DB.Model(&selected.Serie{}).Where("deleted = 0").Where("id IN (?)", sids).Find(&results).Error; err != nil {
		return
	}
	if len(results) > 0 {
		res = make(map[int64]*selected.Serie, len(sids))
		for _, v := range results {
			res[v.ID] = v
		}
	}
	return
}

// Series picks all existing series
func (d *Dao) Series(c context.Context, sType string) (results []*selected.Serie, err error) {
	err = d.DB.Model(&selected.Serie{}).Where("type = ?", sType).Where("deleted = 0").Order("number DESC").Find(&results).Error
	return
}

// UpdateSerieStatus updates the serie's status
func (d *Dao) UpdateSerieStatus(c context.Context, sid int64, status int) (err error) {
	err = d.DB.Model(&selected.Serie{}).Where("id = ?", sid).Update(map[string]int{
		"status": status,
	}).Error
	return
}

// 更新播单 id
func (d *Dao) UpdateSerieMediaId(c context.Context, sid, mediaID int64) (err error) {
	err = d.DB.Model(&selected.Serie{}).Where("id = ?", sid).Update(map[string]interface{}{
		"media_id": mediaID,
	}).Error
	return
}

// AddRes 新增卡片且将其他卡片顺序顺延
func (d *Dao) AddRes(c context.Context, res *selected.Resource) (err error) {
	tx := d.DB.Begin()
	if err = tx.Exec(_updatePos, res.SerieID).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Model(&selected.Resource{}).Create(res).Error; err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit().Error
	return
}

// RejectRes rejects the given card and modify the position of the cards
func (d *Dao) RejectRes(c context.Context, origin *selected.Resource) (err error) {
	tx := d.DB.Begin()
	if err = tx.Exec(_rejectPos, origin.SerieID, origin.Position).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Model(&selected.Resource{}).Where("id = ?", origin.ID).Update(map[string]int{"status": 2}).Error; err != nil { // reject the card
		tx.Rollback()
		return
	}
	err = tx.Commit().Error
	return
}

// DelRes deletes the resource
func (d *Dao) DelRes(c context.Context, id int64) (err error) {
	err = d.DB.Exec("UPDATE selected_resource SET deleted = 1 WHERE id = ?", id).Error
	return
}

// UpdateRes def.
func (d *Dao) UpdateRes(c context.Context, origin *selected.Resource, req *selected.ReqSelEdit) (err error) {
	var (
		tx     = d.DB.Begin()
		maxPos = struct {
			Cnt int
		}{}
	)
	if origin.Rejected() { // 如果卡片被拒绝了，重新捞回排在最后一个
		if err = tx.Model(&selected.Resource{}).Where("serie_id = ?", origin.SerieID).Where("status = 1").Select("MAX(position) AS cnt").Scan(&maxPos).Error; err != nil {
			tx.Rollback()
			return
		}
		if err = tx.Model(&selected.Resource{}).Where("id = ?", req.ID).Update(map[string]int{"position": maxPos.Cnt + 1}).Error; err != nil {
			tx.Rollback()
			return
		}
	}
	if err = tx.Model(&selected.Resource{}).Where("id = ?", req.ID).Update(req.ToMap(origin)).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Commit().Error; err != nil {
		return
	}
	log.Info("UpdateRes Sid %d, CardID %d, Position %d", origin.SerieID, origin.ID, maxPos.Cnt)
	return
}

// CntRes only counts passed cards for sorting
func (d *Dao) CntRes(c context.Context, sid int64) (cnt int, err error) {
	err = d.DB.Model(&selected.Resource{}).Where("serie_id = ?", sid).
		Where("deleted = 0").Where("status = 1").Count(&cnt).Error
	return
}

// SortRes sorts the resources in one serie
func (d *Dao) SortRes(c context.Context, sid int64, cardIDs []int64) (err error) {
	var (
		tx       = d.DB.Begin()
		position = 1
	)
	for _, v := range cardIDs {
		if err = tx.Exec(_sortPos, position, v).Error; err != nil {
			tx.Rollback()
			return
		}
		position++
	}
	err = tx.Commit().Error
	return
}

// SerieValid checks whether the top three items of the given serie have recommend reason
func (d *Dao) SerieValid(c context.Context, sid int64) (valid bool, err error) {
	var cnt int
	if err = d.DB.Model(&selected.Resource{}).Where("serie_id = ?", sid).
		Where("position IN (1,2,3)").Where("rcmd_reason = ?", "").Where("deleted = 0").Where("status = 1").Count(&cnt).Error; err != nil {
		return
	}
	valid = cnt == 0
	return
}

// SeriePass def.
func (d *Dao) SeriePass(c context.Context, sid int64) (err error) {
	err = d.DB.Model(&selected.Serie{}).Where("id = ?", sid).Update(map[string]int{"status": 2}).Error // 2 means passed
	return
}

// SerieUpdate def.
func (d *Dao) SerieUpdate(c context.Context, serie *selected.SerieDB) (err error) {
	err = d.DB.Model(&selected.Serie{}).Update(serie).Error
	return
}

// SerieUpdate def.
func (d *Dao) SerieUpdatePush(_ context.Context, serie *selected.SeriePush) (err error) {
	err = d.DB.Model(&selected.Serie{}).Update(serie).Error
	return
}

// DuplicateCheck def.
func (d *Dao) DuplicateCheck(c context.Context, sid int64, rid int64, rtype string, selfID int64) (cnt int, err error) {
	db := d.DB.Model(&selected.Resource{}).Where("serie_id = ?", sid).Where("deleted = 0").Where("rid = ?", rid).Where("rtype = ?", rtype)
	if selfID != 0 {
		db = db.Where("id != ?", selfID)
	}
	err = db.Count(&cnt).Error
	return
}

// Resources def.
func (d *Dao) Resources(c context.Context, cardIDs []int64) (results map[int64]*selected.SelES, err error) {
	var cards []*selected.SelES
	results = make(map[int64]*selected.SelES, len(cardIDs))
	err = d.DB.Model(&selected.SelES{}).Where(fmt.Sprintf("id IN (%s)", xstr.JoinInts(cardIDs))).Where("deleted = 0").Find(&cards).Error
	for _, v := range cards {
		results[v.ID] = v
	}
	return
}

// SerieRes picks the resources of the given serie
func (d *Dao) PickValidResBySerieID(c context.Context, sid int64) (resources []*selected.Resource, err error) {
	err = d.DB.Model(&selected.Resource{}).Where("serie_id = ?", sid).
		Where("deleted = 0").Where("status = 1").Order("position").Find(&resources).Error
	return
}

// 根据序列 id 获取资源列表
func (d *Dao) PickResBySerieID(c context.Context, sid int64) (resources []*selected.Resource, err error) {
	err = d.DB.Model(&selected.Resource{}).Where("serie_id = ?", sid).Order("position").Find(&resources).Error
	return
}

// SerieUpdateTaskStatus def.
func (d *Dao) SerieUpdateTaskStatus(_ context.Context, serie *selected.Serie) (err error) {
	err = d.DB.Model(&selected.Serie{}).Where("id = ?", serie.ID).Update(map[string]int{
		"task_status": serie.TaskStatus,
	}).Error
	return
}

// SeriesNums 获取已审核的可用的每周必看期数
func (d *Dao) SeriesNums(_ context.Context) (num []int64, err error) {
	var series []*selected.Serie
	num = make([]int64, 0)
	err = d.DB.Model(&selected.Serie{}).
		Where("type = ?", "weekly_selected").Where("deleted = 0").
		Where("media_id != 0").Where("status in (2, 4)").
		Order("number DESC").Scan(&series).Error
	if err != nil {
		return num, nil
	}
	for _, v := range series {
		num = append(num, v.Number)
	}
	return num, err
}

// 每周必看天马推送
func (d *Dao) WeeklySelectedTunnel(c context.Context, number int64) (err error) {
	var res = &struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}
	err = retry.WithAttempts(c, "WeeklySelectedTunnel", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		err = d.HttpClient.Post(c, d.Config.Host.Manager+"/x/admin/popular-job/serie/weekly_selected/tunnel", "", nil, &res)
		return err
	})
	if err != nil {
		return errors.WithMessagef(err, "Dao WeeklySelectedTunnel HttpClient Get error:")
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Error(ecode.Code(res.Code), res.Message)
		return errors.WithMessagef(err, "Dao WeeklySelectedTunnel HttpClient Resspose error:")
	}
	return err
}

// 根据类型获取最新一期的有效数据
func (d *Dao) GetLastValidSerieByType(c context.Context, typ string) (serie *selected.Serie, err error) {
	serie = new(selected.Serie)
	err = d.DB.Model(serie).
		Where("type = ?", typ).Where("deleted = 0").
		Order("number DESC").First(&serie).Error
	if err != nil {
		// 未找到数据，赋空值
		if gorm.IsRecordNotFoundError(err) {
			return new(selected.Serie), nil
		}
		err = errors.WithMessagef(err, "Dao GetLastValidSerieByType type(%s) error:", typ)
		return nil, err
	}
	return
}

// 根据类型获取最新一期的有效数据
func (d *Dao) CreateSerie(c context.Context, serie *selected.Serie) (err error) {
	err = d.DB.Model(new(selected.Serie)).Create(serie).Error
	if err != nil {
		err = errors.WithMessagef(err, "CreateSerie(%+v) error:", serie)
		return
	}
	return
}
