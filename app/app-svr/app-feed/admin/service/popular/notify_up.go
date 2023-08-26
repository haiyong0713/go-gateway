package popular

import (
	"strconv"

	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/util"

	"github.com/jinzhu/gorm"
)

// Card .
type Card struct {
	ID           int64
	CardValue    string
	IsNotifyDone int
}

// TableName .
func (a Card) TableName() string {
	return "popular_card"
}

// CardSet .
type CardSet struct {
	Value string
}

// TableName .
func (a CardSet) TableName() string {
	return "card_set"
}

// NotifyUp notify up
func (s *Service) NotifyUp() (err error) {
	var (
		popularCards          []*Card
		notifyTmp             map[int64]interface{}
		id                    int64
		notifyIDs, popularIDs []int64
	)
	cTimeStr := util.CTimeStr()
	//筛选出没有通知的人
	w := map[string]interface{}{
		"is_delete":      common.NotDeleted,
		"check":          common.Pass,
		"card_type":      common.CardUpRcmdNew,
		"is_notify":      common.Notify,
		"is_notify_done": common.NotifyNotDone,
	}
	if err = s.showDao.DB.Model(&Card{}).Where(w).Where("stime <= ?", cTimeStr).
		Where("etime >= ?", cTimeStr).Find(&popularCards).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
			return
		}
		return
	}
	for _, v := range popularCards {
		popularIDs = append(popularIDs, v.ID)
	}
	if len(popularIDs) == 0 {
		log.Info("NotifyUp is emppty")
		return
	}
	if err = s.showDao.DB.Model(&Card{}).Where("id IN (?)", popularIDs).Updates(map[string]interface{}{"is_notify_done": common.NotifyDone}).Error; err != nil {
		log.Error("dao.SearchWebUpdate Updates(%+v) error(%v)", popularIDs, err)
		return
	}
	notifyTmp = make(map[int64]interface{})
	for _, popular := range popularCards {
		var (
			cardSet []*CardSet
			ids     []int64
		)
		w := map[string]interface{}{
			"deleted": common.NotDeleted,
			"type":    common.CardUpRcmdNew,
		}
		if ids, err = xstr.SplitInts(popular.CardValue); err != nil {
			log.Error("popular service NotifyUp xstr.SplitInts(%s) error(%+v)", popular.CardValue, err)
			return
		}
		if err = s.showDao.DB.Model(&CardSet{}).Where(w).Where("id in (?)", ids).
			Find(&cardSet).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return
		}
		for _, card := range cardSet {
			if card.Value == "" {
				continue
			}
			if id, err = strconv.ParseInt(card.Value, 10, 64); err != nil {
				log.Error("popular service NotifyUp ParseInt(%+v) error(%+v)", card, err)
				return
			}
			notifyTmp[id] = struct{}{}
		}
	}
	if len(notifyTmp) == 0 {
		log.Info("popular service NotifyUp is empty")
		return
	}
	for k := range notifyTmp {
		notifyIDs = append(notifyIDs, k)
	}
	if err = s.messageDao.NotifyPopular(notifyIDs); err != nil {
		return
	}
	return
}
