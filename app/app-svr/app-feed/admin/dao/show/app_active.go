package show

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
)

// AAFindByID app active table find by id
func (d *Dao) AAFindByID(c context.Context, id int64) (active *show.AppActive, err error) {
	active = &show.AppActive{}
	if err = d.DB.Model(&show.AppActive{}).Where("id = ?", id).First(active).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			active = nil
			err = nil
		} else {
			log.Error("dao.ormshow.app_active.findByID error(%v)", err)
		}
	}
	return
}
