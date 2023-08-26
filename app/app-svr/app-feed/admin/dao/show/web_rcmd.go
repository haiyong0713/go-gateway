package show

import (
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
)

// WebRcmdAdd add  web card rcommand
func (d *Dao) WebRcmdAdd(param *show.WebRcmdAP) (err error) {
	if err = d.DB.Create(param).Error; err != nil {
		log.Error("dao.show.WebRcmdAdd error(%v)", err)
		return
	}
	return
}

// WebRcmdUpdate  web recommand update web card
func (d *Dao) WebRcmdUpdate(param *show.WebRcmdUP) (err error) {
	if err = d.DB.Model(&show.WebRcmdUP{}).Save(param).Error; err != nil {
		log.Error("dao.show.WebRcmdUpdate error(%v)", err)
		return
	}
	return
}

// WebRcmdDelete  web recommand delete cweb card
func (d *Dao) WebRcmdDelete(id int64) (err error) {
	up := map[string]interface{}{
		"deleted": common.Deleted,
	}
	if err = d.DB.Model(&show.WebRcmd{}).Where("id = ?", id).Update(up).Error; err != nil {
		log.Error("dao.show.WebRcmdDelete error(%v)", err)
		return
	}
	return
}

// WebRcmdFindByID  web recommand card table find by id
func (d *Dao) WebRcmdFindByID(id int64) (card *show.WebRcmd, err error) {
	card = &show.WebRcmd{}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	if err = d.DB.Model(&show.WebRcmd{}).Where("id = ?", id).Where(w).First(card).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			card = nil
			err = fmt.Errorf("ID为%d的数据不存在", id)
		} else {
			log.Error("dao.WebRcmdFindByID.findByID error(%v)", err)
		}
	}
	return
}

// WebRcmdOption option search web
func (d *Dao) WebRcmdOption(up *show.WebRcmdOption) (err error) {
	if err = d.DB.Model(&show.WebRcmdOption{}).Update(up).Error; err != nil {
		log.Error("dao.WebRcmdOption Updates(%+v) error(%v)", up, err)
	}
	return
}
