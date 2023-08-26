package card

import (
	"context"

	model "go-gateway/app/app-svr/app-feed/admin/model/card"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/ecode"
)

func (d *Dao) ResourceCardAdd(c context.Context, card *model.ResourceCard) (cardId int64, err error) {
	if err = d.DB.Model(&model.ResourceCard{}).Create(card).Error; err != nil {
		return
	}
	return card.Id, nil
}

func (d *Dao) ResourceCardUpdate(c context.Context, card *model.ResourceCard) (oldCard *model.ResourceCard, err error) {
	oldCard = &model.ResourceCard{}
	db := d.DB.Model(&model.ResourceCard{}).
		Where("card_type = ?", card.CardType).
		Where("id = ? AND deleted = ?", card.Id, common.NotDeleted).
		Scan(oldCard).
		Update(map[string]interface{}{
			"title":      card.Title,
			"desc":       card.Desc,
			"cover":      card.Cover,
			"corner":     card.Corner,
			"button":     card.Button,
			"jump_info":  card.JumpInfo,
			"extra_info": card.ExtraInfo,
			"m_uname":    card.MUname,
		})
	if db.RecordNotFound() {
		return nil, ecode.CardNotFound
	}
	if err = db.Error; err != nil {
		return
	}
	return
}

func (d *Dao) ResourceCardDelete(c context.Context, username string, cardId int64, cardType string) (oldCard *model.ResourceCard, err error) {
	oldCard = &model.ResourceCard{}
	db := d.DB.Model(&model.ResourceCard{}).
		Where("card_type = ?", cardType).
		Where("id = ? AND deleted = ?", cardId, common.NotDeleted).
		Scan(oldCard).
		Update(map[string]interface{}{
			"deleted": common.Deleted,
			"m_uname": username,
		})
	if db.RecordNotFound() {
		return nil, ecode.CardNotFound
	}
	if err = db.Error; err != nil {
		return
	}
	return
}

func (d *Dao) ResourceCardQuery(c context.Context, cardId int64, cardType string) (ret *model.ResourceCard, err error) {
	ret = &model.ResourceCard{}
	db := d.DB.Model(&model.ResourceCard{}).
		Where("card_type = ?", cardType).
		Where("id = ? AND deleted = ?", cardId, common.NotDeleted).
		Scan(ret)
	if db.RecordNotFound() {
		return nil, ecode.CardNotFound
	}
	if err = db.Error; err != nil {
		return
	}
	return
}

func (d *Dao) ResourceCardList(c context.Context, cardId int64, cardType string, keyword string, pn, ps int) (total int, list []*model.ResourceCard, err error) {
	var (
		offset, limit int
	)

	if offset, limit, err = d.paginate(pn, ps); err != nil {
		return
	}

	db := d.DB.Model(&model.ResourceCard{}).Where("deleted = ?", common.NotDeleted)
	if cardId > 0 {
		db = db.Where("id = ?", cardId)
	}
	if len(cardType) > 0 {
		db = db.Where("card_type = ?", cardType)
	}
	if len(keyword) > 0 {
		kwd := "%" + keyword + "%"
		db = db.Where("title LIKE ?", kwd)
	}

	if err = db.Count(&total).
		Order("id desc").
		Offset(offset).
		Limit(limit).
		Scan(&list).Error; err != nil {
		return
	}

	return
}

func (d *Dao) paginate(pn, ps int) (offset, limit int, err error) {
	limit = ps
	if pn < 1 {
		err = ecode.PageInvalid
		return
	}

	offset = (pn - 1) * ps
	return
}
