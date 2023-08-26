package show

import (
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
)

// PopRecommendAdd add event topic
func (d *Dao) PopRecommendAdd(param *show.PopRecommendAP) (err error) {
	if err = d.DB.Create(param).Error; err != nil {
		log.Error("dao.show.PopRecommendAdd error(%v)", err)
		return
	}
	return
}

// PopRecommendUpdate update event topic
func (d *Dao) PopRecommendUpdate(param *show.PopRecommendUP) (err error) {
	if err = d.DB.Model(&show.PopRecommendUP{}).Where("id = ?", param.ID).Update(map[string]interface{}{"card_value": param.CardValue, "label": param.Label, "cover_gif": param.CoverGif}).Error; err != nil {
		log.Error("dao.show.PopRecommendUpdate error(%v)", err)
		return
	}
	return
}

// PopRecommendDelete delete cevent topic
func (d *Dao) PopRecommendDelete(id int64) (err error) {
	up := map[string]interface{}{
		"deleted": common.Deleted,
	}
	if err = d.DB.Model(&show.PopRecommend{}).Where("id = ?", id).Update(up).Error; err != nil {
		log.Error("dao.show.PopRecommendDelete error(%v)", err)
		return
	}
	return
}

// PopRFindByID search web card table find by id
func (d *Dao) PopRFindByID(id string) (card *show.PopRecommend, err error) {
	card = &show.PopRecommend{}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	if err = d.DB.Model(&show.PopRecommend{}).Where("card_value = ?", id).Where(w).First(card).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			card = nil
			err = nil
		} else {
			log.Error("dao.PopRFindByID error(%v)", err)
		}
	}
	return
}
