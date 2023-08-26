package show

import (
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
)

const (
	IDLE_RE_VALUE = ""
)

// WebRcmdCardAdd add  web card rcommand
func (d *Dao) WebRcmdCardAdd(param *show.WebRcmdCardAP) (err error) {
	// 运营[阿牧]需求：跳转类型为url，跳转地址为特定字符串时，数据库库存空值，并且不下发该卡片给AI
	if param.ReType == show.WEB_RCMD_RE_TYPE_URL && param.ReValue == d.config.FeedConfig.SkipCardUrl {
		param.ReValue = IDLE_RE_VALUE
	}
	if err = d.DB.Create(param).Error; err != nil {
		log.Error("dao.show.WebRcmdCardAdd error(%v)", err)
		return
	}
	return
}

// WebRcmdCardUpdate  web recommand update web card
func (d *Dao) WebRcmdCardUpdate(param *show.WebRcmdCardUP) (err error) {
	// 运营[阿牧]需求：运营[阿牧]需求：跳转类型为url，跳转地址为特定字符串时，数据库库存空值，并且不下发该卡片给AI
	if param.ReType == show.WEB_RCMD_RE_TYPE_URL && param.ReValue == d.config.FeedConfig.SkipCardUrl {
		param.ReValue = IDLE_RE_VALUE
	}
	if err = d.DB.Model(&show.WebRcmdCardUP{}).
		Where("id = ? AND deleted = ?", param.ID, common.NotDeleted).
		Update(map[string]interface{}{
			"type":     param.Type,
			"title":    param.Title,
			"desc":     param.Desc,
			"cover":    param.Cover,
			"re_type":  param.ReType,
			"re_value": param.ReValue,
		}).Error; err != nil {
		log.Error("dao.show.WebRcmdCardUpdate error(%v)", err)
		return
	}
	return
}

// WebRcmdCardDelete  web recommand delete cweb card
func (d *Dao) WebRcmdCardDelete(id int64) (err error) {
	up := map[string]interface{}{
		"deleted": common.Deleted,
	}
	if err = d.DB.Model(&show.WebRcmdCard{}).Where("id = ?", id).Update(up).Error; err != nil {
		log.Error("dao.show.WebRcmdCardDelete error(%v)", err)
		return
	}
	return
}

// WebRcmdCardFindByID  web recommand card table find by id
func (d *Dao) WebRcmdCardFindByID(id int64) (card *show.WebRcmdCard, err error) {
	card = &show.WebRcmdCard{}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	if err = d.DB.Model(&show.WebRcmdCard{}).Where("id = ?", id).Where(w).First(card).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			card = nil
			err = nil
		} else {
			log.Error("dao.ormshow.WebRcmdCardFindByID.findByID error(%v)", err)
		}
	}
	return
}
