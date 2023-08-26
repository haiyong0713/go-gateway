package show

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
)

const (
	_insertTagSQL = "INSERT INTO popular_channel_tag(top_entrance_id,tag_id,deleted) VALUE(?,?,?) ON DUPLICATE KEY UPDATE top_entrance_id=values(top_entrance_id),tag_id=values(tag_id),deleted=values(deleted)"
)

func (d *Dao) PopChannelTagAddM(c context.Context, id int64, tagID []int64) (err error) {
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
	for _, item := range tagID {
		if err = d.PopChannelTagAdd(c, tx, &show.PopChannelTagAD{
			TagID:         item,
			TopEntranceId: id,
			Deleted:       common.NotDeleted,
		}); err != nil {
			return
		}
	}
	return
}

// PopChannelTagAdd add event topic
func (d *Dao) PopChannelTagAdd(ctx context.Context, db *gorm.DB, param *show.PopChannelTagAD) (err error) {
	if err = db.Model(&show.PopChannelResource{}).Exec(_insertTagSQL, param.TopEntranceId, param.TagID, common.NotDeleted).Error; err != nil {
		log.Error("dao.show.PopChannelTagAdd error(%v)", err)
	}
	return
}

// PopChannelTagDelete delete event topic
func (d *Dao) PopChannelTagDelete(ctx context.Context, topEntranceId, tagID int64) (err error) {
	up := map[string]interface{}{
		"deleted": common.Deleted,
	}
	if err = d.DB.Model(&show.PopChannelTag{}).Where("top_entrance_id=? and tag_id=?", topEntranceId, tagID).Update(up).Error; err != nil {
		log.Error("dao.show.PopChannelTagDelete error(%v)", err)
		return
	}
	return
}

// PopChannelTagNotDelete not delete event topic
func (d *Dao) PopChannelTagNotDelete(ctx context.Context, topEntranceId, tagID int64) (err error) {
	up := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	if err = d.DB.Model(&show.PopChannelTag{}).Where("top_entrance_id=? and tag_id=?", topEntranceId, tagID).Update(up).Error; err != nil {
		log.Error("dao.show.PopChannelTagNotDelete error(%v)", err)
		return
	}
	return
}

// PopCTFindByTEID search channel tags
func (d *Dao) PopCTFindByTEID(ctx context.Context, topEid int64) (res []*show.PopChannelTag, err error) {
	res = []*show.PopChannelTag{}
	if err = d.DB.Model(&show.PopChannelTag{}).Where("top_entrance_id=? and deleted=?", topEid, common.NotDeleted).Find(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			res = nil
			err = nil
		} else {
			log.Error("dao.PopCTFindByTEID error(%v)", err)
		}
	}
	return
}
