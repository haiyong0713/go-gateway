package aggregation

import (
	"context"
	"fmt"

	tag "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-feed/admin/model/aggregation"
	"go-gateway/app/app-svr/app-feed/admin/model/common"

	"github.com/jinzhu/gorm"
)

const (
	_maxTagIDs                = 20
	_aggDeleted               = 4
	_updateAggSQL             = "UPDATE hotword_aggregation SET state=? WHERE id=?"
	_delAggSQL                = "UPDATE hotword_aggregation SET state=4 WHERE id=?"
	_delAggTagRelationSQL     = "UPDATE hotword_aggregation_tag SET state=1 WHERE hotword_id=?"
	_delAggBiliTagRelationSQL = "UPDATE hotword_aggregation_tag SET state=1 WHERE hotword_id=?"
	_insertResourceSQL        = "INSERT INTO hotword_aggregation_resource(hotword_id,oid,tag_id,state) VALUE(?,?,?,?) ON DUPLICATE KEY UPDATE tag_id=values(tag_id),state=values(state)"
	_insertTagSQL             = "INSERT INTO hotword_aggregation_tag(hotword_id,tag_id,state) VALUE(?,?,?) ON DUPLICATE KEY UPDATE state=values(state)"
)

func removeAgg(tx *gorm.DB, id int64) (err error) {
	if err = tx.Exec(_delAggSQL, id).Error; err != nil {
		log.Error("[UpdateAggregation] tx.Exec() id(%d) error(%v)", id, err)
		return
	}
	if err = tx.Exec(_delAggTagRelationSQL, id).Error; err != nil {
		log.Error("[UpdateAggregation] tx.Exec() id(%d) error(%v)", id, err)
	}
	return
}

func addAggTag(tx *gorm.DB, hotID int64, tagID []int64) (err error) {
	if len(tagID) != 0 {
		for _, v := range tagID {
			tag := &aggregation.AggTag{
				TagID:     v,
				HotwordID: hotID,
			}
			if err = tx.Create(tag).Error; err != nil {
				log.Error("[addAggTag] d.DB.Create(tag) error(%v)", err)
				return
			}
		}
	}
	return
}

// AddAggregation .
func (d *Dao) AddAggregation(ctx context.Context, param aggregation.AggPub, tagID []int64) (id int64, err error) {
	tx := d.DB.Begin()
	if err = tx.Create(&param).Error; err != nil {
		tx.Rollback()
		log.Error("[AddAggregation]  d.DB.Create() error(%v)", err)
		return
	}
	if err = addAggTag(tx, param.ID, tagID); err != nil {
		tx.Rollback()
		log.Error("[AddAggregation] addAggTag() error(%v)", err)
		return
	}
	id = param.ID
	err = tx.Commit().Error
	return
}

// UpdateAggregation .
func (d *Dao) UpdateAggregation(ctx context.Context, param aggregation.AggPub, tagID []int64) (err error) {
	tx := d.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Error("panic in Transact:%+v", r)
			if err = tx.Rollback().Error; err != nil {
				log.Error("UpdateAggregation param(%+v) err(%+v)", param, err)
			}
			return
		}
		if err != nil {
			log.Error("UpdateAggregation param(%+v) error(%v)", param, err)
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("UpdateAggregation tx.Rollback() param(%+v) error(%v)", param, err1)
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("UpdateAggregation param(%+v) tx.Commit() error(%v)", param, err)
		}
	}()
	m := map[string]interface{}{
		"hot_title": param.HotTitle,
		"title":     param.Title,
		"subtitle":  param.SubTitle,
		"image":     param.Image,
	}
	if err = tx.Model(&aggregation.AggPub{}).Where("id=?", param.ID).Update(m).Error; err != nil {
		return
	}
	if err = tx.Exec(_delAggBiliTagRelationSQL, param.ID).Error; err != nil {
		return
	}
	for _, item := range tagID {
		if err = d.AggTagAdd(ctx, tx, &aggregation.AggTag{
			TagID:     item,
			HotwordID: param.ID,
			State:     common.StateOK,
		}); err != nil {
			return
		}
	}
	return
}

// AggOperate .
func (d *Dao) AggOperate(ctx context.Context, id int64, state int) (err error) {
	if state == _aggDeleted {
		tx := d.DB.Begin()
		if err = removeAgg(tx, id); err != nil {
			tx.Rollback()
			log.Error("[UpdateAggregation] removeAgg() error(%v)", err)
			return
		}
		if err = tx.Commit().Error; err != nil {
			return
		}
	} else {
		if err = d.DB.Exec(_updateAggSQL, state, id).Error; err != nil {
			log.Error("[AggOperate] tx.Exec() id(%d) error(%v)", id, err)
		}
	}
	return
}

// AggList Aggregation List .
func (d *Dao) AggList(ctx context.Context, param *aggregation.AggListReq, hotwordID []int64) (res *aggregation.AggListReply, err error) {
	var item []*aggregation.AggList
	res = &aggregation.AggListReply{}
	db := d.DB.Model(&aggregation.AggPub{}).Where("state!=4")
	//nolint:gomnd
	if param.State != 5 {
		db = db.Where("state=?", param.State)
	}
	if param.ID != 0 {
		db = db.Where("id=?", param.ID)
	}
	if param.HotTitle != "" {
		db = db.Where("hot_title LIKE ?", "%"+param.HotTitle+"%")
	}
	if len(hotwordID) != 0 {
		db = db.Where(fmt.Sprintf("id IN (%s)", xstr.JoinInts(hotwordID)))
	}
	if param.Order == aggregation.Ctime {
		if param.Sort == aggregation.Desc { // 倒序
			db = db.Order("ctime DESC")
		} else if param.Sort == aggregation.Asc {
			db = db.Order("ctime ASC")
		}
	}
	db.Count(&res.Pager.Total)
	if err = db.Offset((param.Pn - 1) * param.Ps).Limit(param.Ps).Find(&item).Error; err != nil {
		log.Error("[AggList] error(%v)", err)
		return
	}
	res.Pager.Num = param.Pn
	res.Pager.Size = param.Ps
	res.Items = item
	return
}

// FindByTagIDs .
func (d *Dao) FindByTagIDs(ctx context.Context, tagID []int64) (res []*aggregation.AggTag, err error) {
	if err = d.DB.Model(&aggregation.AggTag{}).Where(fmt.Sprintf("tag_id IN (%s)", xstr.JoinInts(tagID))).Where("state!=1").Find(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
			return
		}
		log.Error("[FindByTagIDs] tag_id(%s) error(%v)", xstr.JoinInts(tagID), err)
	}
	return
}

// TagIDByID .
func (d *Dao) TagIDByID(ctx context.Context, hotID int64) (res []*aggregation.AggTag, err error) {
	if err = d.DB.Model(&aggregation.AggTag{}).Where("hotword_id=?", hotID).Where("state!=1").Find(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
			return
		}
		log.Error("[TagIDByID] hotID(%d) error(%v)", hotID, err)
	}
	return
}

// TagIDByName .
func (d *Dao) TagIDByName(ctx context.Context, tagName string) (tagReply *tag.TagReply, err error) {
	if tagReply, err = d.tagClient.TagByName(ctx, &tag.TagByNameReq{Tname: tagName}); err != nil {
		log.Error("[NameByTagID] d.tagClient.TagByName() error(%v)", err)
	}
	return
}

// NameByTagID .
func (d *Dao) NameByTagID(ctx context.Context, tagIDs []int64) (tagsReply *tag.TagsReply, err error) {
	if tagsReply, err = d.tagClient.Tags(ctx, &tag.TagsReq{Tids: tagIDs}); err != nil {
		log.Error("[NameByTagID] d.tagClient.Tag() tag_id(%s) error(%v)", xstr.JoinInts(tagIDs), err)
		return
	}
	if tagsReply == nil || tagsReply.Tags == nil {
		log.Error("[NameByTagID] d.tagClient.Tag() tag_id(%s) nil reply", xstr.JoinInts(tagIDs))
		return nil, ecode.NothingFound
	}
	return
}

// AggView view list .
func (d *Dao) AggView(ctx context.Context, hotID int64) (views []*aggregation.CardList, err error) {
	var (
		res struct {
			Code int                     `json:"code"`
			List []*aggregation.CardList `json:"list"`
		}
	)
	if err = d.client.Get(ctx, fmt.Sprintf(d.AggURL, hotID), "", nil, &res); err != nil {
		log.Error("[AggView] d.client.Get() url(%s) hotID(%d) error(%v)", d.AggURL, hotID, err)
		err = ecode.NothingFound
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("[AggView] d.client.Get() url(%s) hotID(%d) error(%v)", d.AggURL, hotID, err)
		err = ecode.Int(res.Code)
		return
	}
	views = res.List
	return
}

// HotWordCount
func (d *Dao) HotWordCount(ctx context.Context, hotWord string) (count int, err error) {
	err = d.DB.Model(&aggregation.AggPub{}).Where("hot_title=?", hotWord).Where("state!=4").Count(&count).Error
	return
}

// FindNameByID .
func (d *Dao) FindNameByID(ctx context.Context, id int64) (res *aggregation.AggPub, err error) {
	res = &aggregation.AggPub{}
	if err = d.DB.Model(&aggregation.AggPub{}).Where("id=?", id).Where("state!=4").First(&res).Error; err != nil {
		log.Error("[FindNameByID] HotID(%d) is not find. error(%v)", id, err)
	}
	return
}

func (d *Dao) HotwordAggResourceAddM(c context.Context, id int64, rids []int64) (err error) {
	tx := d.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Error("panic in Transact:%+v", r)
			if err = tx.Rollback().Error; err != nil {
				log.Error("HotwordAggResourceAddM id(%d) rids(%v) err(%+v)", id, rids, err)
			}
			return
		}
		if err != nil {
			log.Error("HotwordAggResourceAddM id(%d) rids(%v) err(%+v)", id, rids, err)
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("HotwordAggResourceAddM tx.Rollback() id(%d) rids(%v) err(%+v)", id, rids, err1)
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("HotwordAggResourceAddM id(%d) rids(%v) err(%+v)", id, rids, err)
		}
	}()
	for _, item := range rids {
		if err = d.HotwordAggResourceAdd(c, tx, &aggregation.HotwordAggResource{
			Oid:       item,
			HotwordID: id,
			Deleted:   common.NotDeleted,
			State:     aggregation.DefaultState,
		}); err != nil {
			return
		}
	}
	return
}

func (d *Dao) HotwordAggResourceAdd(ctx context.Context, db *gorm.DB, param *aggregation.HotwordAggResource) (err error) {
	if err = db.Create(param).Error; err != nil {
		log.Error("dao.HotwordAggResourceAdd error(%v)", err)
		return
	}
	return
}

func (d *Dao) HotwordAggResourceState(ctx context.Context, hotwordID, rid, tagID int64, state int) (err error) {
	if err = d.DB.Model(&aggregation.HotwordAggResource{}).Exec(_insertResourceSQL, hotwordID, rid, tagID, state).Error; err != nil {
		log.Error("dao.HotwordAggResourceState hotwordID(%d) rid(%d) tagID(%d) state(%d)error(%v)", hotwordID, rid, tagID, state, err)
		return
	}
	return
}

func (d *Dao) AggTagAddM(c context.Context, id int64, tagID []int64) (err error) {
	tx := d.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Error("panic in Transact:%+v", r)
			if err = tx.Rollback().Error; err != nil {
				log.Error("AggTagAddM id(%d) tagID(%v) err(%+v)", id, tagID, err)
			}
			return
		}
		if err != nil {
			log.Error("AggTagAddM id(%d) tagID(%v) err(%+v)", id, tagID, err)
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("AggTagAddM tx.Rollback() id(%d) tagID(%v) err(%+v)", id, tagID, err1)
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("AggTagAddM id(%d) tagID(%v) err(%+v)", id, tagID, err)
		}
	}()
	for _, item := range tagID {
		if err = d.AggTagAdd(c, tx, &aggregation.AggTag{
			TagID:     item,
			HotwordID: id,
			State:     common.StateOK,
		}); err != nil {
			return
		}
	}
	return
}

func (d *Dao) AggTagAdd(ctx context.Context, db *gorm.DB, param *aggregation.AggTag) (err error) {
	if err = db.Model(&aggregation.AggTag{}).Exec(_insertTagSQL, param.HotwordID, param.TagID, common.StateOK).Error; err != nil {
		log.Error("dao.AggTagAdd param(%+v) error(%v)", param, err)
	}
	return
}

func (d *Dao) AggTagDelete(ctx context.Context, hotwordID, tagID int64) (err error) {
	up := map[string]interface{}{
		"state": common.StateBlock,
	}
	if err = d.DB.Model(&aggregation.AggTag{}).Where("hotword_id=? and tag_id=?", hotwordID, tagID).Update(up).Error; err != nil {
		log.Error("dao.AggTagDelete hotwordID(%d) tagID(%d) error(%v)", hotwordID, tagID, err)
		return
	}
	return
}

func (d *Dao) HwResourceByHwID(ctx context.Context, hotwordID int64) (res []*aggregation.HotwordAggResource, err error) {
	res = []*aggregation.HotwordAggResource{}
	if err = d.DB.Model(&aggregation.HotwordAggResource{}).Where("hotword_id=? and deleted=?", hotwordID, common.NotDeleted).Find(&res).Error; err != nil {
		log.Error("dao.HwRsrByHwID hotwordID(%d) error(%v)", hotwordID, err)
	}
	return
}

func (d *Dao) NamesByTagIDs(c context.Context, tagIDs []int64) (map[int64]*tag.Tag, error) {
	res := make(map[int64]*tag.Tag)
	pag := len(tagIDs)/_maxTagIDs + 1
	for i := 0; i < pag; i++ {
		maxIndex := (i + 1) * _maxTagIDs
		if maxIndex > len(tagIDs) {
			maxIndex = len(tagIDs)
		}
		tagTemp := tagIDs[i*_maxTagIDs : maxIndex]
		resTmp, err := d.NameByTagID(c, tagTemp)
		if err != nil {
			return nil, err
		}
		if len(resTmp.Tags) > 0 {
			for k, v := range resTmp.Tags {
				if _, ok := res[k]; !ok {
					res[k] = v
				}
			}
		}
	}
	return res, nil
}
