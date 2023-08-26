package currency

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/currency"
)

const (
	_deleted    = 1
	_notDeleted = 0
)

// CurrencyList get currency list.
func (s *Service) CurrencyList(c context.Context, pn, ps int64) (list []*currency.CurrItem, count int64, err error) {
	source := s.dao.DB.Model(&currency.Currency{})
	if err = source.Count(&count).Error; err != nil {
		log.Error("CurrencyList count pn(%d) ps(%d) error (%v)", pn, ps, err)
		return
	}
	if count == 0 {
		return
	}
	var (
		currList []*currency.Currency
		currIDs  []int64
		relaList []*currency.Relation
	)
	if err = source.Order("id ASC").Offset((pn - 1) * ps).Limit(ps).Find(&currList).Error; err != nil {
		log.Error("CurrencyList list pn(%d) ps(%d) error (%v)", pn, ps, err)
		return
	}
	for _, v := range currList {
		currIDs = append(currIDs, v.ID)
	}
	if err = s.dao.DB.Model(&currency.Relation{}).Where("currency_id IN (?)", currIDs).Find(&relaList).Error; err != nil {
		log.Error("CurrencyList relation currIDs(%v) error (%v)", currIDs, err)
		return
	}
	relas := make(map[int64][]*currency.Relation, len(relaList))
	for _, v := range relaList {
		relas[v.CurrencyID] = append(relas[v.CurrencyID], v)
	}
	for _, v := range currList {
		tmp := &currency.CurrItem{Currency: v, Relation: make([]*currency.Relation, 0)}
		if val, ok := relas[v.ID]; ok {
			tmp.Relation = val
		}
		list = append(list, tmp)
	}
	return
}

// CurrencyList get currency list.
func (s *Service) CurrencyItem(c context.Context, id int64) (data *currency.CurrItem, err error) {
	curr := &currency.Currency{}
	if err = s.dao.DB.Model(&currency.Currency{}).Where("id = ?", id).First(curr).Error; err != nil {
		log.Error("CurrencyItem id(%d) error (%v)", id, err)
		return
	}
	data = &currency.CurrItem{Currency: curr, Relation: make([]*currency.Relation, 0)}
	var relations []*currency.Relation
	if err = s.dao.DB.Model(&currency.Relation{}).Where("currency_id = ?", id).Find(&relations).Error; err != nil {
		log.Error("CurrencyItem Relation id(%d) error (%v)", id, err)
		return
	}
	if len(relations) > 0 {
		data.Relation = relations
	}
	return
}

// AddCurrency add currency.
func (s *Service) AddCurrency(c context.Context, arg *currency.AddArg) (err error) {
	add := &currency.Currency{
		Name:  arg.Name,
		Unit:  arg.Unit,
		State: arg.State,
	}
	if err = s.dao.DB.Model(&currency.Currency{}).Create(add).Error; err != nil {
		log.Error("AddCurrency s.dao.DB.Model Create(%+v) error(%v)", add, err)
		return
	}
	if err = s.dao.UserCreate(c, add.ID); err != nil {
		return
	}
	err = s.dao.UserLogCreate(c, add.ID)
	return
}

// AddCurrency save currency data.
func (s *Service) SaveCurrency(c context.Context, arg *currency.SaveArg) (err error) {
	return s.dao.SaveCurrency(c, arg)
}

// AddRelation add relation data.
func (s *Service) AddRelation(c context.Context, arg *currency.RelationArg) (err error) {
	curr := new(currency.Currency)
	if err = s.dao.DB.Model(&currency.Currency{}).Where("id = ?", arg.CurrencyID).First(curr).Error; err != nil {
		log.Error("AddRelation check s.dao.DB.Model(%+v) error(%v)", arg, err)
		return
	}
	preData := new(currency.Relation)
	if err = s.dao.DB.Model(&currency.Relation{}).Where("currency_id = ? AND business_id = ? AND foreign_id = ?", arg.CurrencyID, arg.BusinessID, arg.ForeignID).Find(preData).Error; err != nil {
		if err != ecode.NothingFound {
			log.Error("AddRelation count s.dao.DB.Model(%+v) error(%v)", arg, err)
			return
		}
	}
	if preData.ID > 0 {
		if preData.IsDeleted == 0 {
			log.Warn("AddRelation exist(%+v)", arg)
			return
		}
		if err = s.dao.SaveCurrRelation(c, preData.ID, _notDeleted); err != nil {
			log.Error("AddRelation s.dao.SaveCurrRelation(%d) error(%v)", preData.ID, err)
		}
		return
	}
	rela := &currency.Relation{
		CurrencyID: arg.CurrencyID,
		BusinessID: arg.BusinessID,
		ForeignID:  arg.ForeignID,
	}
	if err = s.dao.DB.Model(&currency.Relation{}).Create(rela).Error; err != nil {
		log.Error("AddRelation rela(%+v) error(%v)", rela, err)
	}
	return
}

// DelRelation del relation data.
func (s *Service) DelRelation(c context.Context, id int64) (err error) {
	preData := new(currency.Relation)
	if err = s.dao.DB.Model(&currency.Relation{}).Where("id = ?", id).Find(preData).Error; err != nil {
		log.Error("DelRelation s.dao.DB.Model id(%d) error(%v)", id, err)
		return
	}
	if preData.ID == 0 {
		log.Warn("DelRelation not exist(%d)", id)
		return
	}
	return s.dao.SaveCurrRelation(c, id, _deleted)
}
