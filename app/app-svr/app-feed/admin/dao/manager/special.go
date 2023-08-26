package manager

import (
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/manager"
)

// EventTopicDelete delete cevent topic
func (d *Dao) SpecialCards(ids []int64) (res map[int64]*manager.SpecialCard, err error) {
	var (
		special []*manager.SpecialCard
	)
	if err = d.DB.Model(&manager.SpecialCard{}).Where("id IN (?)", ids).Find(&special).Error; err != nil {
		log.Error("SpecialCards param(%v) error(%v)", ids, err)
		return
	}
	res = make(map[int64]*manager.SpecialCard, len(special))
	for _, v := range special {
		res[v.ID] = v
	}
	return
}
