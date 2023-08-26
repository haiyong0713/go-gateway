package show

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
)

const (
	_defaultState      = 1
	_insertResourceSQL = "INSERT INTO popular_channel_resource(top_entrance_id,rid,tag_id,state) VALUE(?,?,?,?) ON DUPLICATE KEY UPDATE top_entrance_id=values(top_entrance_id),rid=values(rid),tag_id=values(tag_id),state=values(state)"
)

func (d *Dao) PopChannelResourceAddM(c context.Context, id int64, rid []int64) (err error) {
	tx := d.DB.BeginTx(c, nil)
	if tx.Error != nil {
		err = tx.Error
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	for _, item := range rid {
		if err = d.PopChannelResourceAdd(c, tx, &show.PopChannelResourceAD{
			RID:           item,
			TopEntranceId: id,
			Deleted:       common.NotDeleted,
			State:         _defaultState,
		}); err != nil {
			return
		}
	}
	return
}

// PopChannelResourceAdd add event topic
func (d *Dao) PopChannelResourceAdd(ctx context.Context, db *gorm.DB, param *show.PopChannelResourceAD) (err error) {
	if err = db.Create(param).Error; err != nil {
		log.Error("dao.show.PopChannelResourceAdd error(%v)", err)
		return
	}
	return
}

// PopChannelResourceState update state event topic
func (d *Dao) PopChannelResourceState(ctx context.Context, topEntranceId, rid, tagID int64, state int) (err error) {
	if err = d.DB.Model(&show.PopChannelResource{}).Exec(_insertResourceSQL, topEntranceId, rid, tagID, state).Error; err != nil {
		log.Error("dao.show.PopChannelResourceState error(%v)", err)
		return
	}
	return
}

// PopCTFindByTEID search channel resources
func (d *Dao) PopCRFindByTEID(ctx context.Context, topEid int64) (res []*show.PopChannelResource, err error) {
	res = []*show.PopChannelResource{}
	if err = d.DB.Model(&show.PopRecommend{}).Where("top_entrance_id=? and deleted=?", topEid, common.NotDeleted).Find(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			res = nil
			err = nil
		} else {
			log.Error("dao.PopCRFindByTEID error(%v)", err)
		}
	}
	return
}
