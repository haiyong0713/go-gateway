package search

import (
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/card"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	search "go-gateway/app/app-svr/app-feed/admin/model/search"
)

func (d *Dao) NavigationCards(ids []int64) (ret map[int64]*search.NavigationCard, err error) {
	var rawCards []*card.ResourceCard
	if err = d.DB.Model(&card.ResourceCard{}).
		Where("deleted = ? AND id IN (?)", common.NotDeleted, ids).
		Scan(&rawCards).Error; err != nil {
		return
	}

	ret = make(map[int64]*search.NavigationCard)
	for _, c := range rawCards {
		var navCard *card.NavigationCard
		if navCard, err = card.ParseNavigationCard(c); err != nil {
			log.Error("ParseNavigationCard id(%v) error(%v)", c.Id, err)
			continue
		}

		ret[c.Id] = &search.NavigationCard{
			CardId:     navCard.Id,
			Title:      navCard.Title,
			Desc:       navCard.Desc,
			Navigation: navCard.Navigation,
		}
		if navCard.Cover != nil {
			ret[c.Id].CoverType = navCard.Cover.Type
			ret[c.Id].CoverSunUrl = navCard.Cover.SunPic
			ret[c.Id].CoverNightUrl = navCard.Cover.NightPic
			ret[c.Id].CoverWidth = navCard.Cover.Width
			ret[c.Id].CoverHeight = navCard.Cover.Height
		}
		if navCard.Corner != nil {
			ret[c.Id].CornerType = navCard.Corner.Type
			ret[c.Id].CornerText = navCard.Corner.Text
			ret[c.Id].CornerSunUrl = navCard.Corner.SunPic
			ret[c.Id].CornerNightUrl = navCard.Corner.NightPic
			ret[c.Id].CornerWidth = navCard.Corner.Width
			ret[c.Id].CornerHeight = navCard.Corner.Height
		}
		if navCard.Button != nil {
			ret[c.Id].ButtonType = navCard.Button.Type
			ret[c.Id].ButtonText = navCard.Button.Text
			ret[c.Id].ButtonReType = navCard.Button.ReType
			ret[c.Id].ButtonReValue = navCard.Button.ReValue
		}
	}
	return
}
