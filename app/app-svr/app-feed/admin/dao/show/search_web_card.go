package show

import (
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
)

// SearchWebCardAdd add search web card
func (d *Dao) SearchWebCardAdd(param *show.SearchWebCardAP) (err error) {
	if err = d.DB.Create(param).Error; err != nil {
		log.Error("dao.show.SearchWebCardAdd error(%v)", err)
		return
	}
	return
}

// SearchWebCardUpdate search update web card
func (d *Dao) SearchWebCardUpdate(param *show.SearchWebCardUP) (err error) {
	if err = d.DB.Model(&show.SearchWebCardUP{}).Update(param).Error; err != nil {
		log.Error("dao.show.SearchWebCardUpdate error(%v)", err)
		return
	}
	return
}

// SearchWebCardDelete search delete cweb card
func (d *Dao) SearchWebCardDelete(id int64) (err error) {
	up := map[string]interface{}{
		"deleted": common.Deleted,
	}
	if err = d.DB.Model(&show.SearchWebCard{}).Where("id = ?", id).Update(up).Error; err != nil {
		log.Error("dao.show.SearchWebCardDelete error(%v)", err)
		return
	}
	return
}

// SWBFindByID search web card table find by id
func (d *Dao) SWBFindByID(id int64) (card *show.SearchWebCard, err error) {
	card = &show.SearchWebCard{}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	if err = d.DB.Model(&show.SearchWebCard{}).Where("id = ?", id).Where(w).First(card).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			card = nil
			err = nil
		} else {
			log.Error("dao.ormshow.event_topic.findByID error(%v)", err)
		}
	}
	return
}

// SWBFindByID search web card table find by id
func (d *Dao) SWBFindByIDs(ids []int64) (ret []*show.SearchWebCard, err error) {
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	if err = d.DB.Model(&show.SearchWebCard{}).Where("id IN (?)", ids).Where(w).First(&ret).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ret = nil
			err = nil
		} else {
			log.Error("dao.ormshow.event_topic.findByIDs error(%v)", err)
		}
	}
	return
}
