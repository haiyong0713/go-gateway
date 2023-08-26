package manager

import (
	"strconv"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/card"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/manager"
)

// ContentCards .
func (d *Dao) ContentCards(ids []int64) (ret map[int64]*manager.ContentCard, err error) {
	var rawCards []*card.ResourceCard

	if err = d.DBResource.Model(&card.ResourceCard{}).
		Where("deleted = ? AND id IN (?)", common.NotDeleted, ids).
		Find(&rawCards).Error; err != nil {
		log.Error("ResourceCard param(%v) error(%v)", ids, err)
		return
	}

	ret = make(map[int64]*manager.ContentCard)
	for _, c := range rawCards {
		var contCard *card.ContentCard
		if contCard, err = card.ParseContentCard(c); err != nil {
			log.Error("ParseContentCard id(%v) error(%+v)", c.Id, err)
			continue
		}

		ret[c.Id] = &manager.ContentCard{
			ID:     contCard.Id,
			Title:  contCard.Title,
			Weight: 0, // 和AI确认，AI不用会用到该值，manager后台无入口配置，默认传0
		}

		if contCard.Cover != nil {
			ret[c.Id].Cover = contCard.Cover.MCover
		}
		if contCard.Jump != nil {
			ret[c.Id].ReType = int64(contCard.Jump.ReType)
			ret[c.Id].ReValue = contCard.Jump.ReValue
		}
		if contCard.Button != nil {
			ret[c.Id].BtnReType = int64(contCard.Button.ReType)
			ret[c.Id].BtnReValue = contCard.Button.ReValue
		}

		if contCard.Deleted == common.NotDeleted {
			ret[c.Id].State = 1 // 上线
		} else {
			ret[c.Id].State = 0 // 下线
		}

		contents := make([]*manager.Contents, len(contCard.Content))
		for i, v := range contCard.Content {
			contents[i] = &manager.Contents{
				Ctype:  strconv.Itoa(int(v.ReType)),
				Cvalue: v.ReValue,
			}
		}
		ret[c.Id].Contents = contents
	}
	return
}
